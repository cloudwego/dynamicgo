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
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"testing"
	"time"
	"unsafe"

	"github.com/cloudwego/dynamicgo/conv"
	cv "github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/util_test"
	sjson "github.com/cloudwego/dynamicgo/json"
	"github.com/cloudwego/dynamicgo/meta"
	kbase "github.com/cloudwego/dynamicgo/testdata/kitex_gen/base"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example3"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/cloudwego/dynamicgo/thrift/base"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	debugAsyncGC = os.Getenv("SONIC_NO_ASYNC_GC") == ""
)

func TestMain(m *testing.M) {
	go func() {
		if !debugAsyncGC {
			return
		}
		println("Begin GC looping...")
		for {
			runtime.GC()
			debug.FreeOSMemory()
		}
	}()
	time.Sleep(time.Millisecond)
	m.Run()
}

func GetDescByName(method string, isReq bool) *thrift.TypeDescriptor {
	opts := thrift.Options{}
	svc, err := opts.NewDescritorFromPath(util_test.MustGitPath("testdata/idl/example3.thrift"))
	if err != nil {
		panic(err)
	}
	if isReq {
		return svc.Functions()[method].Request().Struct().Fields()[0].Type()
	} else {
		return svc.Functions()[method].Response().Struct().Fields()[0].Type()
	}
}

func getExamplePartialDesc() *thrift.TypeDescriptor {
	svc, err := thrift.NewDescritorFromPath(util_test.MustGitPath("testdata/idl/example3.thrift"))
	if err != nil {
		panic(err)
	}
	return svc.Functions()["PartialMethod"].Response().Struct().Fields()[0].Type()
}

func getExampleInt2Float() *thrift.TypeDescriptor {
	svc, err := thrift.NewDescritorFromPath(util_test.MustGitPath("testdata/idl/example3.thrift"))
	if err != nil {
		panic(err)
	}
	return svc.Functions()["Int2FloatMethod"].Response().Struct().Fields()[0].Type()
}

func getExamplePartialDesc2() *thrift.TypeDescriptor {
	svc, err := thrift.NewDescritorFromPath(util_test.MustGitPath("testdata/idl/example3.thrift"))
	if err != nil {
		panic(err)
	}
	return svc.Functions()["PartialMethod"].Request().Struct().Fields()[0].Type()
}

func getExampleErrorDesc() *thrift.TypeDescriptor {
	svc, err := thrift.NewDescritorFromPath(util_test.MustGitPath("testdata/idl/example3.thrift"))
	if err != nil {
		panic(err)
	}
	return svc.Functions()["ErrorMethod"].Response().Struct().Fields()[0].Type()
}

func getExampleFallbackDesc() *thrift.TypeDescriptor {
	svc, err := thrift.NewDescritorFromPath(util_test.MustGitPath("testdata/idl/example3.thrift"))
	if err != nil {
		panic(err)
	}
	return svc.Functions()["FallbackMethod"].Response().Struct().Fields()[0].Type()
}

func getExample3Data() []byte {
	out, err := ioutil.ReadFile(util_test.MustGitPath("testdata/data/example3resp.bin"))
	if err != nil {
		panic(err)
	}
	return out
}

func getExmaple3JSON() string {
	out, err := ioutil.ReadFile(util_test.MustGitPath("testdata/data/example3resp.json"))
	if err != nil {
		panic(err)
	}
	return string(out)
}

func TestConvThrift2JSON(t *testing.T) {
	desc := thrift.FnResponse(thrift.GetFnDescFromFile("testdata/idl/example3.thrift", "ExampleMethod", thrift.Options{}))
	// js := getExmaple3JSON()
	conv := NewBinaryConv(conv.Options{})
	in := getExample3Data()
	out, err := conv.Do(context.Background(), desc, in)
	if err != nil {
		t.Fatal(err)
	}
	var exp, act example3.ExampleResp
	// require.Nil(t, json.Unmarshal([]byte(js), &exp))
	_, err = exp.FastRead(in)
	require.Nil(t, err)
	require.Nil(t, json.Unmarshal([]byte(out), &act))
	println(string(out))
	assert.Equal(t, exp, act)
	// assert.Equal(t, len(js), len(out))
}

// func getHTTPResponse() *http.HTTPResponse {
// 	return &http.HTTPResponse{
// 		StatusCode: 200,
// 		Header:     map[string][]string{},
// 		Cookies:    map[string]*stdh.Cookie{},
// 		RawBody:    []byte{},
// 	}
// }

