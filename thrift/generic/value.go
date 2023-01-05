/**
 * Copyright 2022 CloudWeGo Authors.
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
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
)

// Value is a generic API wrapper for operations on thrift data.
// Node is the underlying raw bytes.
// Desc is the corresponding type descriptor.
type Value struct {
	Node
	Desc *thrift.TypeDescriptor
}

// NewValue creates a new Value from a raw byte slice.
func NewValue(desc *thrift.TypeDescriptor, src []byte) Value {
	return Value{
		Node: NewNode(desc.Type(), src),
		Desc: desc,
	}
}

// NewValueFromNode copy both Node and TypeDescriptor from another Value.
func (self Value) Fork() Value {
	ret := self
	ret.Node = self.Node.Fork()
	return ret
}

func (self Value) slice(s int, e int, desc *thrift.TypeDescriptor) Value {
	t := desc.Type()
	ret := Value{
		Node: Node{
			t: t,
			l: (e - s),
			v: rt.AddPtr(self.v, uintptr(s)),
		},
		Desc: desc,
	}
	if t == thrift.LIST || t == thrift.SET {
		ret.et = desc.Elem().Type()
	} else if t == thrift.MAP {
		ret.kt = desc.Key().Type()
		ret.et = desc.Elem().Type()
	}
	return ret
}

func searchFieldName(p *thrift.BinaryProtocol, id string, f *thrift.FieldDescriptor) (tt thrift.Type, start int, err error) {
	// if _, err := p.ReadStructBegin(); err != nil {
	// 	return 0, start, wrapError(meta.ErrReadInput, "", err)
	// }
	for {
		_, t, i, err := p.ReadFieldBegin()
		if err != nil {
			return 0, start, wrapError(meta.ErrRead, "", err)
		}
		if t == thrift.STOP {
			return thrift.STRUCT, start, errNotFound
		}
		if f.ID() == thrift.FieldID(i) {
			// if t != f.Type().Type() {
			// 	return 0, start, wrapError(meta.ErrDismatchType, fmt.Sprintf("field '%s' expects type %s, buf got type %s", f.Name(), f.Type().Type(), t), nil)
			// }
			start = p.Read
			tt = t
			break
		}
		if err := p.Skip(t, _SkipMaxDepth, UseNativeSkipForGet); err != nil {
			return thrift.STRUCT, start, wrapError(meta.ErrRead, "", err)
		}
	}
	return
}

// GetByPath searches longitudinally and return a sub value at the given path from the value.
//
// The path is a list of PathFieldId, PathFieldName, PathIndex, PathStrKey, PathBinKey, PathIntKey,
// Each path MUST be a valid path type for current layer (e.g. PathFieldId is only valid for STRUCT).
// Empty path will return the current value.
func (self Value) GetByPath(pathes ...Path) Value {
	if self.Error() != "" {
		return self
	}

	p := thrift.BinaryProtocol{
		Buf: self.raw(),
	}
	start := 0
	desc := self.Desc
	tt := self.t
	isList := tt == thrift.LIST
	var err error

	for i, path := range pathes {
		switch path.t {
		case PathFieldId:
			id := path.id()
			tt, start, err = searchFieldId(&p, id)
			desc = desc.Struct().FieldById(id).Type()
			isList = tt == thrift.LIST
		case PathFieldName:
			id := path.str()
			f := desc.Struct().FieldByKey(id)
			if f == nil {
				return errValue(meta.ErrUnknownField, fmt.Sprintf("field name '%s' is not defined in IDL", id), nil)
			}
			tt, start, err = searchFieldName(&p, id, f)
			desc = f.Type()
			isList = tt == thrift.LIST
		case PathIndex:
			tt, start, err = searchIndex(&p, path.int(), isList)
			desc = desc.Elem()
		case PathStrKey:
			tt, start, err = searchStrKey(&p, path.str())
			desc = desc.Elem()
		case PathIntKey:
			tt, start, err = searchIntKey(&p, path.int())
			desc = desc.Elem()
		case PathBinKey:
			tt, start, err = searchBinKey(&p, path.bin())
			desc = desc.Elem()
		default:
			return errValue(meta.ErrUnsupportedType, fmt.Sprintf("invalid %dth path: %#v", i, p), nil)
		}
		if err != nil {
			// the last one not foud, return start pointer for subsequently inserting operation on `SetByPath()`
			if i == len(pathes)-1 && err == errNotFound {
				return Value{errNotFoundLast(unsafe.Pointer(uintptr(self.v)+uintptr(start)), tt), nil}
			}
			en := err.(Node)
			return errValue(en.ErrCode().Behavior(), "", err)
		}
	}

	if err := p.Skip(desc.Type(), _SkipMaxDepth, UseNativeSkipForGet); err != nil {
		return errValue(meta.ErrRead, "", err)
	}
	return self.slice(start, p.Read, desc)
}

// Field returns a sub node at the given field id from a STRUCT value.
func (self Value) Field(id thrift.FieldID) (v Value) {
	n := self.Node.Field(id)
	if n.IsError() {
		return wrapValue(n, nil)
	}

	f := self.Desc.Struct().FieldById(id)
	if f == nil {
		return errValue(meta.ErrUnknownField, fmt.Sprintf("field id %d is not defined in IDL", id), nil)
	}

	return wrapValue(n, f.Type())
}

// func (self Value) Fields(ids []PathNode, opts *Options) (err error) {
// 	if err := self.should("Fields", thrift.STRUCT); err != "" {
// 		return errValue(meta.ErrUnsupportedType, err, nil)
// 	}

// 	it := self.iterFields()
// 	if it.Err != nil {
// 		return errValue(meta.ErrReadInput, "", it.Err)
// 	}

// 	need := len(ids)
// 	for count := 0; it.HasNext() && count < need; {
// 		var p *PathNode
// 		i, t, s, e := it.Next(opts.UseNativeSkip)
// 		if it.Err != nil {
// 			err = errValue(meta.ErrReadInput, "", it.Err)
// 			goto ret
// 		}
// 		f := self.Desc.Struct().FieldById(i)
// 		if f == nil {
// 			if opts.DisallowUnknow {
// 				err = errValue(meta.ErrUnknownField, fmt.Sprintf("field id %d is not defined in IDL", i), nil)
// 				goto ret
// 			}
// 			continue

// 		}
// 		if t != f.Type().Type() {
// 			err = errValue(meta.ErrDismatchType, fmt.Sprintf("field '%s' expects type %s, buf got type %s", f.Name(), f.Type().Type(), t), nil)
// 			goto ret
// 		}
// 		//TODO: use bitmap to avoid repeatedly scan
// 		for j, id := range ids {
// 			if (id.Path.t == PathFieldId && id.Path.Id() == i) || (id.Path.t == PathFieldName && id.Path.Str() == f.Name()) {
// 				p = &ids[j]
// 				count += 1
// 				break
// 			}
// 		}
// 		if p == nil {
// 			continue
// 		}
// 		p.Node = self.Node.slice(s, e, f.Type().Type())
// 	}

// ret:
// 	// it.Recycle()
// 	return
// }

// FieldByName returns a sub node at the given field name from a STRUCT value.
func (self Value) FieldByName(name string) (v Value) {
	if err := self.should("FieldByName", thrift.STRUCT); err != "" {
		return errValue(meta.ErrUnsupportedType, err, nil)
	}

	f := self.Desc.Struct().FieldByKey(name)
	if f == nil {
		return errValue(meta.ErrUnknownField, fmt.Sprintf("field '%s' is not defined in IDL", name), nil)
	}

	// not found, try to scan the whole bytes
	it := self.iterFields()
	if it.Err != nil {
		return errValue(meta.ErrRead, "", it.Err)
	}
	for it.HasNext() {
		i, t, s, e := it.Next(UseNativeSkipForGet)
		if i == f.ID() {
			if t != f.Type().Type() {
				v = errValue(meta.ErrDismatchType, fmt.Sprintf("field '%s' expects type %s, buf got type %s", f.Name(), f.Type().Type(), t), nil)
				goto ret
			}
			v = self.slice(s, e, f.Type())
			goto ret
		} else if it.Err != nil {
			v = errValue(meta.ErrRead, "", it.Err)
			goto ret
		}
	}

	v = errValue(meta.ErrNotFound, fmt.Sprintf("field '%s' is not found in this value", name), errNotFound)
ret:
	// it.Recycle()
	return
}

// Index returns a sub node at the given index from a LIST value.
func (self Value) Index(i int) (v Value) {
	n := self.Node.Index(i)
	if n.IsError() {
		return wrapValue(n, nil)
	}
	return wrapValue(n, self.Desc.Elem())
}

// GetByStr returns a sub node at the given string key from a MAP value.
func (self Value) GetByStr(key string) (v Value) {
	n := self.Node.GetByStr(key)
	if n.IsError() {
		return wrapValue(n, nil)
	}
	return wrapValue(n, self.Desc.Elem())
}

// GetByInt returns a sub node at the given int key from a MAP value.
func (self Value) GetByInt(key int) (v Value) {
	n := self.Node.GetByInt(key)
	if n.IsError() {
		return wrapValue(n, nil)
	}
	return wrapValue(n, self.Desc.Elem())
}

// func (self Value) GetMany(pathes []PathNode, opts *Options) error {
// 	if len(pathes) == 0 {
// 		return nil
// 	}
// 	p := pathes[0]
// 	switch p.Path.t {
// 	case PathFieldId, PathFieldName:
// 		return self.Fields(pathes, opts)
// 	case PathIndex:
// 		return self.Indexes(pathes, opts)
// 	case PathStrKey, PathIntKey:
// 		return self.Gets(pathes, opts)
// 	case PathObjKey:
// 		panic("not implement yet")
// 	default:
// 		return errValue(meta.ErrUnsupportedType, fmt.Sprintf("invalid path: %#v", p), nil)
// 	}
// }

// SetByPath searches longitudinally and sets a sub value at the given path from the value.
// exist tells whether the node is already exists.
func (self *Value) SetByPath(sub Value, path ...Path) (exist bool, err error) {
	l := len(path)
	if l == 0 {
		*self = sub
		return true, nil
	}
	if err := self.Check(); err != nil {
		return false, err
	}
	if sub.Error() != "" {
		return false, meta.NewError(meta.ErrInvalidParam, "given node is invalid", sub)
	}

	// search source node by path
	v := self.GetByPath(path...)
	if v.IsError() {
		if !v.isErrNotFoundLast() {
			return false, v
		}
		p := path[l-1]
		if p.t == PathFieldName {
			desc, err := GetDescByPath(self.Desc, path[:l-1]...)
			if err != nil {
				return false, err
			}
			f := desc.Struct().FieldByKey(p.str())
			p = NewPathFieldId(f.ID())
		}
		if err := v.setNotFound(p, &sub.Node); err != nil {
			return false, err
		}
	} else {
		exist = true
	}
	err = self.replace(v.Node, sub.Node)
	return
}

// UnsetByPath searches longitudinally and unsets a sub value at the given path from the value.
func (self *Value) UnsetByPath(path ...Path) error {
	l := len(path)
	if l == 0 {
		*self = Value{}
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
	p := path[l-1]
	if p.t == PathFieldName {
		desc, err := GetDescByPath(self.Desc, path[:l-1]...)
		if err != nil {
			return err
		}
		f := desc.Struct().FieldByKey(p.str())
		p = NewPathFieldId(f.ID())
	}
	ret := v.deleteChild(p)
	if ret.IsError() {
		return ret
	}
	return self.replace(ret, Node{t: ret.t})
}

// var pathesPool = sync.Pool{
// 	New: func() interface{} {
// 		return &pathSlice{
// 			a: make([]PathNode, 0, defaultNodeSliceCap),
// 		}
// 	},
// }

// func (self *Value) SetMany(pathes []PathNode, opts *Options) (err error) {
// 	if len(pathes) == 0 {
// 		return nil
// 	}
// 	if err := self.Check(); err != nil {
// 		return err
// 	}

// 	// copy original pathes
// 	ps := pnsPool.Get().(*pnSlice)
// 	if cap(ps.a) < len(pathes) {
// 		ps.a = make([]PathNode, len(pathes))
// 	} else {
// 		ps.a = ps.a[:len(pathes)]
// 	}
// 	copy(ps.a, pathes)
// 	ps.b = pathes

// 	// get original values
// 	if err = self.GetMany(ps.a, opts); err != nil {
// 		goto ret
// 	}

// 	err = self.replaceMany(ps)
// ret:
// 	ps.b = nil
// 	pnsPool.Put(ps)
// 	return
// }

// MarshalTo marshals self value into a sub value descripted by the to descriptor, alse called as "Cutting".
// Usually, the to descriptor is a subset of self descriptor.
func (self Value) MarshalTo(to *thrift.TypeDescriptor, opts *Options) ([]byte, error) {
	var w = thrift.NewBinaryProtocolBuffer()
	var r = thrift.BinaryProtocol{}
	r.Buf = self.raw()
	var from = self.Desc
	if from.Type() != to.Type() {
		return nil, wrapError(meta.ErrDismatchType, "to descriptor dismatches from descriptor", nil)
	}
	if err := marshalTo(&r, w, from, to, opts); err != nil {
		return nil, err
	}
	ret := make([]byte, len(w.Buf))
	copy(ret, w.Buf)
	thrift.FreeBinaryProtocolBuffer(w)
	return ret, nil
}

func marshalTo(read *thrift.BinaryProtocol, write *thrift.BinaryProtocol, from *thrift.TypeDescriptor, to *thrift.TypeDescriptor, opts *Options) error {
	switch t := to.Type(); t {
	case thrift.STRUCT:
		if from == to {
			return nil
		}
		var req *thrift.RequiresBitmap
		if !opts.NotCheckRequireNess {
			req = thrift.NewRequiresBitmap()
			to.Struct().Requires().CopyTo(req)
		}

		for {
			_, tt, i, err := read.ReadFieldBegin()
			if err != nil {
				return wrapError(meta.ErrRead, "", err)
			}

			if tt == thrift.STOP {
				if !opts.NotCheckRequireNess {
					if err := handleUnsets(*req, to.Struct(), write, opts.WriteDefault); err != nil {
						return err
					}
				}
				if err := write.WriteStructEnd(); err != nil {
					return wrapError(meta.ErrWrite, "", err)
				}
				break
			}

			ff := from.Struct().FieldById(thrift.FieldID(i))
			if ff == nil {
				if opts.DisallowUnknow {
					return wrapError(meta.ErrUnknownField, fmt.Sprintf("unknown field %d", i), nil)
				} else {
					// if not set, skip to the next field
					if err := read.Skip(tt, _SkipMaxDepth, opts.UseNativeSkip); err != nil {
						return wrapError(meta.ErrRead, "", err)
					}
					continue
				}
			}

			if tt != ff.Type().Type() {
				return wrapError(meta.ErrDismatchType, fmt.Sprintf("field %d expects type %s, buf got type %s", ff.ID(), ff.Type().Type(), thrift.Type(tt)), nil)
			}
			// find corresponding field in the 'to' descriptor
			// TODO: use field name to find the field
			// var tf *thrift.FieldDescriptor
			// if opts.FieldByName {
			// 	tf = to.Struct().FieldByKey(ff.Name())
			// } else {
			tf := to.Struct().FieldById(thrift.FieldID(i))
			// }
			if tf == nil {
				// if not set, skip to the next field
				if err := read.Skip(tt, _SkipMaxDepth, opts.UseNativeSkip); err != nil {
					return wrapError(meta.ErrRead, "", err)
				}
				continue
			}
			// if tt != tf.Type().Type() {
			// 	return wrapError(meta.ErrDismatchType, fmt.Sprintf("field %d expects type %s, buf got type %s", tf.ID(), tf.Type().Type(), thrift.Type(tt)), nil)
			// }
			if err := write.WriteFieldBegin(tf.Name(), tt, i); err != nil {
				return wrapError(meta.ErrWrite, "", err)
			}

			if !opts.NotCheckRequireNess {
				req.Set(tf.ID(), thrift.OptionalRequireness)
			}

			if err := marshalTo(read, write, ff.Type(), tf.Type(), opts); err != nil {
				return err
			}
		}
		if !opts.NotCheckRequireNess {
			thrift.FreeRequiresBitmap(req)
		}
		return nil
	case thrift.LIST, thrift.SET:
		if from.Type() != thrift.LIST && from.Type() != thrift.SET {
			return wrapError(meta.ErrDismatchType, fmt.Sprintf("expect type LIST or SET, buf got type %s", from.Type()), nil)
		}
		if to.Elem() == from.Elem() {
			goto skip_val
		}
		et, size, err := read.ReadListBegin()
		if err != nil {
			return wrapError(meta.ErrRead, "", err)
		}
		// if et != to.Elem().Type() || et != from.Elem().Type() {
		// 	return wrapError(meta.ErrDismatchType, fmt.Sprintf("expect elem type %s, buf got type %s", from.Elem().Type(), et), nil)
		// }
		if err := write.WriteListBegin(et, size); err != nil {
			return wrapError(meta.ErrWrite, "", err)
		}
		if size == 0 {
			return nil
		}
		for i := 0; i < size; i++ {
			if err := marshalTo(read, write, from.Elem(), to.Elem(), opts); err != nil {
				return err
			}
		}
		return nil
	case thrift.MAP:
		if from.Type() != thrift.MAP {
			return wrapError(meta.ErrDismatchType, fmt.Sprintf("expect type MAP, buf got type %s", from.Type()), nil)
		}
		if to.Elem() == from.Elem() && to.Key() == from.Key() {
			goto skip_val
		}
		kt, et, size, err := read.ReadMapBegin()
		if err != nil {
			return wrapError(meta.ErrRead, "", err)
		}
		// if et != to.Elem().Type() || et != from.Elem().Type() {
		// 	return wrapError(meta.ErrDismatchType, fmt.Sprintf("expect elem type %s, buf got type %s", from.Elem().Type(), et), nil)
		// }
		// if kt != to.Key().Type() || kt != from.Key().Type() {
		// 	return wrapError(meta.ErrDismatchType, fmt.Sprintf("expect key type %s, buf got type %s", from.Key().Type(), kt), nil)
		// }
		if err := write.WriteMapBegin(kt, et, size); err != nil {
			return wrapError(meta.ErrWrite, "", err)
		}
		if size == 0 {
			return nil
		}
		for i := 0; i < size; i++ {
			if err := marshalTo(read, write, from.Key(), to.Key(), opts); err != nil {
				return err
			}
			if err := marshalTo(read, write, from.Elem(), to.Elem(), opts); err != nil {
				return err
			}
		}
		return nil
	}

skip_val:
	if to.Type() != from.Type() {
		return meta.NewError(meta.ErrDismatchType, "to descriptor dismatches from descriptor", nil)
	}
	e := read.Read
	if err := read.Skip(to.Type(), _SkipMaxDepth, opts.UseNativeSkip); err != nil {
		return wrapError(meta.ErrRead, "", err)
	}
	s := read.Read
	write.Buf = append(write.Buf, read.Buf[e:s]...)
	return nil
}

func handleUnsets(req thrift.RequiresBitmap, desc *thrift.StructDescriptor, p *thrift.BinaryProtocol, writeDefault bool) error {
	return req.CheckRequires(desc, writeDefault, func(f *thrift.FieldDescriptor) error {
		if e := p.WriteFieldBegin(f.Name(), f.Type().Type(), (f.ID())); e != nil {
			return e
		}
		return p.WriteEmpty(f.Type())
	})
}
