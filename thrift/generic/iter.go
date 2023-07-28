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

	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
)

var (
	_SkipMaxDepth = types.TB_SKIP_STACK_SIZE - 1
)

func (self Node) iterFields() (fi structIterator) {
	// fi = structIteratorPool.Get().(*StructIterator)
	fi.p.Buf = self.raw()
	// if _, err := fi.p.ReadStructBegin(); err != nil {
	// 	fi.Err = meta.NewError(meta.ErrReadInput, "", err)
	// 	return
	// }
	return
}

func (self Node) iterPairs() (fi mapIterator) {
	buf := self.raw()
	// fi = pairIteratorPool.Get().(*PairIterator)
	fi.p.Buf = buf
	if kt, et, size, err := fi.p.ReadMapBegin(); err != nil {
		fi.Err = meta.NewError(meta.ErrRead, "", err)
		return
	} else {
		fi.size = size
		fi.et = thrift.Type(et)
		fi.kt = thrift.Type(kt)
	}
	return
}

func (self Node) iterElems() (fi listIterator) {
	buf := self.raw()
	// fi = listIteratorPool.Get().(*ListIterator)
	fi.p.Buf = buf
	if et, size, err := fi.p.ReadListBegin(); err != nil {
		fi.Err = meta.NewError(meta.ErrRead, "", err)
		return
	} else {
		fi.size = size
		fi.et = thrift.Type(et)
	}
	return
}

type structIterator struct {
	Err error
	p   thrift.BinaryProtocol
}

func (it structIterator) HasNext() bool {
	return it.Err == nil && it.p.Left() > 0 && (it.p.Buf)[it.p.Read] != byte(thrift.STOP)
}

func (it *structIterator) Next(useNative bool) (id thrift.FieldID, typ thrift.Type, start int, end int) {
	_, typeId, ID, err := it.p.ReadFieldBegin()
	if err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}
	if typeId == thrift.STOP {
		return
	}
	typ = thrift.Type(typeId)
	start = it.p.Read
	err = it.p.Skip(typeId, useNative)
	if err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}
	end = it.p.Read
	err = it.p.ReadFieldEnd()
	if err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}
	id = thrift.FieldID(ID)
	return
}

type listIterator struct {
	Err  error
	size int
	k    int
	et   thrift.Type
	p    thrift.BinaryProtocol
}

func (it listIterator) HasNext() bool {
	return it.Err == nil && it.p.Left() > 0 && it.k < it.size
}

func (it listIterator) Size() int {
	return it.size
}

func (it listIterator) Pos() int {
	return it.k
}

func (it *listIterator) Next(useNative bool) (start int, end int) {
	start = it.p.Read
	err := it.p.Skip(it.et, useNative)
	if err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}
	end = it.p.Read
	it.k++
	return
}

type mapIterator struct {
	Err  error
	size int
	i    int
	kt   thrift.Type
	et   thrift.Type
	p    thrift.BinaryProtocol
}

func (it mapIterator) HasNext() bool {
	return it.Err == nil && it.p.Left() > 0 && it.i < it.size
}

func (it mapIterator) Size() int {
	return it.size
}

func (it mapIterator) Pos() int {
	return it.i
}

func (it *mapIterator) NextBin(useNative bool) (keyStart int, keyBin []byte, start int, end int) {
	keyStart = it.p.Read
	var err error
	ks := it.p.Read
	if err = it.p.Skip(it.kt, useNative); err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}
	keyBin = it.p.Buf[ks:it.p.Read]
	start = it.p.Read
	err = it.p.Skip(it.et, useNative)
	if err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}
	end = it.p.Read
	it.i++
	return
}

func (it *mapIterator) NextStr(useNative bool) (keyStart int, keyString string, start int, end int) {
	keyStart = it.p.Read
	var err error
	if it.kt == thrift.STRING {
		keyString, err = it.p.ReadString(false)
		if err != nil {
			it.Err = meta.NewError(meta.ErrRead, "", err)
			return
		}
	} else {
		it.Err = wrapError(meta.ErrUnsupportedType, "MapIterator.nextStr: key type is not string", nil)
		return
	}
	start = it.p.Read
	err = it.p.Skip(it.et,
		 useNative)
	if err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}
	end = it.p.Read
	it.i++
	return
}