func TestConvThrift2HTTP(t *testing.T) {
	desc := thrift.FnResponse(thrift.GetFnDescFromFile("testdata/idl/example3.thrift", "ExampleMethod", thrift.Options{}))
	expJSON := getExmaple3JSON()
	conv := NewBinaryConv(conv.Options{
		// MapRecurseDepth:    conv.DefaultMaxDepth,
		EnableValueMapping: true,
		EnableHttpMapping:  true,
	})
	in := getExample3Data()
	ctx := context.Background()
	resp := http.NewHTTPResponse()
	ctx = context.WithValue(ctx, cv.CtxKeyHTTPResponse, resp)
	out, err := conv.Do(ctx, desc, in)
	if err != nil {
		t.Fatal(err)
	}
	println(expJSON, string(out))
	chechHelper(t, expJSON, string(out), testOptions{
		JSConv: true,
	})
	// buf, err := io.ReadAll(resp.Response.Body)
	// require.Nil(t, err)
	// chechHelper(t, expJSON, string(buf))
	assert.Equal(t, "true", resp.Header["Heeader"][0])
	if len(resp.Header) != 2 {
		spew.Dump(resp.Header)
	}
	assert.Equal(t, 2, len(resp.Header))
	assert.Equal(t, "cookie", resp.Response.Cookies()[0].Name)
	assert.Equal(t, "-1e-7", resp.Response.Cookies()[0].Value)
	assert.Equal(t, "inner_string", resp.Response.Cookies()[1].Name)
	assert.Equal(t, "hello", resp.Response.Cookies()[1].Value)

	ctx = context.Background()
	resp = http.NewHTTPResponse()
	ctx = context.WithValue(ctx, cv.CtxKeyHTTPResponse, resp)
	out, err = conv.Do(ctx, desc, in)
	if err != nil {
		t.Fatal(err)
	}
	chechHelper(t, expJSON, string(out), testOptions{
		JSConv: true,
	})
	// buf, err = io.ReadAll(resp.Response.Body)
	// require.Nil(t, err)
	// chechHelper(t, expJSON, string(buf))
	assert.Equal(t, "true", resp.Header["Heeader"][0])
	assert.Equal(t, 2, len(resp.Header))
	assert.Equal(t, "cookie", resp.Response.Cookies()[0].Name)
	assert.Equal(t, "-1e-7", resp.Response.Cookies()[0].Value)
	assert.Equal(t, "inner_string", resp.Response.Cookies()[1].Name)
	assert.Equal(t, "hello", resp.Response.Cookies()[1].Value)
}

func chechHelper(t *testing.T, exp string, act string, opts testOptions) {
	var expMap, actMap map[string]interface{}
	require.Nil(t, json.Unmarshal([]byte(exp), &expMap))
	require.Nil(t, json.Unmarshal([]byte(act), &actMap))
	assert.True(t, checkAInB(t, reflect.ValueOf(actMap), reflect.ValueOf(expMap), opts))
}

type testOptions struct {
	JSConv bool
}

func checkAInB(t *testing.T, a reflect.Value, b reflect.Value, opts testOptions) bool {
	if a.IsZero() {
		return true
	}
	if a.Kind() == reflect.Ptr || a.Kind() == reflect.Interface {
		a = a.Elem()
	}
	if b.Kind() == reflect.Ptr || b.Kind() == reflect.Interface {
		b = b.Elem()
	}
	// if a.Type() != b.Type() {
	// 	t.Logf("type mismatch: %v != %v", a.Type(), b.Type())
	// 	return false
	// }
	if a.Kind() != reflect.Map && a.Kind() != reflect.Slice {
		if !reflect.DeepEqual(a.Interface(), b.Interface()) {
			// a := a.Elem()
			// b := b.Elem()
			if opts.JSConv {
				switch a.Kind() {
				case reflect.String:
					x, _ := strconv.ParseFloat(a.String(), 64)
					if x != (b.Float()) {
						return false
					}
					return true
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					x, _ := strconv.Atoi(b.String())
					if x != int(a.Int()) {
						return false
					}
					return true
				}
			}
			t.Logf("value mismatch: %#v != %#v", a.Interface(), b.Interface())
			return false
		}
	}
	if a.Kind() == reflect.Map {
		for _, k := range a.MapKeys() {
			if !checkAInB(t, a.MapIndex(k), b.MapIndex(k), opts) {
				return false
			}
		}
	}
	if a.Kind() == reflect.Slice {
		for i := 0; i < a.Len(); i++ {
			if !checkAInB(t, a.Index(i), b.Index(i), opts) {
				return false
			}
		}
	}
	return true
}

