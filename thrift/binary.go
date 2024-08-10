/**
 * Copyright 2023 CloudWeGo Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package thrift

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/json"
	"github.com/cloudwego/dynamicgo/internal/primitive"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/internal/util"
	"github.com/cloudwego/dynamicgo/meta"
)

// memory resize factor
const (
	// new = old + old >> growSliceFactor
	growBufferFactor = 1

	defaultBufferSize = 4096

	msgHeaderFixedLen = 4 + 4 + 4 + 3
	msgFooterFixedLen = 1
)

// TMessageType is the type of message
type TMessageType int32

const (
	INVALID_TMESSAGE_TYPE TMessageType = 0
	CALL                  TMessageType = 1
	REPLY                 TMessageType = 2
	EXCEPTION             TMessageType = 3
	ONEWAY                TMessageType = 4
)

var (
	errDismatchPrimitive = meta.NewError(meta.ErrDismatchType, "dismatch primitive types", nil)
	errInvalidDataSize   = meta.NewError(meta.ErrInvalidParam, "invalid data size", nil)
	errInvalidVersion    = meta.NewError(meta.ErrInvalidParam, "invalid version in ReadMessageBegin", nil)
	errExceedDepthLimit  = meta.NewError(meta.ErrStackOverflow, "exceed depth limit", nil)
	errInvalidDataType   = meta.NewError(meta.ErrRead, "invalid data type", nil)
	errUnknonwField      = meta.NewError(meta.ErrUnknownField, "unknown field", nil)
	errUnsupportedType   = meta.NewError(meta.ErrUnsupportedType, "unsupported type", nil)
	errNotImplemented    = meta.NewError(meta.ErrNotImplemented, "not implemted type", nil)
)

// must be strict read & strict write
var (
	bpPool = sync.Pool{
		New: func() interface{} {
			return &BinaryProtocol{
				Buf: make([]byte, 0, defaultBufferSize),
			}
		},
	}
)

// NewBinaryProtocol get a new binary protocol from sync.Pool.
func NewBinaryProtocol(buf []byte) *BinaryProtocol {
	bp := bpPool.Get().(*BinaryProtocol)
	bp.Buf = buf
	return bp
}

// NewBinaryProtocolBuffer gets a new binary protocol from sync.Pool
// and reuse the buffer in pool
func NewBinaryProtocolBuffer() *BinaryProtocol {
	bp := bpPool.Get().(*BinaryProtocol)
	return bp
}

// FreeBinaryProtocol resets the buffer and puts the binary protocol back to sync.Pool
func FreeBinaryProtocolBuffer(bp *BinaryProtocol) {
	bp.Reset()
	bpPool.Put(bp)
}

// Recycle put the protocol back to sync.Pool
func (p *BinaryProtocol) Recycle() {
	p.Reset()
	bpPool.Put(p)
}

// BinaryProtocol implements the BinaryProtocol
// see https://github.com/apache/thrift/blob/master/doc/specs/thrift-binary-protocol.md
type BinaryProtocol struct {
	Buf  []byte
	Read int
}

// Reset resets the buffer and read position
func (p *BinaryProtocol) Reset() {
	p.Read = 0
	p.Buf = p.Buf[:0]
}

// RawBuf returns the raw buffer of the protocol
func (p BinaryProtocol) RawBuf() []byte {
	return p.Buf
}

// Left returns the left bytes to read
func (p BinaryProtocol) Left() int {
	return len(p.Buf) - p.Read
}

/**
 * Message related methods
 */

// GetBinaryMessageHeaderAndFooter writes the message parameters into header and footer
func GetBinaryMessageHeaderAndFooter(methodName string, msgTyp TMessageType, structID FieldID, seqID int32) (header []byte, footer []byte, err error) {
	var writer = BinaryProtocol{}

	// write header
	header = make([]byte, 0, msgHeaderFixedLen+len(methodName))
	writer.Buf = header
	err = writer.WriteMessageBegin(methodName, msgTyp, seqID)
	if err != nil {
		return
	}
	err = writer.WriteStructBegin("")
	if err != nil {
		return
	}
	err = writer.WriteFieldBegin("", STRUCT, structID)
	if err != nil {
		return
	}
	header = writer.Buf

	// write footer
	footer = make([]byte, 0, msgFooterFixedLen)
	writer.Buf = footer
	err = writer.WriteFieldEnd()
	if err != nil {
		return
	}
	err = writer.WriteStructEnd()
	if err != nil {
		return
	}
	err = writer.WriteMessageEnd()
	if err != nil {
		return
	}
	footer = writer.Buf

	return
}

// WrapBinaryMessage wraps the message with header and footer and body
func WrapBinaryBody(body []byte, methodName string, msgTyp TMessageType, structID FieldID, seqID int32) ([]byte, error) {
	// write header
	buf := make([]byte, 0, msgHeaderFixedLen+len(methodName)+len(body)+msgFooterFixedLen)
	writer := BinaryProtocol{Buf: buf}
	writer.WriteMessageBegin(methodName, msgTyp, seqID)
	writer.WriteStructBegin("")
	writer.WriteFieldBegin("", STRUCT, structID)
	writer.Buf = append(writer.Buf, body...)
	writer.WriteFieldEnd()
	writer.WriteStructEnd()
	writer.WriteMessageEnd()
	return writer.Buf, nil
}

// UnwrapBinaryMessage unwraps the message parameters from the buf
func UnwrapBinaryMessage(buf []byte) (name string, callType TMessageType, seqID int32, structID FieldID, body []byte, err error) {
	var reader = BinaryProtocol{
		Buf: buf,
	}
	return reader.UnwrapBody()
}

// UnwrapBody unwraps the  message parameters from its buf
func (p BinaryProtocol) UnwrapBody() (string, TMessageType, int32, FieldID, []byte, error) {
	name, rTyp, seqID, err := p.ReadMessageBegin(false)
	if err != nil {
		return name, rTyp, seqID, 0, nil, err
	}
	// read the success struct
	_, typeId, structID, err := p.ReadFieldBegin()
	if err != nil {
		return name, rTyp, seqID, structID, nil, err
	}
	//empty response
	if typeId == STOP {
		return name, rTyp, seqID, structID, []byte{}, nil
	}
	// there's alway a struct stop by success struct
	if p.Read > len(p.Buf)-1 {
		return name, rTyp, seqID, structID, nil, io.EOF
	}
	return name, rTyp, seqID, structID, p.Buf[p.Read : len(p.Buf)-1], err
}

/**
 * Writing Methods
 */

// WriteMessageBegin ...
func (p *BinaryProtocol) WriteMessageBegin(name string, typeID TMessageType, seqID int32) error {
	version := uint32(VERSION_1) | uint32(typeID)
	e := p.WriteI32(int32(version))
	if e != nil {
		return e
	}
	e = p.WriteString(name)
	if e != nil {
		return e
	}
	e = p.WriteI32(seqID)
	return e
}

// WriteMessageEnd ...
func (p *BinaryProtocol) WriteMessageEnd() error {
	return nil
}

// WriteStructBegin ...
func (p *BinaryProtocol) WriteStructBegin(name string) error {
	return nil
}

// WriteStructEnd ...
func (p *BinaryProtocol) WriteStructEnd() error {
	return p.WriteFieldStop()
}

// WriteFieldBegin ...
func (p *BinaryProtocol) WriteFieldBegin(name string, typeID Type, id FieldID) error {
	e := p.WriteByte(byte(typeID))
	if e != nil {
		return e
	}
	e = p.WriteI16(int16(id))
	return e
}

// WriteFieldEnd ...
func (p *BinaryProtocol) WriteFieldEnd() error {
	return nil
}

// WriteFieldStop ...
func (p *BinaryProtocol) WriteFieldStop() error {
	e := p.WriteByte(byte(STOP))
	return e
}

