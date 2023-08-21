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
	ejson "encoding/json"
	"math"
	stdh "net/http"
	"strconv"
	"strings"
	"sync"
	"testing"

	athrift "github.com/apache/thrift/lib/go/thrift"
	"github.com/bytedance/sonic"
	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/conv/j2t"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/json"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/baseline"
	"github.com/cloudwego/kitex/pkg/generic"
	"github.com/cloudwego/kitex/pkg/generic/descriptor"
	gthrift "github.com/cloudwego/kitex/pkg/generic/thrift"
	"github.com/cloudwego/kitex/pkg/remote"
	bthrift "github.com/cloudwego/kitex/pkg/remote/codec/thrift"
	"github.com/stretchr/testify/require"
)

const (
	idlPath = "idl/baseline.thrift"
)

var (
	simpleJSON  = ""
	nestingJSON = ""
)

func init() {
	sobj := getSimpleValue()
	sout, err := ejson.Marshal(sobj)
	if err != nil {
		panic(err)
	}
	simpleJSON = string(sout)
	println("small data size: ", len(simpleJSON))
	var out bytes.Buffer
	ejson.Indent(&out, sout, "", "")
	println(out.String())

	nobj := getNestingValue()
	nout, err := ejson.Marshal(nobj)
	if err != nil {
		panic(err)
	}
	nestingJSON = string(nout)
	println("medium data size: ", len(nestingJSON))
	out.Reset()

	psobj := getPartialSimpleValue()
	psout, err := ejson.Marshal(psobj)
	if err != nil {
		panic(err)
	}
	println("partial small data size: ", len(psout))

	pnobj := getPartialNestingValue()
	pnout, err := ejson.Marshal(pnobj)
	if err != nil {
		panic(err)
	}
	println("partial medium data size: ", len(pnout))
}

type Sample struct {
	name  string
	val   interface{}
	bytes []byte
}

var samples []Sample

var (
	bytesCount  int = 2
	stringCount int = 2
	listCount   int = 16
	mapCount    int = 16
)

func getSamples() []Sample {
	return []Sample{
		{samples[0].name, getSimpleValue(), samples[0].bytes},
		{samples[1].name, getNestingValue(), samples[1].bytes},
		{samples[2].name, getNesting2Value(), samples[2].bytes},
	}
}

func getString() string {
	return strings.Repeat("你好,\b\n\r\t世界", stringCount)
}

