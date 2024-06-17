package binary

import (
	"errors"
	"io"
	"math"
	"sync"
	"unicode/utf8"
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/primitive"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/protowire"
)

// memory resize factor
const (
	defaultBufferSize = 4096
	growBufferFactor  = 1
	defaultListSize   = 5
	speculativeLength = 1 // prefixed bytes length when write complex fields
)

var (
	errDismatchPrimitive  = meta.NewError(meta.ErrDismatchType, "dismatch primitive types", nil)
	errInvalidDataSize    = meta.NewError(meta.ErrInvalidParam, "invalid data size", nil) // not used
	errInvalidTag         = meta.NewError(meta.ErrInvalidParam, "invalid tag in ReadMessageBegin", nil)
	errInvalidFieldNumber = meta.NewError(meta.ErrInvalidParam, "invalid field number", nil)
	errExceedDepthLimit   = meta.NewError(meta.ErrStackOverflow, "exceed depth limit", nil) // not used
	errInvalidDataType    = meta.NewError(meta.ErrRead, "invalid data type", nil)
	errUnknonwField       = meta.NewError(meta.ErrUnknownField, "unknown field", nil)
	errUnsupportedType    = meta.NewError(meta.ErrUnsupportedType, "unsupported type", nil)
	errNotImplemented     = meta.NewError(meta.ErrNotImplemented, "not implemted type", nil) // not used
	ErrConvert            = meta.NewError(meta.ErrConvert, "convert type error", nil)
	errDecodeField        = meta.NewError(meta.ErrRead, "invalid field data", nil)
)

// We append to an empty array rather than a nil []byte to get non-nil zero-length byte slices.
var emptyBuf [0]byte

// Serizalize data to byte array and reuse the memory
type BinaryProtocol struct {
	Buf  []byte
	Read int
}

var (
	bpPool = sync.Pool{
		New: func() interface{} {
			return &BinaryProtocol{
				Buf: make([]byte, 0, defaultBufferSize),
			}
		},
	}
)

func (p *BinaryProtocol) malloc(size int) ([]byte, error) {
	if size <= 0 {
		panic(errors.New("invalid size"))
	}

	l := len(p.Buf)
	c := cap(p.Buf)
	d := l + size

	if d > c {
		c += c >> growBufferFactor
		if d > c {
			c = d * 2
		}
		buf := rt.Growslice(byteType, *(*rt.GoSlice)(unsafe.Pointer(&p.Buf)), c)
		p.Buf = *(*[]byte)(unsafe.Pointer(&buf))
	}
	p.Buf = (p.Buf)[:d]

	return (p.Buf)[l:d], nil
}

// next ...
func (p *BinaryProtocol) next(size int) ([]byte, error) {
	if size <= 0 {
		panic(errors.New("invalid size"))
	}

	l := len(p.Buf)
	d := p.Read + size
	if d > l {
		return nil, io.EOF
	}

	ret := (p.Buf)[p.Read:d]
	p.Read = d
	return ret, nil
}

// Reset resets the buffer and read position
func (p *BinaryProtocol) Reset() {
	p.Read = 0
	p.Buf = p.Buf[:0]
}

// RawBuf returns the raw buffer of the protocol
func (p *BinaryProtocol) RawBuf() []byte {
	return p.Buf
}

// Left returns the left bytes to read
func (p *BinaryProtocol) Left() int {
	return len(p.Buf) - p.Read
}

// BinaryProtocol Method
func NewBinaryProtol(buf []byte) *BinaryProtocol {
	bp := bpPool.Get().(*BinaryProtocol)
	bp.Buf = buf
	return bp
}

func NewBinaryProtocolBuffer() *BinaryProtocol {
	bp := bpPool.Get().(*BinaryProtocol)
	return bp
}

func FreeBinaryProtocol(bp *BinaryProtocol) {
	bp.Reset()
	bpPool.Put(bp)
}

func (p *BinaryProtocol) Recycle() {
	p.Reset()
	bpPool.Put(p)
}

// Append Tag
func (p *BinaryProtocol) AppendTag(num proto.FieldNumber, typ proto.WireType) error {
	tag := uint64(num)<<3 | uint64(typ&7)
	if num > proto.MaxValidNumber || num < proto.MinValidNumber {
		return errInvalidFieldNumber
	}
	p.Buf = protowire.BinaryEncoder{}.EncodeUint64(p.Buf, tag)
	return nil
}

// Append Tag With FieldDescriptor by kind, you must use kind to write tag, because the typedesc when list has no tag
func (p *BinaryProtocol) AppendTagByKind(number proto.FieldNumber, kind proto.ProtoKind) error {
	return p.AppendTag(number, proto.Kind2Wire[kind])
}