// WriteMapBegin ...
func (p *BinaryProtocol) WriteMapBegin(keyType, valueType Type, size int) error {
	e := p.WriteByte(byte(keyType))
	if e != nil {
		return e
	}
	e = p.WriteByte(byte(valueType))
	if e != nil {
		return e
	}
	e = p.WriteI32(int32(size))
	return e
}

// WriteMapBeginWithSizePos writes the map begin, and return the buffer position of the size data
func (p *BinaryProtocol) WriteMapBeginWithSizePos(keyType, valueType Type, size int) (int, error) {
	e := p.WriteByte(byte(keyType))
	if e != nil {
		return 0, e
	}
	e = p.WriteByte(byte(valueType))
	if e != nil {
		return 0, e
	}
	re := len(p.Buf)
	e = p.WriteI32(int32(size))
	return re, e
}

// WriteMapEnd ...
func (p *BinaryProtocol) WriteMapEnd() error {
	return nil
}

// WriteListBegin ...
func (p *BinaryProtocol) WriteListBegin(elemType Type, size int) error {
	e := p.WriteByte(byte(elemType))
	if e != nil {
		return e
	}
	e = p.WriteI32(int32(size))
	return e
}

// WriteListBeginWithSizePos writes the list begin, and return the buffer position of the size data
func (p *BinaryProtocol) WriteListBeginWithSizePos(elemType Type, size int) (int, error) {
	e := p.WriteByte(byte(elemType))
	if e != nil {
		return 0, e
	}
	re := len(p.Buf)
	e = p.WriteI32(int32(size))
	return re, e
}

// WriteListEnd ...
func (p *BinaryProtocol) WriteListEnd() error {
	return nil
}

// WriteSetBegin ...
func (p *BinaryProtocol) WriteSetBegin(elemType Type, size int) error {
	e := p.WriteByte(byte(elemType))
	if e != nil {
		return e
	}
	e = p.WriteI32(int32(size))
	return e
}

// WriteSetEnd ...
func (p *BinaryProtocol) WriteSetEnd() error {
	return nil
}

// WriteBool ...
func (p *BinaryProtocol) WriteBool(value bool) error {
	if value {
		return p.WriteByte(1)
	}
	return p.WriteByte(0)
}

// WriteByte ...
func (p *BinaryProtocol) WriteByte(value byte) error {
	p.Buf = append(p.Buf, byte(value))
	return nil
}

// WriteI16 ...
func (p *BinaryProtocol) WriteI16(value int16) error {
	v, err := p.malloc(2)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint16(v, uint16(value))
	return err
}

// ModifyI16 write int32 into the buffer at the given position
func (p *BinaryProtocol) ModifyI32(pos int, value int32) error {
	old := len(p.Buf)
	if old < pos+4 {
		return fmt.Errorf("not enough space to modify i32")
	}
	p.Buf = p.Buf[:pos]
	p.WriteI32(value)
	p.Buf = p.Buf[:old]
	return nil
}

// WriteI32 ...
func (p *BinaryProtocol) WriteI32(value int32) error {
	v, err := p.malloc(4)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint32(v, uint32(value))
	return err
}

// WriteI64 ...
func (p *BinaryProtocol) WriteI64(value int64) error {
	v, err := p.malloc(8)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint64(v, uint64(value))
	return err
}

// WriteInt ...
func (p *BinaryProtocol) WriteInt(t Type, value int) error {
	switch t {
	case I08:
		return p.WriteByte(byte(value))
	case I16:
		return p.WriteI16(int16(value))
	case I32:
		return p.WriteI32(int32(value))
	case I64:
		return p.WriteI64(int64(value))
	default:
		return errInvalidDataType
	}
}

// WriteDouble ...
func (p *BinaryProtocol) WriteDouble(value float64) error {
	return p.WriteI64(int64(math.Float64bits(value)))
}

// WriteString ...
func (p *BinaryProtocol) WriteString(value string) error {
	len := len(value)
	e := p.WriteI32(int32(len))
	if e != nil {
		return e
	}
	p.Buf = append(p.Buf, value...)
	return nil
}

// WriteBinary ...
func (p *BinaryProtocol) WriteBinary(value []byte) error {
	e := p.WriteI32(int32(len(value)))
	if e != nil {
		return e
	}
	p.Buf = append(p.Buf, value...)
	return nil
}

// malloc ...
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

// WriteDefaultOrEmpty write default value if any, otherwise write zero value
func (p *BinaryProtocol) WriteDefaultOrEmpty(field *FieldDescriptor) error {
	if dv := field.DefaultValue(); dv != nil {
		p.Buf = append(p.Buf, dv.ThriftBinary()...)
		return nil
	}
	return p.WriteEmpty(field.Type())
}

// WriteEmpty write zero value
func (p *BinaryProtocol) WriteEmpty(desc *TypeDescriptor) error {
	switch desc.Type() {
	case BOOL:
		return p.WriteBool(false)
	case BYTE:
		return p.WriteByte(0)
	case I16:
		return p.WriteI16(0)
	case I32:
		return p.WriteI32(0)
	case I64:
		return p.WriteI64(0)
	case DOUBLE:
		return p.WriteDouble(0)
	case STRING:
		return p.WriteString("")
	case LIST, SET:
		if err := p.WriteListBegin(desc.Elem().Type(), 0); err != nil {
			return err
		}
		return p.WriteListEnd()
	case MAP:
		if err := p.WriteMapBegin(desc.Key().Type(), desc.Elem().Type(), 0); err != nil {
			return err
		}
		return p.WriteMapEnd()
	case STRUCT:
		// NOTICE: to avoid self-cycled type dead loop here, just write empty struct
		return p.WriteStructEnd()
	default:
		return errors.New("invalid type")
	}
}

/**
 * Reading methods
 */

// ReadMessageBegin ...
func (p *BinaryProtocol) ReadMessageBegin(copyString bool) (name string, typeID TMessageType, seqID int32, err error) {
	size, e := p.ReadI32()
	if e != nil {
		return "", typeID, 0, errInvalidVersion
	}
	if size > 0 {
		return name, typeID, seqID, errInvalidVersion
	}
	typeID = TMessageType(size & 0x0ff)
	version := int64(int64(size) & VERSION_MASK)
	if version != VERSION_1 {
		return name, typeID, seqID, errInvalidVersion
	}
	name, e = p.ReadString(copyString)
	if e != nil {
		return name, typeID, seqID, errInvalidVersion
	}
	seqID, e = p.ReadI32()
	if e != nil {
		return name, typeID, seqID, errInvalidVersion
	}
	return name, typeID, seqID, nil
}

// ReadMessageEnd ...
func (p *BinaryProtocol) ReadMessageEnd() error {
	return nil
}

// ReadStructBegin ...
func (p *BinaryProtocol) ReadStructBegin() (name string, err error) {
	return
}

// ReadStructEnd ...
func (p *BinaryProtocol) ReadStructEnd() error {
	return nil
}

// ReadFieldBegin ...
func (p *BinaryProtocol) ReadFieldBegin() (name string, typeID Type, id FieldID, err error) {
	t, err := p.ReadByte()
	typeID = Type(t)
	if err != nil {
		return name, typeID, id, err
	}
	if !typeID.Valid() {
		return "", 0, 0, errInvalidDataType
	}
	if t != byte(STOP) {
		var x int16
		x, err = p.ReadI16()
		id = FieldID(x)
	}
	return name, typeID, id, err
}

// ReadFieldEnd ...
func (p *BinaryProtocol) ReadFieldEnd() error {
	return nil
}

