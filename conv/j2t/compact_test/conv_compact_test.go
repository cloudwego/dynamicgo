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

package j2t

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	stdh "net/http"
	"strconv"
	"strings"
	"testing"

	xthrift "github.com/apache/thrift/lib/go/thrift"
	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/conv/j2t"
	"github.com/cloudwego/dynamicgo/http"
	sjson "github.com/cloudwego/dynamicgo/internal/json"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example3"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/null"
	"github.com/cloudwego/dynamicgo/testdata/sample"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/stretchr/testify/require"
)

func TestCompactConvJSON2Thrift(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()
	cv := j2t.NewCompactConv(conv.Options{})
	ctx := context.Background()
	out, err := cv.Do(ctx, desc, data)
	require.Nil(t, err)
	exp := example3.NewExampleReq()
	err = json.Unmarshal(data, exp)
	require.Nil(t, err)

	act := example3.NewExampleReq()
	// consume out buffer
	tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
	pin := protoF(tin)
	err = act.Read(pin)
	require.Equal(t, tin.RemainingBytes(), ^uint64(0))
	// _, err = act.FastRead(out)

	// spw.Dump(act)
	require.NoError(t, err)
	require.Equal(t, exp, act)
}

func TestCompactConvHTTP2Thrift(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()
	exp := example3.NewExampleReq()
	err := json.Unmarshal(data, exp)
	require.Nil(t, err)
	req := getExampleReq(exp, true, data)
	cv := j2t.NewCompactConv(conv.Options{
		EnableHttpMapping: true,
	})
	ctx := context.Background()
	ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
	out, err := cv.Do(ctx, desc, data)
	require.NoError(t, err)

	act := example3.NewExampleReq()

	// consume out buffer
	tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
	pin := protoF(tin)
	err = act.Read(pin)
	require.Equal(t, tin.RemainingBytes(), ^uint64(0))
	// _, err = act.FastRead(out)

	// spw.Dump(act)
	require.NoError(t, err)
	require.Equal(t, exp, act)
}

func TestCompactWriteDefault(t *testing.T) {
	desc := getExampleDesc()
	data := []byte(`{"Path":"<>"}`)
	exp := example3.NewExampleReq()
	data2 := []byte(`{"Path":"<>","Base":{}}`)
	exp.InnerBase = sample.GetEmptyInnerBase3()
	err := json.Unmarshal(data2, exp)
	require.Nil(t, err)
	req := getExampleReq(exp, false, data)

	cv := j2t.NewCompactConv(conv.Options{
		WriteDefaultField: true,
		EnableHttpMapping: true,
	})

	ctx := context.Background()
	ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
	out, err := cv.Do(ctx, desc, data)

	fmt.Printf("%+#v\n", out)
	require.Nil(t, err)
	act := example3.NewExampleReq()

	// consume out buffer
	tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
	pin := protoF(tin)
	err = act.Read(pin)
	require.Equal(t, tin.RemainingBytes(), ^uint64(0))
	// _, err = act.FastRead(out)

	// spw.Dump(act)
	require.NoError(t, err)
	require.Equal(t, exp, act)

	var oout bytes.Buffer
	pout := protoF(xthrift.NewStreamTransportW(&oout))
	err = act.Write(pout)
	require.NoError(t, err)
	pout.Flush(nil)
	fmt.Printf("encode_athrift: %+#v\nencode_nthrift: %+#v\n",
		oout.Bytes(), out)
}

