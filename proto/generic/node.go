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
	t  proto.Type
	et proto.Type
	kt proto.Type
	v  unsafe.Pointer
	l  int
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
func NewComplexNode(t proto.Type, et proto.Type, kt proto.Type, src []byte) (ret Node){
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
		return meta.NewError(meta.ErrDismatchType, fmt.Sprintf("type mismatch: %s != %s", o.t, n.t), nil)
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
func (self *Node) deleteChild(path Path) Node {
	p := binary.BinaryProtocol{}
	p.Buf = self.raw()
	var start, end int
	var tt proto.Type // Need TODO: check if this is necessary
	exist := false
	switch self.t {
	case proto.MESSAGE:
		p.ConsumeTag()
		len,_ := p.ReadLength()
		if path.Type() != PathFieldId {
			return errNode(meta.ErrDismatchType, "", nil)
		}
		messageStart := p.Read
		id := path.id()
		for p.Read < messageStart + len {
			start = p.Read
			fieldNumber, wireType, _, tagErr := p.ConsumeTag()
			if tagErr != nil {
				return errNode(meta.ErrRead, "", tagErr)
			}
			if err := p.Skip(wireType, false); err != nil {
				return errNode(meta.ErrRead, "", err)
			}
			end = p.Read
			if id == fieldNumber {
				exist = true
				break
			}
		}
		if !exist {
			return errNotFound
		}
	case proto.LIST:
		fieldNumber, wireType, _, tagErr := p.ConsumeTagWithoutMove()
		if tagErr != nil {
			return errNode(meta.ErrRead, "", tagErr)
		}
		start := p.Read
		listIndex := 0		
		if path.Type() != PathIndex {
			return errNode(meta.ErrDismatchType, "", nil)
		}
		idx := path.int()
		// packed
		if wireType == proto.BytesType {
			_, _, _, tagErr := p.ConsumeTag()
			if tagErr != nil {
				return errNode(meta.ErrRead, "", tagErr)
			}

			listLen, lengthErr := p.ReadLength()
			if lengthErr != nil {
				return errNode(meta.ErrRead, "", lengthErr)
			}

			for p.Read < start + listLen && listIndex < idx {
				start = p.Read
				if err := p.Skip(wireType, false); err != nil {
					return errNode(meta.ErrRead, "", err)
				}
				end = p.Read
				listIndex++
			}
		} else {
			for p.Read < len(p.Buf) && listIndex < idx{
				start = p.Read
				itemNumber, _, tagLen, _ := p.ConsumeTagWithoutMove()
				if itemNumber != fieldNumber {
					break
				}
				p.Read += tagLen
				if err := p.Skip(wireType, false); err != nil {
					return errNode(meta.ErrRead, "", err)
				}
				end = p.Read
				listIndex++
			}
		}

		if listIndex > idx {
			return errNotFound
		}
	case proto.MAP:
		pairNumber, _, _, _ := p.ConsumeTagWithoutMove()
		if self.kt == proto.STRING {
			key := path.str()
			for p.Read < len(p.Buf) {
				start = p.Read
				itemNumber, _, pairLen, _ := p.ConsumeTagWithoutMove()
				if itemNumber != pairNumber {
					break
				}
				p.Read += pairLen
				p.ReadLength()
				// key
				p.ConsumeTag()
				k, _ := p.ReadString(false)
				// value
				_, valueWire, _, _ := p.ConsumeTag()
				if err := p.Skip(valueWire, false); err != nil {
					return errNode(meta.ErrRead, "", err)
				}
				end = p.Read
				if k == key {
					exist = true
					break
				}
			}
		} else if self.kt.IsInt() {
			key := path.Int()
			pairNumber, _, _, _ := p.ConsumeTagWithoutMove()
			for p.Read < len(p.Buf) {
				start = p.Read
				itemNumber, _, pairLen, _ := p.ConsumeTagWithoutMove()
				if itemNumber != pairNumber {
					break
				}
				p.Read += pairLen
				p.ReadLength()

				// key
				p.ConsumeTag()
				k, _ := p.ReadInt(self.kt)
				//value
				_, valueWire, _, _ := p.ConsumeTag()
				if err := p.Skip(valueWire, false); err != nil {
					return errNode(meta.ErrRead, "", err)
				}
				end = p.Read
				if k == key {
					exist = true
					break
				}
			}
		}
		if !exist {
			return errNotFound
		}
	default:
		return errNotFound
	}
	return Node{
		t: tt,
		v: rt.AddPtr(self.v, uintptr(start)),
		l: end - start,
	}
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
			tag := protowire.AppendVarint(nil, uint64(fdNum) << 3 | uint64(proto.BytesType))
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
		pairTag := protowire.AppendVarint(nil, uint64(fdNum) << 3 | uint64(proto.BytesType))
		buf := path.ToRaw(n.t) // keytag + key
		valueKind := (*desc).MapValue().Kind()
		valueTag := uint64(1) << 3 | uint64(proto.Kind2Wire[valueKind])
		buf = protowire.BinaryEncoder{}.EncodeUint64(buf, valueTag) // + value tag
		src := n.raw() // + value
		buf = append(buf, src...) // key + value
		pairbuf := protowire.BinaryEncoder{}.EncodeUint64(pairTag, uint64(len(buf))) // + pairlen
		pairbuf = append(pairbuf, buf...) // pair tag + pairlen + key + value
		n.l = len(pairbuf)
		n.v = rt.GetBytePtr(pairbuf)
	default:
		return errNode(meta.ErrDismatchType, "simple type node shouldn't have child", nil)
	}
	o.t = n.t
	o.l = 0
	return nil
}