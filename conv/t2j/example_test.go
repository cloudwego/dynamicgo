package t2j

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example3"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/davecgh/go-spew/spew"
)

var opts = conv.Options{
	EnableHttpMapping: true,
}

func ExampleBinaryConv_Do() {
	// get descriptor and data
	desc := thrift.FnResponse(thrift.GetFnDescFromFile("testdata/idl/example3.thrift", "ExampleMethod", thrift.Options{}))
	data := getExample3Data()
	// make BinaryConv
	cv := NewBinaryConv(opts)

	// do conversion
	out, err := cv.Do(context.Background(), desc, data)
	if err != nil {
		panic(err)
	}

	// validate result
	var exp, act example3.ExampleResp
	_, err = exp.FastRead(data)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(out, &act)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(exp, act) {
		panic("not equal")
	}
}

func ExampleHTTPConv_DoInto() {
	// get function descriptor
	desc := thrift.GetFnDescFromFile("testdata/idl/example3.thrift", "ExampleMethod", thrift.Options{})

	// make thrift message
	data := getExample3Data()
	in, err := thrift.WrapBinaryBody(data, "ExampleMethod", thrift.REPLY, thrift.FieldID(0), 1)
	if err != nil {
		panic(err)
	}

	// get http.ResponseSetter
	resp := http.NewHTTPResponse()
	resp.StatusCode = 200

	// make HTTPConv
	cv := NewHTTPConv(meta.EncodingThriftBinary, desc)

	// do conversion
	buf := make([]byte, 0, len(data)*2)
	err = cv.DoInto(context.Background(), resp, in, &buf, opts)
	if err != nil {
		panic(err)
	}

	// validate result
	var act example3.ExampleResp
	err = json.Unmarshal(buf, &act)
	if err != nil {
		panic(err)
	}
	// non-http annotations fields
	spew.Dump(act)
	// http annotations fields
	spew.Dump(resp)
}
