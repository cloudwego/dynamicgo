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

// Type returns the proto type of the node
func (self Node) Type() proto.Type {
	return self.t
}

// ElemType returns the thrift type of a LIST/MAP node's element
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

func (self Node) sliceNodeWithDesc(s int, e int, desc *proto.TypeDescriptor) Node {
	t := desc.Type()
	ret := Node{
		t: t,
		l: (e - s),
		v: rt.AddPtr(self.v, uintptr(s)),
	}
	if t == proto.LIST {
		ret.et = desc.Elem().Type()
	} else if t == proto.MAP {
		ret.kt = desc.Key().Type()
		ret.et = desc.Elem().Type()
	}
	return ret
}

func (self Node) offset() uintptr {
	return uintptr(self.v) + uintptr(self.l)
}

func (self *Node) SetElemType(et proto.Type) {
	self.et = et
}

func (self *Node) SetKeyType(kt proto.Type) {
	self.kt = kt
}

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
	l0 := int(uintptr(o.v) - uintptr(self.v))
	l1 := len(pat)
	l2 := int(self.offset() - uintptr(o.offset()))

	// copy three slices into new buffer
	buf := make([]byte, l0+l1+l2)
	if l0 > 0 {
		copy(buf[:l0], rt.BytesFrom(self.v, l0, l0))
	}
	copy(buf[l0:l0+l1], pat)
	if l2 > 0 {
		copy(buf[l0+l1:l0+l1+l2], rt.BytesFrom(rt.AddPtr(o.v, uintptr(o.l)), l2, l2))
	}

	// replace self's entire buffer
	self.v = rt.GetBytePtr(buf)
	self.l = int(len(buf))
	return nil
}

