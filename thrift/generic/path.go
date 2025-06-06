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

	"github.com/cloudwego/dynamicgo/internal/caching"
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
	return descriptorToPathNode(0, desc, root, opts)
}

func descriptorToPathNode(recurse int, desc *thrift.TypeDescriptor, root *PathNode, opts *Options) error {
	if opts.DescriptorToPathNodeMaxDepth > 0 && recurse > opts.DescriptorToPathNodeMaxDepth {
		return nil
	}
	if desc == nil || root == nil {
		panic("nil pointer")
	}
	t := desc.Type()
	root.Node.t = t
	if t == thrift.STRUCT {
		fs := desc.Struct().Fields()
		ns := root.Next
		if cap(ns) == 0 {
			ns = make([]PathNode, 0, len(fs))
		} else {
			ns = ns[:0]
		}
		for _, f := range fs {
			if f.Required() == thrift.OptionalRequireness && !opts.DescriptorToPathNodeWriteOptional {
				continue
			}
			if f.Required() == thrift.DefaultRequireness && !opts.DescriptorToPathNodeWriteDefualt {
				continue
			}
			var p PathNode
			// if opts.FieldByName {
			// 	p.Path = NewPathFieldName(f.Name())
			// } else {
			p.Path = NewPathFieldId(f.ID())
			// println("field: ", f.ID())
			// }
			if err := descriptorToPathNode(recurse+1, f.Type(), &p, opts); err != nil {
				return err
			}
			ns = append(ns, p)
		}
		root.Next = ns
	} else if t == thrift.LIST || t == thrift.SET {
		if opts.DescriptorToPathNodeArraySize > 0 {
			next := make([]PathNode, opts.DescriptorToPathNodeArraySize)
			for i := range next {
				next[i].Path = NewPathIndex(i)
				if err := descriptorToPathNode(recurse+1, desc.Elem(), &next[i], opts); err != nil {
					return err
				}
			}
			root.Next = next
			root.Node.et = desc.Elem().Type()
		}
	} else if t == thrift.MAP {
		if opts.DescriptorToPathNodeMapSize > 0 {
			next := make([]PathNode, opts.DescriptorToPathNodeMapSize)
			for i := range next {
				if ty := desc.Key().Type(); ty.IsInt() {
					next[i].Path = NewPathIntKey(i) // NOTICE: use index as int key here
				} else if ty == thrift.STRING {
					next[i].Path = NewPathStrKey(strconv.Itoa(i)) // NOTICE: use index as string key here
				} else {
					buf := thrift.NewBinaryProtocol([]byte{})
					_ = buf.WriteEmpty(desc.Key()) // NOTICE: use emtpy value as binary key here
					next[i].Path = NewPathBinKey(buf.Buf)
				}
				if err := descriptorToPathNode(recurse+1, desc.Elem(), &next[i], opts); err != nil {
					return err
				}
			}
			root.Next = next
			root.Node.kt = desc.Key().Type()
			root.Node.et = desc.Elem().Type()
		}
	} else {
		// println("type: ", desc.Type().String())
		buf := thrift.NewBinaryProtocol([]byte{})
		_ = buf.WriteEmpty(desc)
		root.Node = NewNode(desc.Type(), buf.Buf)
	}
	return nil
}

