package testdata

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"testing"

	"github.com/cloudwego/dynamicgo/internal/util_test"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
	"github.com/cloudwego/dynamicgo/proto/generic"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/baseline"
	"github.com/stretchr/testify/require"
	goprotowire "google.golang.org/protobuf/encoding/protowire"
	goproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
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
		vv := v.GetByPath(generic.NewPathFieldId(6))
		require.Nil(b, vv.Check())
		bs, err := vv.Binary()
		require.Nil(b, err)
		require.Equal(b, obj.BinaryField, bs)
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = v.GetByPath(generic.NewPathFieldId(6))
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
		vv := v.GetByPath(generic.NewPathFieldId(6))
		require.Nil(b, vv.Check())
		bs, err := vv.Int()
		require.Nil(b, err)
		require.Equal(b, obj.I64, int64(bs))
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = v.GetByPath(generic.NewPathFieldId(6))
			}
		})
	})
}

func BenchmarkDynamicPbGetOne(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		// build desc、obj
		desc := getPbSimpleDesc()
		obj := getPbSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		// build dynamicpb Message
		message := dynamicpb.NewMessage(*desc)
		if err := goproto.Unmarshal(data, message); err != nil {
			b.Fatal("build dynamicpb failed")
		}
		targetDesc := (*desc).Fields().ByNumber(6)
		if !message.Has(targetDesc) {
			b.Fatal("dynamicpb can't find targetDesc")
		}
		value := message.Get(targetDesc)
		require.Equal(b, obj.BinaryField, value.Bytes())
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = goproto.Unmarshal(data, message)
				_ = message.Get(targetDesc)
				// _ = goproto.Marshal(message)
			}
		})
	})

	b.Run("medium", func(b *testing.B) {
		// build desc、obj
		desc := getPbNestingDesc()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		// build dynamicpb Message
		message := dynamicpb.NewMessage(*desc)
		if err := goproto.Unmarshal(data, message); err != nil {
			b.Fatal("build dynamicpb failed")
		}
		targetDesc := (*desc).Fields().ByNumber(6)
		if !message.Has(targetDesc) {
			b.Fatal("dynamicpb can't find targetDesc")
		}
		value := message.Get(targetDesc)
		require.Equal(b, obj.I64, value.Int())
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = goproto.Unmarshal(data, message)
				_ = message.Get(targetDesc)
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

		opts.UseNativeSkip = false
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
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

		opts.UseNativeSkip = false
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = v.GetMany(ps, &opts)
			}
		})
	})
}

func BenchmarkProtoSetMany(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getPbSimpleDesc()
		obj := getPbSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			panic(ret)
		}

		opts := generic.Options{}
		v := generic.NewRootValue(desc, data)
		ps := []generic.PathNode{
			{Path: generic.NewPathFieldId(1)},
			{Path: generic.NewPathFieldId(3)},
			{Path: generic.NewPathFieldId(6)},
		}

		adress2root := make([]int, 0)
		path2root := make([]generic.Path, 0)
		require.Nil(b, v.GetMany(ps, &opts))
		require.Nil(b, v.SetMany(ps, &opts, &v, adress2root, path2root...))
		opts.UseNativeSkip = false
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = v.SetMany(ps, &opts, &v, adress2root, path2root...)
			}
		})
	})

	b.Run("medium", func(b *testing.B) {
		desc := getPbNestingDesc()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			panic(ret)
		}

		opts := generic.Options{}
		v := generic.NewRootValue(desc, data)
		ps := []generic.PathNode{
			{Path: generic.NewPathFieldId(2)},
			{Path: generic.NewPathFieldId(8)},
			{Path: generic.NewPathFieldId(15)},
		}

		adress2root := make([]int, 0)
		path2root := make([]generic.Path, 0)
		require.Nil(b, v.GetMany(ps, &opts))
		require.Nil(b, v.SetMany(ps, &opts, &v, adress2root, path2root...))
		opts.UseNativeSkip = false
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = v.SetMany(ps, &opts, &v, adress2root, path2root...)
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

		buf, err := n.Marshal(&opts)
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
		for i := 0; i < b.N; i++ {
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
		buf, err := n.Marshal(&opts)
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
		for i := 0; i < b.N; i++ {
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
		_, err := v.SetByPath(n.Node, generic.NewPathFieldId(6))
		require.Nil(b, err)
		nn := v.GetByPath(generic.NewPathFieldId(6))
		require.Equal(b, n.Raw(), nn.Raw())

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = v.SetByPath(n.Node, generic.NewPathFieldId(6))
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
		_, err := v.SetByPath(n.Node, generic.NewPathFieldId(15), generic.NewPathStrKey("0"), generic.NewPathFieldId(6))
		require.Nil(b, err)
		nn := v.GetByPath(generic.NewPathFieldId(15), generic.NewPathStrKey("0"), generic.NewPathFieldId(6))
		require.Equal(b, n.Raw(), nn.Raw())

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = v.SetByPath(n.Node, generic.NewPathFieldId(15), generic.NewPathStrKey("0"), generic.NewPathFieldId(6))
			}
		})
	})
}

