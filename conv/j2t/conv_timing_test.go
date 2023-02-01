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

package j2t

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example3"

	// "github.com/cloudwego/dynamicgo/thrift"
	// athrift "github.com/apache/thrift/lib/go/thrift"
	// kg "github.com/cloudwego/kitex/pkg/generic"
	// kd "github.com/cloudwego/kitex/pkg/generic/descriptor"
	// gthrift "github.com/cloudwego/kitex/pkg/generic/thrift"
	"github.com/stretchr/testify/require"
)

func BenchmarkConvJSON2Thrift_DynamicGo(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()
	cv := NewBinaryConv(conv.Options{})
	ctx := context.Background()
	out, err := cv.Do(ctx, desc, data)
	require.NoError(b, err)
	exp := example3.NewExampleReq()
	err = json.Unmarshal(data, exp)
	require.Nil(b, err)
	act := example3.NewExampleReq()
	_, err = act.FastRead(out)
	require.Nil(b, err)
	require.Equal(b, exp, act)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out = out[:0]
		_ = cv.DoInto(ctx, desc, data, &out)
	}
}

func BenchmarkConvHTTP2Thrift_DynamicGo(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()
	exp := example3.NewExampleReq()
	err := json.Unmarshal(data, exp)
	require.Nil(b, err)
	req := getExampleReq(exp, true, data)
	cv := NewBinaryConv(conv.Options{
		EnableHttpMapping: true,
	})
	ctx := context.Background()
	ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
	out, err := cv.Do(ctx, desc, data)
	require.NoError(b, err)
	act := example3.NewExampleReq()
	_, err = act.FastRead(out)
	require.Nil(b, err)
	require.Equal(b, exp, act)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cv.Do(ctx, desc, data)
	}
}

func BenchmarkConvJSON2Thrift_Parallel_DynamicGo(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()
	cv := NewBinaryConv(conv.Options{})
	ctx := context.Background()
	out, err := cv.Do(ctx, desc, data)
	require.NoError(b, err)
	exp := example3.NewExampleReq()
	err = json.Unmarshal(data, exp)
	require.Nil(b, err)
	act := example3.NewExampleReq()
	_, err = act.FastRead(out)
	require.Nil(b, err)
	require.Equal(b, exp, act)

	b.ResetTimer()
	b.RunParallel(func(b *testing.PB) {
		buf := make([]byte, 0, len(out))
		for b.Next() {
			buf = buf[:0]
			_ = cv.DoInto(ctx, desc, data, &buf)
		}
	})
}

func BenchmarkConvHTTP2Thrift_Parallel_DynamicGo(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()
	exp := example3.NewExampleReq()
	err := json.Unmarshal(data, exp)
	require.Nil(b, err)
	req := getExampleReq(exp, true, data)
	cv := NewBinaryConv(conv.Options{
		EnableHttpMapping: true,
	})
	ctx := context.Background()
	ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
	out, err := cv.Do(ctx, desc, data)
	require.NoError(b, err)
	act := example3.NewExampleReq()
	_, err = act.FastRead(out)
	require.Nil(b, err)
	require.Equal(b, exp, act)

	b.ResetTimer()
	b.RunParallel(func(b *testing.PB) {
		for b.Next() {
			_, _ = cv.Do(ctx, desc, data)
		}
	})
}

// func getKitexGenericDesc() *kd.ServiceDescriptor {
// 	p, err := kg.NewThriftFileProvider(exampleIDLPath)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	return <-p.Provide()
// }

// func wrapKitexGenericRequestPayload(in []byte) []byte {
// 	out := make([]byte, 0, len(in)+4)
// 	p := thrift.NewBinaryProtocol(out)
// 	p.WriteFieldBegin("", athrift.STRUCT, 1)
// 	p.Buf = append(p.Buf, in...)
// 	p.WriteFieldEnd()
// 	p.WriteStructEnd()
// 	return p.Buf
// }

// func BenchmarkThriftMarshalAll_KitexGeneric(b *testing.B) {
// 	b.Run("sequential", func(b *testing.B) {
// 		data := string(getExampleData())
// 		svcDsc := getKitexGenericDesc()
// 		var _args kg.Args
// 		_args.Method = "ExampleMethod"
// 		_args.Request = string(data)
// 		codec, err := gthrift.NewWriteJSON(svcDsc, "ExampleMethod", true)
// 		if err != nil {
// 			b.Fatal(err)
// 		}
// 		var mm = athrift.NewTMemoryBuffer()
// 		bc := athrift.NewTBinaryProtocol(mm, false, true)
// 		if err := codec.Write(context.Background(), bc, data, nil); err != nil {
// 			b.Fatal(err)
// 		}

// 		b.ResetTimer()
// 		for i := 0; i < b.N; i++ {
// 			mm.Reset()
// 			_ = codec.Write(context.Background(), bc, data, nil)
// 		}
// 	})
// }
