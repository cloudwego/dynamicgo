package binary

import (
	"encoding/binary"
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
	"google.golang.org/protobuf/reflect/protoreflect"
)

// memory resize factor
const (
	defaultBufferSize = 4096
	growBufferFactor  = 1
	defaultListSize   = 10
)

var (
	errDismatchPrimitive  = meta.NewError(meta.ErrDismatchType, "dismatch primitive types", nil)
	errInvalidDataSize    = meta.NewError(meta.ErrInvalidParam, "invalid data size", nil)
	errInvalidTag         = meta.NewError(meta.ErrInvalidParam, "invalid tag in ReadMessageBegin", nil)
	errInvalidFieldNumber = meta.NewError(meta.ErrInvalidParam, "invalid field number", nil)
	errExceedDepthLimit   = meta.NewError(meta.ErrStackOverflow, "exceed depth limit", nil)
	errInvalidDataType    = meta.NewError(meta.ErrRead, "invalid data type", nil)
	errUnknonwField       = meta.NewError(meta.ErrUnknownField, "unknown field", nil)
	errUnsupportedType    = meta.NewError(meta.ErrUnsupportedType, "unsupported type", nil)
	errNotImplemented     = meta.NewError(meta.ErrNotImplemented, "not implemted type", nil)
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
func (p *BinaryProtocol) AppendTag(num proto.Number, typ proto.WireType) error {
	tag := uint64(num)<<3 | uint64(typ&7)
	if num > proto.MaxValidNumber || num < proto.MinValidNumber {
		return errInvalidFieldNumber
	}
	p.Buf = protowire.BinaryEncoder{}.EncodeUint64(p.Buf, tag)
	return nil
}

// Append Tag With FieldDescriptor
func (p *BinaryProtocol) AppendTagByDesc(desc *proto.FieldDescriptor) error {
	return p.AppendTag((*desc).Number(), proto.Kind2Wire[(*desc).Kind()])
}

// ConsumeTag parses b as a varint-encoded tag, reporting its length.
func (p *BinaryProtocol) ConsumeTag() (proto.Number, proto.WireType, int, error) {
	v, n := protowire.ConsumeVarint((p.Buf)[p.Read:])
	if n < 0 {
		return 0, 0, n, errInvalidTag
	}
	_, err := p.next(n)
	if v>>3 > uint64(math.MaxInt32) {
		return -1, 0, n, errUnknonwField
	}
	num, typ := proto.Number(v>>3), proto.WireType(v&7)
	if num < proto.MinValidNumber {
		return 0, 0, n, errInvalidFieldNumber
	}
	return num, typ, n, err
}

// ConsumeChildTag parses b as a varint-encoded tag, don't move p.Read
func (p *BinaryProtocol) ConsumeTagWithoutMove() (proto.Number, proto.WireType, int, error) {
	v, n := protowire.ConsumeVarint((p.Buf)[p.Read:])
	if n < 0 {
		return 0, 0, n, errInvalidTag
	}
	if v>>3 > uint64(math.MaxInt32) {
		return -1, 0, n, errUnknonwField
	}
	num, typ := proto.Number(v>>3), proto.WireType(v&7)
	if num < proto.MinValidNumber {
		return 0, 0, n, errInvalidFieldNumber
	}
	return num, typ, n, nil
}

// When encoding length-prefixed fields, we speculatively set aside some number of bytes
// for the length, encode the data, and then encode the length (shifting the data if necessary
// to make room).
const speculativeLength = 1

func AppendSpeculativeLength(b []byte) ([]byte, int) {
	pos := len(b)
	b = append(b, "\x00\x00\x00\x00"[:speculativeLength]...)
	return b, pos
}

func FinishSpeculativeLength(b []byte, pos int) []byte {
	mlen := len(b) - pos - speculativeLength
	msiz := protowire.SizeVarint(uint64(mlen))
	if msiz != speculativeLength {
		for i := 0; i < msiz-speculativeLength; i++ {
			b = append(b, 0)
		}
		copy(b[pos+msiz:], b[pos+speculativeLength:])
		b = b[:pos+msiz+mlen]
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
func (p *BinaryProtocol) WriteI32(value int32) error {
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
func (p *BinaryProtocol) WriteFixed32(value int32) error {
	v, err := p.malloc(4)
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint32(v, uint32(value))
	return err
}

// WriteSfixed32
func (p *BinaryProtocol) WriteSfixed32(value int32) error {
	v, err := p.malloc(4)
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint32(v, uint32(value))
	return err
}

// WriteInt64
func (p *BinaryProtocol) WriteI64(value int64) error {
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
	v, err := p.malloc(8)
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint64(v, value)
	return err
}

// WriteSfixed64
func (p *BinaryProtocol) WriteSfixed64(value int64) error {
	v, err := p.malloc(8)
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint64(v, uint64(value))
	return err
}

// WriteFloat
func (p *BinaryProtocol) WriteFloat(value float32) error {
	v, err := p.malloc(4)
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint32(v, math.Float32bits(float32(value)))
	return err
}

// WriteDouble
func (p *BinaryProtocol) WriteDouble(value float64) error {
	v, err := p.malloc(8)
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint64(v, math.Float64bits(value))
	return err
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

/**
 * WriteList
 */
func (p *BinaryProtocol) WriteList(desc *proto.FieldDescriptor, val interface{}) error {
	vs, ok := val.([]interface{})
	if !ok {
		return errDismatchPrimitive
	}
	// packed List bytes format: [tag][length][(L)V][value][value]...
	fd := *desc
	if fd.IsPacked() && len(vs) > 0 {
		p.AppendTag(fd.Number(), proto.BytesType)
		var pos int
		p.Buf, pos = AppendSpeculativeLength(p.Buf)
		for _, v := range vs {
			if err := p.WriteBaseTypeWithDesc(desc, v, true, false, true); err != nil {
				return err
			}
		}
		p.Buf = FinishSpeculativeLength(p.Buf, pos)
		return nil
	}

	// unpacked List bytes format: [T(L)V][T(L)V]...
	kind := fd.Kind()
	for _, v := range vs {
		// share the same field number for Tag
		if err := p.AppendTag(fd.Number(), proto.Kind2Wire[kind]); err != nil {
			return err
		}

		if err := p.WriteBaseTypeWithDesc(desc, v, true, false, true); err != nil {
			return err
		}
	}
	return nil
}

/**
 * WriteMap
 * Map bytes format: [Pairtag][Pairlength][keyTag(L)V][valueTag(L)V] [Pairtag][Pairlength][T(L)V][T(L)V]...
 * Pairtag = MapFieldnumber << 3 | wiretype:BytesType
 */
func (p *BinaryProtocol) WriteMap(desc *proto.FieldDescriptor, val interface{}) error {
	fd := *desc
	MapKey := fd.MapKey()
	MapValue := fd.MapValue()
	vs, ok := val.(map[interface{}]interface{})
	if !ok {
		return errDismatchPrimitive
	}

	for k, v := range vs {
		p.AppendTag(fd.Number(), proto.BytesType)
		var pos int
		p.Buf, pos = AppendSpeculativeLength(p.Buf)
		p.WriteAnyWithDesc(&MapKey, k, true, false, true)
		p.WriteAnyWithDesc(&MapValue, v, true, false, true)
		p.Buf = FinishSpeculativeLength(p.Buf, pos)
	}

	return nil
}

/**
 * Write Message
 */
func (p *BinaryProtocol) WriteMessageSlow(desc *proto.FieldDescriptor, val interface{}, cast bool, disallowUnknown bool, useFieldName bool) error {
	fields := (*desc).Message().Fields()
	if useFieldName {
		for name, v := range val.(map[string]interface{}) {
			f := fields.ByName(proto.FieldName(name))
			if f == nil {
				if disallowUnknown {
					return errUnknonwField
				}
				// unknown field will skip when writing
				continue
			}
			if err := p.WriteAnyWithDesc(&f, v, cast, disallowUnknown, useFieldName); err != nil {
				return err
			}
		}
	} else {
		for fieldNumber, v := range val.(map[proto.FieldNumber]interface{}) {
			f := fields.ByNumber(fieldNumber)
			if f == nil {
				if disallowUnknown {
					return errUnknonwField
				}
				continue
			}
			if err := p.WriteAnyWithDesc(&f, v, cast, disallowUnknown, useFieldName); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteBaseType with desc, not thread safe
func (p *BinaryProtocol) WriteBaseTypeWithDesc(fd *proto.FieldDescriptor, val interface{}, cast bool, disallowUnknown bool, useFieldName bool) error {
	switch (*fd).Kind() {
	case proto.BoolKind:
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
	case proto.EnumKind:
		v, ok := val.(proto.EnumNumber)
		if !ok {
			return meta.NewError(meta.ErrConvert, "convert enum error", nil)
		}
		p.WriteEnum(v)
	case proto.Int32Kind:
		v, ok := val.(int32)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				vv, err := primitive.ToInt64(v)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
				v = int32(vv)
			}
		}
		p.WriteI32(v)
	case proto.Sint32Kind:
		v, ok := val.(int32)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				vv, err := primitive.ToInt64(v)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
				v = int32(vv)
			}
		}
		p.WriteSint32(v)
	case proto.Uint32Kind:
		v, ok := val.(uint32)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				vv, err := primitive.ToInt64(v)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
				v = uint32(vv)
			}
		}
		p.WriteUint32(v)
	case proto.Int64Kind:
		v, ok := val.(int64)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				v, err = primitive.ToInt64(v)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
			}
		}
		p.WriteI64(v)
	case proto.Sint64Kind:
		v, ok := val.(int64)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				v, err = primitive.ToInt64(v)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
			}
		}
		p.WriteSint64(v)
	case proto.Uint64Kind:
		v, ok := val.(uint64)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				vv, err := primitive.ToInt64(v)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
				v = uint64(vv)
			}
		}
		p.WriteUint64(v)
	case proto.Sfixed32Kind:
		v, ok := val.(int32)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				vv, err := primitive.ToInt64(v)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
				v = int32(vv)
			}
		}
		p.WriteSfixed32(v)
	case proto.Fixed32Kind:
		v, ok := val.(int32)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				vv, err := primitive.ToInt64(v)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
				v = int32(vv)
			}
		}
		p.WriteFixed32(v)
	case proto.FloatKind:
		v, ok := val.(float32)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				vfloat64, err := primitive.ToFloat64(v)
				v = float32(vfloat64)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
			}
		}
		p.WriteFloat(v)
	case proto.Sfixed64Kind:
		v, ok := val.(int64)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				v, err = primitive.ToInt64(v)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
			}
		}
		p.WriteSfixed64(v)
	case proto.Fixed64Kind:
		v, ok := val.(int64)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				v, err = primitive.ToInt64(v)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
			}
		}
		p.WriteSfixed64(v)
	case proto.DoubleKind:
		v, ok := val.(float64)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				v, err = primitive.ToFloat64(v)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
			}
		}
		p.WriteDouble(v)
	case proto.StringKind:
		v, ok := val.(string)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				v, err = primitive.ToString(v)
				if err != nil {
					return meta.NewError(meta.ErrConvert, "", err)
				}
			}
		}
		p.WriteString(v)
	case proto.BytesKind:
		v, ok := val.([]byte)
		if !ok {
			return meta.NewError(meta.ErrConvert, "write bytes kind error", nil)
		}
		p.WriteBytes(v)
	case proto.MessageKind:
		var ok bool
		var pos int
		p.Buf, pos = AppendSpeculativeLength(p.Buf)
		if useFieldName {
			val, ok = val.(map[string]interface{})
		} else {
			val, ok = val.(map[proto.FieldNumber]interface{})
		}
		if !ok {
			return errDismatchPrimitive
		}
		if err := p.WriteMessageSlow(fd, val, cast, disallowUnknown, useFieldName); err != nil {
			return err
		}
		p.Buf = FinishSpeculativeLength(p.Buf, pos)
	default:
		return errUnsupportedType
	}
	return nil
}

