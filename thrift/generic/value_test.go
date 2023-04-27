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
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"reflect"
	"testing"

	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example2"
	"github.com/cloudwego/dynamicgo/testdata/sample"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exampleIDLPath         = "../../testdata/idl/example2.thrift"
	exampleThriftPath      = "../../testdata/data/example2.bin"
	exampleSuperThriftPath = "../../testdata/data/example2super.bin"
)

func GetSampleTree(v Value) PathNode {
	return PathNode{
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
					},
					{
						Path: NewPathFieldId(9),
					},
				},
			},
			{
				Path: NewPathFieldId(255),
				Next: []PathNode{
					{
						Path: NewPathFieldId(5),
					},
				},
			},
		},
	}
}

func GetSampleTreeById(v Value) PathNode {
	return PathNode{
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
					},
					{
						Path: NewPathFieldId(9),
					},
				},
			},
			{
				Path: NewPathFieldId(255),
				Next: []PathNode{
					{
						Path: NewPathFieldId(4),
					},
				},
			},
		},
	}
}

func TestCount(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()
	fmt.Printf("data len: %d\n", len(data))
	v := NewValue(desc, data)
	children := make([]PathNode, 0, 4)
	opts := Options{}
	if err := v.Children(&children, true, &opts); err != nil {
		t.Fatal(err)
	}
	count := 1
	countHelper(&count, children)
	fmt.Printf("nodes count: %d", count)
}

func countHelper(count *int, ps []PathNode) {
	*count += len(ps)
	for _, p := range ps {
		countHelper(count, p.Next)
	}
}

func getExampleDesc() *thrift.TypeDescriptor {
	svc, err := thrift.NewDescriptorFromPath(context.Background(), exampleIDLPath)
	if err != nil {
		panic(err)
	}
	return svc.Functions()["ExampleMethod"].Request().Struct().FieldByKey("req").Type()
}

func getExamplePartialDesc() *thrift.TypeDescriptor {
	svc, err := thrift.NewDescriptorFromPath(context.Background(), exampleIDLPath)
	if err != nil {
		panic(err)
	}
	return svc.Functions()["ExamplePartialMethod"].Request().Struct().FieldByKey("req").Type()
}

func getExampleData() []byte {
	out, err := ioutil.ReadFile(exampleThriftPath)
	if err != nil {
		panic(err)
	}
	return out
}

func getExampleSuperData() []byte {
	out, err := ioutil.ReadFile(exampleSuperThriftPath)
	if err != nil {
		panic(err)
	}
	return out
}

func empty(v interface{}) interface{} {
	et := reflect.TypeOf(v)
	if et.Kind() == reflect.Ptr {
		et = et.Elem()
	}
	return reflect.New(et).Interface()
}

func checkHelper(t *testing.T, exp interface{}, act Node, api string) {
	v := reflect.ValueOf(act)
	f := v.MethodByName(api)
	if f.Kind() != reflect.Func {
		t.Fatalf("method %s not found", api)
	}
	var args []reflect.Value
	if api == "List" || api == "StrMap" || api == "IntMap" || api == "Interface" {
		args = make([]reflect.Value, 1)
		args[0] = reflect.ValueOf(&Options{})
	} else {
		args = make([]reflect.Value, 0)
	}
	rets := f.Call(args)
	if len(rets) != 2 {
		t.Fatal("wrong number of return values")
	}
	require.Nil(t, rets[1].Interface())
	switch api {
	case "List":
		vs := rets[0]
		if vs.Kind() != reflect.Slice {
			t.Fatal("wrong type")
		}
		es := reflect.ValueOf(exp)
		if es.Kind() != reflect.Slice {
			t.Fatal("wrong type")
		}
		for i := 0; i < vs.Len(); i++ {
			vv := vs.Index(i)
			require.Equal(t, cast(es.Index(i).Interface(), vv.Type().Name() == "byte"), vv.Interface())
		}
	case "StrMap":
		vs := rets[0]
		if vs.Kind() != reflect.Map {
			t.Fatal("wrong type")
		}
		es := reflect.ValueOf(exp)
		if es.Kind() != reflect.Map {
			t.Fatal("wrong type")
		}
		ks := vs.MapKeys()
		for i := 0; i < len(ks); i++ {
			vv := vs.MapIndex(ks[i])
			require.Equal(t, cast(es.MapIndex(ks[i]).Interface(), vv.Type().Name() == "byte"), vv.Interface())
		}
	default:
		require.Equal(t, exp, rets[0].Interface())
	}
}

