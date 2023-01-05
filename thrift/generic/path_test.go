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
	"encoding/hex"
	"reflect"
	"testing"

	// "github.com/cloudwego/dynamicgo/meta"

	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example2"
	"github.com/cloudwego/dynamicgo/testdata/sample"

	// "github.com/cloudwego/dynamicgo/testdata/sample"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/stretchr/testify/require"
)

func TestChildren(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()
	v := NewValue(desc, data)

	// t.Run("skip", func(t *testing.T) {
	// 	opts := Options{}
	// 	children := make([]PathNode, 0)
	// 	err := v.Children(&children, false, &opts)
	// 	require.Nil(t, err)
	// 	checkHelper2(t, getExampleValue(v), children, true)
	// })

	t.Run("recurse", func(t *testing.T) {
		children := make([]PathNode, 0)
		opts := Options{}
		err := v.Children(&children, true, &opts)
		require.NoError(t, err)
		require.Equal(t, 4, len(children))
		exp := toInterface2(sample.Example2Obj, false, b2s)
		act := pathNodeToInterface(PathNode{
			Node: v.Node,
			Next: children,
		}, &opts, false)
		// require.Equal(t, exp, act)
		require.True(t, DeepEqual(exp, act))
		require.True(t, DeepEqual(act, exp))
	})

	t.Run("NotScanParentNode", func(t *testing.T) {
		children := make([]PathNode, 0)
		opts := Options{}
		opts.NotScanParentNode = true
		err := v.Children(&children, true, &opts)
		require.NoError(t, err)
		require.Equal(t, 4, len(children))
		exp := toInterface2(sample.Example2Obj, false, b2s)
		act := pathNodeToInterface(PathNode{
			Node: v.Node,
			Next: children,
		}, &opts, false)
		require.True(t, DeepEqual(exp, act))
		require.True(t, DeepEqual(act, exp))
	})
}

var (
	BasePathNode = PathNode{
		Path: NewPathFieldId(255),
		Next: []PathNode{
			{
				Path: NewPathFieldId(1),
			},
			{
				Path: NewPathFieldId(2),
			},
			{
				Path: NewPathFieldId(3),
			},
			{
				Path: NewPathFieldId(4),
			},
			{
				Path: NewPathFieldId(5),
				Next: []PathNode{
					{
						Path: NewPathFieldId(1),
					},
					{
						Path: NewPathFieldId(2),
					},
				},
			},
			{
				Path: NewPathFieldId(6),
			},
		},
	}
	InnerBasePathNode = PathNode{
		Path: NewPathFieldId(3),
		Next: []PathNode{
			{
				Path: NewPathFieldId(1),
			},
			{
				Path: NewPathFieldId(2),
			},
			{
				Path: NewPathFieldId(3),
			},
			{
				Path: NewPathFieldId(4),
			},
			{
				Path: NewPathFieldId(5),
			},
			{
				Path: NewPathFieldId(6),
			},
			{
				Path: NewPathFieldId(7),
			},
			{
				Path: NewPathFieldId(8),
			},
			{
				Path: NewPathFieldId(9),
			},
			{
				Path: NewPathFieldId(10),
			},
			{
				Path: NewPathFieldId(11),
			},
			{
				Path: NewPathFieldId(12),
			},
			{
				Path: NewPathFieldId(13),
			},
			{
				Path: NewPathFieldId(14),
			},
			{
				Path: NewPathFieldId(15),
			},
			{
				Path: NewPathFieldId(16),
			},
			BasePathNode,
		},
	}
	ExamplePathNode = PathNode{
		Path: NewPathFieldName("root"),
		Next: []PathNode{
			{
				Path: NewPathFieldId(1),
			},
			{
				Path: NewPathFieldId(2),
			},
			InnerBasePathNode,
			BasePathNode,
			{
				Path: NewPathFieldId(32767),
			},
		},
	}
)

