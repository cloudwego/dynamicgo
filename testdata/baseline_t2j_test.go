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

//go:generate kitex -module=github.com/cloudwego/dynamicgo idl/baseline.thrift
package testdata

import (
	"bytes"
	"context"
	"encoding/json"
	ejson "encoding/json"
	"sync"
	"testing"

	athrift "github.com/apache/thrift/lib/go/thrift"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/conv/t2j"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/baseline"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/cloudwego/kitex/pkg/generic"
	gthrift "github.com/cloudwego/kitex/pkg/generic/thrift"
	"github.com/stretchr/testify/require"
)

func getSimpleDesc() *thrift.TypeDescriptor {
	svc, err := thrift.NewDescritorFromPath(context.Background(), idlPath)
	if err != nil {
		panic(err)
	}
	return svc.Functions()["SimpleMethod"].Request().Struct().FieldByKey("req").Type()
}

func getPartialSimpleDesc() *thrift.TypeDescriptor {
	svc, err := thrift.NewDescritorFromPath(context.Background(), idlPath)
	if err != nil {
		panic(err)
	}
	return svc.Functions()["PartialSimpleMethod"].Request().Struct().FieldByKey("req").Type()
}

func getNestingDesc() *thrift.TypeDescriptor {
	svc, err := thrift.NewDescritorFromPath(context.Background(), idlPath)
	if err != nil {
		panic(err)
	}
	return svc.Functions()["NestingMethod"].Request().Struct().FieldByKey("req").Type()
}

func getPartialNestingDesc() *thrift.TypeDescriptor {
	svc, err := thrift.NewDescritorFromPath(context.Background(), idlPath)
	if err != nil {
		panic(err)
	}
	return svc.Functions()["PartialNestingMethod"].Request().Struct().FieldByKey("req").Type()
}

func TestThrift2JSON(t *testing.T) {
	t.Run("small", func(t *testing.T) {
		ctx := context.Background()
		desc := getSimpleDesc()
		data := getSimpleValue()
		cv := t2j.NewBinaryConv(conv.Options{})
		in := make([]byte, data.BLength())
		if err := data.FastWriteNocopy(in, nil); err <= 0 {
			t.Fatal(err)
		}

		ret, err := cv.Do(ctx, desc, in)
		if err != nil {
			t.Fatal(err)
		}
		// println(string(ret))
		var v = baseline.NewSimple()
		if err := ejson.Unmarshal(ret, v); err != nil {
			t.Fatal(err)
		}
		require.Equal(t, data, v)
	})
	t.Run("medium", func(t *testing.T) {
		ctx := context.Background()
		desc := getNestingDesc()
		data := getNestingValue()
		cv := t2j.NewBinaryConv(conv.Options{})
		in := make([]byte, data.BLength())
		if err := data.FastWriteNocopy(in, nil); err <= 0 {
			t.Fatal(err)
		}

		ret, err := cv.Do(ctx, desc, in)
		if err != nil {
			t.Fatal(err)
		}
		// println(string(ret))
		var v = baseline.NewNesting()
		if err := json.Unmarshal(ret, v); err != nil {
			t.Fatal(err)
		}
		require.Equal(t, data, v)
	})
}

func TestThrift2JSON_Parallel(t *testing.T) {
	t.Run("small", func(t *testing.T) {
		ctx := context.Background()
		desc := getSimpleDesc()
		data := getSimpleValue()
		cv := t2j.NewBinaryConv(conv.Options{})
		in := make([]byte, data.BLength())
		if err := data.FastWriteNocopy(in, nil); err <= 0 {
			t.Fatal(err)
		}

		wg := sync.WaitGroup{}
		for i := 0; i < Concurrency; i++ {
			wg.Add(1)
			go func(i int) {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("panic: %d\n", i)
					}
				}()
				defer wg.Done()

				ret, err := cv.Do(ctx, desc, in)
				if err != nil {
					t.Fatal(err)
				}
				// println(string(ret))
				var v = baseline.NewSimple()
				if err := ejson.Unmarshal(ret, v); err != nil {
					t.Fatal(err)
				}
				require.Equal(t, data, v)
			}(i)
		}
		wg.Wait()
	})
	t.Run("medium", func(t *testing.T) {
		ctx := context.Background()
		desc := getNestingDesc()
		data := getNestingValue()
		cv := t2j.NewBinaryConv(conv.Options{})
		in := make([]byte, data.BLength())
		if err := data.FastWriteNocopy(in, nil); err <= 0 {
			t.Fatal(err)
		}

		wg := sync.WaitGroup{}
		for i := 0; i < Concurrency; i++ {
			wg.Add(1)
			go func(i int) {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("panic: %d\n", i)
					}
				}()
				defer wg.Done()
				ret, err := cv.Do(ctx, desc, in)
				if err != nil {
					t.Fatal(err)
				}
				// println(string(ret))
				var v = baseline.NewNesting()
				if err := json.Unmarshal(ret, v); err != nil {
					t.Fatal(err)
				}
				require.Equal(t, data, v)
			}(i)
		}
		wg.Wait()
	})
}