func getBytes() []byte {
	return bytes.Repeat([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, bytesCount)
}

func getSimpleValue() *baseline.Simple {
	return &baseline.Simple{
		ByteField:   math.MaxInt8,
		I64Field:    math.MaxInt64,
		DoubleField: math.MaxFloat64,
		I32Field:    math.MaxInt32,
		StringField: getString(),
		BinaryField: getBytes(),
	}
}

func getPartialSimpleValue() *baseline.PartialSimple {
	return &baseline.PartialSimple{
		ByteField:   math.MaxInt8,
		DoubleField: math.MaxFloat64,
		BinaryField: getBytes(),
	}
}

func getNestingValue() *baseline.Nesting {
	var ret = &baseline.Nesting{
		String_:         getString(),
		ListSimple:      []*baseline.Simple{},
		Double:          math.MaxFloat64,
		I32:             math.MaxInt32,
		ListI32:         []int32{},
		I64:             math.MaxInt64,
		MapStringString: map[string]string{},
		SimpleStruct:    getSimpleValue(),
		MapI32I64:       map[int32]int64{},
		ListString:      []string{},
		Binary:          getBytes(),
		MapI64String:    map[int64]string{},
		ListI64:         []int64{},
		Byte:            math.MaxInt8,
		MapStringSimple: map[string]*baseline.Simple{},
	}

	for i := 0; i < listCount; i++ {
		ret.ListSimple = append(ret.ListSimple, getSimpleValue())
		ret.ListI32 = append(ret.ListI32, math.MinInt32)
		ret.ListI64 = append(ret.ListI64, math.MinInt64)
		ret.ListString = append(ret.ListString, getString())
	}

	for i := 0; i < mapCount; i++ {
		ret.MapStringString[strconv.Itoa(i)] = getString()
		ret.MapI32I64[int32(i)] = math.MinInt64
		ret.MapI64String[int64(i)] = getString()
		ret.MapStringSimple[strconv.Itoa(i)] = getSimpleValue()
	}

	return ret
}

func getPartialNestingValue() *baseline.PartialNesting {
	var ret = &baseline.PartialNesting{
		ListSimple:      []*baseline.PartialSimple{},
		SimpleStruct:    getPartialSimpleValue(),
		MapStringSimple: map[string]*baseline.PartialSimple{},
	}

	for i := 0; i < listCount; i++ {
		ret.ListSimple = append(ret.ListSimple, getPartialSimpleValue())
	}

	for i := 0; i < mapCount; i++ {
		ret.MapStringSimple[strconv.Itoa(i)] = getPartialSimpleValue()
	}

	return ret
}

func getNesting2Value() *baseline.Nesting2 {
	var ret = &baseline.Nesting2{
		MapSimpleNesting: map[*baseline.Simple]*baseline.Nesting{},
		SimpleStruct:     getSimpleValue(),
		Byte:             math.MaxInt8,
		Double:           math.MaxFloat64,
		ListNesting:      []*baseline.Nesting{},
		I64:              math.MaxInt64,
		NestingStruct:    getNestingValue(),
		Binary:           getBytes(),
		String_:          getString(),
		SetNesting:       []*baseline.Nesting{},
		I32:              math.MaxInt32,
	}
	for i := 0; i < mapCount; i++ {
		ret.MapSimpleNesting[getSimpleValue()] = getNestingValue()
	}
	for i := 0; i < listCount; i++ {
		ret.ListNesting = append(ret.ListNesting, getNestingValue())
		x := getNestingValue()
		x.I64 = int64(i)
		ret.SetNesting = append(ret.SetNesting, x)
	}
	return ret
}

func getSampleHttpRequest(exp *baseline.Nesting, jbody string) *http.HTTPRequest {
	req := http.NewHTTPRequest()
	hr, err := stdh.NewRequest("POST", "localhost:8080", bytes.NewBufferString(jbody))
	if err != nil {
		panic(err)
	}
	hr.Header.Set("Content-Type", "application/json")
	req.Request = hr
	header := "你好"
	req.Request.Header.Set("String", header)
	exp.String_ = header
	// h2 := "abcdefghijklmnopqrstuvwxyz"
	// req.Header.Set("string_field", h2)
	// exp.SimpleStruct.StringField = h2
	// for i := range exp.ListSimple {
	// 	exp.ListSimple[i].StringField = h2
	// }
	// for k := range exp.MapStringSimple {
	// 	exp.MapStringSimple[k].StringField = h2
	// }

	c := []int64{-1, 0, math.MaxInt64, math.MinInt64}
	cookie := ""
	for i, v := range c {
		cookie += strconv.Itoa(int(v))
		if i != len(c)-1 {
			cookie += ","
		}
	}
	req.AddCookie(&stdh.Cookie{
		Name:  "list_i64",
		Value: cookie,
	})
	exp.ListI64 = c

	param := math.MaxFloat64
	req.Params.Set("double", strconv.FormatFloat(param, 'f', -1, 64))
	exp.Double = param

	q := []int32{-1, 0, math.MaxInt32, math.MinInt32}
	query := ""
	for i, v := range q {
		query += strconv.Itoa(int(v))
		if i != len(q)-1 {
			query += ","
		}
	}
	req.Request.URL.RawQuery = "ListI32=" + query
	exp.ListI32 = q
	return req
}

func getSampleHttpResponse(exp *baseline.Nesting) *http.HTTPResponse {
	req := http.NewHTTPResponse()

	code := int32(401)
	req.StatusCode = int(code)
	exp.I32 = code

	header := "你好"
	req.Header.Set("String", header)
	exp.String_ = header

	c := []int64{-1, 0, math.MaxInt64, math.MinInt64}
	cookie := ""
	for i, v := range c {
		cookie += strconv.Itoa(int(v))
		if i != len(c)-1 {
			cookie += ","
		}
	}
	req.SetCookie("list_i64", cookie)
	exp.ListI64 = c
	return req
}

// func TestThriftEncodeSimple_Load(t *testing.T) {
// 	_, err := ejson.Marshal(baseline.Simple{})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	simple := getSimpleDesc()
// 	// fmt.Printf("%#v", simple)
// 	root, err := json.NewSearcher(simpleJSON).GetByPath()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := root.LoadAll(); err != nil {
// 		t.Fatal(err)
// 	}
// 	enc := &json.ThriftEncoder{Proto: meta.ThriftBinary, Options: &json.Options{
// 		WriteDefault: true,
// 	}}
// 	out, err := enc.Encode(simple, root)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	spew.Dump(out)

// 	stru := baseline.NewSimple()
// 	if _, err := stru.FastRead(out); err != nil {
// 		t.Fatal(err)
// 	}
// 	stru2 := baseline.NewSimple()
// 	if err := sonic.UnmarshalString(simpleJSON, stru2); err != nil {
// 		t.Fatal(err)
// 	}
// 	require.Equal(t, stru2, stru)
// }

func convertI642StringSimple(js string) string {
	n, err := json.NewSearcher(js).GetByPath()
	if err != nil {
		panic(err)
	}
	old := n.Get("I64Field")
	if old.Check() != nil {
		panic(old)
	}
	new, err := old.Int64()
	if err != nil {
		panic(err)
	}
	_, err = n.Set("I64Field", json.NewString(strconv.Itoa(int(new))))
	if err != nil {
		panic(err)
	}
	e, _ := n.Raw()
	return e
}

func convertStr2I64Simple(js string) string {
	n, err := json.NewSearcher(js).GetByPath()
	if err != nil {
		panic(err)
	}
	old := n.Get("I64Field")
	if old.Check() != nil {
		panic(old)
	}
	s, err := old.String()
	if err != nil {
		panic(err)
	}
	_, err = n.Set("I64Field", json.NewNumber(s[1:len(s)-1]))
	if err != nil {
		panic(err)
	}
	e, _ := n.Raw()
	return e
}

func convertI642StringNesting(js string, itoa bool) string {
	n, err := json.NewSearcher(js).GetByPath()
	if err != nil {
		panic(err)
	}
	c := n.Get("SimpleStruct")
	r, _ := c.Raw()
	var new string
	if itoa {
		new = convertI642StringSimple(r)
	} else {
		new = convertStr2I64Simple(r)
	}
	_, err = n.Set("SimpleStruct", json.NewRaw(new))
	if err != nil {
		panic(err)
	}
	a := n.Get("ListSimple")
	if a.Check() != nil {
		panic(a)
	}
	a.ForEach(func(path json.Sequence, node *json.Node) bool {
		r, _ := node.Raw()
		var new string
		if itoa {
			new = convertI642StringSimple(r)
		} else {
			new = convertStr2I64Simple(r)
		}
		*node = json.NewRaw(new)
		return true
	})
	b := n.Get("MapStringSimple")
	if b.Check() != nil {
		panic(b)
	}
	b.ForEach(func(path json.Sequence, node *json.Node) bool {
		r, _ := node.Raw()
		var new string
		if itoa {
			new = convertI642StringSimple(r)
		} else {
			new = convertStr2I64Simple(r)
		}
		*node = json.NewRaw(new)
		return true
	})
	e, _ := n.Raw()
	return e
}

func TestJSON2Thrift_Simple(t *testing.T) {
	_, err := ejson.Marshal(baseline.Simple{})
	if err != nil {
		t.Fatal(err)
	}
	simple := getSimpleDesc()

	stru2 := baseline.NewSimple()
	if err := sonic.UnmarshalString(simpleJSON, stru2); err != nil {
		t.Fatal(err)
	}

	nj := convertI642StringSimple(simpleJSON)
	cv := j2t.NewBinaryConv(conv.Options{
		WriteDefaultField:  true,
		EnableValueMapping: true,
	})
	ctx := context.Background()
	out, err := cv.Do(ctx, simple, []byte(nj))
	require.Nil(t, err)

	stru := baseline.NewSimple()
	if _, err := stru.FastRead(out); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, stru2, stru)
}

