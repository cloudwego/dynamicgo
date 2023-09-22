package generic

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/base"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example2"
	"github.com/stretchr/testify/require"
	goprotowire "google.golang.org/protobuf/encoding/protowire"
	goproto "google.golang.org/protobuf/proto"
)

const (
	exampleIDLPath   = "../../testdata/idl/example2.proto"
	exampleProtoPath = "../../testdata/data/example2_pb.bin"
	// exampleSuperProtoPath = "../../testdata/data/example2super.bin"
)

// parse protofile to get MessageDescriptor
func getExample2Desc() *proto.MessageDescriptor {
	svc, err := proto.NewDescritorFromPath(context.Background(), exampleIDLPath)
	if err != nil {
		panic(err)
	}
	res := (*svc).Methods().ByName("ExampleMethod").Input()

	if res == nil {
		panic("can't find Target MessageDescriptor")
	}
	return &res
}

func getExamplePartialDesc() *proto.MessageDescriptor {
	svc, err := proto.NewDescritorFromPath(context.Background(), exampleIDLPath)
	if err != nil {
		panic(err)
	}
	res := (*svc).Methods().ByName("ExamplePartialMethod").Input()

	if res == nil {
		panic("can't find Target MessageDescriptor")
	}
	return &res
}

func getExample2Data() []byte {
	out, err := ioutil.ReadFile(exampleProtoPath)
	if err != nil {
		panic(err)
	}
	return out
}