// ConsumeTag parses b as a varint-encoded tag, reporting its length.
func (p *BinaryProtocol) ConsumeTag() (proto.FieldNumber, proto.WireType, int, error) {
	v, n := protowire.ConsumeVarint((p.Buf)[p.Read:])
	if n < 0 {
		return 0, 0, n, errInvalidTag
	}
	_, err := p.next(n)
	if v>>3 > uint64(math.MaxInt32) {
		return -1, 0, n, errUnknonwField
	}
	num, typ := proto.FieldNumber(v>>3), proto.WireType(v&7)
	if num < proto.MinValidNumber {
		return 0, 0, n, errInvalidFieldNumber
	}
	return num, typ, n, err
}

// ConsumeChildTag parses b as a varint-encoded tag, don't move p.Read
func (p *BinaryProtocol) ConsumeTagWithoutMove() (proto.FieldNumber, proto.WireType, int, error) {
	v, n := protowire.ConsumeVarint((p.Buf)[p.Read:])
	if n < 0 {
		return 0, 0, n, errInvalidTag
	}
	if v>>3 > uint64(math.MaxInt32) {
		return -1, 0, n, errUnknonwField
	}
	num, typ := proto.FieldNumber(v>>3), proto.WireType(v&7)
	if num < proto.MinValidNumber {
		return 0, 0, n, errInvalidFieldNumber
	}
	return num, typ, n, nil
}

// When encoding length-prefixed fields, we speculatively set aside some number of bytes
// for the length, encode the data, and then encode the length (shifting the data if necessary
// to make room).
func AppendSpeculativeLength(b []byte) ([]byte, int) {
	pos := len(b)
	b = append(b, "\x00\x00\x00\x00"[:speculativeLength]...) // the max length is 4
	return b, pos
}

func FinishSpeculativeLength(b []byte, pos int) []byte {
	mlen := len(b) - pos - speculativeLength
	msiz := protowire.SizeVarint(uint64(mlen))
	if msiz != speculativeLength {
		if cap(b) >= pos+msiz+mlen {
			b = b[:pos+msiz+mlen]
		} else {
			newSlice := make([]byte, pos+msiz+mlen)
			copy(newSlice, b)
			b = newSlice
		}
		copy(b[pos+msiz:], b[pos+speculativeLength:])
	}
	protowire.AppendVarint(b[:pos], uint64(mlen))
	return b
}

/**
 * Write methods
 */

// WriteBool
func (p *BinaryProtocol) WriteBool(value bool) error {
	if value {
		return p.WriteUint64(uint64(1))
	} else {
		return p.WriteUint64(uint64(0))
	}
}

// WriteInt32
func (p *BinaryProtocol) WriteInt32(value int32) error {
	p.Buf = protowire.BinaryEncoder{}.EncodeInt32(p.Buf, value)
	return nil
}

// WriteSint32
func (p *BinaryProtocol) WriteSint32(value int32) error {
	p.Buf = protowire.BinaryEncoder{}.EncodeSint32(p.Buf, value)
	return nil
}

// WriteUint32
func (p *BinaryProtocol) WriteUint32(value uint32) error {
	p.Buf = protowire.BinaryEncoder{}.EncodeUint32(p.Buf, value)
	return nil
}

// Writefixed32
func (p *BinaryProtocol) WriteFixed32(value uint32) error {
	p.Buf = protowire.BinaryEncoder{}.EncodeFixed32(p.Buf, value)
	return nil
}

// WriteSfixed32
func (p *BinaryProtocol) WriteSfixed32(value int32) error {
	p.Buf = protowire.BinaryEncoder{}.EncodeSfixed32(p.Buf, value)
	return nil
}

// WriteInt64
func (p *BinaryProtocol) WriteInt64(value int64) error {
	p.Buf = protowire.BinaryEncoder{}.EncodeInt64(p.Buf, value)
	return nil
}

// WriteSint64
func (p *BinaryProtocol) WriteSint64(value int64) error {
	p.Buf = protowire.BinaryEncoder{}.EncodeSint64(p.Buf, value)
	return nil
}

// WriteUint64
func (p *BinaryProtocol) WriteUint64(value uint64) error {
	p.Buf = protowire.BinaryEncoder{}.EncodeUint64(p.Buf, value)
	return nil
}

// Writefixed64
func (p *BinaryProtocol) WriteFixed64(value uint64) error {
	p.Buf = protowire.BinaryEncoder{}.EncodeFixed64(p.Buf, value)
	return nil
}

// WriteSfixed64
func (p *BinaryProtocol) WriteSfixed64(value int64) error {
	p.Buf = protowire.BinaryEncoder{}.EncodeSfixed64(p.Buf, value)
	return nil
}

// WriteFloat
func (p *BinaryProtocol) WriteFloat(value float32) error {
	p.Buf = protowire.BinaryEncoder{}.EncodeFloat32(p.Buf, value)
	return nil
}

