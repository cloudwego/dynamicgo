package j2t

import (
	"bytes"
	"context"
	"encoding/json"
	stdh "net/http"
	"net/url"
	"strings"
	"testing"

	xthrift "github.com/apache/thrift/lib/go/thrift"
	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/conv/j2t"
	"github.com/cloudwego/dynamicgo/http"
	sjson "github.com/cloudwego/dynamicgo/internal/json"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example3"
	"github.com/cloudwego/dynamicgo/testdata/sample"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/cloudwego/dynamicgo/thrift/base"
	"github.com/stretchr/testify/require"
)

func TestPostFormBody(t *testing.T) {
	desc := getExampleDescByName("PostFormMethod", true, thrift.Options{})
	data := url.Values{
		"form":       []string{"b"},
		"JSON":       []string{`{"a":"中文","b":1}`},
		"inner_form": []string{"1"},
	}
	t.Run("fallback", func(t *testing.T) {
		exp := example3.NewExamplePostForm()
		exp.Query = "a"
		exp.Form = "b"
		exp.JSON = &example3.InnerJSON{
			A:         "中文",
			B:         1,
			InnerForm: 0,
		}
		cv := j2t.NewCompactConv(conv.Options{
			WriteDefaultField:            true,
			EnableHttpMapping:            true,
			ReadHttpValueFallback:        true,
			TracebackRequredOrRootFields: true,
		})
		ctx := context.Background()
		sr, err := stdh.NewRequest("POST", "http://localhost?query=a", strings.NewReader(data.Encode()))
		require.NoError(t, err)
		sr.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req, err := http.NewHTTPRequestFromStdReq(sr)
		require.NoError(t, err)
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
		out, err := cv.Do(ctx, desc, []byte(`{}`))
		require.Nil(t, err)
		act := example3.NewExamplePostForm()

		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = exp.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)

		require.Equal(t, exp, act)
	})
	t.Run("no fallback", func(t *testing.T) {
		exp := example3.NewExamplePostForm()
		exp.Query = "a"
		exp.Form = "b"
		// exp.JSON = &example3.InnerJSON{
		// 	A: "中文",
		// 	B: 1,
		// 	InnerForm: 1,
		// }   //NOTICE: not set since conv data is nil, thus no fallback
		cv := j2t.NewCompactConv(conv.Options{
			WriteDefaultField: false,
			EnableHttpMapping: true,
		})
		ctx := context.Background()
		sr, err := stdh.NewRequest("POST", "http://localhost?query=a", strings.NewReader(data.Encode()))
		require.NoError(t, err)
		sr.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req, err := http.NewHTTPRequestFromStdReq(sr)
		require.NoError(t, err)
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
		out, err := cv.Do(ctx, desc, []byte(``))
		require.Nil(t, err)
		act := example3.NewExamplePostForm()

		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = exp.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)

		require.Equal(t, exp, act)
	})
}

func TestHttpMappingFallback(t *testing.T) {
	desc := getExampleFallbackDesc()
	data := []byte(`{"Msg":"hello","Heeader":"world"}`)
	t.Run("fallback", func(t *testing.T) {
		hr, err := stdh.NewRequest("GET", "http://localhost?query=a", nil)
		req := &http.HTTPRequest{
			Request: hr,
		}
		cv := j2t.NewCompactConv(conv.Options{
			EnableHttpMapping:     true,
			ReadHttpValueFallback: true,
		})
		ctx := context.Background()
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
		out, err := cv.Do(ctx, desc, data)
		require.NoError(t, err)
		exp := example3.NewExampleFallback()
		exp.Msg = "hello"
		exp.Heeader = "world"
		act := example3.NewExampleFallback()

		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = exp.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)

		require.Equal(t, exp, act)
	})
	t.Run("not fallback", func(t *testing.T) {
		hr, err := stdh.NewRequest("GET", "http://localhost?A=a", nil)
		hr.Header.Set("heeader", "中文")
		req := &http.HTTPRequest{
			Request: hr,
		}
		cv := j2t.NewCompactConv(conv.Options{
			EnableHttpMapping: true,
		})
		ctx := context.Background()
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
		out, err := cv.Do(ctx, desc, data)
		require.NoError(t, err)
		exp := example3.NewExampleFallback()
		exp.Msg = "a"
		exp.Heeader = "中文"
		act := example3.NewExampleFallback()

		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = exp.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)

		require.Equal(t, exp, act)
	})
}

