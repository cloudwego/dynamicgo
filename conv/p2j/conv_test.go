package p2j

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"testing"
	"time"
	"unsafe"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/conv/j2p"
	"github.com/cloudwego/dynamicgo/internal/util_test"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/base"
	examplepb "github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	goproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoimpl"
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
	NestedExampleIDLPath  = "testdata/idl/nested.proto"
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

func protoUnmarshal(b []byte, m goproto.Message) error {
	if err := goproto.Unmarshal(b, m); err != nil {
		return err
	}
	// reset internal protobuf state field
	// so that we can require.Equal later without pain

	// use UnsafePointer() >= go1.18
	p := unsafe.Pointer(reflect.ValueOf(m).Pointer())

	// protoimpl.MessageState is always the 1st field
	// reset it for require.Equal
	*(*protoimpl.MessageState)(p) = protoimpl.MessageState{}
	return nil
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
	// use proto.Unmarshal to check proto data validity
	err = protoUnmarshal(in, &exp)
	require.Nil(t, err)
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
	// use proto.Unmarshal to check proto data validity
	err = protoUnmarshal(in, &exp)
	require.Nil(t, err)
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
	in, err := goproto.Marshal(&exp)
	require.NoError(t, err)

	out, err := cv.Do(ctx, desc, in)
	require.NoError(t, err)
	require.Equal(t, `{"Int32":1,"Float64":3.14,"String":"hello","Int64":2,"Subfix":0.92653}`, string(out))

	cv.opts.EnableValueMapping = false
	out, err = cv.Do(ctx, desc, in)
	require.NoError(t, err)
	require.Equal(t, (`{"Int32":1,"Float64":3.14,"String":"hello","Int64":2,"Subfix":0.92653}`), string(out))

	cv.opts.Int642String = true
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

func TestNestedField(t *testing.T) {
	t.Run("nested nested string list example", func(t *testing.T) {
		protoContent := `
		syntax = "proto3";
		message Item {
			repeated string ids = 1;
		}

		message ExampleReq {
			repeated Item items = 1;
		}

		message ExampleResp {
			repeated Item items = 1;
		}

		service Service {
			rpc Example(ExampleReq) returns (ExampleResp);
		}
		`
		dataContent := `{"items":[{"ids":["123"]},{"ids":["456","789"]}]}`
		// itemtag L [ idstag1 L V idstag1 L V ] itemtag L [ idstag1 L V ]

		ctx := context.Background()
		serviceDesc, err := proto.NewDescritorFromContent(ctx, "example.proto", protoContent, map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		reqDesc := serviceDesc.LookupMethodByName("Example").Input()
		conv1 := j2p.NewBinaryConv(conv.Options{})
		reqProto, err := conv1.Do(ctx, reqDesc, []byte(dataContent))

		if err != nil {
			t.Fatal("json marshal to protobuf error")
		}
		conv2 := NewBinaryConv(conv.Options{})
		reqJson, err := conv2.Do(ctx, reqDesc, reqProto)
		if err != nil {
			t.Fatal("protobuf marshal to json error")
		}
		fmt.Println("reqJson:", string(reqJson))
		require.Equal(t, dataContent, string(reqJson))
	})

	t.Run("nested nested map field", func(t *testing.T) {
		protoContent := `
		syntax = "proto3";
		message Item {
			map<string, string> ids = 1;
		}

		message ExampleReq {
			repeated Item items = 1;
		}

		message ExampleResp {
			repeated Item items = 1;
		}

		service Service {
			rpc Example(ExampleReq) returns (ExampleResp);
		}
		`
		dataContent := `{"items":[{"ids":{"123":"456","abc":"def"}},{"ids":{"789":"012"}}]}`

		ctx := context.Background()
		serviceDesc, err := proto.NewDescritorFromContent(ctx, "example.proto", protoContent, map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		reqDesc := serviceDesc.LookupMethodByName("Example").Input()

		conv1 := j2p.NewBinaryConv(conv.Options{})
		reqProto, err := conv1.Do(ctx, reqDesc, []byte(dataContent))
		if err != nil {
			t.Fatal("json marshal to protobuf error")
		}

		conv2 := NewBinaryConv(conv.Options{})
		reqJson, err := conv2.Do(ctx, reqDesc, reqProto)

		if err != nil {
			t.Fatal("protobuf marshal to json error")
		}
		fmt.Println("reqJson:", string(reqJson))
		require.Equal(t, dataContent, string(reqJson))
	})

	t.Run("nested nested map field with int key", func(t *testing.T) {
		protoContent := `
		syntax = "proto3";
		message Item {
			map<int32, string> ids = 1;
		}

		message ExampleReq {
			repeated Item items = 1;
		}

		message ExampleResp {
			repeated Item items = 1;
		}

		service Service {
			rpc Example(ExampleReq) returns (ExampleResp);
		}
		`
		dataContent := `{"items":[{"ids":{"123":"456","789":"012"}},{"ids":{"345":"678"}}]}`

		ctx := context.Background()
		serviceDesc, err := proto.NewDescritorFromContent(ctx, "example.proto", protoContent, map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		reqDesc := serviceDesc.LookupMethodByName("Example").Input()

		conv1 := j2p.NewBinaryConv(conv.Options{})
		reqProto, err := conv1.Do(ctx, reqDesc, []byte(dataContent))
		if err != nil {
			t.Fatal("json marshal to protobuf error")
		}

		conv2 := NewBinaryConv(conv.Options{})
		reqJson, err := conv2.Do(ctx, reqDesc, reqProto)

		if err != nil {
			t.Fatal("protobuf marshal to json error")
		}
		fmt.Println("reqJson:", string(reqJson))
		require.Equal(t, dataContent, string(reqJson))
	})

	t.Run("nested nested int list", func(t *testing.T) {
		protoContent := `
		syntax = "proto3";
		message Item {
			repeated int32 ids = 1;
		}

		message ExampleReq {
			repeated Item items = 1;
		}

		message ExampleResp {
			repeated Item items = 1;
		}

		service Service {
			rpc Example(ExampleReq) returns (ExampleResp);
		}
		`
		dataContent := `{"items":[{"ids":[123,456]},{"ids":[789]}]}`

		ctx := context.Background()
		serviceDesc, err := proto.NewDescritorFromContent(ctx, "example.proto", protoContent, map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		reqDesc := serviceDesc.LookupMethodByName("Example").Input()

		conv1 := j2p.NewBinaryConv(conv.Options{})
		reqProto, err := conv1.Do(ctx, reqDesc, []byte(dataContent))
		if err != nil {
			t.Fatal("json marshal to protobuf error")
		}

		conv2 := NewBinaryConv(conv.Options{})
		reqJson, err := conv2.Do(ctx, reqDesc, reqProto)

		if err != nil {
			t.Fatal("protobuf marshal to json error")
		}
		fmt.Println("reqJson:", string(reqJson))
		require.Equal(t, dataContent, string(reqJson))
	})

	t.Run("map value with nested list", func(t *testing.T) {
		protoContent := `
		syntax = "proto3";
		message Item {
			repeated string name = 1;
		}
		message ItemValue {
			repeated Item ids = 1;
		}

		message ExampleReq {
			map<string, ItemValue> items = 1;
		}

		message ExampleResp {
			repeated Item items = 1;
		}

		service Service {
			rpc Example(ExampleReq) returns (ExampleResp);
		}
		`
		dataContent := `{"items":{"key1":{"ids":[{"name":["123"]},{"name":["456","789"]}]},"key2":{"ids":[{"name":["abc"]}]}}}`

		ctx := context.Background()
		serviceDesc, err := proto.NewDescritorFromContent(ctx, "example.proto", protoContent, map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		reqDesc := serviceDesc.LookupMethodByName("Example").Input()

		conv1 := j2p.NewBinaryConv(conv.Options{})
		reqProto, err := conv1.Do(ctx, reqDesc, []byte(dataContent))
		if err != nil {
			t.Fatal("json marshal to protobuf error")
		}

		conv2 := NewBinaryConv(conv.Options{})
		reqJson, err := conv2.Do(ctx, reqDesc, reqProto)

		if err != nil {
			t.Fatal("protobuf marshal to json error")
		}
		fmt.Println("reqJson:", string(reqJson))
		require.Equal(t, dataContent, string(reqJson))
	})

	t.Run("nested complex struct", func(t *testing.T) {
		protoContent := `
		syntax = "proto3";
		message Item {
			repeated string id = 1;
			bool hasSub = 2;
			repeated int32 value = 3;
			map<string, int32> extra = 4;
			repeated Item subItems = 5;
		}

		message NestedExampleReq {
			repeated Item items = 1;
		}

		message NestedExampleResp {
			repeated Item items = 1;
		}

		service Service {
			rpc Example(NestedExampleReq) returns (NestedExampleResp);
		}
		`
		dataContent := `{"items":[{"id":["123"],"hasSub":false,"value":[1,2,3],"extra":{"key1":10,"key2":20}},{"id":["123"],"hasSub":true,"value":[1,2,3],"extra":{"key1":10,"key2":20},"subItems":[{"id":["sub123","sub456"],"value":[111,222,333],"extra":{"subkey1":10,"subkey2":20}},{"id":["sub789"],"extra":{"SUB7":7,"SUB8":8,"SUB9":9}}]}]}`
		ctx := context.Background()
		serviceDesc, err := proto.NewDescritorFromContent(ctx, "example.proto", protoContent, map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		reqDesc := serviceDesc.LookupMethodByName("Example").Input()

		conv1 := j2p.NewBinaryConv(conv.Options{})
		reqProto, err := conv1.Do(ctx, reqDesc, []byte(dataContent))
		if err != nil {
			t.Fatal("json marshal to protobuf error")
		}
		fmt.Println("reqProto len:", len(reqProto)) // byte length 159

		conv2 := NewBinaryConv(conv.Options{})
		reqJson, err := conv2.Do(ctx, reqDesc, reqProto)
		if err != nil {
			t.Fatal("protobuf marshal to json error")
		}
		fmt.Println("reqJson:", string(reqJson))
		require.Equal(t, dataContent, string(reqJson))
	})

	t.Run("test p2j with offical go protobuf marshal", func(t *testing.T) {
		includeDirs := util_test.MustGitPath("testdata/idl/") // includeDirs is used to find the include files.
		requestDesc := proto.FnRequest(proto.GetFnDescFromFile(NestedExampleIDLPath, "Example", proto.Options{}, includeDirs))
		if requestDesc == nil {
			t.Fatal("descriptor is nil")
		}
		// we also support using p2j with offical marshal protobuf, it ignore default value field
		// in this case "hasSub":false will be ignored
		req := &examplepb.NestedExampleReq{
			Items: []*examplepb.Item{
				{
					Id:     []string{"123"},
					HasSub: false,
					Value:  []int32{1, 2, 3},
					Extra:  map[string]int32{"key1": 10},
				},
				{
					Id:     []string{"123"},
					HasSub: true,
					Value:  []int32{1, 2, 3},
					Extra:  map[string]int32{"key1": 10},
					SubItems: []*examplepb.Item{
						{
							Id:    []string{"sub123"},
							Value: []int32{111, 222, 333},
							Extra: map[string]int32{"subkey1": 10},
						},
						{
							Id:    []string{"sub789"},
							Extra: map[string]int32{"SUB7": 7},
						},
					},
				},
			},
		}

		origindata, err := goproto.Marshal(req)
		if err != nil {
			t.Fatal("offical protobuf marshal error:", err)
		}

		conv2 := NewBinaryConv(conv.Options{})
		ctx := context.Background()
		fmt.Println("reqProto len:", len(origindata))
		officalReqJson, err := conv2.Do(ctx, requestDesc, origindata)
		if err != nil {
			t.Fatal("protobuf marshal to json error")
		}
		fmt.Println("officalReqJson:", string(officalReqJson))
		expOfficaldataContent := `{"items":[{"id":["123"],"value":[1,2,3],"extra":{"key1":10}},{"id":["123"],"hasSub":true,"value":[1,2,3],"extra":{"key1":10},"subItems":[{"id":["sub123"],"value":[111,222,333],"extra":{"subkey1":10}},{"id":["sub789"],"extra":{"SUB7":7}}]}]}`
		require.Equal(t, expOfficaldataContent, string(officalReqJson))
	})

}
