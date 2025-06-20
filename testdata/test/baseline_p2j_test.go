package test

import (
	"context"
	ejson "encoding/json"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/conv/p2j"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/baseline"
	"github.com/cloudwego/prutal"
	"github.com/stretchr/testify/require"
)

func getPbSimpleDesc() *proto.TypeDescriptor {
	return proto.FnRequest(proto.GetFnDescFromFile(protoPath, "SimpleMethod", proto.Options{}))
}

func getPbPartialSimpleDesc() *proto.TypeDescriptor {
	return proto.FnRequest(proto.GetFnDescFromFile(protoPath, "PartialSimpleMethod", proto.Options{}))
}

func getPbNestingDesc() *proto.TypeDescriptor {
	return proto.FnRequest(proto.GetFnDescFromFile(protoPath, "NestingMethod", proto.Options{}))
}

func getPbPartialNestingDesc() *proto.TypeDescriptor {
	return proto.FnRequest(proto.GetFnDescFromFile(protoPath, "PartialNestingMethod", proto.Options{}))
}

func TestProtobuf2JSON(t *testing.T) {
	t.Run("small", func(t *testing.T) {
		ctx := context.Background()
		desc := getPbSimpleDesc()
		data := getPbSimpleValue()
		cv := p2j.NewBinaryConv(conv.Options{})
		in, err := prutal.Marshal(data)
		require.Nil(t, err)

		ret, err := cv.Do(ctx, desc, in)
		require.Nil(t, err)

		// unserialize json to object
		v := &baseline.Simple{}
		if err := ejson.Unmarshal(ret, v); err != nil {
			t.Fatal(err)
		}
		require.Equal(t, data, v)
	})

	t.Run("medium", func(t *testing.T) {
		ctx := context.Background()
		desc := getPbNestingDesc()
		data := getPbNestingValue()
		cv := p2j.NewBinaryConv(conv.Options{})
		in, err := prutal.Marshal(data)
		require.Nil(t, err)

		ret, err := cv.Do(ctx, desc, in)
		require.Nil(t, err)

		// unserialize json to object
		v := &baseline.Nesting{}
		if err := ejson.Unmarshal(ret, v); err != nil {
			t.Fatal(err)
		}
		require.Equal(t, data, v)
	})
}

func BenchmarkProtobuf2JSON_DynamicGo_Raw(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		ctx := context.Background()
		desc := getPbSimpleDesc()
		data := getPbSimpleValue()
		cv := p2j.NewBinaryConv(conv.Options{})
		in, err := prutal.Marshal(data)
		require.Nil(b, err)

		_, err = cv.Do(ctx, desc, in)
		require.Nil(b, err)

		b.SetBytes(int64(len(in)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cv.Do(ctx, desc, in)
		}
	})

	b.Run("medium", func(b *testing.B) {
		ctx := context.Background()
		desc := getPbNestingDesc()
		data := getPbNestingValue()
		cv := p2j.NewBinaryConv(conv.Options{})
		in, err := prutal.Marshal(data)
		require.Nil(b, err)

		_, err = cv.Do(ctx, desc, in)
		require.Nil(b, err)

		b.SetBytes(int64(len(in)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cv.Do(ctx, desc, in)
		}
	})
}

func BenchmarkProtobuf2JSON_SonicAndPrutal(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		v := getPbSimpleValue()
		buf, err := prutal.Marshal(v)
		require.Nil(b, err)

		// build object
		obj := &baseline.Simple{}
		err = prutal.Unmarshal(buf, obj)
		require.Nil(b, err)

		// sonic marshal object
		_, err = sonic.Marshal(obj)
		require.Nil(b, err)

		b.SetBytes(int64(len(buf)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj := &baseline.Simple{}
			_ = prutal.Unmarshal(buf, obj)
			_, _ = sonic.Marshal(obj)
		}
	})

	b.Run("medium", func(b *testing.B) {
		v := getPbNestingValue()
		buf, err := prutal.Marshal(v)
		require.Nil(b, err)

		// use prutal to build object
		obj := &baseline.Nesting{}
		err = prutal.Unmarshal(buf, obj)
		require.Nil(b, err)

		// sonic marshal object
		_, err = sonic.Marshal(obj)
		require.Nil(b, err)

		b.SetBytes(int64(len(buf)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj := &baseline.Nesting{}
			_ = prutal.Unmarshal(buf, obj)
			_, _ = sonic.Marshal(obj)
		}
	})
}
