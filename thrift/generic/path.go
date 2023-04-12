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

package generic

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
)

// PathType is the type of path
type PathType uint8

const (
	// PathFieldId represents a field id of STRUCT type
	PathFieldId PathType = 1 + iota

	// PathFieldName represents a field name of STRUCT type
	// NOTICE: it is only supported by Value
	PathFieldName

	// PathIndex represents a index of LIST\SET type
	PathIndex

	// Path represents a string key of MAP type
	PathStrKey

	// Path represents a int key of MAP type
	PathIntKey

	// Path represents a raw-bytes key of MAP type
	// It is usually used for neither-string-nor-integer type key
	PathBinKey
)

// Path represents the relative position of a sub node in a complex parent node
type Path struct {
	t PathType
	v unsafe.Pointer
	l int
}

// Str returns the string value of a PathFieldName\PathStrKey path
func (self Path) Str() string {
	switch self.t {
	case PathFieldName, PathStrKey:
		return self.str()
	default:
		return ""
	}
}

func (self Path) str() string {
	return *(*string)(unsafe.Pointer(&self.v))
}

// Int returns the int value of a PathIndex\PathIntKey path
func (self Path) Int() int {
	switch self.t {
	case PathIndex, PathIntKey:
		return self.int()
	default:
		return -1
	}
}

func (self Path) int() int {
	return self.l
}

// Id returns the field id of a PathFieldId path
func (self Path) Id() thrift.FieldID {
	switch self.t {
	case PathFieldId:
		return self.id()
	default:
		return thrift.FieldID(0)
	}
}

func (self Path) id() thrift.FieldID {
	return thrift.FieldID(self.l)
}

// Bin returns the raw bytes value of a PathBinKey path
func (self Path) Bin() []byte {
	switch self.t {
	case PathBinKey:
		return self.bin()
	default:
		return nil
	}
}

func (self Path) bin() []byte {
	return rt.BytesFrom(self.v, self.l, self.l)
}

// ToRaw converts underlying value to thrift-encoded bytes
func (self Path) ToRaw(t thrift.Type) []byte {
	switch self.t {
	case PathFieldId:
		ret := make([]byte, 3, 3)
		thrift.BinaryEncoding{}.EncodeByte(ret, byte(t))
		thrift.BinaryEncoding{}.EncodeInt16(ret[1:], int16(self.l))
		return ret
	case PathStrKey:
		ret := make([]byte, 4+self.l, 4+self.l)
		thrift.BinaryEncoding{}.EncodeString(ret, self.str())
		return ret
	case PathIntKey:
		switch t {
		case thrift.I08:
			ret := make([]byte, 1, 1)
			thrift.BinaryEncoding{}.EncodeByte(ret, byte(self.l))
			return ret
		case thrift.I16:
			ret := make([]byte, 2, 2)
			thrift.BinaryEncoding{}.EncodeInt16(ret, int16(self.l))
			return ret
		case thrift.I32:
			ret := make([]byte, 4, 4)
			thrift.BinaryEncoding{}.EncodeInt32(ret, int32(self.l))
			return ret
		case thrift.I64:
			ret := make([]byte, 8, 8)
			thrift.BinaryEncoding{}.EncodeInt64(ret, int64(self.l))
			return ret
		default:
			return nil
		}
	case PathBinKey:
		return rt.BytesFrom(self.v, self.l, self.l)
	default:
		return nil
	}
}

// String returns the string representation of a Path
func (self Path) String() string {
	switch self.t {
	case PathFieldId:
		return "FieldId(" + strconv.Itoa(int(self.id())) + ")"
	case PathFieldName:
		return "FieldName(\"" + self.str() + "\")"
	case PathIndex:
		return "Index(" + strconv.Itoa(self.int()) + ")"
	case PathStrKey:
		return "Key(\"" + self.str() + "\")"
	case PathIntKey:
		return "Key(" + strconv.Itoa(self.int()) + ")"
	case PathBinKey:
		return "Key(" + fmt.Sprintf("%x", self.bin()) + ")"
	default:
		return fmt.Sprintf("unsupported path %#v", self)
	}
}

// Type returns the type of a Path
func (self Path) Type() PathType {
	return self.t
}