func cast(ev interface{}, b bool) interface{} {
	switch ev.(type) {
	case int8:
		if b {
			return byte(ev.(int8))
		}
		return int(ev.(int8))
	case int16:
		return int(ev.(int16))
	case int32:
		return int(ev.(int32))
	case int64:
		return int(ev.(int64))
	case float32:
		return float64(ev.(float32))
	default:
		return ev
	}
}

func TestGet(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()

	exp := example2.NewExampleReq()
	v := NewValue(desc, data)
	_, err := exp.FastRead(data)
	require.Nil(t, err)

	t.Run("GetByStr()", func(t *testing.T) {
		v := v.GetByPath(PathExampleMapStringString...)
		require.Nil(t, v.Check())
		v1, err := v.GetByStr("b").String()
		require.NoError(t, err)
		require.Equal(t, sample.Example2Obj.InnerBase.MapStringString["b"], v1)
		v2, err := v.GetByStr("aa").String()
		require.Error(t, err)
		require.Equal(t, meta.ErrNotFound, err.(Node).ErrCode())
		require.Equal(t, sample.Example2Obj.InnerBase.MapStringString["aa"], v2)
	})

	t.Run("GetByInt()", func(t *testing.T) {
		v := v.GetByPath(PathExampleMapInt32String...)
		require.Nil(t, v.Check())
		v1, err := v.GetByInt(1).String()
		require.NoError(t, err)
		require.Equal(t, sample.Example2Obj.InnerBase.MapInt32String[1], v1)
		v2, err := v.GetByInt(999).String()
		require.Error(t, err)
		require.Equal(t, meta.ErrNotFound, err.(Node).ErrCode())
		require.Equal(t, sample.Example2Obj.InnerBase.MapInt32String[999], v2)
	})

	t.Run("Index()", func(t *testing.T) {
		v := v.GetByPath(PathExampleListInt32...)
		require.Nil(t, v.Check())
		v1, err := v.Index(1).Int()
		require.NoError(t, err)
		require.Equal(t, int(sample.Example2Obj.InnerBase.ListInt32[1]), v1)
		v2 := v.Index(999)
		require.Error(t, v2)
		require.Equal(t, meta.ErrInvalidParam, v2.ErrCode())
	})

	t.Run("FieldByName()", func(t *testing.T) {
		_, err := v.FieldByName("Msg2").String()
		require.NotNil(t, err)
		s, err := v.FieldByName("Msg").String()
		require.Equal(t, *exp.Msg, s)
	})

	t.Run("Field()", func(t *testing.T) {
		xx, err := v.Field(222).Int()
		require.NotNil(t, err)
		require.Equal(t, int(0), xx)
		a := v.Field(3)
		b, err := a.Field(1).Bool()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase.Bool, b)
		c, err := a.Field(2).Byte()
		require.Nil(t, err)
		require.Equal(t, byte(exp.InnerBase.Byte), c)
		d, err := a.Field(3).Int()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase.Int16, int16(d))
		e, err := a.Field(4).Int()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase.Int32, int32(e))
		f, err := a.Field(5).Int()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase.Int64, int64(f))
		g, err := a.Field(6).Float64()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase.Double, float64(g))
		h, err := a.Field(7).String()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase.String_, string(h))
		list := a.Field(8)
		checkHelper(t, exp.InnerBase.ListInt32, list.Node, "List")
		list1, err := a.Field(8).Index(1).Int()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase.ListInt32[1], int32(list1))
		mp := a.Field(9)
		checkHelper(t, exp.InnerBase.MapStringString, mp.Node, "StrMap")
		mp1, err := a.Field(9).GetByStr("b").String()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase.MapStringString["b"], (mp1))
		sp := a.Field(10)
		checkHelper(t, exp.InnerBase.SetInt32_, sp.Node, "List")
		i, err := a.Field(11).Int()
		require.Nil(t, err)
		require.Equal(t, int64(exp.InnerBase.Foo), int64(i))
		mp2, err := a.Field(12).GetByInt(2).String()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase.MapInt32String[2], (mp2))
	})

	t.Run("GetByPath()", func(t *testing.T) {
		exp := sample.Example2Obj.InnerBase.ListInt32[1]

		v1 := v.GetByPath(NewPathFieldId(thrift.FieldID(3)), NewPathFieldId(8), NewPathIndex(1))
		if v1.Error() != "" {
			t.Fatal(v1.Error())
		}
		act, err := v1.Int()
		require.NoError(t, err)
		require.Equal(t, int(exp), act)

		v2 := v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"), NewPathIndex(1))
		if v2.Error() != "" {
			t.Fatal(v2.Error())
		}
		require.Equal(t, v1, v2)
		v3 := v.GetByPath(NewPathFieldId(thrift.FieldID(3)), NewPathFieldName("ListInt32"), NewPathIndex(1))
		if v3.Error() != "" {
			t.Fatal(v3.Error())
		}
		require.Equal(t, v1, v3)
		v4 := v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldId(8), NewPathIndex(1))
		if v4.Error() != "" {
			t.Fatal(v4.Error())
		}
		require.Equal(t, v1, v4)

		exp2 := sample.Example2Obj.InnerBase.MapInt32String[2]
		v5 := v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldId(12), NewPathIntKey(2))
		if v5.Error() != "" {
			t.Fatal(v5.Error())
		}
		act2, err := v5.String()
		require.NoError(t, err)
		require.Equal(t, exp2, act2)

		v6 := v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("MapInt32String"), NewPathIntKey(2))
		if v6.Error() != "" {
			t.Fatal(v6.Error())
		}
		require.Equal(t, v5, v6)

		exp3 := sample.Example2Obj.InnerBase.MapStringString["b"]
		v7 := v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldId(9), NewPathStrKey("b"))
		if v5.Error() != "" {
			t.Fatal(v7.Error())
		}
		act3, err := v7.String()
		require.NoError(t, err)
		require.Equal(t, exp3, act3)

		v8 := v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("MapStringString"), NewPathStrKey("b"))
		if v6.Error() != "" {
			t.Fatal(v8.Error())
		}
		require.Equal(t, v8, v7)
	})
}