func TestCompactWriteRequired(t *testing.T) {
	desc := getExampleDesc()
	data := []byte(`{}`)
	t.Run("JSON", func(t *testing.T) {
		exp := example3.NewExampleReq()
		data2 := []byte(`{"Path":""}`)
		err := json.Unmarshal(data2, exp)
		require.Nil(t, err)
		cv := j2t.NewCompactConv(conv.Options{
			WriteRequireField: true,
		})
		ctx := context.Background()
		out, err := cv.Do(ctx, desc, data)
		require.Nil(t, err)

		act := example3.NewExampleReq()
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = act.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)
		require.Equal(t, exp, act)

	})
	t.Run("http-mapping", func(t *testing.T) {
		exp := example3.NewExampleReq()
		exp.InnerBase = sample.GetEmptyInnerBase3()
		data2 := []byte(`{"Path":"","Base":{}}`)
		err := json.Unmarshal(data2, exp)
		require.Nil(t, err)
		hr, err := stdh.NewRequest("POST", "http://localhost:8888/root", bytes.NewBuffer(data))
		require.NoError(t, err)
		exp.RawUri = hr.URL.String()
		req, err := http.NewHTTPRequestFromStdReq(hr)
		require.NoError(t, err)
		cv := j2t.NewCompactConv(conv.Options{
			WriteRequireField: true,
			WriteDefaultField: true,
			EnableHttpMapping: true,
		})
		ctx := context.Background()
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
		out, err := cv.Do(ctx, desc, data)
		require.Nil(t, err)

		act := example3.NewExampleReq()
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = act.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		spw.Dump(act)
		require.NoError(t, err)
		require.Equal(t, exp, act)
	})
}

func TestCompactRequireness(t *testing.T) {
	desc := getErrorExampleDesc()
	data := []byte(`{}`)
	cv := j2t.NewCompactConv(conv.Options{
		EnableHttpMapping: true,
	})
	ctx := context.Background()
	req := http.NewHTTPRequest()
	req.Request, _ = stdh.NewRequest("GET", "root?query=abc", bytes.NewBuffer(nil))
	ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
	out, err := cv.Do(ctx, desc, data)
	require.Nil(t, err)
	act := example3.NewExampleError()
	// consume out buffer
	tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
	pin := protoF(tin)
	err = act.Read(pin)
	require.Equal(t, tin.RemainingBytes(), ^uint64(0))
	// _, err = act.FastRead(out)
	spw.Dump(act)
	require.NoError(t, err)
	require.Equal(t, req.URL.Query().Get("query"), act.Query)
}

func TestCompactNullJSON2Thrift(t *testing.T) {
	desc := getNullDesc()
	data := getNullData()
	cv := j2t.NewCompactConv(conv.Options{})
	ctx := context.Background()
	out, err := cv.Do(ctx, desc, data)
	require.Nil(t, err)
	exp := null.NewNullStruct()
	err = json.Unmarshal(data, exp)
	require.Nil(t, err)

	var m = map[string]interface{}{}
	err = json.Unmarshal(data, &m)
	require.Nil(t, err)
	spw.Dump(m)

	act := null.NewNullStruct()
	// consume out buffer
	tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
	pin := protoF(tin)
	err = act.Read(pin)
	require.Equal(t, tin.RemainingBytes(), ^uint64(0))
	// _, err = act.FastRead(out)
	spw.Dump(act)
	require.NoError(t, err)
	// require.Equal(t, exp, act)
}

func TestCompactApiBody(t *testing.T) {
	desc := getExampleDescByName("ApiBodyMethod", true, thrift.Options{})
	data := []byte(`{"code":1024,"InnerCode":{}}`)
	cv := j2t.NewCompactConv(conv.Options{
		EnableHttpMapping:            true,
		WriteDefaultField:            true,
		ReadHttpValueFallback:        true,
		TracebackRequredOrRootFields: true,
	})
	ctx := context.Background()
	req, err := stdh.NewRequest("POST", "http://localhost:8888/root", bytes.NewBuffer(data))
	require.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	r, err := http.NewHTTPRequestFromStdReq(req)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, r)
	out, err := cv.Do(ctx, desc, data)
	require.Nil(t, err)
	act := example3.NewExampleApiBody()

	// consume out buffer
	tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
	pin := protoF(tin)
	err = act.Read(pin)
	require.Equal(t, tin.RemainingBytes(), ^uint64(0))
	// _, err = act.FastRead(out)
	// spw.Dump(act)
	require.NoError(t, err)

	require.Equal(t, int64(1024), act.Code)
	require.Equal(t, int16(1024), act.Code2)
	require.Equal(t, int64(1024), act.InnerCode.C1)
	require.Equal(t, int16(0), act.InnerCode.C2)
}

