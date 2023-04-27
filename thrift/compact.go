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
	"io"
	"math"
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

const (
	// proto ID (u8) + ver and typeId (u8)
	// max length varint32 seqId + max length varint32 size RPC name

	compactMsgVarint32MaxLen    = 5
	compactMsgVarint64MaxLen    = 10
	compactMsgHeaderMaxFixedLen = 1 + 1 + compactMsgVarint32MaxLen + compactMsgVarint32MaxLen
	compactMsgFooterFixedLen    = 8 // sparse

	COMPACT_PROTOCOL_ID       = 0x082
	COMPACT_VERSION           = 1
	COMPACT_VERSION_MASK      = 0x1f
	COMPACT_TYPE_MASK         = 0x0E0
	COMPACT_TYPE_BITS         = 0x07
	COMPACT_TYPE_SHIFT_AMOUNT = 5
)

const (
	COMPACT_STOP          = 0x00
	COMPACT_BOOLEAN_TRUE  = 0x01
	COMPACT_BOOLEAN_FALSE = 0x02
	COMPACT_BYTE          = 0x03
	COMPACT_I16           = 0x04
	COMPACT_I32           = 0x05
	COMPACT_I64           = 0x06
	COMPACT_DOUBLE        = 0x07
	COMPACT_BINARY        = 0x08
	COMPACT_LIST          = 0x09
	COMPACT_SET           = 0x0A
	COMPACT_MAP           = 0x0B
	COMPACT_STRUCT        = 0x0C
)

type tcompactType byte

var (
	cpPool = sync.Pool{
		New: func() interface{} {
			return &CompactProtocol{
				Buf:     make([]byte, 0, defaultBufferSize),
				ReadOff: 0,
			}
		},
	}

	ttypeToCompactTypeTable = [...]tcompactType{
		STOP:   COMPACT_STOP,
		BOOL:   COMPACT_BOOLEAN_TRUE,
		BYTE:   COMPACT_BYTE,
		I16:    COMPACT_I16,
		I32:    COMPACT_I32,
		I64:    COMPACT_I64,
		DOUBLE: COMPACT_DOUBLE,
		STRING: COMPACT_BINARY,
		LIST:   COMPACT_LIST,
		SET:    COMPACT_SET,
		MAP:    COMPACT_MAP,
		STRUCT: COMPACT_STRUCT,
	}
)

// NewCompactProtocol get a new compact protocol from sync.Pool
func NewCompactProtocol(buf []byte) *CompactProtocol {
	cp := cpPool.Get().(*CompactProtocol)
	cp.Buf = buf
	return cp
}

// NewCompactProtocolBuffer gets a new compact protocol from sync.Pool
// and reuse the buffer in pool
func NewCompactProtocolBuffer() *CompactProtocol {
	cp := cpPool.Get().(*CompactProtocol)
	return cp
}

// FreCompactProtocol resets the buffer and puts the compact protocol back to sync.Pool
func FreeCompactProtocolBuffer(cp *CompactProtocol) {
	cp.Reset()
	cpPool.Put(cp)
}

// Recycle put the protocol back to sync.Pool
func (p *CompactProtocol) Recycle() {
	FreeCompactProtocolBuffer(p)
}

// CompactProtocol implements the CompactProtocol
type CompactProtocol struct {
	Buf []byte

	// SAFETY: KEEP IN-SYNC WITH LAYOUT IN thrift.h AND thrift/compact.go
	PendingBoolField struct {
		ID    FieldID
		Typ   Type
		Valid bool
		Value bool
	}

	LastFieldID      FieldID
	LastFieldIDStack []FieldID

	ReadOff int
}

// Reset resets the buffer and read position
func (p *CompactProtocol) Reset() {
	p.Buf = p.Buf[:0]
	p.ReadOff = 0
}

func (p *CompactProtocol) RawBuf() []byte {
	return p.Buf
}

// Left returns the left bytes to read
// aka. RemainingBytes left to read
func (p *CompactProtocol) Left() int {
	return len(p.Buf) - p.ReadOff
}

/**
 * Message related methods
 */

