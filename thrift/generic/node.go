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
	"bytes"
	"fmt"
	"sync"
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
)

// Node is a generic wrap of raw thrift data
type Node struct {
	t  thrift.Type
	et thrift.Type
	kt thrift.Type
	v  unsafe.Pointer
	l  int
}

// NewNodeBool create a new node with given type and data
// WARN: it WON'T check the correctness of the data
func NewNode(t thrift.Type, src []byte) Node {
	ret := Node{
		t: t,
		l: (len(src)),
		v: rt.GetBytePtr(src),
	}
	if t == thrift.LIST || t == thrift.SET {
		ret.et = *(*thrift.Type)(unsafe.Pointer(ret.v))
	} else if t == thrift.MAP {
		ret.kt = *(*thrift.Type)(unsafe.Pointer(ret.v))
		ret.et = *(*thrift.Type)(rt.AddPtr(ret.v, uintptr(1)))
	}
	return ret
}

func (self Node) slice(s int, e int, t thrift.Type) Node {
	ret := Node{
		t: t,
		l: (e - s),
		v: rt.AddPtr(self.v, uintptr(s)),
	}
	if t == thrift.LIST || t == thrift.SET {
		ret.et = *(*thrift.Type)(unsafe.Pointer(ret.v))
	} else if t == thrift.MAP {
		ret.kt = *(*thrift.Type)(unsafe.Pointer(ret.v))
		ret.et = *(*thrift.Type)(rt.AddPtr(ret.v, uintptr(1)))
	}
	return ret
}

func (self Node) offset() unsafe.Pointer {
	return rt.AddPtr(self.v, uintptr(self.l))
}

// func NewNodeAny(val interface{}) Node {
// 	vt := reflect.ValueOf(val)
// 	switch vt.Kind() {
// 	case reflect.Int8:
// 		return NewNodeInt8(int8(vt.Int()))
// 	case reflect.Int16:
// 		return NewNodeInt16(int16(vt.Int()))
// 	case reflect.Int32:
// 		return NewNodeInt32(int32(vt.Int()))
// 	case reflect.Int64:
// 		return NewNodeInt64(vt.Int())
// 	case reflect.Float32, reflect.Float64:
// 		return NewNodeDouble(vt.Float())
// 	case reflect.String:
// 		return NewNodeString(vt.String())
// 	case reflect.Bool:
// 		return NewNodeBool(vt.Bool())
// 	case reflect.Slice:
// 		if vt.Type().Elem().Kind() == reflect.Uint8 {
// 			return NewNodeBinary(vt.Bytes())
// 		}
// 		panic("not supported")
// 	default:
// 		panic("not supported")
// 	}
// }

// NewNodeBool converts a bool value to a BOOL node
func NewNodeBool(val bool) Node {
	buf := make([]byte, 1, 1)
	thrift.BinaryEncoding{}.EncodeBool(buf, val)
	return NewNode(thrift.BOOL, buf)
}

// NewNodeInt8 converts a byte value to a BYTE node
func NewNodeByte(val byte) Node {
	buf := make([]byte, 1, 1)
	thrift.BinaryEncoding{}.EncodeByte(buf, val)
	return NewNode(thrift.BYTE, buf)
}

// NewNodeInt16 converts a int16 value to a I16 node
func NewNodeInt16(val int16) Node {
	buf := make([]byte, 2, 2)
	thrift.BinaryEncoding{}.EncodeInt16(buf, val)
	return NewNode(thrift.I16, buf)
}

// NewNodeInt32 converts a int32 value to a I32 node
func NewNodeInt32(val int32) Node {
	buf := make([]byte, 4, 4)
	thrift.BinaryEncoding{}.EncodeInt32(buf, val)
	return NewNode(thrift.I32, buf)
}

// NewNodeInt64 converts a int64 value to a I64 node
func NewNodeInt64(val int64) Node {
	buf := make([]byte, 8, 8)
	thrift.BinaryEncoding{}.EncodeInt64(buf, val)
	return NewNode(thrift.I64, buf)
}

// NewNodeDouble converts a float64 value to a DOUBLE node
func NewNodeDouble(val float64) Node {
	buf := make([]byte, 8, 8)
	thrift.BinaryEncoding{}.EncodeDouble(buf, val)
	return NewNode(thrift.DOUBLE, buf)
}

