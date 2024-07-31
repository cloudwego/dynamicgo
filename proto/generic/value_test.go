package generic

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"testing"

	"github.com/cloudwego/dynamicgo/internal/util_test"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/base"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example3"
	"github.com/stretchr/testify/require"
	goprotowire "google.golang.org/protobuf/encoding/protowire"
	goproto "google.golang.org/protobuf/proto"
)

const (
	exampleIDLPath   = "../../testdata/idl/example3.proto"
	exampleProtoPath = "../../testdata/data/example3_pb.bin"
)

func init() {
	generateBinaryData() // for generating exampleProtoPath
}

// parse protofile to get MessageDescriptor
func getExample3Desc() *proto.TypeDescriptor {
	includeDirs := util_test.MustGitPath("testdata/idl/") // includeDirs is used to find the include files.
	svc, err := proto.NewDescritorFromPath(context.Background(), exampleIDLPath, includeDirs)
	if err != nil {
		panic(err)
	}
	res := svc.LookupMethodByName("ExampleMethod").Input()

	if res == nil {
		panic("can't find Target MessageDescriptor")
	}
	return res
}

func getExamplePartialDesc() *proto.TypeDescriptor {
	includeDirs := util_test.MustGitPath("testdata/idl/") // includeDirs is used to find the include files.
	svc, err := proto.NewDescritorFromPath(context.Background(), exampleIDLPath, includeDirs)
	if err != nil {
		panic(err)
	}
	res := svc.LookupMethodByName("ExamplePartialMethod").Input()

	if res == nil {
		panic("can't find Target MessageDescriptor")
	}
	return res
}

func getExample3Data() []byte {
	out, err := ioutil.ReadFile(exampleProtoPath)
	if err != nil {
		panic(err)
	}
	return out
}

func getExample3Req() *example3.ExampleReq {
	req := example3.ExampleReq{}
	req.Msg = "hello"
	req.Subfix = math.MaxFloat64
	req.InnerBase2 = &example3.InnerBase2{}
	req.InnerBase2.Bool = true
	req.InnerBase2.Uint32 = uint32(123)
	req.InnerBase2.Uint64 = uint64(123)
	req.InnerBase2.Double = float64(22.3)
	req.InnerBase2.String_ = "hello_inner"
	req.InnerBase2.ListInt32 = []int32{12, 13, 14, 15, 16, 17}
	req.InnerBase2.MapStringString = map[string]string{"m1": "aaa", "m2": "bbb", "m3": "ccc", "m4": "ddd"}
	req.InnerBase2.SetInt32 = []int32{200, 201, 202, 203, 204, 205}
	req.InnerBase2.Foo = example3.FOO_FOO_A
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

	req.InnerBase2.Sfixed32 = int32(100)
	req.InnerBase2.Fixed64 = uint64(200)
	req.InnerBase2.Sint32 = int32(300)
	req.InnerBase2.Sint64 = int64(400)
	req.InnerBase2.ListSInt64 = []int64{100, 200, 300}
	req.InnerBase2.ListSInt32 = []int32{121, 400, 514}
	req.InnerBase2.ListSfixed32 = []int32{201, 31, 22}
	req.InnerBase2.ListFixed64 = []uint64{841, 23, 11000}
	req.InnerBase2.MapInt64Sfixed64 = map[int64]int64{2: 2, 3: 3, 100: -100}
	req.InnerBase2.MapStringFixed32 = map[string]uint32{"1": 1, "2": 222222}
	return &req
}