// GetCompactMessageHeaderAndFooter writes the message parameters into header and footer
func GetCompactMessageHeaderAndFooter(methodName string, msgTyp TMessageType, structID FieldID, seqID int32) (header []byte, footer []byte, err error) {
	var writer = CompactProtocol{}

	// write header
	header = make([]byte, 0, compactMsgHeaderMaxFixedLen+len(methodName))
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
	footer = make([]byte, 0, compactMsgFooterFixedLen)
	writer.Buf = footer
	err = writer.WriteFieldEnd()
	if err != nil {
		return
	}
	// TODO: STOP, but what about method EXCEPTION fields?
	err = writer.WriteFieldStop()
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

// WrapCompactMessage wraps the message body with header and footer
func WrapCompactBody(body []byte, methodName string, msgTyp TMessageType, structID FieldID, seqID int32) ([]byte, error) {
	// write header
	buf := make([]byte, 0, 0+
		compactMsgFooterFixedLen+
		compactMsgHeaderMaxFixedLen+len(methodName)+len(body))
	writer := CompactProtocol{Buf: buf}
	writer.WriteMessageBegin(methodName, msgTyp, seqID)
	writer.WriteStructBegin("")

	// SUCCESS field
	writer.WriteFieldBegin("", STRUCT, structID)
	writer.Buf = append(writer.Buf, body...)
	writer.WriteFieldEnd()

	// TODO: STOP, but what about method EXCEPTION fields?
	writer.WriteFieldStop()

	writer.WriteStructEnd()
	writer.WriteMessageEnd()
	return writer.Buf, nil
}

// UnwrapCompactMessage unwraps the message paramters from the buf
func UnwrapCompactMessage(proto meta.Encoding, buf []byte) (name string, callType TMessageType, seqID int32, structID FieldID, body []byte, err error) {
	var reader = CompactProtocol{
		Buf: buf,
	}
	return reader.UnwrapBody()
}

func (p CompactProtocol) UnwrapBody() (string, TMessageType, int32, FieldID, []byte, error) {
	name, rTyp, seqID, err := p.ReadMessageBegin(false)
	if err != nil {
		return name, rTyp, seqID, 0, nil, err
	}
	// read the success struct
	_, _, structID, err := p.ReadFieldBegin()
	if err != nil {
		return name, rTyp, seqID, structID, nil, err
	}
	// There's always a field STOP on RPC Result struct
	if p.ReadOff > len(p.Buf)-1 {
		return name, rTyp, seqID, structID, nil, io.EOF
	}

	return name, rTyp, seqID, 0, p.Buf[p.ReadOff : len(p.Buf)-1], err
}

/**
 * Writing Methods
 */

func (p *CompactProtocol) WriteMessageBegin(name string, typeID TMessageType, seqID int32) error {
	e := p.WriteByte(COMPACT_PROTOCOL_ID)
	if e != nil {
		return e
	}
	e = p.WriteByte((COMPACT_VERSION & COMPACT_VERSION_MASK) |
		((byte(typeID) << COMPACT_TYPE_SHIFT_AMOUNT) & COMPACT_TYPE_MASK))
	if e != nil {
		return e
	}
	_, e = p.writeVarint32(seqID)
	if e != nil {
		return e
	}
	e = p.WriteString(name)
	if e != nil {
		return e
	}
	return nil
}
func (p *CompactProtocol) WriteMessageEnd() error {
	return nil
}

func (p *CompactProtocol) WriteStructBegin(name string) error {
	p.LastFieldIDStack = append(p.LastFieldIDStack, p.LastFieldID)
	p.LastFieldID = 0
	return nil
}
func (p *CompactProtocol) WriteStructEnd() error {
	p.LastFieldID = p.LastFieldIDStack[len(p.LastFieldIDStack)-1]
	p.LastFieldIDStack = p.LastFieldIDStack[:len(p.LastFieldIDStack)-1]
	return nil
}

func (p *CompactProtocol) writeFieldBeginInternal(name string, typeID Type, id FieldID, typeOverride tcompactType) error {
	var typeToWrite tcompactType
	if typeOverride == 0xFF {
		typeToWrite = p.getCompactType(typeID)
	} else {
		typeToWrite = typeOverride
	}
	cid := int(id)
	lid := int(p.LastFieldID)
	if cid > lid && cid-lid <= 15 {
		e := p.WriteByte(byte((cid-lid)<<4) | byte(typeToWrite))
		if e != nil {
			return e
		}
	} else {
		e := p.WriteByte(byte(typeToWrite))
		if e != nil {
			return e
		}
		e = p.WriteI16(int16(id))
		if e != nil {
			return e
		}
	}
	p.LastFieldID = FieldID(cid)
	return nil
}
func (p *CompactProtocol) WriteFieldBegin(name string, typeID Type, id FieldID) error {
	if typeID == BOOL {
		p.PendingBoolField.ID = id
		p.PendingBoolField.Valid = true
		return nil
	}
	return p.writeFieldBeginInternal(name, typeID, id, 0xFF)
}
func (p *CompactProtocol) WriteFieldEnd() error {
	return nil
}
func (p *CompactProtocol) WriteFieldStop() error {
	e := p.WriteByte(byte(COMPACT_STOP))
	if e != nil {
		return e
	}
	return nil
}

func (p *CompactProtocol) WriteMapBegin(keyType, valueType Type, size int) error {
	if size <= 0 {
		return p.WriteByte(COMPACT_STOP)
	}
	_, e := p.writeVarint32(int32(size))
	if e != nil {
		return e
	}
	return p.WriteByte(byte(p.getCompactType(keyType))<<4 | byte(p.getCompactType(valueType)))
}
func (p *CompactProtocol) WriteMapBeginWithSizePos(keyType, valueType Type, size int) (int, error) {
	// TODO: implement this
	panic("not implemented")
}
func (p *CompactProtocol) WriteMapEnd() error {
	return nil
}

func (p *CompactProtocol) writeCollectionBegin(elemType Type, size int) error {
	if size <= 14 {
		return p.WriteByte(byte(int32(size<<4) | int32(p.getCompactType(elemType))))
	}
	e := p.WriteByte(0xf0 | byte(p.getCompactType(elemType)))
	if e != nil {
		return e
	}
	_, e = p.writeVarint32(int32(size))
	return e
}
func (p *CompactProtocol) writeCollectionBeginWithSizePos(elemType Type, size int) (int, error) {
	// TODO: implement this
	panic("not implemented")
}
func (p *CompactProtocol) writeCollectionEnd() error {
	return nil
}

func (c *CompactProtocol) WriteListBegin(elemType Type, size int) error {
	return c.writeCollectionBegin(elemType, size)
}
func (c *CompactProtocol) WriteListBeginWithSizePos(elemType Type, size int) (int, error) {
	return c.writeCollectionBeginWithSizePos(elemType, size)
}
func (c *CompactProtocol) WriteListEnd() error {
	return c.writeCollectionEnd()
}

func (c *CompactProtocol) WriteSetBegin(elemType Type, size int) error {
	return c.writeCollectionBegin(elemType, size)
}
func (c *CompactProtocol) WriteSetBeginWithSizePos(elemType Type, size int) (int, error) {
	return c.writeCollectionBeginWithSizePos(elemType, size)
}
func (c *CompactProtocol) WriteSetEnd() error {
	return c.writeCollectionEnd()
}

func (p *CompactProtocol) WriteBool(value bool) error {
	var v tcompactType = COMPACT_BOOLEAN_FALSE
	if value {
		v = COMPACT_BOOLEAN_TRUE
	}
	if p.PendingBoolField.Valid {
		p.PendingBoolField.Valid = false
		e := p.writeFieldBeginInternal("", BOOL, p.PendingBoolField.ID, v)
		if e != nil {
			return e
		}
		return nil
	}
	e := p.WriteByte(byte(v))
	if e != nil {
		return e
	}
	return nil
}

func (p *CompactProtocol) WriteByte(value byte) error {
	p.Buf = append(p.Buf, byte(value))
	return nil
}

func (p *CompactProtocol) WriteI16(value int16) error {
	_, e := p.writeVarint32(p.int32ToZigzag(int32(value)))
	if e != nil {
		return e
	}
	return nil
}

func (p *CompactProtocol) ModifyI32(pos int, value int32) error {
	// TODO: implement this
	panic("not implemented")
}

func (p *CompactProtocol) WriteI32(value int32) error {
	_, e := p.writeVarint32(p.int32ToZigzag(value))
	if e != nil {
		return e
	}
	return nil
}

func (p *CompactProtocol) WriteI64(value int64) error {
	_, e := p.writeVarint64(p.int64ToZigzag(value))
	if e != nil {
		return e
	}
	return nil
}

func (p *CompactProtocol) WriteInt(t Type, value int) error {
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

func (p *CompactProtocol) WriteDouble(value float64) error {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], math.Float64bits(value))

	obuf, e := p.malloc(8)
	if e != nil {
		return e
	}
	copy(obuf[:], buf[:])
	return nil
}

