package generic

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"testing"

	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
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

func getExample2Req() *example2.ExampleReq {
	req := example2.ExampleReq{}
	req.Msg = "hello"
	req.A = 25
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
	req.InnerBase2.Foo = example2.FOO_FOO_A
	req.InnerBase2.MapInt32String = map[int32]string{1: "aaa", 2: "bbb", 3: "ccc", 4: "ddd"}
	req.InnerBase2.Binary = []byte{0x1, 0x2, 0x3, 0x4}
	req.InnerBase2.MapUint32String = map[uint32]string{uint32(1): "u32aa", uint32(2): "u32bb", uint32(3): "u32cc", uint32(4): "u32dd"}
	req.InnerBase2.MapUint64String = map[uint64]string{uint64(1): "u64aa", uint64(2): "u64bb", uint64(3): "u64cc", uint64(4): "u64dd"}
	req.InnerBase2.MapInt64String = map[int64]string{int64(1): "64aaa", int64(2): "64bbb", int64(3): "64ccc", int64(4): "64ddd"}
	req.InnerBase2.ListString = []string{"111", "222", "333", "44", "51", "6"}
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
	return &req
}

// build binaryData for example2.proto
func generateBinaryData() error {
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


func TestGet(t *testing.T) {
	desc := getExample2Desc()
	data := getExample2Data()
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

	req := getExample2Req()
	t.Run("GetByStr()", func(t *testing.T) {
		v,_ := v.GetByPath(PathExampleMapStringString...)
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
		v,_ := v.GetByPath(PathExampleMapInt32String...)
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
		v,_ := v.GetByPath(PathExampleListInt32...)
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
		DeepEqual(exp.InnerBase2.ListInt32, v)
		
		
		list1, err := a.Field(8).Index(1).Int()
		require.Nil(t, err)
		require.Equal(t, exp.InnerBase2.ListInt32[1], int32(list1))
		mp := a.Field(9)
		vmp, err := mp.StrMap(&Options{})
		require.Nil(t, err)
		fmt.Println(vmp)
		// require.Equal(t, exp.InnerBase2.MapStringString, vmp)
		DeepEqual(exp.InnerBase2.MapStringString, vmp)
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
		DeepEqual(exp.InnerBase2.SetInt32, vsp)
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

		v1,_ := v.GetByPath(NewPathFieldId(proto.FieldNumber(3)), NewPathFieldId(8), NewPathIndex(1))
		if v1.Error() != "" {
			t.Fatal(v1.Error())
		}
		act, err := v1.Int()
		require.NoError(t, err)
		require.Equal(t, int(exp), act)

		v2,_ := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("ListInt32"), NewPathIndex(1))
		if v2.Error() != "" {
			t.Fatal(v2.Error())
		}
		require.Equal(t, v1, v2)
		v3,_ := v.GetByPath(NewPathFieldId(proto.FieldNumber(3)), NewPathFieldName("ListInt32"), NewPathIndex(1))
		if v3.Error() != "" {
			t.Fatal(v3.Error())
		}
		require.Equal(t, v1, v3)
		v4,_ := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(8), NewPathIndex(1))
		if v4.Error() != "" {
			t.Fatal(v4.Error())
		}
		require.Equal(t, v1, v4)

		exp2 := req.InnerBase2.MapInt32String[2]
		v5,_ := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(12), NewPathIntKey(2))
		if v5.Error() != "" {
			t.Fatal(v5.Error())
		}
		act2, err := v5.String()
		require.NoError(t, err)
		require.Equal(t, exp2, act2)

		v6,_ := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapInt32String"), NewPathIntKey(2))
		if v6.Error() != "" {
			t.Fatal(v6.Error())
		}
		require.Equal(t, v5, v6)

		exp3 := req.InnerBase2.MapStringString["m1"]
		v7,_ := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(9), NewPathStrKey("m1"))
		if v5.Error() != "" {
			t.Fatal(v7.Error())
		}
		act3, err := v7.String()
		require.NoError(t, err)
		require.Equal(t, exp3, act3)

		v8,_ := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldName("MapStringString"), NewPathStrKey("m1"))
		if v8.Error() != "" {
			t.Fatal(v8.Error())
		}
		require.Equal(t, v8, v7)

		v9,_ := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(9), NewPathStrKey("m8"))
		require.Error(t, v9.Check())

		v10,_ := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(21), NewPathIndex(2))
		if v10.Error() != "" {
			t.Fatal(v10.Error())
		}
		act10, err := v10.String()
		require.NoError(t, err)
		require.Equal(t, req.InnerBase2.ListString[2], act10)

		v11,_ := v.GetByPath(NewPathFieldName("InnerBase2"), NewPathFieldId(19), NewPathIndex(1))
		if v11.Error() != "" {
			t.Fatal(v11.Error())
		}
		act11, err := v11.Interface(&Options{})
		var exp11 interface{} = req.InnerBase2.ListBase[1]
		DeepEqual(exp11, act11)
	})

}