func TestPathCast(t *testing.T) {
	t.Run("field id", func(t *testing.T) {
		exp := thrift.FieldID(1)
		path := NewPathFieldId(exp)
		require.Equal(t, exp, path.Id())
		require.Equal(t, exp, path.Value())
	})
	t.Run("field name", func(t *testing.T) {
		exp := "name"
		path := NewPathFieldName(exp)
		require.Equal(t, exp, path.Str())
		require.Equal(t, exp, path.Value())
	})
	t.Run("index", func(t *testing.T) {
		exp := 1
		path := NewPathIndex(exp)
		require.Equal(t, exp, path.Int())
		require.Equal(t, exp, path.Value())
	})
	t.Run("key", func(t *testing.T) {
		exp := "key"
		path := NewPathStrKey(exp)
		require.Equal(t, exp, path.Str())
		require.Equal(t, exp, path.Value())
	})
	// t.Run("obj key", func(t *testing.T) {
	// 	exp := struct{ A int }{1}
	// 	path := NewPathObjKey(exp)
	// 	require.Equal(t, exp, path.Value())
	// })
}

// func TestDescriptorToPN(t *testing.T) {
// 	desc := getExampleDesc()

// 	t.Run("field id", func(t *testing.T) {
// 		var opts = &Options{
// 			// OnlyScanStruct: true,
// 		}
// 		var exp = ExamplePathNode

// 		act := PathNode{
// 			Path: NewPathFieldName("root"),
// 		}
// 		require.Nil(t, DescriptorToPathNode(desc, &act, opts))
// 		require.Equal(t, exp, act)
// 	})
// }

func TestPathReuse(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()
	root := NewValue(desc, data)

	obj := ExamplePathNode.Fork()
	require.Equal(t, ExamplePathNode, obj)

	obj.Node = root.Node
	opts := &Options{}
	require.Nil(t, obj.Assgin(true, opts))
	out1, err := obj.Marshal(opts)
	require.Nil(t, err)

	obj.ResetValue()
	require.Equal(t, ExamplePathNode, obj)

	obj.Node = root.Node
	require.Nil(t, obj.Assgin(true, opts))
	out2, err := obj.Marshal(opts)
	require.Nil(t, err)

	require.Equal(t, out1, out2)
}

func TestPathEmpty(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()
	root := NewValue(desc, data)
	opts := &Options{}

	ori := PathNode{
		Path: NewPathFieldName("root"),
		Node: root.Node,
	}
	root.Children(&ori.Next, true, opts)

	obj := ori.Fork()
	obj.Next = append(obj.Next, PathNode{Path: NewPathFieldName("empty")})
	obj.Next[2].Next = append(obj.Next[2].Next, PathNode{Path: NewPathFieldName("empty")})
	obj.Next[2].Next[7].Next = append(obj.Next[2].Next[7].Next, PathNode{Path: NewPathIndex(1024)})
	obj.Next[2].Next[8].Next = append(obj.Next[2].Next[8].Next, PathNode{Path: NewPathStrKey("empty")})
	obj.Node = root.Node

	require.Nil(t, obj.Assgin(true, opts))
	out1, err := obj.Marshal(opts)
	require.Nil(t, err)

	obj = ori.Fork()
	obj.Node = root.Node
	require.Nil(t, obj.Assgin(true, opts))
	out2, err := obj.Marshal(opts)
	require.Nil(t, err)

	require.Equal(t, out1, out2)
}

