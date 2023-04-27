package j2t

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	stdh "net/http"
	"net/url"
	"strconv"
	"strings"

	xthrift "github.com/apache/thrift/lib/go/thrift"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example3"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/davecgh/go-spew/spew"
	thrift_iter "github.com/thrift-iterator/go"
)

const (
	// exampleIDLPath = "../../testdata/idl/example3.thrift"
	// cur
	basePath       = "../../../"
	exampleIDLPath = basePath + "testdata/idl/example3.thrift"
	nullIDLPath    = basePath + "testdata/idl/null.thrift"
	exampleJSON    = basePath + "testdata/data/example3req.json"
	nullJSON       = basePath + "testdata/data/null_pass.json"
	nullerrJSON    = basePath + "testdata/data/null_err.json"
)

var (
	spw = &spew.ConfigState{
		Indent:                  "\t",
		MaxDepth:                0,
		DisableMethods:          true,
		DisablePointerMethods:   false,
		DisablePointerAddresses: false,
		DisableCapacities:       false,
		ContinueOnMethod:        false,
		SortKeys:                false,
		SpewKeys:                false,
	}
	protoF    = xthrift.NewTCompactProtocol
	debugF    = xthrift.NewTDebugProtocolFactory(xthrift.NewTCompactProtocolFactory(), "dbg ").GetProtocol
	tiCompact = thrift_iter.Config{Protocol: thrift_iter.ProtocolCompact, StaticCodegen: false}.Froze()
)

func getExampleDesc() *thrift.TypeDescriptor {
	opts := thrift.Options{}
	svc, err := opts.NewDescritorFromPath(context.Background(), exampleIDLPath)
	if err != nil {
		panic(err)
	}
	return svc.Functions()["ExampleMethod"].Request().Struct().FieldById(1).Type()
}

func getErrorExampleDesc() *thrift.TypeDescriptor {
	opts := thrift.Options{}
	svc, err := opts.NewDescritorFromPath(context.Background(), exampleIDLPath)
	if err != nil {
		panic(err)
	}
	return svc.Functions()["ErrorMethod"].Request().Struct().FieldById(1).Type()
}

func getExampleInt2FloatDesc() *thrift.TypeDescriptor {
	opts := thrift.Options{}
	svc, err := opts.NewDescritorFromPath(context.Background(), exampleIDLPath)
	if err != nil {
		panic(err)
	}
	return svc.Functions()["Int2FloatMethod"].Request().Struct().FieldById(1).Type()
}

func getExampleJSONStringDesc() *thrift.TypeDescriptor {
	opts := thrift.Options{}
	svc, err := opts.NewDescritorFromPath(context.Background(), exampleIDLPath)
	if err != nil {
		panic(err)
	}
	return svc.Functions()["JSONStringMethod"].Request().Struct().FieldById(1).Type()
}

func getExampleFallbackDesc() *thrift.TypeDescriptor {
	opts := thrift.Options{}
	svc, err := opts.NewDescritorFromPath(context.Background(), exampleIDLPath)
	if err != nil {
		panic(err)
	}
	return svc.Functions()["FallbackMethod"].Request().Struct().FieldById(1).Type()
}

func getExampleDescByName(method string, req bool, opts thrift.Options) *thrift.TypeDescriptor {
	svc, err := opts.NewDescritorFromPath(context.Background(), exampleIDLPath)
	if err != nil {
		panic(err)
	}
	if req {
		return svc.Functions()[method].Request().Struct().Fields()[0].Type()
	} else {
		return svc.Functions()[method].Response().Struct().Fields()[0].Type()
	}
}

func getExampleData() []byte {
	out, err := ioutil.ReadFile(exampleJSON)
	if err != nil {
		panic(err)
	}
	return out
}

func getExampleReq(exp *example3.ExampleReq, setIs bool, body []byte) *http.HTTPRequest {
	f := -1.00001
	x := true
	q := []string{"1", "2", "3"}
	p := "<>"
	is := "abcd"
	uv := url.Values{
		"query": []string{strings.Join(q, ",")},
	}
	if setIs {
		uv.Add("inner_query", is)
		exp.InnerBase.InnerQuery = is
		exp.InnerBase.ListInnerBase[0].InnerQuery = is
		exp.InnerBase.MapStringInnerBase["innerx"].InnerQuery = is
	}

	uri := "http://localhost:8888/root?" + uv.Encode()
	hr, err := stdh.NewRequest("POST", uri, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	hr.Header.Set("Content-Type", "application/json")
	req, err := http.NewHTTPRequestFromStdReq(hr)
	if err != nil {
		panic(err)
	}
	req.Params.Set("path", p)
	req.Request.Header.Add("heeader", strconv.FormatBool(x))
	req.AddCookie(&stdh.Cookie{Name: "cookie", Value: strconv.FormatFloat(f, 'f', -1, 64)})
	if setIs {
		req.Request.Header.Add("inner_string", is)
		exp.InnerBase.ListInnerBase[0].String_ = is
		exp.InnerBase.MapStringInnerBase["innerx"].String_ = is
		exp.InnerBase.String_ = is
	}
	exp.Path = p
	exp.Query = q
	exp.Header = &x
	exp.Cookie = &f
	exp.RawUri = uri
	return req
}

func getExampleJSONStringReq(exp *example3.ExampleJSONString) *http.HTTPRequest {
	j := `{"a":"1","b":2}`
	x := "{}"
	a := `["1","2","3"]`
	b := `[1,2,3]`
	c := `{"1":"1","2":"2","3":"3"}`

	qs := url.Values{}
	qs.Add("query", j)
	qs.Add("query2", a)
	hr, err := stdh.NewRequest("POST", "http://localhost:8888/root?"+qs.Encode(), bytes.NewBuffer(nil))
	if err != nil {
		panic(err)
	}
	req := &http.HTTPRequest{
		Request: hr,
	}
	req.AddCookie(&stdh.Cookie{Name: "cookie", Value: x})
	req.AddCookie(&stdh.Cookie{Name: "cookie2", Value: b})
	req.Request.Header.Set("header", j)
	req.Request.Header.Set("header2", c)

	_ = json.Unmarshal([]byte(j), &exp.Query)
	_ = json.Unmarshal([]byte(a), &exp.Query2)
	_ = json.Unmarshal([]byte(x), &exp.Cookie)
	_ = json.Unmarshal([]byte(b), &exp.Cookie2)
	_ = json.Unmarshal([]byte(j), &exp.Header)
	_ = json.Unmarshal([]byte(c), &exp.Header2)
	return req
}

func getNullDesc() *thrift.TypeDescriptor {
	opts := thrift.Options{}
	svc, err := opts.NewDescritorFromPath(context.Background(), nullIDLPath)
	if err != nil {
		panic(err)
	}
	return svc.Functions()["NullTest"].Request().Struct().FieldById(1).Type()
}

func getNullData() []byte {
	out, err := ioutil.ReadFile(nullJSON)
	if err != nil {
		panic(err)
	}
	return out
}

func getNullErrData() []byte {
	out, err := ioutil.ReadFile(nullerrJSON)
	if err != nil {
		panic(err)
	}
	return out
}
