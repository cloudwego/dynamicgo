package p2j

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/internal/util_test"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example2"
	goprotowire "google.golang.org/protobuf/encoding/protowire"
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

	// use kitex_util to check proto data validity
	l := 0
	dataLen := len(in)
	for l < dataLen {
		id, wtyp, tagLen := goprotowire.ConsumeTag(in)
		if tagLen < 0 {
			panic("proto data error format")
		}
		l += tagLen
		in = in[tagLen:]
		offset, err := exp.FastRead(in, int8(wtyp), int32(id))
		if err != nil {
			panic(err)
		}
		in = in[offset:]
		l += offset
	}
	if len(in) != 0 {
		panic("proto data error format")
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
