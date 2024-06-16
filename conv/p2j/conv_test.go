package p2j

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

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

var (
	debugAsyncGC = os.Getenv("SONIC_NO_ASYNC_GC") == ""
)

func TestMain(m *testing.M) {
	go func() {
		if !debugAsyncGC {
			return
		}
		println("Begin GC looping...")
		for {
			runtime.GC()
			debug.FreeOSMemory()
		}
	}()
	time.Sleep(time.Millisecond)
	m.Run()
}

const (
	exampleIDLPath        = "testdata/idl/example2.proto"
	exampleProtoPath      = "testdata/data/example3_pb.bin"
	exampleJSONPath       = "testdata/data/example3req.json"
	basicExampleIDLPath   = "testdata/idl/basic_example.proto"
	basicExampleJSONPath  = "testdata/data/basic_example.json"
	basicExampleProtoPath = "testdata/data/basic_example_pb.bin"
)

func TestBuildData(t *testing.T) {
	if err := saveExampleReqProtoBufData(); err != nil {
		panic("build example3ProtoData failed")
	}
}

func TestBuildBasicExampleData(t *testing.T) {
	if err := saveBasicExampleReqProtoBufData(); err != nil {
		panic("build basicExampleProtoData failed")
	}
}

func TestConvProto2JSON(t *testing.T) {
	includeDirs := util_test.MustGitPath("testdata/idl/") // includeDirs is used to find the include files.
	messageDesc := proto.FnRequest(proto.GetFnDescFromFile(exampleIDLPath, "ExampleMethod", proto.Options{}, includeDirs))
	//js := getExample2JSON()
	cv := NewBinaryConv(conv.Options{})
	in := readExampleReqProtoBufData()
	out, err := cv.Do(context.Background(), messageDesc, in)
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
	assert.Equal(t, exp, act)
}