// WriteDouble
func (p *BinaryProtocol) WriteDouble(value float64) error {
	p.Buf = protowire.BinaryEncoder{}.EncodeDouble(p.Buf, value)
	return nil
}

// WriteString
func (p *BinaryProtocol) WriteString(value string) error {
	if !utf8.ValidString(value) {
		return meta.NewError(meta.ErrInvalidParam, value, nil)
	}
	p.Buf = protowire.BinaryEncoder{}.EncodeString(p.Buf, value)
	return nil
}

// WriteBytes
func (p *BinaryProtocol) WriteBytes(value []byte) error {
	p.Buf = protowire.BinaryEncoder{}.EncodeBytes(p.Buf, value)
	return nil
}

// WriteEnum
func (p *BinaryProtocol) WriteEnum(value proto.EnumNumber) error {
	p.Buf = protowire.BinaryEncoder{}.EncodeInt64(p.Buf, int64(value))
	return nil
}

/*
 * WriteList
 * packed format：[tag][length][value value value value....]
 * unpacked format：[tag][(length)][value][tag][(length)][value][tag][(length)][value]....
 * accpet val type: []interface{}
 */
func (p *BinaryProtocol) WriteList(desc *proto.TypeDescriptor, val interface{}, cast bool, disallowUnknown bool, useFieldName bool) error {
	vs, ok := val.([]interface{})
	if !ok {
		return errDismatchPrimitive
	}
	fieldId := desc.BaseId()
	NeedMessageLen := true
	// packed List bytes format: [tag][length][(L)V][value][value]...
	if desc.IsPacked() && len(vs) > 0 {
		p.AppendTag(fieldId, proto.BytesType)
		var pos int
		p.Buf, pos = AppendSpeculativeLength(p.Buf)
		for _, v := range vs {
			if err := p.WriteBaseTypeWithDesc(desc.Elem(), v, NeedMessageLen, cast, disallowUnknown, useFieldName); err != nil {
				return err
			}
		}
		p.Buf = FinishSpeculativeLength(p.Buf, pos)
		return nil
	}

	// unpacked List bytes format: [T(L)V][T(L)V]...
	for _, v := range vs {
		// share the same field number for Tag
		if err := p.AppendTag(fieldId, proto.BytesType); err != nil {
			return err
		}

		if err := p.WriteBaseTypeWithDesc(desc.Elem(), v, NeedMessageLen, cast, disallowUnknown, useFieldName); err != nil {
			return err
		}
	}
	return nil
}

/*
 * WriteMap
 * Map bytes format: [Pairtag][Pairlength][keyTag(L)V][valueTag(L)V] [Pairtag][Pairlength][T(L)V][T(L)V]...
 * Pairtag = MapFieldnumber << 3 | wiretype, wiertype = proto.BytesType
 * accpet val type: map[string]interface{} or map[int]interface{} or map[interface{}]interface{}
 */
func (p *BinaryProtocol) WriteMap(desc *proto.TypeDescriptor, val interface{}, cast bool, disallowUnknown bool, useFieldName bool) error {
	baseId := desc.BaseId()
	MapKey := desc.Key()
	MapValue := desc.Elem()
	// check val is map[string]interface{} or map[int]interface{} or map[interface{}]interface{}
	var vs map[string]interface{}
	var vs2 map[int]interface{}
	var vs3 map[interface{}]interface{}
	var ok bool

	if vs, ok = val.(map[string]interface{}); !ok {
		if vs2, ok = val.(map[int]interface{}); !ok {
			if vs3, ok = val.(map[interface{}]interface{}); !ok {
				return errDismatchPrimitive
			}
		}
	}
	NeedMessageLen := true
	if vs != nil {
		for k, v := range vs {
			p.AppendTag(baseId, proto.BytesType)
			var pos int
			p.Buf, pos = AppendSpeculativeLength(p.Buf)
			p.AppendTag(1, MapKey.WireType())
			p.WriteString(k)
			p.AppendTag(2, MapValue.WireType())
			p.WriteBaseTypeWithDesc(MapValue, v, cast, NeedMessageLen, disallowUnknown, useFieldName)
			p.Buf = FinishSpeculativeLength(p.Buf, pos)
		}
	} else if vs2 != nil {
		for k, v := range vs2 {
			p.AppendTag(baseId, proto.BytesType)
			var pos int
			p.Buf, pos = AppendSpeculativeLength(p.Buf)
			p.AppendTag(1, MapKey.WireType())
			// notice: may have problem, when k is sfixed64/fixed64 or sfixed32/fixed32 there is no need to use varint
			// we had better add more code to judge the type of k if write fast
			// p.WriteInt64(int64(k))
			p.WriteBaseTypeWithDesc(MapKey, k, NeedMessageLen, cast, disallowUnknown, useFieldName) // the gerneral way
			p.AppendTag(2, MapValue.WireType())
			p.WriteBaseTypeWithDesc(MapValue, v, NeedMessageLen, cast, disallowUnknown, useFieldName)
			p.Buf = FinishSpeculativeLength(p.Buf, pos)
		}
	} else {
		for k, v := range vs3 {
			p.AppendTag(baseId, proto.BytesType)
			var pos int
			p.Buf, pos = AppendSpeculativeLength(p.Buf)
			p.AppendTag(1, MapKey.WireType())
			p.WriteBaseTypeWithDesc(MapKey, k, NeedMessageLen, cast, disallowUnknown, useFieldName) // the gerneral way
			p.AppendTag(2, MapValue.WireType())
			p.WriteBaseTypeWithDesc(MapValue, v, NeedMessageLen, cast, disallowUnknown, useFieldName)
			p.Buf = FinishSpeculativeLength(p.Buf, pos)
		}
	}

	return nil
}