func BenchmarkDynamicpbSetOne(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getPbSimpleDesc()
		obj := getPbSimpleValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		message := dynamicpb.NewMessage(*desc)
		targetDesc := (*desc).Fields().ByNumber(6)
		if err := goproto.Unmarshal(data, message); err != nil {
			b.Fatal("build dynamicpb failed")
		}
		fieldValue := protoreflect.ValueOfBytes(obj.BinaryField)
		message.Set(targetDesc, fieldValue)
		if !message.Has(targetDesc) {
			b.Fatal("dynamicpb can't find targetDesc")
		}
		find := message.Get(targetDesc)
		require.Equal(b, obj.BinaryField, find.Bytes())
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				message = dynamicpb.NewMessage(*desc)
				_ = goproto.Unmarshal(data, message)
				fieldValue := protoreflect.ValueOfBytes(obj.BinaryField)
				message.Set(targetDesc, fieldValue)
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

func BenchmarkProtoMarshalTo_ProtoBufGo(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		obj := getPbSimpleValue()
		data, err := goproto.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
		part := &baseline.PartialSimple{}
		opts := goproto.UnmarshalOptions{DiscardUnknown: true}
		err = opts.Unmarshal(data, part)
		if err != nil {
			b.Fatal(err)
		}

		out, err2 := goproto.Marshal(part)
		if err2 != nil {
			b.Fatal(err2)
		}
		
		desc := getPbSimpleDesc()
		partdesc := getPbPartialSimpleDesc()
		v := generic.NewRootValue(desc, data)
		out2, err3 := v.MarshalTo(partdesc, &generic.Options{})
		if err3 != nil {
			b.Fatal(err3)
		}

		require.Equal(b, len(out), len(out2))

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			part := &baseline.PartialSimple{}
			_ = opts.Unmarshal(data, part)
			_, _ = goproto.Marshal(part)
		}
	})

	b.Run("medium", func(b *testing.B) {
		obj := getPbNestingValue()
		data, err := goproto.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
		part := &baseline.PartialNesting{}
		opts := goproto.UnmarshalOptions{DiscardUnknown: true}
		err = opts.Unmarshal(data, part)
		if err != nil {
			b.Fatal(err)
		}

		out, err2 := goproto.Marshal(part)
		if err2 != nil {
			b.Fatal(err2)
		}

		desc := getPbNestingDesc()
		partdesc := getPbPartialNestingDesc()
		v := generic.NewRootValue(desc, data)
		out2, err3 := v.MarshalTo(partdesc, &generic.Options{})
		if err3 != nil {
			b.Fatal(err3)
		}

		require.Equal(b, len(out), len(out2))

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			part := &baseline.PartialNesting{}
			_ = opts.Unmarshal(data, part)
			_, _ = goproto.Marshal(part)
		}
	})
}