// ReadMapBegin ...
func (p *BinaryProtocol) ReadMapBegin() (kType, vType Type, size int, err error) {
	k, e := p.ReadByte()
	if e != nil {
		err = e
		return
	}
	kType = Type(k)
	if !kType.Valid() {
		return 0, 0, 0, errInvalidDataType
	}

	v, e := p.ReadByte()
	if e != nil {
		err = e
		return
	}
	vType = Type(v)
	if !vType.Valid() {
		return 0, 0, 0, errInvalidDataType
	}

	size32, e := p.ReadI32()
	if e != nil {
		err = e
		return
	}
	if size32 < 0 {
		err = errInvalidDataSize
		return
	}
	size = int(size32)
	return kType, vType, size, nil
}

// ReadMapEnd ...
func (p *BinaryProtocol) ReadMapEnd() error {
	return nil
}

// ReadListBegin ...
func (p *BinaryProtocol) ReadListBegin() (elemType Type, size int, err error) {
	b, e := p.ReadByte()
	if e != nil {
		err = e
		return
	}

	elemType = Type(b)
	if !elemType.Valid() {
		return 0, 0, errInvalidDataType
	}

	size32, e := p.ReadI32()
	if e != nil {
		err = e
		return
	}
	if size32 < 0 {
		err = errInvalidDataSize
		return
	}
	size = int(size32)

	return
}

// ReadListEnd ...
func (p *BinaryProtocol) ReadListEnd() error {
	return nil
}

// ReadSetBegin ...
func (p *BinaryProtocol) ReadSetBegin() (elemType Type, size int, err error) {
	b, e := p.ReadByte()
	if e != nil {
		err = e
		return
	}

	elemType = Type(b)
	if !elemType.Valid() {
		return 0, 0, errInvalidDataType
	}

	size32, e := p.ReadI32()
	if e != nil {
		err = e
		return
	}
	if size32 < 0 {
		err = errInvalidDataSize
		return
	}
	size = int(size32)
	return elemType, size, nil
}

// ReadSetEnd ...
func (p *BinaryProtocol) ReadSetEnd() error {
	return nil
}

// ReadBool ...
func (p *BinaryProtocol) ReadBool() (bool, error) {
	b, e := p.ReadByte()
	v := true
	if b != 1 {
		v = false
	}
	return v, e
}

// ReadByte ...
func (p *BinaryProtocol) ReadByte() (value byte, err error) {
	buf, err := p.next(1)
	if err != nil {
		return value, err
	}
	return byte(buf[0]), err
}

// ReadI16 ...
func (p *BinaryProtocol) ReadI16() (value int16, err error) {
	buf, err := p.next(2)
	if err != nil {
		return value, err
	}
	value = int16(binary.BigEndian.Uint16(buf))
	return value, err
}

// ReadI32 ...
func (p *BinaryProtocol) ReadI32() (value int32, err error) {
	buf, err := p.next(4)
	if err != nil {
		return value, err
	}
	value = int32(binary.BigEndian.Uint32(buf))
	return value, err
}

// ReadI64 ...
func (p *BinaryProtocol) ReadI64() (value int64, err error) {
	buf, err := p.next(8)
	if err != nil {
		return value, err
	}
	value = int64(binary.BigEndian.Uint64(buf))
	return value, err
}

// ReadInt ...
func (p *BinaryProtocol) ReadInt(t Type) (value int, err error) {
	switch t {
	case I08:
		n, err := p.ReadByte()
		return int(n), err
	case I16:
		n, err := p.ReadI16()
		return int(n), err
	case I32:
		n, err := p.ReadI32()
		return int(n), err
	case I64:
		n, err := p.ReadI64()
		return int(n), err
	default:
		return 0, errInvalidDataType
	}
}

// ReadDouble ...
func (p *BinaryProtocol) ReadDouble() (value float64, err error) {
	buf, err := p.next(8)
	if err != nil {
		return value, err
	}
	value = math.Float64frombits(binary.BigEndian.Uint64(buf))
	return value, err
}

// ReadString ...
func (p *BinaryProtocol) ReadString(copy bool) (value string, err error) {
	size, e := p.ReadI32()
	if e != nil {
		return "", e
	}
	if size < 0 || int(size) > len(p.Buf)-p.Read {
		err = errInvalidDataSize
		return
	}

	if copy {
		value = string(*(*[]byte)(unsafe.Pointer(&rt.GoSlice{
			Ptr: rt.IndexPtr(*(*unsafe.Pointer)(unsafe.Pointer(&p.Buf)), byteTypeSize, p.Read),
			Len: int(size),
			Cap: int(size),
		})))
	} else {
		v := (*rt.GoString)(unsafe.Pointer(&value))
		v.Ptr = rt.IndexPtr(*(*unsafe.Pointer)(unsafe.Pointer(&p.Buf)), byteTypeSize, p.Read)
		v.Len = int(size)
	}

	p.Read += int(size)
	return
}

// ReadBinary ...
func (p *BinaryProtocol) ReadBinary(copyBytes bool) (value []byte, err error) {
	size, e := p.ReadI32()
	if e != nil {
		return nil, e
	}
	if size < 0 || int(size) > len(p.Buf)-p.Read {
		return nil, errInvalidDataSize
	}

	if copyBytes {
		value = make([]byte, int(size))
		copy(value, *(*[]byte)(unsafe.Pointer(&rt.GoSlice{
			Ptr: rt.IndexPtr(*(*unsafe.Pointer)(unsafe.Pointer(&p.Buf)), byteTypeSize, p.Read),
			Len: int(size),
			Cap: int(size),
		})))
	} else {
		v := (*rt.GoString)(unsafe.Pointer(&value))
		v.Ptr = rt.IndexPtr(*(*unsafe.Pointer)(unsafe.Pointer(&p.Buf)), byteTypeSize, p.Read)
		v.Len = int(size)
	}

	p.Read += int(size)
	return
}

// ReadStringWithDesc explains thrift data with desc and converts to simple string
func (p *BinaryProtocol) ReadStringWithDesc(desc *TypeDescriptor, buf *[]byte, byteAsUint8 bool, disallowUnknown bool, base64Binary bool) error {
	return p.EncodeText(desc, buf, byteAsUint8, disallowUnknown, base64Binary, true, false)
}