func TestThrift2HTTP(t *testing.T) {
	t.Run("small", func(t *testing.T) {
		desc := getSimpleDesc()
		data := getSimpleValue()
		in := make([]byte, data.BLength())
		if err := data.FastWriteNocopy(in, nil); err <= 0 {
			t.Fatal(err)
		}

		opts := conv.Options{}
		opts.EnableHttpMapping = true
		opts.WriteHttpValueFallback = true
		ctx := context.Background()
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPResponse, http.NewHTTPResponse())
		cv := t2j.NewBinaryConv(opts)
		out, err := cv.Do(ctx, desc, in)
		if err != nil {
			t.Fatal(err)
		}

		// resp := ctx.Value(conv.CtxKeyHTTPResponse).(*http.HTTPResponse)
		// out, err := io.ReadAll(resp.Body)
		// require.Nil(t, err)
		var v = baseline.NewSimple()
		if err := ejson.Unmarshal(out, v); err != nil {
			t.Fatal(err)
		}
		require.Equal(t, data, v)

		opts.EnableValueMapping = true
		cv = t2j.NewBinaryConv(opts)
		ret, err := cv.Do(ctx, desc, in)
		if err != nil {
			t.Fatal(err)
		}
		nr := convertStr2I64Simple(string(ret))
		err = ejson.Unmarshal([]byte(nr), v)
		require.Nil(t, err)
	})

	t.Run("medium", func(t *testing.T) {
		desc := getNestingDesc()
		data := getNestingValue()
		in := make([]byte, data.BLength())
		if err := data.FastWriteNocopy(in, nil); err <= 0 {
			t.Fatal(err)
		}

		opts := conv.Options{}
		opts.EnableHttpMapping = true
		opts.WriteHttpValueFallback = true
		opts.OmitHttpMappingErrors = true

		ctx := context.Background()
		resp := http.NewHTTPResponse()
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPResponse, resp)
		cv := t2j.NewBinaryConv(opts)
		out, err := cv.Do(ctx, desc, in)
		if err != nil {
			t.Fatal(err)
		}
		// out, err := io.ReadAll(resp.Body)
		// require.Nil(t, err)
		var v = baseline.NewNesting()
		if err := json.Unmarshal(out, v); err != nil {
			t.Fatal(err)
		}
		dstr := data.String_
		data.String_ = ""
		di32 := data.I32
		data.I32 = 0
		ls, err := json.Marshal(data.ListI64)
		require.NoError(t, err)
		data.ListI64 = nil
		
		require.Equal(t, data, v)
		require.Equal(t, dstr, resp.Header.Get("String"))
		require.Equal(t, int(di32), resp.StatusCode)
		require.Equal(t, string(ls), resp.Cookies()[0].Value)

		opts.EnableValueMapping = true
		cv = t2j.NewBinaryConv(opts)
		ret, err := cv.Do(ctx, desc, in)
		if err != nil {
			t.Fatal(err)
		}
		v = baseline.NewNesting()
		ret = []byte(convertI642StringNesting(string(ret), false))
		_ = ejson.Unmarshal(ret, v)
		require.Equal(t, string(ls), resp.Cookies()[0].Value)
		require.Equal(t, dstr, resp.Header.Get("String"))
		require.Equal(t, int(di32), resp.StatusCode)
	})
}