/*
 * Write Message
 * accpet val type: map[string]interface{} or map[proto.FieldNumber]interface{}
 * message fields format: [fieldTag(L)V][fieldTag(L)V]...
 */
func (p *BinaryProtocol) WriteMessageFields(desc *proto.MessageDescriptor, val interface{}, cast bool, disallowUnknown bool, useFieldName bool) error {
	NeedMessageLen := true
	if useFieldName {
		for name, v := range val.(map[string]interface{}) {
			f := desc.ByName(name)
			if f == nil {
				if disallowUnknown {
					return errUnknonwField
				}
				// unknown field will skip when writing
				continue
			}
			if !f.IsMap() && !f.IsList() {
				if err := p.AppendTag(f.Number(), proto.Kind2Wire[f.Kind()]); err != nil {
					return meta.NewError(meta.ErrWrite, "append field tag failed", nil)
				}
			}

			if err := p.WriteAnyWithDesc(f.Type(), v, NeedMessageLen, cast, disallowUnknown, useFieldName); err != nil {
				return err
			}
		}
	} else {
		for fieldNumber, v := range val.(map[proto.FieldNumber]interface{}) {
			f := desc.ByNumber(fieldNumber)
			if f == nil {
				if disallowUnknown {
					return errUnknonwField
				}
				continue
			}

			if !f.IsMap() && !f.IsList() {
				if err := p.AppendTag(f.Number(), proto.Kind2Wire[f.Kind()]); err != nil {
					return meta.NewError(meta.ErrWrite, "append field tag failed", nil)
				}
			}

			if err := p.WriteAnyWithDesc(f.Type(), v, NeedMessageLen, cast, disallowUnknown, useFieldName); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteBaseType Fields with FieldDescriptor format: (L)V
func (p *BinaryProtocol) WriteBaseTypeWithDesc(desc *proto.TypeDescriptor, val interface{}, NeedMessageLen bool, cast bool, disallowUnknown bool, useFieldName bool) error {
	switch desc.Type() {
	case proto.BOOL:
		v, ok := val.(bool)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				v, err = primitive.ToBool(val)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
			}
		}
		p.WriteBool(v)
	case proto.ENUM:
		v, ok := val.(proto.EnumNumber)
		if !ok {
			return meta.NewError(meta.ErrConvert, "convert enum error", nil)
		}
		p.WriteEnum(v)
	case proto.INT32:
		v, ok := val.(int32)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				vv, err := primitive.ToInt64(val)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
				v = int32(vv)
			}
		}
		p.WriteInt32(v)
	case proto.SINT32:
		v, ok := val.(int32)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				vv, err := primitive.ToInt64(val)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
				v = int32(vv)
			}
		}
		p.WriteSint32(v)
	case proto.UINT32:
		v, ok := val.(uint32)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				vv, err := primitive.ToInt64(val)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
				v = uint32(vv)
			}
		}
		p.WriteUint32(v)
	case proto.INT64:
		v, ok := val.(int64)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				v, err = primitive.ToInt64(val)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
			}
		}
		p.WriteInt64(v)
	case proto.SINT64:
		v, ok := val.(int64)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				v, err = primitive.ToInt64(val)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
			}
		}
		p.WriteSint64(v)
	case proto.UINT64:
		v, ok := val.(uint64)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				vv, err := primitive.ToInt64(val)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
				v = uint64(vv)
			}
		}
		p.WriteUint64(v)
	case proto.SFIX32:
		v, ok := val.(int32)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				vv, err := primitive.ToInt64(val)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
				v = int32(vv)
			}
		}
		p.WriteSfixed32(v)
	case proto.FIX32:
		v, ok := val.(int32)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				vv, err := primitive.ToInt64(val)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
				v = int32(vv)
			}
		}
		p.WriteFixed32(uint32(v))
	case proto.FLOAT:
		v, ok := val.(float32)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				vfloat64, err := primitive.ToFloat64(val)
				v = float32(vfloat64)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
			}
		}
		p.WriteFloat(v)
	case proto.SFIX64:
		v, ok := val.(int64)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				v, err = primitive.ToInt64(val)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
			}
		}
		p.WriteSfixed64(v)
	case proto.FIX64:
		v, ok := val.(int64)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				v, err = primitive.ToInt64(val)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
			}
		}
		p.WriteFixed64(uint64(v))
	case proto.DOUBLE:
		v, ok := val.(float64)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				v, err = primitive.ToFloat64(val)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
			}
		}
		p.WriteDouble(v)
	case proto.STRING:
		v, ok := val.(string)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				v, err = primitive.ToString(val)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
			}
		}
		p.WriteString(v)
	case proto.BYTE:
		v, ok := val.([]byte)
		if !ok {
			return meta.NewError(meta.ErrConvert, "write bytes kind error", nil)
		}
		p.WriteBytes(v)
	case proto.MESSAGE:
		var ok bool
		var pos int
		// prefix message length
		if NeedMessageLen {
			p.Buf, pos = AppendSpeculativeLength(p.Buf)
		}

		if useFieldName {
			val, ok = val.(map[string]interface{})
		} else {
			val, ok = val.(map[proto.FieldNumber]interface{})
		}
		if !ok {
			return errDismatchPrimitive
		}
		msg := desc.Message()
		if err := p.WriteMessageFields(msg, val, cast, disallowUnknown, useFieldName); err != nil {
			return err
		}
		// write message length
		if NeedMessageLen {
			p.Buf = FinishSpeculativeLength(p.Buf, pos)
		}
	default:
		return errUnsupportedType
	}
	return nil
}

