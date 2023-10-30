package generic

import (
	"fmt"
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
	"github.com/cloudwego/dynamicgo/proto/protowire"
)

type Node struct {
	t    proto.Type
	et   proto.Type
	kt   proto.Type
	v    unsafe.Pointer
	l    int
	size int // only for MAP/LIST element counts
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

// slice returns a new node which is a slice of a simple node
func (self Node) slice(s int, e int, t proto.Type) Node {
	ret := Node{
		t: t,
		l: (e - s),
		v: rt.AddPtr(self.v, uintptr(s)),
	}
	return ret
}

// sliceComplex returns a new node which is a slice of a Complex(can deal with all the type) node
func (self Node) sliceComplex(s int, e int, t proto.Type, kt proto.Type, et proto.Type, size int) Node {
	ret := Node{
		t: t,
		l: (e - s),
		v: rt.AddPtr(self.v, uintptr(s)),
	}
	if t == proto.LIST {
		ret.et = et
		ret.size = size
	} else if t == proto.MAP {
		ret.kt = kt
		ret.et = et
		ret.size = size
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
	return ret
}

func NewNodeBool(val bool) Node {
	buf := make([]byte, 0, 1)
	buf = protowire.BinaryEncoder{}.EncodeBool(buf, val)
	return NewNode(proto.BOOL, buf)
}

func NewNodeByte(val byte) Node {
	buf := make([]byte, 0, 1)
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

// NewComplexNode can deal with all the types
func NewComplexNode(t proto.Type, et proto.Type, kt proto.Type, src []byte) (ret Node) {
	if !t.Valid() {
		panic("invalid node type")
	}
	ret = Node{
		t: t,
		l: (len(src)),
		v: rt.GetBytePtr(src),
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

	return
}

func (self Node) Children(out *[]PathNode, recurse bool, opts *Options, desc *proto.MessageDescriptor) (err error) {
	if self.Error() != "" {
		return self
	}

	if out == nil {
		panic("out is nil")
	}
	// NOTICE: since we only use out for memory reuse, we always reset it to zero.
	p := binary.BinaryProtocol{
		Buf: self.raw(),
	}

	tree := PathNode{
		Node: self,
		Next: (*out)[:0], // NOTICE: we reset it to zero.
	}
	rootDesc := (*desc).(proto.Descriptor)
	err = tree.scanChildren(&p, recurse, opts, &rootDesc, len(p.Buf))
	if err == nil {
		*out = tree.Next
	}
	return err
}

// may have error in deal with byteLength
func (self *Node) replace(o Node, n Node) error {
	// mush ensure target value is same type as source value
	if o.t != n.t {
		return wrapError(meta.ErrDismatchType, fmt.Sprintf("type mismatch: %s != %s", o.t, n.t), nil)
	}
	// export target node to bytes
	pat := n.raw()

	// divide self's buffer into three slices:
	// 1. self's slice before the target node
	// 2. target node's slice
	// 3. self's slice after the target node
	s1 := rt.AddPtr(o.v, uintptr(o.l))
	l0 := int(uintptr(o.v) - uintptr(self.v))
	l1 := len(pat)
	l2 := int(uintptr(self.v) + uintptr(self.l) - uintptr(s1))

	// copy three slices into new buffer
	buf := make([]byte, l0+l1+l2)
	copy(buf[:l0], rt.BytesFrom(self.v, l0, l0))
	copy(buf[l0:l0+l1], pat)
	copy(buf[l0+l1:l0+l1+l2], rt.BytesFrom(s1, l2, l2))

	// replace self's entire buffer
	self.v = rt.GetBytePtr(buf)
	self.l = int(len(buf))
	return nil
}

// have problem in deal with byteLength
func (o *Node) setNotFound(path Path, n *Node, desc *proto.FieldDescriptor) error {
	switch o.kt {
	case proto.MESSAGE:
		tag := path.ToRaw(n.t)
		src := n.raw()
		buf := make([]byte, 0, len(tag)+len(src))
		buf = append(buf, tag...)
		buf = append(buf, src...)
		n.l = len(buf)
		n.v = rt.GetBytePtr(buf)
	case proto.LIST:
		// unpacked need write tag, packed needn't
		if (*desc).IsPacked() == false {
			fdNum := (*desc).Number()
			tag := protowire.AppendVarint(nil, uint64(fdNum)<<3|uint64(proto.BytesType))
			src := n.raw()
			buf := make([]byte, 0, len(tag)+len(src))
			buf = append(buf, tag...)
			buf = append(buf, src...)
			n.l = len(buf)
			n.v = rt.GetBytePtr(buf)
		}
	case proto.MAP:
		// pair tag
		fdNum := (*desc).Number()
		pairTag := protowire.AppendVarint(nil, uint64(fdNum)<<3|uint64(proto.BytesType))
		buf := path.ToRaw(n.t) // keytag + key
		valueKind := (*desc).MapValue().Kind()
		valueTag := uint64(1)<<3 | uint64(proto.Kind2Wire[valueKind])
		buf = protowire.BinaryEncoder{}.EncodeUint64(buf, valueTag)                  // + value tag
		src := n.raw()                                                               // + value
		buf = append(buf, src...)                                                    // key + value
		pairbuf := protowire.BinaryEncoder{}.EncodeUint64(pairTag, uint64(len(buf))) // + pairlen
		pairbuf = append(pairbuf, buf...)                                            // pair tag + pairlen + key + value
		n.l = len(pairbuf)
		n.v = rt.GetBytePtr(pairbuf)
	default:
		return wrapError(meta.ErrDismatchType, "simple type node shouldn't have child", nil)
	}
	o.t = n.t
	o.l = 0
	return nil
}


func (self *Node) replaceMany(ps *pnSlice) error {
	var buf []byte
	// Sort pathes by original value address
	ps.Sort()

	// sequentially set new values into buffer according to sorted pathes
	buf = make([]byte, 0, self.l)
	last := self.v
	for i := 0; i < len(ps.a); i++ {
		lastLen := rt.PtrOffset(ps.a[i].Node.v, last)
		// copy last slice from original buffer
		buf = append(buf, rt.BytesFrom(last, lastLen, lastLen)...)
		// copy new value's buffer into buffer
		buf = append(buf, ps.b[i].Node.raw()...)
		// update last index
		last = ps.a[i].offset()
	}
	if tail := self.offset(); uintptr(last) < uintptr(tail) {
		// copy last slice from original buffer
		buf = append(buf, rt.BytesFrom(last, rt.PtrOffset(tail, last), rt.PtrOffset(tail, last))...)
	}

	self.v = rt.GetBytePtr(buf)
	self.l = int(len(buf))
	return nil
}

func (self Node) Index(idx int) (v Node) {
	if err := self.should("Index", proto.LIST); err != "" {
		return errNode(meta.ErrUnsupportedType, err, nil)
	}

	var s, e int
	it := self.iterElems()
	if it.Err != nil {
		return errNode(meta.ErrRead, "", it.Err)
	}
	isPacked := it.IsPacked()

	// size = 0 maybe list node is in lazyload mode, need to check idx
	if it.size > 0 && idx >= it.size {
		return errNode(meta.ErrInvalidParam, fmt.Sprintf("index %d exceeds list/set bound", idx), nil)
	}

	// read packed list tag and bytelen
	if isPacked {
		if _, _, _, err := it.p.ConsumeTag(); err != nil {
			return errNode(meta.ErrRead, "", err)
		}
		if _, err := it.p.ReadLength(); err != nil {
			return errNode(meta.ErrRead, "", err)
		}
	}

	for j := 0; it.HasNext() && j < idx; j++ {
		it.Next(UseNativeSkipForGet)
	}

	if it.Err != nil {
		return errNode(meta.ErrRead, "", it.Err)
	}

	// when lazy load, size = 0, use it.k which is counted after it.Next() to check valid idx
	if idx > it.k {
		return errNode(meta.ErrInvalidParam, fmt.Sprintf("index '%d' is out of range", idx), nil)
	}

	s, e = it.Next(UseNativeSkipForGet)
	v = self.slice(s, e, self.et)
	return v
}

func (self Node) GetByStr(key string) (v Node) {
	if err := self.should("Get", proto.MAP); err != "" {
		return errNode(meta.ErrUnsupportedType, err, nil)
	}

	if self.kt != proto.STRING {
		return errNode(meta.ErrUnsupportedType, "key type is not string", nil)
	}

	it := self.iterPairs()
	if it.Err != nil {
		return errNode(meta.ErrRead, "", it.Err)
	}

	for it.HasNext() {
		_, s, ss, e := it.NextStr(UseNativeSkipForGet)
		if it.Err != nil {
			v = errNode(meta.ErrRead, "", it.Err)
			goto ret
		}

		if s == key {
			v = self.slice(ss, e, self.et)
			goto ret
		}
	}
	v = errNode(meta.ErrNotFound, fmt.Sprintf("key '%s' is not found in this value", key), errNotFound)
ret:
	return
}


func (self Node) GetByInt(key int) (v Node) {
	if err := self.should("Get", proto.MAP); err != "" {
		return errNode(meta.ErrUnsupportedType, err, nil)
	}

	if !self.kt.IsInt() {
		return errNode(meta.ErrUnsupportedType, "key type is not int", nil)
	}

	it := self.iterPairs()
	if it.Err != nil {
		return errNode(meta.ErrRead, "", it.Err)
	}

	for it.HasNext() {
		_, i, ss, e := it.NextInt(UseNativeSkipForGet)
		if it.Err != nil {
			v = errNode(meta.ErrRead, "", it.Err)
			goto ret
		}
		if i == key {
			v = self.slice(ss, e, self.et)
			goto ret
		}
	}
	v = errNode(meta.ErrNotFound, fmt.Sprintf("key '%d' is not found in this value", key), nil)
ret:
	return
}