func TestCompactError(t *testing.T) {
	desc := getExampleDesc()

	t.Run("ERR_NULL_REQUIRED", func(t *testing.T) {
		desc := getNullDesc()
		data := getNullErrData()
		cv := j2t.NewCompactConv(conv.Options{})
		ctx := context.Background()
		_, err := cv.Do(ctx, desc, data)
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrMissRequiredField, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "missing required field 3"))
		// require.True(t, strings.Contains(msg, "near "+strconv.Itoa(strings.Index(string(data), `"Null3": null`)+13)))
	})

	t.Run("INVALID_CHAR", func(t *testing.T) {
		data := `{xx}`
		cv := j2t.NewCompactConv(conv.Options{})
		ctx := context.Background()
		_, err := cv.Do(ctx, desc, []byte(data))
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrRead, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "invalid char 'x' for state J2T_OBJ_0"))
		// require.True(t, strings.Contains(msg, "near 2"))
	})

	t.Run("ERR_INVALID_NUMBER_FMT", func(t *testing.T) {
		desc := getExampleInt2FloatDesc()
		data := []byte(`{"Float64":1.x1}`)
		cv := j2t.NewCompactConv(conv.Options{
			EnableValueMapping: true,
		})
		ctx := context.Background()
		_, err := cv.Do(ctx, desc, data)
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrConvert, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "unexpected number type"))
		// require.True(t, strings.Contains(msg, "near 41"))
	})

	t.Run("ERR_UNSUPPORT_THRIFT_TYPE", func(t *testing.T) {
		desc := getErrorExampleDesc()
		data := []byte(`{"MapInnerBaseInnerBase":{"a":"a"}`)
		cv := j2t.NewCompactConv(conv.Options{})
		ctx := context.Background()
		_, err := cv.Do(ctx, desc, data)
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrUnsupportedType, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "unsupported thrift type STRUCT"))
		// require.True(t, strings.Contains(msg, "near 32"))
	})

	t.Run("ERR_DISMATCH_TYPE", func(t *testing.T) {
		data := getExampleData()
		n, err := sjson.NewSearcher(string(data)).GetByPath()
		require.Nil(t, err)
		exist, err := n.Set("code_code", sjson.NewString("1.1"))
		require.True(t, exist)
		require.Nil(t, err)
		data, err = n.MarshalJSON()
		require.Nil(t, err)
		cv := j2t.NewCompactConv(conv.Options{})
		ctx := context.Background()
		_, err = cv.Do(ctx, desc, data)
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrDismatchType, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "expect type I64 but got type 11"))
	})

	t.Run("ERR_UNKNOWN_FIELD", func(t *testing.T) {
		desc := getErrorExampleDesc()
		data := []byte(`{"UnknownField":"1"}`)
		cv := j2t.NewCompactConv(conv.Options{
			DisallowUnknownField: true,
		})
		ctx := context.Background()
		_, err := cv.Do(ctx, desc, data)
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrUnknownField, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "unknown field 'UnknownField'"))
	})

	t.Run("ERR_UNKNOWN_FIELD", func(t *testing.T) {
		desc := getErrorExampleDesc()
		data := []byte(`{"UnknownField":"1"}`)
		cv := j2t.NewCompactConv(conv.Options{
			DisallowUnknownField: true,
		})
		ctx := context.Background()
		_, err := cv.Do(ctx, desc, data)
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrUnknownField, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "unknown field 'UnknownField'"))
	})

	t.Run("ERR_DECODE_BASE64", func(t *testing.T) {
		desc := getErrorExampleDesc()
		data := []byte(`{"Base64":"xxx"}`)
		cv := j2t.NewCompactConv(conv.Options{})
		ctx := context.Background()
		_, err := cv.Do(ctx, desc, data)
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrRead, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "decode base64 error: "))
	})

	t.Run("ERR_RECURSE_EXCEED_MAX", func(t *testing.T) {
		// desc := getExampleInt2FloatDesc()
		// src := []byte(`{}`)
		// cv := j2t.NewCompactConv(conv.Options{})
		// ctx := context.Background()
		// buf := make([]byte, 0, 1)
		// mock := MockConv{
		// 	sp:        types.MAX_RECURSE + 1,
		// 	reqsCache: 1,
		// 	keyCache:  1,
		// 	dcap:      800,
		// }
		// TODO: IMPLEMENT THIS!
		// err := mock.do(&cv, ctx, src, desc, &buf, nil, true)
		// require.Error(t, err)
		// msg := err.Error()
		// require.Equal(t, meta.ErrStackOverflow, err.(meta.Error).Code.Behavior())
		// require.True(t, strings.Contains(msg, "stack "+strconv.Itoa(types.MAX_RECURSE+1)+" overflow"))
	})
}