// WriteAnyWithDesc explain desc and val and write them into buffer
//   - LIST will be converted from []interface{}
//   - MAP will be converted from map[string]interface{} or map[int]interface{} or map[interface{}]interface{}
//   - MESSAGE will be converted from map[FieldNumber]interface{} or map[string]interface{}
func (p *BinaryProtocol) WriteAnyWithDesc(desc *proto.TypeDescriptor, val interface{}, NeedMessageLen bool, cast bool, disallowUnknown bool, useFieldName bool) error {
	switch {
	case desc.IsList():
		return p.WriteList(desc, val, cast, disallowUnknown, useFieldName)
	case desc.IsMap():
		return p.WriteMap(desc, val, cast, disallowUnknown, useFieldName)
	default:
		return p.WriteBaseTypeWithDesc(desc, val, NeedMessageLen, cast, disallowUnknown, useFieldName)
	}
}

/**
 * Read methods
 */

// ReadByte
func (p *BinaryProtocol) ReadByte() (value byte, err error) {
	buf, err := p.next(1)
	if err != nil {
		return value, err
	}
	return byte(buf[0]), err
}

// ReadBool
func (p *BinaryProtocol) ReadBool() (bool, error) {
	v, n := protowire.BinaryDecoder{}.DecodeBool((p.Buf)[p.Read:])
	if n < 0 {
		return false, errDecodeField
	}
	_, err := p.next(n)
	return v, err
}

// ReadInt containing INT32, SINT32, SFIX32, INT64, SINT64, SFIX64, UINT32, UINT64
func (p *BinaryProtocol) ReadInt(t proto.Type) (value int, err error) {
	switch t {
	case proto.INT32:
		n, err := p.ReadInt32()
		return int(n), err
	case proto.SINT32:
		n, err := p.ReadSint32()
		return int(n), err
	case proto.SFIX32:
		n, err := p.ReadSfixed32()
		return int(n), err
	case proto.INT64:
		n, err := p.ReadInt64()
		return int(n), err
	case proto.SINT64:
		n, err := p.ReadSint64()
		return int(n), err
	case proto.SFIX64:
		n, err := p.ReadSfixed64()
		return int(n), err
	case proto.UINT32:
		n, err := p.ReadUint32()
		return int(n), err
	case proto.UINT64:
		n, err := p.ReadUint64()
		return int(n), err
	default:
		return 0, errInvalidDataType
	}
}

// ReadI32
func (p *BinaryProtocol) ReadInt32() (int32, error) {
	value, n := protowire.BinaryDecoder{}.DecodeInt32((p.Buf)[p.Read:])
	if n < 0 {
		return value, errDecodeField
	}
	_, err := p.next(n)
	return value, err
}

// ReadSint32
func (p *BinaryProtocol) ReadSint32() (int32, error) {
	value, n := protowire.BinaryDecoder{}.DecodeSint32((p.Buf)[p.Read:])
	if n < 0 {
		return value, errDecodeField
	}
	_, err := p.next(n)
	return value, err
}