// WriteAnyWithDesc explain desc and val and write them into buffer
//   - LIST will be converted from []interface{}
//   - MAP will be converted from map[string]interface{} or map[int]interface{}
//   - STRUCT will be converted from map[FieldNumber]interface{} or map[string]interface{}
func (p *BinaryProtocol) WriteAnyWithDesc(desc *proto.FieldDescriptor, val interface{}, cast bool, disallowUnknown bool, useFieldName bool) error {
	fd := *desc
	switch {
	case fd.IsList():
		return p.WriteList(desc, val)
	case fd.IsMap():
		return p.WriteMap(desc, val)
	default:
		if e := p.AppendTag(proto.Number(fd.Number()), proto.Kind2Wire[fd.Kind()]); e != nil {
			return meta.NewError(meta.ErrWrite, "append field tag failed", nil)
		}
		return p.WriteBaseTypeWithDesc(desc, val, cast, disallowUnknown, useFieldName)
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

// ReadInt ...
func (p *BinaryProtocol) ReadInt(t proto.Type) (value int, err error) {
	switch t {
	case proto.INT32:
		n, err := p.ReadI32()
		return int(n), err
	case proto.SINT32:
		n, err := p.ReadSint32()
		return int(n), err
	case proto.SFIX32:
		n, err := p.ReadSfixed32()
		return int(n), err
	case proto.INT64:
		n, err := p.ReadI64()
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
func (p *BinaryProtocol) ReadI32() (int32, error) {
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
func (p *BinaryProtocol) ReadI64() (int64, error) {
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
	value, n := protowire.BinaryDecoder{}.DecodeFixed32((p.Buf)[p.Read:])
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
		return append(emptyBuf[:], value...), errDecodeField
	}
	_, err := p.next(all)
	return append(emptyBuf[:], value...), err
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

// ReadList，List format：
//
//	packed：[tag][length][value value value value....]
//	normal：[tag][(length)][value][tag][(length)][value][tag][(length)][value]....
func (p *BinaryProtocol) ReadList(desc *proto.FieldDescriptor, copyString bool, disallowUnknonw bool, useFieldName bool) ([]interface{}, error) {
	fieldNumber, _, _, listTagErr := p.ConsumeTag()
	if listTagErr != nil {
		return nil, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
	}
	list := make([]interface{}, 0, 1)
	// packed list
	if (*desc).IsPacked() {
		// read length
		length, err := p.ReadLength()
		if err != nil {
			return nil, err
		}
		// read list
		start := p.Read
		for p.Read < start+length {
			v, err := p.ReadBaseTypeWithDesc(desc, copyString, disallowUnknonw, useFieldName)
			if err != nil {
				return nil, err
			}
			list = append(list, v)
		}
	} else {
		// normal Type : [tag][(length)][value][tag][(length)][value][tag][(length)][value]....
		v, err := p.ReadBaseTypeWithDesc(desc, copyString, disallowUnknonw, useFieldName)
		if err != nil {
			return nil, err
		}
		list = append(list, v)

		for p.Read < len(p.Buf) {
			// don't move p.Read and judge whether readList completely
			elementFieldNumber, _, n, err := p.ConsumeTagWithoutMove()
			if err != nil {
				return nil, err
			}
			if elementFieldNumber != fieldNumber {
				break
			}
			p.Read += n
			v, err := p.ReadBaseTypeWithDesc(desc, copyString, disallowUnknonw, useFieldName)
			if err != nil {
				return nil, err
			}
			list = append(list, v)
		}
	}
	return list, nil
}

func (p *BinaryProtocol) ReadPair(keyDesc *proto.FieldDescriptor, valueDesc *proto.FieldDescriptor, copyString bool, disallowUnknonw bool, useFieldName bool) (interface{}, interface{}, error) {
	key, err := p.ReadAnyWithDesc(keyDesc, copyString, disallowUnknonw, useFieldName)
	if err != nil {
		return nil, nil, err
	}

	value, err := p.ReadAnyWithDesc(valueDesc, copyString, disallowUnknonw, useFieldName)
	if err != nil {
		return nil, nil, err
	}
	return key, value, nil
}

// ReadMap
// Map format ：[Pairtag][Pairlength][keyTag(L)V][valueTag(L)V] [Pairtag][Pairlength][keyTag(L)V][valueTag(L)V]....
func (p *BinaryProtocol) ReadMap(desc *proto.FieldDescriptor, copyString bool, disallowUnknonw bool, useFieldName bool) (map[interface{}]interface{}, error) {
	// make a map
	fieldNumber, mapWireType, _, mapTagErr := p.ConsumeTag()
	if mapTagErr != nil {
		return nil, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
	}

	if mapWireType != proto.BytesType {
		return nil, meta.NewError(meta.ErrRead, "mapWireType is not BytesType", nil)
	}

	map_kv := make(map[interface{}]interface{})
	keyDesc := (*desc).MapKey()
	valueDesc := (*desc).MapValue()

	if _, lengthErr := p.ReadLength(); lengthErr != nil {
		return nil, lengthErr
	}
	// read first Pair
	key, value, pairReadErr := p.ReadPair(&keyDesc, &valueDesc, copyString, disallowUnknonw, useFieldName)
	if pairReadErr != nil {
		return nil, pairReadErr
	}

	map_kv[key] = value

	for p.Read < len(p.Buf) {
		pairNumber, _, n, pairTagErr := p.ConsumeTagWithoutMove()
		if pairTagErr != nil {
			return nil, pairTagErr
		}

		if pairNumber != fieldNumber {
			break
		}
		p.Read += n
		if _, pairLenErr := p.ReadLength(); pairLenErr != nil {
			return nil, pairLenErr
		}
		key, value, pairReadErr := p.ReadPair(&keyDesc, &valueDesc, copyString, disallowUnknonw, useFieldName)
		if pairReadErr != nil {
			return nil, pairReadErr
		}
		map_kv[key] = value
	}

	return map_kv, nil
}

// ReadAnyWithDesc read any type by desc and val, the first Tag is parsed outside
//   - LIST/SET will be converted to []interface{}
//   - MAP will be converted from map[string]interface{} or map[int]interface{}
//   - STRUCT will be converted from map[FieldID]interface{}
func (p *BinaryProtocol) ReadAnyWithDesc(desc *proto.FieldDescriptor, copyString bool, disallowUnknonw bool, useFieldName bool) (interface{}, error) {
	switch {
	case (*desc).IsList():
		return p.ReadList(desc, copyString, disallowUnknonw, useFieldName)
	case (*desc).IsMap():
		return p.ReadMap(desc, copyString, disallowUnknonw, useFieldName)
	default:
		_, _, _, err := p.ConsumeTag()
		if err != nil {
			return nil, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
		}
		return p.ReadBaseTypeWithDesc(desc, copyString, disallowUnknonw, useFieldName)
	}
}

// ReadBaseType with desc, not thread safe
func (p *BinaryProtocol) ReadBaseTypeWithDesc(desc *proto.FieldDescriptor, copyString bool, disallowUnknown bool, useFieldName bool) (interface{}, error) {
	switch (*desc).Kind() {
	case protoreflect.BoolKind:
		v, e := p.ReadBool()
		return v, e
	case protoreflect.EnumKind:
		v, e := p.ReadEnum()
		return v, e
	case protoreflect.Int32Kind:
		v, e := p.ReadI32()
		return v, e
	case protoreflect.Sint32Kind:
		v, e := p.ReadSint32()
		return v, e
	case protoreflect.Uint32Kind:
		v, e := p.ReadUint32()
		return v, e
	case protoreflect.Fixed32Kind:
		v, e := p.ReadFixed32()
		return v, e
	case protoreflect.Sfixed32Kind:
		v, e := p.ReadSfixed32()
		return v, e
	case protoreflect.Int64Kind:
		v, e := p.ReadI64()
		return v, e
	case protoreflect.Sint64Kind:
		v, e := p.ReadI64()
		return v, e
	case protoreflect.Uint64Kind:
		v, e := p.ReadUint64()
		return v, e
	case protoreflect.Fixed64Kind:
		v, e := p.ReadFixed64()
		return v, e
	case protoreflect.Sfixed64Kind:
		v, e := p.ReadSfixed64()
		return v, e
	case protoreflect.FloatKind:
		v, e := p.ReadFloat()
		return v, e
	case protoreflect.DoubleKind:
		v, e := p.ReadDouble()
		return v, e
	case protoreflect.StringKind:
		v, e := p.ReadString(false)
		return v, e
	case protoreflect.BytesKind:
		v, e := p.ReadBytes()
		return v, e
	case protoreflect.MessageKind:
		length, messageLengthErr := p.ReadLength()
		if messageLengthErr != nil {
			return nil, messageLengthErr
		}
		if length == 0 {
			return nil, nil
		}

		fd := *desc
		fields := fd.Message().Fields()
		var retFieldID map[proto.FieldNumber]interface{}
		var retString map[string]interface{}
		if useFieldName {
			retString = make(map[string]interface{}, fields.Len())
		} else {
			retFieldID = make(map[proto.FieldNumber]interface{}, fields.Len())
		}
		// read repeat until sumLength equals MessageLength
		start := p.Read
		for p.Read < start+length {
			fieldNumber, wireType, tagLen, fieldTagErr := p.ConsumeTagWithoutMove()
			if fieldTagErr != nil {
				return nil, fieldTagErr
			}
			field := fields.ByNumber(fieldNumber)
			if field == nil {
				if disallowUnknown {
					return nil, errUnknonwField
				}
				p.Read += tagLen // move Read cross tag
				// skip unknown field value
				p.Skip(wireType, false)
				continue
			}
			v, fieldErr := p.ReadAnyWithDesc(&field, copyString, disallowUnknown, useFieldName)
			if fieldErr != nil {
				return nil, fieldErr
			}
			if useFieldName {
				retString[field.TextName()] = v
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