// PathNodeToInterface convert a pathnode to a interface
func PathNodeToInterface(tree PathNode, opts *Options, useParent bool) interface{} {
	switch tree.Node.Type() {
	case thrift.STRUCT:
		if len(tree.Next) == 0 && useParent {
			vv, err := tree.Node.Interface(opts)
			if err != nil {
				panic(err)
			}
			return vv
		}
		ret := make(map[int]interface{}, len(tree.Next))
		for _, v := range tree.Next {
			vv := PathNodeToInterface(v, opts, useParent)
			if vv == nil {
				continue
			}
			ret[int(v.Path.Id())] = vv
		}
		return ret
	case thrift.LIST, thrift.SET:
		if len(tree.Next) == 0 && useParent {
			vv, err := tree.Node.Interface(opts)
			if err != nil {
				panic(err)
			}
			return vv
		}
		ret := make([]interface{}, len(tree.Next))
		for _, v := range tree.Next {
			vv := PathNodeToInterface(v, opts, useParent)
			if vv == nil {
				continue
			}
			ret[v.Path.Int()] = vv
		}
		return ret
	case thrift.MAP:
		if len(tree.Next) == 0 && useParent {
			vv, err := tree.Node.Interface(opts)
			if err != nil {
				panic(err)
			}
			return vv
		}
		if kt := tree.Node.kt; kt == thrift.STRING {
			ret := make(map[string]interface{}, len(tree.Next))
			for _, v := range tree.Next {
				vv := PathNodeToInterface(v, opts, useParent)
				if vv == nil {
					continue
				}
				ret[v.Path.Str()] = vv
			}
			return ret
		} else if kt.IsInt() {
			ret := make(map[int]interface{}, len(tree.Next))
			for _, v := range tree.Next {
				vv := PathNodeToInterface(v, opts, useParent)
				if vv == nil {
					continue
				}
				ret[v.Path.Int()] = vv
			}
			return ret
		} else {
			ret := make(map[interface{}]interface{}, len(tree.Next))
			for _, v := range tree.Next {
				kp := PathNode{
					Node: NewNode(kt, v.Path.Bin()),
				}
				kv := PathNodeToInterface(kp, opts, true)
				if kv == nil {
					continue
				}
				vv := PathNodeToInterface(v, opts, useParent)
				if vv == nil {
					continue
				}
				switch x := kv.(type) {
				case map[string]interface{}:
					ret[&x] = vv
				case map[int]interface{}:
					ret[&x] = vv
				case map[interface{}]interface{}:
					ret[&x] = vv
				case []interface{}:
					ret[&x] = vv
				default:
					ret[kv] = vv
				}
			}
			return ret
		}
	case thrift.STOP:
		return nil
	default:
		ret, err := tree.Node.Interface(opts)
		if err != nil {
			panic(err)
		}
		return ret
	}
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

func guardPathNodeSlice(con *[]PathNode, l int) {
	c := cap(*con)
	if l >= c {
		tmp := make([]PathNode, len(*con), l+DefaultNodeSliceCap)
		copy(tmp, *con)
		*con = tmp
	}
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
	guardPathNodeSlice(&con, l)
	if l >= len(con) {
		con = con[:l+1]
	}
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
		if e := p.SkipType(et); e != nil {
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
	*cp = cap(con)
	return v, nil
}

// Error returns non-empty string if the PathNode has error
func (self PathNode) Error() string {
	return self.Node.Error()
}

// Check returns non-nil error if the PathNode has error
func (self *PathNode) Check() error {
	if self == nil {
		return errNotFound
	}
	if self.Node.t == thrift.ERROR {
		return self
	}
	return nil
}

func (self *PathNode) should(op string, t thrift.Type) *PathNode {
	if self == nil {
		return errPathNode(meta.ErrNotFound, op, nil)
	}
	if self.Node.t != t {
		return errPathNode(meta.ErrDismatchType, op, nil)
	}
	return nil
}

func (self *PathNode) should2(op string, t thrift.Type, t2 thrift.Type) *PathNode {
	if self == nil {
		return errPathNode(meta.ErrNotFound, op, nil)
	}
	if self.Node.t != t && self.Node.t != t2 {
		return errPathNode(meta.ErrDismatchType, op, nil)
	}
	return nil
}

func getStrHash(next *[]PathNode, key string, N int) *PathNode {
	h := int(caching.StrHash(key) % uint64(N))
	s := (*PathNode)(rt.IndexPtr(*(*unsafe.Pointer)(unsafe.Pointer(next)), sizePathNode, h))
	for s.Path.t == PathStrKey {
		if s.Path.str() == key {
			return s
		}
		h = (h + 1) % N
		s = (*PathNode)(unsafe.Pointer(uintptr(unsafe.Pointer(s)) + sizePathNode))
	}
	return nil
}

func seekIntHash(next unsafe.Pointer, key uint64, N int) int {
	h := int(key % uint64(N))
	s := (*PathNode)(rt.IndexPtr(next, sizePathNode, h))
	for s.Path.t != 0 {
		h = (h + 1) % N
		s = (*PathNode)(rt.AddPtr(unsafe.Pointer(s), sizePathNode))
	}
	return h
}

func getIntHash(next *[]PathNode, key uint64, N int) *PathNode {
	h := int(key % uint64(N))
	s := (*PathNode)(rt.IndexPtr(*(*unsafe.Pointer)(unsafe.Pointer(next)), sizePathNode, h))
	for s.Path.t == PathIntKey {
		if uint64(s.Path.int()) == key {
			return s
		}
		h = (h + 1) % N
		s = (*PathNode)(rt.AddPtr(unsafe.Pointer(s), sizePathNode))
	}
	return nil
}

// GetByInt get the child node by string. Only support MAP with string-type key.
//
// If opts.StoreChildrenByHash is true, it will try to use hash (O(1)) to search the key.
// However, if the map size has changed, it may fallback to O(n) search.
func (self *PathNode) GetByStr(key string, opts *Options) *PathNode {
	if err := self.should("GetStrKey() only support MAP", thrift.MAP); err != nil {
		return err
	}
	if self.Node.kt != thrift.STRING {
		return errPathNode(meta.ErrDismatchType, "GetStrKey() only support MAP with string key", nil)
	}
	// fast path: use hash to find the key.
	if opts.StoreChildrenByHash {
		n, _ := self.Node.len()
		N := n * 2
		// TODO: cap may change after Set. Use better way to store hash size
		if cap(self.Next) >= N {
			if s := getStrHash(&self.Next, key, N); s != nil {
				return s
			}
		}
		// not find, maybe hash size has changed, try to search from the beginning.
	}
	for i := range self.Next {
		v := &self.Next[i]
		if v.Path.t == PathStrKey && v.Path.str() == key {
			return v
		}
	}
	return nil
}

// SetByStr set the child node by string. Only support MAP with string-type key.
// If the key already exists, it will be overwritten and return true.
//
// If opts.StoreChildrenByHash is true, it will try to use hash (O(1)) to search the key.
// However, if the map hash size has changed, it may fallback to O(n) search.
func (self *PathNode) SetByStr(key string, val Node, opts *Options) (bool, error) {
	if err := self.should("SetStrKey() only support MAP", thrift.MAP); err != nil {
		return false, err
	}
	if self.Node.kt != thrift.STRING {
		return false, errPathNode(meta.ErrDismatchType, "GetStrKey() only support MAP with string key", nil)
	}
	// fast path: use hash to find the key.
	if opts.StoreChildrenByHash {
		n, _ := self.Node.len()
		N := n * 2
		// TODO: cap may change after Set. Use better way to store hash size
		if cap(self.Next) >= N {
			if s := getStrHash(&self.Next, key, N); s != nil {
				s.Node = val
				return true, nil
			}
		}
		// not find, maybe hash size has changed, try to search from the beginning.
	}
	for i := range self.Next {
		v := &self.Next[i]
		if v.Path.t == PathStrKey && v.Path.str() == key {
			v.Node = val
			return true, nil
		}
	}
	self.Next = append(self.Next, PathNode{
		Path: NewPathStrKey(key),
		Node: val,
	})
	return false, nil
}

// GetByInt get the child node by integer. Only support MAP with integer-type key.
//
// If opts.StoreChildrenByHash is true, it will try to use hash (O(1)) to search the key.
// However, if the map size has changed, it may fallback to O(n) search.
func (self *PathNode) GetByInt(key int, opts *Options) *PathNode {
	if err := self.should("GetByInt() only support MAP", thrift.MAP); err != nil {
		return err
	}
	if !self.Node.kt.IsInt() {
		return errPathNode(meta.ErrDismatchType, "GetByInt() only support MAP with integer key", nil)
	}
	// fast path: use hash to find the key.
	if opts.StoreChildrenByHash {
		// TODO: size may change after Set. Use better way to store hash size
		n, _ := self.Node.len()
		N := n * 2
		if cap(self.Next) >= N {
			if s := getIntHash(&self.Next, uint64(key), N); s != nil {
				return s
			}
		}
		// not find, maybe hash size has changed, try to search from the beginning.
	}
	for i := range self.Next {
		v := &self.Next[i]
		if v.Path.t == PathIntKey && v.Path.int() == key {
			return v
		}
	}
	return nil
}

// SetByInt set the child node by integer. Only support MAP with integer-type key.
// If the key already exists, it will be overwritten and return true.
//
// If opts.StoreChildrenByHash is true, it will try to use hash (O(1)) to search the key.
// However, if the map hash size has changed, it may fallback to O(n) search.
func (self *PathNode) SetByInt(key int, val Node, opts *Options) (bool, error) {
	if err := self.should("SetByInt() only support MAP", thrift.MAP); err != nil {
		return false, err
	}
	if !self.Node.kt.IsInt() {
		return false, errPathNode(meta.ErrDismatchType, "SetByInt() only support MAP with integer key", nil)
	}
	// fast path: use hash to find the key.
	if opts.StoreChildrenByHash {
		n, _ := self.Node.len()
		N := n * 2
		if cap(self.Next) >= N {
			if s := getIntHash(&self.Next, uint64(key), N); s != nil {
				s.Node = val
				return true, nil
			}
		}
		// not find, maybe hash size has changed, try to search from the beginning.
	}
	for i := range self.Next {
		v := &self.Next[i]
		if v.Path.t == PathIntKey && v.Path.int() == key {
			v.Node = val
			return true, nil
		}
	}
	self.Next = append(self.Next, PathNode{
		Path: NewPathIntKey(key),
		Node: val,
	})
	return false, nil
}

// Field get the child node by field id. Only support STRUCT.
//
// If opts.StoreChildrenById is true, it will try to use id (O(1)) as index to search the key.
// However, if the struct fields have changed, it may fallback to O(n) search.
func (self *PathNode) Field(id thrift.FieldID, opts *Options) *PathNode {
	if err := self.should("GetById() only support STRUCT", thrift.STRUCT); err != nil {
		return err
	}
	// fast path: use id to find the key.
	if opts.StoreChildrenById && int(id) <= StoreChildrenByIdShreshold {
		v := &self.Next[id]
		if v.Path.t != 0 && v.Path.id() == id {
			return v
		}
	}
	// slow path: use linear search to find the id.
	for i := StoreChildrenByIdShreshold; i < len(self.Next); i++ {
		v := &self.Next[i]
		if v.Path.t == PathFieldId && v.Path.id() == id {
			return v
		}
	}
	for i := 0; i < len(self.Next) && i < StoreChildrenByIdShreshold; i++ {
		v := &self.Next[i]
		if v.Path.t == PathFieldId && v.Path.id() == id {
			return v
		}
	}
	return nil
}

// SetField set the child node by field id. Only support STRUCT.
// If the key already exists, it will be overwritten and return true.
//
// If opts.StoreChildrenById is true, it will try to use id (O(1)) as index to search the key.
// However, if the struct fields have changed, it may fallback to O(n) search.
func (self *PathNode) SetField(id thrift.FieldID, val Node, opts *Options) (bool, error) {
	if err := self.should("GetById() only support STRUCT", thrift.STRUCT); err != nil {
		return false, err
	}
	// fast path: use id to find the key.
	if opts.StoreChildrenById && int(id) <= StoreChildrenByIdShreshold {
		v := &self.Next[id]
		exist := v.Path.t != 0
		v.Node = val
		return exist, nil
	}
	// slow path: use linear search to find the id.
	for i := StoreChildrenByIdShreshold; i < len(self.Next); i++ {
		v := &self.Next[i]
		if v.Path.t == PathFieldId && v.Path.id() == id {
			v.Node = val
			return true, nil
		}
	}
	for i := 0; i < len(self.Next) && i < StoreChildrenByIdShreshold; i++ {
		v := &self.Next[i]
		if v.Path.t == PathFieldId && v.Path.id() == id {
			v.Node = val
			return true, nil
		}
	}
	// not find, append to the end.
	self.Next = append(self.Next, PathNode{
		Path: NewPathFieldId(id),
		Node: val,
	})
	return false, nil
}

func (self *PathNode) scanChildren(p *thrift.BinaryProtocol, recurse bool, opts *Options) error {
	var con = self.Next[:0]
	var v *PathNode
	var err error
	l := len(con)
	c := cap(con)

	var tp = StoreChildrenByIdShreshold
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
			// OPT: store children by id here, thus we can use id as index to access children.
			if opts.StoreChildrenById {
				// if id is larger than the threshold, we store children after the threshold.
				if int(id) < StoreChildrenByIdShreshold {
					l = int(id)
				} else {
					l = tp
					tp += 1
				}
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
		var conAddr unsafe.Pointer
		var N int

		if kt == thrift.STRING {
			// fast path: use hash to store the key.
			if opts.StoreChildrenByHash && size > StoreChildrenByIntHashShreshold {
				// NOTE: we use original count*2 as the capacity of the hash table.
				N = size * 2
				guardPathNodeSlice(&con, N-1)
				conAddr = *(*unsafe.Pointer)(unsafe.Pointer(&con))
				c = N
			}
			for i := 0; i < size; i++ {
				key, e := p.ReadString(false)
				if e != nil {
					return errNode(meta.ErrRead, "", e)
				}
				// fast path: use hash to store the key.
				if N != 0 {
					l = seekIntHash(conAddr, caching.StrHash(key), N)
				}
				v, err = self.handleChild(&con, &l, &c, p, recurse, opts, et)
				if err != nil {
					return err
				}
				v.Path = NewPathStrKey(key)
			}
		} else if kt.IsInt() {
			// fast path: use hash to store the key.
			if opts.StoreChildrenByHash && size > StoreChildrenByIntHashShreshold {
				// NOTE: we use original count*2 as the capacity of the hash table.
				N = size * 2
				guardPathNodeSlice(&con, N-1)
				conAddr = *(*unsafe.Pointer)(unsafe.Pointer(&con))
				c = N
			}
			for i := 0; i < size; i++ {
				key, e := p.ReadInt(kt)
				if e != nil {
					return errNode(meta.ErrRead, "", e)
				}
				// fast path: use hash to store the key.
				if N != 0 {
					l = seekIntHash(conAddr, uint64(key), N)
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
				if e := p.SkipType(kt); e != nil {
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