// Value returns the equivalent go interface of a Path
func (self Path) Value() interface{} {
	switch self.t {
	case PathFieldId:
		return self.id()
	case PathFieldName, PathStrKey:
		return self.str()
	case PathIndex, PathIntKey:
		return self.int()
	case PathBinKey:
		return self.bin()
	default:
		return nil
	}
}

// NewPathFieldId creates a PathFieldId path
func NewPathFieldId(id thrift.FieldID) Path {
	return Path{
		t: PathFieldId,
		l: int(id),
	}
}

// NewPathFieldName creates a PathFieldName path
func NewPathFieldName(name string) Path {
	return Path{
		t: PathFieldName,
		v: *(*unsafe.Pointer)(unsafe.Pointer(&name)),
		l: len(name),
	}
}

// NewPathIndex creates a PathIndex path
func NewPathIndex(index int) Path {
	return Path{
		t: PathIndex,
		l: index,
	}
}

// NewPathStrKey creates a PathStrKey path
func NewPathStrKey(key string) Path {
	return Path{
		t: PathStrKey,
		v: *(*unsafe.Pointer)(unsafe.Pointer(&key)),
		l: len(key),
	}
}

// NewPathIntKey creates a PathIntKey path
func NewPathIntKey(key int) Path {
	return Path{
		t: PathIntKey,
		l: key,
	}
}

// NewPathBinKey creates a PathBinKey path
func NewPathBinKey(key []byte) Path {
	return Path{
		t: PathBinKey,
		v: *(*unsafe.Pointer)(unsafe.Pointer(&key)),
		l: len(key),
	}
}

// GetDescByPath searches longitudinally and returns the sub descriptor of the desc specified by path
func GetDescByPath(desc *thrift.TypeDescriptor, path ...Path) (ret *thrift.TypeDescriptor, err error) {
	ret = desc
	for _, p := range path {
		switch desc.Type() {
		case thrift.STRUCT:
			switch p.Type() {
			case PathFieldId:
				f := desc.Struct().FieldById(thrift.FieldID(p.l))
				if f == nil {
					return nil, errNode(meta.ErrUnknownField, fmt.Sprintf("unknown field %d", p.l), nil)
				}
				ret = f.Type()
			case PathFieldName:
				f := desc.Struct().FieldByKey(p.str())
				if f == nil {
					return nil, errNode(meta.ErrUnknownField, fmt.Sprintf("unknown field %s", p.str()), nil)
				}
				ret = f.Type()
			default:
				return nil, errNode(meta.ErrInvalidParam, "", nil)
			}
		case thrift.LIST, thrift.SET, thrift.MAP:
			ret = desc.Elem()
		default:
			return nil, errNode(meta.ErrUnsupportedType, "", err)
		}
		if ret == nil {
			return nil, errNode(meta.ErrNotFound, "", err)
		}
	}
	return
}

// PathNode is a three node of DOM tree
type PathNode struct {
	Path
	Node
	Next []PathNode
}

var pathNodePool = sync.Pool{
	New: func() interface{} {
		return &PathNode{}
	},
}

// NewPathNode get a new PathNode from memory pool
func NewPathNode() *PathNode {
	return pathNodePool.Get().(*PathNode)
}

// FreePathNode put a PathNode back to memory pool
func FreePathNode(p *PathNode) {
	p.Path = Path{}
	p.Node = Node{}
	p.Next = p.Next[:0]
	pathNodePool.Put(p)
}

// DescriptorToPathNode converts a thrift type descriptor to a DOM, assgining path to root
// NOTICE: it only recursively converts STRUCT type
func DescriptorToPathNode(desc *thrift.TypeDescriptor, root *PathNode, opts *Options) error {
	if desc == nil || root == nil {
		panic("nil pointer")
	}
	if desc.Type() == thrift.STRUCT {
		fs := desc.Struct().Fields()
		ns := root.Next
		if cap(ns) == 0 {
			ns = make([]PathNode, 0, len(fs))
		} else {
			ns = ns[:0]
		}
		for _, f := range fs {
			var p PathNode
			// if opts.FieldByName {
			// 	p.Path = NewPathFieldName(f.Name())
			// } else {
			p.Path = NewPathFieldId(f.ID())
			// }
			if err := DescriptorToPathNode(f.Type(), &p, opts); err != nil {
				return err
			}
			ns = append(ns, p)
		}
		root.Next = ns

	}
	return nil
}

