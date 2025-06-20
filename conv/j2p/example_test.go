package j2p

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example2"
	"google.golang.org/protobuf/proto"
)

var opts = conv.Options{}

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
	exp := &example2.ExampleReq{}
	err = json.Unmarshal(data, exp)
	if err != nil {
		panic(err)
	}
	act := &example2.ExampleReq{}
	err = proto.Unmarshal(out, act)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(exp, act) {
		panic("not equal")
	}
}
