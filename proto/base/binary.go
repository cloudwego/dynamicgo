package base

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"reflect"
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
)

var byteType = rt.UnpackType(reflect.TypeOf(byte(0)))

// map from proto.ProtoKind to proto.WireType
var wireTypes = map[proto.ProtoKind]proto.WireType{
	proto.BoolKind:     proto.VarintType,
	proto.EnumKind:     proto.VarintType,
	proto.Int32Kind:    proto.VarintType,
	proto.Sint32Kind:   proto.VarintType,
	proto.Uint32Kind:   proto.VarintType,
	proto.Int64Kind:    proto.VarintType,
	proto.Sint64Kind:   proto.VarintType,
	proto.Uint64Kind:   proto.VarintType,
	proto.Sfixed32Kind: proto.Fixed32Type,
	proto.Fixed32Kind:  proto.Fixed32Type,
	proto.FloatKind:    proto.Fixed32Type,
	proto.Sfixed64Kind: proto.Fixed64Type,
	proto.Fixed64Kind:  proto.Fixed64Type,
	proto.DoubleKind:   proto.Fixed64Type,
	proto.StringKind:   proto.BytesType,
	proto.BytesKind:    proto.BytesType,
	proto.MessageKind:  proto.BytesType,
	proto.GroupKind:    proto.StartGroupType,
}

var (
	errDismatchPrimitive = meta.NewError(meta.ErrDismatchType, "dismatch primitive types", nil)
	errInvalidDataSize   = meta.NewError(meta.ErrInvalidParam, "invalid data size", nil)
	errInvalidTag        = meta.NewError(meta.ErrInvalidParam, "invalid tag in ReadMessageBegin", nil)
	errExceedDepthLimit  = meta.NewError(meta.ErrStackOverflow, "exceed depth limit", nil)
	errInvalidDataType   = meta.NewError(meta.ErrRead, "invalid data type", nil)
	errUnknonwField      = meta.NewError(meta.ErrUnknownField, "unknown field", nil)
	errUnsupportedType   = meta.NewError(meta.ErrUnsupportedType, "unsupported type", nil)
	errNotImplemented    = meta.NewError(meta.ErrNotImplemented, "not implemted type", nil)
	errCodeFieldNumber   = meta.NewError(meta.ErrConvert, "invalid field number", nil)
	errDecodeField       = meta.NewError(meta.ErrRead, "cannot parse invalid wire-format data", nil)
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

// Append Tag
func (p *BinaryProtocol) AppendTag(num proto.Number, typ proto.WireType) error {
	tag := uint64(num)<<3 | uint64(typ&7)
	p.Buf = protowire.BinaryEncoder{}.EncodeInt64(p.Buf, int64(tag))
	return nil
}

// ConsumeTag parses b as a varint-encoded tag, reporting its length.
func (p *BinaryProtocol) ConsumeTag() (proto.Number, proto.WireType, int, error) {
	v, n := protowire.ConsumeVarint((p.Buf)[p.Read:])
	_, err := p.next(n)
	if n < 0 {
		return 0, 0, n, errInvalidTag
	}
	if v>>3 > uint64(math.MaxInt32) {
		return -1, 0, n, errUnknonwField
	}
	num, typ := proto.Number(v>>3), proto.WireType(v&7)
	if num < proto.MinValidNumber {
		return 0, 0, n, errCodeFieldNumber
	}
	return num, typ, n, err
}

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
	binary.BigEndian.PutUint32(v, uint32(value))
	return err
}

// WriteSfixed32
func (p *BinaryProtocol) WriteSfixed32(value int32) error {
	v, err := p.malloc(4)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint32(v, uint32(value))
	return err
}

// WriteInt64
func (p *BinaryProtocol) WriteI64(value int64) error {
	protowire.BinaryEncoder{}.EncodeInt64(p.Buf, value)
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
	binary.BigEndian.PutUint64(v, value)
	return err
}

// WriteSfixed64
func (p *BinaryProtocol) WriteSfixed64(value int64) error {
	v, err := p.malloc(8)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint64(v, uint64(value))
	return err
}

// WriteFloat
func (p *BinaryProtocol) WriteFloat(value float64) error {
	v, err := p.malloc(4)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint32(v, math.Float32bits(float32(value)))
	return err
}