// EncodeText reads thrift data with descriptor, and converts it to a specail text-protocol string:
// This protocol is similar to JSON, excepts its key (or field id) IS NOT QUOTED unless it is a string type:
//   - LIST/SET's all elements will be joined with ',',
//     and if asJson is true the entiry value will be wrapped by '[' (start) and ']' (end).
//   - MAP's each pair of key and value will be binded with ':', all elements will be joined with ',',
//     and if asJson is true the entiry value will be wrapped by '{' (start) and '}' (end).
//   - STRUCT's each pair of field (name or id) and value will be binded with ':', all elements will be joined with ',',
//     and if asJson is true the entiry value will be wrapped by '{' (start) and '}' (end).
//   - STRING (including key) will be wrapped by '"' if asJson is true.
func (p *BinaryProtocol) EncodeText(desc *TypeDescriptor, buf *[]byte, byteAsUint8 bool, disallowUnknown bool, base64Binary bool, useFieldName bool, asJson bool) error {
	switch desc.Type() {
	case BOOL:
		b, err := p.ReadBool()
		if err != nil {
			return err
		}
		*buf = strconv.AppendBool(*buf, b)
		return nil
	case BYTE:
		b, err := p.ReadByte()
		if err != nil {
			return err
		}
		if byteAsUint8 {
			*buf = strconv.AppendInt(*buf, int64(uint8(b)), 10)
			return nil
		} else {
			*buf = strconv.AppendInt(*buf, int64(b), 10)
			return nil
		}
	case I16:
		i, err := p.ReadI16()
		if err != nil {
			return err
		}
		*buf = json.EncodeInt64(*buf, int64(i))
		return nil
	case I32:
		i, err := p.ReadI32()
		if err != nil {
			return err
		}
		*buf = json.EncodeInt64(*buf, int64(i))
		return nil
	case I64:
		i, err := p.ReadI64()
		if err != nil {
			return err
		}
		*buf = json.EncodeInt64(*buf, i)
		return nil
	case DOUBLE:
		f, err := p.ReadDouble()
		if err != nil {
			return err
		}
		*buf = json.EncodeFloat64(*buf, f)
		return nil
	case STRING:
		if base64Binary && desc.IsBinary() {
			vs, err := p.ReadBinary(false)
			if err != nil {
				return err
			}
			if !asJson {
				*buf = json.EncodeBase64(*buf, vs)
				return nil
			}
			*buf = json.EncodeBaniry(*buf, vs)
		} else {
			vs, err := p.ReadString(false)
			if err != nil {
				return err
			}
			if !asJson {
				*buf = append(*buf, vs...)
				return nil
			}
			*buf = json.EncodeString(*buf, vs)
		}
		return nil
	case SET, LIST:
		elemType, size, e := p.ReadSetBegin()
		if e != nil {
			return e
		}
		et := desc.Elem()
		if et.Type() != elemType {
			return errDismatchPrimitive
		}
		if asJson {
			*buf = append(*buf, '[')
		}
		for i := 0; i < size; i++ {
			if e := p.EncodeText(et, buf, byteAsUint8, disallowUnknown, base64Binary, useFieldName, asJson); e != nil {
				return e
			}
			if i != size-1 {
				*buf = append(*buf, ',')
			}
		}
		if asJson {
			*buf = append(*buf, ']')
		}
		return nil
	case MAP:
		keyType, valueType, size, e := p.ReadMapBegin()
		if e != nil {
			return e
		}
		et := desc.Elem()
		if et.Type() != valueType || keyType != desc.Key().Type() {
			return errDismatchPrimitive
		}
		if asJson {
			*buf = append(*buf, '{')
		}
		for i := 0; i < size; i++ {
			if e := p.EncodeText(desc.Key(), buf, byteAsUint8, disallowUnknown, base64Binary, useFieldName, asJson); e != nil {
				return e
			}
			*buf = append(*buf, ':')
			if e := p.EncodeText(desc.Elem(), buf, byteAsUint8, disallowUnknown, base64Binary, useFieldName, asJson); e != nil {
				return e
			}
			if i != size-1 {
				*buf = append(*buf, ',')
			}
		}
		if asJson {
			*buf = append(*buf, '}')
		}
		return nil
	case STRUCT:
		st := desc.Struct()
		if asJson {
			*buf = append(*buf, '{')
		}
		hasVal := false
		for {
			_, typ, id, err := p.ReadFieldBegin()
			if err != nil {
				return err
			}
			if typ == STOP {
				break
			}
			if !hasVal {
				hasVal = true
			} else {
				*buf = append(*buf, ',')
			}
			field := st.FieldById(id)
			if field == nil {
				if !disallowUnknown {
					return errUnknonwField
				}
				continue
			}
			if !useFieldName {
				*buf = json.EncodeInt64(*buf, int64(id))
			} else {
				*buf = append(*buf, field.Alias()...)
			}
			*buf = append(*buf, ':')
			if err := p.EncodeText(field.Type(), buf, byteAsUint8, disallowUnknown, base64Binary, useFieldName, asJson); err != nil {
				return err
			}
		}
		if asJson {
			*buf = append(*buf, '}')
		}
		return nil
	default:
		return errUnsupportedType
	}
}

// ReadAnyWithDesc explains thrift data with descriptor and converts it to go interface{}
//   - LIST/SET will be converted to []interface{}
//   - MAP will be converted to map[string]interface{} or map[int]interface{}
//     or map[interface{}]interface (depends on its key type)
//   - STRUCT will be converted to map[FieldID]interface{}
func (p *BinaryProtocol) ReadAnyWithDesc(desc *TypeDescriptor, byteAsUint8 bool, copyString bool, disallowUnknonw bool, useFieldName bool) (interface{}, error) {
	switch desc.Type() {
	case STOP:
		return nil, nil
	case BOOL:
		return p.ReadBool()
	case BYTE:
		v, e := p.ReadByte()
		if e != nil {
			return nil, e
		}
		if !byteAsUint8 {
			return int8(v), nil
		}
		return v, nil
	case I16:
		return p.ReadI16()
	case I32:
		return p.ReadI32()
	case I64:
		return p.ReadI64()
	case DOUBLE:
		return p.ReadDouble()
	case STRING:
		if desc.IsBinary() {
			return p.ReadBinary(copyString)
		} else {
			return p.ReadString(copyString)
		}
	case SET, LIST:
		elemType, size, e := p.ReadSetBegin()
		if e != nil {
			return nil, e
		}
		et := desc.Elem()
		if et.Type() != elemType {
			return nil, errDismatchPrimitive
		}
		ret := make([]interface{}, 0, size)
		for i := 0; i < size; i++ {
			v, e := p.ReadAnyWithDesc(et, byteAsUint8, copyString, disallowUnknonw, useFieldName)
			if e != nil {
				return nil, e
			}
			ret = append(ret, v)
		}
		return ret, p.ReadSetEnd()
	case MAP:
		var ret interface{}
		keyType, valueType, size, e := p.ReadMapBegin()
		if e != nil {
			return nil, e
		}
		et := desc.Elem()
		if et.Type() != valueType || keyType != desc.Key().Type() {
			return nil, errDismatchPrimitive
		}
		if keyType == STRING {
			m := make(map[string]interface{}, size)
			for i := 0; i < size; i++ {
				kv, e := p.ReadString(false)
				if e != nil {
					return nil, e
				}
				vv, e := p.ReadAnyWithDesc(et, byteAsUint8, copyString, disallowUnknonw, useFieldName)
				if e != nil {
					return nil, e
				}
				m[kv] = vv
			}
			ret = m
		} else if keyType.IsInt() {
			m := make(map[int]interface{}, size)
			for i := 0; i < size; i++ {
				kv, e := p.ReadInt(keyType)
				if e != nil {
					return nil, e
				}
				vv, e := p.ReadAnyWithDesc(et, byteAsUint8, copyString, disallowUnknonw, useFieldName)
				if e != nil {
					return nil, e
				}
				m[kv] = vv
			}
			ret = m
		} else {
			m := make(map[interface{}]interface{})
			for i := 0; i < size; i++ {
				kv, e := p.ReadAnyWithDesc(desc.Key(), byteAsUint8, copyString, disallowUnknonw, useFieldName)
				if e != nil {
					return nil, e
				}
				vv, e := p.ReadAnyWithDesc(et, byteAsUint8, copyString, disallowUnknonw, useFieldName)
				if e != nil {
					return nil, e
				}
				switch x := kv.(type) {
				case map[string]interface{}:
					m[&x] = vv
				case map[int]interface{}:
					m[&x] = vv
				case map[interface{}]interface{}:
					m[&x] = vv
				case []interface{}:
					m[&x] = vv
				case map[FieldID]interface{}:
					m[&x] = vv
				default:
					m[kv] = vv
				}
			}
			ret = m
		}
		return ret, p.ReadMapEnd()
	case STRUCT:
		st := desc.Struct()
		var ret map[FieldID]interface{}
		var ret2 map[string]interface{}
		if useFieldName {
			ret2 = make(map[string]interface{}, len(st.Fields()))
		} else {
			ret = make(map[FieldID]interface{}, len(st.Fields()))
		}
		for {
			_, typ, id, err := p.ReadFieldBegin()
			if err != nil {
				return nil, err
			}
			if typ == STOP {
				if useFieldName {
					return ret2, nil
				} else {
					return ret, nil
				}
			}
			next := st.FieldById(id)
			if next == nil {
				if disallowUnknonw {
					return nil, errUnknonwField
				}
				if err := p.Skip(typ, false); err != nil {
					return nil, err
				}
				continue
			}
			vv, err := p.ReadAnyWithDesc(next.Type(), byteAsUint8, copyString, disallowUnknonw, useFieldName)
			if err != nil {
				return nil, err
			}
			if useFieldName {
				ret2[next.Alias()] = vv
			} else {
				ret[id] = vv
			}
		}
	default:
		return nil, errUnsupportedType
	}
}