var Concurrency = 1000

func TestJSON2Thrift_Simple_Parallel(t *testing.T) {
	_, err := ejson.Marshal(baseline.Simple{})
	if err != nil {
		t.Fatal(err)
	}
	desc := getSimpleDesc()
	stru2 := baseline.NewSimple()
	if err := sonic.UnmarshalString(simpleJSON, stru2); err != nil {
		t.Fatal(err)
	}
	nj := convertI642StringSimple(simpleJSON)

	cv := j2t.NewBinaryConv(conv.Options{
		WriteDefaultField:  true,
		EnableValueMapping: true,
	})

	wg := sync.WaitGroup{}
	for i := 0; i < Concurrency; i++ {
		wg.Add(1)
		go func(i int) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic: %d\n%s", i, nj)
				}
			}()
			defer wg.Done()
			ctx := context.Background()
			out, err := cv.Do(ctx, desc, []byte(nj))
			require.Nil(t, err)

			stru := baseline.NewSimple()
			if _, err := stru.FastRead(out); err != nil {
				t.Fatal(err)
			}
			require.Equal(t, stru2, stru)
		}(i)
	}

	wg.Wait()
}

// func TestThriftEncodeNesting_Load(t *testing.T) {
// 	nesting := getNestingDesc()
// 	// fmt.Printf("%#v", nesting)
// 	root, err := json.NewSearcher(nestingJSON).GetByPath()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := root.LoadAll(); err != nil {
// 		t.Fatal(err)
// 	}
// 	// js, err := root.MarshalJSON()
// 	// println(string(js))
// 	enc := &json.ThriftEncoder{Proto: meta.ThriftBinary, Options: &json.Options{
// 		WriteDefault: true,
// 	}}
// 	out, err := enc.Encode(nesting, root)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	// spew.Dump(out)