// Assgin assigns self's raw Value according to its Next Path,
// which must be set before calling this method.
func (self *PathNode) Assgin(recurse bool, opts *Options) error {
	if self.Node.IsError() {
		return self.Node
	}
	if len(self.Next) == 0 {
		return nil
	}
	if err := self.Node.GetMany(self.Next, opts); err != nil {
		return err
	}
	if !recurse {
		return nil
	}
	for _, n := range self.Next {
		if n.Node.IsEmpty() {
			continue
		}
		if err := n.assgin(_SkipMaxDepth, opts); err != nil {
			return err
		}
	}
	return nil
}

func (self *PathNode) assgin(depth int, opts *Options) error {
	if self.Node.IsError() {
		return self.Node
	}
	if len(self.Next) == 0 || depth == 0 {
		return nil
	}
	if err := self.Node.GetMany(self.Next, opts); err != nil {
		return err
	}
	for _, n := range self.Next {
		if n.Node.IsEmpty() {
			continue
		}
		if err := n.assgin(depth-1, opts); err != nil {
			break
		}
	}
	return nil
}

// Load loads self's all children ( and children's children if recurse is true) into self.Next,
// no matter whether self.Next is empty or set before (will be reset).
// NOTICE: if opts.NotScanParentNode is true, the parent nodes (PathNode.Node) of complex (map/list/struct) type won't be assgined data
func (self *PathNode) Load(recurse bool, opts *Options) error {
	if self == nil {
		panic("nil PathNode")
	}
	if self.Error() != "" {
		return self
	}
	self.Next = self.Next[:0]
	p := thrift.BinaryProtocol{
		Buf: self.Node.raw(),
	}
	return self.scanChildren(&p, recurse, opts)
}

// Fork deeply copy self and its children to a new PathNode
func (self PathNode) Fork() PathNode {
	n := PathNode{
		Path: self.Path,
		Node: self.Node,
	}
	for _, c := range self.Next {
		n.Next = append(n.Next, c.Fork())
	}
	return n
}

// CopyTo deeply copy self and its children to a PathNode
func (self PathNode) CopyTo(to *PathNode) {
	to.Path = self.Path
	to.Node = self.Node
	if cap(to.Next) < len(self.Next) {
		to.Next = make([]PathNode, len(self.Next))
	}
	to.Next = to.Next[:len(self.Next)]
	for i, c := range self.Next {
		c.CopyTo(&to.Next[i])
	}
}

// ResetValue resets self's node and its children's node
func (self *PathNode) ResetValue() {
	self.Node = Node{}
	for i := range self.Next {
		self.Next[i].ResetValue()
	}
}

// ResetAll resets self and its children, including path and node both
func (self *PathNode) ResetAll() {
	for i := range self.Next {
		self.Next[i].ResetAll()
	}
	self.Node = Node{}
	self.Path = Path{}
	self.Next = self.Next[:0]
}

// Marshal marshals self to thrift bytes
func (self PathNode) Marshal(opt *Options) (out []byte, err error) {
	p := thrift.NewBinaryProtocolBuffer()
	err = self.marshal(p, opt)
	if err == nil {
		out = make([]byte, len(p.Buf))
		copy(out, p.Buf)
	}
	thrift.FreeBinaryProtocolBuffer(p)
	return
}

// MarshalIntoBuffer marshals self to thrift bytes into a buffer
func (self PathNode) MarshalIntoBuffer(out *[]byte, opt *Options) error {
	p := thrift.BinaryProtocol{
		Buf: *out,
	}
	if err := self.marshal(&p, opt); err != nil {
		return err
	}
	*out = p.Buf
	return nil
}