func TestHttpConvError(t *testing.T) {
	desc := getErrorExampleDesc()
	t.Run("nil required", func(t *testing.T) {
		data := []byte(`{}`)
		hr, err := stdh.NewRequest("GET", "http://localhost", nil)
		require.Nil(t, err)
		req := &http.HTTPRequest{
			Request: hr,
		}
		cv := j2t.NewCompactConv(conv.Options{
			EnableHttpMapping: true,
		})
		ctx := context.Background()
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
		_, err = cv.Do(ctx, desc, data)
		require.Error(t, err)
		require.Equal(t, meta.ErrMissRequiredField, err.(meta.Error).Code.Behavior())
	})

	t.Run("write default", func(t *testing.T) {
		data := []byte(`{}`)
		hr, err := stdh.NewRequest("GET", "http://localhost?query=a", nil)
		require.Nil(t, err)
		hr.Header = stdh.Header{}
		req := &http.HTTPRequest{
			Request: hr,
		}
		cv := j2t.NewBinaryConv(conv.Options{
			EnableHttpMapping: true,
			WriteDefaultField: true,
		})
		ctx := context.Background()
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
		out, err := cv.Do(ctx, desc, data)
		require.NoError(t, err)
		var exp = example3.NewExampleError()
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = exp.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)
		require.Equal(t, "a", exp.Query)
		require.Equal(t, "", exp.Header)
	})

	t.Run("dismathed type", func(t *testing.T) {
		data := []byte(`{}`)
		hr, err := stdh.NewRequest("GET", "http://localhost?query=a&q2=1.5", nil)
		require.Nil(t, err)
		req := &http.HTTPRequest{
			Request: hr,
		}
		cv := j2t.NewCompactConv(conv.Options{
			EnableHttpMapping: true,
		})
		ctx := context.Background()
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
		_, err = cv.Do(ctx, desc, data)
		require.Error(t, err)
		require.Equal(t, meta.ErrConvert, err.(meta.Error).Code.Behavior())
	})
}

func TestJSONString(t *testing.T) {
	desc := getExampleJSONStringDesc()
	data := []byte(``)
	exp := example3.NewExampleJSONString()
	req := getExampleJSONStringReq(exp)
	cv := j2t.NewCompactConv(conv.Options{
		EnableHttpMapping: true,
	})
	ctx := context.Background()
	ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
	out, err := cv.Do(ctx, desc, data)
	require.NoError(t, err)

	act := example3.NewExampleJSONString()
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

func TestThriftRequestBase(t *testing.T) {
	desc := getExampleDescByName("ExampleMethod", true, thrift.Options{
		// NOTICE: must set options.EnableThriftBase to true
		EnableThriftBase: true,
	})
	cv := j2t.NewCompactConv(conv.Options{
		EnableThriftBase:  true,
		WriteDefaultField: true,
	})
	ctx := context.Background()
	b := base.NewBase()
	b.Caller = "caller"
	b.Extra = map[string]string{
		"key": "value",
	}
	b.TrafficEnv = &base.TrafficEnv{
		Env: "env",
	}
	ctx = context.WithValue(ctx, conv.CtxKeyThriftReqBase, b)
	app, err := json.Marshal(b)
	require.NoError(t, err)
	data := getExampleData()

	t.Run("context base", func(t *testing.T) {
		root, _ := sjson.NewSearcher(string(data)).GetByPath()
		_, err := root.Unset("Base")
		require.NoError(t, err)
		str, _ := root.Raw()
		in := []byte(str)
		out, err := cv.Do(ctx, desc, in)
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

		exp := example3.NewExampleReq()
		_, err = root.Set("Base", sjson.NewRaw(string(app)))
		require.NoError(t, err)
		str, _ = root.Raw()
		err = json.Unmarshal([]byte(str), exp)
		require.Nil(t, err)
		require.Equal(t, exp, act)
	})

	// NOTICE: when both body and context base are set, body base will be used
	t.Run("ctx + json base", func(t *testing.T) {
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

		exp := example3.NewExampleReq()
		err = json.Unmarshal(data, exp)
		require.Nil(t, err)
		require.Equal(t, exp, act)
	})
}

func TestCompactEmptyConvHTTP2Thrift(t *testing.T) {
	desc := getExampleDesc()
	data := []byte(``)
	exp := example3.NewExampleReq()
	req := getExampleReq(exp, false, data)
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
	err = exp.Read(pin)
	require.Equal(t, tin.RemainingBytes(), ^uint64(0))
	// _, err = act.FastRead(out)
	// spw.Dump(act)
	require.NoError(t, err)
	require.Equal(t, exp, act)
}

func TestCompactNoBodyStruct(t *testing.T) {
	desc := getExampleDescByName("NoBodyStructMethod", true, thrift.Options{
		UseDefaultValue: true,
	})
	in := []byte(`{}`)
	req, err := http.NewHTTPRequestFromUrl("GET", "http://localhost?b=1", nil)
	require.NoError(t, err)
	cv := j2t.NewCompactConv(conv.Options{
		EnableHttpMapping: true,
	})
	ctx := context.WithValue(context.Background(), conv.CtxKeyHTTPRequest, req)
	out, err := cv.Do(ctx, desc, in)
	require.Nil(t, err)
	act := &example3.ExampleNoBodyStruct{}
	// consume out buffer
	tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
	pin := protoF(tin)
	err = act.Read(pin)
	require.Equal(t, tin.RemainingBytes(), ^uint64(0))
	// _, err = act.FastRead(out)
	// spw.Dump(act)
	require.NoError(t, err)
	exp := example3.NewExampleNoBodyStruct()
	exp.NoBodyStruct = example3.NewNoBodyStruct()
	B := int32(1)
	exp.NoBodyStruct.B = &B
	require.Equal(t, exp, act)
}

func TestAGWDynamicBody(t *testing.T) {
	desc := getExampleDescByName("DynamicStructMethod", true, thrift.Options{})
	exp := example3.NewExampleDynamicStruct()
	exp.Query = "1"
	exp.JSON = "[1,2,3]"
	exp.InnerStruct = &example3.InnerStruct{
		InnerJSON: `{"a":"中文","b":1}`,
		Must:      "2",
	}
	t.Run("no http-mapping", func(t *testing.T) {
		data := `{"Query":"1","json":[1,2,3],"inner_struct":{"inner_json":{"a":"中文","b":1},"Must":"2"}}`
		cv := j2t.NewCompactConv(conv.Options{
			EnableValueMapping:    true,
			EnableHttpMapping:     false,
			WriteRequireField:     true,
			ReadHttpValueFallback: true,
		})
		ctx := context.Background()
		out, err := cv.Do(ctx, desc, []byte(data))
		require.Nil(t, err)
		act := example3.NewExampleDynamicStruct()

		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = exp.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)

		require.Equal(t, exp, act)
	})
	t.Run("http-mapping", func(t *testing.T) {
		data := `{"json":[1,2,3],"inner_struct":{"inner_json":{"a":"中文","b":1}}}`
		cv := j2t.NewCompactConv(conv.Options{
			EnableValueMapping:           true,
			EnableHttpMapping:            true,
			WriteRequireField:            true,
			ReadHttpValueFallback:        true,
			TracebackRequredOrRootFields: true,
		})
		ctx := context.Background()
		req, err := stdh.NewRequest("GET", "http://localhost?query=1&Must=2", nil)
		require.NoError(t, err)
		rr, err := http.NewHTTPRequestFromStdReq(req)
		require.NoError(t, err)
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, rr)
		out, err := cv.Do(ctx, desc, []byte(data))
		require.Nil(t, err)
		act := example3.NewExampleDynamicStruct()

		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := protoF(tin)
		err = exp.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		// spw.Dump(act)
		require.NoError(t, err)

		require.Equal(t, exp, act)
	})
}

