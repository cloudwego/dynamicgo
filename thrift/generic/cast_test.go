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
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/testdata/sample"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/stretchr/testify/require"
)

var (
	PathInnerBase              = NewPathFieldName("InnerBase")
	PathNotExist               = []Path{NewPathFieldName("NotExist")}
	PathExampleBool            = []Path{PathInnerBase, NewPathFieldId(thrift.FieldID(1))}
	PathExampleByte            = []Path{PathInnerBase, NewPathFieldId(thrift.FieldID(2))}
	PathExampleInt16           = []Path{PathInnerBase, NewPathFieldId(thrift.FieldID(3))}
	PathExampleInt32           = []Path{PathInnerBase, NewPathFieldId(thrift.FieldID(4))}
	PathExampleInt64           = []Path{PathInnerBase, NewPathFieldId(thrift.FieldID(5))}
	PathExampleDouble          = []Path{PathInnerBase, NewPathFieldId(thrift.FieldID(6))}
	PathExampleString          = []Path{PathInnerBase, NewPathFieldId(thrift.FieldID(7))}
	PathExampleListInt32       = []Path{PathInnerBase, NewPathFieldId(thrift.FieldID(8))}
	PathExampleMapStringString = []Path{PathInnerBase, NewPathFieldId(thrift.FieldID(9))}
	PathExampleSetInt32_       = []Path{PathInnerBase, NewPathFieldId(thrift.FieldID(10))}
	PathExampleFoo             = []Path{PathInnerBase, NewPathFieldId(thrift.FieldID(11))}
	PathExampleMapInt32String  = []Path{PathInnerBase, NewPathFieldId(thrift.FieldID(12))}
	PathExampleBinary          = []Path{PathInnerBase, NewPathFieldId(thrift.FieldID(13))}
	PathExampleBase            = []Path{PathInnerBase, NewPathFieldId(thrift.FieldID(255))}
)

func TestCastInt8(t *testing.T) {
	var exp = byte(129)
	var n = NewNodeByte(exp)
	act, err := n.Int()
	require.Nil(t, err)
	require.Equal(t, int(exp), act)
}