func TestSetByPath(t *testing.T) {
	desc := getExample2Desc()
	data := getExample2Data()
	v := NewRootValue(desc, data)
	ds := (*desc).Fields().ByName("Subfix")
	d2 := (*desc).Fields().ByName("InnerBase2").Message().Fields().ByName("Base").Message().Fields().ByName("Extra").MapValue()
	d3 := (*desc).Fields().ByName("InnerBase2").Message().Fields().ByName("ListInt32")
	e, err := v.SetByPath(v)
	require.True(t, e)
	require.Nil(t, err)

	t.Run("replace", func(t *testing.T) {
		s, _ := v.GetByPath(NewPathFieldName("Subfix"))
		require.Empty(t, s.Error())
		f, _ := s.Float64()
		require.Equal(t, math.MaxFloat64, f)
		exp := float64(-0.1)
		e, err := v.SetByPath(Value{NewNodeDouble(exp),nil,&ds}, NewPathFieldName("Subfix"))
		require.True(t, e)
		require.Nil(t, err)
		s, _ = v.GetByPath(NewPathFieldName("Subfix"))
		require.Empty(t, s.Error())
		f, _ = s.Float64()
		require.Equal(t, exp, f)

		exp2 := "中文"
		p := binary.NewBinaryProtocolBuffer()
		p.WriteString(exp2)
		e, err2 := v.SetByPath(NewValue(&d2, p.Buf), NewPathFieldName("InnerBase2"), NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("1b"))
		require.True(t, e)
		require.Nil(t, err2)
		s2, _ := v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("b"))
		require.Empty(t, s2.Error())
		f2, _ := s2.String()
		require.Equal(t, exp2, f2)
		e, err2 = v.SetByPath(NewValue(&d2, p.Buf), NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("b"))
		require.True(t, e)
		require.Nil(t, err2)
		s2,_  = v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("b"))
		require.Empty(t, s2.Error())
		f2, _ = s2.String()
		require.Equal(t, exp2, f2)

		exp3 := int32(math.MinInt32) + 1
		v3 := Value{NewNodeInt32(exp3), nil, &d3}
		ps := []Path{NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"), NewPathIndex(2)}
		parent,_ := v.GetByPath(ps[:len(ps)-1]...)
		l3, err := parent.Len()
		require.Nil(t, err)
		e, err = v.SetByPath(v3, ps...)
		require.True(t, e)
		require.Nil(t, err)
		s3, _ := v.GetByPath(ps...)
		act3, _ := s3.Int()
		require.Equal(t, exp3, int32(act3))
		parent, _ = v.GetByPath(ps[:len(ps)-1]...)
		l3a, err := parent.Len()
		require.Nil(t, err)
		require.Equal(t, l3, l3a)
	})

	t.Run("insert", func(t *testing.T) {
		s,_ := v.GetByPath(NewPathFieldName("A"))
		require.True(t, s.IsErrNotFound())
		exp := int32(-1024)
		v1 := Value{NewNodeInt32(exp), nil, &d3}
		e, err := v.SetByPath(v1, NewPathFieldName("A"))
		require.False(t, e)
		require.Nil(t, err)
		s, _ = v.GetByPath(NewPathFieldName("A"))
		require.Empty(t, s.Error())
		act, _ := s.Int()
		require.Equal(t, exp, int32(act))

		s2, _ := v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("x"))
		require.True(t, s2.IsErrNotFound())
		exp2 := "中文\bb"
		v2 := Value{NewNodeString(exp2), nil, &d3}
		e, err2 := v.SetByPath(v2, NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("x"))
		require.False(t, e)
		require.Nil(t, err2)
		s2, _ = v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("x"))
		require.Empty(t, s2.Error())
		act2, _ := s2.String()
		require.Equal(t, exp2, act2)

		parent, _ := v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"))
		l3, err := parent.Len()
		require.Nil(t, err)
		s3, _ := v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"), NewPathIndex(1024))
		require.True(t, s3.IsErrNotFound())
		exp3 := rand.Int31()
		v3 := Value{NewNodeInt32(exp3), nil, &d3}
		e, err = v.SetByPath(v3, NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"), NewPathIndex(1024))
		require.False(t, e)
		require.NoError(t, err)
		s3, _ = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		act3, _ := s3.Int()
		require.Equal(t, exp3, int32(act3))
		parent, _ = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"))
		l3a, err := parent.Len()
		require.Nil(t, err)
		require.Equal(t, l3+1, l3a)
		exp3 = rand.Int31()
		v3 = Value{NewNodeInt32(exp3), nil, &d3}
		e, err = v.SetByPath(v3, NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"), NewPathIndex(1024))
		require.False(t, e)
		require.NoError(t, err)
		s3, _ = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"), NewPathIndex(0))
		act3, _ = s3.Int()
		require.Equal(t, exp3, int32(act3))
		parent, _ = v.GetByPath(NewPathFieldName("InnerBase"), NewPathFieldName("ListInt32"))
		l3a, err = parent.Len()
		require.Nil(t, err)
		require.Equal(t, l3+2, l3a)
	})
}