func TestThriftResponseBase(t *testing.T) {
	desc := thrift.FnResponse(thrift.GetFnDescFromFile("testdata/idl/example3.thrift", "ExampleMethod", thrift.Options{
		EnableThriftBase: true,
	}))
	js := getExmaple3JSON()
	conv := NewBinaryConv(conv.Options{
		EnableThriftBase: true,
	})
	in := getExample3Data()

	ctx := context.Background()
	out, err := conv.Do(ctx, desc, in)
	require.NoError(t, err)
	var exp, act = new(example3.ExampleResp), new(example3.ExampleResp)
	require.Nil(t, json.Unmarshal([]byte(js), &exp))
	require.Nil(t, json.Unmarshal([]byte(out), &act))
	require.Equal(t, exp, act)

	b := base.NewBaseResp()
	require.NoError(t, err)
	ctx = context.WithValue(ctx, cv.CtxKeyThriftRespBase, b)
	out, err = conv.Do(ctx, desc, in)
	require.NoError(t, err)
	exp, act = new(example3.ExampleResp), new(example3.ExampleResp)
	require.Nil(t, json.Unmarshal([]byte(js), &exp))
	require.Nil(t, json.Unmarshal([]byte(out), &act))
	require.NotEqual(t, exp, act)
	n, err := sjson.NewSearcher(js).GetByPath("BaseResp")
	bj, err := n.Raw()
	require.NoError(t, err)
	require.Nil(t, json.Unmarshal([]byte(bj), &act.BaseResp))
	require.Equal(t, (*base.BaseResp)(unsafe.Pointer(act.BaseResp)), b)
}

func TestAGWBodyDynamic(t *testing.T) {
	conv := NewBinaryConv(conv.Options{
		EnableValueMapping: true,
	})
	desc := getExampleErrorDesc()
	exp := example3.NewExampleErrorResp()
	exp.Int64 = 1
	exp.Xjson = `{"b":1}`
	ctx := context.Background()
	in := make([]byte, exp.BLength())
	_ = exp.FastWriteNocopy(in, nil)
	out, err := conv.Do(ctx, desc, in)
	require.NoError(t, err)
	expj := (`{"Int64":1,"Xjson":{"b":1}}`)
	require.Equal(t, expj, string(out))

	conv.opts.EnableValueMapping = false
	out, err = conv.Do(ctx, desc, in)
	require.NoError(t, err)
	require.Equal(t, (`{"Int64":1,"Xjson":"{\"b\":1}"}`), string(out))
}

func TestInt2String(t *testing.T) {
	conv := NewBinaryConv(conv.Options{
		EnableValueMapping: true,
	})
	desc := getExampleInt2Float()
	exp := example3.NewExampleInt2Float()
	exp.Int32 = 1
	exp.Int64 = 2
	exp.Float64 = 3.14
	exp.String_ = "hello"
	exp.Subfix = 0.92653
	ctx := context.Background()
	in := make([]byte, exp.BLength())
	_ = exp.FastWriteNocopy(in, nil)

	out, err := conv.Do(ctx, desc, in)
	require.NoError(t, err)
	require.Equal(t, `{"Int32":"1","Float64":"3.14","Int64":2,"Subfix":0.92653,"中文":"hello"}`, string(out))

	conv.opts.EnableValueMapping = false
	out, err = conv.Do(ctx, desc, in)
	require.NoError(t, err)
	require.Equal(t, (`{"Int32":1,"Float64":3.14,"Int64":2,"Subfix":0.92653,"中文":"hello"}`), string(out))

	conv.opts.String2Int64 = true
	out, err = conv.Do(ctx, desc, in)
	require.NoError(t, err)
	require.Equal(t, (`{"Int32":1,"Float64":3.14,"Int64":"2","Subfix":0.92653,"中文":"hello"}`), string(out))
}

