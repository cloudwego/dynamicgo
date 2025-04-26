package j2p

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"testing"
	"time"

	"github.com/bytedance/sonic/ast"
	"github.com/stretchr/testify/require"
	goprotowire "google.golang.org/protobuf/encoding/protowire"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/internal/util_test"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/base"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example2"
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
	exampleIDLPath      = "testdata/idl/example2.proto"
	exampleJSON         = "testdata/data/example2req.json"
	exampleProtoPath    = "testdata/data/example2_pb.bin" // not used
	basicExampleJSON    = "testdata/data/basic_example.json"
	basicExampleIDLPath = "testdata/idl/basic_example.proto"
	exampleEmptyIDLPath = "testdata/idl/empty_example.proto"
)

func TestBuildExampleJSONData(t *testing.T) {
	buildExampleJSONData()
}

func TestBuildBasicExampleJSONData(t *testing.T) {
	buildBasicExampleJSONData()
}

func TestConvJSON2Protobf(t *testing.T) {
	// buildExampleJSONData()
	desc := getExampleDesc()
	data := getExampleData()
	// pdata, _ := ioutil.ReadFile(util_test.MustGitPath(exampleProtoPath))
	// fmt.Println(pdata)
	cv := NewBinaryConv(conv.Options{})
	ctx := context.Background()
	// get protobuf-encode bytes
	out, err := cv.Do(ctx, desc, data)
	require.Nil(t, err)
	exp := &example2.ExampleReq{}
	// unmarshal target struct
	err = json.Unmarshal(data, exp)
	require.Nil(t, err)
	act := &example2.ExampleReq{}
	l := 0
	// fmt.Print(out)
	dataLen := len(out)
	// fastRead to get target struct
	for l < dataLen {
		id, wtyp, tagLen := goprotowire.ConsumeTag(out)
		if tagLen < 0 {
			t.Fatal("test failed")
		}
		l += tagLen
		out = out[tagLen:]
		offset, err := act.FastRead(out, int8(wtyp), int32(id))
		require.Nil(t, err)
		out = out[offset:]
		l += offset
	}
	require.Nil(t, err)
	// compare exp and act struct
	require.Equal(t, exp, act)
}

func TestConvJSON2Protobf_BasicExample(t *testing.T) {
	// buildExampleJSONData()
	desc := getBasicExampleDesc()
	data := getBasicExampleData()
	// pdata, _ := ioutil.ReadFile(util_test.MustGitPath(exampleProtoPath))
	// fmt.Println(pdata)
	cv := NewBinaryConv(conv.Options{})
	ctx := context.Background()
	// get protobuf-encode bytes
	out, err := cv.Do(ctx, desc, data)
	require.Nil(t, err)
	exp := &base.BasicExample{}
	// unmarshal target struct
	err = json.Unmarshal(data, exp)
	require.Nil(t, err)
	act := &base.BasicExample{}
	l := 0
	// fmt.Print(out)
	dataLen := len(out)
	// fastRead to get target struct
	for l < dataLen {
		id, wtyp, tagLen := goprotowire.ConsumeTag(out)
		fmt.Println("id", id)
		fmt.Println("w", wtyp)
		if tagLen < 0 {
			t.Fatal("test failed")
		}
		l += tagLen
		out = out[tagLen:]
		offset, err := act.FastRead(out, int8(wtyp), int32(id))
		require.Nil(t, err)
		out = out[offset:]
		l += offset
	}
	require.Nil(t, err)
	// compare exp and act struct
	fmt.Println(exp)
	fmt.Println("----------------")
	fmt.Println(act)
	require.Equal(t, exp, act)
}

func getBasicExampleDesc() *proto.TypeDescriptor {
	opts := proto.Options{}
	includeDirs := util_test.MustGitPath("testdata/idl/") // includeDirs is used to find the include files.
	svc, err := opts.NewDescriptorFromPath(context.Background(), util_test.MustGitPath(basicExampleIDLPath), includeDirs)
	if err != nil {
		panic(err)
	}
	res := (*svc).LookupMethodByName("ExampleMethod").Input()

	if res == nil {
		panic("can't find Target MessageDescriptor")
	}
	return res
}

func getBasicExampleData() []byte {
	out, err := ioutil.ReadFile(util_test.MustGitPath(basicExampleJSON))
	if err != nil {
		panic(err)
	}
	return out
}

