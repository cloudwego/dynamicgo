package generic

import (
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/protowire"
)

type Node struct {
	t  proto.Type
	et proto.Type
	kt proto.Type
	v  unsafe.Pointer
	l  int
}

// Fork forks the node to a new node, copy underlying data as well
func (self Node) Fork() Node {
	ret := self
	buf := make([]byte, self.l, self.l)
	copy(buf, rt.BytesFrom(self.v, self.l, self.l))
	ret.v = rt.GetBytePtr(buf)
	return ret
}

// Type returns the thrift type of the node
func (self Node) Type() proto.Type {
	return self.t
}

// ElemType returns the thrift type of a LIST/SET/MAP node's element
func (self Node) ElemType() proto.Type {
	return self.et
}

// KeyType returns the thrift type of a MAP node's key
func (self Node) KeyType() proto.Type {
	return self.kt
}

func (self Node) slice(s int, e int, t proto.Type) Node {
	ret := Node{
		t: t,
		l: (e - s),
		v: rt.AddPtr(self.v, uintptr(s)),
	}
	if t == proto.LIST {
		ret.et = *(*proto.Type)(unsafe.Pointer(ret.v))
	} else if t == proto.MAP {
		ret.kt = *(*proto.Type)(unsafe.Pointer(ret.v))
		ret.et = *(*proto.Type)(rt.AddPtr(ret.v, uintptr(1)))
	}
	return ret
}

func (self Node) offset() unsafe.Pointer {
	return rt.AddPtr(self.v, uintptr(self.l))
}

// NewNode method: creates a new node from a byte slice
func NewNode(t proto.Type, src []byte) Node {
	ret := Node{
		t: t,
		l: (len(src)),
		v: rt.GetBytePtr(src),
	}
	if t == proto.LIST {
		ret.et = *(*proto.Type)(unsafe.Pointer(ret.v))
	} else if t == proto.MAP {
		ret.kt = *(*proto.Type)(unsafe.Pointer(ret.v))
		ret.et = *(*proto.Type)(rt.AddPtr(ret.v, uintptr(1)))
	}
	return ret
}

func NewNodeBool(val bool) Node {
	buf := make([]byte, 0, 1)
	buf = protowire.BinaryEncoder{}.EncodeBool(buf, val)
	return NewNode(proto.BOOL, buf)
}

func NewNodeByte(val byte) Node {
	buf := make([]byte, 0,1)
	buf = protowire.BinaryEncoder{}.EncodeByte(buf, val)
	return NewNode(proto.BYTE, buf)
}

func NewNodeEnum(val int32) Node {
	buf := make([]byte, 0, 4)
	buf = protowire.BinaryEncoder{}.EncodeEnum(buf, val)
	return NewNode(proto.ENUM, buf)
}

func NewNodeInt32(val int32) Node {
	buf := make([]byte, 0, 4)
	buf = protowire.BinaryEncoder{}.EncodeInt32(buf, val)
	return NewNode(proto.INT32, buf)
}

func NewNodeSint32(val int32) Node {
	buf := make([]byte, 0, 4)
	buf = protowire.BinaryEncoder{}.EncodeSint32(buf, val)
	return NewNode(proto.SINT32, buf)
}

func NewNodeUint32(val uint32) Node {
	buf := make([]byte, 0, 4)
	buf = protowire.BinaryEncoder{}.EncodeUint32(buf, val)
	return NewNode(proto.UINT32, buf)
}

func NewNodeFixed32(val uint32) Node {
	buf := make([]byte, 0, 4)
	buf = protowire.BinaryEncoder{}.EncodeFixed32(buf, val)
	return NewNode(proto.FIX32, buf)
}

func NewNodeSfixed32(val int32) Node {
	buf := make([]byte, 0, 4)
	buf = protowire.BinaryEncoder{}.EncodeSfixed32(buf, val)
	return NewNode(proto.SFIX32, buf)
}

func NewNodeInt64(val int64) Node {
	buf := make([]byte, 0, 8)
	buf = protowire.BinaryEncoder{}.EncodeInt64(buf, val)
	return NewNode(proto.INT64, buf)
}

func NewNodeSint64(val int64) Node {
	buf := make([]byte, 0, 8)
	buf = protowire.BinaryEncoder{}.EncodeSint64(buf, val)
	return NewNode(proto.SINT64, buf)
}

func NewNodeUint64(val uint64) Node {
	buf := make([]byte, 0, 8)
	buf = protowire.BinaryEncoder{}.EncodeUint64(buf, val)
	return NewNode(proto.UINT64, buf)
}

func NewNodeFixed64(val uint64) Node {
	buf := make([]byte, 0, 8)
	buf = protowire.BinaryEncoder{}.EncodeFixed64(buf, val)
	return NewNode(proto.FIX64, buf)
}

func NewNodeSfixed64(val int64) Node {
	buf := make([]byte, 0, 8)
	buf = protowire.BinaryEncoder{}.EncodeSfixed64(buf, val)
	return NewNode(proto.SFIX64, buf)
}

func NewNodeFloat(val float32) Node {
	buf := make([]byte, 0, 4)
	buf = protowire.BinaryEncoder{}.EncodeFloat32(buf, val)
	return NewNode(proto.FLOAT, buf)
}

func NewNodeDouble(val float64) Node {
	buf := make([]byte, 0, 8)
	buf = protowire.BinaryEncoder{}.EncodeDouble(buf, val)
	return NewNode(proto.DOUBLE, buf)
}

func NewNodeString(val string) Node {
	buf := make([]byte, 0, len(val)+4)
	buf = protowire.BinaryEncoder{}.EncodeString(buf, val)
	return NewNode(proto.STRING, buf)
}

func NewNodeBytes(val []byte) Node {
	buf := make([]byte, 0, len(val)+4)
	buf = protowire.BinaryEncoder{}.EncodeBytes(buf, val)
	return NewNode(proto.BYTE, buf)
}

func NewComplexNode(t proto.Type, et proto.Type, kt proto.Type) (ret Node){
	if !t.Valid() {
		panic("invalid node type")
	}
	switch t {
	case proto.LIST:
		if !et.Valid() {
			panic("invalid element type")
		}
		ret.et = et
	case proto.MAP:
		if !et.Valid() {
			panic("invalid element type")
		}
		if !kt.Valid() {
			panic("invalid key type")
		}
		ret.et = et
		ret.kt = kt
	}
	ret.t = t
	return
}