// ReadUint32
func (p *BinaryProtocol) ReadUint32() (uint32, error) {
	value, n := protowire.BinaryDecoder{}.DecodeUint32((p.Buf)[p.Read:])
	if n < 0 {
		return value, errDecodeField
	}
	_, err := p.next(n)
	return value, err
}

// ReadI64
func (p *BinaryProtocol) ReadInt64() (int64, error) {
	value, n := protowire.BinaryDecoder{}.DecodeInt64((p.Buf)[p.Read:])
	if n < 0 {
		return value, errDecodeField
	}
	_, err := p.next(n)
	return value, err
}

// ReadSint64
func (p *BinaryProtocol) ReadSint64() (int64, error) {
	value, n := protowire.BinaryDecoder{}.DecodeSint64((p.Buf)[p.Read:])
	if n < 0 {
		return value, errDecodeField
	}
	_, err := p.next(n)
	return value, err
}

// ReadUint64
func (p *BinaryProtocol) ReadUint64() (uint64, error) {
	value, n := protowire.BinaryDecoder{}.DecodeUint64((p.Buf)[p.Read:])
	if n < 0 {
		return value, errDecodeField
	}
	_, err := p.next(n)
	return value, err
}

// ReadVarint
func (p *BinaryProtocol) ReadVarint() (uint64, error) {
	value, n := protowire.BinaryDecoder{}.DecodeUint64((p.Buf)[p.Read:])
	if n < 0 {
		return value, errDecodeField
	}
	_, err := p.next(n)
	return value, err
}

// ReadFixed32
func (p *BinaryProtocol) ReadFixed32() (int32, error) {
	value, n := protowire.BinaryDecoder{}.DecodeFixed32((p.Buf)[p.Read:])
	if n < 0 {
		return int32(value), errDecodeField
	}
	_, err := p.next(n)
	return int32(value), err
}

// ReadSFixed32
func (p *BinaryProtocol) ReadSfixed32() (int32, error) {
	value, n := protowire.BinaryDecoder{}.DecodeSfixed32((p.Buf)[p.Read:])
	if n < 0 {
		return int32(value), errDecodeField
	}
	_, err := p.next(n)
	return int32(value), err
}

// ReadFloat
func (p *BinaryProtocol) ReadFloat() (float32, error) {
	value, n := protowire.BinaryDecoder{}.DecodeFloat32((p.Buf)[p.Read:])
	if n < 0 {
		return value, errDecodeField
	}
	_, err := p.next(n)
	return value, err
}

// ReadFixed64
func (p *BinaryProtocol) ReadFixed64() (int64, error) {
	value, n := protowire.BinaryDecoder{}.DecodeFixed64((p.Buf)[p.Read:])
	if n < 0 {
		return int64(value), errDecodeField
	}
	_, err := p.next(n)
	return int64(value), err
}

// ReadSFixed64
func (p *BinaryProtocol) ReadSfixed64() (int64, error) {
	value, n := protowire.BinaryDecoder{}.DecodeFixed64((p.Buf)[p.Read:])
	if n < 0 {
		return int64(value), errDecodeField
	}
	_, err := p.next(n)
	return int64(value), err
}

// ReadDouble
func (p *BinaryProtocol) ReadDouble() (float64, error) {
	value, n := protowire.BinaryDecoder{}.DecodeFixed64((p.Buf)[p.Read:])
	if n < 0 {
		return math.Float64frombits(value), errDecodeField
	}
	_, err := p.next(n)
	return math.Float64frombits(value), err
}

// ReadBytes return bytesData and the sum length of L、V in TLV
func (p *BinaryProtocol) ReadBytes() ([]byte, error) {
	value, n, all := protowire.BinaryDecoder{}.DecodeBytes((p.Buf)[p.Read:])
	if n < 0 {
		return emptyBuf[:], errDecodeField
	}
	_, err := p.next(all)
	return value, err
}

// ReadLength return dataLength, and move pointer in the begin of data
func (p *BinaryProtocol) ReadLength() (int, error) {
	value, n := protowire.BinaryDecoder{}.DecodeUint64((p.Buf)[p.Read:])
	if n < 0 {
		return 0, errDecodeField
	}
	_, err := p.next(n)
	return int(value), err
}

// ReadString
func (p *BinaryProtocol) ReadString(copy bool) (value string, err error) {
	bytes, n, all := protowire.BinaryDecoder{}.DecodeBytes((p.Buf)[p.Read:])
	if n < 0 {
		return "", errDecodeField
	}
	if copy {
		value = string(bytes)
	} else {
		v := (*rt.GoString)(unsafe.Pointer(&value))
		v.Ptr = rt.IndexPtr(*(*unsafe.Pointer)(unsafe.Pointer(&p.Buf)), byteTypeSize, p.Read+n)
		v.Len = int(all - n)
	}
	_, err = p.next(all)
	return
}

