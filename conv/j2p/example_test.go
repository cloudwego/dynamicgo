package j2p

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example2"
	"google.golang.org/protobuf/encoding/protowire"
)

var opts = conv.Options{}

func ExampleBinaryConv_Do() {
	// get descriptor and data
	messageDesc := getExampleDesc()
	desc := (*messageDesc).(proto.Descriptor)
	data := getExampleData()

	// make BinaryConv
	cv := NewBinaryConv(opts)

	// do conversion
	out, err := cv.Do(context.Background(), &desc, data)
	if err != nil {
		panic(err)
	}

	// validate result
	exp := &example2.ExampleReq{}
	err = json.Unmarshal(data, exp)
	if err != nil {
		panic(err)
	}
	act := &example2.ExampleReq{}
	l := 0
	dataLen := len(out)
	// fastRead to get target struct
	for l < dataLen {
		id, wtyp, tagLen := protowire.ConsumeTag(out)
		if tagLen < 0 {
			panic("parseTag failed")
		}
		l += tagLen
		out = out[tagLen:]
		offset, err := act.FastRead(out, int8(wtyp), int32(id))
		if err != nil {
			panic(err)
		}
		out = out[offset:]
		l += offset
	}
	if !reflect.DeepEqual(exp, act) {
		panic("not equal")
	}
}
