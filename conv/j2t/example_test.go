package j2t

import (
	"context"
	"encoding/json"
	stdhttp "net/http"
	"reflect"
	"strings"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example3"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/davecgh/go-spew/spew"
)

var opts = conv.Options{
	EnableValueMapping: true,
	EnableHttpMapping:  true,
}

func ExampleBinaryConv_Do() {
	// get descriptor and data
	desc := getExampleDesc()
	data := getExampleData()

	// make BinaryConv
	cv := NewBinaryConv(opts)

	// do conversion
	out, err := cv.Do(context.Background(), desc, data)
	if err != nil {
		panic(err)
	}

	// validate result
	exp := example3.NewExampleReq()
	err = json.Unmarshal(data, exp)
	if err != nil {
		panic(err)
	}
	act := example3.NewExampleReq()
	_, err = act.FastRead(out)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(exp, act) {
		panic("not equal")
	}
}

func ExampleHTTPConv_DoInto() {
	// get function descriptor
	svc, err := thrift.NewDescritorFromPath(context.Background(), exampleIDLPath)
	if err != nil {
		panic(err)
	}
	fn := svc.Functions()["ExampleMethod"]

	// make HTTPConv
	cv := NewHTTPConv(meta.EncodingThriftBinary, fn)

	// make std http request
	jdata := `{"msg":"hello","InnerBase":{}}`
	stdreq, err := stdhttp.NewRequest("POST",
		"http://localhost:8080/example?query=1,2,3&inner_query=中文",
		strings.NewReader(jdata))
	if err != nil {
		panic(err)
	}
	stdreq.Header.Set("Content-Type", "application/json")
	stdreq.Header.Set("heeader", "true")
	stdreq.Header.Set("inner_string", "igorned")

	// wrap std http request as http.RequestGetter
	req, err := http.NewHTTPRequestFromStdReq(
		stdreq,
		http.Param{Key: "path", Value: "OK"},
		http.Param{Key: "inner_string", Value: "priority"},
	)

	// allocate output buffer
	// TIPS: json data size is usually 1.5x times of thrift data size
	buf := make([]byte, 0, len(jdata)*2/3)

	// do conversion
	err = cv.DoInto(context.Background(), req, &buf, opts)
	if err != nil {
		panic(err)
	}

	// unwrap thrift message
	p := thrift.NewBinaryProtocol(buf)
	method, mType, seqID, reqID, stru, err := p.UnwrapBody()
	println(method, mType, seqID, reqID) // ExampleMethod 1 0 1

	// validate result
	act := example3.NewExampleReq()
	_, err = act.FastRead(stru)
	if err != nil {
		panic(err)
	}
	spew.Dump(act)
}
