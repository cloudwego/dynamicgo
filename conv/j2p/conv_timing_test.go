package j2p

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example2"
	"github.com/stretchr/testify/require"
	goprotowire "google.golang.org/protobuf/encoding/protowire"
)

func BenchmarkConvJSON2Protobuf_DynamicGo(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()
	cv := NewBinaryConv(conv.Options{})
	ctx := context.Background()
	// dynamicgo exec json2pb
	out, err := cv.Do(ctx, desc, data)
	require.Nil(b, err)

	// unmarshal json to get pbObj
	exp := example2.ExampleReq{}
	err = json.Unmarshal(data, &exp)
	require.Nil(b, err)
	act := example2.ExampleReq{}
	l := 0
	dataLen := len(out)
	for l < dataLen {
		id, wtyp, tagLen := goprotowire.ConsumeTag(out)
		if tagLen < 0 {
			b.Fatal("error pb data format")
		}
		l += tagLen
		out = out[tagLen:]
		offset, err := act.FastRead(out, int8(wtyp), int32(id))
		require.Nil(b, err)
		out = out[offset:]
		l += offset
	}
	require.Equal(b, exp, act)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out = out[:0]
		_ = cv.DoInto(ctx, desc, data, &out)
	}
}

func BenchmarkConvJSON2Protobuf_Parallel_DynamicGo(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()
	cv := NewBinaryConv(conv.Options{})
	ctx := context.Background()
	// dynamicgo exec json2pb
	out, err := cv.Do(ctx, desc, data)
	require.Nil(b, err)

	// unmarshal json to get pbObj
	exp := example2.ExampleReq{}
	err = json.Unmarshal(data, &exp)
	require.Nil(b, err)
	act := example2.ExampleReq{}
	l := 0
	dataLen := len(out)
	for l < dataLen {
		id, wtyp, tagLen := goprotowire.ConsumeTag(out)
		if tagLen < 0 {
			b.Fatal("error pb data format")
		}
		l += tagLen
		out = out[tagLen:]
		offset, err := act.FastRead(out, int8(wtyp), int32(id))
		require.Nil(b, err)
		out = out[offset:]
		l += offset
	}
	require.Equal(b, exp, act)

	b.ResetTimer()
	b.RunParallel(func(b *testing.PB) {
		buf := make([]byte, 0, dataLen)
		for b.Next() {
			buf := buf[:0]
			_ = cv.DoInto(ctx, desc, data, &buf)
		}
	})
}
