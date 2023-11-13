package p2j

import (
	"context"
	"testing"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/proto"
)

func BenchmarkProtobuf2JSON_DynamicGo(t *testing.B) {
	messageDesc := proto.FnRequest(proto.GetFnDescFromFile("testdata/idl/example2.proto", "ExampleMethod", proto.Options{}))
	desc, ok := (*messageDesc).(proto.Descriptor)
	if !ok {
		t.Fatal("invalid descriptor")
	}
	conv := NewBinaryConv(conv.Options{})
	in := readExampleReqProtoBufData()
	ctx := context.Background()
	out, err := conv.Do(ctx, &desc, in)
	//print(string(out))
	if err != nil {
		t.Fatal(err)
	}
	t.SetBytes(int64(len(in)))
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		out = out[:0]
		_ = conv.DoInto(ctx, &desc, in, &out)
	}
}

func BenchmarkProtobuf2JSON_Parallel_DynamicGo(t *testing.B) {
	messageDesc := proto.FnRequest(proto.GetFnDescFromFile("testdata/idl/example2.proto", "ExampleMethod", proto.Options{}))
	desc, ok := (*messageDesc).(proto.Descriptor)
	if !ok {
		t.Fatal("invalid descriptor")
	}
	conv := NewBinaryConv(conv.Options{})
	in := readExampleReqProtoBufData()
	ctx := context.Background()
	out, err := conv.Do(ctx, &desc, in)
	//print(string(out))
	if err != nil {
		t.Fatal(err)
	}
	t.SetBytes(int64(len(in)))
	t.ResetTimer()
	t.RunParallel(func(p *testing.PB) {
		buf := make([]byte, len(out))
		for p.Next() {
			buf = buf[:0]
			_ = conv.DoInto(ctx, &desc, in, &buf)
		}
	})
}
