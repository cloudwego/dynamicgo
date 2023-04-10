package thrift

import (
	"context"

	"github.com/apache/thrift/lib/go/thrift"
)

type CompatCompactProtocol struct {
	CompactProtocol
}

var _ thrift.TProtocol = (*CompatCompactProtocol)(nil)

func (cp *CompatCompactProtocol) WriteMessageBegin(name string, typ thrift.TMessageType, seqID int32) error {
	return cp.CompactProtocol.WriteMessageBegin(name, TMessageType(typ), int32(seqID))
}

func (cp *CompatCompactProtocol) WriteFieldBegin(name string, typ thrift.TType, id int16) error {
	return cp.CompactProtocol.WriteFieldBegin(name, Type(typ), FieldID(id))
}

func (cp *CompatCompactProtocol) WriteMapBegin(keyTyp, elemTyp thrift.TType, size int) error {
	return cp.CompactProtocol.WriteMapBegin(Type(keyTyp), Type(elemTyp), size)
}

func (cp *CompatCompactProtocol) WriteListBegin(elemTyp thrift.TType, size int) error {
	return cp.CompactProtocol.WriteListBegin(Type(elemTyp), size)
}

func (cp *CompatCompactProtocol) WriteSetBegin(elemType thrift.TType, size int) error {
	return cp.CompactProtocol.WriteSetBegin(Type(elemType), size)
}

func (cp *CompatCompactProtocol) WriteByte(v int8) error {
	return cp.CompactProtocol.WriteByte(byte(v))
}

/* -------- */

func (cp *CompatCompactProtocol) ReadMessageBegin() (string, thrift.TMessageType, int32, error) {
	name, typeID, seqID, err := cp.CompactProtocol.ReadMessageBegin(true)
	if err != nil {
		return "", thrift.INVALID_TMESSAGE_TYPE, 0, err
	}
	return name, thrift.TMessageType(typeID), seqID, nil
}

func (cp *CompatCompactProtocol) ReadFieldBegin() (string, thrift.TType, int16, error) {
	name, typeID, fieldID, err := cp.CompactProtocol.ReadFieldBegin()
	if err != nil {
		return "", thrift.STOP, 0, err
	}
	return name, thrift.TType(typeID), int16(fieldID), nil
}

func (cp *CompatCompactProtocol) ReadMapBegin() (thrift.TType, thrift.TType, int, error) {
	keyType, elemType, size, err := cp.CompactProtocol.ReadMapBegin()
	if err != nil {
		return thrift.STOP, thrift.STOP, 0, err
	}
	return thrift.TType(keyType), thrift.TType(elemType), size, err
}

func (cp *CompatCompactProtocol) ReadListBegin() (thrift.TType, int, error) {
	elemType, size, err := cp.CompactProtocol.ReadListBegin()
	if err != nil {
		return thrift.STOP, 0, err
	}
	return thrift.TType(elemType), size, nil
}

func (cp *CompatCompactProtocol) ReadSetBegin() (thrift.TType, int, error) {
	elemType, size, err := cp.CompactProtocol.ReadListBegin()
	if err != nil {
		return thrift.STOP, 0, err
	}
	return thrift.TType(elemType), size, nil
}

func (cp *CompatCompactProtocol) ReadByte() (int8, error) {
	return cp.ReadByte()
}

func (cp *CompatCompactProtocol) ReadString() (string, error) {
	return cp.CompactProtocol.ReadString(true)
}

func (cp *CompatCompactProtocol) ReadBinary() ([]byte, error) {
	return cp.CompactProtocol.ReadBinary(true)
}

func (cp *CompatCompactProtocol) Skip(typ thrift.TType) error {
	return cp.CompactProtocol.SkipGo(Type(typ), MaxSkipDepth)
}

func (cp *CompatCompactProtocol) Transport() thrift.TTransport {
	panic("DO NOT CALL")
}

func (cp *CompatCompactProtocol) Flush(ctx context.Context) error {
	return nil
}