// ReadEnum
func (p *BinaryProtocol) ReadEnum() (proto.EnumNumber, error) {
	value, n := protowire.BinaryDecoder{}.DecodeUint64((p.Buf)[p.Read:])
	if n < 0 {
		return 0, errDecodeField
	}
	_, err := p.next(n)
	return proto.EnumNumber(value), err
}

// ReadList
func (p *BinaryProtocol) ReadList(desc *proto.TypeDescriptor, copyString bool, disallowUnknown bool, useFieldName bool) ([]interface{}, error) {
	hasMessageLen := true
	elemetdesc := desc.Elem()
	// Read ListTag
	fieldNumber, _, _, listTagErr := p.ConsumeTag()
	if listTagErr != nil {
		return nil, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
	}
	list := make([]interface{}, 0, defaultListSize)
	// packed list
	if desc.IsPacked() {
		// read length
		length, err := p.ReadLength()
		if err != nil {
			return nil, err
		}
		// read list
		start := p.Read
		for p.Read < start+length {
			v, err := p.ReadBaseTypeWithDesc(elemetdesc, hasMessageLen, copyString, disallowUnknown, useFieldName)
			if err != nil {
				return nil, err
			}
			list = append(list, v)
		}
	} else {
		// unpacked list
		v, err := p.ReadBaseTypeWithDesc(elemetdesc, hasMessageLen, copyString, disallowUnknown, useFieldName)
		if err != nil {
			return nil, err
		}
		list = append(list, v)

		for p.Read < len(p.Buf) {
			// don't move p.Read and judge whether readList completely
			elementFieldNumber, _, tagLen, err := p.ConsumeTagWithoutMove()
			if err != nil {
				return nil, err
			}
			if elementFieldNumber != fieldNumber {
				break
			}

			if _, moveTagErr := p.next(tagLen); moveTagErr != nil {
				return nil, moveTagErr
			}

			v, err := p.ReadBaseTypeWithDesc(elemetdesc, hasMessageLen, copyString, disallowUnknown, useFieldName)
			if err != nil {
				return nil, err
			}
			list = append(list, v)
		}
	}
	return list, nil
}

func (p *BinaryProtocol) ReadPair(keyDesc *proto.TypeDescriptor, valueDesc *proto.TypeDescriptor, copyString bool, disallowUnknown bool, useFieldName bool) (interface{}, interface{}, error) {
	hasMessageLen := true
	if _, _, _, err := p.ConsumeTag(); err != nil {
		return nil, nil, err
	}

	key, err := p.ReadBaseTypeWithDesc(keyDesc, hasMessageLen, copyString, disallowUnknown, useFieldName)
	if err != nil {
		return nil, nil, err
	}

	if _, _, _, err := p.ConsumeTag(); err != nil {
		return nil, nil, err
	}
	value, err := p.ReadBaseTypeWithDesc(valueDesc, hasMessageLen, copyString, disallowUnknown, useFieldName)
	if err != nil {
		return nil, nil, err
	}
	return key, value, nil
}

// ReadMap
func (p *BinaryProtocol) ReadMap(desc *proto.TypeDescriptor, copyString bool, disallowUnknown bool, useFieldName bool) (map[interface{}]interface{}, error) {
	// read first kv pair tag
	fieldNumber, mapWireType, _, mapTagErr := p.ConsumeTag()
	if mapTagErr != nil {
		return nil, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
	}

	if mapWireType != proto.BytesType {
		return nil, meta.NewError(meta.ErrRead, "mapWireType is not BytesType", nil)
	}

	map_kv := make(map[interface{}]interface{})
	keyDesc := desc.Key()
	valueDesc := desc.Elem()

	// read first pair length
	if _, lengthErr := p.ReadLength(); lengthErr != nil {
		return nil, lengthErr
	}
	// read first Pair
	key, value, pairReadErr := p.ReadPair(keyDesc, valueDesc, copyString, disallowUnknown, useFieldName)
	if pairReadErr != nil {
		return nil, pairReadErr
	}
	// add first pair
	map_kv[key] = value

	// check remain pairs
	for p.Read < len(p.Buf) {
		pairNumber, _, tagLen, pairTagErr := p.ConsumeTagWithoutMove()
		if pairTagErr != nil {
			return nil, pairTagErr
		}
		if pairNumber != fieldNumber {
			break
		}

		if _, moveTagErr := p.next(tagLen); moveTagErr != nil {
			return nil, moveTagErr
		}

		if _, pairLenErr := p.ReadLength(); pairLenErr != nil {
			return nil, pairLenErr
		}
		key, value, pairReadErr := p.ReadPair(keyDesc, valueDesc, copyString, disallowUnknown, useFieldName)
		if pairReadErr != nil {
			return nil, pairReadErr
		}
		map_kv[key] = value
	}

	return map_kv, nil
}