func TestThrift2HTTP_Parallel(t *testing.T) {
	t.Run("small", func(t *testing.T) {
		desc := getSimpleDesc()
		opts := conv.Options{}
		opts.EnableHttpMapping = true
		opts.WriteHttpValueFallback = true
		cv := t2j.NewBinaryConv(opts)

		wg := sync.WaitGroup{}
		for i := 0; i < Concurrency; i++ {
			wg.Add(1)
			go func(i int) {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("panic: %d\n", i)
					}
				}()
				defer wg.Done()
				data := getSimpleValue()
				in := make([]byte, data.BLength())
				if err := data.FastWriteNocopy(in, nil); err <= 0 {
					t.Fatal(err)
				}
				ctx := context.Background()
				ctx = context.WithValue(ctx, conv.CtxKeyHTTPResponse, http.NewHTTPResponse())
				out, err := cv.Do(ctx, desc, in)
				require.NoError(t, err)
				var v = baseline.NewSimple()
				if err := ejson.Unmarshal(out, v); err != nil {
					t.Fatal(err)
				}
				require.Equal(t, data, v)
			}(i)
		}
		wg.Wait()
	})

	t.Run("medium", func(t *testing.T) {
		desc := getNestingDesc()
		opts := conv.Options{}
		opts.EnableHttpMapping = true
		opts.WriteHttpValueFallback = true
		opts.OmitHttpMappingErrors = true
		cv := t2j.NewBinaryConv(opts)

		wg := sync.WaitGroup{}
		for i := 0; i < Concurrency; i++ {
			wg.Add(1)
			go func(i int) {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("panic: %d\n", i)
					}
				}()
				defer wg.Done()
				data := getNestingValue()
				in := make([]byte, data.BLength())
				if err := data.FastWriteNocopy(in, nil); err <= 0 {
					t.Fatal(err)
				}
				ctx := context.Background()
				resp := http.NewHTTPResponse()
				ctx = context.WithValue(ctx, conv.CtxKeyHTTPResponse, resp)
				out, err := cv.Do(ctx, desc, in)
				if err != nil {
					t.Fatal(err)
				}
				var v = baseline.NewNesting()
				if err := json.Unmarshal(out, v); err != nil {
					t.Fatal(err)
				}
				dstr := data.String_
				data.String_ = ""
				di32 := data.I32
				data.I32 = 0
				ls, err := json.Marshal(data.ListI64)
				require.NoError(t, err)
				data.ListI64 = nil
				
				require.Equal(t, data, v)
				require.Equal(t, dstr, resp.Header.Get("String"))
				require.Equal(t, int(di32), resp.StatusCode)
				require.Equal(t, string(ls), resp.Cookies()[0].Value)
			}(i)
		}
		wg.Wait()
	})
}

