package testdata

import (
	"bytes"
	"context"
	ejson "encoding/json"
	"math"
	"strconv"
	"sync"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/conv/j2p"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/baseline"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/protowire"
)

const (
	protoPath = "testdata/idl/baseline.proto"
)

var (
	simplePbJSON  = ""
	nestingPbJSON = ""
)

func init() {
	// build simpleJSON data
	sobj := getPbSimpleValue()
	sout, err := ejson.Marshal(sobj)
	if err != nil {
		panic(err)
	}
	simplePbJSON = string(sout)
	println("small data size: ", len(simplePbJSON))
	var out bytes.Buffer
	ejson.Indent(&out, sout, "", "")
	println(out.String())

	// build nestingJSON data
	nobj := getPbNestingValue()
	nout, err := ejson.Marshal(nobj)
	if err != nil {
		panic(err)
	}
	nestingPbJSON = string(nout)
	println("medium data size: ", len(nestingPbJSON))
	out.Reset()

	psobj := getPartialSimpleValue()
	psout, err := ejson.Marshal(psobj)
	if err != nil {
		panic(err)
	}
	println("partial small data size: ", len(psout))

	pnobj := getPartialNestingValue()
	pnout, err := ejson.Marshal(pnobj)
	if err != nil {
		panic(err)
	}
	println("partial medium data size: ", len(pnout))
}

func getPbSimpleValue() *baseline.Simple {
	return &baseline.Simple{
		ByteField:   []byte{math.MaxInt8},
		I64Field:    math.MaxInt64,
		DoubleField: math.MaxFloat64,
		I32Field:    math.MaxInt32,
		StringField: getString(),
		BinaryField: getBytes(),
		ListString:  []string{"aaaa", "bbbb", "cccc"},
	}
}

func getPbPartialSimpleValue() *baseline.PartialSimple {
	return &baseline.PartialSimple{
		ByteField:   []byte{math.MaxInt8},
		DoubleField: math.MaxFloat64,
		BinaryField: getBytes(),
	}
}

func getPbNestingValue() *baseline.Nesting {
	var ret = &baseline.Nesting{
		String_:         getString(),
		ListSimple:      []*baseline.Simple{},
		Double:          math.MaxFloat64,
		I32:             math.MaxInt32,
		ListI32:         []int32{},
		I64:             math.MaxInt64,
		MapStringString: map[string]string{},
		SimpleStruct:    getPbSimpleValue(),
		MapI32I64:       map[int32]int64{},
		ListString:      []string{},
		// Binary:          getBytes(),
		MapI64String:    map[int64]string{},
		ListI64:         []int64{},
		Byte:            []byte{math.MaxInt8},
		MapStringSimple: map[string]*baseline.Simple{},
	}
	for i := 0; i < listCount; i++ {
		ret.ListSimple = append(ret.ListSimple, getPbSimpleValue())
		ret.ListI32 = append(ret.ListI32, math.MinInt32)
		ret.ListI64 = append(ret.ListI64, math.MinInt64)
		ret.ListString = append(ret.ListString, getString())
	}

	for i := 0; i < mapCount; i++ {
		ret.MapStringString[strconv.Itoa(i)] = getString()
		ret.MapI32I64[int32(i)] = math.MinInt64
		ret.MapI64String[int64(i)] = getString()
		ret.MapStringSimple[strconv.Itoa(i)] = getPbSimpleValue()
	}

	return ret
}

func getPbPartialNestingValue() *baseline.PartialNesting {
	var ret = &baseline.PartialNesting{
		ListSimple:      []*baseline.PartialSimple{},
		SimpleStruct:    getPbPartialSimpleValue(),
		MapStringSimple: map[string]*baseline.PartialSimple{},
	}
	for i := 0; i < listCount; i++ {
		ret.ListSimple = append(ret.ListSimple, getPbPartialSimpleValue())
	}

	for i := 0; i < mapCount; i++ {
		ret.MapStringSimple[strconv.Itoa(i)] = getPbPartialSimpleValue()
	}

	return ret
}

