package testdata

import (
	"context"
	"testing"

	"github.com/cloudwego/dynamicgo/internal/util_test"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
	"github.com/cloudwego/dynamicgo/proto/generic"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/baseline"
	"github.com/stretchr/testify/require"
	goprotowire "google.golang.org/protobuf/encoding/protowire"
	goproto "google.golang.org/protobuf/proto"
)

func init() {
	sobj := getPbSimpleValue()
	println("small protobuf data size: ", sobj.Size())

	psobj := getPbPartialSimpleValue()
	println("small protobuf data size: ", psobj.Size())

	nobj := getPbNestingValue()
	println("medium protobuf data size: ", nobj.Size())

	pnobj := getPbPartialNestingValue()
	println("medium protobuf data size: ", pnobj.Size())
}

func getPbServiceDescriptor() *proto.ServiceDescriptor {
	svc, err := proto.Options{}.NewDescriptorFromPath(context.Background(), util_test.MustGitPath(protoPath))
	if err != nil {
		panic(err)
	}
	return svc
}


func BenchmarkProtoSkip(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getPbSimpleDesc()
		obj := getPbSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		p := binary.NewBinaryProtol(data)
		for p.Left() > 0 {
			fieldNumber, wireType, _, err := p.ConsumeTag()
			if err != nil {
				b.Fatal(err)
			}

			if (*desc).Fields().ByNumber(fieldNumber) == nil {
				b.Fatal("field not found")
			}

			if err := p.Skip(wireType, false); err != nil {
				b.Fatal(err)
			}
		}

		require.Equal(b, len(data), p.Read)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				p := binary.NewBinaryProtol(data)
				for p.Left() > 0 {
					_, wireType, _, _ := p.ConsumeTag()
					_ = p.Skip(wireType, false)
				}
			}			
		})
	})

	b.Run("medium", func(b *testing.B) {
		desc := getPbNestingDesc()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		p := binary.NewBinaryProtol(data)
		for p.Left() > 0 {
			fieldNumber, wireType, _, err := p.ConsumeTag()
			if err != nil {
				b.Fatal(err)
			}

			if (*desc).Fields().ByNumber(fieldNumber) == nil {
				b.Fatal("field not found")
			}

			if err := p.Skip(wireType, false); err != nil {
				b.Fatal(err)
			}
		}

		require.Equal(b, len(data), p.Read)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				p := binary.NewBinaryProtol(data)
				for p.Left() > 0 {
					_, wireType, _, _ := p.ConsumeTag()
					_ = p.Skip(wireType, false)
				}
			}			
		})
	})
}


func BenchmarkProtoGetOne(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getPbSimpleDesc()
		obj := getPbSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		v := generic.NewRootValue(desc, data)
		vv, _ := v.GetByPath(generic.NewPathFieldId(6))
		require.Nil(b, vv.Check())
		bs, err := vv.Binary()
		require.Nil(b, err)
		require.Equal(b, obj.BinaryField, bs)
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = v.GetByPath(generic.NewPathFieldId(6))
			}
		})
	})

	b.Run("medium", func(b *testing.B) {
		desc := getPbNestingDesc()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		v := generic.NewRootValue(desc, data)
		vv, _ := v.GetByPath(generic.NewPathFieldId(6))
		require.Nil(b, vv.Check())
		bs, err := vv.Int()
		require.Nil(b, err)
		require.Equal(b, obj.I64, int64(bs))
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = v.GetByPath(generic.NewPathFieldId(6))
			}
		})
	})
}


func BenchmarkProtoGetMany(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getPbSimpleDesc()
		obj := getPbSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		opts := generic.Options{}

		v := generic.NewRootValue(desc, data)

		ps := []generic.PathNode{
			{Path: generic.NewPathFieldId(1)},
			{Path: generic.NewPathFieldId(3)},
			{Path: generic.NewPathFieldId(6)},
		}

		err := v.GetMany(ps, &opts)
		require.Nil(b, err)

		b.Run("go", func(b *testing.B) {
			opts.UseNativeSkip = false
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = v.GetMany(ps, &opts)
			}
		})
	})

	b.Run("medium", func(b *testing.B) {
		desc := getPbNestingDesc()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		opts := generic.Options{}

		v := generic.NewRootValue(desc, data)

		ps := []generic.PathNode{
			{Path: generic.NewPathFieldId(2)},
			{Path: generic.NewPathFieldId(8)},
			{Path: generic.NewPathFieldId(15)},
		}

		err := v.GetMany(ps, &opts)
		require.Nil(b, err)

		b.Run("go", func(b *testing.B) {
			opts.UseNativeSkip = false
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = v.GetMany(ps, &opts)
			}
		})
	})
}