func BenchmarkProtoMarshalAll_KitexFast(b *testing.B) {
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

func BenchmarkProtoMarshalPartial_KitexFast(b *testing.B) {
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

		// fast read check
		exp := baseline.PartialSimple{}
		dataLen := len(data)
		l := 0
		for l < dataLen {
			id, wtyp, tagLen := goprotowire.ConsumeTag(data)
			if tagLen < 0 {
				b.Fatal("test failed")
			}
			l += tagLen
			data = data[tagLen:]
			offset, err := exp.FastRead(data, int8(wtyp), int32(id))
			require.Nil(b, err)
			data = data[offset:]
			l += offset
		}

		data2 := make([]byte, exp.Size())
		ret2 := exp.FastWrite(data2)
		if ret2 != len(data2) {
			b.Fatal(ret2)
		}
			
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			exp := baseline.PartialSimple{}
			dataLen := len(data)
			l := 0
			for l < dataLen {
				id, wtyp, tagLen := goprotowire.ConsumeTag(data)
				if tagLen < 0 {
					b.Fatal("test failed")
				}
				l += tagLen
				data = data[tagLen:]
				offset, _ := exp.FastRead(data, int8(wtyp), int32(id))
				data = data[offset:]
				l += offset
			}
			data2 := make([]byte, exp.Size())
			_ = exp.FastWrite(data2)
			
		}
	})

	b.Run("medium", func(b *testing.B) {
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			panic(ret)
		}

		// fast read check
		exp := baseline.PartialNesting{}
		dataLen := len(data)
		l := 0
		for l < dataLen {
			id, wtyp, tagLen := goprotowire.ConsumeTag(data)
			if tagLen < 0 {
				b.Fatal("test failed")
			}
			l += tagLen
			data = data[tagLen:]
			offset, err := exp.FastRead(data, int8(wtyp), int32(id))
			require.Nil(b, err)
			data = data[offset:]
			l += offset
		}

		data2 := make([]byte, exp.Size())
		ret2 := exp.FastWrite(data2)
		if ret2 != len(data2) {
			b.Fatal(ret2)
		}
			
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			exp := baseline.PartialNesting{}
			dataLen := len(data)
			l := 0
			for l < dataLen {
				id, wtyp, tagLen := goprotowire.ConsumeTag(data)
				if tagLen < 0 {
					b.Fatal("test failed")
				}
				l += tagLen
				data = data[tagLen:]
				offset, _ := exp.FastRead(data, int8(wtyp), int32(id))
				data = data[offset:]
				l += offset
			}
			data2 := make([]byte, exp.Size())
			_ = exp.FastWrite(data2)
		}
	})
}

func BenchmarkProtoMarshalAll_DynamicGo(b *testing.B) {
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

func BenchmarkProtoMarshalPartial_DynamicGo(b *testing.B) {
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

func BenchmarkProtoUnmarshalAllDynamicGoGet_New(b *testing.B) {
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
				out := []generic.PathNode{}
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
				_ = v.Children(&out, true, opts, desc)
			}
		})
	})
}

func BenchmarkProtoUnmarshalPartialDynamicGoGet_New(b *testing.B) {
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
				out := []generic.PathNode{}
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
				_ = v.Children(&out, true, opts, desc)
			}
		})
	})

}