// 	stru := baseline.NewNesting()
// 	if _, err := stru.FastRead(out); err != nil {
// 		t.Fatal(err)
// 	}
// 	stru2 := baseline.NewNesting()
// 	if err := sonic.UnmarshalString(nestingJSON, stru2); err != nil {
// 		t.Fatal(err)
// 	}
// 	require.Equal(t, stru, stru2)
// 	// fmt.Printf("%#v", *stru)
// }

func TestHTTP2Thrift_Nesting(t *testing.T) {
	nesting := getNestingDesc()
	// fmt.Printf("%#v", nesting)
	stru2 := baseline.NewNesting()
	if err := ejson.Unmarshal([]byte(nestingJSON), stru2); err != nil {
		t.Fatal(err)
	}

	req := getSampleHttpRequest(stru2, nestingJSON)
	ctx := context.WithValue(context.Background(), conv.CtxKeyHTTPRequest, req)
	cv := j2t.NewBinaryConv(conv.Options{
		WriteDefaultField:            true,
		EnableHttpMapping:            true,
		EnableValueMapping:           true,
		TracebackRequredOrRootFields: true,
		ReadHttpValueFallback:        true,
	})
	nj := convertI642StringNesting(nestingJSON, true)
	out, err := cv.Do(ctx, nesting, []byte(nj))
	require.Nil(t, err)

	stru := baseline.NewNesting()
	if _, err := stru.FastRead(out); err != nil {
		t.Fatal(err)
	}

	require.Equal(t, stru2, stru)
	// fmt.Printf("%#v", *stru)
}

func TestHTTP2Thrift_Nesting_Parallel(t *testing.T) {
	nesting := getNestingDesc()
	// fmt.Printf("%#v", nesting)

	cv := j2t.NewBinaryConv(conv.Options{
		WriteDefaultField:            true,
		EnableHttpMapping:            true,
		EnableValueMapping:           true,
		TracebackRequredOrRootFields: true,
		ReadHttpValueFallback:        true,
		OmitHttpMappingErrors:        true,
	})
	nj := convertI642StringNesting(nestingJSON, true)
	println(nj)

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
			stru2 := baseline.NewNesting()
			if err := ejson.Unmarshal([]byte(nestingJSON), stru2); err != nil {
				t.Fatal(err)
			}
			req := getSampleHttpRequest(stru2, nestingJSON)
			ctx := context.WithValue(context.Background(), conv.CtxKeyHTTPRequest, req)
			out, err := cv.Do(ctx, nesting, []byte(nj))
			require.Nil(t, err)
			stru := baseline.NewNesting()
			if _, err := stru.FastRead(out); err != nil {
				t.Fatal(err)
			}
			require.Equal(t, stru2, stru)
		}(i)
	}
	wg.Wait()
}

// func BenchmarkJSON2Thrift_DynamicGo_Load(b *testing.B) {
// 	b.Run("small", func(b *testing.B) {
// 		simple := getSimpleDesc()
// 		enc := &json.ThriftEncoder{Proto: meta.ThriftBinary, Options: &json.Options{
// 			WriteDefault: true,
// 		}}
// 		jSimple, _ := json.NewSearcher(simpleJSON).GetByPath()
// 		jSimple.LoadAll()
// 		out, err := enc.Encode(simple, jSimple)
// 		if err != nil {
// 			b.Fatal(err)
// 		}
// 		b.SetBytes(int64(len(out)))
// 		b.ResetTimer()
// 		for i := 0; i < b.N; i++ {
// 			jSimple := json.NewRaw(simpleJSON)
// 			jSimple.LoadAll()
// 			_, _ = enc.Encode(simple, jSimple)
// 		}
// 	})