func (self PathNode) marshal(p *thrift.BinaryProtocol, opts *Options) error {
	if self.IsError() {
		return self.Node
	}
	if len(self.Next) == 0 {
		p.Buf = append(p.Buf, self.raw()...)
		return nil
	}

	var err error
	// desc := self.Node.d
	switch self.Node.t {
	case thrift.STRUCT:
		// err = p.WriteStructBegin("")
		// if err != nil {
		// 	return wrapError(meta.ErrWrite, "", err)
		// }
		for _, v := range self.Next {
			// NOTICE: we skip if the value is empty.
			if v.IsEmpty() {
				continue
			}
			err = p.WriteFieldBegin("", v.Node.t, (v.Path.id()))
			if err != nil {
				return wrapError(meta.ErrWrite, "", err)
			}
			if err = v.marshal(p, opts); err != nil {
				return unwrapError(fmt.Sprintf("field %d  marshal failed", v.Path.id()), err)
			}
			err = p.WriteFieldEnd()
		}
		// NOTICE: we don't check if required fields unset.
		err = p.WriteStructEnd()
		if err != nil {
			return wrapError(meta.ErrWrite, "", err)
		}
	case thrift.LIST, thrift.SET:
		size := len(self.Next)
		rewrite, err := p.WriteListBeginWithSizePos(self.et, size)
		// if found emtpy value, we must rewite size
		if err != nil {
			return wrapError(meta.ErrWrite, "", err)
		}
		for _, v := range self.Next {
			// NOTICE: we rewite list size if the value is empty.
			if v.IsEmpty() {
				size -= 1
				p.ModifyI32(rewrite, int32(size))
				continue
			}
			if err = v.marshal(p, opts); err != nil {
				return unwrapError(fmt.Sprintf("element %v of list marshal failed", v.Path.v), err)
			}
		}
		// err = p.WriteSetEnd()
		// if err != nil {
		// 	return wrapError(meta.ErrWrite, "", err)
		// }
	case thrift.MAP:
		size := len(self.Next)
		rewrite, err := p.WriteMapBeginWithSizePos(self.kt, self.et, len(self.Next))
		if err != nil {
			return wrapError(meta.ErrWrite, "", err)
		}
		if kt := self.kt; kt == thrift.STRING {
			for _, v := range self.Next {
				// NOTICE: we skip if the value is empty.
				if v.IsEmpty() {
					size -= 1
					p.ModifyI32(rewrite, int32(size))
					continue
				}
				if v.Path.Type() != PathStrKey {
					return errNode(meta.ErrDismatchType, "path key must be string", nil)
				}
				if err = p.WriteString(v.Path.str()); err != nil {
					return wrapError(meta.ErrWrite, fmt.Sprintf("key %v of map marshal failed", v.Path.str()), err)
				}
				if err = v.marshal(p, opts); err != nil {
					return unwrapError(fmt.Sprintf("element %v marshal failed", v.Path.str()), err)
				}
			}

		} else if kt.IsInt() {
			for _, v := range self.Next {
				// NOTICE: we skip if the value is empty.
				if v.IsEmpty() {
					size -= 1
					p.ModifyI32(rewrite, int32(size))
					continue
				}
				if v.Path.Type() != PathIntKey {
					return errNode(meta.ErrDismatchType, "path key must be int", nil)
				}
				if err = p.WriteInt(kt, v.Path.int()); err != nil {
					return wrapError(meta.ErrWrite, fmt.Sprintf("key %v of map marshal failed", v.Path.int()), err)
				}
				if err = v.marshal(p, opts); err != nil {
					return unwrapError(fmt.Sprintf("element %v marshal failed", v.Path.int()), err)
				}
			}
		} else {
			for _, v := range self.Next {
				// NOTICE: we skip if the value is empty.
				if v.IsEmpty() {
					size -= 1
					p.ModifyI32(rewrite, int32(size))
					continue
				}
				if v.Path.Type() != PathBinKey {
					return errNode(meta.ErrDismatchType, "path key must be binary", nil)
				}
				p.Buf = append(p.Buf, v.Path.bin()...)
				if err = v.marshal(p, opts); err != nil {
					return unwrapError(fmt.Sprintf("element %v marshal failed", v.Path.bin()), err)
				}
			}
		}
		// err = p.WriteMapEnd()
		// if err != nil {
		// 	return wrapError(meta.ErrWrite, "", err)
		// }

	case thrift.BOOL, thrift.I08, thrift.I16, thrift.I32, thrift.I64, thrift.DOUBLE, thrift.STRING:
		p.Buf = append(p.Buf, self.raw()...)
	default:
		return meta.NewError(meta.ErrUnsupportedType, "", nil)
	}

	return err
}

type pnSlice struct {
	a []PathNode
	b []PathNode
}

func (self pnSlice) Len() int {
	return len(self.a)
}