func BenchmarkProtoUnmarshalAllDynamicGoGet_ReuseMemory(b *testing.B) {
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

func BenchmarkProtoUnmarshalPartialDynamicGoGet_ReuseMemory(b *testing.B) {
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

const (
	factor         = 1.0 // change set/get field ratio
	defaultBufSize = 512
)

func sizeNestingField(value reflect.Value, id int) int {
	name := fmt.Sprintf("%s%s", "SizeField", strconv.FormatInt(int64(id), 10))
	method := value.MethodByName(name)
	argument := []reflect.Value{}
	ans := method.Call(argument)
	return int(ans[0].Int())
}

func encNestingField(value reflect.Value, id int, b []byte) error {
	name := fmt.Sprintf("%s%s", "FastWriteField", strconv.FormatInt(int64(id), 10))
	method := value.MethodByName(name)
	argument := []reflect.Value{reflect.ValueOf(b)}
	method.Call(argument)
	return nil
}

func parseBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(key); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func collectMarshalData(id int, value reflect.Value, b *[]byte) error {
	size := sizeNestingField(value, id)
	data := make([]byte, size)
	if err := encNestingField(value, id, data); err != nil {
		return err
	}
	*b = append(*b, data...)
	return nil
}

func buildBinaryProtocolByFieldId(id int, p *binary.BinaryProtocol, value reflect.Value, desc *proto.FieldDescriptor) error {
	var err error
	size := sizeNestingField(value, id)
	data := make([]byte, size)
	if err = encNestingField(value, id, data); err != nil {
		return err
	}
	if !(*desc).IsList() && !(*desc).IsMap() {
		data = data[1:] // skip Tag because the test case tag len is 1, we just use 1, it is not always 1, please use comsumevarint
	}
	p.Buf = data
	return nil
}

func BenchmarkProtoRationGet(b *testing.B) {
	b.Run("ration", func(b *testing.B) {
		desc := getPbNestingDesc()
		fieldNums := (*desc).Fields().Len()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		v := generic.NewRootValue(desc, data)
		testNums := int(math.Ceil(float64(fieldNums) * factor))
		value := reflect.ValueOf(obj)
		for id := 1; id <= testNums; id++ {
			vv := v.GetByPath(generic.NewPathFieldId(proto.Number(id)))
			require.Nil(b, vv.Check())
			size := sizeNestingField(value, id)
			data := make([]byte, size)
			if err := encNestingField(value, id, data); err != nil {
				b.Fatal("encNestingField failed, fieldId: {}", id)
			}
			// skip Tag
			t := vv.Node.Type()
			if t != proto.LIST && t != proto.MAP {
				data = data[1:] // skip Tag because the test case tag len is 1, we just use 1, it is not always 1, please use comsumevarint
			}
			vdata := vv.Raw()
			for j := 0; j < size; j++ {
				if !bytes.Equal(vdata[:j], data[:j]) {
					fmt.Print(j)
				}
			}
			if t != proto.MAP {
				require.Equal(b, data, vdata)
			} else {
				require.Equal(b, len(data), len(vdata))
			}
		}
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				option := generic.Options{}
				for id := 1; id <= testNums; id++ {
					vv := v.GetByPath(generic.NewPathFieldId(proto.Number(id)))
					vv.Interface(&option)
				}
			}
		})
	})
}

func BenchmarkDynamicpbRationGet(b *testing.B) {
	b.Run("ration", func(b *testing.B) {
		desc := getPbNestingDesc()
		fieldNums := (*desc).Fields().Len()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		testNums := int(math.Ceil(float64(fieldNums) * factor))
		// build dynamicpb Message
		message := dynamicpb.NewMessage(*desc)
		if err := goproto.Unmarshal(data, message); err != nil {
			b.Fatal("build dynamicpb failed")
		}
		value := reflect.ValueOf(obj)
		for id := 1; id <= testNums; id++ {
			// dynamicpb read data
			targetDesc := (*desc).Fields().ByNumber(proto.Number(id))
			if !message.Has(targetDesc) {
				b.Fatal("dynamicpb can't find targetDesc")
			}
			_ = message.Get(targetDesc)
			// fastRead data
			size := sizeNestingField(value, id)
			data := make([]byte, size)
			if err := encNestingField(value, id, data); err != nil {
				b.Fatal("encNestingField failed, fieldId: {}", id)
			}
		}
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = goproto.Unmarshal(data, message)
				for id := 1; id <= testNums; id++ {
					targetDesc := (*desc).Fields().ByNumber(proto.Number(id))
					v := message.Get(targetDesc)
					v.Interface()
				}
			}
		})
	})
}



// func GetBytes(p *binary.BinaryProtocol, obj *baseline.Nesting, id int, desc *proto.FieldDescriptor) (err error) {
// 	switch id {
// 	case 1:
// 		err = p.WriteString(obj.String_)
// 	case 2:
// 		err = p.WriteListFast(desc, obj.ListSimple)
// 	case 3:
// 		err = p.WriteDouble(obj.Double)
// 	case 4:
// 		err = p.WriteI32(obj.I32)
// 	case 5:
// 		err = p.WriteListFast(desc, obj.ListI32)
// 	case 6:
// 		err = p.WriteI64(obj.I64)
// 	case 7:
// 		err = p.WriteMap(desc, obj.MapStringString)
// 	case 8:
// 		err = p.WriteMessageSlow(desc, obj.SimpleStruct, false, false, false)
// 	case 9:
// 		err = p.WriteMapFast(desc, obj.MapI32I64)
// 	case 10:
// 		err = p.WriteListFast(desc, obj.ListString)
// 	case 11:
// 		err = p.WriteBytes(obj.Binary)
// 	case 12:
// 		err = p.WriteMapFast(desc, obj.MapI64String)
// 	case 13:
// 		err = p.WriteListFast(desc, obj.ListI64)
// 	case 14:
// 		err = p.WriteBytes(obj.Byte)
// 	case 15:
// 		err = p.WriteMapFast(desc, obj.MapStringSimple)
// 	}
// 	return err
// }