func (p *CompactProtocol) WriteString(value string) error {
	_, e := p.writeVarint32(int32(len(value)))
	if e != nil {
		return e
	}
	if len(value) <= 0 {
		return nil
	}
	p.Buf = append(p.Buf, value...)
	return nil
}

func (p *CompactProtocol) WriteBinary(value []byte) error {
	_, e := p.writeVarint32(int32(len(value)))
	if e != nil {
		return e
	}
	if len(value) <= 0 {
		return nil
	}
	p.Buf = append(p.Buf, value...)
	return nil
}

// WriteDefaultOrEmpty write defaultvalue if any, otherwise write zero value
func (p *CompactProtocol) WriteDefaultOrEmpty(field *FieldDescriptor) error {
	if dv := field.DefaultValue(); dv != nil {
		p.Buf = append(p.Buf, dv.ThriftCompact()...)
		return nil
	}
	return p.WriteEmpty(field.Type())
}

// WriteEmpty write zero value
func (p *CompactProtocol) WriteEmpty(desc *TypeDescriptor) error {
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
		if err := p.WriteStructBegin(""); err != nil {
			return err
		}
		for _, f := range desc.Struct().Fields() {
			if f.Required() == OptionalRequireness {
				continue
			}
			if err := p.WriteFieldBegin(f.Name(), f.Type().Type(), f.ID()); err != nil {
				return err
			}
			if err := p.WriteEmpty(f.Type()); err != nil {
				return err
			}
			if err := p.WriteFieldEnd(); err != nil {
				return err
			}
		}
		return p.WriteStructEnd()
	default:
		return errors.New("invalid type")
	}
}