func TestCastAPI(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()
	var errUnsupportedType = errNode(meta.ErrUnsupportedType, "", nil)
	var errNotFound = errNode(meta.ErrUnknownField, "", nil)
	// var errInvalidParam = errNode(meta.ErrInvalidParam, "", nil)
	var opts = &Options{}
	cases := []testAPICase{
		{path: PathNotExist, api: "Bool", exp: false, err: errNotFound},
		{path: PathNotExist, api: "Byte", exp: byte(0), err: errNotFound},
		{path: PathNotExist, api: "Int", exp: int(0), err: errNotFound},
		{path: PathNotExist, api: "Float64", exp: float64(0), err: errNotFound},
		{path: PathNotExist, api: "String", exp: "", err: errNotFound},
		{path: PathNotExist, api: "Binary", exp: []byte(nil), err: errNotFound},
		{path: PathNotExist, api: "List", exp: []interface{}(nil), err: errNotFound, args: []interface{}{opts}},
		{path: PathNotExist, api: "StrMap", exp: map[string]interface{}(nil), err: errNotFound, args: []interface{}{opts}},
		{path: PathNotExist, api: "IntMap", exp: map[int]interface{}(nil), err: errNotFound, args: []interface{}{opts}},
		{path: PathNotExist, api: "Interface", exp: nil, err: errNotFound, args: []interface{}{opts}},

		{path: PathExampleBool, api: "Bool", exp: sample.Example2Obj.InnerBase.Bool},
		{path: PathExampleBool, api: "Byte", exp: byte(0), err: errUnsupportedType},
		{path: PathExampleBool, api: "Int", exp: int(0), err: errUnsupportedType},
		{path: PathExampleBool, api: "Float64", exp: float64(0), err: errUnsupportedType},
		{path: PathExampleBool, api: "String", exp: "", err: errUnsupportedType},
		{path: PathExampleBool, api: "Binary", exp: []byte(nil), err: errUnsupportedType},
		{path: PathExampleBool, api: "List", exp: []interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleBool, api: "StrMap", exp: map[string]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleBool, api: "IntMap", exp: map[int]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleBool, api: "Interface", exp: sample.Example2Obj.InnerBase.Bool, args: []interface{}{opts}},

		{path: PathExampleByte, api: "Bool", exp: false, err: errUnsupportedType},
		{path: PathExampleByte, api: "Byte", exp: byte(sample.Example2Obj.InnerBase.Byte)},
		{path: PathExampleByte, api: "Int", exp: int(sample.Example2Obj.InnerBase.Byte)},
		{path: PathExampleByte, api: "Float64", exp: float64(0), err: errUnsupportedType},
		{path: PathExampleByte, api: "String", exp: "", err: errUnsupportedType},
		{path: PathExampleByte, api: "Binary", exp: []byte(nil), err: errUnsupportedType},
		{path: PathExampleByte, api: "List", exp: []interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleByte, api: "StrMap", exp: map[string]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleByte, api: "IntMap", exp: map[int]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleByte, api: "Interface", exp: int(sample.Example2Obj.InnerBase.Byte), args: []interface{}{opts}},

		{path: PathExampleInt16, api: "Bool", exp: false, err: errUnsupportedType},
		{path: PathExampleInt16, api: "Byte", exp: byte(0), err: errUnsupportedType},
		{path: PathExampleInt16, api: "Int", exp: int(sample.Example2Obj.InnerBase.Int16)},
		{path: PathExampleInt16, api: "Float64", exp: float64(0), err: errUnsupportedType},
		{path: PathExampleInt16, api: "String", exp: "", err: errUnsupportedType},
		{path: PathExampleInt16, api: "Binary", exp: []byte(nil), err: errUnsupportedType},
		{path: PathExampleInt16, api: "List", exp: []interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleInt16, api: "StrMap", exp: map[string]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleInt16, api: "IntMap", exp: map[int]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleInt16, api: "Interface", exp: int(sample.Example2Obj.InnerBase.Int16), args: []interface{}{opts}},

		{path: PathExampleInt32, api: "Bool", exp: false, err: errUnsupportedType},
		{path: PathExampleInt32, api: "Byte", exp: byte(0), err: errUnsupportedType},
		{path: PathExampleInt32, api: "Int", exp: int(sample.Example2Obj.InnerBase.Int32)},
		{path: PathExampleInt32, api: "Float64", exp: float64(0), err: errUnsupportedType},
		{path: PathExampleInt32, api: "String", exp: "", err: errUnsupportedType},
		{path: PathExampleInt32, api: "Binary", exp: []byte(nil), err: errUnsupportedType},
		{path: PathExampleInt32, api: "List", exp: []interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleInt32, api: "StrMap", exp: map[string]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleInt32, api: "IntMap", exp: map[int]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleInt32, api: "Interface", exp: int(sample.Example2Obj.InnerBase.Int32), args: []interface{}{opts}},

		{path: PathExampleInt64, api: "Bool", exp: false, err: errUnsupportedType},
		{path: PathExampleInt64, api: "Byte", exp: byte(0), err: errUnsupportedType},
		{path: PathExampleInt64, api: "Int", exp: int(sample.Example2Obj.InnerBase.Int64)},
		{path: PathExampleInt64, api: "Float64", exp: float64(0), err: errUnsupportedType},
		{path: PathExampleInt64, api: "String", exp: "", err: errUnsupportedType},
		{path: PathExampleInt64, api: "Binary", exp: []byte(nil), err: errUnsupportedType},
		{path: PathExampleInt64, api: "List", exp: []interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleInt64, api: "StrMap", exp: map[string]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleInt64, api: "IntMap", exp: map[int]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleInt64, api: "Interface", exp: int(sample.Example2Obj.InnerBase.Int64), args: []interface{}{opts}},

		{path: PathExampleDouble, api: "Bool", exp: false, err: errUnsupportedType},
		{path: PathExampleDouble, api: "Byte", exp: byte(0), err: errUnsupportedType},
		{path: PathExampleDouble, api: "Int", exp: int(0), err: errUnsupportedType},
		{path: PathExampleDouble, api: "Float64", exp: sample.Example2Obj.InnerBase.Double},
		{path: PathExampleDouble, api: "String", exp: "", err: errUnsupportedType},
		{path: PathExampleDouble, api: "Binary", exp: []byte(nil), err: errUnsupportedType},
		{path: PathExampleDouble, api: "List", exp: []interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleDouble, api: "StrMap", exp: map[string]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleDouble, api: "IntMap", exp: map[int]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleDouble, api: "Interface", exp: float64(sample.Example2Obj.InnerBase.Double), args: []interface{}{opts}},

		{path: PathExampleString, api: "Bool", exp: false, err: errUnsupportedType},
		{path: PathExampleString, api: "Byte", exp: byte(0), err: errUnsupportedType},
		{path: PathExampleString, api: "Int", exp: int(0), err: errUnsupportedType},
		{path: PathExampleString, api: "Float64", exp: float64(0), err: errUnsupportedType},
		{path: PathExampleString, api: "String", exp: sample.Example2Obj.InnerBase.String_},
		{path: PathExampleString, api: "Binary", exp: []byte(sample.Example2Obj.InnerBase.String_)},
		{path: PathExampleString, api: "List", exp: []interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleString, api: "StrMap", exp: map[string]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleString, api: "IntMap", exp: map[int]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleString, api: "Interface", exp: string(sample.Example2Obj.InnerBase.String_), args: []interface{}{opts}},

		{path: PathExampleBinary, api: "Bool", exp: false, err: errUnsupportedType},
		{path: PathExampleBinary, api: "Byte", exp: byte(0), err: errUnsupportedType},
		{path: PathExampleBinary, api: "Int", exp: int(0), err: errUnsupportedType},
		{path: PathExampleBinary, api: "Float64", exp: float64(0), err: errUnsupportedType},
		{path: PathExampleBinary, api: "String", exp: string(sample.Example2Obj.InnerBase.Binary)},
		{path: PathExampleBinary, api: "Binary", exp: []byte(sample.Example2Obj.InnerBase.Binary)},
		{path: PathExampleBinary, api: "List", exp: []interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleBinary, api: "StrMap", exp: map[string]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleBinary, api: "IntMap", exp: map[int]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleBinary, api: "Interface", exp: string(sample.Example2Obj.InnerBase.Binary), args: []interface{}{opts}},

		{path: PathExampleFoo, api: "Bool", exp: false, err: errUnsupportedType},
		{path: PathExampleFoo, api: "Byte", exp: byte(0), err: errUnsupportedType},
		{path: PathExampleFoo, api: "Int", exp: int(sample.Example2Obj.InnerBase.Foo)},
		{path: PathExampleFoo, api: "Float64", exp: float64(0), err: errUnsupportedType},
		{path: PathExampleFoo, api: "String", exp: "", err: errUnsupportedType},
		{path: PathExampleFoo, api: "Binary", exp: []byte(nil), err: errUnsupportedType},
		{path: PathExampleFoo, api: "List", exp: []interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleFoo, api: "StrMap", exp: map[string]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleFoo, api: "IntMap", exp: map[int]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleFoo, api: "Interface", exp: int(sample.Example2Obj.InnerBase.Foo), args: []interface{}{opts}},

		{path: PathExampleListInt32, api: "Bool", exp: false, err: errUnsupportedType},
		{path: PathExampleListInt32, api: "Byte", exp: byte(0), err: errUnsupportedType},
		{path: PathExampleListInt32, api: "Int", exp: int(0), err: errUnsupportedType},
		{path: PathExampleListInt32, api: "Float64", exp: float64(0), err: errUnsupportedType},
		{path: PathExampleListInt32, api: "String", exp: "", err: errUnsupportedType},
		{path: PathExampleListInt32, api: "Binary", exp: []byte(nil), err: errUnsupportedType},
		{path: PathExampleListInt32, api: "List", exp: toInterface(sample.Example2Obj.InnerBase.ListInt32), args: []interface{}{opts}},
		{path: PathExampleListInt32, api: "StrMap", exp: map[string]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleListInt32, api: "IntMap", exp: map[int]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleListInt32, api: "Interface", exp: toInterface(sample.Example2Obj.InnerBase.ListInt32), args: []interface{}{opts}},

		{path: PathExampleSetInt32_, api: "Bool", exp: false, err: errUnsupportedType},
		{path: PathExampleSetInt32_, api: "Byte", exp: byte(0), err: errUnsupportedType},
		{path: PathExampleSetInt32_, api: "Int", exp: int(0), err: errUnsupportedType},
		{path: PathExampleSetInt32_, api: "Float64", exp: float64(0), err: errUnsupportedType},
		{path: PathExampleSetInt32_, api: "String", exp: "", err: errUnsupportedType},
		{path: PathExampleSetInt32_, api: "Binary", exp: []byte(nil), err: errUnsupportedType},
		{path: PathExampleSetInt32_, api: "List", exp: toInterface(sample.Example2Obj.InnerBase.ListInt32), args: []interface{}{opts}},
		{path: PathExampleSetInt32_, api: "StrMap", exp: map[string]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleSetInt32_, api: "IntMap", exp: map[int]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleSetInt32_, api: "Interface", exp: toInterface(sample.Example2Obj.InnerBase.ListInt32), args: []interface{}{opts}},

		{path: PathExampleMapInt32String, api: "Bool", exp: false, err: errUnsupportedType},
		{path: PathExampleMapInt32String, api: "Byte", exp: byte(0), err: errUnsupportedType},
		{path: PathExampleMapInt32String, api: "Int", exp: int(0), err: errUnsupportedType},
		{path: PathExampleMapInt32String, api: "Float64", exp: float64(0), err: errUnsupportedType},
		{path: PathExampleMapInt32String, api: "String", exp: "", err: errUnsupportedType},
		{path: PathExampleMapInt32String, api: "Binary", exp: []byte(nil), err: errUnsupportedType},
		{path: PathExampleMapInt32String, api: "List", exp: []interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleMapInt32String, api: "StrMap", exp: map[string]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleMapInt32String, api: "IntMap", exp: toInterface(sample.Example2Obj.InnerBase.MapInt32String), args: []interface{}{opts}},
		{path: PathExampleMapInt32String, api: "Interface", exp: toInterface(sample.Example2Obj.InnerBase.MapInt32String), args: []interface{}{opts}},

		{path: PathExampleMapStringString, api: "Bool", exp: false, err: errUnsupportedType},
		{path: PathExampleMapStringString, api: "Byte", exp: byte(0), err: errUnsupportedType},
		{path: PathExampleMapStringString, api: "Int", exp: int(0), err: errUnsupportedType},
		{path: PathExampleMapStringString, api: "Float64", exp: float64(0), err: errUnsupportedType},
		{path: PathExampleMapStringString, api: "String", exp: "", err: errUnsupportedType},
		{path: PathExampleMapStringString, api: "Binary", exp: []byte(nil), err: errUnsupportedType},
		{path: PathExampleMapStringString, api: "List", exp: []interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleMapStringString, api: "StrMap", exp: toInterface(sample.Example2Obj.InnerBase.MapStringString), args: []interface{}{opts}},
		{path: PathExampleMapStringString, api: "IntMap", exp: map[int]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleMapStringString, api: "Interface", exp: toInterface(sample.Example2Obj.InnerBase.MapStringString), args: []interface{}{opts}},
		{path: PathExampleMapStringString, api: "Interface", exp: toInterface2(sample.Example2Obj.InnerBase.MapStringString, false, 2), args: []interface{}{&Options{
			CastStringAsBinary: true,
		}}},

		{path: PathExampleBase, api: "Bool", exp: false, err: errUnsupportedType},
		{path: PathExampleBase, api: "Byte", exp: byte(0), err: errUnsupportedType},
		{path: PathExampleBase, api: "Int", exp: int(0), err: errUnsupportedType},
		{path: PathExampleBase, api: "Float64", exp: float64(0), err: errUnsupportedType},
		{path: PathExampleBase, api: "String", exp: "", err: errUnsupportedType},
		{path: PathExampleBase, api: "Binary", exp: []byte(nil), err: errUnsupportedType},
		{path: PathExampleBase, api: "List", exp: []interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleBase, api: "StrMap", exp: map[string]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleBase, api: "IntMap", exp: map[int]interface{}(nil), err: errUnsupportedType, args: []interface{}{opts}},
		{path: PathExampleBase, api: "Interface", exp: toInterface2(sample.Example2Obj.InnerBase.Base, false, 0), args: []interface{}{opts}},
		{path: PathExampleBase, api: "Interface", exp: toInterface2(sample.Example2Obj.InnerBase.Base, true, 0), args: []interface{}{&Options{
			MapStructById: true,
		}}},
	}
	root := NewValue(desc, data)
	for _, c := range cases {
		t.Run(fmt.Sprintf(c.path[len(c.path)-1].String())+c.api, func(t *testing.T) {
			APIHelper(t, &root, nil, c)
		})
	}
}

