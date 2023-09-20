package generic

import (
	"context"

	"github.com/cloudwego/dynamicgo/proto"
)

const (
	exampleIDLPath   = "../../testdata/idl/example2.proto"
	exampleProtoPath = "../../testdata/idl/example2_pb.bin"
	// exampleSuperProtoPath = "../../testdata/data/example2super.bin"
)

// parse protofile to get MessageDescriptor
func getExampleDesc() *proto.FieldDescriptor {
	svc, err := proto.NewDescritorFromPath(context.Background(), exampleIDLPath)
	if err != nil {
		panic(err)
	}
	res := (*svc).Methods().ByName("ExampleMethod").Input()

	if res == nil {
		panic("can't find Target MessageDescriptor")
	}
	return nil
}