func (p *CompactProtocol) writeVarint32Internal(buf []byte, v int32) (int, error) {
	idx := 0
	for {
		if (v & ^0x7F) == 0 {
			buf[idx] = byte(v)
			idx++
			break
		} else {
			buf[idx] = byte((v & 0x7F) | 0x80)
			idx++
			u := uint32(v)
			v = int32(u >> 7)
		}
	}
	return idx, nil
}

func (p *CompactProtocol) writeVarint32(v int32) (int, error) {
	var buf [compactMsgVarint32MaxLen]byte
	var idx int
	var err error
	idx, err = p.writeVarint32Internal(buf[:], v)
	if err != nil {
		return 0, err
	}
	obuf, err := p.malloc(idx)
	if err != nil {
		return 0, err
	}
	copy(obuf[0:idx], buf[0:idx])
	return idx, nil
}

func (p *CompactProtocol) writeVarint64Internal(buf []byte, v int64) (int, error) {
	idx := 0
	for {
		if (v & ^0x7F) == 0 {
			buf[idx] = byte(v)
			idx++
			break
		} else {
			buf[idx] = byte((v & 0x7F) | 0x80)
			idx++
			u := uint64(v)
			v = int64(u >> 7)
		}
	}
	return idx, nil
}

func (p *CompactProtocol) writeVarint64(v int64) (int, error) {
	var buf [compactMsgVarint64MaxLen]byte
	var idx int
	var err error
	idx, err = p.writeVarint64Internal(buf[:], v)
	if err != nil {
		return 0, err
	}
	obuf, err := p.malloc(idx)
	if err != nil {
		return 0, err
	}
	copy(obuf[0:idx], buf[0:idx])
	return idx, nil
}

func (p *CompactProtocol) int32ToZigzag(n int32) int32 {
	return (n << 1) ^ (n >> 31)
}
func (p *CompactProtocol) int64ToZigzag(n int64) int64 {
	return (n << 1) ^ (n >> 63)
}

/**
 * Reading methods
 */

// ReadMessageBegin ...
func (p *CompactProtocol) ReadMessageBegin(copyString bool) (name string, typeID TMessageType, seqID int32, err error) {
	var protoID byte
	protoID, err = p.ReadByte()
	if err != nil {
		return
	}
	if protoID != COMPACT_PROTOCOL_ID {
		return name, typeID, seqID, errInvalidVersion
	}
	var versionAndType byte
	versionAndType, err = p.ReadByte()
	if err != nil {
		return
	}
	version := versionAndType & COMPACT_VERSION_MASK
	typeID = TMessageType((versionAndType >> COMPACT_TYPE_SHIFT_AMOUNT) & COMPACT_TYPE_BITS)
	if version != COMPACT_VERSION {
		err = errInvalidVersion
		return
	}
	seqID, err = p.readVarint32()
	if err != nil {
		return
	}
	name, err = p.ReadString(copyString)
	return
}

// ReadMessageEnd ...
func (p *CompactProtocol) ReadMessageEnd() error {
	return nil
}

func (p *CompactProtocol) ReadStructBegin() (name string, err error) {
	p.LastFieldIDStack = append(p.LastFieldIDStack, p.LastFieldID)
	p.LastFieldID = 0
	return
}
func (p *CompactProtocol) ReadStructEnd() error {
	p.LastFieldID = p.LastFieldIDStack[len(p.LastFieldIDStack)-1]
	p.LastFieldIDStack = p.LastFieldIDStack[0 : len(p.LastFieldIDStack)-1]
	return nil
}