// have problem in deal with byteLength
func (o *Node) setNotFound(path Path, n *Node, desc *proto.TypeDescriptor) error {
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
		if desc.IsPacked() == false {
			fdNum := desc.BaseId()
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
		fdNum := desc.BaseId()
		pairTag := protowire.AppendVarint(nil, uint64(fdNum)<<3|uint64(proto.BytesType))
		buf := path.ToRaw(n.t) // keytag + key
		valueWireType := desc.Elem().WireType()
		valueTag := uint64(1)<<3 | uint64(valueWireType)
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
	offset := int(0)
	for i := 0; i < len(ps.a); i++ {
		// copy (a[i-1]tail, a[i]head) into buffer
		if offset < self.l {
			lastp := rt.AddPtr(self.v, uintptr(offset))
			lastLen := rt.PtrOffset(uintptr(ps.a[i].Node.v), uintptr(lastp))
			buf = append(buf, rt.BytesFrom(lastp, lastLen, lastLen)...)
		}
		// copy new value's buffer into buffer
		buf = append(buf, ps.b[i].Node.raw()...)
		// update last index
		offset = rt.PtrOffset(ps.a[i].offset(), uintptr(self.v))
	}
	if offset < self.l {
		// copy last slice from original buffer
		l := self.l - offset
		buf = append(buf, rt.BytesFrom(rt.AddPtr(self.v, uintptr(offset)), l, l)...)
	}

	self.v = rt.GetBytePtr(buf)
	self.l = int(len(buf))
	return nil
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

// NewComplexNode can deal with all the types, only if the src is a valid byte slice
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

// returns all the children of a node, when recurse is false, it switch to lazyload mode, only direct children are returned
func (self Node) Children(out *[]PathNode, recurse bool, opts *Options, desc *proto.TypeDescriptor) (err error) {
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
	err = tree.scanChildren(&p, recurse, opts, desc, len(p.Buf))
	if err == nil {
		*out = tree.Next
	}
	return err
}

// Get idx element of a LIST node
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

	// size = 0 maybe list node is in lazyload mode, need to check idx with it.k
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

// Get string key of a MAP node
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

// Get int key of a MAP node
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

func (self Node) Field(id proto.FieldNumber, rootLayer bool, msgDesc *proto.MessageDescriptor) (v Node, f *proto.FieldDescriptor) {
	if err := self.should("Field", proto.MESSAGE); err != "" {
		return errNode(meta.ErrUnsupportedType, err, nil), nil
	}

	fd := msgDesc.ByNumber(id)

	if fd == nil {
		return errNode(meta.ErrUnknownField, fmt.Sprintf("field '%d' is not defined in IDL", id), nil), nil
	}

	it := self.iterFields()

	if it.Err != nil {
		return errNode(meta.ErrRead, "", it.Err), nil
	}

	if !rootLayer {
		if _, err := it.p.ReadLength(); err != nil {
			return errNode(meta.ErrRead, "", err), nil
		}

		if it.Err != nil {
			return errNode(meta.ErrRead, "", it.Err), nil
		}
	}

	for it.HasNext() {
		i, wt, s, e, tagPos := it.Next(UseNativeSkipForGet)
		if i == fd.Number() {
			typDesc := fd.Type()
			if typDesc.IsMap() || typDesc.IsList() {
				it.p.Read = tagPos
				if _, err := it.p.SkipAllElements(i, typDesc.IsPacked()); err != nil {
					return errNode(meta.ErrRead, "SkipAllElements in LIST/MAP failed", err), nil
				}
				s = tagPos
				e = it.p.Read

				v = self.sliceNodeWithDesc(s, e, typDesc)
				goto ret
			}

			t := proto.Kind2Wire[fd.Kind()]
			if wt != t {
				v = errNode(meta.ErrDismatchType, fmt.Sprintf("field '%s' expects type %s, buf got type %s", fd.Name(), t, wt), nil)
				goto ret
			}
			v = self.sliceNodeWithDesc(s, e, typDesc)
			goto ret
		} else if it.Err != nil {
			v = errNode(meta.ErrRead, "", it.Err)
			goto ret
		}
	}

	v = errNode(meta.ErrNotFound, fmt.Sprintf("field '%d' is not found in this value", id), errNotFound)
ret:
	return v, fd
}

func (self Node) Fields(ids []PathNode, rootLayer bool, msgDesc *proto.MessageDescriptor, opts *Options) error {
	if err := self.should("Fields", proto.MESSAGE); err != "" {
		return errNode(meta.ErrUnsupportedType, err, nil)
	}

	if len(ids) == 0 {
		return nil
	}

	if opts.ClearDirtyValues {
		for i := range ids {
			ids[i].Node = Node{}
		}
	}

	it := self.iterFields()
	if it.Err != nil {
		return errNode(meta.ErrRead, "", it.Err)
	}

	if !rootLayer {
		if _, err := it.p.ReadLength(); err != nil {
			return errNode(meta.ErrRead, "", err)
		}

		if it.Err != nil {
			return errNode(meta.ErrRead, "", it.Err)
		}
	}

	need := len(ids)
	for count := 0; it.HasNext() && count < need; {
		var p *PathNode
		i, _, s, e, tagPos := it.Next(UseNativeSkipForGet)
		if it.Err != nil {
			return errNode(meta.ErrRead, "", it.Err)
		}
		f := msgDesc.ByNumber(i)
		typDesc := f.Type()
		if typDesc.IsMap() || typDesc.IsList() {
			it.p.Read = tagPos
			if _, err := it.p.SkipAllElements(i, typDesc.IsPacked()); err != nil {
				return errNode(meta.ErrRead, "SkipAllElements in LIST/MAP failed", err)
			}
			s = tagPos
			e = it.p.Read
		}

		v := self.sliceNodeWithDesc(s, e, typDesc)

		//TODO: use bitmap to avoid repeatedly scan
		for j, id := range ids {
			if id.Path.t == PathFieldId && id.Path.id() == i {
				p = &ids[j]
				count += 1
				break
			}
		}

		if p != nil {
			p.Node = v
		}
	}

	return nil
}

func (self Node) Indexes(ins []PathNode, opts *Options) error {
	if err := self.should("Indexes", proto.LIST); err != "" {
		return errNode(meta.ErrUnsupportedType, err, nil)
	}

	if len(ins) == 0 {
		return nil
	}

	if opts.ClearDirtyValues {
		for i := range ins {
			ins[i].Node = Node{}
		}
	}

	it := self.iterElems()
	if it.Err != nil {
		return errNode(meta.ErrRead, "", it.Err)
	}

	need := len(ins)
	IsPacked := it.IsPacked()

	// read packed list tag and bytelen
	if IsPacked {
		if _, _, _, err := it.p.ConsumeTag(); err != nil {
			return errNode(meta.ErrRead, "", err)
		}
		if _, err := it.p.ReadLength(); err != nil {
			return errNode(meta.ErrRead, "", err)
		}
	}

	for count, i := 0, 0; it.HasNext() && count < need; i++ {
		s, e := it.Next(UseNativeSkipForGet)
		if it.Err != nil {
			return errNode(meta.ErrRead, "", it.Err)
		}

		// check if the index is in the pathes
		var p *PathNode
		for j, id := range ins {
			if id.Path.t != PathIndex {
				continue
			}
			k := id.Path.int()
			if k >= it.size {
				continue
			}
			if k == i {
				p = &ins[j]
				count += 1
				break
			}
		}
		// PathNode is found
		if p != nil {
			p.Node = self.slice(s, e, self.et)
		}
	}
	return nil
}

func (self Node) Gets(keys []PathNode, opts *Options) error {
	if err := self.should("Gets", proto.MAP); err != "" {
		return errNode(meta.ErrUnsupportedType, err, nil)
	}

	if len(keys) == 0 {
		return nil
	}

	if opts.ClearDirtyValues {
		for i := range keys {
			keys[i].Node = Node{}
		}
	}

	et := self.et
	it := self.iterPairs()
	if it.Err != nil {
		return errNode(meta.ErrRead, "", it.Err)
	}

	need := len(keys)
	for count := 0; it.HasNext() && count < need; {
		for j, id := range keys {
			if id.Path.Type() == PathStrKey {
				exp := id.Path.str()
				_, key, s, e := it.NextStr(UseNativeSkipForGet)
				if it.Err != nil {
					return errNode(meta.ErrRead, "", it.Err)
				}
				if key == exp {
					keys[j].Node = self.slice(s, e, et)
					count += 1
					break
				}
			} else if id.Path.Type() == PathIntKey {
				exp := id.Path.int()
				_, key, s, e := it.NextInt(UseNativeSkipForGet)
				if it.Err != nil {
					return errNode(meta.ErrRead, "", it.Err)
				}
				if key == exp {
					keys[j].Node = self.slice(s, e, et)
					count += 1
					break
				}
			}
		}
	}
	return nil
}
