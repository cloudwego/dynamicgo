package base

import (
	"github.com/cloudwego/dynamicgo/proto"
)

var typeSize = [256]int{
	proto.NIL:	-1, // nil
	proto.DOUBLE:	-1,
	proto.FLOAT:  -1,
	proto.INT64:  -1,
	proto.UINT64: -1,
	proto.INT32:  -1,
	proto.FIX64:  -1,
	proto.FIX32:  -1,
	proto.BOOL:   -1,
	proto.STRING: -1,
	proto.GROUP:  -1, // deprecated
	proto.MESSAGE: -1,
	proto.BYTE:   -1,
	proto.UINT32: -1,
	proto.ENUM:   -1,
	proto.SFIX32: -1,
	proto.SFIX64: -1,
	proto.SINT32: -1,
	proto.SINT64: -1,
	proto.LIST:   -1,
	proto.MAP:	-1,
}

const MaxSkipDepth = 1023

func (p *BinaryProtocol) Skip(fieldType proto.Type, useNative bool) (err error) {
	return p.SkipGo(fieldType, MaxSkipDepth)
}

// TypeSize returns the size of the given type.
// -1 means variable size (LIST, SET, MAP, STRING)
// 0 means unknown type
func TypeSize(t proto.Type) int {
	return typeSize[t]
}

// SkipGo skips over the value for the given type using Go implementation.
func (p *BinaryProtocol) SkipGo(fieldType proto.Type, maxDepth int) (err error) {
	if maxDepth <= 0 {
		return errExceedDepthLimit
	}
	switch fieldType {
	case proto.DOUBLE:
		_, err = p.ReadDouble()
		return
	case proto.FLOAT:
		_, err = p.ReadFloat()
		return
	case proto.INT64:
		_, err = p.ReadI64()
		return
	case proto.INT32:
		_, err = p.ReadI32()
		return
	case proto.FIX64:
		_, err = p.ReadFixed64()
		return
	case proto.FIX32:
		_, err = p.ReadFixed32()
		return
	case proto.BOOL:
		_, err = p.ReadBool()
		return
	case proto.STRING:
		_, err = p.ReadString(false)
		return
	case proto.BYTE:
		_, err = p.ReadBytes()
		return
	case proto.UINT32:
		_, err = p.ReadUint32()
		return
	case proto.ENUM:
		_, err = p.ReadI32()
		return
	case proto.UINT64:
		_, err = p.ReadUint64()
		return
	case proto.SFIX32:
		_, err = p.ReadSfixed32()
		return
	case proto.SFIX64:
		_, err = p.ReadSfixed64()
		return
	case proto.SINT32:
		_, err = p.ReadSint32()
		return
	case proto.SINT64:
		_, err = p.ReadSint64()
		return
	case proto.LIST:
		return err
	case proto.MAP:
		return err
	case proto.MESSAGE:
		return err
	default:
		return err
	}
}