// build binaryData for example3.proto
func generateBinaryData() error {
	req := getExample3Req()
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
	if checkExist(exampleProtoPath) == true {
		if err := os.Remove(exampleProtoPath); err != nil {
			panic("remove exampleProtoPath failed")
		}
	}
	file, err = os.Create(exampleProtoPath)
	if err != nil {
		panic("create protoBinaryFile failed")
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
	t.Run("Normal", func(t *testing.T) {
		desc := getExample3Desc()
		data := getExample3Data()
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
	})

	// stored unknown node
	t.Run("Partial", func(t *testing.T) {
		desc := getExamplePartialDesc()
		data := getExample3Data()
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
	})
}

func countHelper(count *int, ps []PathNode) {
	*count += len(ps)
	for _, p := range ps {
		countHelper(count, p.Next)
	}
}

func TestMarshalTo(t *testing.T) {
	desc := getExample3Desc()
	data := getExample3Data()
	partial := getExamplePartialDesc()

	exp := example3.ExampleReq{}
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
		offset, err := exp.FastRead(data, int8(wtyp), int32(id))
		require.Nil(t, err)
		data = data[offset:]
		l += offset
	}
	if len(data) != 0 {
		t.Fatal("test failed")
	}

	t.Run("ById", func(t *testing.T) {
		t.Run("TestMapStringString", func(t *testing.T) {
			opts := &Options{}
			buf, err := v.MarshalTo(partial, opts)
			require.Nil(t, err)
			ep := example3.ExampleReqPartial{}
			bufLen := len(buf)

			l := 0
			for l < bufLen {
				id, wtyp, tagLen := goprotowire.ConsumeTag(buf)
				if tagLen < 0 {
					t.Fatal("test failed")
				}
				l += tagLen
				buf = buf[tagLen:]
				offset, err := ep.FastRead(buf, int8(wtyp), int32(id))
				require.Nil(t, err)
				buf = buf[offset:]
				l += offset
			}
			if len(buf) != 0 {
				t.Fatal("test failed")
			}

			act := toInterface(ep)
			exp := toInterface(exp)
			require.False(t, deepEqual(act, exp))
			handlePartialMapStringString2(act.(map[int]interface{})[3].(map[int]interface{}))
			require.True(t, deepEqual(act, exp))
			require.Nil(t, ep.InnerBase2.MapStringString2)
		})
	})

	t.Run("unknown", func(t *testing.T) {
		data := getExample3Data()
		v := NewRootValue(desc, data)
		exist, err := v.SetByPath(NewNodeString("Insert"), NewPathFieldId(1024))
		require.False(t, exist)
		require.Nil(t, err)
		data = v.raw()
		v.Node = NewNode(proto.MESSAGE, data)
		t.Run("allow unknown", func(t *testing.T) {
			opts := &Options{DisallowUnknown: false}
			buf, err := v.MarshalTo(partial, opts)
			require.NoError(t, err)
			ep := example3.ExampleReqPartial{}
			bufLen := len(buf)

			l := 0
			for l < bufLen {
				id, wtyp, tagLen := goprotowire.ConsumeTag(buf)
				if tagLen < 0 {
					t.Fatal("test failed")
				}
				l += tagLen
				buf = buf[tagLen:]
				offset, err := ep.FastRead(buf, int8(wtyp), int32(id))
				require.Nil(t, err)
				buf = buf[offset:]
				l += offset
			}
			if len(buf) != 0 {
				t.Fatal("test failed")
			}

			act := toInterface(ep)
			exp := toInterface(exp)
			require.False(t, deepEqual(act, exp))
			handlePartialMapStringString2(act.(map[int]interface{})[3].(map[int]interface{}))
			require.True(t, deepEqual(act, exp))
		})
		t.Run("disallow unknown", func(t *testing.T) {
			opts := &Options{DisallowUnknown: true}
			_, err := v.MarshalTo(partial, opts)
			require.Error(t, err)
		})
	})
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

func TestGet(t *testing.T) {
	desc := getExample3Desc()
	data := getExample3Data()
	exp := example3.ExampleReq{}
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
		offset, err := exp.FastRead(data, int8(wtyp), int32(id))
		require.Nil(t, err)
		data = data[offset:]
		l += offset
	}

	if len(data) != 0 {
		t.Fatal("test failed")
	}

	req := getExample3Req()
	t.Run("GetByStr()", func(t *testing.T) {
		v := v.GetByPath(PathExampleMapStringString...)
		require.Nil(t, v.Check())
		v1, err := v.GetByStr("m1").String()
		require.NoError(t, err)
		require.Equal(t, req.InnerBase2.MapStringString["m1"], v1)
		v2, err := v.GetByStr("m8").String()
		require.Error(t, err)
		require.Equal(t, meta.ErrNotFound, err.(Node).ErrCode())
		require.Equal(t, req.InnerBase2.MapStringString["m8"], v2)
	})

	t.Run("GetByInt()", func(t *testing.T) {
		v := v.GetByPath(PathExampleMapInt32String...)
		require.Nil(t, v.Check())
		v1, err := v.GetByInt(1).String()
		require.NoError(t, err)
		require.Equal(t, req.InnerBase2.MapInt32String[1], v1)
		v2, err := v.GetByInt(999).String()
		require.Error(t, err)
		require.Equal(t, meta.ErrNotFound, err.(Node).ErrCode())
		require.Equal(t, req.InnerBase2.MapInt32String[999], v2)
	})

	t.Run("Index()", func(t *testing.T) {
		v := v.GetByPath(PathExampleListInt32...)
		require.Nil(t, v.Check())
		v1, err := v.Index(1).Int()
		require.NoError(t, err)
		require.Equal(t, int(req.InnerBase2.ListInt32[1]), v1)
		v2 := v.Index(999)
		require.Error(t, v2)
		require.Equal(t, meta.ErrInvalidParam, v2.ErrCode())
	})

	t.Run("FieldByName()", func(t *testing.T) {
		_, err := v.FieldByName("Msg2").String()
		require.NotNil(t, err)
		s, err := v.FieldByName("Msg").String()
		require.Equal(t, exp.Msg, s)
	})

	t.Run("Field()", func(t *testing.T) {
		xx, err := v.Field(222).Int()
		require.NotNil(t, err)
		require.Equal(t, int(0), xx)
		a := v.Field(3)
		b, err := a.Field(1).Bool()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase2.Bool, b)
		c, err := a.Field(2).Uint()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase2.Uint32, uint32(c))
		d, err := a.Field(3).Uint()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase2.Uint64, uint64(d))

		e, err := a.Field(4).Int()
		require.NotNil(t, err)
		require.Equal(t, int(0), e)

		f, err := a.Field(5).Int()
		require.NotNil(t, err)
		require.Equal(t, int(0), f)

		g, err := a.Field(6).Float64()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase2.Double, float64(g))
		h, err := a.Field(7).String()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase2.String_, string(h))
		list := a.Field(8)
		v, err := list.List(&Options{})
		require.Nil(t, err)
		// vint32 := make([]int32, 0, 6)
		// for _, vv := range v {
		// 	if vvv, ok := vv.(int32); ok{
		// 		vint32 = append(vint32, vvv)
		// 	}
		// }
		// require.Equal(t, exp.InnerBase2.ListInt32, vint32)
		// checkHelper(t, exp.InnerBase2.ListInt32, list, "List") why error?
		deepEqual(exp.InnerBase2.ListInt32, v)

		list1, err := a.Field(8).Index(1).Int()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase2.ListInt32[1], int32(list1))
		mp := a.Field(9)
		vmp, err := mp.StrMap(&Options{})
		require.Nil(t, err)
		fmt.Println(vmp)
		// require.Equal(t, exp.InnerBase2.MapStringString, vmp)
		deepEqual(exp.InnerBase2.MapStringString, vmp)
		checkHelper(t, exp.InnerBase2.MapStringString, mp, "StrMap")
		mp1, err := a.Field(9).GetByStr("m1").String()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase2.MapStringString["m1"], (mp1))
		sp := a.Field(10)
		vsp, err := sp.List(&Options{})
		require.Nil(t, err)
		fmt.Println(vsp)
		// require.Equal(t, exp.InnerBase2.SetInt32, vsp)
		// checkHelper(t, exp.InnerBase2.SetInt32, vsp, "List")
		deepEqual(exp.InnerBase2.SetInt32, vsp)
		i, err := a.Field(11).Int()
		require.NotNil(t, err)
		require.Equal(t, 0, i)
		mp2, err := a.Field(12).GetByInt(2).String()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase2.MapInt32String[2], (mp2))
		s, err := a.Field(21).Index(2).String()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase2.ListString[2], s)
	})

	t.Run("GetByPath()", func(t *testing.T) {
		exp := req.InnerBase2.ListInt32[1]

		v1 := v.GetByPath(NewPathFieldId(proto.FieldNumber(3)), NewPathFieldId(8), NewPathIndex(1))
		if v1.Error() != "" {
			t.Fatal(v1.Error())
		}
		act, err := v1.Int()
		require.NoError(t, err)
		require.Equal(t, int(exp), act)

		v2 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(1))
		if v2.Error() != "" {
			t.Fatal(v2.Error())
		}
		require.Equal(t, v1, v2)
		v3 := v.GetByPath(NewPathFieldId(proto.FieldNumber(3)), NewPathFieldName("ListInt32"), NewPathIndex(1))
		if v3.Error() != "" {
			t.Fatal(v3.Error())
		}
		require.Equal(t, v1, v3)
		v4 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(8), NewPathIndex(1))
		if v4.Error() != "" {
			t.Fatal(v4.Error())
		}
		require.Equal(t, v1, v4)

		exp2 := req.InnerBase2.MapInt32String[2]
		v5 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(12), NewPathIntKey(2))
		if v5.Error() != "" {
			t.Fatal(v5.Error())
		}
		act2, err := v5.String()
		require.NoError(t, err)
		require.Equal(t, exp2, act2)

		v6 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapInt32String"), NewPathIntKey(2))
		if v6.Error() != "" {
			t.Fatal(v6.Error())
		}
		require.Equal(t, v5, v6)

		exp3 := req.InnerBase2.MapStringString["m1"]
		v7 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(9), NewPathStrKey("m1"))
		if v5.Error() != "" {
			t.Fatal(v7.Error())
		}
		act3, err := v7.String()
		require.NoError(t, err)
		require.Equal(t, exp3, act3)

		v8 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapStringString"), NewPathStrKey("m1"))
		if v8.Error() != "" {
			t.Fatal(v8.Error())
		}
		require.Equal(t, v8, v7)

		v9 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(9), NewPathStrKey("m8"))
		require.Error(t, v9.Check())

		v10 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(21), NewPathIndex(2))
		if v10.Error() != "" {
			t.Fatal(v10.Error())
		}
		act10, err := v10.String()
		require.NoError(t, err)
		require.Equal(t, req.InnerBase2.ListString[2], act10)

		v11 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(19), NewPathIndex(1))
		if v11.Error() != "" {
			t.Fatal(v11.Error())
		}
		act11, err := v11.Interface(&Options{})
		var exp11 interface{} = req.InnerBase2.ListBase[1]
		deepEqual(exp11, act11)
	})

	t.Run("TestGetfixedNumber", func(t *testing.T) {
		// fixed64
		exp := req.InnerBase2.Fixed64
		ve := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("Fixed64"))
		if ve.Error() != "" {
			t.Fatal(ve.Error())
		}
		act, err := ve.Uint()
		fmt.Println(exp)
		fmt.Println(act)
		require.NoError(t, err)
		require.Equal(t, uint(exp), act)

		// repeated sint64
		exp1 := req.InnerBase2.ListSInt64[1]
		v1 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(26), NewPathIndex(1))
		if v1.Error() != "" {
			t.Fatal(v1.Error())
		}
		act1, err := v1.Int()
		fmt.Println(exp1)
		fmt.Println(act1)
		require.NoError(t, err)
		require.Equal(t, int(exp1), act1)

		// repeated sint32
		exp2 := req.InnerBase2.ListSInt32[2]
		v2 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(27), NewPathIndex(2))
		if v2.Error() != "" {
			t.Fatal(v2.Error())
		}
		act2, err := v2.Int()
		fmt.Println(exp2)
		fmt.Println(act2)
		require.NoError(t, err)
		require.Equal(t, int(exp2), act2)

		// repeated sfixed32
		exp3 := req.InnerBase2.ListSfixed32[0]
		v3 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(28), NewPathIndex(0))
		if v3.Error() != "" {
			t.Fatal(v3.Error())
		}
		act3, err := v3.Int()
		fmt.Println(exp3)
		fmt.Println(act3)
		require.NoError(t, err)
		require.Equal(t, int(exp3), act3)

		// repeated fixed64
		exp4 := req.InnerBase2.ListFixed64[2]
		v4 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListFixed64"), NewPathIndex(2))
		if v4.Error() != "" {
			t.Fatal(v4.Error())
		}
		act4, err := v4.Uint()
		fmt.Println(exp4)
		fmt.Println(act4)
		require.NoError(t, err)
		require.Equal(t, uint(exp4), act4)

		// MapInt64Sfixed64
		exp5 := req.InnerBase2.MapInt64Sfixed64[100]
		v5 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapInt64Sfixed64"), NewPathIntKey(100))
		act5, err := v5.int()
		fmt.Println(exp5)
		fmt.Println(act5)
		require.NoError(t, err)
		require.Equal(t, int(exp5), act5)

		// MapStringFixed32
		exp6 := req.InnerBase2.MapStringFixed32["2"]
		v6 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapStringFixed32"), NewPathStrKey("2"))
		act6, err := v6.Uint()
		fmt.Println(exp6)
		fmt.Println(act6)
		require.NoError(t, err)
		require.Equal(t, uint(exp6), act6)

		// Sfixed32
		exp7 := req.InnerBase2.Sfixed32
		v7 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("Sfixed32"))
		if v7.Error() != "" {
			t.Fatal(v7.Error())
		}
		act7, err := v7.Int()
		fmt.Println(exp7)
		fmt.Println(act7)
		require.NoError(t, err)
		require.Equal(t, int(exp7), act7)
	})
}