// WriteStringWithDesc explain simple string val with desc and convert to thrift data
func (p *BinaryProtocol) WriteStringWithDesc(val string, desc *TypeDescriptor, disallowUnknown bool, base64Binary bool) error {
	return p.DecodeText(val, desc, disallowUnknown, base64Binary, true, false)
}

// DecodeText decode special text-encoded val with desc and write it into buffer
// The encoding of val should be compatible with `EncodeText()`
// WARNING: this function is not fully implemented, only support json-encoded string for LIST/MAP/SET/STRUCT
func (p *BinaryProtocol) DecodeText(val string, desc *TypeDescriptor, disallowUnknown bool, base64Binary bool, useFieldName bool, asJson bool) error {
	switch desc.Type() {
	case STRING:
		if asJson {
			v, err := strconv.Unquote(val)
			if err != nil {
				return err
			}
			val = v
		}
		if base64Binary && desc.IsBinary() {
			v, err := base64.StdEncoding.DecodeString(val)
			if err != nil {
				return err
			}
			val = rt.Mem2Str(v)
		}
		return p.WriteString(val)
	case BOOL:
		v, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		return p.WriteBool(v)
	case BYTE:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		return p.WriteByte(byte(i))
	case I16:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		return p.WriteI16(int16(i))
	case I32:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		return p.WriteI32(int32(i))
	case I64:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		return p.WriteI64(i)
	case DOUBLE:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		return p.WriteDouble(f)
	case LIST, SET:
		if !asJson {
			// OPT: Optimize this using json ast in-place parser
			vs := strings.Split(val, ",")
			if err := p.WriteListBegin(desc.Elem().Type(), len(vs)); err != nil {
				return err
			}
			for _, v := range vs {
				err := p.DecodeText(v, desc.Elem(), disallowUnknown, base64Binary, useFieldName, asJson)
				if err != nil {
					return err
				}
			}
			return p.WriteListEnd()
		} else {
			// OPT: Optimize this using json ast in-place parser
			vs := []interface{}{}
			if err := util.SonicUseInt64.UnmarshalFromString(val, &vs); err != nil {
				return err
			}
			if err := p.WriteListBegin(desc.Elem().Type(), len(vs)); err != nil {
				return err
			}
			for _, v := range vs {
				if err := p.WriteAnyWithDesc(desc.Elem(), v, true, false, true); err != nil {
					return err
				}
			}
			return p.WriteListEnd()
		}
	case MAP:
		//TODO: implement it for non-json
		if !asJson {
			return errNotImplemented
		}
		// OPT: Optimize this using json ast in-place parser
		vs := map[string]interface{}{}
		if err := util.SonicUseInt64.UnmarshalFromString(val, &vs); err != nil {
			return err
		}
		if err := p.WriteMapBegin(desc.Key().Type(), desc.Elem().Type(), len(vs)); err != nil {
			return err
		}
		for k, v := range vs {
			err := p.DecodeText(k, desc.Key(), disallowUnknown, base64Binary, useFieldName, asJson)
			if err != nil {
				return err
			}
			if err := p.WriteAnyWithDesc(desc.Elem(), v, true, false, true); err != nil {
				return err
			}
		}
		return p.WriteMapEnd()
	case STRUCT:
		//TODO: implement it for non-json
		if !asJson {
			return errNotImplemented
		}
		// OPT: Optimize this using json ast in-place parser
		var v = make(map[string]interface{})
		err := util.SonicUseInt64.UnmarshalFromString(val, &v)
		if err != nil {
			return err
		}
		return p.WriteAnyWithDesc(desc, v, true, false, true)
	default:
		return errDismatchPrimitive
	}
}

// GoType2ThriftType a go primitive type to a thrift type
// The rules is:
//   - bool -> BOOL
//   - byte/int8 -> BYTE
//   - int16 -> I16
//   - int32 -> I32
//   - int64/int -> I64
//   - int -> I64
//   - float64/float32 -> DOUBLE
//   - string/[]byte -> STRING
//   - []interface{} -> LIST
//   - map[FieldID]interface{} -> STRUCT
//   - map[(int|string|interface{})]interface{} -> MAP
func GoType2ThriftType(val interface{}) (Type, error) {
	_, ok := val.(map[FieldID]interface{})
	if ok {
		return STRUCT, nil
	}
	_, ok = val.([]byte)
	if ok {
		return STRING, nil
	}
	switch reflect.TypeOf(val).Kind() {
	case reflect.Bool:
		return BOOL, nil
	case reflect.Int8, reflect.Uint8:
		return BYTE, nil
	case reflect.Int16, reflect.Uint16:
		return I16, nil
	case reflect.Int32, reflect.Uint32:
		return I32, nil
	case reflect.Int64, reflect.Uint64, reflect.Int, reflect.Uint:
		return I64, nil
	case reflect.Float64:
		return DOUBLE, nil
	case reflect.String:
		return STRING, nil
	case reflect.Slice:
		return LIST, nil
	case reflect.Map:
		return MAP, nil
	case reflect.Struct:
		return STRUCT, nil
	case reflect.Ptr:
		return GoType2ThriftType(reflect.ValueOf(val).Elem().Interface())
	default:
		return STOP, errUnsupportedType
	}
}

// ReadAny reads a thrift value from buffer and convert it to go primitive type
// It basicallly obeys rules in `GoType2ThriftType`.
// Specially,
//   - For INT(8/16/32/64) type, the return type is corresponding int8/int16/int32/int64 by default;
//   - For MAP type, the output key type could be string, int or interface{}, depends on the input key's thrift type.
//   - for STRUCT type, the return type is map[thrift.FieldID]interface{}.
func (p *BinaryProtocol) ReadAny(typ Type, strAsBinary bool, byteAsInt8 bool) (interface{}, error) {
	switch typ {
	case BOOL:
		return p.ReadBool()
	case BYTE:
		if byteAsInt8 {
			n, e := p.ReadByte()
			return int8(n), e
		}
		return p.ReadByte()
	case I16:
		return p.ReadI16()
	case I32:
		return p.ReadI32()
	case I64:
		return p.ReadI64()
	case DOUBLE:
		return p.ReadDouble()
	case STRING:
		if strAsBinary {
			return p.ReadBinary(false)
		}
		return p.ReadString(false)
	case LIST, SET:
		elemType, size, e := p.ReadListBegin()
		if e != nil {
			return nil, e
		}
		ret := make([]interface{}, 0, size)
		for i := 0; i < size; i++ {
			v, e := p.ReadAny(elemType, strAsBinary, byteAsInt8)
			if e != nil {
				return nil, e
			}
			ret = append(ret, v)
		}
		return ret, p.ReadListEnd()
	case MAP:
		keyType, valueType, size, e := p.ReadMapBegin()
		if e != nil {
			return nil, e
		}
		if keyType == STRING {
			ret := make(map[string]interface{}, size)
			for i := 0; i < size; i++ {
				k, e := p.ReadString(false)
				if e != nil {
					return nil, e
				}
				v, e := p.ReadAny(valueType, strAsBinary, byteAsInt8)
				if e != nil {
					return nil, e
				}
				ret[k] = v
			}
			return ret, p.ReadMapEnd()
		} else if keyType.IsInt() {
			ret := make(map[int]interface{}, size)
			for i := 0; i < size; i++ {
				k, e := p.ReadInt(keyType)
				if e != nil {
					return nil, e
				}
				v, e := p.ReadAny(valueType, strAsBinary, byteAsInt8)
				if e != nil {
					return nil, e
				}
				ret[k] = v
			}
			return ret, p.ReadMapEnd()
		} else {
			m := make(map[interface{}]interface{}, size)
			for i := 0; i < size; i++ {
				k, e := p.ReadAny(keyType, strAsBinary, byteAsInt8)
				if e != nil {
					return nil, e
				}
				v, e := p.ReadAny(valueType, strAsBinary, byteAsInt8)
				if e != nil {
					return nil, e
				}
				switch x := k.(type) {
				case map[string]interface{}:
					m[&x] = v
				case map[int]interface{}:
					m[&x] = v
				case map[interface{}]interface{}:
					m[&x] = v
				case []interface{}:
					m[&x] = v
				case map[FieldID]interface{}:
					m[&x] = v
				default:
					m[k] = v
				}
			}
			return m, p.ReadMapEnd()
		}
	case STRUCT:
		ret := make(map[FieldID]interface{})
		for {
			_, typ, id, err := p.ReadFieldBegin()
			if err != nil {
				return nil, err
			}
			if typ == STOP {
				return ret, nil
			}
			v, e := p.ReadAny(typ, strAsBinary, byteAsInt8)
			if e != nil {
				return nil, e
			}
			ret[id] = v
		}
	default:
		return nil, errUnsupportedType
	}
}

