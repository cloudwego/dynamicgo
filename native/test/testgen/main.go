package main

import (
	"context"

	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/davecgh/go-spew/spew"
)

func main() {

	svc, err := thrift.NewDescriptorFromPath(
		context.Background(),
		"../../../testdata/idl/example2.thrift",
	)
	if err != nil {
		panic(err)
	}

	methDesc, err := svc.LookupFunctionByMethod("ExampleMethod")
	if err != nil {
		panic(err)
	}

	spew.Dump(methDesc)

}