func BenchmarkProtoRationSetBefore(b *testing.B) {
	b.Run("ration", func(b *testing.B) {
		desc := getPbNestingDesc()
		fieldNums := (*desc).Fields().Len()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		v := generic.NewRootValue(desc, nil)
		opts := generic.Options{}
		ps := make([]generic.PathNode, 0)
		testNums := int(math.Ceil(float64(fieldNums) * factor))
		mObj := make([]byte, 0)
		value := reflect.ValueOf(obj)
		for id := 1; id <= testNums; id++ {
			p := binary.NewBinaryProtocolBuffer()
			fieldDesc := (*desc).Fields().ByNumber(proto.Number(id))
			if err := collectMarshalData(id, value, &mObj); err != nil {
				b.Fatal("collect MarshalData failed")
			}
			if err := buildBinaryProtocolByFieldId(id, p, value, &fieldDesc); err != nil {
				b.Fatal("build BinaryProtocolByFieldId failed")
			}
			field := (*desc).Fields().ByNumber(proto.Number(id))
			n := generic.NewValue(&field, p.Buf)
			_, err := v.SetByPath(n.Node, generic.NewPathFieldId(proto.Number(id)))
			require.Nil(b, err)
			nn := v.GetByPath(generic.NewPathFieldId(proto.Number(id)))
			nndata := nn.Raw()
			ndata := n.Raw()
			if nn.Type() != proto.MAP {
				require.Equal(b, nndata, ndata)
			} else {
				require.Equal(b, len(nndata), len(ndata))
			}
			pnode := generic.PathNode{
				Path: generic.NewPathFieldId(proto.FieldNumber(id)),
				Node: n.Node,
			}
			ps = append(ps, pnode)
			p.Recycle()
		}
		n := generic.PathNode{
			Path: generic.NewPathFieldId(1),
			Node: v.Node,
			Next: ps,
		}
		mProto, err := n.Marshal(&opts)
		if err != nil {
			b.Fatal("marshal PathNode failed")
		}
		require.Equal(b, len(mProto), len(mObj))
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				opts := generic.Options{}
				v := generic.NewRootValue(desc, nil)
				ps := make([]generic.PathNode, 0)
				for id := 1; id <= testNums; id++ {
					p := binary.NewBinaryProtocolBuffer()
					fieldDesc := (*desc).Fields().ByNumber(proto.Number(id))
					_ = buildBinaryProtocolByFieldId(id, p, value, &fieldDesc)
					field := (*desc).Fields().ByNumber(proto.Number(id))
					n := generic.NewValue(&field, p.Buf)
					_, _ = v.SetByPath(n.Node, generic.NewPathFieldId(proto.Number(id)))
					pnode := generic.PathNode{
						Path: generic.NewPathFieldId(proto.FieldNumber(id)),
						Node: n.Node,
					}
					ps = append(ps, pnode)
					p.Recycle()
				}
				n := generic.PathNode{
					Path: generic.NewPathFieldId(1),
					Node: v.Node,
					Next: ps,
				}
				_, _ = n.Marshal(&opts)
			}
		})
	})
}