func getExampleDesc() *proto.TypeDescriptor {
	opts := proto.Options{}
	includeDirs := util_test.MustGitPath("testdata/idl/") // includeDirs is used to find the include files.
	svc, err := opts.NewDescriptorFromPath(context.Background(), util_test.MustGitPath(exampleIDLPath), includeDirs)
	if err != nil {
		panic(err)
	}
	res := (*svc).LookupMethodByName("ExampleMethod").Input()

	if res == nil {
		panic("can't find Target MessageDescriptor")
	}
	return res
}

func getExampleData() []byte {
	out, err := ioutil.ReadFile(util_test.MustGitPath(exampleJSON))
	if err != nil {
		panic(err)
	}
	return out
}

func getExample2Req() *example2.ExampleReq {
	req := new(example2.ExampleReq)
	req.Msg = "hello"
	req.A = 25
	req.InnerBase2 = &example2.InnerBase2{}
	req.InnerBase2.Bool = true
	req.InnerBase2.Uint32 = uint32(123)
	req.InnerBase2.Uint64 = uint64(123)
	req.InnerBase2.Int32 = int32(123)
	req.InnerBase2.SInt64 = int64(123)
	req.InnerBase2.Double = float64(22.3)
	req.InnerBase2.String_ = "hello_inner"
	req.InnerBase2.ListInt32 = []int32{12, 13, 14, 15, 16, 17}
	req.InnerBase2.MapStringString = map[string]string{"m1": "aaa", "m2": "bbb"}
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
	return req
}

