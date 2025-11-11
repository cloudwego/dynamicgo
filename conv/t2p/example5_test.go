package t2p

import (
	"sync"
	"testing"

	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/base"
	thriftExample5 "github.com/cloudwego/dynamicgo/testdata/kitex_gen/example5"
	pbase "github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/base"
	pbExample5 "github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example5"
	"github.com/cloudwego/dynamicgo/thrift"
	tgeneric "github.com/cloudwego/dynamicgo/thrift/generic"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestT2P_Example5_InnerBase validates that t2p produces valid pb bytes for example5.InnerBase
func TestT2P_Example5_RawNode(t *testing.T) {
	// 1) 构造 thrift 结构体并序列化为 thrift 二进制
	buf := generate(t)

	// 2) 构造 thrift raw node，并输出 pb 字节
	root := &tgeneric.PathNode{Node: tgeneric.NewNode(thrift.STRUCT, buf)}
	// root.Load(true, &tgeneric.Options{})
	var pbBytes []byte
	conv := &PathNodeToBytesConv{}
	require.NoError(t, conv.Do(root, &pbBytes))
	require.NotEmpty(t, pbBytes)

	// 3) 使用 pb 结构体对 pb 字节反序列化并校验
	validate(t, pbBytes)
}

func TestT2P_Example5_DOM(t *testing.T) {
	// 1) 构造 thrift 结构体并序列化为 thrift 二进制
	buf := generate(t)

	// 2) 构造 thrift DOM，并输出 pb 字节
	root := &tgeneric.PathNode{Node: tgeneric.NewNode(thrift.STRUCT, buf)}
	require.NoError(t, root.Load(true, &tgeneric.Options{}))

	var pbBytes []byte
	conv := &PathNodeToBytesConv{}
	require.NoError(t, conv.Do(root, &pbBytes))
	require.NotEmpty(t, pbBytes)

	// 3) 使用 pb 结构体对 pb 字节反序列化并校验
	validate(t, pbBytes)
}

func generate(t testing.TB) []byte {
	// 1) 构造 thrift 结构体并序列化为 thrift 二进制
	th := thriftExample5.NewInnerBase()
	th.Bool = true
	th.Int32 = 123
	th.Int64 = -456
	th.Double = 3.14159
	th.String_ = "hello"
	th.ListString = []string{"a", "b"}
	th.MapStringString = map[string]string{"k": "v"}
	th.SetInt32_ = []int32{1, 2, 3}
	th.MapInt32String = map[int32]string{1: "x"}
	th.Binary = []byte{0x00, 0xFF, 0x10}
	th.MapInt64String = map[int64]string{7: "vv"}
	inner := thriftExample5.NewInnerBase()
	inner.Bool = false
	th.ListInnerBase = []*thriftExample5.InnerBase{inner}
	th.MapStringInnerBase = map[string]*thriftExample5.InnerBase{"kk": inner}
	th.Base = &base.Base{TrafficEnv: &base.TrafficEnv{
		Env: "prod",
	}}

	buf := make([]byte, th.BLength())
	ret := th.FastWriteNocopy(buf, nil)
	require.Greater(t, ret, int(0), "FastWriteNocopy failed")
	return buf[:ret]
}

func validate(t *testing.T, pbBytes []byte) {
	act := &pbExample5.InnerBase{}
	// Try unmarshal
	if err := proto.Unmarshal(pbBytes, act); err != nil {
		t.Fatalf("proto unmarshal error: %v; bytes=%x", err, pbBytes)
	}

	// 构造期望 pb 结构体，字段与 thrift 值一致
	exp := &pbExample5.InnerBase{
		Bool:            true,
		Int32:           123,
		Int64:           -456,
		Double:          3.14159,
		String_:         "hello",
		ListString:      []string{"a", "b"},
		MapStringString: map[string]string{"k": "v"},
		SetInt32:        []int32{1, 2, 3},
		MapInt32String:  map[int32]string{1: "x"},
		Binary:          []byte{0x00, 0xFF, 0x10},
		MapInt64String:  map[int64]string{7: "vv"},
		Base: &pbase.Base{
			TrafficEnv: &pbase.TrafficEnv{
				Env: "prod",
			},
		},
	}
	// 嵌套字段
	exp.ListInnerBase = []*pbExample5.InnerBase{&pbExample5.InnerBase{Bool: false}}
	exp.MapStringInnerBase = map[string]*pbExample5.InnerBase{"kk": &pbExample5.InnerBase{Bool: false}}

	require.Equal(t, exp.Bool, act.Bool)
	require.Equal(t, exp.Int32, act.Int32)
	require.Equal(t, exp.Int64, act.Int64)
	require.InDelta(t, exp.Double, act.Double, 1e-9)
	require.Equal(t, exp.String_, act.String_)
	require.Equal(t, exp.ListString, act.ListString)
	require.Equal(t, exp.MapStringString, act.MapStringString)
	require.Equal(t, exp.SetInt32, act.SetInt32)
	require.Equal(t, exp.MapInt32String, act.MapInt32String)
	require.Equal(t, exp.Binary, act.Binary)
	require.Equal(t, exp.MapInt64String, act.MapInt64String)
	// 嵌套校验
	require.Equal(t, 1, len(act.ListInnerBase))
	require.Equal(t, false, act.ListInnerBase[0].Bool)
	require.Equal(t, 1, len(act.MapStringInnerBase))
	require.Equal(t, false, act.MapStringInnerBase["kk"].Bool)
	require.Equal(t, exp.Base.TrafficEnv.Env, act.Base.TrafficEnv.Env)
}

func BenchmarkConv_DOM(b *testing.B) {
	b.ReportAllocs()
	buf := generate(b)
	if len(buf) == 0 {
		b.Fatalf("generateThriftBytes failed")
	}
	conv := &PathNodeToBytesConv{}
	root := &tgeneric.PathNode{Node: tgeneric.NewNode(thrift.STRUCT, buf)}
	if err := root.Load(true, &tgeneric.Options{}); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var out = bytesPool.Get().(*[]byte)
		if err := conv.Do(root, out); err != nil {
			b.Fatal(err)
		}
		putBytes(out)
	}
}

var bytesPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 0, 1024)
		return &b
	},
}

func putBytes(b *[]byte) {
	*b = (*b)[:0]
	bytesPool.Put(b)
}

func BenchmarkConv_RawNode(b *testing.B) {
	b.ReportAllocs()
	buf := generate(b)
	if len(buf) == 0 {
		b.Fatalf("generateThriftBytes failed")
	}
	conv := &PathNodeToBytesConv{}
	root := &tgeneric.PathNode{Node: tgeneric.NewNode(thrift.STRUCT, buf)}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var out = bytesPool.Get().(*[]byte)
		if err := conv.Do(root, out); err != nil {
			b.Fatal(err)
		}
		putBytes(out)
	}
}

func BenchmarkProtoMarshal(b *testing.B) {
	b.ReportAllocs()

	exp := &pbExample5.InnerBase{
		Bool:            true,
		Int32:           123,
		Int64:           -456,
		Double:          3.14159,
		String_:         "hello",
		ListString:      []string{"a", "b"},
		MapStringString: map[string]string{"k": "v"},
		SetInt32:        []int32{1, 2, 3},
		MapInt32String:  map[int32]string{1: "x"},
		Binary:          []byte{0x00, 0xFF, 0x10},
		MapInt64String:  map[int64]string{7: "vv"},
		Base: &pbase.Base{
			TrafficEnv: &pbase.TrafficEnv{
				Env: "prod",
			},
		},
	}
	exp.ListInnerBase = []*pbExample5.InnerBase{&pbExample5.InnerBase{Bool: false}}
	exp.MapStringInnerBase = map[string]*pbExample5.InnerBase{"kk": &pbExample5.InnerBase{Bool: false}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pbBytes, err := proto.Marshal(exp)
		if err != nil {
			b.Fatal(err)
		}
		_ = pbBytes
	}
}
