package testdata

import (
	"context"
	ejson "encoding/json"
	"errors"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/conv/p2j"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/baseline"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func getPbSimpleDesc() *protoreflect.MessageDescriptor {
	return proto.FnRequest(proto.GetFnDescFromFile(protoPath, "SimpleMethod", proto.Options{}))
}

func getPbPartialSimpleDesc() *protoreflect.MessageDescriptor {
	return proto.FnRequest(proto.GetFnDescFromFile(protoPath, "PartialSimpleMethod", proto.Options{}))
}

func getPbNestingDesc() *protoreflect.MessageDescriptor {
	return proto.FnRequest(proto.GetFnDescFromFile(protoPath, "NestingMethod", proto.Options{}))
}

func getPbPartialNestingDesc() *protoreflect.MessageDescriptor {
	return proto.FnRequest(proto.GetFnDescFromFile(protoPath, "PartialNestingMethod", proto.Options{}))
}

func FastReadPbSimpleObject(bakBuf []byte, obj *baseline.Simple) error {
	l := 0
	datalen := len(bakBuf)
	for l < datalen {
		id, wtyp, tagLen := protowire.ConsumeTag(bakBuf)
		if tagLen < 0 {
			return errors.New("proto data error format")
		}
		l += tagLen
		bakBuf = bakBuf[tagLen:]
		offset, err := obj.FastRead(bakBuf, int8(wtyp), int32(id))
		if err != nil {
			return err
		}
		bakBuf = bakBuf[offset:]
		l += offset
	}
	if len(bakBuf) != 0 {
		return errors.New("proto data error format")
	}
	return nil
}

func FastReadPbMediumObject(bakBuf []byte, obj *baseline.Nesting) error {
	l := 0
	datalen := len(bakBuf)
	for l < datalen {
		id, wtyp, tagLen := protowire.ConsumeTag(bakBuf)
		if tagLen < 0 {
			return errors.New("proto data error format")
		}
		l += tagLen
		bakBuf = bakBuf[tagLen:]
		offset, err := obj.FastRead(bakBuf, int8(wtyp), int32(id))
		if err != nil {
			return err
		}
		bakBuf = bakBuf[offset:]
		l += offset
	}
	if len(bakBuf) != 0 {
		return errors.New("proto data error format")
	}
	return nil
}

func TestProtobuf2JSON(t *testing.T) {
	t.Run("small", func(t *testing.T) {
		ctx := context.Background()
		desc := getPbSimpleDesc()
		// kitex to build pbData, and serialize into bytes
		data := getPbSimpleValue()
		cv := p2j.NewBinaryConv(conv.Options{})
		in := make([]byte, data.Size())
		data.FastWrite(in)

		ret, err := cv.Do(ctx, desc, in)
		if err != nil {
			t.Fatal(err)
		}
		println(string(ret))

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
		// kitex to build pbData, and serialize into bytes
		data := getPbNestingValue()
		cv := p2j.NewBinaryConv(conv.Options{})
		in := make([]byte, data.Size())
		data.FastWrite(in)

		ret, err := cv.Do(ctx, desc, in)
		if err != nil {
			t.Fatal(err)
		}
		println(string(ret))

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
		// kitex to build pbData, and serialize into bytes
		data := getPbSimpleValue()
		cv := p2j.NewBinaryConv(conv.Options{})
		in := make([]byte, data.Size())
		data.FastWrite(in)

		_, err := cv.Do(ctx, desc, in)
		if err != nil {
			b.Fatal(err)
		}
		b.SetBytes(int64(len(in)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cv.Do(ctx, desc, in)
		}
	})

	b.Run("medium", func(b *testing.B) {
		ctx := context.Background()
		desc := getPbNestingDesc()
		// kitex to build pbData, and serialize into bytes
		data := getPbNestingValue()
		cv := p2j.NewBinaryConv(conv.Options{})
		in := make([]byte, data.Size())
		data.FastWrite(in)

		_, err := cv.Do(ctx, desc, in)
		if err != nil {
			b.Fatal(err)
		}
		b.SetBytes(int64(len(in)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cv.Do(ctx, desc, in)
		}
	})
}

// func BenchmarkProtobuf2JSON_KitexGeneric(b *testing.B) {
// 	p, err := generic.NewPbContentProvider(idlPath)
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	svcDesc := <-p.Provide()

// 	b.Run("small", func(b *testing.B) {
// 		codec := gproto.NewMessage()
// 		data := getPbSimpleValue()
// 		in := make([]byte, data.Size())
// 		data.FastWrite(in)
// 		in =
// 	})
// }

func BenchmarkProtobuf2JSON_SonicAndKitex(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		v := getPbSimpleValue()
		var buf = make([]byte, v.Size())
		v.FastWrite(buf)

		// kitex_fastRead build object
		obj := &baseline.Simple{}
		if err := FastReadPbSimpleObject(buf, obj); err != nil {
			b.Fatal(err)
		}

		// sonic marshal object
		_, err := sonic.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
		// println(string(out))

		b.SetBytes(int64(len(buf)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj := &baseline.Simple{}
			_ = FastReadPbSimpleObject(buf, obj)
			_, _ = sonic.Marshal(obj)
		}
	})

	b.Run("medium", func(b *testing.B) {
		v := getPbNestingValue()
		var buf = make([]byte, v.Size())
		v.FastWrite(buf)

		// kitex_fastRead build object
		obj := &baseline.Nesting{}
		if err := FastReadPbMediumObject(buf, obj); err != nil {
			b.Fatal(err)
		}

		// sonic marshal object
		_, err := sonic.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
		// println(string(out))

		b.SetBytes(int64(len(buf)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj := &baseline.Nesting{}
			_ = FastReadPbMediumObject(buf, obj)
			_, _ = sonic.Marshal(obj)
		}
	})
}

func BenchmarkProtobuf2JSON_ProtoBufGo(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		data := getPbSimpleValue()
		out, err := protojson.Marshal(data.ProtoReflect().Interface())
		if err != nil {
			b.Fatal(err)
		}
		// println(string(out))

		b.SetBytes(int64(len(out)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = protojson.Marshal(data.ProtoReflect().Interface())
		}
	})

	b.Run("medium", func(b *testing.B) {
		data := getPbNestingValue()
		out, err := protojson.Marshal(data.ProtoReflect().Interface())
		if err != nil {
			b.Fatal(err)
		}
		// println(string(out))

		b.SetBytes(int64(len(out)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = protojson.Marshal(data.ProtoReflect().Interface())
		}
	})
}
