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

package t2j

import (
	"context"
	"testing"

	"github.com/cloudwego/dynamicgo/conv"
	cv "github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/thrift"
)

func BenchmarkThrift2JSON_DynamicGo(t *testing.B) {
	desc := thrift.FnResponse(thrift.GetFnDescFromFile("testdata/idl/example3.thrift", "ExampleMethod", thrift.Options{}))
	conv := NewBinaryConv(conv.Options{})
	in := getExample3Data()
	ctx := context.Background()
	out, err := conv.Do(ctx, desc, in)
	if err != nil {
		t.Fatal(err)
	}
	t.SetBytes(int64(len(in)))
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		out = out[:0]
		_ = conv.DoInto(ctx, desc, in, &out)
	}
}

func BenchmarkThrift2HTTP_DynamicGo(b *testing.B) {
	desc := thrift.FnResponse(thrift.GetFnDescFromFile("testdata/idl/example3.thrift", "ExampleMethod", thrift.Options{}))
	conv := NewBinaryConv(conv.Options{
		EnableValueMapping: true,
		EnableHttpMapping:  true,
	})
	in := getExample3Data()
	ctx := context.Background()
	resp := http.NewHTTPResponse()
	ctx = context.WithValue(ctx, cv.CtxKeyHTTPResponse, resp)
	out, err := conv.DoHTTP(ctx, desc, in)
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(in)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		resp := http.NewHTTPResponse()
		ctx = context.WithValue(ctx, cv.CtxKeyHTTPResponse, resp)
		out = out[:0]
		_, _ = conv.DoHTTP(ctx, desc, in)
	}
}

func BenchmarkThrift2JSON_Parallel_DynamicGo(t *testing.B) {
	desc := thrift.FnResponse(thrift.GetFnDescFromFile("testdata/idl/example3.thrift", "ExampleMethod", thrift.Options{}))
	conv := NewBinaryConv(conv.Options{})
	in := getExample3Data()
	ctx := context.Background()
	out, err := conv.Do(ctx, desc, in)
	if err != nil {
		t.Fatal(err)
	}

	t.SetBytes(int64(len(in)))
	t.ResetTimer()
	t.RunParallel(func(p *testing.PB) {
		buf := make([]byte, len(out))
		for p.Next() {
			buf = buf[:0]
			_ = conv.DoInto(ctx, desc, in, &buf)
		}
	})
}

func BenchmarkThrift2HTTP_Parallel_DynamicGo(b *testing.B) {
	desc := thrift.FnResponse(thrift.GetFnDescFromFile("testdata/idl/example3.thrift", "ExampleMethod", thrift.Options{}))
	conv := NewBinaryConv(conv.Options{
		EnableValueMapping: true,
		EnableHttpMapping:  true,
	})
	in := getExample3Data()
	ctx := context.Background()
	resp := http.NewHTTPResponse()
	ctx = context.WithValue(ctx, cv.CtxKeyHTTPResponse, resp)
	_, err := conv.DoHTTP(ctx, desc, in)
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(int64(len(in)))
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			ctx := context.Background()
			resp := http.NewHTTPResponse()
			ctx = context.WithValue(ctx, cv.CtxKeyHTTPResponse, resp)
			_, _ = conv.DoHTTP(ctx, desc, in)
		}
	})
}