func TestSetByPath(t *testing.T) {
	desc := getExample3Desc()
	data := getExample3Data()
	v := NewRootValue(desc, data)
	v2 := NewNode(proto.MESSAGE, data)
	d2 := desc.Message().ByName("InnerBase2").Message().ByName("Base").Message().ByName("Extra").MapValue()
	e, err := v.SetByPath(v2)
	require.True(t, e)
	require.Nil(t, err)

	t.Run("replace", func(t *testing.T) {
		s := v.GetByPath(NewPathFieldName("Subfix"))
		require.Empty(t, s.Error())
		f, _ := s.Float64()
		require.Equal(t, math.MaxFloat64, f)
		exp := float64(-0.1)
		e, err := v.SetByPath(NewNodeDouble(exp), NewPathFieldName("Subfix"))
		require.True(t, e)
		require.Nil(t, err)
		s = v.GetByPath(NewPathFieldName("Subfix"))
		require.Empty(t, s.Error())
		f, _ = s.Float64()
		require.Equal(t, exp, f)

		exp2 := "中文"
		p := binary.NewBinaryProtocolBuffer()
		p.WriteString(exp2)
		vx := NewValue(d2, p.Buf)
		e, err2 := v.SetByPath(vx.Node, NewPathFieldName("InnerBase2"), NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("1b"))
		require.True(t, e)
		require.Nil(t, err2)
		s2 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("1b"))
		require.Empty(t, s2.Error())
		f2, _ := s2.String()
		require.Equal(t, exp2, f2)
		e, err2 = v.SetByPath(NewNode(proto.STRING, p.Buf), NewPathFieldName("InnerBase2"), NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("2b"))
		require.True(t, e)
		require.Nil(t, err2)
		s2 = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("2b"))
		require.Empty(t, s2.Error())
		f2, _ = s2.String()
		require.Equal(t, exp2, f2)

		exp3 := int32(math.MinInt32) + 1
		// v3 := Value{NewNodeInt32(exp3), nil, &d3}
		v3 := NewNodeInt32(exp3)
		ps := []Path{NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(2)}
		parent := v.GetByPath(ps[:len(ps)-1]...)
		l3, err := parent.Len()
		require.Nil(t, err)
		e, err = v.SetByPath(v3, ps...)
		require.True(t, e)
		require.Nil(t, err)
		s3 := v.GetByPath(ps...)
		act3, _ := s3.Int()
		require.Equal(t, exp3, int32(act3))
		parent = v.GetByPath(ps[:len(ps)-1]...)
		l3a, err := parent.Len()
		require.Nil(t, err)
		require.Equal(t, l3, l3a)
	})

	t.Run("insert", func(t *testing.T) {
		s := v.GetByPath(NewPathFieldName("A"))
		require.True(t, s.IsErrNotFound())
		exp := int32(-1024)
		// v1 := Value{NewNodeInt32(exp), nil, &d3}
		v1 := NewNodeInt32(exp)
		e, err := v.SetByPath(v1, NewPathFieldName("A"))
		require.False(t, e)
		require.Nil(t, err)
		s = v.GetByPath(NewPathFieldName("A"))
		require.Empty(t, s.Error())
		act, _ := s.Int()
		require.Equal(t, exp, int32(act))

		s2 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("x"))
		require.True(t, s2.IsErrNotFound())
		exp2 := "中文xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\bb"
		// v2 := Value{NewNodeString(exp2), nil, &d3}
		v2 := NewNodeString(exp2)
		e, err2 := v.SetByPath(v2, NewPathFieldName("InnerBase2"), NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("x"))
		require.False(t, e)
		require.Nil(t, err2)
		s2 = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("x"))
		require.Empty(t, s2.Error())
		act2, _ := s2.String()
		require.Equal(t, exp2, act2)

		parent := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"))
		l3, err := parent.Len()
		require.Nil(t, err)
		s3 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(1024))
		require.True(t, s3.IsErrNotFound())
		exp3 := rand.Int31()
		// v3 := Value{NewNodeInt32(exp3), nil, &d3}
		v3 := NewNodeInt32(exp3)
		e, err = v.SetByPath(v3, NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(1024))
		require.False(t, e)
		require.NoError(t, err)
		s3 = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(6)) // 6 + 1
		act3, _ := s3.Int()
		require.Equal(t, exp3, int32(act3))
		parent = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"))
		l3a, err := parent.Len()
		require.Nil(t, err)
		require.Equal(t, l3+1, l3a)
		exp3 = rand.Int31()
		// v3 = Value{NewNodeInt32(exp3), nil, &d3}
		v3 = NewNodeInt32(exp3)
		e, err = v.SetByPath(v3, NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(1024))
		require.False(t, e)
		require.NoError(t, err)
		s3 = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(7))
		act3, _ = s3.Int()
		require.Equal(t, exp3, int32(act3))
		parent = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"))
		l3a, err = parent.Len()
		require.Nil(t, err)
		require.Equal(t, l3+2, l3a)
		exp4 := "hello world!"
		p := binary.NewBinaryProtocolBuffer()
		p.WriteString(exp4)
		// v4 := Value{NewNode(proto.STRING, p.Buf), nil, &d4}
		v4 := NewNode(proto.STRING, p.Buf)
		e, err = v.SetByPath(v4, NewPathFieldName("InnerBase2"), NewPathFieldName("ListString"), NewPathIndex(1024))
		require.False(t, e)
		require.NoError(t, err)
		s4 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListString"), NewPathIndex(6))
		act4, _ := s4.String()
		deepEqual(exp4, act4)
	})
}