func getBasicExampleReq() *base.BasicExample {
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

func buildExampleJSONData() error {
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
	absoluteExampleJSONPath := util_test.MustGitPath(exampleJSON)
	if checkExist(absoluteExampleJSONPath) == true {
		if err = os.Remove(absoluteExampleJSONPath); err != nil {
			panic("delete protoJSONFile failed")
		}
	}
	file, err = os.Create(absoluteExampleJSONPath)
	if err != nil {
		panic("create protoJSONFile failed")
	}
	defer file.Close()
	if _, err := file.WriteString(string(data)); err != nil {
		panic("write protoJSONData failed")
	}
	return nil
}

func buildBasicExampleJSONData() error {
	req := getBasicExampleReq()
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
	absoluteExampleJSONPath := util_test.MustGitPath(basicExampleJSON)
	if checkExist(absoluteExampleJSONPath) == true {
		if err = os.Remove(absoluteExampleJSONPath); err != nil {
			panic("delete protoJSONFile failed")
		}
	}
	file, err = os.Create(absoluteExampleJSONPath)
	if err != nil {
		panic("create protoJSONFile failed")
	}
	defer file.Close()
	if _, err := file.WriteString(string(data)); err != nil {
		panic("write protoJSONData failed")
	}
	return nil
}

func getExampleInt2Float() *proto.TypeDescriptor {
	includeDirs := util_test.MustGitPath("testdata/idl/") // includeDirs is used to find the include files.
	svc, err := proto.NewDescritorFromPath(context.Background(), util_test.MustGitPath(exampleIDLPath), includeDirs)
	if err != nil {
		panic(err)
	}
	return (*svc).LookupMethodByName("Int2FloatMethod").Output()
}

func TestUnknownFields(t *testing.T) {
	t.Run("disallow", func(t *testing.T) {
		desc := getExampleDesc()
		data := getExampleData()
		root := ast.NewRaw(string(data))
		root.SetAny("UNKNOWN", 1)
		data, _ = root.MarshalJSON()
		cv := NewBinaryConv(conv.Options{
			DisallowUnknownField: true,
		})
		ctx := context.Background()
		_, err := cv.Do(ctx, desc, data)
		require.Error(t, err)
	})
	t.Run("allow", func(t *testing.T) {
		desc := getExampleDesc()
		data := getExampleData()
		root := ast.NewRaw(string(data))
		root.Set("UNKNOWN-null", ast.NewNull())
		root.SetAny("UNKNOWN-num", 1)
		root.SetAny("UNKNOWN-bool", true)
		root.SetAny("UNKNOWN-string", "a")
		root.Set("UNKNOWN-object", ast.NewRaw(`{"1":1}`))
		root.Set("UNKNOWN-array", ast.NewRaw(`[1]`))
		base := root.Get("InnerBase2")
		base.SetAny("UNKNOWN-sub", 1)
		data, _ = root.MarshalJSON()
		println(string(data))
		cv := NewBinaryConv(conv.Options{
			DisallowUnknownField: false,
		})
		ctx := context.Background()
		out, err := cv.Do(ctx, desc, data)
		require.NoError(t, err)
		exp := &example2.ExampleReq{}
		// unmarshal target struct
		err = json.Unmarshal(data, exp)
		require.NoError(t, err)
		act := &example2.ExampleReq{}
		l := 0
		// fmt.Print(out)
		dataLen := len(out)
		// fastRead to get target struct
		for l < dataLen {
			id, wtyp, tagLen := goprotowire.ConsumeTag(out)
			if tagLen < 0 {
				t.Fatal("test failed")
			}
			l += tagLen
			out = out[tagLen:]
			offset, err := act.FastRead(out, int8(wtyp), int32(id))
			require.Nil(t, err)
			out = out[offset:]
			l += offset
		}
		require.Nil(t, err)
		require.Equal(t, exp, act)
	})
}

func TestFloat2Int(t *testing.T) {
	t.Run("double2Int", func(t *testing.T) {
		desc := getExampleInt2Float()
		data := []byte(`{"Int32":2.229e+2}`)
		cv := NewBinaryConv(conv.Options{})
		ctx := context.Background()
		out, err := cv.Do(ctx, desc, data)
		require.NoError(t, err)
		exp := example2.ExampleInt2Float{}
		l := 0
		// fmt.Print(out)
		dataLen := len(out)
		// fastRead to get target struct
		for l < dataLen {
			id, wtyp, tagLen := goprotowire.ConsumeTag(out)
			if tagLen < 0 {
				t.Fatal("test failed")
			}
			l += tagLen
			out = out[tagLen:]
			offset, err := exp.FastRead(out, int8(wtyp), int32(id))
			require.Nil(t, err)
			out = out[offset:]
			l += offset
		}
		require.Nil(t, err)
		require.Equal(t, exp.Int32, int32(222))
	})
	t.Run("int2double", func(t *testing.T) {
		desc := getExampleInt2Float()
		data := []byte(`{"Float64":` + strconv.Itoa(math.MaxInt64) + `}`)
		cv := NewBinaryConv(conv.Options{})
		ctx := context.Background()
		out, err := cv.Do(ctx, desc, data)
		require.NoError(t, err)
		exp := example2.ExampleInt2Float{}
		l := 0
		// fmt.Print(out)
		dataLen := len(out)
		// fastRead to get target struct
		for l < dataLen {
			id, wtyp, tagLen := goprotowire.ConsumeTag(out)
			if tagLen < 0 {
				t.Fatal("test failed")
			}
			l += tagLen
			out = out[tagLen:]
			offset, err := exp.FastRead(out, int8(wtyp), int32(id))
			require.Nil(t, err)
			out = out[offset:]
			l += offset
		}
		require.Nil(t, err)
		require.Equal(t, exp.Float64, float64(math.MaxInt64))
	})
}

func getEmptyExampleDesc() *proto.TypeDescriptor {
	opts := proto.Options{}
	includeDirs := util_test.MustGitPath("testdata/idl/") // includeDirs is used to find the include files.
	svc, err := opts.NewDescriptorFromPath(context.Background(), util_test.MustGitPath(exampleEmptyIDLPath), includeDirs)
	if err != nil {
		panic(err)
	}
	res := (*svc).LookupMethodByName("EmptyListMethodTest").Input()

	if res == nil {
		panic("can't find Target MessageDescriptor")
	}
	return res
}

func TestEmptyList(t *testing.T) {
	t.Run("empty unpacked list field with string", func(t *testing.T) {
		desc := getEmptyExampleDesc()
		data := []byte(`{"Msg": "msg", "Query": [], "Header": false, "Code": 0}`)
		cv := NewBinaryConv(conv.Options{})
		ctx := context.Background()
		out, err := cv.Do(ctx, desc, data)
		require.NoError(t, err)
		exp := example.ExampleEmptyReq{}
		l := 0
		// fmt.Print(out)
		dataLen := len(out)
		// fastRead to get target struct
		for l < dataLen {
			id, wtyp, tagLen := goprotowire.ConsumeTag(out)
			if tagLen < 0 {
				t.Fatal("test failed")
			}
			l += tagLen
			out = out[tagLen:]
			offset, err := exp.FastRead(out, int8(wtyp), int32(id))
			require.Nil(t, err)
			out = out[offset:]
			l += offset
		}
		require.Nil(t, err)
		exceptMsg := "msg"
		require.Equal(t, exp, example.ExampleEmptyReq{
			Msg:    exceptMsg,
			Header: false,
			Code:   0,
		})
	})

	t.Run("empty packed list field with int", func(t *testing.T) {
		desc := getEmptyExampleDesc()
		data := []byte(`{"Msg": "msg", "Query": ["xxx","123"], "Header": false, "Code": 0, "QueryInt": [], "QueryDouble": []}`)
		cv := NewBinaryConv(conv.Options{})
		ctx := context.Background()
		out, err := cv.Do(ctx, desc, data)
		require.NoError(t, err)
		exp := example.ExampleEmptyReq{}
		l := 0
		// fmt.Print(out)
		dataLen := len(out)
		// fastRead to get target struct
		for l < dataLen {
			id, wtyp, tagLen := goprotowire.ConsumeTag(out)
			if tagLen < 0 {
				t.Fatal("test failed")
			}
			l += tagLen
			out = out[tagLen:]
			offset, err := exp.FastRead(out, int8(wtyp), int32(id))
			require.Nil(t, err)
			out = out[offset:]
			l += offset
		}
		require.Nil(t, err)
		exceptMsg := "msg"
		require.Equal(t, exp, example.ExampleEmptyReq{
			Msg:    exceptMsg,
			Header: false,
			Code:   0,
			Query:  []string{"xxx", "123"},
		})
	})

	t.Run("empty packed list field with double", func(t *testing.T) {
		desc := getEmptyExampleDesc()
		data := []byte(`{"Msg": "msg", "Query": ["a","b"], "Header": false, "Code": 0, "QueryInt": [1,2,3], "QueryDouble": []}`)
		cv := NewBinaryConv(conv.Options{})
		ctx := context.Background()
		out, err := cv.Do(ctx, desc, data)
		require.NoError(t, err)
		exp := example.ExampleEmptyReq{}
		l := 0
		// fmt.Print(out)
		dataLen := len(out)
		// fastRead to get target struct
		for l < dataLen {
			id, wtyp, tagLen := goprotowire.ConsumeTag(out)
			if tagLen < 0 {
				t.Fatal("test failed")
			}
			l += tagLen
			out = out[tagLen:]
			offset, err := exp.FastRead(out, int8(wtyp), int32(id))
			require.Nil(t, err)
			out = out[offset:]
			l += offset
		}
		require.Nil(t, err)
		exceptMsg := "msg"
		require.Equal(t, exp, example.ExampleEmptyReq{
			Msg:      exceptMsg,
			Header:   false,
			Code:     0,
			Query:    []string{"a", "b"},
			QueryInt: []int32{1, 2, 3},
		})
	})

	t.Run("full list", func(t *testing.T) {
		desc := getEmptyExampleDesc()
		data := []byte(`{"Msg": "msg", "Query": ["aaa","bbb"], "Header": false, "Code": 0, "QueryInt": [1,2,3], "QueryDouble": [1.1,2.2,3.3]}`)
		cv := NewBinaryConv(conv.Options{})
		ctx := context.Background()
		out, err := cv.Do(ctx, desc, data)
		require.NoError(t, err)
		exp := example.ExampleEmptyReq{}
		l := 0
		// fmt.Print(out)
		dataLen := len(out)
		// fastRead to get target struct
		for l < dataLen {
			id, wtyp, tagLen := goprotowire.ConsumeTag(out)
			if tagLen < 0 {
				t.Fatal("test failed")
			}
			l += tagLen
			out = out[tagLen:]
			offset, err := exp.FastRead(out, int8(wtyp), int32(id))
			require.Nil(t, err)
			out = out[offset:]
			l += offset
		}
		require.Nil(t, err)
		exceptMsg := "msg"
		require.Equal(t, exp, example.ExampleEmptyReq{
			Msg:         exceptMsg,
			Header:      false,
			Code:        0,
			Query:       []string{"aaa", "bbb"},
			QueryInt:    []int32{1, 2, 3},
			QueryDouble: []float64{1.1, 2.2, 3.3},
		})
	})

	t.Run("empty list field inner struct", func(t *testing.T) {
		desc := getExampleDesc()
		data := []byte(`{"Msg":"hello","InnerBase2":{"Uint32":1,"ListInt32":[]}}`)
		cv := NewBinaryConv(conv.Options{})
		ctx := context.Background()
		out, err := cv.Do(ctx, desc, data)
		require.NoError(t, err)
		exp := example2.ExampleReq{}
		l := 0
		// fmt.Print(out)
		dataLen := len(out)
		// fastRead to get target struct
		for l < dataLen {
			id, wtyp, tagLen := goprotowire.ConsumeTag(out)
			if tagLen < 0 {
				t.Fatal("test failed")
			}
			l += tagLen
			out = out[tagLen:]
			offset, err := exp.FastRead(out, int8(wtyp), int32(id))
			require.Nil(t, err)
			out = out[offset:]
			l += offset
		}
		require.Nil(t, err)
		require.Equal(t, exp, example2.ExampleReq{
			Msg: "hello",
			InnerBase2: &example2.InnerBase2{
				Uint32: 1,
			},
		})
	})
}