func (p *CompactProtocol) ReadFieldBegin() (name string, typeID Type, id FieldID, err error) {
	var t byte
	t, err = p.ReadByte()
	if err != nil {
		return
	}

	// if it's a stop, then we can return immediately, as the struct is over.
	if Type(t&0x0f) == STOP {
		return "", STOP, 0, nil
	}

	// mask off the 4 MSB of the type header. it could contain a field id delta.
	modifier := int16((t & 0xf0) >> 4)
	if modifier == 0 {
		var id2 int16
		id2, err = p.ReadI16()
		if err != nil {
			return
		}
		if id2 < 0 {
			err = errInvalidDataSize
			return
		}
		id = FieldID(id2)
	} else {
		id = FieldID(int16(p.LastFieldID) + modifier)
	}

	var valid bool
	typeID, valid = tcompactType(t & 0x0F).getTType()
	if !valid {
		err = errInvalidDataType
		return
	}
	return
}
func (p *CompactProtocol) ReadFieldEnd() error {
	return nil
}

func (p *CompactProtocol) ReadMapBegin() (keyType, elemType Type, size int, err error) {
	size32, e := p.readVarint32()
	if e != nil {
		err = e
		return
	}
	if size32 < 0 {
		err = errInvalidDataSize
		return
	}
	size = int(size32)

	keyAndValueType := byte(STOP)
	if size != 0 {
		keyAndValueType, err = p.ReadByte()
		if err != nil {
			return
		}
	}

	var valid bool
	keyType, valid = tcompactType(keyAndValueType >> 4).getTType()
	if !valid {
		err = errInvalidDataType
		return
	}
	elemType, valid = tcompactType(keyAndValueType & 0x0F).getTType()
	if !valid {
		err = errInvalidDataType
		return
	}
	return
}
func (p *CompactProtocol) ReadMapEnd() error {
	return nil
}

func (p *CompactProtocol) readCollectionBegin() (elemType Type, size int, err error) {
	var sizeAndType byte
	sizeAndType, err = p.ReadByte()
	if err != nil {
		return
	}
	size = int((sizeAndType >> 4) & 0x0F)
	if size == 15 {
		var size2 int32
		size2, err = p.readVarint32()
		if err != nil {
			return
		}
		if size2 < 0 {
			err = errInvalidDataSize
			return
		}
		size = int(size2)
	}
	var valid bool
	elemType, valid = tcompactType(sizeAndType).getTType()
	if !valid {
		return 0, 0, errInvalidDataType
	}
	return
}
func (p *CompactProtocol) readCollectionEnd() error {
	return nil
}

func (p *CompactProtocol) ReadListBegin() (elemType Type, size int, err error) {
	return p.readCollectionBegin()
}
func (p *CompactProtocol) ReadListEnd() error {
	return p.readCollectionEnd()
}

func (p *CompactProtocol) ReadSetBegin() (elemType Type, size int, err error) {
	return p.readCollectionBegin()
}
func (p *CompactProtocol) ReadSetEnd() error {
	return p.readCollectionEnd()
}

func (p *CompactProtocol) ReadBool() (bool, error) {
	if p.PendingBoolField.Valid {
		p.PendingBoolField.Valid = false
		return p.PendingBoolField.Value, nil
	}
	v, err := p.ReadByte()
	return v == COMPACT_BOOLEAN_TRUE, err
}

// ReadByte ...
func (p *CompactProtocol) ReadByte() (value byte, err error) {
	buf, err := p.next(1)
	if err != nil {
		return value, err
	}
	return byte(buf[0]), err
}

func (p *CompactProtocol) ReadI16() (value int16, err error) {
	v, err := p.ReadI16()
	return int16(v), err
}

func (p *CompactProtocol) ReadI32() (value int32, err error) {
	value, err = p.readVarint32()
	if err != nil {
		return
	}
	value = p.zigzagToInt32(value)
	return
}

func (p *CompactProtocol) ReadI64() (value int64, err error) {
	value, err = p.readVarint64()
	if err != nil {
		return
	}
	value = p.zigzagToInt64(value)
	return
}