func TestUnsetByPath(t *testing.T) {
	desc := getExample3Desc()
	data := getExample3Data()
	r := NewRootValue(desc, data)
	req := getExample3Req()
	v := r.Fork()
	err := v.UnsetByPath()
	require.Nil(t, err)

	// test UnsetByPath without varint_length change
	t.Run("no_parent_length_change", func(t *testing.T) {
		v = r.Fork()
		n := v.GetByPath(NewPathFieldName("Msg"))
		exp, _ := n.String()
		require.False(t, n.IsError())
		err = v.UnsetByPath(NewPathFieldName("Msg"))
		require.NoError(t, err)
		n = v.GetByPath(NewPathFieldName("Msg"))
		require.True(t, n.IsErrNotFound())
		p := binary.NewBinaryProtocolBuffer()
		p.WriteString(exp)
		// Msg := Value{NewNode(proto.STRING, p.Buf), nil, &d1}
		Msg := NewNode(proto.STRING, p.Buf)
		exist, _ := v.SetByPath(Msg, NewPathFieldName("Msg"))
		require.False(t, exist)
		n = v.GetByPath(NewPathFieldName("Msg"))
		act, _ := n.String()
		require.Equal(t, exp, act)
		v = r.Fork()
		n = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("String"))
		require.False(t, n.IsError())
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("String"))
		require.NoError(t, err)
	})

	// test UnsetByPath with varint_length change
	t.Run("parent_length_change", func(t *testing.T) {
		v = r.Fork()
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(8))
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(9))
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(10))
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(11))
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(20))
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(13))
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(14))
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(15))
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(16))
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(17))
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(18))
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(19))
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(21))
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(255))
		n := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(12))
		require.False(t, n.IsError())
		x, _ := n.IntMap(&Options{})
		deepEqual(x, req.InnerBase2.MapInt32String)
	})

	t.Run("delete_list_index", func(t *testing.T) {
		v = r.Fork()
		n := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		require.False(t, n.IsError())
		exp, _ := n.Int()
		fmt.Println(exp)
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		require.NoError(t, err)

		k := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"))
		fmt.Println(k.Len())

		n = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		require.False(t, n.IsError())
		exp, _ = n.Int()
		fmt.Println(exp)
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		require.NoError(t, err)

		n = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		require.False(t, n.IsError())
		exp, _ = n.Int()
		fmt.Println(exp)
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		require.NoError(t, err)

		n = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		require.False(t, n.IsError())
		exp, _ = n.Int()
		fmt.Println(exp)
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		require.NoError(t, err)

		n = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		require.False(t, n.IsError())
		exp, _ = n.Int()
		fmt.Println(exp)
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		require.NoError(t, err)

		n = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		require.False(t, n.IsError())
		exp, _ = n.Int()
		fmt.Println(exp)
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		require.NoError(t, err)

		n = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		require.True(t, n.IsError())

		n = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapStringString"), NewPathStrKey("m1"))
		act, _ := n.string()
		require.Equal(t, "aaa", act)
	})
	t.Run("delete_list", func(t *testing.T) {
		v = r.Fork()
		n := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListBase"))
		require.False(t, n.IsError())
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListBase"))
		require.NoError(t, err)
		n = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListBase"))
		require.True(t, n.IsErrNotFound())
		n = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListString"))
		require.False(t, n.IsError())
		act, _ := n.List(&Options{})
		fmt.Println(act)
		n = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"))
		require.False(t, n.IsError())
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"))
		require.NoError(t, err)
		n = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"))
		require.True(t, n.IsErrNotFound())
		mp := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapStringString"))
		vmp, err := mp.StrMap(&Options{})
		require.Nil(t, err)
		fmt.Println(vmp)
	})

	t.Run("delete_map", func(t *testing.T) {
		v = r.Fork()
		n := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapStringString"))
		require.False(t, n.IsError())
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapStringString"))
		require.NoError(t, err)
		n = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapStringString"))
		require.True(t, n.IsErrNotFound())
		n = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapInt32String"))
		require.False(t, n.IsError())
		len1, err := n.Len()
		require.NoError(t, err)
		act, _ := n.IntMap(&Options{}) // TODO: can not convert map[interface{}]interface{} to map[int]interface{}
		fmt.Println(act)
		deepEqual(req.InnerBase2.MapInt32String, act)

		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapInt32String"), NewPathIntKey(1))
		require.NoError(t, err)
		n2 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapInt32String"))
		require.False(t, n.IsError())
		len2, err := n2.Len()
		require.NoError(t, err)
		act, _ = n2.IntMap(&Options{})
		fmt.Println(act)
		require.Equal(t, len1-1, len2)

		n3 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapUint32String"), NewPathIntKey(1))
		require.False(t, n3.IsError())
		act3, _ := n3.String()
		require.Equal(t, req.InnerBase2.MapUint32String[1], act3)
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapUint32String"), NewPathIntKey(1))
		require.NoError(t, err)
		mp4 := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapUint32String"))
		require.False(t, mp4.IsError())
		v, err := mp4.IntMap(&Options{})
		require.NoError(t, err)
		fmt.Println(v)
		deepEqual(v, req.InnerBase2.MapUint32String)
	})

	t.Run("delete_field", func(t *testing.T) {
		v = r.Fork()
		bytelen := v.l
		n := v.GetByPath(NewPathFieldName("Msg"))
		require.False(t, n.IsError())
		field1Len := n.l
		err := v.UnsetByPath(NewPathFieldName("Msg"))
		require.NoError(t, err)
		require.Equal(t, bytelen-field1Len-1, v.l) // 1 is the calculated varint taglen of Msg
		n = v.GetByPath(NewPathFieldName("InnerBase2"))
		require.False(t, n.IsError())
		field2Len := n.l
		err = v.UnsetByPath(NewPathFieldName("InnerBase2"))
		require.NoError(t, err)
		require.Equal(t, bytelen-field1Len-field2Len-1-1, v.l)
		fmt.Println(v.l)
		n = v.GetByPath(NewPathFieldName("Subfix"))
		require.False(t, n.IsError())
		act, _ := n.Float64()
		require.Equal(t, math.MaxFloat64, act)
	})

}