// WriteDouble
func (p *BinaryProtocol) WriteDouble(value float64) error {
	v, err := p.malloc(8)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint64(v, math.Float64bits(value))
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
	protowire.BinaryEncoder{}.EncodeInt64(p.Buf, int64(value))
	return nil
}

/**
 * WriteList
 */
func (p *BinaryProtocol) WriteList() error {
	return nil
}

/**
 * WriteMap
 */
func (p *BinaryProtocol) WriteMap() error {
	return nil
}

/**
 * Write Message
 */
func (p *BinaryProtocol) WriteMessageSlow(desc proto.FieldDescriptor, vs map[string]interface{}, cast bool, disallowUnknown bool, useFieldName bool) error {
	for id, v := range vs {
		f := desc.Message().Fields().ByName(protoreflect.Name(id))
		if f == nil {
			if disallowUnknown {
				return errUnknonwField
			}
			continue
		}
		if e := p.WriteAnyWithDesc(f, v, cast, disallowUnknown, useFieldName); e != nil {
			return e
		}
	}
	return nil
}

// When encoding length-prefixed fields, we speculatively set aside some number of bytes
// for the length, encode the data, and then encode the length (shifting the data if necessary
// to make room).
const speculativeLength = 1

func appendSpeculativeLength(b []byte) int {
	pos := len(b)
	b = append(b, "\x00\x00\x00\x00"[:speculativeLength]...)
	return pos
}