func BenchmarkProtoRationSet(b *testing.B) {
	b.Run("ration", func(b *testing.B) {
		desc := getPbNestingDesc()
		fieldNums := (*desc).Fields().Len()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		testNums := int(math.Ceil(float64(fieldNums) * factor))
		objRoot := generic.NewRootValue(desc, data)
		newRoot := generic.NewRootValue(desc, nil)
		ps := make([]generic.PathNode, 0)
		p := binary.NewBinaryProtocolBuffer()
		mObj := make([]byte, 0)
		value := reflect.ValueOf(obj)
		for id := 1; id <= testNums; id++ {
			if err := collectMarshalData(id, value, &mObj); err != nil {
				b.Fatal("collect MarshalData failed")
			}
			x := objRoot.GetByPath(generic.NewPathFieldId(proto.Number(id)))
			require.Nil(b, x.Check())
			m, err := x.Interface(&generic.Options{MapStructById: true})
			require.Nil(b, err)
			field := (*desc).Fields().ByNumber(proto.Number(id))
			if field.IsMap() {
				err = p.WriteMap(&field, m, true, false, false)
			} else if field.IsList() {
				err = p.WriteList(&field, m, true, false, false)
			} else {
				// bacause basic type node buf no tag, just LV
				err = p.WriteBaseTypeWithDesc(&field, m, true, false, false)
			}
			// fmt.Println(id)
			require.Nil(b, err)
			newValue := generic.NewValue(&field, p.Buf)
			if field.IsMap() || field.IsList() {
				newValue = generic.NewComplexValue(&field, p.Buf)
			}
			// fmt.Println(id)
			require.Equal(b, len(newValue.Raw()), len(x.Raw()))
			_, err = newRoot.SetByPath(newValue.Node, generic.NewPathFieldId(proto.Number(id)))
			require.Nil(b, err)
			vv := newRoot.GetByPath(generic.NewPathFieldId(proto.Number(id)))
			require.Equal(b, len(newValue.Raw()), len(vv.Raw()))
			p.Recycle()
			ps = append(ps, generic.PathNode{ Path: generic.NewPathFieldId(proto.FieldNumber(id)), Node: newValue.Node })
		}

		n := generic.PathNode{
			Path: generic.NewPathFieldId(1),
			Node: newRoot.Node,
			Next: ps,
		}
		mProto, err := n.Marshal(&generic.Options{})
		if err != nil {
			b.Fatal("marshal PathNode failed")
		}
		require.Equal(b, len(mProto), len(mObj))

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				objRoot := generic.NewRootValue(desc, data)
				newRoot := generic.NewRootValue(desc, nil)
				ps := make([]generic.PathNode, 0)
				p := binary.NewBinaryProtocolBuffer()
				opt := &generic.Options{MapStructById: true}
				for id := 1; id <= testNums; id++ {
					x := objRoot.GetByPath(generic.NewPathFieldId(proto.Number(id)))
					newValue := x.Fork()
					_, err = newRoot.SetByPath(newValue.Node, generic.NewPathFieldId(proto.Number(id)))
					p.Recycle()
					ps = append(ps, generic.PathNode{Path: generic.NewPathFieldId(proto.FieldNumber(id))})
				}
				_ = newRoot.GetMany(ps,opt)
				n := generic.PathNode{
					Path: generic.NewPathFieldId(1),
					Node: newRoot.Node,
					Next: ps,
				}

				_, _ = n.Marshal(opt)
			}
		})

	})
}