// WriteAny write any go primitive type to thrift data, and return top level thrift type
// It basically obeys rules in `GoType2ThriftType`.
// Specially,
//   - for MAP type, the key type should be string or int8/int16/int32/int64/int or interface{}.
//   - for STRUCT type, the val type should be map[thrift.FieldID]interface{}.
func (p *BinaryProtocol) WriteAny(val interface{}, sliceAsSet bool) (Type, error) {
	switch v := val.(type) {
	case bool:
		return BOOL, p.WriteBool(v)
	case byte:
		return BYTE, p.WriteByte(v)
	case int8:
		return BYTE, p.WriteByte(byte(v))
	case int16:
		return I16, p.WriteI16(v)
	case int32:
		return I32, p.WriteI32(v)
	case int64:
		return I64, p.WriteI64(v)
	case int:
		return I64, p.WriteI64(int64(v))
	case float64:
		return DOUBLE, p.WriteDouble(v)
	case float32:
		return DOUBLE, p.WriteDouble(float64(v))
	case string:
		return STRING, p.WriteString(v)
	case []byte:
		return STRING, p.WriteBinary(v)
	case []interface{}:
		if len(v) == 0 {
			return 0, fmt.Errorf("empty []interface is not supported")
		}
		et, e := GoType2ThriftType(v[0])
		if e != nil {
			return 0, e
		}
		if sliceAsSet {
			e = p.WriteSetBegin(et, len(v))
			if e != nil {
				return 0, e
			}
		} else {
			e = p.WriteListBegin(et, len(v))
			if e != nil {
				return 0, e
			}
		}
		for _, vv := range v {
			if _, e := p.WriteAny(vv, sliceAsSet); e != nil {
				return 0, e
			}
		}
		return LIST, p.WriteListEnd()
	case map[string]interface{}:
		if len(v) == 0 {
			return 0, fmt.Errorf("empty map[string]interface is not supported")
		}
		var firstVal interface{}
		for _, vv := range v {
			firstVal = vv
			break
		}
		et, e := GoType2ThriftType(firstVal)
		if e != nil {
			return 0, e
		}
		e = p.WriteMapBegin(STRING, et, len(v))
		if e != nil {
			return 0, e
		}
		for k, vv := range v {
			if e := p.WriteString(k); e != nil {
				return 0, e
			}
			if _, e := p.WriteAny(vv, sliceAsSet); e != nil {
				return 0, e
			}
		}
		return MAP, p.WriteMapEnd()
	case map[byte]interface{}, map[int]interface{}, map[int8]interface{}, map[int16]interface{}, map[int32]interface{}, map[int64]interface{}:
		vr := reflect.ValueOf(v)
		if vr.Len() == 0 {
			return 0, fmt.Errorf("empty map[int]interface{} is not supported")
		}
		it := vr.MapRange()
		it.Next()
		firstKey := it.Key().Interface()
		firstVal := it.Value().Interface()
		kt, e := GoType2ThriftType(firstKey)
		if e != nil {
			return 0, e
		}
		et, e := GoType2ThriftType(firstVal)
		if e != nil {
			return 0, e
		}
		e = p.WriteMapBegin(kt, et, vr.Len())
		if e != nil {
			return 0, e
		}
		if _, e := p.WriteAny(firstKey, sliceAsSet); e != nil {
			return 0, e
		}
		if _, e := p.WriteAny(firstVal, sliceAsSet); e != nil {
			return 0, e
		}
		for it.Next() {
			if _, e := p.WriteAny(it.Key().Interface(), sliceAsSet); e != nil {
				return 0, e
			}
			if _, e := p.WriteAny(it.Value().Interface(), sliceAsSet); e != nil {
				return 0, e
			}
		}
		return MAP, p.WriteMapEnd()
	case map[interface{}]interface{}:
		if len(v) == 0 {
			return 0, fmt.Errorf("empty map[int]interface{} is not supported")
		}
		var firstVal, firstKey interface{}
		for kk, vv := range v {
			firstVal = vv
			firstKey = kk
			break
		}
		kt, e := GoType2ThriftType(firstKey)
		if e != nil {
			return 0, e
		}
		et, e := GoType2ThriftType(firstVal)
		if e != nil {
			return 0, e
		}
		e = p.WriteMapBegin(kt, et, len(v))
		if e != nil {
			return 0, e
		}
		for k, vv := range v {
			switch kt := k.(type) {
			case *map[string]interface{}:
				if _, err := p.WriteAny(*kt, sliceAsSet); err != nil {
					return 0, err
				}
			case *map[int]interface{}:
				if _, err := p.WriteAny(*kt, sliceAsSet); err != nil {
					return 0, err
				}
			case *map[int8]interface{}:
				if _, err := p.WriteAny(*kt, sliceAsSet); err != nil {
					return 0, err
				}
			case *map[int16]interface{}:
				if _, err := p.WriteAny(*kt, sliceAsSet); err != nil {
					return 0, err
				}
			case *map[int32]interface{}:
				if _, err := p.WriteAny(*kt, sliceAsSet); err != nil {
					return 0, err
				}
			case *map[int64]interface{}:
				if _, err := p.WriteAny(*kt, sliceAsSet); err != nil {
					return 0, err
				}
			case *map[FieldID]interface{}:
				if _, err := p.WriteAny(*kt, sliceAsSet); err != nil {
					return 0, err
				}
			case *map[interface{}]interface{}:
				if _, err := p.WriteAny(*kt, sliceAsSet); err != nil {
					return 0, err
				}
			case *[]interface{}:
				if _, err := p.WriteAny(*kt, sliceAsSet); err != nil {
					return 0, err
				}
			default:
				if _, err := p.WriteAny(k, sliceAsSet); err != nil {
					return 0, err
				}
			}
			if _, e := p.WriteAny(vv, sliceAsSet); e != nil {
				return 0, e
			}
		}
		return MAP, p.WriteMapEnd()
	case map[FieldID]interface{}:
		e := p.WriteStructBegin("")
		if e != nil {
			return 0, e
		}
		for k, vv := range v {
			ft, e := GoType2ThriftType(vv)
			if e != nil {
				return 0, e
			}
			if e := p.WriteFieldBegin("", ft, k); e != nil {
				return 0, e
			}
			if _, e := p.WriteAny(vv, sliceAsSet); e != nil {
				return 0, e
			}
		}
		return STRUCT, p.WriteFieldStop()
	default:
		return 0, errUnsupportedType
	}
}