func (self *pnSlice) Swap(i, j int) {
	self.a[i], self.a[j] = self.a[j], self.a[i]
	self.b[i], self.b[j] = self.b[j], self.b[i]
}

func (self pnSlice) Less(i, j int) bool {
	return int(uintptr(self.a[i].Node.v)) < int(uintptr(self.a[j].Node.v))
}

func (self *pnSlice) Sort() {
	sort.Sort(self)
}

func (self *PathNode) handleChild(in *[]PathNode, lp *int, cp *int, p *thrift.BinaryProtocol, recurse bool, opts *Options, et thrift.Type) (*PathNode, error) {
	var con = *in
	var l = *lp
	var c = *cp
	if l == c {
		c = c + DefaultNodeSliceCap
		tmp := make([]PathNode, l, c)
		copy(tmp, con)
		con = tmp
	}
	con = con[:l+1]
	v := &con[l]
	l += 1

	ss := p.Read
	buf := p.Buf

	if recurse && (et.IsComplex() && opts.NotScanParentNode) {
		v.Node = Node{
			t: et,
			l: 0,
			v: unsafe.Pointer(uintptr(self.Node.v) + uintptr(ss)),
		}
	} else {
		if e := p.Skip(et, opts.UseNativeSkip); e != nil {
			return nil, errNode(meta.ErrRead, "", e)
		}
		v.Node = self.slice(ss, p.Read, et)
	}

	if recurse && et.IsComplex() {
		p.Buf = p.Buf[ss:]
		p.Read = 0
		if err := v.scanChildren(p, recurse, opts); err != nil {
			return nil, err
		}
		p.Buf = buf
		p.Read = ss + p.Read
	}

	*in = con
	*lp = l
	*cp = c
	return v, nil
}

func (self *PathNode) scanChildren(p *thrift.BinaryProtocol, recurse bool, opts *Options) error {
	var con = self.Next[:0]
	var v *PathNode
	var err error
	l := len(con)
	c := cap(con)

	switch self.Node.t {
	case thrift.STRUCT:
		// name, err := p.ReadStructBegin()
		for {
			_, et, id, e := p.ReadFieldBegin()
			if e != nil {
				return errNode(meta.ErrRead, "", e)
			}
			if et == thrift.STOP {
				break
			}
			v, err = self.handleChild(&con, &l, &c, p, recurse, opts, et)
			if err != nil {
				return err
			}
			v.Path = NewPathFieldId(thrift.FieldID(id))
		}
	case thrift.LIST, thrift.SET:
		et, size, e := p.ReadListBegin()
		if e != nil {
			return errNode(meta.ErrRead, "", e)
		}
		self.et = et
		for i := 0; i < size; i++ {
			v, err = self.handleChild(&con, &l, &c, p, recurse, opts, et)
			if err != nil {
				return err
			}
			v.Path = NewPathIndex(i)
		}
	case thrift.MAP:
		kt, et, size, e := p.ReadMapBegin()
		if e != nil {
			return errNode(meta.ErrRead, "", e)
		}
		self.et = et
		self.kt = kt
		if kt == thrift.STRING {
			for i := 0; i < size; i++ {
				key, e := p.ReadString(false)
				if e != nil {
					return errNode(meta.ErrRead, "", e)
				}
				v, err = self.handleChild(&con, &l, &c, p, recurse, opts, et)
				if err != nil {
					return err
				}
				v.Path = NewPathStrKey(key)
			}
		} else if kt.IsInt() {
			for i := 0; i < size; i++ {
				key, e := p.ReadInt(kt)
				if e != nil {
					return errNode(meta.ErrRead, "", e)
				}
				v, err = self.handleChild(&con, &l, &c, p, recurse, opts, et)
				if err != nil {
					return err
				}
				v.Path = NewPathIntKey(key)
			}
		} else {
			for i := 0; i < size; i++ {
				ks := p.Read
				if e := p.Skip(kt, opts.UseNativeSkip); e != nil {
					return errNode(meta.ErrRead, "", e)
				}
				ke := p.Read
				v, err = self.handleChild(&con, &l, &c, p, recurse, opts, et)
				if err != nil {
					return err
				}
				v.Path = NewPathBinKey(p.Buf[ks:ke])
			}
		}
	default:
		return errNode(meta.ErrUnsupportedType, "", nil)
	}

	self.Next = con
	return nil
}