func TestConvProto2JSON_BasicExample(t *testing.T) {
	includeDirs := util_test.MustGitPath("testdata/idl/") // includeDirs is used to find the include files.
	messageDesc := proto.FnRequest(proto.GetFnDescFromFile(basicExampleIDLPath, "ExampleMethod", proto.Options{}, includeDirs))

	cv := NewBinaryConv(conv.Options{})
	in := readBasicExampleReqProtoBufData()
	out, err := cv.Do(context.Background(), messageDesc, in)
	if err != nil {
		t.Fatal(err)
	}
	exp := base.BasicExample{}
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
	var act base.BasicExample
	require.Nil(t, json.Unmarshal([]byte(out), &act))
	fmt.Println(exp)
	fmt.Println("-------------")
	fmt.Println(act)
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
	req.InnerBase2.Int32 = int32(123)
	req.InnerBase2.SInt64 = int64(123)
	req.InnerBase2.Double = float64(22.3)
	req.InnerBase2.String_ = "hello_inner"
	req.InnerBase2.ListInt32 = []int32{12, 13, 14, 15, 16, 17}
	req.InnerBase2.MapStringString = map[string]string{"m1": "aaa", "m2": "bbb", "m3": "ccc", "m4": "ddd"}
	req.InnerBase2.ListSInt64 = []int64{200, 201, 202, 203, 204, 205}
	req.InnerBase2.Foo = example2.FOO_FOO_A
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

func constructBasicExampleReqObject() *base.BasicExample {
	req := new(base.BasicExample)
	req = &base.BasicExample{
		Int32:        123,
		Int64:        123,
		Uint32:       uint32(math.MaxInt32),
		Uint64:       uint64(math.MaxInt64),
		Sint32:       123,
		Sint64:       123,
		Sfixed32:     123,
		Sfixed64:     123,
		Fixed32:      123,
		Fixed64:      123,
		Float:        123.123,
		Double:       123.123,
		Bool:         true,
		Str:          "hello world!",
		Bytes:        []byte{0x1, 0x2, 0x3, 0x4},
		ListInt32:    []int32{100, 200, 300, 400, 500},
		ListInt64:    []int64{100, 200, 300, 400, 500},
		ListUint32:   []uint32{100, 200, 300, 400, 500},
		ListUint64:   []uint64{100, 200, 300, 400, 500},
		ListSint32:   []int32{100, 200, 300, 400, 500},
		ListSint64:   []int64{100, 200, 300, 400, 500},
		ListSfixed32: []int32{100, 200, 300, 400, 500},
		ListSfixed64: []int64{100, 200, 300, 400, 500},
		ListFixed32:  []uint32{100, 200, 300, 400, 500},
		ListFixed64:  []uint64{100, 200, 300, 400, 500},
		ListFloat:    []float32{1.1, 2.2, 3.3, 4.4, 5.5},
		ListDouble:   []float64{1.1, 2.2, 3.3, 4.4, 5.5},
		ListBool:     []bool{true, false, true, false, true},
		ListString:   []string{"a1", "b2", "c3", "d4", "e5"},
		ListBytes:    [][]byte{{0x1, 0x2, 0x3, 0x4}, {0x5, 0x6, 0x7, 0x8}},
		MapInt64SINT32: map[int64]int32{
			0:             0,
			math.MaxInt64: math.MaxInt32,
			math.MinInt64: math.MinInt32,
		},
		MapInt64Sfixed32: map[int64]int32{
			0:             0,
			math.MaxInt64: math.MaxInt32,
			math.MinInt64: math.MinInt32,
		},
		MapInt64Fixed32: map[int64]uint32{
			math.MaxInt64: uint32(math.MaxInt32),
			math.MinInt64: 0,
		},
		MapInt64Uint32: map[int64]uint32{
			0:             0,
			math.MaxInt64: uint32(math.MaxInt32),
			math.MinInt64: 0,
		},
		MapInt64Double: map[int64]float64{
			0:             0,
			math.MaxInt64: math.MaxFloat64,
			math.MinInt64: math.SmallestNonzeroFloat64,
		},
		MapInt64Bool: map[int64]bool{
			0:             false,
			math.MaxInt64: true,
			math.MinInt64: false,
		},
		MapInt64String: map[int64]string{
			0:             "0",
			math.MaxInt64: "max",
			math.MinInt64: "min",
		},
		MapInt64Bytes: map[int64][]byte{
			0:             {0x0},
			math.MaxInt64: {0x1, 0x2, 0x3, 0x4},
			math.MinInt64: {0x5, 0x6, 0x7, 0x8},
		},
		MapInt64Float: map[int64]float32{
			0:             0,
			math.MaxInt64: math.MaxFloat32,
			math.MinInt64: math.SmallestNonzeroFloat32,
		},
		MapInt64Int32: map[int64]int32{
			0:             0,
			math.MaxInt64: math.MaxInt32,
			math.MinInt64: math.MinInt32,
		},
		MapstringSINT64: map[string]int64{
			"0":   0,
			"max": math.MaxInt64,
			"min": math.MinInt64,
		},
		MapstringSfixed64: map[string]int64{
			"0":   0,
			"max": math.MaxInt64,
			"min": math.MinInt64,
		},
		MapstringFixed64: map[string]uint64{
			"max": uint64(math.MaxInt64),
			"min": 0,
		},
		MapstringUint64: map[string]uint64{
			"max": uint64(math.MaxInt64),
			"min": 0,
		},
		MapstringDouble: map[string]float64{
			"0":   0,
			"max": math.MaxFloat64,
			"min": math.SmallestNonzeroFloat64,
		},
		MapstringBool: map[string]bool{
			"0":   false,
			"max": true,
		},
		MapstringString: map[string]string{
			"0":   "0",
			"max": "max",
			"min": "min",
		},
		MapstringBytes: map[string][]byte{
			"0":   {0x0},
			"max": {0x1, 0x2, 0x3, 0x4},
			"min": {0x5, 0x6, 0x7, 0x8},
		},
		MapstringFloat: map[string]float32{
			"0":   0,
			"max": math.MaxFloat32,
			"min": math.SmallestNonzeroFloat32,
		},
		MapstringInt64: map[string]int64{
			"0":   0,
			"max": math.MaxInt64,
			"min": math.MinInt64,
		},
	}

	return req
}

// marshal ExampleReq Object to ProtoBinary, and write binaryData into exampleProtoPath
func saveExampleReqProtoBufData() error {
	req := constructExampleReqObject()
	data, err := goproto.Marshal(req)
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

// construct BasicExampleReq Object
func saveBasicExampleReqProtoBufData() error {
	req := constructBasicExampleReqObject()
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
	absoluteExampleProtoPath := util_test.MustGitPath(basicExampleProtoPath)
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

func readBasicExampleReqProtoBufData() []byte {
	out, err := ioutil.ReadFile(util_test.MustGitPath(basicExampleProtoPath))
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

func getExampleInt2Float() *proto.TypeDescriptor {
	includeDirs := util_test.MustGitPath("testdata/idl/") // includeDirs is used to find the include files.
	svc, err := proto.NewDescritorFromPath(context.Background(), util_test.MustGitPath(exampleIDLPath), includeDirs)
	if err != nil {
		panic(err)
	}
	return (*svc).LookupMethodByName("Int2FloatMethod").Output()
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

func getExampleReqPartialDesc() *proto.TypeDescriptor {
	includeDirs := util_test.MustGitPath("testdata/idl/") // includeDirs is used to find the include files.
	return proto.FnRequest(proto.GetFnDescFromFile(exampleIDLPath, "ExamplePartialMethod", proto.Options{}, includeDirs))
}

func getExampleRespPartialDesc() *proto.TypeDescriptor {
	includeDirs := util_test.MustGitPath("testdata/idl/") // includeDirs is used to find the include files.
	return proto.FnResponse(proto.GetFnDescFromFile(exampleIDLPath, "ExamplePartialMethod2", proto.Options{}, includeDirs))
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
