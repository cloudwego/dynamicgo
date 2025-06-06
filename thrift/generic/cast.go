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
	// "unicode/utf8"
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
)

func (self Node) should(api string, t1 thrift.Type) string {
	if self.t == thrift.STOP || self.t == thrift.ERROR {
		return self.Error()
	}
	if self.t == t1 {
		return ""
	}
	return fmt.Sprintf("API `%s` only supports %+v type", api, t1)
}

func (self Node) should2(api string, t1 thrift.Type, t2 thrift.Type) string {
	if self.t == thrift.STOP || self.t == thrift.ERROR {
		return self.Error()
	}
	if self.t == t1 || self.t == t2 {
		return ""
	}
	return fmt.Sprintf("API `%s` only supports %s or %s type", api, t1, t2)
}

// Len returns the element count of container-kind type (LIST/SET/MAP)
func (self Node) Len() (int, error) {
	if self.IsError() {
		return 0, self
	}
	return self.len()
}

func (self Node) len() (int, error) {
	switch self.t {
	case thrift.LIST, thrift.SET:
		b := rt.BytesFrom(unsafe.Pointer(uintptr(self.v)+uintptr(1)), 4, 4)
		return int(thrift.BinaryEncoding{}.DecodeInt32(b)), nil
	case thrift.MAP:
		b := rt.BytesFrom(unsafe.Pointer(uintptr(self.v)+uintptr(2)), 4, 4)
		return int(thrift.BinaryEncoding{}.DecodeInt32(b)), nil
	default:
		return -1, errNode(meta.ErrUnsupportedType, "", nil)
	}
}

func (self Node) raw() []byte {
	return rt.BytesFrom(self.v, self.l, self.l)
}

// Return its underlying raw data
func (self Node) Raw() []byte {
	if self.Error() != "" {
		return nil
	}
	return self.raw()
}

// Byte returns the byte value contained by a I8/BYTE node
func (self Node) Byte() (byte, error) {
	if self.IsError() {
		return 0, self
	}
	return self.byte()
}

func (self Node) byte() (byte, error) {
	switch self.t {
	case thrift.BYTE:
		return byte(thrift.BinaryEncoding{}.DecodeByte(rt.BytesFrom(self.v, int(self.l), int(self.l)))), nil
	default:
		return 0, errNode(meta.ErrUnsupportedType, "", nil)
	}
}

// Bool returns the bool value contained by a BOOL node
func (self Node) Bool() (bool, error) {
	if self.IsError() {
		return false, self
	}
	return self.bool()
}

func (self Node) bool() (bool, error) {
	switch self.t {
	case thrift.BOOL:
		return thrift.BinaryEncoding{}.DecodeBool(rt.BytesFrom(self.v, int(self.l), int(self.l))), nil
	default:
		return false, errNode(meta.ErrUnsupportedType, "", nil)
	}
}

// Int returns the int value contaned by a I8/I16/I32/I64 node
func (self Node) Int() (int, error) {
	if self.IsError() {
		return 0, self
	}
	return self.int()
}

func (self Node) int() (int, error) {
	buf := rt.BytesFrom(self.v, int(self.l), int(self.l))
	switch self.t {
	case thrift.I08:
		return int(thrift.BinaryEncoding{}.DecodeByte(buf)), nil
	case thrift.I16:
		return int(thrift.BinaryEncoding{}.DecodeInt16(buf)), nil
	case thrift.I32:
		return int(thrift.BinaryEncoding{}.DecodeInt32(buf)), nil
	case thrift.I64:
		return int(thrift.BinaryEncoding{}.DecodeInt64(buf)), nil
	default:
		return 0, errNode(meta.ErrUnsupportedType, "", nil)
	}
}

// Float64 returns the float64 value contained by a DOUBLE node
func (self Node) Float64() (float64, error) {
	if self.IsError() {
		return 0, self
	}
	return self.float64()
}

