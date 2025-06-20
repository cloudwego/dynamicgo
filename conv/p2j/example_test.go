package p2j

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/internal/util_test"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example2"
	goproto "google.golang.org/protobuf/proto"
)

var opts = conv.Options{}

func ExampleBinaryConv_Do() {
	// get descriptor and data
	includeDirs := util_test.MustGitPath("testdata/idl/") // includeDirs is used to find the include files.
	desc := proto.FnRequest(proto.GetFnDescFromFile(exampleIDLPath, "ExampleMethod", proto.Options{}, includeDirs))

	// make BinaryConv
	cv := NewBinaryConv(conv.Options{})
	in := readExampleReqProtoBufData()

	// do conversion
	out, err := cv.Do(context.Background(), desc, in)
	if err != nil {
		panic(err)
	}
	exp := example2.ExampleReq{}

	// use proto.Unmarshal to check proto data validity
	err = goproto.Unmarshal(readExampleReqProtoBufData(), &exp)
	if err != nil {
		panic(err)
	}

	// validate result
	var act example2.ExampleReq
	err = json.Unmarshal([]byte(out), &act)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(exp, act) {
		panic("not equal")
	}
}