func TestFloat2Int(t *testing.T) {
	t.Run("double2int", func(t *testing.T) {
		desc := getExampleInt2FloatDesc()
		data := []byte(`{"Int32":2.229e+2}`)
		cv := j2t.NewCompactConv(conv.Options{})
		ctx := context.Background()
		out, err := cv.Do(ctx, desc, data)
		require.NoError(t, err)
		exp := example3.NewExampleInt2Float()
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = exp.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)
		require.Equal(t, exp.Int32, int32(222))
	})
	t.Run("int2double", func(t *testing.T) {
		desc := getExampleInt2FloatDesc()
		data := []byte(`{"Float64":` + strconv.Itoa(math.MaxInt64) + `}`)
		cv := j2t.NewCompactConv(conv.Options{})
		ctx := context.Background()
		out, err := cv.Do(ctx, desc, data)
		require.NoError(t, err)
		exp := example3.NewExampleInt2Float()
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = exp.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)
		require.Equal(t, exp.Float64, float64(math.MaxInt64))
	})
}

// TODO: IMPL TestStateMachineOOM

func TestCompactString2Int(t *testing.T) {
	desc := getExampleInt2FloatDesc()
	cv := j2t.NewCompactConv(conv.Options{
		String2Int64: true,
	})
	t.Run("converting", func(t *testing.T) {
		data := []byte(`{"Int32":"", "Float64":"1.1", "中文": 123.3}`)
		cv.SetOptions(conv.Options{
			EnableValueMapping: true,
		})
		ctx := context.Background()
		out, err := cv.Do(ctx, desc, data)
		require.Nil(t, err)
		exp := example3.NewExampleInt2Float()
		exp.Int32 = 0
		exp.Float64 = 1.1
		exp.String_ = "123.3"
		act := example3.NewExampleInt2Float()
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = act.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)
		require.Equal(t, exp, act)
	})

	t.Run("no-converting", func(t *testing.T) {
		data := []byte(`{"Int32":222, "Float64":1.1, "中文": "123.3"}`)
		cv.SetOptions(conv.Options{
			EnableValueMapping: true,
		})
		ctx := context.Background()
		out, err := cv.Do(ctx, desc, data)
		require.Nil(t, err)
		exp := example3.NewExampleInt2Float()
		exp.Int32 = 222
		exp.Float64 = 1.1
		exp.String_ = "123.3"
		act := example3.NewExampleInt2Float()
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = act.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)
		require.Equal(t, exp, act)
	})

	t.Run("option Int64AsString", func(t *testing.T) {
		data := []byte(`{"Int32":"222","Int64":"333"}`)
		cv.SetOptions(conv.Options{
			String2Int64: true,
		})
		ctx := context.Background()
		out, err := cv.Do(ctx, desc, data)
		require.Nil(t, err)
		exp := example3.NewExampleInt2Float()
		exp.Int64 = 333
		exp.Int32 = 222
		act := example3.NewExampleInt2Float()
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = act.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)
		require.Equal(t, exp, act)
	})
}

func TestNobodyRequiredFields(t *testing.T) {
	path := "a/b/main.thrift"
	content := `
	namespace go kitex.test.server
	struct Base {
		1: required string required_field
	}

	service InboxService {
		string ExampleMethod(1: Base req)
	}
	`
	includes := map[string]string{
		path: content,
	}
	p, err := thrift.NewDescritorFromContent(context.Background(), path, content, includes, true)
	if err != nil {
		t.Fatal(err)
	}
	desc := p.Functions()["ExampleMethod"].Request().Struct().Fields()[0].Type()
	cv := j2t.NewCompactConv(conv.Options{
		EnableHttpMapping:     true,
		WriteRequireField:     true,
		ReadHttpValueFallback: true,
	})
	ctx := context.Background()
	req, err := http.NewHTTPRequestFromUrl("GET", "http://localhost?required_field=1", nil)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
	out, err := cv.Do(ctx, desc, nil)
	require.Nil(t, err)
	fmt.Printf("%+v", out)
}