func TestJSON2Protobuf_Simple(t *testing.T) {
	_, err := ejson.Marshal(baseline.Simple{})
	if err != nil {
		t.Fatal(err)
	}
	simple := getPbSimpleDesc()
	// unmarshal json to get pb obj
	stru2 := baseline.Simple{}
	if err := sonic.UnmarshalString(simplePbJSON, &stru2); err != nil {
		t.Fatal(err)
	}
	// convert json to pb bytes
	// nj := convertI642StringSimple(simplePbJSON)    // may have error here
	cv := j2p.NewBinaryConv(conv.Options{
		WriteDefaultField: true,
		EnableHttpMapping: false,
	})
	ctx := context.Background()
	out, err := cv.Do(ctx, simple, []byte(simplePbJSON))
	require.Nil(t, err)

	// kitex read pb bytes to pb obj
	stru := baseline.Simple{}
	l := 0
	dataLen := len(out)
	for l < dataLen {
		id, wtyp, tagLen := protowire.ConsumeTag(out)
		if tagLen < 0 {
			t.Fatal("proto data error format")
		}
		l += tagLen
		out = out[tagLen:]
		offset, err := stru.FastRead(out, int8(wtyp), int32(id))
		require.Nil(t, err)
		out = out[offset:]
		l += offset
	}
	if len(out) != 0 {
		t.Fatal("proto data error format")
	}
	require.Equal(t, stru2, stru)
}

func TestJSON2Protobuf_Simple_Parallel(t *testing.T) {
	_, err := ejson.Marshal(baseline.Simple{})
	if err != nil {
		t.Fatal(err)
	}
	simple := getPbSimpleDesc()
	// unmarshal json to get pb obj
	stru2 := baseline.Simple{}
	if err := sonic.UnmarshalString(simplePbJSON, &stru2); err != nil {
		t.Fatal(err)
	}
	// convert json to pb bytes
	// nj := convertI642StringSimple(simplePbJSON)    // may have error here
	cv := j2p.NewBinaryConv(conv.Options{
		WriteDefaultField: true,
		EnableHttpMapping: false,
	})
	wg := sync.WaitGroup{}
	for i := 0; i < Concurrency; i++ {
		wg.Add(1)
		go func(i int) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic: %d\n%s", i, simplePbJSON)
				}
			}()
			defer wg.Done()
			ctx := context.Background()
			out, err := cv.Do(ctx, simple, []byte(simplePbJSON))
			require.Nil(t, err)

			stru := baseline.Simple{}
			l := 0
			dataLen := len(out)
			for l < dataLen {
				id, wtyp, tagLen := protowire.ConsumeTag(out)
				if tagLen < 0 {
					t.Fatal("proto data error format")
				}
				l += tagLen
				out = out[tagLen:]
				offset, err := stru.FastRead(out, int8(wtyp), int32(id))
				require.Nil(t, err)
				out = out[offset:]
				l += offset
			}
			if len(out) != 0 {
				t.Fatal("proto data error format")
			}
			require.Equal(t, stru2, stru)
		}(i)
	}
	wg.Wait()
}

func TestJSON2Protobuf_Nesting(t *testing.T) {
	_, err := ejson.Marshal(baseline.Nesting{})
	if err != nil {
		t.Fatal(err)
	}
	nesting := getPbNestingDesc()
	// unmarshal json to get pb obj
	stru2 := baseline.Nesting{}
	if err := sonic.UnmarshalString(nestingPbJSON, &stru2); err != nil {
		t.Fatal(err)
	}
	// convert json to pb bytes
	// nj := convertI642StringSimple(simplePbJSON)    // may have error here
	cv := j2p.NewBinaryConv(conv.Options{
		WriteDefaultField: true,
		EnableHttpMapping: false,
	})
	ctx := context.Background()
	out, err := cv.Do(ctx, nesting, []byte(nestingPbJSON))
	require.Nil(t, err)

	// kitex read pb bytes to pb obj
	stru := baseline.Nesting{}
	l := 0
	dataLen := len(out)
	for l < dataLen {
		id, wtyp, tagLen := protowire.ConsumeTag(out)
		if tagLen < 0 {
			t.Fatal("proto data error format")
		}
		l += tagLen
		out = out[tagLen:]
		offset, err := stru.FastRead(out, int8(wtyp), int32(id))
		require.Nil(t, err)
		out = out[offset:]
		l += offset
	}
	if len(out) != 0 {
		t.Fatal("proto data error format")
	}
	require.Equal(t, stru2, stru)
}