func finishSpeculativeLength(b []byte, pos int) []byte {
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

// WriteAnyWithDesc explain desc and val and write them into buffer
//   - LIST/SET will be converted from []interface{}
//   - MAP will be converted from map[string]interface{} or map[int]interface{}
//   - STRUCT will be converted from map[FieldID]interface{}
func (p *BinaryProtocol) WriteAnyWithDesc(desc proto.FieldDescriptor, val interface{}, cast bool, disallowUnknown bool, useFieldName bool) error {
	switch {
	case desc.IsList():
		return p.WriteList()
	case desc.IsMap():
		return p.WriteMap()
	default:
		if e := p.AppendTag(proto.Number(desc.Number()), wireTypes[desc.Kind()]); e != nil {
			return meta.NewError(meta.ErrWrite, "AppendTag failed", nil)
		}
		return p.WriteBaseTypeWithDesc(desc, val, cast, disallowUnknown, useFieldName)
	}
}

// WriteBaseType with desc, not thread safe
func (p *BinaryProtocol) WriteBaseTypeWithDesc(fd proto.FieldDescriptor, val interface{}, cast bool, disallowUnknown bool, useFieldName bool) error {
	switch fd.Kind() {
	case protoreflect.Kind(proto.BoolKind):
		v, ok := val.(bool)
		if !ok {
			var err error
			v, err = primitive.ToBool(val)
			if err != nil {
				return err
			}
		}
		p.WriteBool(v)
	case protoreflect.Kind(proto.EnumKind):
		v, ok := val.(proto.EnumNumber)
		if !ok {
			return meta.NewError(meta.ErrConvert, "WriteEnum error", nil)
		}
		p.WriteEnum(v)
	case protoreflect.Kind(proto.Int32Kind):
		v, ok := val.(int32)
		if !ok {
			var err error
			vv, err := primitive.ToInt64(v)
			if err != nil {
				return err
			}
			v = int32(vv)
		}
		p.WriteI32(v)
	case protoreflect.Kind(proto.Sint32Kind):
		v, ok := val.(int32)
		if !ok {
			var err error
			vv, err := primitive.ToInt64(v)
			if err != nil {
				return err
			}
			v = int32(vv)
		}
		p.WriteSint32(v)
	case protoreflect.Kind(proto.Uint32Kind):
		v, ok := val.(uint32)
		if !ok {
			var err error
			vv, err := primitive.ToInt64(v)
			if err != nil {
				return err
			}
			v = uint32(vv)
		}
		p.WriteUint32(v)
	case protoreflect.Kind(proto.Int64Kind):
		v, ok := val.(int64)
		if !ok {
			var err error
			v, err = primitive.ToInt64(v)
			if err != nil {
				return err
			}
		}
		p.WriteI64(v)
	case protoreflect.Kind(proto.Sint64Kind):
		v, ok := val.(int64)
		if !ok {
			var err error
			v, err = primitive.ToInt64(v)
			if err != nil {
				return err
			}
		}
		p.WriteSint64(v)
	case protoreflect.Kind(proto.Uint64Kind):
		v, ok := val.(uint64)
		if !ok {
			var err error
			vv, err := primitive.ToInt64(v)
			if err != nil {
				return err
			}
			v = uint64(vv)
		}
		p.WriteUint64(v)
	case protoreflect.Kind(proto.Sfixed32Kind):
		v, ok := val.(int32)
		if !ok {
			var err error
			vv, err := primitive.ToInt64(v)
			if err != nil {
				return err
			}
			v = int32(vv)
		}
		p.WriteSfixed32(v)
	case protoreflect.Kind(proto.Fixed32Kind):
		v, ok := val.(int32)
		if !ok {
			var err error
			vv, err := primitive.ToInt64(v)
			if err != nil {
				return err
			}
			v = int32(vv)
		}
		p.WriteFixed32(v)
	case protoreflect.Kind(proto.FloatKind):
		v, ok := val.(float64)
		if !ok {
			var err error
			v, err = primitive.ToFloat64(v)
			if err != nil {
				return err
			}
		}
		p.WriteFloat(v)
	case protoreflect.Kind(proto.Sfixed64Kind):
		v, ok := val.(int64)
		if !ok {
			var err error
			v, err = primitive.ToInt64(v)
			if err != nil {
				return err
			}
		}
		p.WriteSfixed64(v)
	case protoreflect.Kind(proto.Fixed64Kind):
		v, ok := val.(int64)
		if !ok {
			var err error
			v, err = primitive.ToInt64(v)
			if err != nil {
				return err
			}
		}
		p.WriteSfixed64(v)
	case protoreflect.Kind(proto.DoubleKind):
		v, ok := val.(float64)
		if !ok {
			var err error
			v, err = primitive.ToFloat64(v)
			if err != nil {
				return err
			}
		}
		p.WriteDouble(v)
	case protoreflect.Kind(proto.StringKind):
		v, ok := val.(string)
		if !ok {
			return meta.NewError(meta.ErrConvert, string(fd.FullName()), nil)
		}
		p.WriteString(v)
	case protoreflect.Kind(proto.BytesKind):
		v, ok := val.([]byte)
		if !ok {
			return errors.New("WriteBytesType error")
		}
		p.WriteBytes(v)
	case protoreflect.Kind(proto.MessageKind):
		var pos int
		var err error
		vs, ok := val.(map[string]interface{})
		if !ok {
			return errDismatchPrimitive
		}
		pos = appendSpeculativeLength(p.Buf)
		err = p.WriteMessageSlow(fd, vs, cast, disallowUnknown, useFieldName)
		if err != nil {
			return err
		}
		p.Buf = finishSpeculativeLength(p.Buf, pos)
	// case protoreflect.Kind(proto.GroupKind):
	// 	var err error
	// 	b, err = o.marshalMessage(b, v.Message())
	// 	if err != nil {
	// 		return b, err
	// 	}
	// 	b = protowire.AppendVarint(b, protowire.EncodeTag(fd.Number(), protowire.EndGroupType))
	default:
		return errUnsupportedType
	}
	return nil
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

// ReadBytes
func (p *BinaryProtocol) ReadBytes() ([]byte, error) {
	value, n := protowire.BinaryDecoder{}.DecodeBytes((p.Buf)[p.Read:])
	if n < 0 {
		return append(emptyBuf[:], value...), errDecodeField
	}
	_, err := p.next(n)
	return append(emptyBuf[:], value...), err
}

// ReadLength
func (p *BinaryProtocol) ReadLength() (int, error) {
	value, n := protowire.BinaryDecoder{}.DecodeUint64((p.Buf)[p.Read:])
	if n < 0 {
		return 0, errDecodeField
	}
	_, err := p.next(n)
	return int(value), err
}

// ReadString
func (p *BinaryProtocol) ReadString() (string, error) {
	value, n := protowire.BinaryDecoder{}.DecodeBytes((p.Buf)[p.Read:])
	if n < 0 {
		return "", errDecodeField
	}
	_, err := p.next(n)
	return string(value), err
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