func BenchmarkProtoMarshalMany(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getPbSimpleDesc()
		obj := getPbSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		opts := generic.Options{}

		v := generic.NewRootValue(desc, data)
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
		
		buf ,err := n.Marshal(&opts)
		require.Nil(b, err)
		exp := baseline.PartialSimple{}
		dataLen := len(buf)
		l := 0
		for l < dataLen {
			id, wtyp, tagLen := goprotowire.ConsumeTag(buf)
			if tagLen < 0 {
				b.Fatal("test failed")
			}
			l += tagLen
			buf = buf[tagLen:]
			offset, err := exp.FastRead(buf, int8(wtyp), int32(id))
			require.Nil(b, err)
			buf = buf[offset:]
			l += offset
		}
		if len(buf) != 0 {
			b.Fatal("test failed")
		}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i:= 0; i < b.N; i++ {
			_, _ = n.Marshal(&opts)
		}
	})

	b.Run("medium", func(b *testing.B) {
		desc := getPbNestingDesc()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		opts := generic.Options{}
		v := generic.NewRootValue(desc, data)
		ps := []generic.PathNode{
			{Path: generic.NewPathFieldId(2)},
			{Path: generic.NewPathFieldId(8)},
			{Path: generic.NewPathFieldId(15)},
		}
		err := v.GetMany(ps, &opts)
		require.Nil(b, err)
		n := generic.PathNode{
			Path: generic.NewPathFieldId(2),
			Node: v.Node,
			Next: ps,
		}
		buf ,err := n.Marshal(&opts)
		require.Nil(b, err)
		exp := baseline.PartialNesting{}
		dataLen := len(buf)
		l := 0
		for l < dataLen {
			id, wtyp, tagLen := goprotowire.ConsumeTag(buf)
			if tagLen < 0 {
				b.Fatal("test failed")
			}
			l += tagLen
			buf = buf[tagLen:]
			offset, err := exp.FastRead(buf, int8(wtyp), int32(id))
			require.Nil(b, err)
			buf = buf[offset:]
			l += offset
		}
		if len(buf) != 0 {
			b.Fatal("test failed")
		}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i:= 0; i < b.N; i++ {
			_, _ = n.Marshal(&opts)
		}
	})
}

func BenchmarkProtoMarshalTo(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getPbSimpleDesc()
		part := getPbPartialSimpleDesc()
		obj := getPbSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		opts := generic.Options{}
		v := generic.NewRootValue(desc, data)
		out, err := v.MarshalTo(part, &opts)
		require.Nil(b, err)
		// fast read check
		exp := baseline.PartialSimple{}
		dataLen := len(out)
		l := 0
		for l < dataLen {
			id, wtyp, tagLen := goprotowire.ConsumeTag(out)
			if tagLen < 0 {
				b.Fatal("test failed")
			}
			l += tagLen
			out = out[tagLen:]
			offset, err := exp.FastRead(out, int8(wtyp), int32(id))
			require.Nil(b, err)
			out = out[offset:]
			l += offset
		}
		if len(out) != 0 {
			b.Fatal("test failed")
		}

		opts = generic.Options{
			UseNativeSkip: false,
		}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()

		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = v.MarshalTo(part, &opts)
			}
		})
	})

	b.Run("medium", func(b *testing.B) {
		desc := getPbNestingDesc()
		part := getPbPartialNestingDesc()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		opts := generic.Options{}
		v := generic.NewRootValue(desc, data)
		out, err := v.MarshalTo(part, &opts)
		require.Nil(b, err)
		// fast read check
		exp := baseline.PartialSimple{}
		dataLen := len(out)
		l := 0
		for l < dataLen {
			id, wtyp, tagLen := goprotowire.ConsumeTag(out)
			if tagLen < 0 {
				b.Fatal("test failed")
			}
			l += tagLen
			out = out[tagLen:]
			offset, err := exp.FastRead(out, int8(wtyp), int32(id))
			require.Nil(b, err)
			out = out[offset:]
			l += offset
		}
		if len(out) != 0 {
			b.Fatal("test failed")
		}

		opts = generic.Options{
			UseNativeSkip: false,
		}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()

		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = v.MarshalTo(part, &opts)
			}
		})
	})
}