func TestJSON2Protobuf_Nesting_Parallel(t *testing.T) {
	_, err := ejson.Marshal(baseline.Nesting{})
	if err != nil {
		t.Fatal(err)
	}
	nesting := getPbNestingDesc()
	// unmarshal json to get pb obj
	stru2 := baseline.Nesting{}
	if err := sonic.UnmarshalString(nestingPbJSON, &stru2); err != nil {
		t.Fatal(err)
	}
	// convert json to pb bytes
	// nj := convertI642StringSimple(simplePbJSON)    // may have error here
	cv := j2p.NewBinaryConv(conv.Options{
		WriteDefaultField: true,
		EnableHttpMapping: false,
	})

	wg := sync.WaitGroup{}
	for i := 0; i < Concurrency; i++ {
		wg.Add(1)
		go func(i int) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic: %d\n%s", i, nestingPbJSON)
				}
			}()
			defer wg.Done()
			ctx := context.Background()
			out, err := cv.Do(ctx, nesting, []byte(nestingPbJSON))
			require.Nil(t, err)

			stru := baseline.Nesting{}
			l := 0
			dataLen := len(out)
			for l < dataLen {
				id, wtyp, tagLen := protowire.ConsumeTag(out)
				if tagLen < 0 {
					t.Fatal("proto data error format")
				}
				l += tagLen
				out = out[tagLen:]
				offset, err := stru.FastRead(out, int8(wtyp), int32(id))
				require.Nil(t, err)
				out = out[offset:]
				l += offset
			}
			if len(out) != 0 {
				t.Fatal("proto data error format")
			}
			require.Equal(t, stru2, stru)
		}(i)
	}
	wg.Wait()
}

func BenchmarkJSON2Protobuf_DynamicGo_Raw(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		simple := getPbSimpleDesc()
		cv := j2p.NewBinaryConv(conv.Options{
			WriteDefaultField:  false,
			EnableValueMapping: true,
		})
		// nj := []byte(convertI642StringSimple(simpleJSON))
		ctx := context.Background()
		out, err := cv.Do(ctx, simple, []byte(simplePbJSON))
		require.Nil(b, err)

		b.SetBytes(int64(len(out)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cv.Do(ctx, simple, []byte(simplePbJSON))
		}
	})

	b.Run("medium", func(b *testing.B) {
		nesting := getPbNestingDesc()
		cv := j2p.NewBinaryConv(conv.Options{
			WriteDefaultField:  false,
			EnableValueMapping: true,
		})
		// println(string(nestingPbJSON))
		// nj := []byte(convertI642StringSimple(simpleJSON))
		ctx := context.Background()
		out, err := cv.Do(ctx, nesting, []byte(nestingPbJSON))
		require.Nil(b, err)

		b.SetBytes(int64(len(out)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cv.Do(ctx, nesting, []byte(nestingPbJSON))
		}
	})
}

func BenchmarkJSON2Protobuf_SonicAndKitex(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		v := baseline.Simple{}
		if err := sonic.UnmarshalString(simplePbJSON, &v); err != nil {
			b.Fatal(err)
		}
		var buf = make([]byte, v.Size())
		v.FastWrite(buf)

		b.SetBytes(int64(len(buf)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v := baseline.Simple{}
			_ = v.Size()
			_ = sonic.UnmarshalString(simplePbJSON, &v)
			v.FastWrite(buf)
		}
	})

	b.Run("medium", func(b *testing.B) {
		v := baseline.Nesting{}
		if err := sonic.UnmarshalString(nestingPbJSON, &v); err != nil {
			b.Fatal(err)
		}
		var buf = make([]byte, v.Size())
		v.FastWrite(buf)

		b.SetBytes(int64(len(buf)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v := baseline.Nesting{}
			_ = v.Size()
			_ = sonic.UnmarshalString(nestingPbJSON, &v)
			v.FastWrite(buf)
		}
	})
}

func BenchmarkJSON2Protobuf_ProtoBufGo(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		v := baseline.Simple{}
		if err := protojson.Unmarshal([]byte(simplePbJSON), v.ProtoReflect().Interface()); err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = protojson.Unmarshal([]byte(simplePbJSON), v.ProtoReflect().Interface())
		}
	})

	b.Run("medium", func(b *testing.B) {
		v := baseline.Nesting{}
		if err := protojson.Unmarshal([]byte(nestingPbJSON), v.ProtoReflect().Interface()); err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = protojson.Unmarshal([]byte(nestingPbJSON), v.ProtoReflect().Interface())
		}
	})
}
