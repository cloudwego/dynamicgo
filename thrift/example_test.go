package thrift

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/cloudwego/dynamicgo/meta"
)

func ExampleNewDescritorFromPath() {
	// default Options
	p1, err := NewDescriptorFromPath(context.Background(), "../testdata/idl/example.thrift")
	if err != nil {
		panic(err)
	}
	r1, _ := p1.LookupFunctionByMethod("ExampleMethod")
	fmt.Printf("%#v\n", r1.Response())

	// With Options.ParseFunctionMode = ParseRequestOnly
	p2, err := Options{
		ParseFunctionMode: meta.ParseRequestOnly,
	}.NewDescritorFromPath(context.Background(), "../testdata/idl/example.thrift")
	if err != nil {
		panic(err)
	}
	r2, _ := p2.LookupFunctionByMethod("ExampleMethod")
	fmt.Printf("%#v\n", r2.Response())
}

func ExampleNewDescritorFromContent() {
	path := "a/b/main.thrift"
	content := `
	include "base/base.thrift"
	namespace go test.server

	service InboxService {
		base.BaseResp ExampleMethod(1: base.Base req)
	}`
	base := `
	namespace py base
	namespace go base
	namespace java com.bytedance.thrift.base
	
	struct TrafficEnv {
		1: bool Open = false,
		2: string Env = "",
	}
	
	struct Base {
		1: string LogID = "",
		2: string Caller = "",
		3: string Addr = "",
		4: string Client = "",
		5: optional TrafficEnv TrafficEnv,
		6: optional map<string, string> Extra,
	}
	
	struct BaseResp {
		1: string StatusMessage = "",
		2: i32 StatusCode = 0,
		3: optional map<string, string> Extra,
	}
	`

	includes := map[string]string{
		path:                   content,
		"a/b/base/base.thrift": base,
	}
	// default Options
	p1, err := NewDescritorFromContent(context.Background(), path, content, includes, true)
	if err != nil {
		panic(err)
	}
	r1, _ := p1.LookupFunctionByMethod("ExampleMethod")
	fmt.Printf("%#v\n", r1.Response())

	// use relative path here
	delete(includes, "a/b/base/base.thrift")
	includes["base/base.thrift"] = base
	// With Options.ParseFunctionMode = ParseRequestOnly
	p2, err := Options{
		ParseFunctionMode: meta.ParseRequestOnly,
	}.NewDescritorFromContent(context.Background(), path, content, includes, false)
	if err != nil {
		panic(err)
	}
	r2, _ := p2.LookupFunctionByMethod("ExampleMethod")
	fmt.Printf("%#v\n", r2.Response())
}

func ExampleBinaryProtocol_ReadAnyWithDesc() {
	p1, err := NewDescriptorFromPath(context.Background(), "../testdata/idl/example2.thrift")
	if err != nil {
		panic(err)
	}
	example2ReqDesc := p1.Functions()["ExampleMethod"].Request().Struct().FieldById(1).Type()
	data, err := ioutil.ReadFile("../testdata/data/example2.bin")
	if err != nil {
		panic(err)
	}

	p := NewBinaryProtocol(data)
	v, err := p.ReadAnyWithDesc(example2ReqDesc, false, false, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", v)
	p = NewBinaryProtocolBuffer()
	err = p.WriteAnyWithDesc(example2ReqDesc, v, true, true, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%x", p.RawBuf())
}
