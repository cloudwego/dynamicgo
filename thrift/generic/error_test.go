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
	"testing"
	"unsafe"

	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/stretchr/testify/require"
)

func setWrongData(v Value, pathes ...Path) {
	c := v.GetByPath(pathes...)
	if c.Check() != nil {
		panic(c.Error())
	}

	switch c.Type() {
	case thrift.STRUCT:
		p := (*[3]byte)(unsafe.Pointer(uintptr(c.v)))
		*p = [3]byte{0xff, 0xff, 0xff}
	case thrift.MAP:
		p := (*[6]byte)(unsafe.Pointer(uintptr(c.v)))
		*p = [6]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	case thrift.LIST, thrift.SET:
		p := (*[5]byte)(unsafe.Pointer(uintptr(c.v)))
		*p = [5]byte{0xff, 0xff, 0xff, 0xff, 0xff}
	default:
		panic("not implement yet")
	}
}

func TestError(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()

	errRead1 := meta.NewError(meta.ErrRead, "", nil)
	errRead2 := errNode(meta.ErrRead, "", nil)
	errUnsupportedType := errValue(meta.ErrUnsupportedType, "", nil)
	errNode := NewNode(thrift.ERROR, nil)
	errUnknownField := errValue(meta.ErrUnknownField, "", nil)

	t.Run("struct", func(t *testing.T) {
		opts := &Options{}
		cases := []testAPICase{
			{name: "", api: "Field", args: []interface{}{thrift.FieldID(1)}, exp: nil, err: errRead2},
			{name: "", api: "FieldByName", args: []interface{}{"xx"}, exp: nil, err: errUnknownField},
			{name: "", api: "Index", args: []interface{}{1}, exp: nil, err: errUnsupportedType},
			{name: "", api: "GetByStr", args: []interface{}{"xx"}, exp: nil, err: errUnsupportedType},
			{name: "", api: "GetByInt", args: []interface{}{1}, exp: nil, err: errUnsupportedType},
			{name: "", api: "Interface", args: []interface{}{opts}, exp: nil, err: errNode},
			{name: "", api: "Children", args: []interface{}{&[]PathNode{}, false, opts}, exp: nil, err: errRead1},
		}
		data := []byte(string(data))
		v := NewValue(desc, data)
		n := v.GetByPath(PathInnerBase)
		setWrongData(v, PathInnerBase)
		for _, c := range cases {
			t.Run(c.api, func(t *testing.T) {
				c.path = []Path{PathInnerBase}
				APIHelper(t, nil, &n, c)
			})
		}
	})

	t.Run("strmap", func(t *testing.T) {
		opts := &Options{}
		cases := []testAPICase{
			{name: "", api: "Field", args: []interface{}{thrift.FieldID(1)}, exp: nil, err: errUnsupportedType},
			{name: "", api: "FieldByName", args: []interface{}{"xx"}, exp: nil, err: errUnsupportedType},
			{name: "", api: "Index", args: []interface{}{1}, exp: nil, err: errUnsupportedType},
			{name: "", api: "GetByStr", args: []interface{}{"a"}, exp: nil, err: errRead2},
			{name: "", api: "GetByInt", args: []interface{}{1}, exp: nil, err: errUnsupportedType},
			{name: "", api: "Interface", args: []interface{}{opts}, exp: map[string]interface{}(nil), err: errRead1},
			{name: "", api: "Children", args: []interface{}{&[]PathNode{}, false, opts}, exp: nil, err: errRead1},
		}
		data := []byte(string(data))
		v := NewValue(desc, data)
		n := v.GetByPath(PathExampleMapStringString...)
		setWrongData(v, PathExampleMapStringString...)
		for _, c := range cases {
			t.Run(c.api, func(t *testing.T) {
				c.path = []Path{PathInnerBase}
				APIHelper(t, nil, &n, c)
			})
		}
	})

	t.Run("intmap", func(t *testing.T) {
		opts := &Options{}
		cases := []testAPICase{
			{name: "", api: "Field", args: []interface{}{thrift.FieldID(1)}, exp: nil, err: errUnsupportedType},
			{name: "", api: "FieldByName", args: []interface{}{"xx"}, exp: nil, err: errUnsupportedType},
			{name: "", api: "Index", args: []interface{}{1}, exp: nil, err: errUnsupportedType},
			{name: "", api: "GetByStr", args: []interface{}{"a"}, exp: nil, err: errUnsupportedType},
			{name: "", api: "GetByInt", args: []interface{}{1}, exp: nil, err: errRead2},
			{name: "", api: "Interface", args: []interface{}{opts}, exp: map[int]interface{}(nil), err: errRead1},
			{name: "", api: "Children", args: []interface{}{&[]PathNode{}, false, opts}, exp: nil, err: errRead1},
		}
		data := []byte(string(data))
		v := NewValue(desc, data)
		n := v.GetByPath(PathExampleMapInt32String...)
		setWrongData(v, PathExampleMapInt32String...)
		for _, c := range cases {
			t.Run(c.api, func(t *testing.T) {
				c.path = []Path{PathInnerBase}
				APIHelper(t, nil, &n, c)
			})
		}
	})

	t.Run("list", func(t *testing.T) {
		opts := &Options{}
		cases := []testAPICase{
			{name: "", api: "Field", args: []interface{}{thrift.FieldID(1)}, exp: nil, err: errUnsupportedType},
			{name: "", api: "FieldByName", args: []interface{}{"xx"}, exp: nil, err: errUnsupportedType},
			{name: "", api: "Index", args: []interface{}{1}, exp: nil, err: errRead2},
			{name: "", api: "GetByStr", args: []interface{}{"a"}, exp: nil, err: errUnsupportedType},
			{name: "", api: "GetByInt", args: []interface{}{1}, exp: nil, err: errUnsupportedType},
			{name: "", api: "Interface", args: []interface{}{opts}, exp: []interface{}(nil), err: errRead1},
			{name: "", api: "Children", args: []interface{}{&[]PathNode{}, false, opts}, exp: nil, err: errRead1},
		}
		data := []byte(string(data))
		v := NewValue(desc, data)
		n := v.GetByPath(PathExampleListInt32...)
		setWrongData(v, PathExampleListInt32...)
		for _, c := range cases {
			t.Run(c.api, func(t *testing.T) {
				c.path = []Path{PathInnerBase}
				APIHelper(t, nil, &n, c)
			})
		}
	})
}

