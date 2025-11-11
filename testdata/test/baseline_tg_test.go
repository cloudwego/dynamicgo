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

package test

import (
	"testing"

	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/baseline"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/cloudwego/dynamicgo/thrift/generic"
	kg "github.com/cloudwego/kitex/pkg/generic"
	kd "github.com/cloudwego/kitex/pkg/generic/descriptor"
	"github.com/stretchr/testify/require"
)

func init() {
	sobj := getSimpleValue()
	println("small thrift data size: ", sobj.BLength())

	psobj := getPartialSimpleValue()
	println("partial small thrift data size: ", psobj.BLength())

	nobj := getNestingValue()
	println("medium thrift data size: ", nobj.BLength())

	pnobj := getPartialNestingValue()
	println("partial medium thrift data size: ", pnobj.BLength())
}

func getKitexGenericDesc() *kd.ServiceDescriptor {
	p, err := kg.NewThriftFileProvider(idlPath)
	if err != nil {
		panic(err.Error())
	}
	return <-p.Provide()
}

func BenchmarkThriftMarshalAll_KitexFast(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		obj := getSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data = data[:obj.BLength()]
			_ = obj.FastWriteNocopy(data, nil)
		}
	})
	b.Run("medium", func(b *testing.B) {
		obj := getNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data = data[:obj.BLength()]
			_ = obj.FastWriteNocopy(data, nil)
		}
	})
}

func BenchmarkThriftUnmarshalAll_KitexFast(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		obj := getSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = obj.FastRead(data)
		}
	})
	b.Run("medium", func(b *testing.B) {
		obj := getNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = obj.FastRead(data)
		}
	})
}

func BenchmarkThriftMarshalPartial_KitexFast(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		obj := getPartialSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data = data[:obj.BLength()]
			_ = obj.FastWriteNocopy(data, nil)
		}
	})
	b.Run("medium", func(b *testing.B) {
		obj := getPartialNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			data = data[:obj.BLength()]
			_ = obj.FastWriteNocopy(data, nil)
		}
	})
}

func BenchmarkThriftUnmarshalPartial_KitexFast(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		obj := getPartialSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = obj.FastRead(data)
		}
	})
	b.Run("medium", func(b *testing.B) {
		obj := getPartialNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = obj.FastRead(data)
		}
	})
}

func BenchmarkThriftMarshalTo_KitexFast(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		obj := getSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		pobj := getPartialSimpleValue()
		_, err := pobj.FastRead(data)
		require.Nil(b, err)
		pdata := make([]byte, pobj.BLength())
		ret = pobj.FastWriteNocopy(pdata, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pobj.FastRead(data)
			_ = pobj.FastWriteNocopy(pdata, nil)
		}
	})
	b.Run("medium", func(b *testing.B) {
		obj := getPartialNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		pobj := getPartialNestingValue()
		_, err := pobj.FastRead(data)
		require.Nil(b, err)
		pdata := make([]byte, pobj.BLength())
		ret = pobj.FastWriteNocopy(pdata, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = pobj.FastRead(data)
			_ = pobj.FastWriteNocopy(pdata, nil)
		}
	})
}

func BenchmarkThriftSkip(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getSimpleDesc()
		obj := getSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		p := thrift.NewBinaryProtocol(data)
		err := p.SkipType(desc.Type())
		require.Nil(b, err)
		require.Equal(b, len(data), p.Read)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.Read = 0
			_ = p.SkipType(desc.Type())
		}
	})

	b.Run("medium", func(b *testing.B) {
		desc := getNestingDesc()
		obj := getNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		p := thrift.NewBinaryProtocol(data)
		err := p.SkipType(desc.Type())
		require.Nil(b, err)
		require.Equal(b, len(data), p.Read)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.Read = 0
			_ = p.SkipType(desc.Type())
		}
	})
}

func BenchmarkThriftGetOne(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getSimpleDesc()
		obj := getSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		v := generic.NewValue(desc, data)
		vv := v.GetByPath(generic.NewPathFieldId(6))
		require.Nil(b, vv.Check())

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = v.GetByPath(generic.NewPathFieldId(6))
		}
	})

	b.Run("medium", func(b *testing.B) {
		desc := getNestingDesc()
		obj := getNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		v := generic.NewValue(desc, data)
		vv := v.GetByPath(generic.NewPathFieldId(15), generic.NewPathStrKey("15"), generic.NewPathFieldId(6))
		require.Nil(b, vv.Check())

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = v.GetByPath(generic.NewPathFieldId(15), generic.NewPathStrKey("15"), generic.NewPathFieldId(6))
		}
	})
}