func BenchmarkProtoSetOne(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getPbSimpleDesc()
		obj := getPbSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		v := generic.NewRootValue(desc, data)
		p := binary.NewBinaryProtocolBuffer()
		p.WriteBytes(obj.BinaryField)
		fd6 := (*desc).Fields().ByNumber(6)
		n := generic.NewValue(&fd6, p.Buf)
		_, err := v.SetByPath(n, generic.NewPathFieldId(6))
		require.Nil(b, err)
		nn, _ := v.GetByPath(generic.NewPathFieldId(6))
		require.Equal(b, n.Raw(), nn.Raw())

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = v.SetByPath(n, generic.NewPathFieldId(6))
			}
		})
	})

	b.Run("medium", func(b *testing.B) {
		desc := getPbNestingDesc()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		v := generic.NewRootValue(desc, data)
		p := binary.NewBinaryProtocolBuffer()
		p.WriteBytes(obj.MapStringSimple["0"].BinaryField)
		fd15 := (*desc).Fields().ByNumber(15).MapValue().Message().Fields().ByNumber(6)
		n := generic.NewValue(&fd15, p.Buf)
		_, err := v.SetByPath(n, generic.NewPathFieldId(15), generic.NewPathStrKey("0"), generic.NewPathFieldId(6))
		require.Nil(b, err)
		nn, _ := v.GetByPath(generic.NewPathFieldId(15), generic.NewPathStrKey("0"), generic.NewPathFieldId(6))
		require.Equal(b, n.Raw(), nn.Raw())

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = v.SetByPath(n, generic.NewPathFieldId(15), generic.NewPathStrKey("0"), generic.NewPathFieldId(6))
			}
		})
	})
}

func BenchmarkProtoMarshalAll_ProtoBufGo(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		obj := getPbSimpleValue()
		data, err := goproto.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = goproto.Marshal(obj)
		}
	})

	b.Run("medium", func(b *testing.B) {
		obj := getPbNestingValue()
		data, err := goproto.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
		
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = goproto.Marshal(obj)
		}
	})
}

func BenchmarkProtoMarshalPartial_ProtoBufGo(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		obj := getPbPartialSimpleValue()
		data, err := goproto.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = goproto.Marshal(obj)
		}
	})

	b.Run("medium", func(b *testing.B) {
		obj := getPbPartialNestingValue()
		data, err := goproto.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
		
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = goproto.Marshal(obj)
		}
	})
}

func BenchmarkProtoUnmarshalAll_ProtoBufGo(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		obj := getPbSimpleValue()
		data, err := goproto.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
		v := &baseline.Simple{}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = goproto.Unmarshal(data, v)
		}
	})

	b.Run("medium", func(b *testing.B) {
		obj := getPbNestingValue()
		data, err := goproto.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
		v := &baseline.Nesting{}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = goproto.Unmarshal(data, v)
		}
	})
}

func BenchmarkProtoUnmarshalPartial_ProtoBufGo(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		obj := getPbPartialSimpleValue()
		data, err := goproto.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
		v := &baseline.PartialSimple{}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = goproto.Unmarshal(data, v)
		}
	})

	b.Run("medium", func(b *testing.B) {
		obj := getPbPartialNestingValue()
		data, err := goproto.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
		v := &baseline.PartialNesting{}
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = goproto.Unmarshal(data, v)
		}
	})
}


func BenchmarkProtoMarshallAll_KitexFast(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		obj := getPbSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		b.SetBytes(int64(len(data)))
		result := make([]byte, obj.Size())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = obj.FastWrite(result)
		}
	})

	b.Run("medium", func(b *testing.B) {
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		b.SetBytes(int64(len(data)))
		result := make([]byte, obj.Size())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = obj.FastWrite(result)
		}
	})
}

func BenchmarkProtoMarshallPartial_KitexFast(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		obj := getPbPartialSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		b.SetBytes(int64(len(data)))
		result := make([]byte, obj.Size())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = obj.FastWrite(result)
		}
	})

	b.Run("medium", func(b *testing.B) {
		obj := getPbPartialNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		b.SetBytes(int64(len(data)))
		result := make([]byte, obj.Size())
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = obj.FastWrite(result)
		}
	})
}

func BenchmarkProtoUnmarshalAll_KitexFast(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		obj := getPbSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		b.SetBytes(int64(len(data)))
		exp := &baseline.Simple{}
		dataLen := len(data)
		l := 0
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for l < dataLen {
				id, wtyp, tagLen := goprotowire.ConsumeTag(data)
				data = data[tagLen:]
				l += tagLen
				offset, _ := exp.FastRead(data, int8(wtyp), int32(id))
				data = data[offset:]
				l += offset
			}
		}
	})

	b.Run("medium", func(b *testing.B) {
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		b.SetBytes(int64(len(data)))
		exp := &baseline.Nesting{}
		dataLen := len(data)
		l := 0
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for l < dataLen {
				id, wtyp, tagLen := goprotowire.ConsumeTag(data)
				data = data[tagLen:]
				l += tagLen
				offset, _ := exp.FastRead(data, int8(wtyp), int32(id))
				data = data[offset:]
				l += offset
			}
		}
	})
}