func TestHttpMappingFallback(t *testing.T) {
	conv := NewBinaryConv(cv.Options{})
	t.Run("not as extra", func(t *testing.T) {
		desc := getExampleFallbackDesc()
		exp := example3.NewExampleFallback()
		exp.Msg = "hello"
		exp.Heeader = "world"
		in := make([]byte, exp.BLength())
		_ = exp.FastWriteNocopy(in, nil)
		expJSON := `{"Msg":"hello"}`
		conv.SetOptions(cv.Options{
			EnableHttpMapping:  true,
			HttpMappingAsExtra: false,
		})
		ctx := context.Background()
		resp := http.NewHTTPResponse()
		ctx = context.WithValue(ctx, cv.CtxKeyHTTPResponse, resp)
		out, err := conv.Do(ctx, desc, in)
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, expJSON, string(out))
		assert.Equal(t, "world", resp.Header["Heeader"][0])
	})
	t.Run("as extra", func(t *testing.T) {
		desc := getExampleFallbackDesc()
		exp := example3.NewExampleFallback()
		exp.Msg = "hello"
		exp.Heeader = "world"
		in := make([]byte, exp.BLength())
		_ = exp.FastWriteNocopy(in, nil)
		expJSON := `{"Msg":"hello","Heeader":"world"}`
		conv.SetOptions(cv.Options{
			EnableHttpMapping:  true,
			HttpMappingAsExtra: true,
		})
		ctx := context.Background()
		resp := http.NewHTTPResponse()
		ctx = context.WithValue(ctx, cv.CtxKeyHTTPResponse, resp)
		out, err := conv.Do(ctx, desc, in)
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, expJSON, string(out))
		assert.Equal(t, "world", resp.Header["Heeader"][0])
	})
}

func TestWriteEmpty(t *testing.T) {
	desc := thrift.FnResponse(thrift.GetFnDescFromFile("testdata/idl/example3.thrift", "ExampleMethod", thrift.Options{}))
	p := thrift.NewBinaryProtocolBuffer()
	p.WriteStructBegin("ExampleResp")
	p.WriteFieldBegin("Status", thrift.I32, 3)
	expStatus := int32(23)
	p.WriteI32(expStatus)
	p.WriteFieldEnd()
	p.WriteStructEnd()
	data := p.Buf
	conv := NewBinaryConv(conv.Options{
		WriteDefaultField: true,
		EnableHttpMapping: true,
	})
	ctx := context.Background()
	resp := http.NewHTTPResponse()
	ctx = context.WithValue(ctx, cv.CtxKeyHTTPResponse, resp)
	out, err := conv.Do(ctx, desc, data)
	if err != nil {
		t.Fatal(err)
	}
	var exp, act example3.ExampleResp
	_, err = exp.FastRead(data)
	require.NoError(t, err)
	require.Nil(t, json.Unmarshal([]byte(out), &act))
	assert.NotEqual(t, exp, act)
	assert.Equal(t, int32(0), act.Status)
	assert.Equal(t, expStatus, int32(resp.StatusCode))
	assert.Equal(t, &kbase.BaseResp{
		StatusMessage: "",
		StatusCode:    0,
		Extra:         map[string]string{},
	}, act.BaseResp)
}

func TestUnknowFields(t *testing.T) {
	t.Run("top", func(t *testing.T) {
		desc := getExamplePartialDesc()
		// js := getExmaple3JSON()
		conv := NewBinaryConv(conv.Options{
			DisallowUnknownField: true,
		})
		in := getExample3Data()
		_, err := conv.Do(context.Background(), desc, in)
		require.Error(t, err)
		require.Equal(t, meta.ErrUnknownField, err.(meta.Error).Code.Behavior())
	})

	t.Run("nested", func(t *testing.T) {
		desc := getExamplePartialDesc2()
		// js := getExmaple3JSON()
		conv := NewBinaryConv(conv.Options{
			DisallowUnknownField: true,
		})
		in := getExample3Data()
		_, err := conv.Do(context.Background(), desc, in)
		require.Error(t, err)
		require.Equal(t, meta.ErrUnknownField, err.(meta.Error).Code.Behavior())
	})

	t.Run("skip top", func(t *testing.T) {
		desc := getExamplePartialDesc()
		// js := getExmaple3JSON()
		conv := NewBinaryConv(conv.Options{
			DisallowUnknownField: false,
		})
		in := getExample3Data()
		_, err := conv.Do(context.Background(), desc, in)
		require.NoError(t, err)
	})

	t.Run("skip nested", func(t *testing.T) {
		desc := getExamplePartialDesc2()
		// js := getExmaple3JSON()
		conv := NewBinaryConv(conv.Options{
			DisallowUnknownField: false,
		})
		in := getExample3Data()
		_, err := conv.Do(context.Background(), desc, in)
		require.NoError(t, err)
	})
}

