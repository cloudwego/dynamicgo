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
	"testing"

	"github.com/cloudwego/dynamicgo/testdata/sample"
	"github.com/stretchr/testify/require"
)

func TestForeach(t *testing.T) {
	var iterOpts = &Options{
		IterateStructByName: true,
	}

	desc := getExampleDesc()
	data := getExampleData()
	root := NewValue(desc, data)

	var exp = PathNode{
		Path: NewPathFieldName("root"),
		Node: root.Node,
		Next: []PathNode{},
	}
	require.Nil(t, root.Children(&exp.Next, true, iterOpts))

	t.Run("Node", func(t *testing.T) {
		var handler func(path Path, node Node) bool
		var tree = PathNode{
			Path: NewPathFieldName("root"),
			Node: root.Node,
		}
		var cur = &tree
		handler = func(path Path, node Node) bool {
			cur.Next = append(cur.Next, PathNode{
				Path: path,
				Node: node,
			})
			if node.Type().IsComplex() {
				old := cur
				cur = &cur.Next[len(cur.Next)-1]
				require.Nil(t, node.Foreach(handler, iterOpts))
				cur = old
			}
			return true
		}
		require.NoError(t, root.Node.Foreach(handler, iterOpts))
		require.Equal(t, exp, tree)

		cout := 0
		handler3 := func(path Path, node Node) bool {
			cout += 1
			return false
		}
		vv := root.GetByPath(NewPathFieldName("InnerBase"))
		require.Nil(t, vv.Check())
		require.NoError(t, vv.Node.Foreach(handler3, iterOpts))
		require.Equal(t, 1, cout)

		cout = 0
		vv = root.GetByPath(PathExampleListInt32...)
		require.Nil(t, vv.Check())
		require.NoError(t, vv.Node.Foreach(handler3, iterOpts))
		require.Equal(t, 1, cout)

		cout = 0
		vv = root.GetByPath(PathExampleMapInt32String...)
		require.Nil(t, vv.Check())
		require.NoError(t, vv.Node.Foreach(handler3, iterOpts))
		require.Equal(t, 1, cout)

		cout = 0
		vv = root.GetByPath(PathExampleMapStringString...)
		require.Nil(t, vv.Check())
		require.NoError(t, vv.Node.Foreach(handler3, iterOpts))
		require.Equal(t, 1, cout)
	})

	t.Run("Value", func(t *testing.T) {
		handler := func(path Path, val Value) bool {
			vv, err := val.Interface(iterOpts)
			require.NoError(t, err)
			t.Logf("Path: %v, Value: %v", path, vv)
			return true
		}
		require.NoError(t, root.Foreach(handler, iterOpts))

		handler2 := func(path Path, val Value) bool {
			require.True(t, path.Type() == PathFieldName)
			t.Logf("handler2: %v", path)
			return true
		}
		vv := root.GetByPath(NewPathFieldName("InnerBase"))
		require.Nil(t, vv.Check())
		require.NoError(t, vv.Foreach(handler2, iterOpts))

		cout := 0
		handler3 := func(path Path, val Value) bool {
			cout += 1
			return false
		}
		cout = 0
		vv = root.GetByPath(PathExampleListInt32...)
		require.Nil(t, vv.Check())
		require.NoError(t, vv.Foreach(handler3, iterOpts))
		require.Equal(t, 1, cout)

		cout = 0
		vv = root.GetByPath(PathExampleMapInt32String...)
		require.Nil(t, vv.Check())
		require.NoError(t, vv.Foreach(handler3, iterOpts))
		require.Equal(t, 1, cout)

		cout = 0
		vv = root.GetByPath(PathExampleMapStringString...)
		require.Nil(t, vv.Check())
		require.NoError(t, vv.Foreach(handler3, iterOpts))
		require.Equal(t, 1, cout)
	})
	
}

func TestForeachKV(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()
	root := NewValue(desc, data)
	opts := &Options{}
	
	t.Run("Node", func(t *testing.T) {
		v := root.GetByPath(PathExampleListInt32...)
		require.Nil(t, v.Check())
		err := v.Node.ForeachKV(func(key, val Node) bool { return true }, opts)
		require.Error(t, err)
		v = root.GetByPath(PathExampleMapInt32String...)
		require.Nil(t, v.Check())
		err = v.Node.ForeachKV(func(key, val Node) bool {
			k, err := key.Int()
			require.NoError(t, err)
			v, err := val.String()
			require.NoError(t, err)
			exp := sample.Example2Obj.InnerBase.MapInt32String[int32(k)]
			require.Equal(t, exp, v)
			return true
		}, opts)
		require.NoError(t, err)
	})

	t.Run("Value", func(t *testing.T) {
		v := root.GetByPath(PathExampleListInt32...)
		require.Nil(t, v.Check())
		err := v.Node.ForeachKV(func(key, val Node) bool { return true }, opts)
		require.Error(t, err)
		v = root.GetByPath(PathExampleMapInt32String...)
		require.Nil(t, v.Check())
		err = v.ForeachKV(func(key, val Value) bool {
			k, err := key.Int()
			require.NoError(t, err)
			v, err := val.String()
			require.NoError(t, err)
			exp := sample.Example2Obj.InnerBase.MapInt32String[int32(k)]
			require.Equal(t, exp, v)
			return true
		}, opts)
		require.NoError(t, err)
	})
	
}