func BenchmarkProtoUnmarshalPartial_KitexFast(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		obj := getPbPartialSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		b.SetBytes(int64(len(data)))
		exp := &baseline.PartialSimple{}
		dataLen := len(data)
		l := 0
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for l < dataLen {
				id, wtyp, tagLen := goprotowire.ConsumeTag(data)
				data = data[tagLen:]
				l += tagLen
				offset, _ := exp.FastRead(data, int8(wtyp), int32(id))
				data = data[offset:]
				l += offset
			}
		}
	})

	b.Run("medium", func(b *testing.B) {
		obj := getPbPartialNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		b.SetBytes(int64(len(data)))
		exp := &baseline.PartialNesting{}
		dataLen := len(data)
		l := 0
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for l < dataLen {
				id, wtyp, tagLen := goprotowire.ConsumeTag(data)
				data = data[tagLen:]
				l += tagLen
				offset, _ := exp.FastRead(data, int8(wtyp), int32(id))
				data = data[offset:]
				l += offset
			}
		}
	})
}

func BenchmarkProtoMarshalTo_KitexFast(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		obj := getPbSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		pobj := getPbPartialSimpleValue()
		plen := pobj.FastWrite(data)
		pdata := make([]byte, pobj.Size())
		ret = pobj.FastWrite(pdata)
		if ret != len(pdata) {
			b.Fatal(ret)
		}
		require.Equal(b, plen, pobj.Size())
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = obj.FastWrite(data)
			_ = pobj.FastWrite(pdata)
		}
	})

	b.Run("medium", func(b *testing.B) {
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		pobj := getPbPartialNestingValue()
		plen := pobj.FastWrite(data)
		pdata := make([]byte, pobj.Size())
		ret = pobj.FastWrite(pdata)
		if ret != len(pdata) {
			b.Fatal(ret)
		}
		require.Equal(b, plen, pobj.Size())
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = obj.FastWrite(data)
			_ = pobj.FastWrite(pdata)
		}
	})
}


func BenchmarkProtoMarshallAll_DynamicGo(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getPbSimpleDesc()
		obj := getPbSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		v := generic.NewRootValue(desc, data)
		p := generic.PathNode{
			Node: v.Node,
		}
		opts := &generic.Options{}
		require.Nil(b, p.Load(true, opts, desc))
		out, err := p.Marshal(opts)
		require.Nil(b, err)
		require.Equal(b, data, out)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = p.Marshal(opts)
		}
	})

	b.Run("medium", func(b *testing.B) {
		desc := getPbNestingDesc()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		v := generic.NewRootValue(desc, data)
		p := generic.PathNode{
			Node: v.Node,
		}
		opts := &generic.Options{}
		require.Nil(b, p.Load(true, opts, desc))
		out, err := p.Marshal(opts)
		require.Nil(b, err)
		require.Equal(b, data, out)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = p.Marshal(opts)
		}
	})
}

func BenchmarkProtoMarshallPartial_DynamicGo(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getPbPartialSimpleDesc()
		obj := getPbPartialSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		v := generic.NewRootValue(desc, data)
		p := generic.PathNode{
			Node: v.Node,
		}
		opts := &generic.Options{}
		require.Nil(b, p.Load(true, opts, desc))
		out, err := p.Marshal(opts)
		require.Nil(b, err)
		require.Equal(b, data, out)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = p.Marshal(opts)
		}
	})

	b.Run("medium", func(b *testing.B) {
		desc := getPbPartialNestingDesc()
		obj := getPbPartialNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		v := generic.NewRootValue(desc, data)
		p := generic.PathNode{
			Node: v.Node,
		}
		opts := &generic.Options{}
		require.Nil(b, p.Load(true, opts, desc))
		out, err := p.Marshal(opts)
		require.Nil(b, err)
		require.Equal(b, data, out)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = p.Marshal(opts)
		}
	})
}