func TestWrapError(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()
	opts := Options{}
	v := NewValue(desc, data)
	tree := PathNode{
		Path: NewPathFieldName("root"),
		Node: v.Node,
		Next: []PathNode{
			{
				Path: NewPathFieldId(3),
				Node: v.GetByPath(PathInnerBase).Node,
				Next: []PathNode{
					{
						Path: NewPathFieldId(9),
						Node: v.GetByPath(PathExampleMapStringString...).Node,
					},
					{
						Path: NewPathFieldId(12),
						Node: v.GetByPath(PathExampleMapInt32String...).Node,
					},
					{
						Path: NewPathFieldId(8),
						Node: v.GetByPath(PathExampleListInt32...).Node,
					},
				},
			},
		},
	}
	e := meta.NewError(meta.NewErrorCode(meta.ErrNotFound, meta.THRIFT), "", nil)
	ev := errNode(meta.ErrRead, "", e)
	t.Run("struct", func(t *testing.T) {
		tree := tree.Fork()
		tree.Next[0].Node = ev
		out, err := tree.Marshal(&opts)
		require.Nil(t, out)
		require.NotNil(t, err)
		require.Equal(t, meta.ErrRead, err.(meta.Error).Code.Behavior())
	})

	t.Run("strmap", func(t *testing.T) {
		tree := tree.Fork()
		tree.Next[0].Next[0].Next = []PathNode{{Path: NewPathStrKey("x"), Node: ev}}
		out, err := tree.Marshal(&opts)
		require.Nil(t, out)
		require.NotNil(t, err)
		require.Equal(t, meta.ErrRead, err.(meta.Error).Code.Behavior())
	})

	t.Run("intmap", func(t *testing.T) {
		tree := tree.Fork()
		tree.Next[0].Next[1].Next = []PathNode{{Path: NewPathIntKey(999), Node: ev}}
		out, err := tree.Marshal(&opts)
		require.Nil(t, out)
		require.NotNil(t, err)
		require.Equal(t, meta.ErrRead, err.(meta.Error).Code.Behavior())
	})

	t.Run("list", func(t *testing.T) {
		tree := tree.Fork()
		tree.Next[0].Next[2].Next = []PathNode{{Path: NewPathIndex(999), Node: ev}}
		out, err := tree.Marshal(&opts)
		require.Nil(t, out)
		require.NotNil(t, err)
		require.Equal(t, meta.ErrRead, err.(meta.Error).Code.Behavior())
	})
}