func TestNobodyRequiredFields(t *testing.T) {
	desc := GetDescByName("Base64BinaryMethod", true)

	t.Run("base64 encode", func(t *testing.T) {
		conv := NewBinaryConv(conv.Options{
			EnableHttpMapping: true,
			NoBase64Binary:    false,
		})
		exp := example3.NewExampleBase64Binary()
		exp.Binary = []byte("hello")
		exp.Binary2 = []byte("world")
		in := make([]byte, exp.BLength())
		_ = exp.FastWriteNocopy(in, nil)
		resp := http.NewHTTPResponse()
		ctx := context.WithValue(context.Background(), cv.CtxKeyHTTPResponse, resp)
		out, err := conv.Do(ctx, desc, in)
		require.NoError(t, err)

		act := example3.NewExampleBase64Binary()
		err = json.Unmarshal(out, act)
		require.NoError(t, err)
		assert.Equal(t, exp.Binary, act.Binary)
		assert.Equal(t, base64.StdEncoding.EncodeToString(exp.Binary2), resp.Response.Header["Binary2"][0])
	})

	t.Run("no base64 encode", func(t *testing.T) {
		conv := NewBinaryConv(conv.Options{
			EnableHttpMapping: true,
			NoBase64Binary:    true,
		})
		exp := example3.NewExampleBase64Binary()
		exp.Binary = []byte("hello")
		exp.Binary2 = []byte("world")
		in := make([]byte, exp.BLength())
		_ = exp.FastWriteNocopy(in, nil)
		resp := http.NewHTTPResponse()
		ctx := context.WithValue(context.Background(), cv.CtxKeyHTTPResponse, resp)
		out, err := conv.Do(ctx, desc, in)
		require.NoError(t, err)

		act := struct {
			Binary string `json:"Binary"`
		}{}
		err = json.Unmarshal(out, &act)
		require.NoError(t, err)
		assert.Equal(t, string(exp.Binary), act.Binary)
		assert.Equal(t, string(exp.Binary2), resp.Response.Header["Binary2"][0])
	})
}

func TestJSONString(t *testing.T) {
	desc := thrift.FnResponse(thrift.GetFnDescFromFile("testdata/idl/example3.thrift", "JSONStringMethod", thrift.Options{}))
	exp := example3.NewExampleJSONString()
	eobj := &example3.JSONObject{
		A: "1",
		B: 2,
	}
	exp.Header = eobj
	exp.Header2 = map[int32]string{1: "1", 2: "2"}
	exp.Cookie = &example3.JSONObject{}
	exp.Cookie2 = []int32{1, 2}
	in := make([]byte, exp.BLength())
	_ = exp.FastWriteNocopy(in, nil)

	conv := NewBinaryConv(conv.Options{
		EnableHttpMapping: true,
	})
	ctx := context.Background()
	resp := http.NewHTTPResponse()
	ctx = context.WithValue(ctx, cv.CtxKeyHTTPResponse, resp)
	out, err := conv.Do(ctx, desc, in)
	require.NoError(t, err)

	act := example3.NewExampleJSONString()
	require.Equal(t, "{\"Query\":{},\"Query2\":[]}", string(out))
	require.NoError(t, json.Unmarshal([]byte(resp.Response.Header.Get("header")), &act.Header))
	require.Equal(t, exp.Header, act.Header)
	require.NoError(t, json.Unmarshal([]byte(resp.Response.Header.Get("header2")), &act.Header2))
	require.Equal(t, exp.Header2, act.Header2)
	// require.NoError(t, json.Unmarshal([]byte(resp.Cookies()[0].Value), &act.Cookie))
	// require.Equal(t, exp.Cookie, act.Cookie)
	require.NoError(t, json.Unmarshal([]byte(resp.Cookies()[1].Value), &act.Cookie2))
	require.Equal(t, exp.Cookie2, act.Cookie2)
}