type testAPICase struct {
	name      string
	path      []Path
	api       string
	args      []interface{}
	exp       interface{}
	err       error
	strictErr bool
}

func APIHelper(t *testing.T, root *Value, val *Value, c testAPICase) {
	if val == nil {
		t := root.GetByPath(c.path...)
		val = &t
	}
	// require.Nil(t, val.Check())
	vt := reflect.ValueOf(val)
	mt := vt.MethodByName(c.api)

	var args []reflect.Value
	if len(c.args) > 0 {
		args = make([]reflect.Value, len(c.args))
		for i, arg := range c.args {
			args[i] = reflect.ValueOf(arg)
		}
	}

	rs := mt.Call(args)
	var e interface{}
	if len(rs) == 2 {
		e = rs[1].Interface()
	} else {
		e = rs[0].Interface()
	}

	require.Equal(t, c.err == nil, e == nil)
	if e != nil {
		if n, ok := c.err.(Value); ok {
			n2, ok2 := e.(Value)
			if !ok2 {
				n3, ok3 := e.(Node)
				if !ok3 {
					t.Fatal("not a node erro ")
				} else {
					require.Equal(t, n.ErrCode(), n3.ErrCode())
				}
			}
			require.Equal(t, n.ErrCode(), n2.ErrCode())
		} else if nn, ok := c.err.(meta.Error); ok {
			nn2, ok2 := e.(meta.Error)
			if !ok2 {
				n3, ok3 := e.(Node)
				if !ok3 {
					t.Fatal("not a node erro ")
				} else {
					require.Equal(t, nn.Code.Behavior(), n3.ErrCode().Behavior())
				}
			} else {
				require.True(t, ok2)
				require.Equal(t, nn.Code.Behavior(), nn2.Code.Behavior())
			}
		}
		if c.strictErr {
			require.Equal(t, c.err, e.(error))
		}
	}
	if len(rs) >= 2 {
		require.Equal(t, c.exp, rs[0].Interface())
	}
}

