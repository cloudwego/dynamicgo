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
	"github.com/cloudwego/dynamicgo/meta"
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
	if err := saveExampleReqProtoBufData(); err != nil {
		panic("build example3ProtoData failed")
	}
	if err := saveExampleReqJSONData(); err != nil {
		panic("build example3JSONData failed")
	}
}

func TestConvProto3JSON(t *testing.T) {
	messageDesc := proto.FnRequest(proto.GetFnDescFromFile(exampleIDLPath, "ExampleMethod", proto.Options{}))
	desc, ok := (*messageDesc).(proto.Descriptor)
	if !ok {
		t.Fatal("invalid descrptor")
	}
	//js := getExample2JSON()
	cv := NewBinaryConv(conv.Options{})
	in := readExampleReqProtoBufData()
	out, err := cv.Do(context.Background(), &desc, in)
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

// construct ExampleReq Object
func constructExampleReqObject() *example2.ExampleReq {
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

// marshal ExampleReq Object to ProtoBinary, and write binaryData into exampleProtoPath
func saveExampleReqProtoBufData() error {
	req := constructExampleReqObject()
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

// read ProtoBuf's data in binary format from exampleProtoPath
func readExampleReqProtoBufData() []byte {
	out, err := ioutil.ReadFile(util_test.MustGitPath(exampleProtoPath))
	if err != nil {
		panic(err)
	}
	return out
}

// marshal ExampleReq Object to JsonBinary, and write binaryData into exampleJSONPath
func saveExampleReqJSONData() error {
	req := constructExampleReqObject()
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

// read JSON's data in binary format from exampleJSONPath
func readExampleReqJSONData() string {
	out, err := ioutil.ReadFile(util_test.MustGitPath(exampleJSONPath))
	if err != nil {
		panic(err)
	}
	return string(out)
}

func getExampleInt2Float() *proto.Descriptor {
	svc, err := proto.NewDescritorFromPath(context.Background(), util_test.MustGitPath(exampleIDLPath))
	if err != nil {
		panic(err)
	}
	fieldDesc := (*svc).Methods().ByName("Int2FloatMethod").Output()
	desc, ok := fieldDesc.(proto.Descriptor)
	if ok {
		return &desc
	}
	return nil
}

func TestInt2String(t *testing.T) {
	cv := NewBinaryConv(conv.Options{})
	desc := getExampleInt2Float()
	exp := example2.ExampleInt2Float{}
	exp.Int32 = 1
	exp.Int64 = 2
	exp.Float64 = 3.14
	exp.String_ = "hello"
	exp.Subfix = 0.92653
	ctx := context.Background()
	in := make([]byte, exp.Size())
	exp.FastWrite(in)

	out, err := cv.Do(ctx, desc, in)
	require.NoError(t, err)
	require.Equal(t, `{"Int32":1,"Float64":3.14,"String":"hello","Int64":2,"Subfix":0.92653}`, string(out))

	cv.opts.EnableValueMapping = false
	out, err = cv.Do(ctx, desc, in)
	require.NoError(t, err)
	require.Equal(t, (`{"Int32":1,"Float64":3.14,"String":"hello","Int64":2,"Subfix":0.92653}`), string(out))

	cv.opts.String2Int64 = true
	out, err = cv.Do(ctx, desc, in)
	require.NoError(t, err)
	require.Equal(t, (`{"Int32":1,"Float64":3.14,"String":"hello","Int64":"2","Subfix":0.92653}`), string(out))
}

func getExampleReqPartialDesc() *proto.Descriptor {
	messageDesc := proto.FnRequest(proto.GetFnDescFromFile(exampleIDLPath, "ExamplePartialMethod", proto.Options{}))
	desc, ok := (*messageDesc).(proto.Descriptor)
	if !ok {
		return nil
	}
	return &desc
}

func getExampleRespPartialDesc() *proto.Descriptor {
	messageDesc := proto.FnResponse(proto.GetFnDescFromFile(exampleIDLPath, "ExamplePartialMethod2", proto.Options{}))
	desc, ok := (*messageDesc).(proto.Descriptor)
	if !ok {
		return nil
	}
	return &desc
}

// construct ExampleResp Object
func getExampleResp() *example2.ExampleResp {
	resp := example2.ExampleResp{}
	resp.Msg = "messagefist"
	resp.RequiredField = "hello"
	resp.BaseResp = &base.BaseResp{}
	resp.BaseResp.StatusMessage = "status1"
	resp.BaseResp.StatusCode = 32
	resp.BaseResp.Extra = map[string]string{"1b": "aaa", "2b": "bbb", "3b": "ccc", "4b": "ddd"}
	return &resp
}

func TestUnknowFields(t *testing.T) {
	t.Run("top", func(t *testing.T) {
		cv := NewBinaryConv(conv.Options{
			DisallowUnknownField: true,
		})
		resp := getExampleResp()
		data, err := goproto.Marshal(resp.ProtoReflect().Interface())
		if err != nil {
			t.Fatal("marshal protobuf data failed")
		}
		partialRespDesc := getExampleRespPartialDesc()
		_, err = cv.Do(context.Background(), partialRespDesc, data)
		require.Error(t, err)
		require.Equal(t, meta.ErrUnknownField, err.(meta.Error).Code.Behavior())
	})

	t.Run("nested", func(t *testing.T) {
		cv := NewBinaryConv(conv.Options{
			DisallowUnknownField: true,
		})
		partialReqDesc := getExampleReqPartialDesc()
		in := readExampleReqProtoBufData()
		_, err := cv.Do(context.Background(), partialReqDesc, in)
		require.Error(t, err)
		require.Equal(t, meta.ErrUnknownField, err.(meta.Error).Code.Behavior())
	})

	t.Run("skip top", func(t *testing.T) {
		cv := NewBinaryConv(conv.Options{
			DisallowUnknownField: false,
		})
		resp := getExampleResp()
		data, err := goproto.Marshal(resp.ProtoReflect().Interface())
		if err != nil {
			t.Fatal("marshal protobuf data failed")
		}
		partialRespDesc := getExampleRespPartialDesc()
		_, err = cv.Do(context.Background(), partialRespDesc, data)
		require.NoError(t, err)
	})

	t.Run("skip nested", func(t *testing.T) {
		cv := NewBinaryConv(conv.Options{
			DisallowUnknownField: false,
		})
		partialReqDesc := getExampleReqPartialDesc()
		in := readExampleReqProtoBufData()
		_, err := cv.Do(context.Background(), partialReqDesc, in)
		require.NoError(t, err)
	})
}