func (p *CompactProtocol) ReadInt(t Type) (value int, err error) {
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

func (p *CompactProtocol) ReadDouble() (value float64, err error) {
	buf, err := p.next(8)
	if err != nil {
		return value, err
	}
	value = math.Float64frombits(p.bytesToUint64(buf))
	return
}

func (p *CompactProtocol) ReadString(copy bool) (value string, err error) {
	length, err := p.readVarint32()
	if err != nil {
		return value, err
	}
	if length < 0 || int(length) > len(p.Buf)-p.ReadOff {
		return value, errInvalidDataSize
	}

	if copy {
		value = string(*(*[]byte)(unsafe.Pointer(&rt.GoSlice{
			Ptr: rt.IndexPtr(*(*unsafe.Pointer)(unsafe.Pointer(&p.Buf)), byteTypeSize, p.ReadOff),
			Len: int(length),
			Cap: int(length),
		})))
	} else {
		v := (*rt.GoString)(unsafe.Pointer(&value))
		v.Ptr = rt.IndexPtr(*(*unsafe.Pointer)(unsafe.Pointer(&p.Buf)), byteTypeSize, p.ReadOff)
		v.Len = int(length)
	}

	p.ReadOff += int(length)
	return
}

// ReadBinary ...
func (p *CompactProtocol) ReadBinary(copyBytes bool) (value []byte, err error) {
	length, err := p.readVarint32()
	if err != nil {
		return value, err
	}
	if length < 0 || int(length) > len(p.Buf)-p.ReadOff {
		return nil, errInvalidDataSize
	}

	if copyBytes {
		value = make([]byte, int(length))
		copy(value, *(*[]byte)(unsafe.Pointer(&rt.GoSlice{
			Ptr: rt.IndexPtr(*(*unsafe.Pointer)(unsafe.Pointer(&p.Buf)), byteTypeSize, p.ReadOff),
			Len: int(length),
			Cap: int(length),
		})))
	} else {
		v := (*rt.GoString)(unsafe.Pointer(&value))
		v.Ptr = rt.IndexPtr(*(*unsafe.Pointer)(unsafe.Pointer(&p.Buf)), byteTypeSize, p.ReadOff)
		v.Len = int(length)
	}

	p.ReadOff += int(length)

	return
}

// Given a tCompactType constant, convert it to its corresponding
// TType value.
func (p tcompactType) getTType() (Type, bool) {
	switch byte(p) & 0x0f {
	case COMPACT_STOP:
		return STOP, true
	case COMPACT_BOOLEAN_TRUE, COMPACT_BOOLEAN_FALSE:
		return BOOL, true
	case COMPACT_BYTE:
		return BYTE, true
	case COMPACT_I16:
		return I16, true
	case COMPACT_I32:
		return I32, true
	case COMPACT_I64:
		return I64, true
	case COMPACT_DOUBLE:
		return DOUBLE, true
	case COMPACT_BINARY:
		return STRING, true
	case COMPACT_LIST:
		return LIST, true
	case COMPACT_SET:
		return SET, true
	case COMPACT_MAP:
		return MAP, true
	case COMPACT_STRUCT:
		return STRUCT, true
	}
	return COMPACT_STOP, false
}

// Convert from zigzag int to int.
func (p *CompactProtocol) zigzagToInt32(n int32) int32 {
	u := uint32(n)
	return int32(u>>1) ^ -(n & 1)
}

// Convert from zigzag long to long.
func (p *CompactProtocol) zigzagToInt64(n int64) int64 {
	u := uint64(n)
	return int64(u>>1) ^ -(n & 1)
}

func (p *CompactProtocol) bytesToUint64(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}
func (p *CompactProtocol) readVarint32() (int32, error) {
	v, err := p.readVarint64()
	return int32(v), err
}

func (p *CompactProtocol) readVarint64() (int64, error) {
	shift := uint(0)
	result := int64(0)
	for {
		b, err := p.ReadByte()
		if err != nil {
			return 0, err
		}
		result |= int64(b&0x7F) << int64(shift)
		if (b & 0x80) != 0x80 {
			break
		}
		shift += 7
	}
	return result, nil
}

/**
 * Extend
 */

// ReadStringWithDesc explains thrift data with desc and converts to simple string
func (p *CompactProtocol) ReadStringWithDesc(desc *TypeDescriptor, buf *[]byte, byteAsUint8 bool, disallowUnknown bool, base64Binary bool) error {
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

func (p *CompactProtocol) EncodeText(desc *TypeDescriptor, buf *[]byte, byteAsUint8 bool, disallowUnknown bool, base64Binary bool, useFieldName bool, asJson bool) error {
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
				*buf = json.EncodeString(*buf, field.Alias())
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
func (p *CompactProtocol) ReadAnyWithDesc(desc *TypeDescriptor, copyString bool, disallowUnknonw bool, useFieldName bool) (interface{}, error) {
	switch desc.Type() {
	case STOP:
		return nil, nil
	case BOOL:
		return p.ReadBool()
	case BYTE:
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
			v, e := p.ReadAnyWithDesc(et, copyString, disallowUnknonw, useFieldName)
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
				vv, e := p.ReadAnyWithDesc(et, copyString, disallowUnknonw, useFieldName)
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
				vv, e := p.ReadAnyWithDesc(et, copyString, disallowUnknonw, useFieldName)
				if e != nil {
					return nil, e
				}
				m[kv] = vv
			}
			ret = m
		} else {
			m := make(map[interface{}]interface{})
			for i := 0; i < size; i++ {
				kv, e := p.ReadAnyWithDesc(desc.Key(), copyString, disallowUnknonw, useFieldName)
				if e != nil {
					return nil, e
				}
				vv, e := p.ReadAnyWithDesc(et, copyString, disallowUnknonw, useFieldName)
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
				if !disallowUnknonw {
					return nil, errUnknonwField
				}
				continue
			}
			vv, err := p.ReadAnyWithDesc(next.Type(), copyString, disallowUnknonw, useFieldName)
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
func (p *CompactProtocol) WriteStringWithDesc(val string, desc *TypeDescriptor, disallowUnknown bool, base64Binary bool) error {
	return p.DecodeText(val, desc, disallowUnknown, base64Binary, true, false)
}

// DecodeText decode special text-encoded val with desc and write it into buffer
// The encoding of val should be compatible with `EncodeText()`
// WARNING: this function is not fully implemented, only support json-encoded string for LIST/MAP/SET/STRUCT
func (p *CompactProtocol) DecodeText(val string, desc *TypeDescriptor, disallowUnknown bool, base64Binary bool, useFieldName bool, asJson bool) error {
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

// WriteAnyWithDesc explain desc and val and write them into buffer
//   - LIST/SET will be converted from []interface{}
//   - MAP will be converted from map[string]interface{} or map[int]interface{}
//   - STRUCT will be converted from map[FieldID]interface{}
func (p *CompactProtocol) WriteAnyWithDesc(desc *TypeDescriptor, val interface{}, cast bool, disallowUnknown bool, useFieldName bool) error {
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
			return nil
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
			return nil
		}
	default:
		return errUnsupportedType
	}
}

// TODO: technically impossible to implement EncodeXXX since CompactProtocol is stateful.
// func (CompactProtocol) EncodeBool(b []byte, v )

// EncodeByte encodes a byte value.
func (p CompactProtocol) EncodeByte(b []byte, v byte) {
	b[0] = byte(v)
}

// EncodeI16 encodes a int16 value.
func (p CompactProtocol) EncodeI16(b []byte, v int16) (written int) {
	written, _ = p.writeVarint32Internal(b, p.int32ToZigzag(int32(v)))
	return
}

// EncodeI32 encodes a int32 value.
func (p CompactProtocol) EncodeI32(b []byte, v int32) (written int) {
	written, _ = p.writeVarint32Internal(b, p.int32ToZigzag(v))
	return
}

// EncodeI64 encodes a int64 value.
func (p CompactProtocol) EncodeI64(b []byte, v int64) (written int) {
	written, _ = p.writeVarint64Internal(b, v)
	return
}

// EncodeDouble encodes a double value.
func (p CompactProtocol) EncodeDouble(b []byte, v float64) {
	p.Buf = b
	p.WriteDouble(v)
}

// EncodeString encodes a string value.
func (p CompactProtocol) EncodeString(b []byte, v string) {
	p.Buf = b
	p.WriteString(v)
}

// EncodeBinary encodes a binary value.
func (p CompactProtocol) EncodeBinary(b []byte, v []byte) {
	p.Buf = b
	p.WriteBinary(v)
}

func (p CompactProtocol) EncodeFieldBegin(b []byte, t Type, id FieldID) {
	p.Buf = b
	p.WriteFieldBegin("", t, id)
}

// TODO: technically impossible to implement EncodeXXX since CompactProtocol is stateful.
// func (CompactProtocol) DecodeBool(b []byte, v )

func (CompactProtocol) DecodeByte(b []byte) byte {
	return b[0]
}

func (p CompactProtocol) DecodeI16(b []byte) int16 {
	p.Buf = b
	value, _ := p.ReadI16()
	return value
}

func (p CompactProtocol) DecodeI32(b []byte) int32 {
	p.Buf = b
	value, _ := p.ReadI32()
	return value
}

func (p CompactProtocol) DecodeI64(b []byte) int64 {
	p.Buf = b
	value, _ := p.ReadI64()
	return value
}

func (p CompactProtocol) DecodeDouble(b []byte) float64 {
	p.Buf = b
	value, _ := p.ReadDouble()
	return value
}

func (p CompactProtocol) DecodeString(b []byte) string {
	p.Buf = b
	value, _ := p.ReadString(false)
	return value
}

func (p CompactProtocol) DecodeBytes(b []byte) []byte {
	p.Buf = b
	value, _ := p.ReadBinary(false)
	return value
}

/**
 * Misc
 */

var compactTypeSize = [256]int{
	STOP:   -1,
	VOID:   -1,
	BOOL:   -1,
	I08:    1,
	I16:    -1,
	I32:    -1,
	I64:    -1,
	DOUBLE: 8,
	STRING: -1,
	STRUCT: -1,
	MAP:    -1,
	SET:    -1,
	LIST:   -1,
	UTF8:   -1,
	UTF16:  -1,
}

// TypeSize returns the size of the given type.
// -1 means variable size (LIST, SET, MAP, STRING)
// 0 means unknown type
func TypeSizeCompact(t Type) int {
	return compactTypeSize[t]
}

// SkipGo skips over the value for the given type using Go implementation.
func (p *CompactProtocol) SkipGo(fieldType Type, maxDepth int) (err error) {
	if maxDepth <= 0 {
		return errExceedDepthLimit
	}
	switch fieldType {
	case BOOL:
		_, err = p.ReadBool()
		return
	case BYTE:
		_, err = p.ReadByte()
		return
	case I16:
		_, err = p.ReadI16()
		return
	case I32:
		_, err = p.ReadI32()
		return
	case I64:
		_, err = p.ReadI64()
		return
	case DOUBLE:
		_, err = p.ReadDouble()
		return
	case STRING:
		_, err = p.ReadString(false)
		return
	case STRUCT:
		// if _, err = p.ReadStructBegin(); err != nil {
		// 	return err
		// }
		for {
			_, typeId, _, _ := p.ReadFieldBegin()
			if typeId == STOP {
				break
			}
			//fastpath
			if n := typeSize[typeId]; n > 0 {
				p.ReadOff += n
				if p.ReadOff > len(p.Buf) {
					return io.EOF
				}
				continue
			}
			err := p.SkipGo(typeId, maxDepth-1)
			if err != nil {
				return err
			}
			p.ReadFieldEnd()
		}
		return p.ReadStructEnd()
	case MAP:
		keyType, valueType, size, err := p.ReadMapBegin()
		if err != nil {
			return err
		}
		//fastpath
		if k, v := compactTypeSize[keyType], compactTypeSize[valueType]; k > 0 && v > 0 {
			p.ReadOff += (k + v) * size
			if p.ReadOff > len(p.Buf) {
				return io.EOF
			}
		} else {
			if size > len(p.Buf)-p.ReadOff {
				return errInvalidDataSize
			}
			for i := 0; i < size; i++ {
				err := p.SkipGo(keyType, maxDepth-1)
				if err != nil {
					return err
				}
				err = p.SkipGo(valueType, maxDepth-1)
				if err != nil {
					return err
				}
			}
		}
		return p.ReadMapEnd()
	case SET, LIST:
		elemType, size, err := p.ReadListBegin()
		if err != nil {
			return err
		}
		//fastpath
		if v := compactTypeSize[elemType]; v > 0 {
			p.ReadOff += v * size
			if p.ReadOff > len(p.Buf) {
				return io.EOF
			}
		} else {
			if size > len(p.Buf)-p.ReadOff {
				return errInvalidDataSize
			}
			for i := 0; i < size; i++ {
				err := p.SkipGo(elemType, maxDepth-1)
				if err != nil {
					return err
				}
			}
		}
		return p.ReadListEnd()
	default:
		return
	}
}

// next ...
func (p *CompactProtocol) next(size int) ([]byte, error) {
	if size <= 0 {
		panic(errors.New("invalid size"))
	}

	l := len(p.Buf)
	d := p.ReadOff + size
	if d > l {
		return nil, io.EOF
	}

	ret := (p.Buf)[p.ReadOff:d]
	p.ReadOff = d
	return ret, nil
}

// malloc ...
func (p *CompactProtocol) malloc(size int) ([]byte, error) {
	if size <= 0 {
		panic(errors.New("invalid dize"))
	}

	l := len(p.Buf)
	c := cap(p.Buf)
	d := l + size

	if d > c {
		c *= c >> growBufferFactor
		if d > c {
			c = d * 2
		}
		buf := rt.Growslice(byteType, *(*rt.GoSlice)(unsafe.Pointer(&p.Buf)), c)
		p.Buf = *(*[]byte)(unsafe.Pointer(&buf))
	}
	p.Buf = (p.Buf)[:d]

	return (p.Buf)[l:d], nil
}

func (p *CompactProtocol) getCompactType(typeID Type) tcompactType {
	if int(typeID) >= len(ttypeToCompactTypeTable) {
		return COMPACT_STOP
	}
	return ttypeToCompactTypeTable[typeID]
}