func handlePartialMapStringString2(p map[int]interface{}) {
	delete(p, 127)
	if f18 := p[18]; f18 != nil {
		for i := range f18.([]interface{}) {
			pv := f18.([]interface{})[i].(map[int]interface{})
			handlePartialMapStringString2(pv)
		}
	}
	if f19 := p[19]; f19 != nil {
		for k := range f19.(map[interface{}]interface{}) {
			handlePartialMapStringString2(*(k.(*map[int]interface{})))
			vv := f19.(map[interface{}]interface{})[k].(map[int]interface{})
			handlePartialMapStringString2(vv)
		}
	}
}

func TestMarshalTo(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()
	partial := getExamplePartialDesc()

	exp := example2.NewExampleReq()
	v := NewValue(desc, data)
	_, err := exp.FastRead(data)
	require.Nil(t, err)

	t.Run("ById", func(t *testing.T) {
		t.Run("WriteDefault", func(t *testing.T) {
			opts := &Options{WriteDefault: true}
			buf, err := v.MarshalTo(partial, opts)
			require.Nil(t, err)
			ep := example2.NewExampleReqPartial()
			_, err = ep.FastRead(buf)
			require.Nil(t, err)
			act := toInterface(ep)
			exp := toInterface(exp)
			require.False(t, DeepEqual(act, exp))
			handlePartialMapStringString2(act.(map[int]interface{})[3].(map[int]interface{}))
			require.True(t, DeepEqual(act, exp))
			require.NotNil(t, ep.InnerBase.MapStringString2)
		})
		t.Run("NotWriteDefault", func(t *testing.T) {
			opts := &Options{}
			buf, err := v.MarshalTo(partial, opts)
			require.Nil(t, err)
			ep := example2.NewExampleReqPartial()
			_, err = ep.FastRead(buf)
			require.Nil(t, err)
			act := toInterface(ep)
			exp := toInterface(exp)
			require.False(t, DeepEqual(act, exp))
			handlePartialMapStringString2(act.(map[int]interface{})[3].(map[int]interface{}))
			require.True(t, DeepEqual(act, exp))
			require.Nil(t, ep.InnerBase.MapStringString2)
		})
	})

	t.Run("unknown", func(t *testing.T) {
		n := NewNode(thrift.STRUCT, data)
		exist, err := n.SetByPath(NewNodeString("Insert"), NewPathFieldId(1024))
		require.Nil(t, err)
		require.False(t, exist)
		data = n.raw()
		v.Node = NewNode(thrift.STRUCT, data)
		t.Run("allow unknown", func(t *testing.T) {
			opts := &Options{DisallowUnknow: false}
			buf, err := v.MarshalTo(partial, opts)
			require.NoError(t, err)
			ep := example2.NewExampleReqPartial()
			_, err = ep.FastRead(buf)
			require.Nil(t, err)
			act := toInterface(ep)
			exp := toInterface(exp)
			require.False(t, DeepEqual(act, exp))
			handlePartialMapStringString2(act.(map[int]interface{})[3].(map[int]interface{}))
			require.True(t, DeepEqual(act, exp))
		})
		t.Run("disallow unknown", func(t *testing.T) {
			opts := &Options{DisallowUnknow: true}
			_, err := v.MarshalTo(partial, opts)
			require.Error(t, err)
		})
	})

	// t.Run("ByName", func(t *testing.T) {
	// 	opts := &Options{WriteDefault: true}
	// 	// opts.FieldByName = true
	// 	buf, err := v.MarshalTo(partial, opts)
	// 	require.Nil(t, err)
	// 	ep := example2.NewExampleReqPartial()
	// 	_, err = ep.FastRead(buf)
	// 	require.Nil(t, err)
	// 	checkHelper3(t, exp, ep)
	// 	require.NotNil(t, ep.InnerBase.MapStringString2)
	// })
}