func BenchmarkThriftGetMany(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getSimpleDesc()
		obj := getSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		opts := generic.Options{
			// StopScanOnError: true,
		}
		v := generic.NewValue(desc, data)
		ps := []generic.PathNode{
			{Path: generic.NewPathFieldId(1)},
			{Path: generic.NewPathFieldId(3)},
			{Path: generic.NewPathFieldId(6)},
		}
		err := v.GetMany(ps, &opts)
		require.Nil(b, err)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = v.GetMany(ps, &opts)
		}

	})

	b.Run("medium", func(b *testing.B) {
		desc := getNestingDesc()
		obj := getNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		opts := generic.Options{
			// StopScanOnError: true,
		}
		v := generic.NewValue(desc, data)
		ps := []generic.PathNode{
			{Path: generic.NewPathFieldId(2)},
			{Path: generic.NewPathFieldId(8)},
			{Path: generic.NewPathFieldId(15)},
		}
		err := v.GetMany(ps, &opts)
		require.Nil(b, err)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = v.GetMany(ps, &opts)
		}
	})
}

func BenchmarkThriftMarshalMany(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getSimpleDesc()
		obj := getSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		opts := generic.Options{
			// StopScanOnError: true,
		}
		v := generic.NewValue(desc, data)
		ps := []generic.PathNode{
			{Path: generic.NewPathFieldId(1)},
			{Path: generic.NewPathFieldId(3)},
			{Path: generic.NewPathFieldId(6)},
		}
		err := v.GetMany(ps, &opts)
		require.Nil(b, err)
		n := generic.PathNode{
			Path: generic.NewPathFieldId(1),
			Node: v.Node,
			Next: ps,
		}
		buf, err := n.Marshal(&opts)
		require.Nil(b, err)
		exp := baseline.NewPartialSimple()
		_, err = exp.FastRead(buf)
		require.Nil(b, err)
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = n.Marshal(&opts)
		}
	})

	b.Run("medium", func(b *testing.B) {
		desc := getNestingDesc()
		obj := getNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		opts := generic.Options{
			// StopScanOnError: true,
		}
		v := generic.NewValue(desc, data)
		ps := []generic.PathNode{
			{Path: generic.NewPathFieldId(2)},
			{Path: generic.NewPathFieldId(8)},
			{Path: generic.NewPathFieldId(15)},
		}
		err := v.GetMany(ps, &opts)
		require.Nil(b, err)
		n := generic.PathNode{
			Path: generic.NewPathFieldId(1),
			Node: v.Node,
			Next: ps,
		}
		buf, err := n.Marshal(&opts)
		require.Nil(b, err)
		exp := baseline.NewPartialNesting()
		_, err = exp.FastRead(buf)
		require.Nil(b, err)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = n.Marshal(&opts)
		}
	})
}

func BenchmarkThriftGetAll_New(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getSimpleDesc()
		obj := getSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		v := generic.NewValue(desc, data)
		out := []generic.PathNode{}
		require.Nil(b, v.Children(&out, false, &generic.Options{}))

		opts := &generic.Options{}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			out := []generic.PathNode{}
			_ = v.Children(&out, true, opts)
		}

	})

	b.Run("medium", func(b *testing.B) {
		desc := getNestingDesc()
		obj := getNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		v := generic.NewValue(desc, data)
		out := make([]generic.PathNode, 0, 16)
		require.Nil(b, v.Children(&out, false, &generic.Options{}))

		opts := &generic.Options{}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			out := []generic.PathNode{}
			_ = v.Children(&out, true, opts)
		}

	})
}

func BenchmarkThriftGetAll_ReuseMemory(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getSimpleDesc()
		obj := getSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		v := generic.NewValue(desc, data)
		r := generic.NewPathNode()
		r.Node = v.Node
		require.Nil(b, r.Load(true, &generic.Options{}))
		r.ResetAll()
		generic.FreePathNode(r)

		opts := &generic.Options{}
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r := generic.NewPathNode()
			r.Node = v.Node
			_ = r.Load(true, opts)
			r.ResetAll()
			generic.FreePathNode(r)
		}

		b.Run("normal", func(b *testing.B) {
			opts := &generic.Options{}
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r := generic.NewPathNode()
				r.Node = v.Node
				_ = r.Load(true, opts)
				r.ResetAll()
				generic.FreePathNode(r)
			}
		})
		b.Run("not_scan_parent", func(b *testing.B) {
			opts := &generic.Options{
				NotScanParentNode: true,
			}
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r := generic.NewPathNode()
				r.Node = v.Node
				_ = r.Load(true, opts)
				r.ResetAll()
				generic.FreePathNode(r)
			}
		})
	})

	b.Run("medium", func(b *testing.B) {
		desc := getNestingDesc()
		obj := getNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		v := generic.NewValue(desc, data)
		r := generic.NewPathNode()
		r.Node = v.Node
		require.Nil(b, r.Load(true, &generic.Options{}))
		r.ResetAll()
		generic.FreePathNode(r)

		b.Run("normal", func(b *testing.B) {
			opts := &generic.Options{}
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r := generic.NewPathNode()
				r.Node = v.Node
				_ = r.Load(true, opts)
				r.ResetAll()
				generic.FreePathNode(r)
			}
		})

		b.Run("not_scan_parent", func(b *testing.B) {
			opts := &generic.Options{
				NotScanParentNode: true,
			}
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r := generic.NewPathNode()
				r.Node = v.Node
				_ = r.Load(true, opts)
				r.ResetAll()
				generic.FreePathNode(r)
			}
		})
	})
}