func BenchmarkProtoGetAll_New(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getPbSimpleDesc()
		obj := getPbSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		v := generic.NewRootValue(desc, data)
		out := []generic.PathNode{}
		require.Nil(b, v.Children(&out, false, &generic.Options{UseNativeSkip: false}, desc))

		b.Run("go", func(b *testing.B) {
			opts := &generic.Options{
				UseNativeSkip: false,
				// OnlyScanStruct: true,
			}
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				out = out[:0]
				_ = v.Children(&out, true, opts, desc)
			}
		})
	})

	b.Run("medium", func(b *testing.B) {
		desc := getPbNestingDesc()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		v := generic.NewRootValue(desc, data)
		out := []generic.PathNode{}
		require.Nil(b, v.Children(&out, false, &generic.Options{UseNativeSkip: false}, desc))

		b.Run("go", func(b *testing.B) {
			opts := &generic.Options{
				UseNativeSkip: false,
			}
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				out := []generic.PathNode{}
				_ = v.Children(&out,true, opts, desc)
			}
		})
	})
}

func BenchmarkProtoGetPartial_New(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getPbPartialSimpleDesc()
		obj := getPbPartialSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret, len(data))
		}

		v := generic.NewRootValue(desc, data)
		out := []generic.PathNode{}
		require.Nil(b, v.Children(&out, false, &generic.Options{UseNativeSkip: false}, desc))

		b.Run("go", func(b *testing.B) {
			opts := &generic.Options{
				UseNativeSkip: false,
				// OnlyScanStruct: true,
			}
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				out = out[:0]
				_ = v.Children(&out, true, opts, desc)
			}
		})
	})

	b.Run("medium", func(b *testing.B) {
		desc := getPbPartialNestingDesc()
		obj := getPbPartialNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret, len(data))
		}

		v := generic.NewRootValue(desc, data)
		out := []generic.PathNode{}
		require.Nil(b, v.Children(&out, false, &generic.Options{UseNativeSkip: false}, desc))

		b.Run("go", func(b *testing.B) {
			opts := &generic.Options{
				UseNativeSkip: false,
			}
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				out := []generic.PathNode{}
				_ = v.Children(&out,true, opts, desc)
			}
		})
	})

}

func BenchmarkProtoGetAll_ReuseMemory(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getPbSimpleDesc()
		obj := getPbSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret, len(data))
		}
		v := generic.NewRootValue(desc, data)
		r := generic.NewPathNode()
		r.Node = v.Node
		r.Load(true, &generic.Options{UseNativeSkip: false}, desc)
		r.ResetAll()
		generic.FreePathNode(r)

		b.Run("go", func(b *testing.B) {
			opts := &generic.Options{
				UseNativeSkip: false,
			}
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r := generic.NewPathNode()
				r.Node = v.Node
				r.Load(true, opts, desc)
				r.ResetAll()
				generic.FreePathNode(r)
			}
		})

	})

	b.Run("medium", func(b *testing.B) {
		desc := getPbNestingDesc()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret, len(data))
		}
		v := generic.NewRootValue(desc, data)
		r := generic.NewPathNode()
		r.Node = v.Node
		r.Load(true, &generic.Options{UseNativeSkip: false}, desc)
		r.ResetAll()
		generic.FreePathNode(r)

		b.Run("go", func(b *testing.B) {
			opts := &generic.Options{
				UseNativeSkip: false,
			}
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r := generic.NewPathNode()
				r.Node = v.Node
				r.Load(true, opts, desc)
				r.ResetAll()
				generic.FreePathNode(r)
			}
		})
	})
}


func BenchmarkProtoGetPartial_ReuseMemory(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getPbPartialSimpleDesc()
		obj := getPbPartialSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret, len(data))
		}
		v := generic.NewRootValue(desc, data)
		r := generic.NewPathNode()
		r.Node = v.Node
		r.Load(true, &generic.Options{UseNativeSkip: false}, desc)
		r.ResetAll()
		generic.FreePathNode(r)

		b.Run("go", func(b *testing.B) {
			opts := &generic.Options{
				UseNativeSkip: false,
			}
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r := generic.NewPathNode()
				r.Node = v.Node
				r.Load(true, opts, desc)
				r.ResetAll()
				generic.FreePathNode(r)
			}
		})

	})

	b.Run("medium", func(b *testing.B) {
		desc := getPbPartialNestingDesc()
		obj := getPbPartialNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret, len(data))
		}
		v := generic.NewRootValue(desc, data)
		r := generic.NewPathNode()
		r.Node = v.Node
		r.Load(true, &generic.Options{UseNativeSkip: false}, desc)
		r.ResetAll()
		generic.FreePathNode(r)

		b.Run("go", func(b *testing.B) {
			opts := &generic.Options{
				UseNativeSkip: false,
			}
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r := generic.NewPathNode()
				r.Node = v.Node
				r.Load(true, opts, desc)
				r.ResetAll()
				generic.FreePathNode(r)
			}
		})
	})
}