func (it *mapIterator) NextInt(useNative bool) (keyStart int, keyInt int, start int, end int) {
	keyStart = it.p.Read
	switch it.kt {
	case thrift.I08:
		n, err := it.p.ReadByte()
		keyInt = int(n)
		if err != nil {
			it.Err = meta.NewError(meta.ErrRead, "", err)
			return
		}
	case thrift.I16:
		n, err := it.p.ReadI16()
		if err != nil {
			it.Err = meta.NewError(meta.ErrRead, "", err)
			return
		}
		keyInt = int(n)
	case thrift.I32:
		n, err := it.p.ReadI32()
		if err != nil {
			it.Err = meta.NewError(meta.ErrRead, "", err)
			return
		}
		keyInt = int(n)
	case thrift.I64:
		n, err := it.p.ReadI64()
		if err != nil {
			it.Err = meta.NewError(meta.ErrRead, "", err)
			return
		}
		keyInt = int(n)
	default:
		it.Err = wrapError(meta.ErrUnsupportedType, "MapIterator.nextInt: key type is not int", nil)
	}
	start = it.p.Read
	if err := it.p.Skip(it.et,
		 useNative); err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}
	end = it.p.Read
	it.i++
	return
}

// Foreach scan each element of a complex type (LIST/SET/MAP/STRUCT),
// and call handler sequentially with corresponding path and value
func (self Value) Foreach(handler func(path Path, val Value) bool, opts *Options) error {
	switch self.t {
	case thrift.STRUCT:
		it := self.iterFields()
		if it.Err != nil {
			return it.Err
		}
		for it.HasNext() {
			id, typ, start, end := it.Next(opts.UseNativeSkip)
			if it.Err != nil {
				return it.Err
			}
			f := self.Desc.Struct().FieldById(id)
			if f == nil {
				if opts.DisallowUnknow {
					return wrapError(meta.ErrUnknownField, fmt.Sprintf("unknown field %d", id), nil)
				}
				continue
			}
			if f.Type().Type() != typ {
				return wrapError(meta.ErrDismatchType, fmt.Sprintf("field %d type %v mismatch read type %v", id, f.Type().Type(), typ), nil)
			}
			v := self.slice(start, end, f.Type())
			var p Path
			if opts.IterateStructByName && f != nil {
				p = NewPathFieldName(f.Name())
			} else {
				p = NewPathFieldId(id)
			}
			if !handler(p, v) {
				return nil
			}
		}
	case thrift.LIST, thrift.SET:
		it := self.iterElems()
		if it.Err != nil {
			return it.Err
		}
		et := self.Desc.Elem()
		for i := 0; it.HasNext(); i++ {
			start, end := it.Next(opts.UseNativeSkip)
			if it.Err != nil {
				return it.Err
			}
			v := self.slice(start, end, et)
			p := NewPathIndex(i)
			if !handler(p, v) {
				return nil
			}
		}
	case thrift.MAP:
		it := self.iterPairs()
		if it.Err != nil {
			return it.Err
		}
		kt := self.Desc.Key()
		if kt.Type() != it.kt {
			return wrapError(meta.ErrDismatchType, "dismatched descriptor key type and binary type", nil)
		}
		et := self.Desc.Elem()
		if et.Type() != it.et {
			return wrapError(meta.ErrDismatchType, "dismatched descriptor elem type and binary type", nil)
		}
		if kt.Type() == thrift.STRING {
			for i := 0; it.HasNext(); i++ {
				_, keyString, start, end := it.NextStr(opts.UseNativeSkip)
				if it.Err != nil {
					return it.Err
				}
				v := self.slice(start, end, et)
				p := NewPathStrKey(keyString)
				if !handler(p, v) {
					return nil
				}
			}
		} else if kt.Type().IsInt() {
			for i := 0; it.HasNext(); i++ {
				_, keyInt, start, end := it.NextInt(opts.UseNativeSkip)
				if it.Err != nil {
					return it.Err
				}
				v := self.slice(start, end, et)
				p := NewPathIntKey(keyInt)
				if !handler(p, v) {
					return nil
				}
			}
		} else {
			for i := 0; it.HasNext(); i++ {
				_, keyBin, start, end := it.NextBin(opts.UseNativeSkip)
				if it.Err != nil {
					return it.Err
				}
				v := self.slice(start, end, et)
				p := NewPathBinKey(keyBin)
				if !handler(p, v) {
					return nil
				}
			}
		}
	}
	return nil
}