func TestTreeGet(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()

	exp := example2.NewExampleReq()
	v := NewValue(desc, data)
	tree := PathNode{
		Node: v.Node,
		Next: []PathNode{
			{
				Path: NewPathFieldId(1),
			},
			{
				Path: NewPathFieldId(3),
				Next: []PathNode{
					{
						Path: NewPathFieldId(1),
					},
					{
						Path: NewPathFieldId(8),
						Next: []PathNode{
							{
								Path: NewPathIndex(1),
							},
						},
					},
					{
						Path: NewPathFieldId(9),
						Next: []PathNode{
							{
								Path: NewPathStrKey("b"),
							},
						},
					},
					{
						Path: NewPathFieldId(12),
						Next: []PathNode{
							{
								Path: NewPathIntKey(2),
							},
						},
					},
				},
			},
			{
				Path: NewPathFieldId(255),
				Next: []PathNode{
					{
						Path: NewPathFieldId(2),
					},
				},
			},
		},
	}
	opts := Options{}
	err := tree.Assgin(true, &opts)
	require.NoError(t, err)

	_, err = exp.FastRead(data)
	require.Nil(t, err)

	expM2 := map[int]interface{}{}
	for k, v := range exp.InnerBase.MapInt32String {
		expM2[int(k)] = v
	}
	checkHelper(t, *exp.Msg, tree.Next[0].Node, "String")
	checkHelper(t, exp.InnerBase.Bool, tree.Next[1].Next[0].Node, "Bool")
	checkHelper(t, exp.InnerBase.ListInt32, tree.Next[1].Next[1].Node, "List")
	checkHelper(t, exp.InnerBase.MapStringString, tree.Next[1].Next[2].Node, "StrMap")
	checkHelper(t, expM2, tree.Next[1].Next[3].Node, "IntMap")
	checkHelper(t, exp.Base.Client, tree.Next[2].Next[0].Node, "String")

	out, err := tree.Marshal(&opts)
	require.Nil(t, err)
	println(hex.Dump(out))
}

func TestTreeMarshal(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()

	v := NewValue(desc, data)
	tree := PathNode{
		Node: v.Node,
		Next: []PathNode{
			{
				Path: NewPathFieldId(1),
			},
			{
				Path: NewPathFieldId(3),
				Next: []PathNode{
					{
						Path: NewPathFieldId(1),
					},
					{
						Path: NewPathFieldId(8),
						Next: []PathNode{
							{
								Path: NewPathIndex(1),
							},
						},
					},
					{
						Path: NewPathFieldId(9),
						Next: []PathNode{
							{
								Path: NewPathStrKey("b"),
							},
						},
					},
					{
						Path: NewPathFieldId(12),
						Next: []PathNode{
							{
								Path: NewPathIntKey(2),
							},
						},
					},
				},
			},
			{
				Path: NewPathFieldId(255),
				Next: []PathNode{
					{
						Path: NewPathFieldId(2),
					},
				},
			},
		},
	}
	opts := Options{}
	err := tree.Assgin(true, &opts)
	require.NoError(t, err)

	out, err := tree.Marshal(&opts)
	require.Nil(t, err)
	// spew.Dump(out)
	exp := example2.NewExampleReq()
	_, err = exp.FastRead(out)
	require.Nil(t, err)

	x := v.GetByPath(PathExampleByte...)
	tt := PathNode{
		Path: NewPathFieldName("Msg"),
		Node: x.Node,
		Next: []PathNode{
			tree,
		},
	}
	out, err = tt.Marshal(&opts)
	require.Nil(t, err)
	require.Equal(t, x.Raw(), out)
}

func getExampleValue(v Value) []PathNode {
	return []PathNode{
		{
			Path: NewPathFieldId(1),
			Node: v.GetByPath(NewPathFieldName("Msg")).Node,
		},
		{
			Path: NewPathFieldId(3),
			Node: v.GetByPath(NewPathFieldName("InnerBase")).Node,
		},
		{
			Path: NewPathFieldId(255),
			Node: v.GetByPath(NewPathFieldName("Base")).Node,
		},
		{
			Path: NewPathFieldId(32767),
			Node: v.GetByPath(NewPathFieldName("Subfix")).Node,
		},
	}
}