// 	b.Run("medium", func(b *testing.B) {
// 		nesting := getNestingDesc()
// 		enc := &json.ThriftEncoder{Proto: meta.ThriftBinary, Options: &json.Options{
// 			WriteDefault: true,
// 		}}
// 		jNesting, _ := json.NewSearcher(nestingJSON).GetByPath()
// 		jNesting.LoadAll()
// 		out, err := enc.Encode(nesting, jNesting)
// 		if err != nil {
// 			b.Fatal(err)
// 		}
// 		b.SetBytes(int64(len(out)))
// 		b.ResetTimer()
// 		for i := 0; i < b.N; i++ {
// 			jNesting := json.NewRaw(nestingJSON)
// 			jNesting.LoadAll()
// 			_, _ = enc.Encode(nesting, jNesting)
// 		}
// 	})
// }

// func BenchmarkJSON2Thrift_DynamicGo_Raw(b *testing.B) {

// 	b.Run("small", func(b *testing.B) {
// 		simple := getSimpleDesc()
// 		enc := &json.ThriftEncoder{Proto: meta.ThriftBinary, Options: &json.Options{
// 			WriteDefault: true,
// 		}}
// 		jSimple := json.NewRaw(simpleJSON)
// 		out, err := enc.Encode(simple, jSimple)
// 		if err != nil {
// 			b.Fatal(err)
// 		}
// 		b.SetBytes(int64(len(out)))
// 		b.ResetTimer()
// 		for i := 0; i < b.N; i++ {
// 			jSimple := json.NewRaw(simpleJSON)
// 			_, _ = enc.Encode(simple, jSimple)
// 		}
// 	})

// 	b.Run("medium", func(b *testing.B) {
// 		nesting := getNestingDesc()
// 		enc := &json.ThriftEncoder{Proto: meta.ThriftBinary, Options: &json.Options{
// 			WriteDefault: true,
// 		}}
// 		jNesting, _ := json.NewSearcher(nestingJSON).GetByPath()
// 		out, err := enc.Encode(nesting, jNesting)
// 		if err != nil {
// 			b.Fatal(err)
// 		}
// 		b.SetBytes(int64(len(out)))
// 		b.ResetTimer()
// 		for i := 0; i < b.N; i++ {
// 			jNesting := json.NewRaw(nestingJSON)
// 			_, _ = enc.Encode(nesting, jNesting)
// 		}
// 	})
// }

const BufferSize = 4096

func BenchmarkJSON2Thrift_KitexGeneric(b *testing.B) {
	p, err := generic.NewThriftFileProvider(idlPath)
	if err != nil {
		b.Fatal(err)
	}
	svcDsc := <-p.Provide()

	b.Run("small", func(b *testing.B) {
		var _args generic.Args
		_args.Method = "SimpleMethod"
		_args.Request = simpleJSON
		codec, err := gthrift.NewWriteJSON(svcDsc, "SimpleMethod", true)
		if err != nil {
			b.Fatal(err)
		}
		var mm = athrift.NewTMemoryBuffer()
		bc := athrift.NewTBinaryProtocol(mm, false, true)
		if err := codec.Write(context.Background(), bc, simpleJSON, nil); err != nil {
			b.Fatal(err)
		}

		b.SetBytes(int64(len(mm.Bytes())))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mm.Reset()
			_ = codec.Write(context.Background(), bc, simpleJSON, nil)
		}
	})

	b.Run("medium", func(b *testing.B) {
		var _args generic.Args
		_args.Method = "NestingMethod"
		_args.Request = nestingJSON
		codec, err := gthrift.NewWriteJSON(svcDsc, "NestingMethod", true)
		if err != nil {
			b.Fatal(err)
		}
		var mm = athrift.NewTMemoryBuffer()
		bc := athrift.NewTBinaryProtocol(mm, false, true)
		if err := codec.Write(context.Background(), bc, nestingJSON, nil); err != nil {
			b.Fatal(err)
		}

		b.SetBytes(int64(len(mm.Bytes())))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mm.Reset()
			_ = codec.Write(context.Background(), bc, nestingJSON, nil)
		}
	})
}