func TestBase64Decode(t *testing.T) {
	desc := getExampleDescByName("Base64BinaryMethod", true, thrift.Options{})
	t.Run("base64 decode", func(t *testing.T) {
		exp := example3.NewExampleBase64Binary()
		exp.Binary = []byte("hello")
		in, err := json.Marshal(exp)
		require.Nil(t, err)
		cv := j2t.NewCompactConv(conv.Options{
			EnableHttpMapping: true,
			NoBase64Binary:    false,
		})
		req, err := http.NewHTTPRequestFromUrl("GET", "http://localhost", nil)
		require.NoError(t, err)
		req.Request.Header.Set("Binary2", base64.StdEncoding.EncodeToString([]byte("world")))
		ctx := context.WithValue(context.Background(), conv.CtxKeyHTTPRequest, req)
		out, err := cv.Do(ctx, desc, in)
		require.Nil(t, err)

		act := example3.NewExampleBase64Binary()
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = act.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)
		require.Equal(t, []byte("hello"), act.Binary)
		require.Equal(t, []byte("world"), act.Binary2)
	})

	t.Run("no base64 decode", func(t *testing.T) {
		in := []byte(`{"Binary":"hello"}`)
		cv := j2t.NewCompactConv(conv.Options{
			EnableHttpMapping: true,
			NoBase64Binary:    true,
		})
		req, err := http.NewHTTPRequestFromUrl("GET", "http://localhost", nil)
		require.NoError(t, err)
		req.Request.Header.Set("Binary2", "world")
		ctx := context.WithValue(context.Background(), conv.CtxKeyHTTPRequest, req)
		out, err := cv.Do(ctx, desc, in)
		require.Nil(t, err)

		act := example3.NewExampleBase64Binary()
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = act.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)
		require.Equal(t, []byte("hello"), act.Binary)
		require.Equal(t, []byte("world"), act.Binary2)
	})
}

func TestDefaultValue(t *testing.T) {
	desc := getExampleDescByName("DefaultValueMethod", true, thrift.Options{
		UseDefaultValue: true,
	})
	in := []byte(`{}`)
	t.Run("default value", func(t *testing.T) {
		cv := j2t.NewCompactConv(conv.Options{
			WriteDefaultField: true,
			EnableHttpMapping: false,
		})
		out, err := cv.Do(context.Background(), desc, in)
		require.Nil(t, err)
		act := &example3.ExampleDefaultValue{}
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = act.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)
		exp := example3.NewExampleDefaultValue()
		require.Equal(t, exp, act)
	})
	t.Run("default value + http mapping", func(t *testing.T) {
		cv := j2t.NewCompactConv(conv.Options{
			WriteDefaultField: true,
			EnableHttpMapping: true,
		})
		req, err := http.NewHTTPRequestFromUrl("GET", "http://localhost", nil)
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), conv.CtxKeyHTTPRequest, req)
		out, err := cv.Do(ctx, desc, in)
		require.Nil(t, err)
		act := &example3.ExampleDefaultValue{}
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = act.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)
		exp := example3.NewExampleDefaultValue()
		require.Equal(t, exp, act)
	})
	t.Run("zero value", func(t *testing.T) {
		desc := getExampleDescByName("DefaultValueMethod", true, thrift.Options{
			UseDefaultValue: false,
		})
		cv := j2t.NewCompactConv(conv.Options{
			WriteDefaultField: true,
			EnableHttpMapping: false,
		})
		req, err := http.NewHTTPRequestFromUrl("GET", "http://localhost", nil)
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), conv.CtxKeyHTTPRequest, req)
		out, err := cv.Do(ctx, desc, in)
		require.Nil(t, err)
		act := &example3.ExampleDefaultValue{}
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = act.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)

		exp := &example3.ExampleDefaultValue{}
		require.Equal(t, exp, act)
	})
}