func BenchmarkThriftMarshalAll(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getSimpleDesc()
		obj := getSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		v := generic.NewValue(desc, data)
		p := generic.PathNode{
			Node: v.Node,
		}
		opts := &generic.Options{}
		require.Nil(b, p.Load(true, opts))
		out, err := p.Marshal(opts)
		require.Nil(b, err)
		off, err := obj.FastRead(out)
		require.Nil(b, err)
		require.Equal(b, off, len(out))

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = p.Marshal(opts)
		}
	})

	b.Run("medium", func(b *testing.B) {
		desc := getNestingDesc()
		obj := getNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		v := generic.NewValue(desc, data)
		p := generic.PathNode{
			Node: v.Node,
		}
		opts := &generic.Options{}
		require.Nil(b, p.Load(true, opts))
		out, err := p.Marshal(opts)
		require.Nil(b, err)
		off, err := obj.FastRead(out)
		require.Nil(b, err)
		require.Equal(b, off, len(out))

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = p.Marshal(opts)
		}
	})
}

func BenchmarkThriftSetOne(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getSimpleDesc()
		obj := getSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		v := generic.NewValue(desc, data)
		p := thrift.NewBinaryProtocolBuffer()
		p.WriteBinary(obj.BinaryField)
		n := generic.NewValue(desc.Struct().FieldById(6).Type(), p.Buf)
		_, err := v.SetByPath(n, generic.NewPathFieldId(6))
		require.Nil(b, err)
		nn := v.GetByPath(generic.NewPathFieldId(6))
		require.Equal(b, n.Raw(), nn.Raw())

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = v.SetByPath(n, generic.NewPathFieldId(6))
		}
	})

	b.Run("medium", func(b *testing.B) {
		desc := getNestingDesc()
		obj := getNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		v := generic.NewValue(desc, data)
		p := thrift.NewBinaryProtocolBuffer()
		p.WriteBinary(obj.MapStringSimple["15"].BinaryField)
		n := generic.NewValue(desc.Struct().FieldById(15).Type().Elem().Struct().FieldById(6).Type(), p.Buf)
		_, err := v.SetByPath(n, generic.NewPathFieldId(15), generic.NewPathStrKey("15"), generic.NewPathFieldId(6))
		require.Nil(b, err)
		nn := v.GetByPath(generic.NewPathFieldId(15), generic.NewPathStrKey("15"), generic.NewPathFieldId(6))
		require.Equal(b, n.Raw(), nn.Raw())

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = v.SetByPath(n, generic.NewPathFieldId(15), generic.NewPathStrKey("15"), generic.NewPathFieldId(6))
		}
	})
}

func BenchmarkThriftSetMany(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getSimpleDesc()
		obj := getSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		opts := generic.Options{}
		v := generic.NewValue(desc, data)
		ps := []generic.PathNode{
			{Path: generic.NewPathFieldId(1)},
			{Path: generic.NewPathFieldId(3)},
			{Path: generic.NewPathFieldId(6)},
		}
		require.Nil(b, v.GetMany(ps, &opts))
		require.Nil(b, v.SetMany(ps, &opts))

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = v.SetMany(ps, &opts)
		}
	})

	b.Run("medium", func(b *testing.B) {
		desc := getNestingDesc()
		obj := getNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		opts := generic.Options{}
		v := generic.NewValue(desc, data)
		ps := []generic.PathNode{
			{Path: generic.NewPathFieldId(2)},
			{Path: generic.NewPathFieldId(8)},
			{Path: generic.NewPathFieldId(15)},
		}
		require.Nil(b, v.GetMany(ps, &opts))
		require.Nil(b, v.SetMany(ps, &opts))

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = v.SetMany(ps, &opts)
		}
	})
}

func BenchmarkThriftMarshalTo(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getSimpleDesc()
		part := getPartialSimpleDesc()
		obj := getSimpleValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		opts := generic.Options{}
		v := generic.NewValue(desc, data)
		out, err := v.MarshalTo(part, &opts)
		require.Nil(b, err)
		exp := baseline.NewPartialSimple()
		_, err = exp.FastRead(out)
		require.Nil(b, err)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = v.MarshalTo(part, &opts)
		}
	})

	b.Run("medium", func(b *testing.B) {
		desc := getNestingDesc()
		part := getPartialNestingDesc()
		obj := getNestingValue()
		data := make([]byte, obj.BLength())
		ret := obj.FastWriteNocopy(data, nil)
		if ret < 0 {
			b.Fatal(ret)
		}
		opts := generic.Options{}
		v := generic.NewValue(desc, data)
		out, err := v.MarshalTo(part, &opts)
		require.NoError(b, err)
		exp := baseline.NewPartialNesting()
		_, err = exp.FastRead(out)
		require.Nil(b, err)

		b.Run("not_check_requireness", func(b *testing.B) {
			opts := generic.Options{
				NotCheckRequireNess: true,
			}
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = v.MarshalTo(part, &opts)
			}
		})
		b.Run("check_requireness", func(b *testing.B) {
			opts := generic.Options{
				NotCheckRequireNess: false,
			}
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = v.MarshalTo(part, &opts)
			}
		})
	})
}