func (self Node) float64() (float64, error) {
	switch self.t {
	case thrift.DOUBLE:
		return thrift.BinaryEncoding{}.DecodeDouble(rt.BytesFrom(self.v, int(self.l), int(self.l))), nil
	default:
		return 0, errNode(meta.ErrUnsupportedType, "", nil)
	}
}

// String returns the string value contianed by a STRING node
func (self Node) String() (string, error) {
	if self.IsError() {
		return "", self
	}
	return self.string()
}

func (self Node) string() (string, error) {
	switch self.t {
	case thrift.STRING:
		str := thrift.BinaryEncoding{}.DecodeString(rt.BytesFrom(self.v, int(self.l), int(self.l)))
		// if self.d.IsBinary() {
		// 	if !utf8.Valid(rt.Str2Mem(str)) {
		// 		return "", errNode(meta.ErrInvalidParam, "invalid utf8 string", nil)
		// 	}
		// }
		return str, nil
	default:
		return "", errNode(meta.ErrUnsupportedType, "", nil)
	}
}

// Binary returns the bytes value contained by a BINARY node
func (self Node) Binary() ([]byte, error) {
	if self.IsError() {
		return nil, self
	}
	return self.binary()
}

func (self Node) binary() ([]byte, error) {
	switch self.t {
	case thrift.STRING:
		return thrift.BinaryEncoding{}.DecodeBytes(rt.BytesFrom(self.v, int(self.l), int(self.l))), nil
	default:
		return nil, errNode(meta.ErrUnsupportedType, "", nil)
	}
}

// List returns interface elements contained by a LIST/SET node
func (self Node) List(opts *Options) ([]interface{}, error) {
	if self.IsError() {
		return nil, self
	}
	if err := self.should2("List", thrift.LIST, thrift.SET); err != "" {
		return nil, errNode(meta.ErrUnsupportedType, err, nil)
	}

	it := self.iterElems()
	if it.Err != nil {
		return nil, it.Err
	}
	ret := make([]interface{}, 0, it.Size())
	for it.HasNext() {
		s, e := it.Next(false)
		if it.Err != nil {
			return nil, it.Err
		}
		v := self.slice(s, e, self.et)
		vv, err := v.Interface(opts)
		if err != nil {
			return ret, err
		}
		ret = append(ret, vv)
	}
	return ret, it.Err
}

// StrMap returns the string keys and interface elements contained by a MAP<STRING,XX> node
func (self Node) StrMap(opts *Options) (map[string]interface{}, error) {
	if self.IsError() {
		return nil, self
	}
	if err := self.should("StrMap", thrift.MAP); err != "" {
		return nil, errNode(meta.ErrUnsupportedType, err, nil)
	}
	it := self.iterPairs()
	if it.Err != nil {
		return nil, it.Err
	}
	if self.kt != thrift.STRING {
		return nil, errNode(meta.ErrUnsupportedType, "key type must by STRING", nil)
	}
	ret := make(map[string]interface{}, it.Size())
	for it.HasNext() {
		_, ks, s, e := it.NextStr(false)
		if it.Err != nil {
			return nil, it.Err
		}
		v := self.slice(s, e, self.et)
		vv, err := v.Interface(opts)
		if err != nil {
			return ret, err
		}
		ret[ks] = vv
	}
	return ret, it.Err
}

// StrMap returns the integer keys and interface elements contained by a MAP<I8|I16|I32|I64,XX> node
func (self Node) IntMap(opts *Options) (map[int]interface{}, error) {
	if self.IsError() {
		return nil, self
	}
	if err := self.should("IntMap", thrift.MAP); err != "" {
		return nil, errNode(meta.ErrUnsupportedType, err, nil)
	}
	if !self.kt.IsInt() {
		return nil, errNode(meta.ErrUnsupportedType, "key must be INT type", nil)
	}
	it := self.iterPairs()
	if it.Err != nil {
		return nil, it.Err
	}
	ret := make(map[int]interface{}, it.Size())
	for it.HasNext() {
		_, ks, s, e := it.NextInt(false)
		if it.Err != nil {
			return nil, it.Err
		}
		v := self.slice(s, e, self.et)
		vv, err := v.Interface(opts)
		if err != nil {
			return ret, err
		}
		ret[ks] = vv
	}
	return ret, it.Err
}