func toInterface(v interface{}) interface{} {
	return toInterface2(v, false, none)
}

var bytesType = reflect.TypeOf([]byte{})

const (
	none int = 0
	b2s  int = 1
	s2b  int = 2
)

func toInterface2(v interface{}, fieldId bool, byte2string int) interface{} {
	vt := reflect.ValueOf(v)
	if vt.Kind() == reflect.Ptr {
		if vt.IsNil() {
			return nil
		}
		vt = vt.Elem()
	}
	if k := vt.Kind(); k == reflect.Slice || k == reflect.Array {
		if vt.Type() == bytesType {
			if byte2string == b2s {
				return string(vt.Bytes())
			} else {
				return vt.Bytes()
			}
		}
		var r = make([]interface{}, 0, vt.Len())
		for i := 0; i < vt.Len(); i++ {
			vv := toInterface2(vt.Index(i).Interface(), fieldId, byte2string)
			if vv != nil {
				r = append(r, vv)
			}
		}
		return r
	} else if k == reflect.Map {
		if kt := vt.Type().Key().Kind(); kt == reflect.String {
			var r = make(map[string]interface{}, vt.Len())
			for _, k := range vt.MapKeys() {
				vv := toInterface2(vt.MapIndex(k).Interface(), fieldId, byte2string)
				if vv != nil {
					r[k.String()] = vv
				}
			}
			return r
		} else if kt == reflect.Int || kt == reflect.Int8 || kt == reflect.Int16 || kt == reflect.Int32 || kt == reflect.Int64 {
			var r = make(map[int]interface{}, vt.Len())
			for _, k := range vt.MapKeys() {
				vv := toInterface2(vt.MapIndex(k).Interface(), fieldId, byte2string)
				if vv != nil {
					r[int(k.Int())] = vv
				}
			}
			return r
		} else {
			var r = make(map[interface{}]interface{}, vt.Len())
			for _, k := range vt.MapKeys() {
				kv := toInterface2(k.Interface(), fieldId, byte2string)
				vv := toInterface2(vt.MapIndex(k).Interface(), fieldId, byte2string)
				if kv != nil && vv != nil {
					switch t := kv.(type) {
					case map[string]interface{}:
						r[&t] = vv
					case map[int]interface{}:
						r[&t] = vv
					case map[interface{}]interface{}:
						r[&t] = vv
					case []interface{}:
						r[&t] = vv
					default:
						r[kv] = vv
					}
				}
			}
			return r
		}
	} else if k == reflect.Struct {
		var r interface{}
		if fieldId {
			r = map[thrift.FieldID]interface{}{}
		} else {
			r = map[int]interface{}{}
		}
		for i := 0; i < vt.NumField(); i++ {
			field := vt.Type().Field(i)
			tag := field.Tag.Get("thrift")
			ts := strings.Split(tag, ",")
			id := i
			if len(ts) > 1 {
				id, _ = strconv.Atoi(ts[1])
			}
			vv := toInterface2(vt.Field(i).Interface(), fieldId, byte2string)
			if vv != nil {
				if fieldId {
					r.(map[thrift.FieldID]interface{})[thrift.FieldID(id)] = vv
				} else {
					r.(map[int]interface{})[int(id)] = vv
				}
			}
		}
		return r
	} else if k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64 {
		return int(vt.Int())
	} else if k == reflect.String {
		if byte2string == s2b {
			return []byte(vt.String())
		} else {
			return vt.String()
		}
	}
	return vt.Interface()
}

func specialFieldName(name string) string {
	if len(name) > 0 && name[len(name)-1] == '_' {
		return name[:len(name)-1]
	}
	return name
}
