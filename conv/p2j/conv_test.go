package p2j

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"testing"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/internal/util_test"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/base"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	goprotowire "google.golang.org/protobuf/encoding/protowire"
	goproto "google.golang.org/protobuf/proto"
)

const (
	exampleIDLPath   = "testdata/idl/example2.proto"
	exampleProtoPath = "testdata/data/example3_pb.bin"
	exampleJSONPath  = "testdata/data/example3req.json"
)

func TestBuildData(t *testing.T) {
	if err := generateExample3Data(); err != nil {
		panic("build example3ProtoData failed")
	}
	if err := generateExmaple3JSON(); err != nil {
		panic("build example3JSONData failed")
	}
}

func TestConvProto3JSON(t *testing.T) {
	desc := proto.FnRequest(proto.GetFnDescFromFile(exampleIDLPath, "ExampleMethod", proto.Options{}))
	//js := getExample2JSON()
	cv := NewBinaryConv(conv.Options{})
	in := getExample3Data()
	out, err := cv.Do(context.Background(), desc, in)
	if err != nil {
		t.Fatal(err)
	}
	exp := example2.ExampleReq{}
	// use kitex_util to check proto data validity
	l := 0
	dataLen := len(in)
	for l < dataLen {
		id, wtyp, tagLen := goprotowire.ConsumeTag(in)
		if tagLen < 0 {
			t.Fatal("proto data error format")
		}
		l += tagLen
		in = in[tagLen:]
		offset, err := exp.FastRead(in, int8(wtyp), int32(id))
		require.Nil(t, err)
		in = in[offset:]
		l += offset
	}
	if len(in) != 0 {
		t.Fatal("proto data error format")
	}
	// check json data validity, convert it into act struct
	var act example2.ExampleReq
	require.Nil(t, json.Unmarshal([]byte(out), &act))
	println(string(out))
	assert.Equal(t, exp, act)
}

func getExample2Req() *example2.ExampleReq {
	req := example2.ExampleReq{}
	req.Msg = "hello"
	req.Subfix = math.MaxFloat64
	req.InnerBase2 = &example2.InnerBase2{}
	req.InnerBase2.Bool = true
	req.InnerBase2.Uint32 = uint32(123)
	req.InnerBase2.Uint64 = uint64(123)
	req.InnerBase2.Double = float64(22.3)
	req.InnerBase2.String_ = "hello_inner"
	req.InnerBase2.ListInt32 = []int32{12, 13, 14, 15, 16, 17}
	req.InnerBase2.MapStringString = map[string]string{"m1": "aaa", "m2": "bbb", "m3": "ccc", "m4": "ddd"}
	req.InnerBase2.SetInt32 = []int32{200, 201, 202, 203, 204, 205}
	// req.InnerBase2.Foo = example2.FOO_FOO_A
	req.InnerBase2.MapInt32String = map[int32]string{1: "aaa", 2: "bbb", 3: "ccc", 4: "ddd"}
	req.InnerBase2.Binary = []byte{0x1, 0x2, 0x3, 0x4}
	req.InnerBase2.MapUint32String = map[uint32]string{uint32(1): "u32aa", uint32(2): "u32bb", uint32(3): "u32cc", uint32(4): "u32dd"}
	req.InnerBase2.MapUint64String = map[uint64]string{uint64(1): "u64aa", uint64(2): "u64bb", uint64(3): "u64cc", uint64(4): "u64dd"}
	req.InnerBase2.MapInt64String = map[int64]string{int64(1): "64aaa", int64(2): "64bbb", int64(3): "64ccc", int64(4): "64ddd"}
	req.InnerBase2.ListString = []string{"111", "222", "333", "44", "51", "6"}
	req.InnerBase2.ListBase = []*base.Base{{
		LogID:  "logId",
		Caller: "caller",
		Addr:   "addr",
		Client: "client",
		TrafficEnv: &base.TrafficEnv{
			Open: false,
			Env:  "env",
		},
		Extra: map[string]string{"1a": "aaa", "2a": "bbb", "3a": "ccc", "4a": "ddd"},
	}, {
		LogID:  "logId2",
		Caller: "caller2",
		Addr:   "addr2",
		Client: "client2",
		TrafficEnv: &base.TrafficEnv{
			Open: true,
			Env:  "env2",
		},
		Extra: map[string]string{"1a": "aaa2", "2a": "bbb2", "3a": "ccc2", "4a": "ddd2"},
	}}
	req.InnerBase2.MapInt64Base = map[int64]*base.Base{int64(1): {
		LogID:  "logId",
		Caller: "caller",
		Addr:   "addr",
		Client: "client",
		TrafficEnv: &base.TrafficEnv{
			Open: false,
			Env:  "env",
		},
		Extra: map[string]string{"1a": "aaa", "2a": "bbb", "3a": "ccc", "4a": "ddd"},
	}, int64(2): {
		LogID:  "logId2",
		Caller: "caller2",
		Addr:   "addr2",
		Client: "client2",
		TrafficEnv: &base.TrafficEnv{
			Open: true,
			Env:  "env2",
		},
		Extra: map[string]string{"1a": "aaa2", "2a": "bbb2", "3a": "ccc2", "4a": "ddd2"},
	}}
	req.InnerBase2.MapStringBase = map[string]*base.Base{"1": {
		LogID:  "logId",
		Caller: "caller",
		Addr:   "addr",
		Client: "client",
		TrafficEnv: &base.TrafficEnv{
			Open: false,
			Env:  "env",
		},
		Extra: map[string]string{"1a": "aaa", "2a": "bbb", "3a": "ccc", "4a": "ddd"},
	}, "2": {
		LogID:  "logId2",
		Caller: "caller2",
		Addr:   "addr2",
		Client: "client2",
		TrafficEnv: &base.TrafficEnv{
			Open: true,
			Env:  "env2",
		},
		Extra: map[string]string{"1a": "aaa2", "2a": "bbb2", "3a": "ccc2", "4a": "ddd2"},
	}}
	req.InnerBase2.Base = &base.Base{}
	req.InnerBase2.Base.LogID = "logId"
	req.InnerBase2.Base.Caller = "caller"
	req.InnerBase2.Base.Addr = "addr"
	req.InnerBase2.Base.Client = "client"
	req.InnerBase2.Base.TrafficEnv = &base.TrafficEnv{}
	req.InnerBase2.Base.TrafficEnv.Open = false
	req.InnerBase2.Base.TrafficEnv.Env = "env"
	req.InnerBase2.Base.Extra = map[string]string{"1b": "aaa", "2b": "bbb", "3b": "ccc", "4b": "ddd"}
	return &req
}