func checkHelper3(t *testing.T, exp, act interface{}) {
	ev := reflect.ValueOf(exp)
	av := reflect.ValueOf(act)
	if ev.Kind() != av.Kind() {
		t.Fatalf("%v != %v", ev.Kind(), av.Kind())
	}
	if ev.Kind() == reflect.Ptr || av.Kind() == reflect.Ptr {
		ev = ev.Elem()
		av = av.Elem()
		checkHelper3(t, ev.Interface(), av.Interface())
		return
	}
	switch av.Kind() {
	case reflect.Struct:
		for i := 0; i < ev.NumField(); i++ {
			ef := ev.Type().Field(i)
			af := av.FieldByName(ef.Name)
			if reflect.DeepEqual(af, reflect.Value{}) {
				continue
			}
			checkHelper3(t, ev.Field(i).Interface(), af.Interface())
		}
	default:
		assert.Equal(t, exp, act)
	}
}

func TestSetMany(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()
	opts := &Options{
		UseNativeSkip: true,
	}

	r := NewValue(desc, data)
	d1 := desc.Struct().FieldByKey("Msg").Type()
	d2 := desc.Struct().FieldByKey("Subfix").Type()

	t.Run("insert", func(t *testing.T) {
		v := r.Fork()
		exp1 := int32(-1024)
		n1 := NewNodeInt32(exp1)
		err := v.SetMany([]PathNode{
			{
				Path: NewPathFieldId(2),
				Node: n1,
			},
		}, opts)
		require.Nil(t, err)
		v1 := v.GetByPath(NewPathFieldName("A"))
		act1, err := v1.Int()
		require.NoError(t, err)
		require.Equal(t, exp1, int32(act1))

		exp21 := int32(math.MinInt32)
		exp22 := int32(math.MaxInt32)
		n21 := NewNodeInt32(exp21)
		n22 := NewNodeInt32(exp22)
		vv := v.GetByPath(PathExampleListInt32...)
		l2, err := vv.Len()
		require.NoError(t, err)
		err = vv.SetMany([]PathNode{
			{
				Path: NewPathIndex(1024),
				Node: n21,
			},
			{
				Path: NewPathIndex(1024),
				Node: n22,
			},
		}, opts)
		require.NoError(t, err)
		v21 := vv.GetByPath(NewPathIndex(0))
		act21, err := v21.Int()
		require.NoError(t, err)
		require.Equal(t, exp21, int32(act21))
		v22 := vv.GetByPath(NewPathIndex(1))
		act22, err := v22.Int()
		require.NoError(t, err)
		require.Equal(t, exp22, int32(act22))
		ll2, err := vv.Len()
		require.Nil(t, err)
		require.Equal(t, l2+2, ll2)
	})

	t.Run("replace", func(t *testing.T) {
		v := r.Fork()
		exp1 := "exp1"
		exp2 := float64(-255.0001)
		p := thrift.NewBinaryProtocolBuffer()
		p.WriteString(exp1)

		v1 := NewValue(d1, []byte(string(p.Buf)))
		p.Buf = p.Buf[:0]
		p.WriteDouble(exp2)
		v2 := NewValue(d2, []byte(string(p.Buf)))
		err := v.SetMany([]PathNode{
			{
				Path: NewPathFieldId(1),
				Node: v1.Node,
			},
			{
				Path: NewPathFieldId(32767),
				Node: v2.Node,
			},
		}, opts)
		require.Nil(t, err)

		exp := example2.NewExampleReq()
		if _, err := exp.FastRead(v.Raw()); err != nil {
			t.Fatal(err)
		}
		require.Equal(t, *exp.Msg, exp1)
		require.Equal(t, exp.Subfix, exp2)

		vx := v.GetByPath(NewPathFieldName("Base"))
		vv := vx
		err = vv.SetMany([]PathNode{
			{
				Path: NewPathFieldId(1),
				Node: v.FieldByName("Base").FieldByName("LogID").Node,
			},
			{
				Path: NewPathFieldId(2),
				Node: v.FieldByName("Base").FieldByName("Caller").Node,
			},
			{
				Path: NewPathFieldId(3),
				Node: v.FieldByName("Base").FieldByName("Addr").Node,
			},
			{
				Path: NewPathFieldId(4),
				Node: v.FieldByName("Base").FieldByName("Client").Node,
			},
			{
				Path: NewPathFieldId(5),
				Node: v.FieldByName("Base").FieldByName("TrafficEnv").Node,
			},
			{
				Path: NewPathFieldId(6),
				Node: v.FieldByName("Base").FieldByName("Extra").Node,
			},
		}, opts)
		require.Nil(t, err)
		require.Equal(t, vx.raw(), vv.raw())

		inner := v.GetByPath(NewPathFieldName("InnerBase"))
		p = thrift.NewBinaryProtocolBuffer()
		e1 := false
		p.WriteBool(e1)
		v1 = NewValue(d1, []byte(string(p.Buf)))
		p.Buf = p.Buf[:0]
		e2 := float64(-255.0001)
		p.WriteDouble(e2)
		v2 = NewValue(d1, []byte(string(p.Buf)))
		p.Buf = p.Buf[:0]
		err = inner.SetMany([]PathNode{
			{
				Path: NewPathFieldId(1),
				Node: v1.Node,
			},
			{
				Path: NewPathFieldId(6),
				Node: v2.Node,
			},
			{
				Path: NewPathFieldId(255),
				Node: vx.Node,
			},
		}, opts)
		require.Nil(t, err)
		// spew.Dump(inner.Raw())
		expx := example2.NewInnerBase()
		if _, err := expx.FastRead(inner.Raw()); err != nil {
			t.Fatal(err)
		}
		require.Equal(t, expx.Bool, e1)
		require.Equal(t, expx.Double, e2)
		require.Equal(t, expx.Base, exp.Base)
	})

}

