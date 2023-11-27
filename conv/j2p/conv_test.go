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

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/internal/util_test"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/base"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example2"
	"github.com/stretchr/testify/require"
	goprotowire "google.golang.org/protobuf/encoding/protowire"
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
	exampleIDLPath   = "testdata/idl/example2.proto"
	exampleJSON      = "testdata/data/example2req.json"
	exampleProtoPath = "testdata/data/example2_pb.bin"
)

func TestConvJSON2Protobf(t *testing.T) {
	buildExampleJSONData()
	desc := getExampleDesc()
	data := getExampleData()
	pdata, _ := ioutil.ReadFile(util_test.MustGitPath(exampleProtoPath))
	fmt.Println(pdata)
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
	req.InnerBase2.Double = float64(22.3)
	req.InnerBase2.String_ = "hello_inner"
	req.InnerBase2.ListInt32 = []int32{12, 13, 14, 15, 16, 17}
	req.InnerBase2.MapStringString = map[string]string{"m1": "aaa", "m2": "bbb"}
	req.InnerBase2.SetInt32 = []int32{200, 201, 202, 203, 204, 205}
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

func getExampleInt2Float() *proto.TypeDescriptor {
	includeDirs := util_test.MustGitPath("testdata/idl/") // includeDirs is used to find the include files.
	svc, err := proto.NewDescritorFromPath(context.Background(), util_test.MustGitPath(exampleIDLPath), includeDirs)
	if err != nil {
		panic(err)
	}
	return (*svc).LookupMethodByName("Int2FloatMethod").Output()
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
