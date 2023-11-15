package p2j

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example2"
	goprotowire "google.golang.org/protobuf/encoding/protowire"
)

var opts = conv.Options{}

func ExampleBinaryConv_Do() {
	// get descriptor and data
	messageDesc := proto.FnRequest(proto.GetFnDescFromFile(exampleIDLPath, "ExampleMethod", proto.Options{}))
	desc, ok := (*messageDesc).(proto.Descriptor)
	if !ok {
		panic("invalid descrptor")
	}

	// make BinaryConv
	cv := NewBinaryConv(conv.Options{})
	in := readExampleReqProtoBufData()

	// do conversion
	out, err := cv.Do(context.Background(), &desc, in)
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