// NewNodeString converts a string value to a STRING node
func NewNodeString(val string) Node {
	l := len(val) + 4
	buf := make([]byte, l, l)
	thrift.BinaryEncoding{}.EncodeString(buf, val)
	return NewNode(thrift.STRING, buf)
}

// NewNodeBinary converts a []byte value to a STRING node
func NewNodeBinary(val []byte) Node {
	l := len(val) + 4
	buf := make([]byte, l, l)
	thrift.BinaryEncoding{}.EncodeBinary(buf, val)
	return NewNode(thrift.STRING, buf)
}

// NewTypedNode creates a new Node with the given typ,
// including element type (for LIST/SET/MAP) and key type (for MAP)
func NewTypedNode(typ thrift.Type, et thrift.Type, kt thrift.Type) (ret Node){
	if !typ.Valid() {
		panic("invalid node type")
	}
	switch typ {
	case thrift.LIST, thrift.SET:
		if !et.Valid() {
			panic("invalid element type")
		}
		ret.et = et
	case thrift.MAP:
		if !et.Valid() {
			panic("invalid element type")
		}
		if !kt.Valid() {
			panic("invalid key type")
		}
		ret.et = et
		ret.kt = kt
	}
	ret.t = typ
	return 
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
func (self Node) Type() thrift.Type {
	return self.t
}

// ElemType returns the thrift type of a LIST/SET/MAP node's element
func (self Node) ElemType() thrift.Type {
	return self.et
}

// KeyType returns the thrift type of a MAP node's key
func (self Node) KeyType() thrift.Type {
	return self.kt
}

func searchFieldId(p *thrift.BinaryProtocol, id thrift.FieldID) (tt thrift.Type, start int, err error) {
	// if _, err := p.ReadStructBegin(); err != nil {
	// 	return 0, start, errNode(meta.ErrReadInput, "", err)
	// }
	for {
		_, t, i, err := p.ReadFieldBegin()
		if err != nil {
			return 0, start, errNode(meta.ErrRead, "", err)
		}
		if t == thrift.STOP {
			return thrift.STRUCT, start, errNotFound
		}
		if id == thrift.FieldID(i) {
			start = p.Read
			tt = t
			break
		}
		if err := p.Skip(t, UseNativeSkipForGet); err != nil {
			return 0, start, errNode(meta.ErrRead, "", err)
		}
	}
	return
}

func searchIndex(p *thrift.BinaryProtocol, id int, isList bool) (tt thrift.Type, start int, err error) {
	et, size, err := p.ReadListBegin()
	if err != nil {
		return 0, start, errNode(meta.ErrRead, "", err)
	}
	if id >= size {
		if isList {
			return thrift.LIST, p.Read, errNotFound
		} else {
			return thrift.SET, p.Read, errNotFound
		}
	}

	//TODO: use direct index calculation for fixed-size list
	for i := 0; i < id; i++ {
		if err := p.Skip(et, UseNativeSkipForGet); err != nil {
			return 0, start, errNode(meta.ErrRead, "", err)
		}
	}
	start = p.Read
	tt = et
	return
}

func searchStrKey(p *thrift.BinaryProtocol, id string) (tt thrift.Type, start int, err error) {
	kt, et, size, err := p.ReadMapBegin()
	if err != nil {
		return 0, start, errNode(meta.ErrRead, "", err)
	}
	if kt != thrift.STRING {
		return 0, start, errNode(meta.ErrDismatchType, "map key must be string type", nil)
	}
	start = p.Read
	found := false
	for i := 0; i < size; i++ {
		key, err := p.ReadString(false)
		if err != nil {
			return 0, start, errNode(meta.ErrRead, "", err)
		}
		if key == id {
			start = p.Read
			tt = et
			found = true
			break
		}
		if err := p.Skip(et, UseNativeSkipForGet); err != nil {
			return 0, start, errNode(meta.ErrRead, "", err)
		}
	}
	if !found {
		return thrift.MAP, start, errNotFound
	}
	return
}

func searchBinKey(p *thrift.BinaryProtocol, id []byte) (tt thrift.Type, start int, err error) {
	kt, et, size, err := p.ReadMapBegin()
	if err != nil {
		return 0, start, errNode(meta.ErrRead, "", err)
	}
	start = p.Read
	found := false
	for i := 0; i < size; i++ {
		ks := p.Read
		if err := p.Skip(kt, UseNativeSkipForGet); err != nil {
			return 0, start, errNode(meta.ErrRead, "", err)
		}
		key := p.Buf[ks:p.Read]
		if bytes.Equal(key, id) {
			start = p.Read
			tt = et
			found = true
			break
		}
		if err := p.Skip(et, UseNativeSkipForGet); err != nil {
			return 0, start, errNode(meta.ErrRead, "", err)
		}
	}
	if !found {
		return thrift.MAP, start, errNotFound
	}
	return
}

func searchIntKey(p *thrift.BinaryProtocol, id int) (tt thrift.Type, start int, err error) {
	kt, et, size, err := p.ReadMapBegin()
	if err != nil {
		return 0, start, errNode(meta.ErrRead, "", err)
	}
	start = p.Read
	found := false
	if !kt.IsInt() {
		return 0, start, errNode(meta.ErrDismatchType, fmt.Sprintf("map key type %s is not supported", kt), nil)
	}
	for i := 0; i < size; i++ {
		key, err := p.ReadInt(kt)
		if err != nil {
			return 0, start, errNode(meta.ErrRead, "", err)
		}
		if key == id {
			start = p.Read
			tt = et
			found = true
			break
		}
		if err := p.Skip(et, UseNativeSkipForGet); err != nil {
			return 0, start, errNode(meta.ErrRead, "", err)
		}
	}
	if !found {
		return thrift.MAP, start, errNotFound
	}
	return
}

// GetByPath searches longitudinally and return a sub node at the given path from the node.
//
// The path is a list of PathFieldId, PathIndex, PathStrKey, PathBinKey, PathIntKey,
// Each path MUST be a valid path type for current layer (e.g. PathFieldId is only valid for STRUCT).
// Empty path will return the current node.
func (self Node) GetByPath(pathes ...Path) Node {
	if self.Error() != "" {
		return self
	}
	if len(pathes) == 0 {
		return self
	}

	p := thrift.BinaryProtocol{
		Buf: self.raw(),
	}
	start := 0
	tt := self.t
	isList := self.t == thrift.LIST
	var err error

	for i, path := range pathes {
		switch path.t {
		case PathFieldId:
			tt, start, err = searchFieldId(&p, path.id())
			isList = self.t == thrift.LIST
		case PathFieldName:
			return errNode(meta.ErrUnsupportedType, "", nil)
		case PathIndex:
			tt, start, err = searchIndex(&p, path.int(), isList)
		case PathStrKey:
			tt, start, err = searchStrKey(&p, path.str())
		case PathIntKey:
			tt, start, err = searchIntKey(&p, path.int())
		case PathBinKey:
			tt, start, err = searchBinKey(&p, path.bin())
		default:
			return errNode(meta.ErrUnsupportedType, fmt.Sprintf("invalid %dth path: %#v", i, p), nil)
		}
		if err != nil {
			// the last one not foud, return start pointer for inserting on `SetByPath()` here
			if i == len(pathes)-1 && err == errNotFound {
				return errNotFoundLast(unsafe.Pointer(uintptr(self.v)+uintptr(start)), tt)
			}
			en := err.(Node)
			return errNode(en.ErrCode().Behavior(), "", err)
		}
	}

	if err := p.Skip(tt, UseNativeSkipForGet); err != nil {
		return errNode(meta.ErrRead, "", err)
	}
	return self.slice(start, p.Read, tt)
}

// Field returns a sub node at the given field id from a STRUCT node.
func (self Node) Field(id thrift.FieldID) (v Node) {
	if err := self.should("Field", thrift.STRUCT); err != "" {
		return errNode(meta.ErrUnsupportedType, err, nil)
	}
	it := self.iterFields()
	if it.Err != nil {
		return errNode(meta.ErrRead, "", it.Err)
	}
	for it.HasNext() {
		i, t, s, e := it.Next(UseNativeSkipForGet)
		if i == id {
			v = self.slice(s, e, t)
			goto ret
		} else if it.Err != nil {
			v = errNode(meta.ErrRead, "", it.Err)
			goto ret
		}
	}
	v = errNode(meta.ErrNotFound, fmt.Sprintf("field %d is not found in this value", id), errNotFound)
ret:
	// it.Recycle()
	return
}

// Index returns a sub node at the given index from a LIST/SET node.
func (self Node) Index(i int) (v Node) {
	if err := self.should2("Index", thrift.LIST, thrift.SET); err != "" {
		return errNode(meta.ErrUnsupportedType, err, nil)
	}

	// et := self.d.Elem()
	var s, e int
	it := self.iterElems()
	if it.Err != nil {
		return errNode(meta.ErrRead, "", it.Err)
	}
	if i >= it.size {
		v = errNode(meta.ErrInvalidParam, fmt.Sprintf("index %d exceeds list/set bound", i), nil)
		goto ret
	}
	for j := 0; it.HasNext() && j < i; j++ {
		it.Next(UseNativeSkipForGet)
	}
	if it.Err != nil {
		return errNode(meta.ErrRead, "", it.Err)
	}

	s, e = it.Next(UseNativeSkipForGet)
	v = self.slice(s, e, self.et)
ret:
	// it.Recycle()
	return
}

// GetByInt returns a sub node at the given int key from a MAP<STRING,xx> node.
func (self Node) GetByStr(key string) (v Node) {
	if err := self.should("Get", thrift.MAP); err != "" {
		return errNode(meta.ErrUnsupportedType, err, nil)
	}

	// et := self.d.Elem()
	// kt := self.d.Key()
	if self.kt != thrift.STRING {
		return errNode(meta.ErrUnsupportedType, "Get() only supports STRING key type", nil)
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

// GetByInt returns a sub node at the given int key from a MAP<INT,xx> node.
func (self Node) GetByInt(key int) (v Node) {
	if err := self.should("Get", thrift.MAP); err != "" {
		return errNode(meta.ErrUnsupportedType, err, nil)
	}

	// et := self.d.Elem()
	// kt := self.d.Key()
	if !self.kt.IsInt() {
		return errNode(meta.ErrUnsupportedType, "Get() only supports integer key type", nil)
	}

	it := self.iterPairs()
	if it.Err != nil {
		return errNode(meta.ErrRead, "", it.Err)
	}
	for it.HasNext() {
		_, s, ss, e := it.NextInt(UseNativeSkipForGet)
		if it.Err != nil {
			v = errNode(meta.ErrRead, "", it.Err)
			goto ret
		}
		if s == key {
			v = self.slice(ss, e, self.et)
			goto ret
		}
	}
	v = errNode(meta.ErrNotFound, fmt.Sprintf("key '%d' is not found in this value", key), nil)
ret:
	return
}

// GetByInt returns a sub node at the given bytes key from a MAP node.
// The key must be deep equal (bytes.Equal) to the key in the map.
func (self Node) GetByRaw(key []byte) (v Node) {
	if err := self.should("Get", thrift.MAP); err != "" {
		return errNode(meta.ErrUnsupportedType, err, nil)
	}

	it := self.iterPairs()
	if it.Err != nil {
		return errNode(meta.ErrRead, "", it.Err)
	}
	for it.HasNext() {
		_, s, ss, e := it.NextBin(UseNativeSkipForGet)
		if it.Err != nil {
			v = errNode(meta.ErrRead, "", it.Err)
			goto ret
		}
		if bytes.Equal(s, key) {
			v = self.slice(ss, e, self.et)
			goto ret
		}
	}
	v = errNode(meta.ErrNotFound, fmt.Sprintf("key '%v' is not found in this value", key), errNotFound)
ret:
	return
}

// Fields returns all sub nodes ids along with the given int path from a STRUCT node.
func (self Node) Fields(ids []PathNode, opts *Options) (err error) {
	if err := self.should("Fields", thrift.STRUCT); err != "" {
		return errValue(meta.ErrUnsupportedType, err, nil)
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
		return errValue(meta.ErrRead, "", it.Err)
	}

	need := len(ids)
	for count := 0; it.HasNext() && count < need; {
		var p *PathNode
		i, t, s, e := it.Next(opts.UseNativeSkip)
		if it.Err != nil {
			return errValue(meta.ErrRead, "", it.Err)
		}
		//TODO: use bitmap to avoid repeatedly scan
		for j, id := range ids {
			if id.Path.t == PathFieldId && id.Path.id() == i {
				p = &ids[j]
				count += 1
				break
			}
		}
		if p != nil {
			p.Node = self.slice(s, e, t)
		}
	}
	// it.Recycle()
	return
}

// Indexes returns all sub nodes along with the given int path from a LIST/SET node.
func (self Node) Indexes(ins []PathNode, opts *Options) (err error) {
	if err := self.should2("Indexes", thrift.LIST, thrift.SET); err != "" {
		return errValue(meta.ErrUnsupportedType, err, nil)
	}
	if len(ins) == 0 {
		return nil
	}

	if opts.ClearDirtyValues {
		for i := range ins {
			ins[i].Node = Node{}
		}
	}

	et := self.et
	it := self.iterElems()
	if it.Err != nil {
		return errValue(meta.ErrRead, "", it.Err)
	}

	need := len(ins)
	for count, i := 0, 0; it.HasNext() && count < need; i++ {
		s, e := it.Next(opts.UseNativeSkip)
		if it.Err != nil {
			return errValue(meta.ErrRead, "", it.Err)
		}
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
		if p != nil {
			p.Node = self.slice(s, e, et)
		}
	}
	// it.Recycle()
	return
}

// Gets returns all sub nodes along with the given key (PathStrKey|PathIntKey|PathBinKey) path from a MAP node.
func (self Node) Gets(keys []PathNode, opts *Options) (err error) {
	if err := self.should("Get", thrift.MAP); err != "" {
		return errValue(meta.ErrUnsupportedType, err, nil)
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
		return errValue(meta.ErrRead, "", it.Err)
	}
	need := len(keys)
	for count := 0; it.HasNext() && count < need; {
		for j, id := range keys {
			if id.Path.Type() == PathStrKey {
				exp := id.Path.str()
				_, s, v, e := it.NextStr(opts.UseNativeSkip)
				if it.Err != nil {
					return errValue(meta.ErrRead, "", it.Err)
				}
				if exp == s {
					p := &keys[j]
					count += 1
					p.Node = self.slice(v, e, et)
				}
			} else if id.Path.Type() == PathIntKey {
				exp := id.Path.int()
				_, s, v, e := it.NextInt(opts.UseNativeSkip)
				if it.Err != nil {
					return errValue(meta.ErrRead, "", it.Err)
				}
				if exp == s {
					p := &keys[j]
					count += 1
					p.Node = self.slice(v, e, et)
				}
			} else {
				exp := id.Path.bin()
				_, s, v, e := it.NextBin(opts.UseNativeSkip)
				if it.Err != nil {
					return errValue(meta.ErrRead, "", it.Err)
				}
				if bytes.Equal(exp, s) {
					p := &keys[j]
					count += 1
					p.Node = self.slice(v, e, et)
				}
			}
		}
	}
	// it.Recycle()
	return
}

// GetMany searches transversely and returns all the sub nodes along with the given pathes.
func (self Node) GetMany(pathes []PathNode, opts *Options) error {
	if len(pathes) == 0 {
		return nil
	}
	return self.getMany(pathes, opts.ClearDirtyValues, opts)
}

func (self Node) getMany(pathes []PathNode, clearDirty bool, opts *Options) error {
	if clearDirty {
		for i := range pathes {
			pathes[i].Node = Node{}
		}
	}
	p := pathes[0]
	switch p.Path.t {
	case PathFieldId:
		return self.Fields(pathes, opts)
	case PathIndex:
		return self.Indexes(pathes, opts)
	case PathStrKey, PathIntKey, PathBinKey:
		return self.Gets(pathes, opts)
	default:
		return errValue(meta.ErrUnsupportedType, fmt.Sprintf("invalid path: %#v", p), nil)
	}
}

// GetTree returns a tree of all sub nodes along with the given path on the tree.
// It supports longitudinally search (like GetByPath) and transversely search (like GetMany) both.
func (self Node) GetTree(tree *PathNode, opts *Options) error {
	if tree == nil {
		panic("nil PathNode tree")
	}
	tree.Node = self
	return tree.Assgin(true, opts)
}

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

func (o *Node) setNotFound(path Path, n *Node) error {
	switch o.kt {
	case thrift.STRUCT:
		// add field bytes
		key := path.ToRaw(n.t)
		src := n.raw()
		buf := make([]byte, 0, len(key)+len(src))
		buf = append(buf, key...)
		buf = append(buf, src...)
		n.l = len(buf)
		n.v = rt.GetBytePtr(buf)
	case thrift.LIST, thrift.SET:
		// modify the original size
		buf := rt.BytesFrom(rt.SubPtr(o.v, uintptr(4)), 4, 4)
		size := int(thrift.BinaryEncoding{}.DecodeInt32(buf))
		thrift.BinaryEncoding{}.EncodeInt32(buf, int32(size+1))
	case thrift.MAP:
		// modify the original size
		buf := rt.BytesFrom(rt.SubPtr(o.v, uintptr(4)), 4, 4)
		size := int(thrift.BinaryEncoding{}.DecodeInt32(buf))
		thrift.BinaryEncoding{}.EncodeInt32(buf, int32(size+1))
		// add key bytes
		key := path.ToRaw(n.t)
		src := n.raw()
		buf = make([]byte, 0, len(key)+len(src))
		buf = append(buf, key...)
		buf = append(buf, src...)
		n.l = len(buf)
		n.v = rt.GetBytePtr(buf)
	default:
		return errNode(meta.ErrDismatchType, "simple type node shouldn't have child", nil)
	}
	o.t = n.t
	o.l = 0
	return nil
}

// SetByPath sets the sub node at the given path.
func (self *Node) SetByPath(sub Node, path ...Path) (exist bool, err error) {
	if len(path) == 0 {
		*self = sub
		return true, nil
	}
	if sub.IsError() {
		return false, sub
	}
	if err := self.Check(); err != nil {
		return false, err
	}
	// search source node by path
	var v = self.GetByPath(path...)
	if v.IsError() {
		if !v.isErrNotFoundLast() {
			return false, v
		}
		if err := v.setNotFound(path[len(path)-1], &sub); err != nil {
			return false, err
		}
	} else {
		exist = true
	}
	err = self.replace(v, sub)
	return
}

var pnsPool = sync.Pool{
	New: func() interface{} {
		return &pnSlice{
			a: make([]PathNode, 0, DefaultNodeSliceCap),
		}
	},
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

// SetMany searches longitudinally and sets the sub nodes at the given pathes.
func (self *Node) SetMany(pathes []PathNode, opts *Options) (err error) {
	if len(pathes) == 0 {
		return nil
	}
	if err := self.Check(); err != nil {
		return err
	}

	// copy original pathes
	ps := pnsPool.Get().(*pnSlice)
	if cap(ps.a) < len(pathes) {
		ps.a = make([]PathNode, len(pathes))
	} else {
		ps.a = ps.a[:len(pathes)]
	}
	copy(ps.a, pathes)
	ps.b = pathes

	// get original values
	if err = self.getMany(ps.a, true, opts); err != nil {
		goto ret
	}
	// handle not found values
	for i, a := range ps.a {
		if a.IsEmpty() {
			var sp unsafe.Pointer
			switch self.t {
			case thrift.STRUCT:
				sp = self.v
			case thrift.LIST, thrift.SET:
				// element type (1 B) + size (4B)
				sp = rt.AddPtr(self.v, 5)
			case thrift.MAP:
				// element type (1 B) + size (4B)
				sp = rt.AddPtr(self.v, 6)
			}
			ps.a[i].Node = errNotFoundLast(sp, self.t)
			ps.a[i].Node.setNotFound(a.Path, &ps.b[i].Node)
		}
	}

	err = self.replaceMany(ps)
ret:
	ps.b = nil
	pnsPool.Put(ps)
	return
}

func (self *Node) deleteChild(path Path) Node {
	p := thrift.BinaryProtocol{}
	p.Buf = self.raw()
	var s, e int
	var tt thrift.Type
	switch self.t {
	case thrift.STRUCT:
		if path.Type() != PathFieldId {
			return errNode(meta.ErrDismatchType, "", nil)
		}
		id := path.id()
		for {
			s = p.Read
			_, t, i, err := p.ReadFieldBegin()
			if err != nil {
				return errNode(meta.ErrRead, "", err)
			}
			if t == thrift.STOP {
				return errNotFound
			}
			if err := p.Skip(t, UseNativeSkipForGet); err != nil {
				return errNode(meta.ErrRead, "", err)
			}
			e = p.Read
			if id == i {
				tt = t
				break
			}
		}
	case thrift.LIST, thrift.SET:
		if path.Type() != PathIndex {
			return errNode(meta.ErrDismatchType, "", nil)
		}
		id := path.int()
		et, size, err := p.ReadListBegin()
		if err != nil {
			return errNode(meta.ErrRead, "", err)
		}
		if id >= size {
			return errNotFound
		}
		tt = et
		if err := p.ModifyI32(p.Read-4, int32(size-1)); err != nil {
			return errNode(meta.ErrWrite, "", err)
		}
		if d := thrift.TypeSize(et); d > 0 {
			s = p.Read + d*id
			e = s + d
		} else {
			for i := 0; i < id; i++ {
				if err := p.Skip(et, UseNativeSkipForGet); err != nil {
					return errNode(meta.ErrRead, "", err)
				}
			}
			s = p.Read
			if err := p.Skip(et, UseNativeSkipForGet); err != nil {
				return errNode(meta.ErrRead, "", err)
			}
			e = p.Read
		}
	case thrift.MAP:
		kt, et, size, err := p.ReadMapBegin()
		if err != nil {
			return errNode(meta.ErrRead, "", err)
		}
		id := path.ToRaw(kt)
		if id == nil {
			return errNode(meta.ErrInvalidParam, "", nil)
		}
		if err := p.ModifyI32(p.Read-4, int32(size-1)); err != nil {
			return errNode(meta.ErrWrite, "", err)
		}
		tt = et
		for i := 0; i < size; i++ {
			s = p.Read
			if err := p.Skip(kt, UseNativeSkipForGet); err != nil {
				return errNode(meta.ErrRead, "", err)
			}
			key := p.Buf[s:p.Read]
			if err := p.Skip(et, UseNativeSkipForGet); err != nil {
				return errNode(meta.ErrRead, "", err)
			}
			e = p.Read
			if bytes.Equal(key, id) {
				break
			}
		}
	}

	return Node{
		t: tt,
		v: rt.AddPtr(self.v, uintptr(s)),
		l: e - s,
	}
}

func (self *Node) UnsetByPath(path ...Path) error {
	l := len(path)
	if l == 0 {
		*self = Node{}
		return nil
	}
	if err := self.Check(); err != nil {
		return err
	}
	// search parent node by path
	var v = self.GetByPath(path[:l-1]...)
	if v.IsError() {
		if v.IsErrNotFound() {
			return nil
		}
		return v
	}
	ret := v.deleteChild(path[l-1])
	if ret.IsError() {
		return ret
	}
	return self.replace(ret, Node{t: ret.t})
}

// Children loads all its children and children's children recursively (if recurse is true).
// out is used for store children, and it is always reset to zero length before use.
//
// NOTICE: if opts.NotScanParentNode is true, the parent nodes (PathNode.Node) of complex (LIST/SET/MAP/STRUCT) type won't be assgined data
func (self Node) Children(out *[]PathNode, recurse bool, opts *Options) (err error) {
	if self.Error() != "" {
		return self
	}
	if out == nil {
		panic("out is nil")
	}
	// NOTICE: since we only use out for memory reuse, we always reset it to zero.
	var p = thrift.BinaryProtocol{
		Buf: self.raw(),
	}
	var tree = PathNode{
		Node: self,
		Next: (*out)[:0], // NOTICE: we reset it to zero.
	}

	err = tree.scanChildren(&p, recurse, opts)
	if err == nil {
		*out = tree.Next
	}
	return err
}