func generateExample3Data() error {
	req := getExample2Req()
	data, err := goproto.Marshal(req.ProtoReflect().Interface())
	if err != nil {
		panic("goproto marshal data failed")
	}
	checkExist := func(path string) bool {
		_, err := os.Stat(path)
		if err != nil {
			if os.IsExist(err) {
				return true
			}
			return false
		}
		return true
	}
	var file *os.File
	absoluteExampleProtoPath := util_test.MustGitPath(exampleProtoPath)
	if checkExist(absoluteExampleProtoPath) == true {
		if err := os.Remove(absoluteExampleProtoPath); err != nil {
			panic("delete protoBinaryFile failed")
		}
	}
	file, err = os.Create(absoluteExampleProtoPath)
	if err != nil {
		panic("create protoBinaryFile failed")
	}
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		panic("write protoBinary data failed")
	}
	return nil
}

func getExample3Data() []byte {
	out, err := ioutil.ReadFile(util_test.MustGitPath(exampleProtoPath))
	if err != nil {
		panic(err)
	}
	return out
}

func generateExmaple3JSON() error {
	req := getExample2Req()
	data, err := json.Marshal(req)
	if err != nil {
		panic(fmt.Sprintf("buildExampleJSONData failed, err: %v", err.Error()))
	}
	checkExist := func(path string) bool {
		_, err := os.Stat(path)
		if err != nil {
			if os.IsExist(err) {
				return true
			}
			return false
		}
		return true
	}
	var file *os.File
	absoluteExampleJSONPath := util_test.MustGitPath(exampleJSONPath)
	if checkExist(absoluteExampleJSONPath) == true {
		if err := os.Remove(absoluteExampleJSONPath); err != nil {
			panic("delete protoBinaryFile failed")
		}
	}
	file, err = os.Create(absoluteExampleJSONPath)
	if err != nil {
		panic("create protoBinaryFile failed")
	}
	defer file.Close()
	if _, err := file.WriteString(string(data)); err != nil {
		panic("write protoJSONData failed")
	}
	return nil
}

func getExample3JSON() string {
	out, err := ioutil.ReadFile(util_test.MustGitPath(exampleJSONPath))
	if err != nil {
		panic(err)
	}
	return string(out)
}