func BenchmarkProtoRationSetByInterface(b *testing.B) {
	b.Run("ration", func(b *testing.B) {
		desc := getPbNestingDesc()
		fieldNums := (*desc).Fields().Len()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		testNums := int(math.Ceil(float64(fieldNums) * factor))
		objRoot := generic.NewRootValue(desc, data)
		newRoot := generic.NewRootValue(desc, nil)
		ps := make([]generic.PathNode, 0)
		p := binary.NewBinaryProtocolBuffer()
		mObj := make([]byte, 0)
		value := reflect.ValueOf(obj)
		for id := 1; id <= testNums; id++ {
			if err := collectMarshalData(id, value, &mObj); err != nil {
				b.Fatal("collect MarshalData failed")
			}
			x := objRoot.GetByPath(generic.NewPathFieldId(proto.Number(id)))
			require.Nil(b, x.Check())
			m, err := x.Interface(&generic.Options{MapStructById: true})
			require.Nil(b, err)
			field := (*desc).Fields().ByNumber(proto.Number(id))
			if field.IsMap() {
				err = p.WriteMap(&field, m, true, false, false)
			} else if field.IsList() {
				err = p.WriteList(&field, m, true, false, false)
			} else {
				// bacause basic type node buf no tag, just LV
				err = p.WriteBaseTypeWithDesc(&field, m, true, false, false)
			}
			// fmt.Println(id)
			require.Nil(b, err)
			newValue := generic.NewValue(&field, p.Buf)
			if field.IsMap() || field.IsList() {
				newValue = generic.NewComplexValue(&field, p.Buf)
			}
			// fmt.Println(id)
			require.Equal(b, len(newValue.Raw()), len(x.Raw()))
			_, err = newRoot.SetByPath(newValue.Node, generic.NewPathFieldId(proto.Number(id)))
			require.Nil(b, err)
			vv := newRoot.GetByPath(generic.NewPathFieldId(proto.Number(id)))
			require.Equal(b, len(newValue.Raw()), len(vv.Raw()))
			p.Recycle()
			ps = append(ps, generic.PathNode{ Path: generic.NewPathFieldId(proto.FieldNumber(id)), Node: newValue.Node })
		}

		n := generic.PathNode{
			Path: generic.NewPathFieldId(1),
			Node: newRoot.Node,
			Next: ps,
		}
		mProto, err := n.Marshal(&generic.Options{})
		if err != nil {
			b.Fatal("marshal PathNode failed")
		}
		require.Equal(b, len(mProto), len(mObj))

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				objRoot := generic.NewRootValue(desc, data)
				newRoot := generic.NewRootValue(desc, nil)
				ps := make([]generic.PathNode, 0)
				p := binary.NewBinaryProtocolBuffer()
				opt := &generic.Options{MapStructById: true}
				for id := 1; id <= testNums; id++ {
					x := objRoot.GetByPath(generic.NewPathFieldId(proto.Number(id)))
					m, _ := x.Interface(opt)
					field := (*desc).Fields().ByNumber(proto.Number(id))
					if field.IsMap() {
						_ = p.WriteMap(&field, m, true, false, false)
					} else if field.IsList() {
						_ = p.WriteList(&field, m, true, false, false)
					} else {
						// bacause basic type node buf no tag, just LV
						_ = p.WriteBaseTypeWithDesc(&field, m, true, false, false)
					}
					newValue := generic.NewValue(&field, p.Buf)
					// can not set when set complex value if use GetMany
					// if field.IsList() {
					// 	et := proto.FromProtoKindToType(field.Kind(), false, false)
					// 	newValue.SetElemType(et)
					// } else if field.IsMap() {
					// 	kt := proto.FromProtoKindToType(field.MapKey().Kind(), false, false)
					// 	newValue.SetKeyType(kt)
					// 	et := proto.FromProtoKindToType(field.MapValue().Kind(), false, false)
					// 	newValue.SetElemType(et)
					// }
					_, err = newRoot.SetByPath(newValue.Node, generic.NewPathFieldId(proto.Number(id)))
					p.Recycle()
					ps = append(ps, generic.PathNode{Path: generic.NewPathFieldId(proto.FieldNumber(id))})
				}
				_ = newRoot.GetMany(ps,opt)
				n := generic.PathNode{
					Path: generic.NewPathFieldId(1),
					Node: newRoot.Node,
					Next: ps,
				}

				_, _ = n.Marshal(opt)
			}
		})

	})
}

func BenchmarkProtoRationSetMany(b *testing.B) {
	b.Run("ration", func(b *testing.B) {
		desc := getPbNestingDesc()
		fieldNums := (*desc).Fields().Len()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}
		testNums := int(math.Ceil(float64(fieldNums) * factor))
		objRoot := generic.NewRootValue(desc, data)
		newRoot := generic.NewRootValue(desc, nil)
		ps := make([]generic.PathNode, testNums)
		opts := generic.Options{}
		mObj := make([]byte, 0)
		value := reflect.ValueOf(obj)
		
		for id := 1; id <= testNums; id++ {
			if err := collectMarshalData(id, value, &mObj); err != nil {
				b.Fatal("collect MarshalData failed")
			}
			ps[id-1] = generic.PathNode{Path: generic.NewPathFieldId(proto.FieldNumber(id))}
		}
		err := objRoot.GetMany(ps, &opts)
		if err != nil {
			b.Fatal("getMany failed")
		}
		adress2root := []int{}
		path2root := []generic.Path{}
		err = newRoot.SetMany(ps, &opts, &newRoot, adress2root, path2root...)
		if err != nil {
			b.Fatal("getMany failed")
		}
		err = newRoot.GetMany(ps, &generic.Options{ClearDirtyValues: true})
		if err != nil {
			b.Fatal("getMany failed")
		}

		n := generic.PathNode{
			Path: generic.NewPathFieldId(1),
			Node: newRoot.Node,
			Next: ps,
		}
		mProto, err := n.Marshal(&generic.Options{})
		if err != nil {
			b.Fatal("marshal PathNode failed")
		}
		require.Equal(b, len(mProto), len(mObj))
		fmt.Println(len(mProto))
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				objRoot := generic.NewRootValue(desc, data)
				newRoot := generic.NewRootValue(desc, nil)
				ps := make([]generic.PathNode, testNums)
				for id := 1; id <= testNums; id++ {
					ps[id-1] = generic.PathNode{Path: generic.NewPathFieldId(proto.FieldNumber(id))}
				}
				_ = objRoot.GetMany(ps, &opts)
				adress2root := make([]int, 0)
				path2root := make([]generic.Path, 0)
				newRoot.SetMany(ps, &opts, &newRoot, adress2root, path2root...)
				_ = newRoot.GetMany(ps, &generic.Options{ClearDirtyValues: true})
				n := generic.PathNode{
					Path: generic.NewPathFieldId(1),
					Node: newRoot.Node,
					Next: ps,
				}
				_, _ = n.Marshal(&generic.Options{})
				
			}
		})
	})
}