func TestOptionalDefaultValue(t *testing.T) {
	desc := getExampleDescByName("OptionalDefaultValueMethod", true, thrift.Options{
		SetOptionalBitmap: true,
		UseDefaultValue:   true,
	})
	in := []byte(`{}`)
	t.Run("write default + write optional", func(t *testing.T) {
		cv := j2t.NewCompactConv(conv.Options{
			WriteRequireField:  true,
			WriteDefaultField:  true,
			EnableHttpMapping:  false,
			WriteOptionalField: true,
		})
		out, err := cv.Do(context.Background(), desc, in)
		require.Nil(t, err)
		act := &example3.ExampleOptionalDefaultValue{}
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = act.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)
		exp := example3.NewExampleOptionalDefaultValue()
		exp.E = new(string)
		exp.F = new(string)
		require.Equal(t, exp, act)
	})
	t.Run("not write optional", func(t *testing.T) {
		cv := j2t.NewCompactConv(conv.Options{
			WriteRequireField: true,
			WriteDefaultField: false,
			EnableHttpMapping: false,
		})
		out, err := cv.Do(context.Background(), desc, in)
		require.Nil(t, err)
		act := &example3.ExampleOptionalDefaultValue{}
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = act.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)
		exp := example3.NewExampleOptionalDefaultValue()
		exp.A = ""
		exp.C = 0
		require.Equal(t, exp, act)
	})
	t.Run("write default", func(t *testing.T) {
		cv := j2t.NewCompactConv(conv.Options{
			WriteRequireField: false,
			WriteDefaultField: true,
			EnableHttpMapping: false,
		})
		_, err := cv.Do(context.Background(), desc, in)
		require.Error(t, err)
	})
	t.Run("write default + http mapping", func(t *testing.T) {
		in := []byte(`{"B":1}`)
		cv := j2t.NewCompactConv(conv.Options{
			WriteRequireField: false,
			WriteDefaultField: true,
			EnableHttpMapping: true,
		})
		req, err := http.NewHTTPRequestFromUrl("GET", "http://localhost", nil)
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), conv.CtxKeyHTTPRequest, req)
		_, err = cv.Do(ctx, desc, in)
		require.Error(t, err)
	})
	t.Run("write default + write optional + http mapping", func(t *testing.T) {
		cv := j2t.NewCompactConv(conv.Options{
			WriteRequireField:  true,
			WriteDefaultField:  true,
			EnableHttpMapping:  true,
			WriteOptionalField: true,
		})
		req, err := http.NewHTTPRequestFromUrl("GET", "http://localhost", nil)
		require.NoError(t, err)
		ctx := context.WithValue(context.Background(), conv.CtxKeyHTTPRequest, req)
		out, err := cv.Do(ctx, desc, in)
		require.Nil(t, err)
		act := &example3.ExampleOptionalDefaultValue{}
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = act.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)
		exp := example3.NewExampleOptionalDefaultValue()
		exp.E = new(string)
		exp.F = new(string)
		require.Equal(t, exp, act)
	})
}

func TestCompactSimpleArgs(t *testing.T) {
	cv := j2t.NewCompactConv(conv.Options{})

	t.Run("string", func(t *testing.T) {
		desc := getExampleDescByName("String", true, thrift.Options{})
		p := thrift.NewCompactProtocolBuffer()
		p.WriteString("hello")
		exp := p.Buf
		out, err := cv.Do(context.Background(), desc, []byte(`"hello"`))
		require.NoError(t, err)
		require.Equal(t, exp, out)
	})

	t.Run("no quoted string", func(t *testing.T) {
		desc := getExampleDescByName("String", true, thrift.Options{})
		p := thrift.NewCompactProtocolBuffer()
		p.WriteString(`hel\lo`)
		exp := p.Buf
		out, err := cv.Do(context.Background(), desc, []byte(`hel\lo`))
		require.NoError(t, err)
		require.Equal(t, exp, out)
	})

	t.Run("int", func(t *testing.T) {
		desc := getExampleDescByName("I64", true, thrift.Options{})
		p := thrift.NewCompactProtocolBuffer()
		p.WriteI64(math.MaxInt64)
		exp := p.Buf
		out, err := cv.Do(context.Background(), desc, []byte(strconv.Itoa(math.MaxInt64)))
		require.NoError(t, err)
		require.Equal(t, exp, out)
	})
}