func getInnerBase(v Value) []PathNode {
	return []PathNode{
		{
			Path: NewPathFieldId(1),
			Node: v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("LogID")).Node,
		},
		{
			Path: NewPathFieldId(2),
			Node: v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Caller")).Node,
		},
		{
			Path: NewPathFieldId(3),
			Node: v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Addr")).Node,
		},
		{
			Path: NewPathFieldId(4),
			Node: v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Client")).Node,
		},
		{
			Path: NewPathFieldId(5),
			Node: v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("TrafficEnv")).Node,
			Next: []PathNode{
				{
					Path: NewPathFieldId(1),
					Node: v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("TrafficEnv"), NewPathFieldName("Open")).Node,
				},
				{
					Path: NewPathFieldId(2),
					Node: v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("TrafficEnv"), NewPathFieldName("Env")).Node,
				},
			},
		},
		{
			Path: NewPathFieldId(6),
			Node: v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Extra")).Node,
			Next: []PathNode{
				{
					Path: NewPathStrKey("a"),
					Node: v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("a")).Node,
				},
				{
					Path: NewPathStrKey("b"),
					Node: v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("b")).Node,
				},
				{
					Path: NewPathStrKey("c"),
					Node: v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("c")).Node,
				},
			},
		},
	}
}

func checkHelper2(t *testing.T, exp []PathNode, act []PathNode, checkNode bool) {
	require.Equal(t, len(exp), len(act))
	for i := range exp {
		t.Logf("Path: %s\n", exp[i].Path)
		found := false
		for j := range act {
			if exp[i].Path.String() == act[j].Path.String() {
				if checkNode {
					require.Equal(t, exp[i].Node, act[j].Node)
				}
				checkHelper2(t, exp[i].Next, act[j].Next, checkNode)
				found = true
				break
			}
		}
		require.True(t, found)
	}
}

// func toPathNode(v interface{}, ret *PathNode) {
// 	vt := reflect.ValueOf(v)
// 	if vt.Kind() == reflect.Ptr {
// 		if vt.IsNil() {
// 			return
// 		}
// 		vt = vt.Elem()
// 	}

// 	if k := vt.Kind(); k == reflect.Slice || k == reflect.Array {
// 		if vt.Type() == bytesType || vt.Len() == 0 {
// 			return
// 		}
// 		if len(ret.Next) < vt.Len() {
// 			ret.Next = make([]PathNode, vt.Len())
// 		}
// 		var r *PathNode
// 		for i := 0; i < vt.Len(); i++ {
// 			r = &r.Next[i]
// 			r.Path = NewPathIndex(i)
// 			toPathNode(vt.Index(i).Interface(), r)
// 		}
// 		return
// 	} else if k == reflect.Map {
// 		if vt.Len() == 0 {
// 			return
// 		}
// 		if len(ret.Next) < vt.Len() {
// 			ret.Next = make([]PathNode, vt.Len())
// 		}
// 		if kt := vt.Type().Key().Kind(); kt == reflect.String {
// 			var r *PathNode
// 			for i, k := range vt.MapKeys() {
// 				r = &r.Next[i]
// 				r.Path = NewPathStrKey(k.String())
// 				toPathNode(vt.MapIndex(k).Interface(), r)
// 			}
// 			return
// 		} else if kt == reflect.Int || kt == reflect.Int8 || kt == reflect.Int16 || kt == reflect.Int32 || kt == reflect.Int64 {
// 			var r *PathNode
// 			for i, k := range vt.MapKeys() {
// 				r = &r.Next[i]
// 				r.Path = NewPathIntKey(int(k.Int()))
// 				toPathNode(vt.MapIndex(k).Interface(), r)
// 			}
// 			return
// 		}
// 	} else if k == reflect.Struct {
// 		var r *PathNode
// 		if vt.NumField() == 0 {
// 			return
// 		}
// 		if len(ret.Next) < vt.NumField() {
// 			ret.Next = make([]PathNode, vt.NumField())
// 		}
// 		for i := 0; i < vt.NumField(); i++ {
// 			r = &r.Next[i]
// 			field := vt.Type().Field(i)
// 			tag := field.Tag.Get("thrift")
// 			ts := strings.Split(tag, ",")
// 			id := i
// 			if len(ts) > 1 {
// 				id, _ = strconv.Atoi(ts[1])
// 			}
// 			r.Path = NewPathFieldId(thrift.FieldID(id))
// 			toPathNode(vt.Field(i).Interface(), r)
// 		}
// 		return
// 	} else {
// 		ret.Node = NewNodeAny(vt.Interface())
// 		return
// 	}
// }