func BenchmarkJSON2Thrift_SonicAndKitex(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		v := baseline.NewSimple()
		if err := sonic.UnmarshalString(simpleJSON, v); err != nil {
			b.Fatal(err)
		}
		var buf = make([]byte, v.BLength())
		if err := v.FastWriteNocopy(buf, nil); err <= 0 {
			b.Fatal(err)
		}

		b.SetBytes(int64(len(buf)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v := baseline.NewSimple()
			_ = v.BLength()
			_ = sonic.UnmarshalString(simpleJSON, v)
			v.FastWriteNocopy(buf, nil)
		}
	})

	b.Run("medium", func(b *testing.B) {
		v := baseline.NewNesting()
		if err := sonic.UnmarshalString(nestingJSON, v); err != nil {
			b.Fatal(err)
		}
		var buf = make([]byte, v.BLength())
		if err := v.FastWriteNocopy(buf, nil); err <= 0 {
			b.Fatal(err)
		}

		b.SetBytes(int64(len(buf)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v := baseline.NewNesting()
			_ = v.BLength()
			_ = sonic.UnmarshalString(nestingJSON, v)
			v.FastWriteNocopy(buf, nil)
		}
	})
}

func BenchmarkHTTP2Thrift_DynamicGo_Raw(b *testing.B) {
	b.Run("small/value_mapping", func(b *testing.B) {
		simple := getSimpleDesc()
		cv := j2t.NewBinaryConv(conv.Options{
			WriteDefaultField:  true,
			EnableValueMapping: true,
		})
		nj := []byte(convertI642StringSimple(simpleJSON))
		ctx := context.Background()
		out, err := cv.Do(ctx, simple, nj)
		require.Nil(b, err)

		b.SetBytes(int64(len(out)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cv.Do(ctx, simple, nj)
		}
	})

	b.Run("medium/value_mapping", func(b *testing.B) {
		nesting := getNestingDesc()
		req := getSampleHttpRequest(baseline.NewNesting(), nestingJSON)
		cv := j2t.NewBinaryConv(conv.Options{
			WriteDefaultField:  true,
			EnableValueMapping: true,
		})
		nj := []byte(convertI642StringNesting(nestingJSON, true))
		ctx := context.Background()
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
		out, err := cv.Do(ctx, nesting, nj)
		require.Nil(b, err)

		b.SetBytes(int64(len(out)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cv.Do(ctx, nesting, nj)
		}
	})

	b.Run("medium/http+value_mapping", func(b *testing.B) {
		nesting := getNestingDesc()
		req := getSampleHttpRequest(baseline.NewNesting(), nestingJSON)
		ctx := context.Background()
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
		cv := j2t.NewBinaryConv(conv.Options{
			WriteDefaultField:  true,
			EnableHttpMapping:  true,
			EnableValueMapping: true,
		})
		nj := []byte(convertI642StringNesting(nestingJSON, true))
		out, err := cv.Do(ctx, nesting, nj)
		require.Nil(b, err)

		b.SetBytes(int64(len(out)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cv.Do(ctx, nesting, nj)
		}
	})
}

// func getKitexHttpRequest(req *descriptor.HTTPRequest) {
// 	header := "你好"
// 	req.Header.Set("String", header)
// 	// exp.String_ = header
// 	// h2 := "abcdefghijklmnopqrstuvwxyz"
// 	// req.Header.Set("string_field", h2)
// 	// exp.SimpleStruct.StringField = h2
// 	// for i := range exp.ListSimple {
// 	// 	exp.ListSimple[i].StringField = h2
// 	// }
// 	// for k := range exp.MapStringSimple {
// 	// 	exp.MapStringSimple[k].StringField = h2
// 	// }

// 	c := []int64{-1, 0, math.MaxInt64, math.MinInt64}
// 	cookie := ""
// 	for i, v := range c {
// 		cookie += strconv.Itoa(int(v))
// 		if i != len(c)-1 {
// 			cookie += ","
// 		}
// 	}
// 	req.Cookies["list_i64"] = cookie
// 	// exp.ListI64 = c

// 	// param := math.MaxFloat64
// 	// req.Params.Set("double", strconv.FormatFloat(param, 'f', -1, 64))
// 	// exp.Double = param

// 	q := []int32{-1, 0, math.MaxInt32, math.MinInt32}
// 	query := ""
// 	for i, v := range q {
// 		query += strconv.Itoa(int(v))
// 		if i != len(q)-1 {
// 			query += ","
// 		}
// 	}
// 	req.Query.Set("ListI32", query)
// 	// exp.ListI32 = q

// 	// exp.I32 = 0

// 	var helper = func(sim map[string]interface{}) {
// 		sim["I32Field"] = int32(sim["I32Field"].(int64))
// 		sim["ByteField"] = int8(sim["ByteField"].(int64))
// 	}
// 	for _, v := range req.Body["MapStringSimple"].(map[string]interface{}) {
// 		helper(v.(map[string]interface{}))
// 	}
// 	for _, v := range req.Body["ListSimple"].([]interface{}) {
// 		helper(v.(map[string]interface{}))
// 	}
// 	helper(req.Body["SimpleStruct"].(map[string]interface{}))
// 	req.Body["Byte"] = int8(req.Body["Byte"].(int64))

// }

func BenchmarkHTTP2Thrift_KitexGeneric(b *testing.B) {
	p, err := generic.NewThriftFileProvider(idlPath)
	if err != nil {
		b.Fatal(err)
	}
	svcDsc := <-p.Provide()
	svcDsc.Functions["NestingMethod"].Request.Struct.FieldsByName["req"].Type.Struct.FieldsByName["Double"].HTTPMapping = nil

	b.Run("small", func(b *testing.B) {
		codec := gthrift.NewWriteHTTPRequest(svcDsc)
		req := &descriptor.HTTPRequest{}
		req.Request, err = stdh.NewRequest("POST", "/simple", nil)
		if err != nil {
			b.Fatal(err)
		}
		jc := sonic.Config{
			UseInt64: true,
		}.Froze()
		if err := jc.UnmarshalFromString(simpleJSON, &req.Body); err != nil {
			b.Fatal(err)
		}
		req.Body["I32Field"] = int32(req.Body["I32Field"].(int64))
		req.Body["ByteField"] = int8(req.Body["ByteField"].(int64))

		buf := remote.NewWriterBuffer(BufferSize)
		bc := bthrift.NewBinaryProtocol(buf)
		if err := codec.Write(context.Background(), bc, req, gthrift.NewBase()); err != nil {
			b.Fatal(err)
		}
		out, _ := buf.Bytes()
		exp := baseline.NewSimple()
		if _, err := exp.FastRead(out); err != nil {
			b.Fatal(err)
		}

		b.SetBytes(int64(len(out)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := remote.NewWriterBuffer(BufferSize)
			bc := bthrift.NewBinaryProtocol(buf)
			if err = codec.Write(context.Background(), bc, req, gthrift.NewBase()); err != nil {
				b.Fatal(err)
			}
		}
	})

	// b.Run("medium", func(b *testing.B) {
	// 	codec := gthrift.NewWriteHTTPRequest(svcDsc)
	// 	jc := sonic.Config{
	// 		UseInt64: true,
	// 	}.Froze()
	// 	var body map[string]interface{}
	// 	if err := jc.UnmarshalFromString(nestingJSON, &body); err != nil {
	// 		b.Fatal(err)
	// 	}
	// 	req := &descriptor.HTTPRequest{
	// 		Header:  map[string][]string{},
	// 		Query:   map[string][]string{},
	// 		Cookies: map[string]string{},
	// 	}
	// 	req.Body = body
	// 	getKitexHttpRequest(req)
	// 	req.Method = "POST"
	// 	req.Path = "/nesting"

	// 	buf := remote.NewWriterBuffer(BufferSize)
	// 	bc := bthrift.NewBinaryProtocol(buf)
	// 	if err := codec.Write(context.Background(), bc, req, gthrift.NewBase()); err != nil {
	// 		b.Fatal(err)
	// 	}
	// 	out, _ := buf.Bytes()
	// 	exp := baseline.NewNesting()
	// 	if _, err := exp.FastRead(out); err != nil {
	// 		b.Fatal(err)
	// 	}

	// 	b.SetBytes(int64(len(out)))
	// 	b.ResetTimer()
	// 	for i := 0; i < b.N; i++ {
	// 		buf := remote.NewWriterBuffer(BufferSize)
	// 		bc := bthrift.NewBinaryProtocol(buf)
	// 		if err = codec.Write(context.Background(), bc, req, gthrift.NewBase()); err != nil {
	// 			b.Fatal(err)
	// 		}
	// 	}
	// })
}