func BenchmarkDynamicpbRationSet(b *testing.B) {
	b.Run("ration", func(b *testing.B) {
		desc := getPbNestingDesc()
		fieldNums := (*desc).Fields().Len()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		objMessage := dynamicpb.NewMessage(*desc)
		if err := goproto.Unmarshal(data, objMessage); err != nil {
			b.Fatal("build dynamicpb failed")
		}

		testNums := int(math.Ceil(float64(fieldNums) * factor))

		message := dynamicpb.NewMessage(*desc)
		for id := 1; id <= testNums; id++ {
			targetDesc := (*desc).Fields().ByNumber(proto.Number(id))
			fieldValue := objMessage.Get(targetDesc)
			v := fieldValue.Interface()
			newValue := protoreflect.ValueOf(v)
			message.Set(targetDesc, newValue)
		}
		out, err := goproto.Marshal(message)
		if err != nil {
			b.Fatal("Dynamicpb marshal failed")
		}
		fmt.Println(len(out))
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				objMessage := dynamicpb.NewMessage(*desc)
				if err := goproto.Unmarshal(data, objMessage); err != nil {
					b.Fatal("build dynamicpb failed")
				}
				message := dynamicpb.NewMessage(*desc)
				for id := 1; id <= testNums; id++ {
					targetDesc := (*desc).Fields().ByNumber(proto.Number(id))
					fieldValue := objMessage.Get(targetDesc)
					message.Set(targetDesc, fieldValue)
				}
				_, _ = goproto.Marshal(message)
			}
		})
	})
}

func BenchmarkDynamicpbRationSetByInterface(b *testing.B) {
	b.Run("ration", func(b *testing.B) {
		desc := getPbNestingDesc()
		fieldNums := (*desc).Fields().Len()
		obj := getPbNestingValue()
		data := make([]byte, obj.Size())
		ret := obj.FastWrite(data)
		if ret != len(data) {
			b.Fatal(ret)
		}

		objMessage := dynamicpb.NewMessage(*desc)
		if err := goproto.Unmarshal(data, objMessage); err != nil {
			b.Fatal("build dynamicpb failed")
		}

		testNums := int(math.Ceil(float64(fieldNums) * factor))

		message := dynamicpb.NewMessage(*desc)
		for id := 1; id <= testNums; id++ {
			targetDesc := (*desc).Fields().ByNumber(proto.Number(id))
			fieldValue := objMessage.Get(targetDesc)
			v := fieldValue.Interface()
			newValue := protoreflect.ValueOf(v)
			message.Set(targetDesc, newValue)
		}
		out, err := goproto.Marshal(message)
		if err != nil {
			b.Fatal("Dynamicpb marshal failed")
		}
		fmt.Println(len(out))
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				objMessage := dynamicpb.NewMessage(*desc)
				if err := goproto.Unmarshal(data, objMessage); err != nil {
					b.Fatal("build dynamicpb failed")
				}
				message := dynamicpb.NewMessage(*desc)
				for id := 1; id <= testNums; id++ {
					targetDesc := (*desc).Fields().ByNumber(proto.Number(id))
					fieldValue := objMessage.Get(targetDesc)
					v := fieldValue.Interface()
					newValue := protoreflect.ValueOf(v)
					message.Set(targetDesc, newValue)
				}
				_, _ = goproto.Marshal(message)
			}
		})
	})
}