func BenchmarkThrift2JSON_DynamicGo_Raw(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		ctx := context.Background()
		desc := getSimpleDesc()
		cv := t2j.NewBinaryConv(conv.Options{})
		data := getSimpleValue()
		in := make([]byte, data.BLength())
		if err := data.FastWriteNocopy(in, nil); err <= 0 {
			b.Fatal(err)
		}
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
		desc := getNestingDesc()
		data := getNestingValue()
		cv := t2j.NewBinaryConv(conv.Options{})
		in := make([]byte, data.BLength())
		if err := data.FastWriteNocopy(in, nil); err <= 0 {
			b.Fatal(err)
		}
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

func wrapKitexGenericRequestPayload(in []byte) []byte {
	out := make([]byte, 0, len(in)+4)
	p := thrift.NewBinaryProtocol(out)
	p.WriteFieldBegin("", athrift.STRUCT, 1)
	p.Buf = append(p.Buf, in...)
	p.WriteFieldEnd()
	p.WriteStructEnd()
	return p.Buf
}

func wrapKitexGenericResponsePayload(in []byte) []byte {
	out := make([]byte, 0, len(in)+4)
	p := thrift.NewBinaryProtocol(out)
	p.WriteFieldBegin("", athrift.STRUCT, 0)
	p.Buf = append(p.Buf, in...)
	p.WriteFieldEnd()
	p.WriteStructEnd()
	return p.Buf
}

func BenchmarkThrift2JSON_KitexGeneric(b *testing.B) {
	p, err := generic.NewThriftFileProvider(idlPath)
	if err != nil {
		b.Fatal(err)
	}
	svcDsc := <-p.Provide()

	b.Run("small", func(b *testing.B) {
		codec := gthrift.NewReadJSON(svcDsc, false)
		data := getSimpleValue()
		in := make([]byte, data.BLength())
		if err := data.FastWriteNocopy(in, nil); err <= 0 {
			b.Fatal(err)
		}
		in = wrapKitexGenericRequestPayload(in)
		var mm = athrift.NewStreamTransportR(bytes.NewBuffer(in))
		bc := athrift.NewTBinaryProtocol(mm, false, false)
		v, err := codec.Read(context.Background(), "SimpleMethod", bc)
		if err != nil {
			b.Fatal(err)
		}
		_ = v
		// spew.Printf("%#+v\n", v)

		b.SetBytes(int64(len(in)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var mm = athrift.NewStreamTransportR(bytes.NewBuffer(in))
			bc := athrift.NewTBinaryProtocol(mm, false, false)
			_, _ = codec.Read(context.Background(), "SimpleMethod", bc)
		}
	})

	b.Run("medium", func(b *testing.B) {
		codec := gthrift.NewReadJSON(svcDsc, false)
		data := getNestingValue()
		in := make([]byte, data.BLength())
		if err := data.FastWriteNocopy(in, nil); err <= 0 {
			b.Fatal(err)
		}
		in = wrapKitexGenericRequestPayload(in)
		mm := athrift.NewStreamTransportR(bytes.NewBuffer(in))
		bc := athrift.NewTBinaryProtocol(mm, false, false)
		v, err := codec.Read(context.Background(), "NestingMethod", bc)
		if err != nil {
			b.Fatal(err)
		}
		_ = v
		// spew.Printf("%#+v\n", v)

		b.SetBytes(int64(len(in)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mm = athrift.NewStreamTransportR(bytes.NewBuffer(in))
			bc = athrift.NewTBinaryProtocol(mm, false, false)
			_, _ = codec.Read(context.Background(), "NestingMethod", bc)
		}
	})
}

func BenchmarkThrift2JSON_SonicAndKitex(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		v := getSimpleValue()
		var buf = make([]byte, v.BLength())
		if err := v.FastWriteNocopy(buf, nil); err <= 0 {
			b.Fatal(err)
		}

		obj := baseline.NewSimple()
		_, err := obj.FastRead(buf)
		if err != nil {
			b.Fatal(err)
		}
		_, err = sonic.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
		// println(string(out))

		b.SetBytes(int64(len(buf)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj := baseline.NewSimple()
			_, _ = obj.FastRead(buf)
			_, _ = sonic.Marshal(obj)
		}
	})

	b.Run("medium", func(b *testing.B) {
		v := getNestingValue()
		var buf = make([]byte, v.BLength())
		if err := v.FastWriteNocopy(buf, nil); err <= 0 {
			b.Fatal(err)
		}

		obj := baseline.NewNesting()
		_, err := obj.FastRead(buf)
		if err != nil {
			b.Fatal(err)
		}
		_, err = sonic.Marshal(obj)
		if err != nil {
			b.Fatal(err)
		}
		// println(string(out))

		b.SetBytes(int64(len(buf)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj := baseline.NewNesting()
			_, _ = obj.FastRead(buf)
			_, _ = sonic.Marshal(obj)
		}
	})
}

func BenchmarkThrift2HTTP_DynamicGo(t *testing.B) {
	t.Run("small/value_mapping", func(t *testing.B) {
		desc := getSimpleDesc()
		data := getSimpleValue()
		in := make([]byte, data.BLength())
		if err := data.FastWriteNocopy(in, nil); err <= 0 {
			t.Fatal(err)
		}

		opts := conv.Options{}
		opts.EnableValueMapping = false
		opts.OmitHttpMappingErrors = true
		ctx := context.Background()
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPResponse, http.NewHTTPResponse())
		cv := t2j.NewBinaryConv(opts)
		ret, err := cv.Do(ctx, desc, in)
		if err != nil {
			t.Fatal(err)
		}
		ret = []byte(convertStr2I64Simple(string(ret)))
		if err := ejson.Unmarshal(ret, baseline.NewSimple()); err != nil {
			t.Fatal(err)
		}

		t.SetBytes(int64(len(in)))
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			_, err = cv.Do(ctx, desc, in)
		}
	})

	t.Run("medium/http_mapping", func(t *testing.B) {
		desc := getNestingDesc()
		data := getNestingValue()
		in := make([]byte, data.BLength())
		if err := data.FastWriteNocopy(in, nil); err <= 0 {
			t.Fatal(err)
		}

		opts := conv.Options{}
		opts.EnableHttpMapping = true
		opts.EnableValueMapping = false
		opts.OmitHttpMappingErrors = true
		opts.NoCopyString = true

		ctx := context.Background()
		resp := http.NewHTTPResponse()
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPResponse, resp)
		cv := t2j.NewBinaryConv(opts)
		ret, err := cv.Do(ctx, desc, in)
		if err != nil {
			t.Fatal(err)
		}
		if err := ejson.Unmarshal(ret, baseline.NewNesting()); err != nil {
			t.Fatal(err)
		}

		t.SetBytes(int64(len(in)))
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			_, err = cv.Do(ctx, desc, in)
		}
	})

	t.Run("medium/http+value_mapping", func(t *testing.B) {
		desc := getNestingDesc()
		data := getNestingValue()
		in := make([]byte, data.BLength())
		if err := data.FastWriteNocopy(in, nil); err <= 0 {
			t.Fatal(err)
		}

		opts := conv.Options{}
		opts.EnableHttpMapping = true
		opts.EnableValueMapping = true
		opts.OmitHttpMappingErrors = true
		opts.NoCopyString = true

		ctx := context.Background()
		resp := http.NewHTTPResponse()
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPResponse, resp)
		cv := t2j.NewBinaryConv(opts)
		ret, err := cv.Do(ctx, desc, in)
		if err != nil {
			t.Fatal(err)
		}
		ret = []byte(convertI642StringNesting(string(ret), false))
		if err := ejson.Unmarshal(ret, baseline.NewNesting()); err != nil {
			t.Fatal(err)
		}

		t.SetBytes(int64(len(in)))
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			_, err = cv.Do(ctx, desc, in)
		}
	})
}