// WriteAnyWithDesc explain desc and val and write them into buffer
//   - LIST/SET will be converted from []interface{}
//   - MAP will be converted from map[string]interface{} or map[int]interface{}
//   - STRUCT will be converted from map[FieldID]interface{}
func (p *BinaryProtocol) WriteAnyWithDesc(desc *TypeDescriptor, val interface{}, cast bool, disallowUnknown bool, useFieldName bool) error {
	switch desc.Type() {
	case STOP:
		return nil
	case BOOL:
		v, ok := val.(bool)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				var err error
				v, err = primitive.ToBool(val)
				if err != nil {
					return err
				}
			}
		}
		return p.WriteBool(v)
	case BYTE:
		v, ok := val.(byte)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				vv, err := primitive.ToInt64(val)
				if err != nil {
					return err
				}
				v = byte(vv)
			}
		}
		return p.WriteByte(v)
	case I16:
		v, ok := val.(int16)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				vv, err := primitive.ToInt64(val)
				if err != nil {
					return err
				}
				v = int16(vv)
			}
		}
		return p.WriteI16(v)
	case I32:
		v, ok := val.(int32)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				vv, err := primitive.ToInt64(val)
				if err != nil {
					return err
				}
				v = int32(vv)
			}
		}
		return p.WriteI32(v)
	case I64:
		v, ok := val.(int64)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				vv, err := primitive.ToInt64(val)
				if err != nil {
					return err
				}
				v = int64(vv)
			}
		}
		return p.WriteI64(v)
	case DOUBLE:
		v, ok := val.(float64)
		if !ok {
			if !cast {
				return errDismatchPrimitive
			} else {
				vv, err := primitive.ToFloat64(val)
				if err != nil {
					return err
				}
				v = float64(vv)
			}
		}
		return p.WriteDouble(v)
	case STRING:
		v, ok := val.(string)
		if !ok {
			vv, ok := val.([]byte)
			if !ok {
				if !cast {
					return errDismatchPrimitive
				} else {
					vv, err := primitive.ToString(val)
					if err != nil {
						return err
					}
					v = string(vv)
				}
			}
			return p.WriteBinary(vv)
		}
		return p.WriteString(v)
	case SET, LIST:
		vs, ok := val.([]interface{})
		if !ok {
			return errDismatchPrimitive
		}
		e := p.WriteSetBegin(desc.Elem().Type(), len(vs))
		if e != nil {
			return e
		}
		for _, v := range vs {
			if e := p.WriteAnyWithDesc(desc.Elem(), v, cast, disallowUnknown, useFieldName); e != nil {
				return e
			}
		}
		return p.WriteSetEnd()
	case MAP:
		if kt := desc.Key().Type(); kt == STRING {
			vs, ok := val.(map[string]interface{})
			if !ok {
				return errDismatchPrimitive
			}
			e := p.WriteMapBegin(desc.Key().Type(), desc.Elem().Type(), len(vs))
			if e != nil {
				return e
			}
			for k, v := range vs {
				if e := p.WriteString(k); e != nil {
					return e
				}
				if e := p.WriteAnyWithDesc(desc.Elem(), v, cast, disallowUnknown, useFieldName); e != nil {
					return e
				}
			}
		} else if kt.IsInt() {
			vi, ok := val.(map[int]interface{})
			if ok {
				e := p.WriteMapBegin(desc.Key().Type(), desc.Elem().Type(), len(vi))
				if e != nil {
					return e
				}
				for k, v := range vi {
					if e := p.WriteInt(kt, k); e != nil {
						return e
					}
					if e := p.WriteAnyWithDesc(desc.Elem(), v, cast, disallowUnknown, useFieldName); e != nil {
						return e
					}
				}
				return nil
			}
			v2, ok := val.(map[int8]interface{})
			if ok {
				e := p.WriteMapBegin(desc.Key().Type(), desc.Elem().Type(), len(v2))
				if e != nil {
					return e
				}
				for k, v := range v2 {
					if e := p.WriteInt(kt, int(k)); e != nil {
						return e
					}
					if e := p.WriteAnyWithDesc(desc.Elem(), v, cast, disallowUnknown, useFieldName); e != nil {
						return e
					}
				}
				return nil
			}
			v3, ok := val.(map[int16]interface{})
			if ok {
				e := p.WriteMapBegin(desc.Key().Type(), desc.Elem().Type(), len(v3))
				if e != nil {
					return e
				}
				for k, v := range v3 {
					if e := p.WriteInt(kt, int(k)); e != nil {
						return e
					}
					if e := p.WriteAnyWithDesc(desc.Elem(), v, cast, disallowUnknown, useFieldName); e != nil {
						return e
					}
				}
				return nil
			}
			v4, ok := val.(map[int32]interface{})
			if ok {
				e := p.WriteMapBegin(desc.Key().Type(), desc.Elem().Type(), len(v4))
				if e != nil {
					return e
				}
				for k, v := range v4 {
					if e := p.WriteInt(kt, int(k)); e != nil {
						return e
					}
					if e := p.WriteAnyWithDesc(desc.Elem(), v, cast, disallowUnknown, useFieldName); e != nil {
						return e
					}
				}
				return nil
			}
			v5, ok := val.(map[int64]interface{})
			if ok {
				e := p.WriteMapBegin(desc.Key().Type(), desc.Elem().Type(), len(v5))
				if e != nil {
					return e
				}
				for k, v := range v5 {
					if e := p.WriteInt(kt, int(k)); e != nil {
						return e
					}
					if e := p.WriteAnyWithDesc(desc.Elem(), v, cast, disallowUnknown, useFieldName); e != nil {
						return e
					}
				}
				return nil
			}
			return errDismatchPrimitive
		} else {
			vv, ok := val.(map[interface{}]interface{})
			if !ok {
				return errDismatchPrimitive
			}
			for k, v := range vv {
				switch kt := k.(type) {
				case *map[string]interface{}:
					if err := p.WriteAnyWithDesc(desc.Key(), *kt, cast, disallowUnknown, useFieldName); err != nil {
						return err
					}
				case *map[int]interface{}:
					if err := p.WriteAnyWithDesc(desc.Key(), *kt, cast, disallowUnknown, useFieldName); err != nil {
						return err
					}
				case *map[int8]interface{}:
					if err := p.WriteAnyWithDesc(desc.Key(), *kt, cast, disallowUnknown, useFieldName); err != nil {
						return err
					}
				case *map[int16]interface{}:
					if err := p.WriteAnyWithDesc(desc.Key(), *kt, cast, disallowUnknown, useFieldName); err != nil {
						return err
					}
				case *map[int32]interface{}:
					if err := p.WriteAnyWithDesc(desc.Key(), *kt, cast, disallowUnknown, useFieldName); err != nil {
						return err
					}
				case *map[int64]interface{}:
					if err := p.WriteAnyWithDesc(desc.Key(), *kt, cast, disallowUnknown, useFieldName); err != nil {
						return err
					}
				case *map[FieldID]interface{}:
					if err := p.WriteAnyWithDesc(desc.Key(), *kt, cast, disallowUnknown, useFieldName); err != nil {
						return err
					}
				case *map[interface{}]interface{}:
					if err := p.WriteAnyWithDesc(desc.Key(), *kt, cast, disallowUnknown, useFieldName); err != nil {
						return err
					}
				case *[]interface{}:
					if err := p.WriteAnyWithDesc(desc.Key(), *kt, cast, disallowUnknown, useFieldName); err != nil {
						return err
					}
				default:
					if err := p.WriteAnyWithDesc(desc.Key(), k, cast, disallowUnknown, useFieldName); err != nil {
						return err
					}
				}
				if e := p.WriteAnyWithDesc(desc.Elem(), v, cast, disallowUnknown, useFieldName); e != nil {
					return e
				}
			}
		}
		return nil
	case STRUCT:
		if useFieldName {
			vs, ok := val.(map[string]interface{})
			if !ok {
				return errDismatchPrimitive
			}
			e := p.WriteStructBegin(desc.Name())
			if e != nil {
				return e
			}
			for id, v := range vs {
				f := desc.Struct().FieldByKey(id)
				if f == nil {
					if disallowUnknown {
						return errUnknonwField
					}
					continue
				}
				if e := p.WriteFieldBegin(f.Alias(), f.Type().Type(), f.ID()); e != nil {
					return e
				}
				if e := p.WriteAnyWithDesc(f.Type(), v, cast, disallowUnknown, useFieldName); e != nil {
					return e
				}
				if e := p.WriteFieldEnd(); e != nil {
					return e
				}
			}
		} else {
			vs, ok := val.(map[FieldID]interface{})
			if !ok {
				return errDismatchPrimitive
			}
			e := p.WriteStructBegin(desc.Name())
			if e != nil {
				return e
			}
			// var r = NewRequiresBitmap()
			// desc.Struct().Requires().CopyTo(r)
			for id, v := range vs {
				f := desc.Struct().FieldById(id)
				if f == nil {
					if disallowUnknown {
						return errUnknonwField
					}
					continue
				}
				// r.Set(f.ID(), OptionalRequireness)
				if e := p.WriteFieldBegin(f.Alias(), f.Type().Type(), f.ID()); e != nil {
					return e
				}
				if e := p.WriteAnyWithDesc(f.Type(), v, cast, disallowUnknown, useFieldName); e != nil {
					return e
				}
				if e := p.WriteFieldEnd(); e != nil {
					return e
				}
			}
			// if e = r.CheckRequires(desc.Struct(), false, nil); e != nil {
			// 	return e
			// }
			// FreeRequiresBitmap(r)
		}
		return p.WriteStructEnd()
	default:
		return errUnsupportedType
	}
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