// ForeachKV scan each element of a MAP type, and call handler sequentially with corresponding key and value
func (self Value) ForeachKV(handler func(key Value, val Value) bool, opts *Options) error {
	if err := self.should("MAP", thrift.MAP); err != "" {
		return wrapError(meta.ErrUnsupportedType, err, nil)
	}
	switch self.t {
	case thrift.MAP:
		p := thrift.BinaryProtocol{Buf: self.raw()}
		kt, et, size, e := p.ReadMapBegin()
		if e != nil {
			return errNode(meta.ErrRead, "", nil)
		}
		kd := self.Desc.Key()
		if kd.Type() != kt {
			return wrapError(meta.ErrDismatchType, "dismatched descriptor key type and binary type", nil)
		}
		ed := self.Desc.Elem()
		if ed.Type() != et {
			return wrapError(meta.ErrDismatchType, "dismatched descriptor elem type and binary type", nil)
		}
		for i := 0; i < size; i++ {
			ks := p.Read
			if err := p.Skip(kt, opts.UseNativeSkip); err != nil {
				return errNode(meta.ErrRead, "", nil)
			}
			key := self.slice(ks, p.Read, kd)
			es := p.Read
			if err := p.Skip(et, opts.UseNativeSkip); err != nil {
				return errNode(meta.ErrRead, "", nil)
			}
			val := self.slice(es, p.Read, ed)
			if !handler(key, val) {
				return nil
			}
		}
		return nil
	default:
		return errNode(meta.ErrUnsupportedType, "", nil)
	}
}

// Foreach scan each element of a complex type (LIST/SET/MAP/STRUCT),
// and call handler sequentially with corresponding path and node
func (self Node) Foreach(handler func(path Path, node Node) bool, opts *Options) error {
	switch self.t {
	case thrift.STRUCT:
		it := self.iterFields()
		if it.Err != nil {
			return it.Err
		}
		for it.HasNext() {
			id, typ, start, end := it.Next(opts.UseNativeSkip)
			if it.Err != nil {
				return it.Err
			}
			v := self.slice(start, end, typ)
			var p Path
			p = NewPathFieldId(id)
			if !handler(p, v) {
				return nil
			}
		}
	case thrift.LIST, thrift.SET:
		it := self.iterElems()
		if it.Err != nil {
			return it.Err
		}
		for i := 0; it.HasNext(); i++ {
			start, end := it.Next(opts.UseNativeSkip)
			if it.Err != nil {
				return it.Err
			}
			v := self.slice(start, end, self.et)
			p := NewPathIndex(i)
			if !handler(p, v) {
				return nil
			}
		}
	case thrift.MAP:
		it := self.iterPairs()
		if it.Err != nil {
			return it.Err
		}
		kt := self.kt
		if kt == thrift.STRING {
			for i := 0; it.HasNext(); i++ {
				_, keyString, start, end := it.NextStr(opts.UseNativeSkip)
				if it.Err != nil {
					return it.Err
				}
				v := self.slice(start, end, self.et)
				p := NewPathStrKey(keyString)
				if !handler(p, v) {
					return nil
				}
			}
		} else if kt.IsInt() {
			for i := 0; it.HasNext(); i++ {
				_, keyInt, start, end := it.NextInt(opts.UseNativeSkip)
				if it.Err != nil {
					return it.Err
				}
				v := self.slice(start, end, self.et)
				p := NewPathIntKey(keyInt)
				if !handler(p, v) {
					return nil
				}
			}
		} else {
			for i := 0; it.HasNext(); i++ {
				_, keyBin, start, end := it.NextBin(opts.UseNativeSkip)
				if it.Err != nil {
					return it.Err
				}
				v := self.slice(start, end, self.et)
				p := NewPathBinKey(keyBin)
				if !handler(p, v) {
					return nil
				}
			}
		}
	}
	return nil
}

// ForeachKV scan each element of a MAP type, and call handler sequentially with corresponding key and value
func (self Node) ForeachKV(handler func(key Node, val Node) bool, opts *Options) error {
	switch self.t {
	case thrift.MAP:
		p := thrift.BinaryProtocol{Buf: self.raw()}
		kt, et, size, e := p.ReadMapBegin()
		if e != nil {
			return errNode(meta.ErrRead, "", nil)
		}
		for i := 0; i < size; i++ {
			ks := p.Read
			if err := p.Skip(kt, opts.UseNativeSkip); err != nil {
				return errNode(meta.ErrRead, "", nil)
			}
			key := self.slice(ks, p.Read, kt)
			es := p.Read
			if err := p.Skip(et, opts.UseNativeSkip); err != nil {
				return errNode(meta.ErrRead, "", nil)
			}
			val := self.slice(es, p.Read, et)
			if !handler(key, val) {
				return nil
			}
		}
		return nil
	default:
		return errNode(meta.ErrUnsupportedType, "", nil)
	}
}