func pathNodeToInterface(tree PathNode, opts *Options, useParent bool) interface{} {
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
			vv := pathNodeToInterface(v, opts, useParent)
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
			vv := pathNodeToInterface(v, opts, useParent)
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
				vv := pathNodeToInterface(v, opts, useParent)
				if vv == nil {
					continue
				}
				ret[v.Path.Str()] = vv
			}
			return ret
		} else if kt.IsInt() {
			ret := make(map[int]interface{}, len(tree.Next))
			for _, v := range tree.Next {
				vv := pathNodeToInterface(v, opts, useParent)
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
				kv := pathNodeToInterface(kp, opts, true)
				if kv == nil {
					continue
				}
				vv := pathNodeToInterface(v, opts, useParent)
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

func DeepEqual(exp interface{}, act interface{}) bool {
	switch ev := exp.(type) {
	case map[int]interface{}:
		av, ok := act.(map[int]interface{})
		if !ok {
			return false
		}
		for k, v := range ev {
			vv, ok := av[k]
			if !ok {
				return false
			}
			if !DeepEqual(v, vv) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		av, ok := act.(map[string]interface{})
		if !ok {
			return false
		}
		for k, v := range ev {
			vv, ok := av[k]
			if !ok {
				return false
			}
			if !DeepEqual(v, vv) {
				return false
			}
		}
		return true
	case map[interface{}]interface{}:
		av, ok := act.(map[interface{}]interface{})
		if !ok {
			return false
		}
		if len(ev) == 0 {
			return true
		}
		erv := reflect.ValueOf(ev)
		arv := reflect.ValueOf(av)
		eks := erv.MapKeys()
		aks := arv.MapKeys()
		isPointer := eks[0].Elem().Kind() == reflect.Ptr
		if !isPointer {
			for k, v := range ev {
				vv, ok := av[k]
				if !ok {
					return false
				}
				if !DeepEqual(v, vv) {
					return false
				}
			}
		} else {
			for _, ek := range eks {
				found := false
				for _, ak := range aks {
					if DeepEqual(ek.Elem().Elem().Interface(), ak.Elem().Elem().Interface()) {
						found = true
						evv := erv.MapIndex(ek)
						avv := arv.MapIndex(ak)
						if !DeepEqual(evv.Interface(), avv.Interface()) {
							return false
						}
					}
					if !found {
						return false
					}
				}
			}
		}
		return true
	case []interface{}:
		av, ok := act.([]interface{})
		if !ok {
			return false
		}
		for i, v := range ev {
			vv := av[i]
			if !DeepEqual(v, vv) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(exp, act)
	}
}

func TestDeepEqual(t *testing.T) {
	a := map[interface{}]interface{}{
		float64(0.1): "A",
		float64(0.2): "B",
		float64(0.3): "C",
		float64(0.4): "D",
		float64(0.5): "E",
		float64(0.6): "F",
		float64(0.7): "G",
		float64(0.8): "H",
		float64(0.9): "I",
	}
	b := map[interface{}]interface{}{
		float64(0.4): "D",
		float64(0.8): "H",
		float64(0.7): "G",
		float64(0.5): "E",
		float64(0.6): "F",
		float64(0.9): "I",
		float64(0.2): "B",
		float64(0.1): "A",
		float64(0.3): "C",
	}
	for i := 0; i < 10; i++ {
		require.Equal(t, a, b)
	}
	require.True(t, DeepEqual(a, b))
}

func TestUnknownFields(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleSuperData()
	v := NewValue(desc, data)

	t.Run("Children()", func(t *testing.T) {
		// t.Run("allow", func(t *testing.T) {
		children := make([]PathNode, 0)
		opts := Options{}
		err := v.Children(&children, true, &opts)
		require.Nil(t, err)
		act := pathNodeToInterface(PathNode{Node: v.Node, Next: children}, &opts, false)
		exp := toInterface2(sample.Example2Super, false, b2s)
		if !DeepEqual(exp, act) {
			t.Fatal()
		}
		if !DeepEqual(act, exp) {
			t.Fatal()
		}
		// require.Equal(t, exp, act)
		// })

		// t.Run("disallow", func(t *testing.T) {
		// children := make([]PathNode, 0)
		// opts := Options{
		// 	DisallowUnknow: true,
		// }
		// err := v.Children(&children, false, &opts)
		// require.NotNil(t, err)
		// require.Equal(t, meta.ErrUnknownField, err.(meta.Error).Code.Behavior())
		// })
	})

	t.Run("Assgin(true, )", func(t *testing.T) {
		// t.Run("allow", func(t *testing.T) {
		opts := Options{
			// DisallowUnknow: false,
		}
		path := PathNode{
			Node: v.Node,
		}
		err := DescriptorToPathNode(desc, &path, &opts)
		if err != nil {
			t.Fatal(err)
		}
		err = path.Assgin(true, &opts)
		require.NoError(t, err)
		act := pathNodeToInterface(path, &opts, true)
		exp := toInterface2(sample.Example2Obj, false, b2s)
		if !DeepEqual(exp, act) {
			t.Fatal()
		}
		if !DeepEqual(act, exp) {
			t.Fatal()
		}

		// })
		// t.Run("disallow", func(t *testing.T) {
		// 	opts := Options{
		// 		DisallowUnknow: true,
		// 	}
		// 	path := PathNode{
		// 		Node: v.Node,
		// 	}
		// 	err := DescriptorToPathNode(desc, &path, &opts)
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}
		// 	err = path.Assgin(true, &opts)
		// 	require.Error(t, err)
		// 	require.Equal(t, meta.ErrUnknownField, err.(Value).ErrCode())
		// })
	})

	// t.Run("Interface()/ByName", func(t *testing.T) {
	// 	// t.Run("allow", func(t *testing.T) {
	// 	opts := Options{
	// 		DisallowUnknow: false,
	// 		StructByName:   true,
	// 	}
	// 	ret, err := v.Interface(&opts)
	// 	require.NoError(t, err)
	// 	rv := toInterface2(sample.Example2Obj, false, true)
	// 	require.Equal(t, rv, ret)
	// })
	// t.Run("disallow", func(t *testing.T) {
	// 	opts := Options{
	// 		DisallowUnknow: true,
	// 		StructByName:   true,
	// 	}
	// 	_, err := v.Interface(&opts)
	// 	require.Error(t, err)
	// 	require.Equal(t, meta.ErrUnknownField, err.(Value).ErrCode())
	// })
	// })

	t.Run("Interface()", func(t *testing.T) {
		// t.Run("allow", func(t *testing.T) {
		opts := Options{
			// DisallowUnknow: false,
			// StructByName:   false,
		}
		ret, err := v.Interface(&opts)
		require.NoError(t, err)
		rv := toInterface2(sample.Example2Super, false, b2s)
		if !DeepEqual(rv, ret) {
			t.Fatal()
		}
		if !DeepEqual(ret, rv) {
			t.Fatal()
		}
		// })
		// t.Run("disallow", func(t *testing.T) {
		// 	opts := Options{
		// 		DisallowUnknow: true,
		// 		StructByName:   false,
		// 	}
		// 	_, err := v.Interface(&opts)
		// 	require.Error(t, err)
		// 	require.Equal(t, meta.ErrUnknownField, err.(Value).ErrCode())
		// })
	})

	t.Run("Marshal()", func(t *testing.T) {
		children := make([]PathNode, 0)
		opts := Options{}
		err := v.Children(&children, true, &opts)
		require.Nil(t, err)
		tree := PathNode{
			Node: v.Node,
			Next: children,
		}
		out, err := tree.Marshal(&opts)
		require.NoError(t, err)

		act := example2.NewExampleSuper()
		_, err = act.FastRead(out)
		require.NoError(t, err)

		// require.Equal(t, sample.Example2Super, act)
	})
}