// BinaryEncoding is the implementation of Encoding for binary encoding.
type BinaryEncoding struct{}

func (b BinaryEncoding) EncodeEmpty(typ, et, kt Type, buf []byte) ([]byte, error) {
	switch typ {
	case BOOL:
		rt.GuardSlice(&buf, 1)
		b.EncodeBool(buf[len(buf):len(buf)+1], false)
		buf = buf[:len(buf)+1]
	case DOUBLE:
		rt.GuardSlice(&buf, 8)
		b.EncodeDouble(buf[len(buf):len(buf)+8], 0)
		buf = buf[:len(buf)+8]
	case I08:
		rt.GuardSlice(&buf, 1)
		b.EncodeByte(buf[len(buf):len(buf)+1], 0)
		buf = buf[:len(buf)+1]
	case I16:
		rt.GuardSlice(&buf, 2)
		b.EncodeInt16(buf[len(buf):len(buf)+2], 0)
		buf = buf[:len(buf)+2]
	case I32:
		rt.GuardSlice(&buf, 4)
		b.EncodeInt32(buf[len(buf):len(buf)+4], 0)
		buf = buf[:len(buf)+4]
	case I64:
		rt.GuardSlice(&buf, 8)
		b.EncodeInt64(buf[len(buf):len(buf)+8], 0)
		buf = buf[:len(buf)+8]
	case STRING:
		rt.GuardSlice(&buf, 4)
		b.EncodeString(buf[len(buf):len(buf)+4], "")
		buf = buf[:len(buf)+4]
	case STRUCT:
		rt.GuardSlice(&buf, 1)
		buf = append(buf, 0)
	case MAP:
		rt.GuardSlice(&buf, 6)
		buf = append(buf, byte(kt))
		buf = append(buf, byte(et))
		b.EncodeInt32(buf[len(buf):len(buf)+4], 0)
		buf = buf[:len(buf)+4]
	case LIST, SET:
		rt.GuardSlice(&buf, 5)
		buf = append(buf, byte(et))
		b.EncodeInt32(buf[len(buf):len(buf)+4], 0)
		buf = buf[:len(buf)+4]
	default:
		return nil, errUnsupportedType
	}
	return buf, nil
}

// EncodeBool encodes a bool value.
func (BinaryEncoding) EncodeBool(b []byte, v bool) {
	if v {
		b[0] = 1
	} else {
		b[0] = 0
	}
}

// EncodeByte encodes a byte value.
func (BinaryEncoding) EncodeByte(b []byte, v byte) {
	b[0] = byte(v)
}

// EncodeInt16 encodes a int16 value.
func (BinaryEncoding) EncodeInt16(b []byte, v int16) {
	binary.BigEndian.PutUint16(b, uint16(v))
}

// EncodeInt32 encodes a int32 value.
func (BinaryEncoding) EncodeInt32(b []byte, v int32) {
	binary.BigEndian.PutUint32(b, uint32(v))
}

// EncodeInt64 encodes a int64 value.
func (BinaryEncoding) EncodeInt64(b []byte, v int64) {
	binary.BigEndian.PutUint64(b, uint64(v))
}

func (BinaryEncoding) EncodeDouble(b []byte, v float64) {
	binary.BigEndian.PutUint64(b, math.Float64bits(v))
}

// EncodeString encodes a string value.
func (BinaryEncoding) EncodeString(b []byte, v string) {
	binary.BigEndian.PutUint32(b, uint32(len(v)))
	copy(b[4:], v)
}

// EncodeBinary encodes a binary value.
func (BinaryEncoding) EncodeBinary(b []byte, v []byte) {
	binary.BigEndian.PutUint32(b, uint32(len(v)))
	copy(b[4:], v)
}

// EncodeFieldBegin encodes a field begin.
func (BinaryEncoding) EncodeFieldBegin(b []byte, t Type, id FieldID) {
	b[0] = byte(t)
	binary.BigEndian.PutUint16(b[1:], uint16(id))
}

// EncodeFieldEnd encodes a field end.
func (BinaryEncoding) DecodeBool(b []byte) bool {
	return int8(b[0]) == 1
}

// DecodeByte decodes a byte value.
func (BinaryEncoding) DecodeByte(b []byte) byte {
	return byte(b[0])
}

// DecodeInt16 decodes a int16 value.
func (BinaryEncoding) DecodeInt16(b []byte) int16 {
	return int16(binary.BigEndian.Uint16(b))
}

// DecodeInt32 decodes a int32 value.
func (BinaryEncoding) DecodeInt32(b []byte) int32 {
	return int32(binary.BigEndian.Uint32(b))
}

// DecodeInt64 decodes a int64 value.
func (BinaryEncoding) DecodeInt64(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

// DecodeDouble decodes a double value.
func (BinaryEncoding) DecodeDouble(b []byte) float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(b))
}

// DecodeString decodes a string value.
func (d BinaryEncoding) DecodeString(b []byte) (value string) {
	size := d.DecodeInt32(b)
	v := (*rt.GoString)(unsafe.Pointer(&value))
	v.Ptr = rt.IndexPtr(*(*unsafe.Pointer)(unsafe.Pointer(&b)), byteTypeSize, 4)
	v.Len = int(size)
	return
}

// DecodeBinary decodes a binary value.
func (d BinaryEncoding) DecodeBytes(b []byte) (value []byte) {
	size := d.DecodeInt32(b)
	v := (*rt.GoSlice)(unsafe.Pointer(&value))
	v.Ptr = rt.IndexPtr(*(*unsafe.Pointer)(unsafe.Pointer(&b)), byteTypeSize, 4)
	v.Len = int(size)
	v.Cap = int(size)
	return
}