func TestSetMany(t *testing.T) {
	desc := getExample3Desc()
	data := getExample3Data()
	opts := &Options{
		UseNativeSkip: true,
	}
	r := NewRootValue(desc, data)
	d1 := desc.Message().ByName("Msg").Type()
	d2 := desc.Message().ByName("Subfix").Type()
	address := make([]int, 0)
	pathes := make([]Path, 0)
	PathInnerBase = NewPathFieldName("InnerBase2")
	PathExampleListInt32 = []Path{PathInnerBase, NewPathFieldId(proto.FieldNumber(8))}

	t.Run("insert", func(t *testing.T) {
		v := r.Fork()
		exp1 := int32(-1024)
		n1 := NewNodeInt32(exp1)
		err := v.SetMany([]PathNode{
			{
				Path: NewPathFieldId(2),
				Node: n1,
			},
		}, opts, &v, address, pathes...)
		require.Nil(t, err)
		v1 := v.GetByPath(NewPathFieldName("A"))
		act1, err := v1.Int()
		require.NoError(t, err)
		require.Equal(t, exp1, int32(act1))

		exp21 := int32(math.MinInt32)
		exp22 := int32(math.MaxInt32)
		n21 := NewNodeInt32(exp21)
		n22 := NewNodeInt32(exp22)
		vv, listInt2root := v.GetByPathWithAddress(PathExampleListInt32...)
		l2, err := vv.Len()
		require.NoError(t, err)
		// the last value of path2root and address2root is only a flag not using real value
		path2root := []Path{NewPathFieldName("InnerBase2"), NewPathFieldId(proto.FieldNumber(8)), NewPathIndex(1024)}
		address2root := append(listInt2root, 0)

		err = vv.SetMany([]PathNode{
			{
				Path: NewPathIndex(1024),
				Node: n21,
			},
			{
				Path: NewPathIndex(1024),
				Node: n22,
			},
		}, opts, &v, address2root, path2root...)
		require.NoError(t, err)
		v21 := vv.GetByPath(NewPathIndex(6))
		act21, err := v21.Int()
		require.NoError(t, err)
		require.Equal(t, exp21, int32(act21))
		v22 := vv.GetByPath(NewPathIndex(7))
		act22, err := v22.Int()
		require.NoError(t, err)
		require.Equal(t, exp22, int32(act22))
		ll2, err := vv.Len()
		require.Nil(t, err)
		require.Equal(t, l2+2, ll2)

		vx, base2root := v.GetByPathWithAddress(NewPathFieldName("InnerBase2"), NewPathFieldName("ListBase"))
		vv = vx
		m1, e := vv.Len()
		require.Nil(t, e)
		// the last value of path2root and address2root is only a flag not using real value
		path2root = []Path{NewPathFieldName("InnerBase2"), NewPathFieldName("ListBase"), NewPathIndex(1024)}
		address2root = append(base2root, 0)
		err = vv.SetMany([]PathNode{
			{
				Path: NewPathIndex(1024),
				Node: v.FieldByName("InnerBase2").FieldByName("Base").Node,
			}}, opts, &v, address2root, path2root...)
		require.Nil(t, err)
		vx = v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListBase"))
		m2, e := vx.Len()
		require.Nil(t, e)
		require.Equal(t, m1+1, m2)
	})

	t.Run("replace", func(t *testing.T) {
		vRoot := r.Fork()
		exp1 := "exp1"
		exp2 := float64(-225.0001)
		p := binary.NewBinaryProtocolBuffer()
		p.WriteString(exp1)

		v1 := NewValue(d1, []byte(string(p.Buf)))
		p.Buf = p.Buf[:0]
		p.WriteDouble(exp2)
		v2 := NewValue(d2, []byte(string(p.Buf)))
		err := vRoot.SetMany([]PathNode{
			{
				Path: NewPathFieldId(1),
				Node: v1.Node,
			},
			{
				Path: NewPathFieldId(32767),
				Node: v2.Node,
			},
		}, opts, &vRoot, address, pathes...)
		require.Nil(t, err)

		exp := example3.ExampleReq{}
		// fast read
		dataLen := vRoot.l
		data := vRoot.raw()
		l := 0
		for l < dataLen {
			id, wtyp, tagLen := goprotowire.ConsumeTag(data)
			if tagLen < 0 {
				t.Fatal("test failed")
			}
			l += tagLen
			data = data[tagLen:]
			offset, err := exp.FastRead(data, int8(wtyp), int32(id))
			require.Nil(t, err)
			data = data[offset:]
			l += offset
		}

		require.Equal(t, exp.Msg, exp1)
		require.Equal(t, exp.Subfix, exp2)

		vx, base2root := vRoot.GetByPathWithAddress(NewPathFieldName("InnerBase2"), NewPathFieldName("Base"))
		vv := vx
		// the last value of path2root and address2root is only a flag not using real value
		path2root := []Path{NewPathFieldName("InnerBase2"), NewPathFieldName("Base"), NewPathFieldId(1024)}
		address2root := append(base2root, 0)
		err = vv.SetMany([]PathNode{
			{
				Path: NewPathFieldId(1),
				Node: vRoot.FieldByName("InnerBase2").FieldByName("Base").FieldByName("LogID").Node,
			},
			{
				Path: NewPathFieldId(2),
				Node: vRoot.FieldByName("InnerBase2").FieldByName("Base").FieldByName("Caller").Node,
			},
			{
				Path: NewPathFieldId(3),
				Node: vRoot.FieldByName("InnerBase2").FieldByName("Base").FieldByName("Addr").Node,
			},
			{
				Path: NewPathFieldId(4),
				Node: vRoot.FieldByName("InnerBase2").FieldByName("Base").FieldByName("Client").Node,
			},
			{
				Path: NewPathFieldId(5),
				Node: vRoot.FieldByName("InnerBase2").FieldByName("Base").FieldByName("TrafficEnv").Node,
			},
			{
				Path: NewPathFieldId(6),
				Node: vRoot.FieldByName("InnerBase2").FieldByName("Base").FieldByName("Extra").Node,
			},
		}, opts, &vRoot, address2root, path2root...)
		require.Nil(t, err)
		require.Equal(t, vx.raw(), vv.raw())

		inner, inner2root := vRoot.GetByPathWithAddress(NewPathFieldName("InnerBase2"))
		p = binary.NewBinaryProtocolBuffer()
		e1 := false
		p.WriteBool(e1)
		v1 = NewValue(d1, []byte(string(p.Buf)))
		p.Buf = p.Buf[:0]
		e2 := float64(-255.0001)
		p.WriteDouble(e2)
		v2 = NewValue(d1, []byte(string(p.Buf)))
		p.Buf = p.Buf[:0]

		// the last value of path2root and address2root is only a flag not using real value
		path2root = []Path{NewPathFieldName("InnerBase2"), NewPathFieldId(1024)}
		address2root = append(inner2root, 0)

		err = inner.SetMany([]PathNode{
			{
				Path: NewPathFieldId(1),
				Node: v1.Node,
			},
			{
				Path: NewPathFieldId(6),
				Node: v2.Node,
			},
			{
				Path: NewPathFieldId(255),
				Node: vx.Node,
			},
		}, opts, &vRoot, address2root, path2root...)
		require.Nil(t, err)

		expx := example3.InnerBase2{}
		// fast read
		data = inner.raw()
		byteLen, offset := goprotowire.ConsumeVarint(data)
		fmt.Println(byteLen)

		data = data[offset:]
		dataLen = len(data)
		l = 0
		for l < dataLen {
			id, wtyp, tagLen := goprotowire.ConsumeTag(data)
			if tagLen < 0 {
				t.Fatal("test failed")
			}
			l += tagLen
			data = data[tagLen:]
			offset, err := expx.FastRead(data, int8(wtyp), int32(id))
			require.Nil(t, err)
			data = data[offset:]
			l += offset
		}

		require.Equal(t, expx.Bool, e1)
		require.Equal(t, expx.Double, e2)
		require.Equal(t, expx.Base, exp.InnerBase2.Base)
	})

}