func TestDefaultValue(t *testing.T) {
	desc := thrift.FnResponse(thrift.GetFnDescFromFile("testdata/idl/example3.thrift", "DefaultValueMethod", thrift.Options{
		UseDefaultValue: true,
	}))
	in := []byte{byte(thrift.STOP)}
	t.Run("default value", func(t *testing.T) {
		conv := NewBinaryConv(conv.Options{
			EnableHttpMapping: true,
			WriteDefaultField: true,
		})
		resp := http.NewHTTPResponse()
		ctx := context.WithValue(context.Background(), cv.CtxKeyHTTPResponse, resp)
		out, err := conv.Do(ctx, desc, in)
		require.NoError(t, err)
		act := &example3.ExampleDefaultValue{}
		err = json.Unmarshal(out, act)
		require.NoError(t, err)

		exp := example3.NewExampleDefaultValue()
		exp.C = 0
		exp.D = ""
		assert.Equal(t, exp, act)
		require.Equal(t, "1.2", resp.Response.Header.Get("c"))
		require.Equal(t, "const string", resp.Response.Cookies()[0].Value)
	})
	t.Run("zero value", func(t *testing.T) {
		desc := thrift.FnResponse(thrift.GetFnDescFromFile("testdata/idl/example3.thrift", "DefaultValueMethod", thrift.Options{
			UseDefaultValue: false,
		}))
		conv := NewBinaryConv(conv.Options{
			EnableHttpMapping: true,
			WriteDefaultField: true,
		})
		resp := http.NewHTTPResponse()
		ctx := context.WithValue(context.Background(), cv.CtxKeyHTTPResponse, resp)
		out, err := conv.Do(ctx, desc, in)
		require.NoError(t, err)
		act := &example3.ExampleDefaultValue{}
		err = json.Unmarshal(out, act)
		require.NoError(t, err)

		exp := &example3.ExampleDefaultValue{}
		assert.Equal(t, exp, act)
		require.Equal(t, "0", resp.Response.Header.Get("c"))
		require.Equal(t, "", resp.Response.Cookies()[0].Value)
	})
}

func TestOptionalDefaultValue(t *testing.T) {
	in := []byte{byte(thrift.STOP)}
	t.Run("write required", func(t *testing.T) {
		desc := thrift.FnResponse(thrift.GetFnDescFromFile("testdata/idl/example3.thrift", "OptionalDefaultValueMethod", thrift.Options{
			UseDefaultValue: true,
		}))
		conv := NewBinaryConv(conv.Options{
			EnableHttpMapping: true,
			WriteRequireField: true,
		})
		resp := http.NewHTTPResponse()
		ctx := context.WithValue(context.Background(), cv.CtxKeyHTTPResponse, resp)
		out, err := conv.Do(ctx, desc, in)
		require.NoError(t, err)
		act := &example3.ExampleOptionalDefaultValue{}
		err = json.Unmarshal(out, act)
		require.NoError(t, err)

		exp := example3.NewExampleOptionalDefaultValue()
		exp.A = ""
		exp.C = 0
		exp.D = ""
		assert.Equal(t, exp, act)
		require.Equal(t, "const string", resp.Response.Cookies()[0].Value)
	})
	t.Run("write required + write default + write optional", func(t *testing.T) {
		desc := thrift.FnResponse(thrift.GetFnDescFromFile("testdata/idl/example3.thrift", "OptionalDefaultValueMethod", thrift.Options{
			SetOptionalBitmap: true,
			UseDefaultValue:   true,
		}))
		conv := NewBinaryConv(conv.Options{
			EnableHttpMapping:  true,
			WriteRequireField:  true,
			WriteDefaultField:  true,
			WriteOptionalField: true,
		})
		resp := http.NewHTTPResponse()
		ctx := context.WithValue(context.Background(), cv.CtxKeyHTTPResponse, resp)
		out, err := conv.Do(ctx, desc, in)
		require.NoError(t, err)
		act := &example3.ExampleOptionalDefaultValue{}
		err = json.Unmarshal(out, act)
		require.NoError(t, err)

		exp := example3.NewExampleOptionalDefaultValue()
		exp.C = 0
		exp.D = ""
		exp.E = new(string)
		assert.Equal(t, exp, act)
		require.Equal(t, "1.2", resp.Response.Header.Get("c"))
		require.Equal(t, "const string", resp.Response.Cookies()[0].Value)
		require.Equal(t, `""`, resp.Response.Header.Get("f"))
	})
}