// InterfaceMap returns the interface keys and interface elements contained by a MAP node.
// If the key type is complex (LIST/SET/MAP/STRUCT),
// it will be stored using its pointer since its value are not supported by Go
func (self Node) InterfaceMap(opts *Options) (map[interface{}]interface{}, error) {
	if self.IsError() {
		return nil, self
	}
	if err := self.should("InterfaceMap", thrift.MAP); err != "" {
		return nil, errNode(meta.ErrUnsupportedType, err, nil)
	}
	// if self.kt.IsInt() || self.kt == thrift.STRING {
	// 	return nil, errNode(meta.ErrUnsupportedType, "key must be non-integer-nor-string type", nil)
	// }
	it := self.iterPairs()
	if it.Err != nil {
		return nil, it.Err
	}
	ret := make(map[interface{}]interface{}, it.Size())
	for it.HasNext() {
		_, ks, s, e := it.NextBin(false)
		if it.Err != nil {
			return nil, it.Err
		}
		k := NewNode(self.kt, ks)
		kv, err := k.Interface(opts)
		if err != nil {
			return ret, err
		}
		v := self.slice(s, e, self.et)
		vv, err := v.Interface(opts)
		if err != nil {
			return ret, err
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
		case map[thrift.FieldID]interface{}:
			ret[&x] = vv
		default:
			ret[kv] = vv
		}
	}
	return ret, it.Err
}

// Interface returns the go interface value contained by a node.
// If the node is a STRUCT, it will return a map[thrift.FieldID]interface{}
// If it is a map, it will return map[int|string|interface{}]interface{}, which depends on the key type
func (self Node) Interface(opts *Options) (interface{}, error) {
	switch self.t {
	case thrift.ERROR:
		return nil, self
	case thrift.BOOL:
		return self.bool()
	case thrift.I08, thrift.I16, thrift.I32, thrift.I64:
		return self.int()
	case thrift.DOUBLE:
		return self.float64()
	case thrift.STRING:
		if opts.CastStringAsBinary {
			return self.binary()
		}
		return self.string()
	case thrift.LIST, thrift.SET:
		return self.List(opts)
	case thrift.MAP:
		if kt := self.kt; kt == thrift.STRING {
			return self.StrMap(opts)
		} else if kt.IsInt() {
			return self.IntMap(opts)
		} else {
			return self.InterfaceMap(opts)
		}
	case thrift.STRUCT:
		it := self.iterFields()
		if it.Err != nil {
			return nil, it.Err
		}
		var ret1 map[thrift.FieldID]interface{}
		var ret2 map[int]interface{}
		if opts.MapStructById {
			ret1 = make(map[thrift.FieldID]interface{}, DefaultNodeSliceCap)
		} else {
			ret2 = make(map[int]interface{}, DefaultNodeSliceCap)
		}
		for it.HasNext() {
			id, t, s, e := it.Next(false)
			if it.Err != nil {
				return nil, it.Err
			}
			v := self.slice(s, e, t)
			vv, err := v.Interface(opts)
			if err != nil {
				return nil, err
			}
			if opts.MapStructById {
				ret1[thrift.FieldID(id)] = vv
			} else {
				ret2[int(id)] = vv
			}
		}
		if opts.MapStructById {
			return ret1, it.Err
		} else {
			return ret2, it.Err
		}
		// }
	default:
		return 0, errNode(meta.ErrUnsupportedType, "", nil)
	}
}