// ReadAnyWithDesc read any type by desc and val, the first Tag is parsed outside when use ReadBaseTypeWithDesc
//   - LIST/SET will be converted to []interface{}
//   - MAP will be converted to map[string]interface{} or map[int]interface{} or map[interface{}]interface{}
//   - MESSAGE will be converted to map[proto.FieldNumber]interface{} or map[string]interface{}
func (p *BinaryProtocol) ReadAnyWithDesc(desc *proto.TypeDescriptor, hasMessageLen bool, copyString bool, disallowUnknown bool, useFieldName bool) (interface{}, error) {
	switch {
	case desc.IsList():
		return p.ReadList(desc, copyString, disallowUnknown, useFieldName)
	case desc.IsMap():
		return p.ReadMap(desc, copyString, disallowUnknown, useFieldName)
	default:
		return p.ReadBaseTypeWithDesc(desc, hasMessageLen, copyString, disallowUnknown, useFieldName)
	}
}

// ReadBaseType with desc, not thread safe
func (p *BinaryProtocol) ReadBaseTypeWithDesc(desc *proto.TypeDescriptor, hasMessageLen bool, copyString bool, disallowUnknown bool, useFieldName bool) (interface{}, error) {
	switch desc.Type() {
	case proto.BOOL:
		v, e := p.ReadBool()
		return v, e
	case proto.ENUM:
		v, e := p.ReadEnum()
		return v, e
	case proto.INT32:
		v, e := p.ReadInt32()
		return v, e
	case proto.SINT32:
		v, e := p.ReadSint32()
		return v, e
	case proto.UINT32:
		v, e := p.ReadUint32()
		return v, e
	case proto.FIX32:
		v, e := p.ReadFixed32()
		return v, e
	case proto.SFIX32:
		v, e := p.ReadSfixed32()
		return v, e
	case proto.INT64:
		v, e := p.ReadInt64()
		return v, e
	case proto.SINT64:
		v, e := p.ReadInt64()
		return v, e
	case proto.UINT64:
		v, e := p.ReadUint64()
		return v, e
	case proto.FIX64:
		v, e := p.ReadFixed64()
		return v, e
	case proto.SFIX64:
		v, e := p.ReadSfixed64()
		return v, e
	case proto.FLOAT:
		v, e := p.ReadFloat()
		return v, e
	case proto.DOUBLE:
		v, e := p.ReadDouble()
		return v, e
	case proto.STRING:
		v, e := p.ReadString(copyString)
		return v, e
	case proto.BYTE:
		v, e := p.ReadBytes()
		return v, e
	case proto.MESSAGE:
		messageLength := len(p.Buf) - p.Read
		if hasMessageLen {
			length, messageLengthErr := p.ReadLength()
			if messageLengthErr != nil {
				return nil, messageLengthErr
			}
			if length == 0 {
				return nil, nil
			}
			messageLength = length
		}

		fd := *desc
		msg := fd.Message()
		var retFieldID map[proto.FieldNumber]interface{}
		var retString map[string]interface{}
		if useFieldName {
			retString = make(map[string]interface{}, msg.FieldsCount())
		} else {
			retFieldID = make(map[proto.FieldNumber]interface{}, msg.FieldsCount())
		}
		// read repeat until sumLength equals MessageLength
		start := p.Read
		for p.Read < start+messageLength {
			fieldNumber, wireType, tagLen, fieldTagErr := p.ConsumeTagWithoutMove()
			if fieldTagErr != nil {
				return nil, fieldTagErr
			}
			field := msg.ByNumber(fieldNumber)
			if field == nil {
				if disallowUnknown {
					return nil, errUnknonwField
				}
				// skip unknown field TLV
				if _, moveTagErr := p.next(tagLen); moveTagErr != nil {
					return nil, moveTagErr
				}
				p.Skip(wireType, false)
				continue
			}
			if !field.IsList() && !field.IsMap() {
				p.next(tagLen)
			}
			// sub message has message length, must be true
			hasMsgLen := true
			v, fieldErr := p.ReadAnyWithDesc(field.Type(), hasMsgLen, copyString, disallowUnknown, useFieldName)
			if fieldErr != nil {
				return nil, fieldErr
			}
			if useFieldName {
				retString[string(field.Name())] = v
			} else {
				retFieldID[field.Number()] = v
			}
		}
		if useFieldName {
			return retString, nil
		} else {
			return retFieldID, nil
		}
	default:
		return nil, errInvalidDataType
	}
}