func TestUnsetByPath(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()
	r := NewValue(desc, data)

	v := r.Fork()
	err := v.UnsetByPath()
	require.Nil(t, err)
	require.True(t, v.IsEmpty())

	v = r.Fork()
	n := v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Caller"))
	require.False(t, n.IsError())
	err = v.UnsetByPath(NewPathFieldName("Base"), NewPathFieldName("Caller"))
	require.NoError(t, err)
	n = v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Caller"))
	require.True(t, n.IsErrNotFound())

	v = r.Fork()
	l, err := v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32")).Len()
	require.NoError(t, err)
	n = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"), NewPathIndex(l-1))
	require.False(t, n.IsError())
	err = v.UnsetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"), NewPathIndex(l-1))
	require.NoError(t, err)
	n = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"), NewPathIndex(l-1))
	require.True(t, n.IsErrNotFound())
	ll, err := v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32")).Len()
	require.NoError(t, err)
	require.Equal(t, l-1, ll)

	v = r.Fork()
	l, err = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("MapStringString")).Len()
	require.NoError(t, err)
	n = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("MapStringString"), NewPathStrKey("a"))
	require.False(t, n.IsError())
	err = v.UnsetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("MapStringString"), NewPathStrKey("a"))
	require.NoError(t, err)
	n = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("MapStringString"), NewPathStrKey("a"))
	require.True(t, n.IsErrNotFound())
	ll, err = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("MapStringString")).Len()
	require.NoError(t, err)
	require.Equal(t, l-1, ll)

	v = r.Fork()
	l, err = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("MapInt32String")).Len()
	require.NoError(t, err)
	n = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("MapInt32String"), NewPathIntKey(1))
	require.False(t, n.IsError())
	err = v.UnsetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("MapInt32String"), NewPathIntKey(1))
	require.NoError(t, err)
	n = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("MapInt32String"), NewPathIntKey(1))
	require.True(t, n.IsErrNotFound())
	ll, err = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("MapInt32String")).Len()
	require.NoError(t, err)
	require.Equal(t, l-1, ll)
}