func TestCompactBodyFallbackToHttp(t *testing.T) {
	desc := getExampleDesc()
	data := []byte(`{}`)
	hr, err := stdh.NewRequest("POST", "http://localhost:8888/root?Msg=a&Subfix=1", bytes.NewBuffer(nil))
	hr.Header.Set("Base", `{"LogID":"c"}`)
	hr.Header.Set("Path", `b`)
	// NOTICE: optional field will be ignored
	hr.Header.Set("Extra", `{"x":"y"}`)
	require.NoError(t, err)
	req := &http.HTTPRequest{
		Request: hr,
		BodyMap: map[string]string{
			"InnerBase": `{"Bool":true}`,
		},
	}

	t.Run("write default", func(t *testing.T) {
		edata := []byte(`{"Base":{"LogID":"c"},"Subfix":1,"Path":"b","InnerBase":{"Bool":true}}`)
		exp := example3.NewExampleReq()
		exp.InnerBase = sample.GetEmptyInnerBase3()
		exp.RawUri = req.GetUri()
		err = json.Unmarshal(edata, exp)
		require.Nil(t, err)
		cv := j2t.NewCompactConv(conv.Options{
			EnableHttpMapping:            true,
			WriteDefaultField:            true,
			ReadHttpValueFallback:        true,
			TracebackRequredOrRootFields: true,
		})
		ctx := context.Background()
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
		out, err := cv.Do(ctx, desc, data)
		require.NoError(t, err)

		// {
		// 	var b bytes.Buffer
		// 	tout := xthrift.NewStreamTransportW(&b)
		// 	pout := protoF(tout)
		// 	exp.Write(pout)
		// 	pout.Flush(nil)

		// 	fmt.Printf("exp bytes: %+#v\n", b.Bytes())
		// 	fmt.Printf("act bytes: %+#v\n", out)
		// }

		act := example3.NewExampleReq()
		// consume out buffer
		tin := xthrift.NewStreamTransportR(bytes.NewReader(out))
		pin := debugF(tin)
		err = act.Read(pin)
		require.Equal(t, tin.RemainingBytes(), ^uint64(0))
		// _, err = act.FastRead(out)
		require.NoError(t, err)
		require.Equal(t, "a", exp.Query)
		require.Equal(t, "", exp.Header)
	})

	t.Run("not write default", func(t *testing.T) {
		edata := []byte(`{"Base":{"LogID":"c"},"Subfix":1,"Path":"b","InnerBase":{"Bool":true}}`)
		exp := example3.NewExampleReq()
		exp.RawUri = req.GetUri()
		err = json.Unmarshal(edata, exp)
		require.Nil(t, err)
		cv := j2t.NewCompactConv(conv.Options{
			WriteRequireField:            true,
			EnableHttpMapping:            true,
			WriteDefaultField:            false,
			ReadHttpValueFallback:        true,
			TracebackRequredOrRootFields: true,
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
		require.Error(t, err)
		require.Equal(t, meta.ErrConvert, err.(meta.Error).Code.Behavior())
	})
}