// build binaryData for example2.proto
func generateBinaryData() error {
	req := example2.ExampleReq{}
	req.Msg = "hello"
	req.A = 25
	req.InnerBase2 = &example2.InnerBase2{}
	req.InnerBase2.Bool = true
	req.InnerBase2.Uint32 = uint32(123)
	req.InnerBase2.Uint64 = uint64(123)
	req.InnerBase2.Double = float64(22.3)
	req.InnerBase2.String_ = "hello_inner"
	req.InnerBase2.ListInt32 = []int32{12, 13, 14, 15, 16, 17}
	req.InnerBase2.MapStringString = map[string]string{"m1": "aaa", "m2": "bbb", "m3": "ccc", "m4": "ddd"}
	req.InnerBase2.SetInt32 = []int32{200, 201, 202, 203, 204, 205}
	req.InnerBase2.Foo = example2.FOO_FOO_A
	req.InnerBase2.MapInt32String = map[int32]string{1: "aaa", 2: "bbb", 3: "ccc", 4: "ddd"}
	req.InnerBase2.Binary = []byte{0x1, 0x2, 0x3, 0x4}
	req.InnerBase2.MapUint32String = map[uint32]string{uint32(1): "u32aa", uint32(2): "u32bb", uint32(3): "u32cc", uint32(4): "u32dd"}
	req.InnerBase2.MapUint64String = map[uint64]string{uint64(1): "u64aa", uint64(2): "u64bb", uint64(3): "u64cc", uint64(4): "u64dd"}
	req.InnerBase2.MapInt64String = map[int64]string{int64(1): "64aaa", int64(2): "64bbb", int64(3): "64ccc", int64(4): "64ddd"}
	req.InnerBase2.ListBase = []*base.Base{{
		LogID: "logId",
		Caller: "caller",
		Addr: "addr",
		Client: "client",
		TrafficEnv: &base.TrafficEnv{
			Open: false,
			Env: "env",
		},
		Extra: map[string]string{"1a": "aaa", "2a": "bbb", "3a": "ccc", "4a": "ddd"},
	}, {
		LogID: "logId2",
		Caller: "caller2",
		Addr: "addr2",
		Client: "client2",
		TrafficEnv: &base.TrafficEnv{
			Open: true,
			Env: "env2",
		},
		Extra: map[string]string{"1a": "aaa2", "2a": "bbb2", "3a": "ccc2", "4a": "ddd2"},
	}}
	req.InnerBase2.MapInt64Base = map[int64]*base.Base{int64(1): {
		LogID: "logId",
		Caller: "caller",
		Addr: "addr",
		Client: "client",
		TrafficEnv: &base.TrafficEnv{
			Open: false,
			Env: "env",
		},
		Extra: map[string]string{"1a": "aaa", "2a": "bbb", "3a": "ccc", "4a": "ddd"},
	}, int64(2): {
		LogID: "logId2",
		Caller: "caller2",
		Addr: "addr2",
		Client: "client2",
		TrafficEnv: &base.TrafficEnv{
			Open: true,
			Env: "env2",
		},
		Extra: map[string]string{"1a": "aaa2", "2a": "bbb2", "3a": "ccc2", "4a": "ddd2"},
	}}
	req.InnerBase2.MapStringBase = map[string]*base.Base{"1": {
		LogID: "logId",
		Caller: "caller",
		Addr: "addr",
		Client: "client",
		TrafficEnv: &base.TrafficEnv{
			Open: false,
			Env: "env",
		},
		Extra: map[string]string{"1a": "aaa", "2a": "bbb", "3a": "ccc", "4a": "ddd"},
	}, "2": {
		LogID: "logId2",
		Caller: "caller2",
		Addr: "addr2",
		Client: "client2",
		TrafficEnv: &base.TrafficEnv{
			Open: true,
			Env: "env2",
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
	if checkExist(exampleProtoPath) == false {
		file, err = os.Create(exampleProtoPath)
		if err != nil {
			panic("create protoBinaryFile failed")
		}
	} else {
		file, err = os.OpenFile(exampleProtoPath, os.O_RDWR, 0666)
		if err != nil {
			panic("open protoBinaryFile failed")
		}
	}
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		panic("write protoBinary data failed")
	}
	return nil
}

func TestCreateValue(t *testing.T) {
	generateBinaryData()
}

func TestCount(t *testing.T) {
	desc := getExample2Desc()
	data := getExample2Data()
	fmt.Printf("data len: %d\n", len(data))
	v := NewRootValue(desc, data)
	children := make([]PathNode, 0, 4)
	opts := Options{}
	if err := v.Children(&children, true, &opts, desc); err != nil {
		t.Fatal(err)
	}
	count := 1
	countHelper(&count, children)
	fmt.Printf("nodes count: %d", count)
}

func countHelper(count *int, ps []PathNode) {
	*count += len(ps)
	for _, p := range ps {
		countHelper(count, p.Next)
	}
}


func TestMarshalTo(t *testing.T) {
	desc := getExample2Desc()
	data := getExample2Data()
	partial := getExamplePartialDesc()

	exp := example2.ExampleReq{}
	v := NewRootValue(desc, data)
	dataLen := len(data)
	l := 0
	for l < dataLen {
		id, wtyp, tagLen := goprotowire.ConsumeTag(data)
		if tagLen < 0 {
			t.Fatal("test failed")
		}
		l += tagLen
		data = data[tagLen:]
		offset, err := exp.FastRead(data,int8(wtyp),int32(id))
		require.Nil(t, err)
		data = data[offset:]
		l += offset
	}
	if len(data) != 0 {
		t.Fatal("test failed")
	}

	t.Run("ById", func(t *testing.T) {
		t.Run("WriteDefault", func(t *testing.T) {
			opts := &Options{WriteDefault: true}
			buf, err := v.MarshalTo(partial, opts)
			require.Nil(t, err)
			ep := example2.ExampleReqPartial{}

			bufLen := len(buf)
			l := 0
			for l < bufLen {
				id, wtyp, tagLen := goprotowire.ConsumeTag(buf)
				if tagLen < 0 {
					t.Fatal("test failed")
				}
				l += tagLen
				buf = buf[tagLen:]
				offset, err := ep.FastRead(buf,int8(wtyp),int32(id))
				require.Nil(t, err)
				buf = buf[offset:]
				l += offset
			}
			if len(buf) != 0 {
				t.Fatal("test failed")
			}
			
			act := toInterface(ep)
			exp := toInterface(exp)
			require.False(t, DeepEqual(act, exp))
			handlePartialMapStringString2(act.(map[int]interface{})[3].(map[int]interface{}))
			require.True(t, DeepEqual(act, exp))
			// TODO
			// require.NotNil(t, ep.InnerBase2.MapStringString2)
		})
		t.Run("NotWriteDefault", func(t *testing.T) {
			opts := &Options{}
			buf, err := v.MarshalTo(partial, opts)
			require.Nil(t, err)
			ep := example2.ExampleReqPartial{}
			bufLen := len(buf)
			
			l := 0
			for l < bufLen {
				id, wtyp, tagLen := goprotowire.ConsumeTag(buf)
				if tagLen < 0 {
					t.Fatal("test failed")
				}
				l += tagLen
				buf = buf[tagLen:]
				offset, err := ep.FastRead(buf,int8(wtyp),int32(id))
				require.Nil(t, err)
				buf = buf[offset:]
				l += offset
			}
			if len(buf) != 0 {
				t.Fatal("test failed")
			}

			act := toInterface(ep)
			exp := toInterface(exp)
			require.False(t, DeepEqual(act, exp))
			handlePartialMapStringString2(act.(map[int]interface{})[3].(map[int]interface{}))
			require.True(t, DeepEqual(act, exp))
			require.Nil(t, ep.InnerBase2.MapStringString2)
		})
	})

	// TODO: test unknown

}

func handlePartialMapStringString2(p map[int]interface{}) {
	delete(p, 127)

	if f18 := p[18]; f18 != nil {
		for i := range f18.([]interface{}) {
			pv := f18.([]interface{})[i].(map[int]interface{})
			handlePartialMapStringString2(pv)
		}
	}

}