func BenchmarkThrift2HTTP_KitexGeneric(b *testing.B) {
	p, err := generic.NewThriftFileProvider(idlPath)
	if err != nil {
		b.Fatal(err)
	}
	svcDsc := <-p.Provide()

	b.Run("small/http+value_mapping", func(b *testing.B) {
		codec := gthrift.NewReadHTTPResponse(svcDsc)
		data := getSimpleValue()
		in := make([]byte, data.BLength())
		if err := data.FastWriteNocopy(in, nil); err <= 0 {
			b.Fatal(err)
		}
		in = wrapKitexGenericResponsePayload(in)
		var mm = athrift.NewStreamTransportR(bytes.NewBuffer(in))
		bc := athrift.NewTBinaryProtocol(mm, false, false)
		v, err := codec.Read(context.Background(), "SimpleMethod", bc)
		if err != nil {
			b.Fatal(err)
		}
		_ = v
		// spew.Printf("%#+v\n", v)

		b.SetBytes(int64(len(in)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var mm = athrift.NewStreamTransportR(bytes.NewBuffer(in))
			bc := athrift.NewTBinaryProtocol(mm, false, false)
			_, _ = codec.Read(context.Background(), "SimpleMethod", bc)
		}
	})

	b.Run("medium/http+value_mapping", func(b *testing.B) {
		codec := gthrift.NewReadHTTPResponse(svcDsc)
		data := getNestingValue()
		in := make([]byte, data.BLength())
		if err := data.FastWriteNocopy(in, nil); err <= 0 {
			b.Fatal(err)
		}
		in = wrapKitexGenericResponsePayload(in)
		mm := athrift.NewStreamTransportR(bytes.NewBuffer(in))
		bc := athrift.NewTBinaryProtocol(mm, false, false)
		v, err := codec.Read(context.Background(), "NestingMethod", bc)
		if err != nil {
			b.Fatal(err)
		}
		_ = v
		// spew.Printf("%#+v\n", v)

		b.SetBytes(int64(len(in)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mm = athrift.NewStreamTransportR(bytes.NewBuffer(in))
			bc = athrift.NewTBinaryProtocol(mm, false, false)
			_, _ = codec.Read(context.Background(), "NestingMethod", bc)
		}
	})
}