func TestSetByPath(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()
	v := NewValue(desc, data)
	ds := desc.Struct().FieldByKey("Subfix").Type()
	d2 := desc.Struct().FieldByKey("Base").Type().Struct().FieldByKey("Extra").Type().Elem()
	d3 := desc.Struct().FieldByKey("InnerBase").Type().Struct().FieldByKey("ListInt32").Type().Elem()

	e, err := v.SetByPath(v)
	require.True(t, e)
	require.Nil(t, err)

	t.Run("replace", func(t *testing.T) {
		s := v.GetByPath(NewPathFieldName("Subfix"))
		require.Empty(t, s.Error())
		f, _ := s.Float64()
		require.Equal(t, math.MaxFloat64, f)
		exp := float64(-0.1)
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, math.Float64bits(exp))
		e, err := v.SetByPath(NewValue(ds, buf), NewPathFieldName("Subfix"))
		require.True(t, e)
		require.Nil(t, err)
		s = v.GetByPath(NewPathFieldName("Subfix"))
		require.Empty(t, s.Error())
		f, _ = s.Float64()
		require.Equal(t, exp, f)

		exp2 := "中文"
		p := thrift.NewBinaryProtocolBuffer()
		p.WriteString(exp2)
		e, err2 := v.SetByPath(NewValue(d2, p.Buf), NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("b"))
		require.True(t, e)
		require.Nil(t, err2)
		s2 := v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("b"))
		require.Empty(t, s2.Error())
		f2, _ := s2.String()
		require.Equal(t, exp2, f2)
		e, err2 = v.SetByPath(NewValue(d2, p.Buf), NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("b"))
		require.True(t, e)
		require.Nil(t, err2)
		s2 = v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("b"))
		require.Empty(t, s2.Error())
		f2, _ = s2.String()
		require.Equal(t, exp2, f2)

		exp3 := int32(math.MinInt32) + 1
		v3 := Value{NewNodeInt32(exp3), d3}
		ps := []Path{NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"), NewPathIndex(2)}
		parent := v.GetByPath(ps[:len(ps)-1]...)
		l3, err := parent.Len()
		require.Nil(t, err)
		e, err = v.SetByPath(v3, ps...)
		require.True(t, e)
		require.Nil(t, err)
		s3 := v.GetByPath(ps...)
		act3, _ := s3.Int()
		require.Equal(t, exp3, int32(act3))
		parent = v.GetByPath(ps[:len(ps)-1]...)
		l3a, err := parent.Len()
		require.Nil(t, err)
		require.Equal(t, l3, l3a)
	})

	t.Run("insert", func(t *testing.T) {
		s := v.GetByPath(NewPathFieldName("A"))
		require.True(t, s.IsErrNotFound())
		exp := int32(-1024)
		v1 := Value{NewNodeInt32(exp), d3}
		e, err := v.SetByPath(v1, NewPathFieldName("A"))
		require.False(t, e)
		require.Nil(t, err)
		s = v.GetByPath(NewPathFieldName("A"))
		require.Empty(t, s.Error())
		act, _ := s.Int()
		require.Equal(t, exp, int32(act))

		s2 := v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("x"))
		require.True(t, s2.IsErrNotFound())
		exp2 := "中文\bb"
		v2 := Value{NewNodeString(exp2), d3}
		e, err2 := v.SetByPath(v2, NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("x"))
		require.False(t, e)
		require.Nil(t, err2)
		s2 = v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("x"))
		require.Empty(t, s2.Error())
		act2, _ := s2.String()
		require.Equal(t, exp2, act2)

		parent := v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"))
		l3, err := parent.Len()
		require.Nil(t, err)
		s3 := v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"), NewPathIndex(1024))
		require.True(t, s3.IsErrNotFound())
		exp3 := rand.Int31()
		v3 := Value{NewNodeInt32(exp3), d3}
		e, err = v.SetByPath(v3, NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"), NewPathIndex(1024))
		require.False(t, e)
		require.NoError(t, err)
		s3 = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		act3, _ := s3.Int()
		require.Equal(t, exp3, int32(act3))
		parent = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"))
		l3a, err := parent.Len()
		require.Nil(t, err)
		require.Equal(t, l3+1, l3a)
		exp3 = rand.Int31()
		v3 = Value{NewNodeInt32(exp3), d3}
		e, err = v.SetByPath(v3, NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"), NewPathIndex(1024))
		require.False(t, e)
		require.NoError(t, err)
		s3 = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		act3, _ = s3.Int()
		require.Equal(t, exp3, int32(act3))
		parent = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"))
		l3a, err = parent.Len()
		require.Nil(t, err)
		require.Equal(t, l3+2, l3a)
	})
}
