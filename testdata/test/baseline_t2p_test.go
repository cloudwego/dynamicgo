package test

import (
	"testing"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/conv/t2p"
	"github.com/cloudwego/dynamicgo/internal/util_test"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
	tbase "github.com/cloudwego/dynamicgo/testdata/kitex_gen/base"
	texample5 "github.com/cloudwego/dynamicgo/testdata/kitex_gen/example5"
	pbbase "github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/base"
	pbexample5 "github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example5"
	"github.com/cloudwego/dynamicgo/testdata/sample"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/cloudwego/dynamicgo/thrift/generic"
	goproto "google.golang.org/protobuf/proto"
)

var (
	sampleDepth = 3
	sampleWidth = 10
)

func getThriftBytes() []byte {
	obj := sample.GetThriftInnerBase5(sampleWidth, sampleDepth)
	tbytes := make([]byte, obj.BLength())
	if n := obj.FastWriteNocopy(tbytes, nil); n != obj.BLength() {
		panic("marshal thrift object failed")
	}
	// fmt.Println("thrift sample, width:", sampleWidth, "depth:", sampleDepth, "bytes size:", len(tbytes))
	return tbytes
}

// Benchmark comparing manual struct-based Thrift -> PB conversion by field assignment,
// followed by protobuf marshaling to bytes.
func BenchmarkThrift2Proto_Struct(b *testing.B) {
	tbytes := getThriftBytes()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// unmarshal thrift bytes to thrift struct
		var obj texample5.InnerBase
		if _, err := obj.FastRead(tbytes); err != nil {
			b.Fatalf("unmarshal thrift object failed: %v", err)
		}

		// simulate DSL on thrift struct
		var pbMsg pbexample5.InnerBase
		copyStruct(&obj, &pbMsg)

		// serialize to protobuf bytes
		pbbytes, err := goproto.Marshal(&pbMsg)
		if err != nil {
			b.Fatalf("marshal pb message failed: %v", err)
		}
		_ = pbbytes
	}
}

// copyStruct performs a deep field-by-field assignment from thrift InnerBase
// (kitex_gen/example5.InnerBase) into protobuf InnerBase (kitex_gen/pb/example5.InnerBase).
func copyStruct(th *texample5.InnerBase, pb *pbexample5.InnerBase) {
	if th == nil {
		return
	}

	// scalar fields
	pb.Bool = th.Bool
	pb.Int32 = th.Int32
	pb.Int64 = th.Int64
	pb.Double = th.Double
	pb.String_ = th.String_
	pb.Binary = th.Binary

	// list fields
	if th.ListString != nil {
		pb.ListString = make([]string, len(th.ListString))
		copy(pb.ListString, th.ListString)
	}
	if th.SetInt32_ != nil {
		pb.SetInt32 = make([]int32, len(th.SetInt32_))
		copy(pb.SetInt32, th.SetInt32_)
	}

	// map fields
	if th.MapStringString != nil {
		pb.MapStringString = make(map[string]string, len(th.MapStringString))
		for k, v := range th.MapStringString {
			pb.MapStringString[k] = v
		}
	}
	if th.MapInt32String != nil {
		pb.MapInt32String = make(map[int32]string, len(th.MapInt32String))
		for k, v := range th.MapInt32String {
			pb.MapInt32String[k] = v
		}
	}
	if th.MapInt64String != nil {
		pb.MapInt64String = make(map[int64]string, len(th.MapInt64String))
		for k, v := range th.MapInt64String {
			pb.MapInt64String[k] = v
		}
	}

	// nested list of InnerBase
	if th.ListInnerBase != nil {
		pb.ListInnerBase = make([]*pbexample5.InnerBase, len(th.ListInnerBase))
		for i := range th.ListInnerBase {
			tmp := &pbexample5.InnerBase{}
			pb.ListInnerBase[i] = tmp
			copyStruct(th.ListInnerBase[i], tmp)
		}
	}
	// nested map of InnerBase
	if th.MapStringInnerBase != nil {
		pb.MapStringInnerBase = make(map[string]*pbexample5.InnerBase, len(th.MapStringInnerBase))
		for k, v := range th.MapStringInnerBase {
			tmp := &pbexample5.InnerBase{}
			pb.MapStringInnerBase[k] = tmp
			copyStruct(v, tmp)
		}
	}

	// Base field (optional)
	if th.Base != nil {
		pb.Base = convertBase(th.Base)
	}

	return
}

func convertBase(th *tbase.Base) *pbbase.Base {
	if th == nil {
		return nil
	}
	pb := &pbbase.Base{}
	pb.LogID = th.LogID
	pb.Caller = th.Caller
	pb.Addr = th.Addr
	pb.Client = th.Client
	if th.TrafficEnv != nil {
		pb.TrafficEnv = &pbbase.TrafficEnv{Open: th.TrafficEnv.Open, Env: th.TrafficEnv.Env}
	}
	if th.Extra != nil {
		pb.Extra = make(map[string]string, len(th.Extra))
		for k, v := range th.Extra {
			pb.Extra[k] = v
		}
	}
	return pb
}

var opts = &generic.Options{}

// Benchmark comparing DOM-based Thrift -> PB conversion by field assignment,
// followed by protobuf marshaling to bytes.
func BenchmarkThrift2Proto_DOM(b *testing.B) {
	// get sample thrift object
	tbytes := getThriftBytes()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// unmashal thrift bytes to thrift DOM
		doc := generic.NewPathNode()
		// defer generic.FreePathNode(doc)
		doc.Node = generic.NewNode(thrift.STRUCT, tbytes)
		if err := doc.Load(true, opts); err != nil {
			b.Fatalf("load thrift doc failed: %v", err)
		}

		// simulate DSL on thrift DOM
		var doc2 = generic.NewPathNode()
		// defer generic.FreePathNode(doc2)
		if err := copyDOM(doc, doc2); err != nil {
			b.Fatalf("simulate DSL failed: %v", err)
		}

		// convert thrift DOM to proto bytes
		out := conv.NewBytes()
		defer conv.FreeBytes(out)
		cv := t2p.PathNodeToBytesConv{}
		if err := cv.Do(doc2, out); err != nil {
			b.Fatalf("convert thrift doc to proto bytes failed: %v", err)
		}
	}
}

// copyDOM deep copy a thrift DOM to another thrift DOM.
func copyDOM(in *generic.PathNode, out *generic.PathNode) error {
	// // query operations
	// f7 := in.Field(7, opts)
	// if f7 == nil {
	// 	return errors.New("field 7 not found")
	// }

	// // cast operation
	// v7, err := f7.Node.String()
	// if err != nil {
	// 	return err
	// }
	// if v7 != "hello" {
	// 	return errors.New("field 7 value not match")
	// }

	// // secondary query operation
	// m := in.Field(9, opts)
	// if m == nil {
	// 	return errors.New("field 9 not found")
	// }
	// if v := m.GetByStr("k", opts); v == nil {
	// 	return errors.New("field 9.k not found")
	// }

	// // assign operation
	// if _, err := m.SetByStr("k", f7.Node, opts); err != nil {
	// 	return err
	// }

	// // insert operation
	// if _, err := m.SetByStr("a", generic.NewNodeString("b"), opts); err != nil {
	// 	return err
	// }

	// assgin operations
	in.CopyTo(out)
	return nil
}

func BenchmarkThrift2Proto_Map(b *testing.B) {
	// get sample thrift object
	tbytes := getThriftBytes()

	tdesc := thrift.FnRequest(thrift.GetFnDescFromFile("testdata/idl/example5.thrift", "ExampleMethod", thrift.Options{}))
	pdesc := proto.FnRequest(proto.GetFnDescFromFile("testdata/idl/example5.proto", "ExampleMethod", proto.Options{}, util_test.MustGitPath("testdata/idl/")))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// unmarshal thrift bytes to thrift map
		t := thrift.BinaryProtocol{Buf: tbytes}
		tmap, err := t.ReadAnyWithDesc(tdesc, false, false, false, true)
		if err != nil {
			b.Fatalf("read thrift struct failed: %v", err)
		}

		// simulate DSL on thrift map
		var pmap interface{}
		copyAny(tmap, &pmap)

		// serialize to protobuf bytes
		p := binary.NewBinaryProtocolBuffer()
		if err := p.WriteAnyWithDesc(pdesc, pmap, false, false, false, true); err != nil {
			b.Fatalf("write proto struct failed: %v", err)
		}
		_ = p.Buf
	}
}

func copyAny(tmap any, pmap *any) {
	switch tm := tmap.(type) {
	case map[string]interface{}:
		*pmap = make(map[string]interface{}, len(tm))
		for k, v := range tm {
			// recurse copy map if value is map
			var tmp any
			copyAny(v, &tmp)
			(*pmap).(map[string]interface{})[k] = tmp
		}
	case map[interface{}]interface{}:
		*pmap = make(map[interface{}]interface{}, len(tm))
		for k, v := range tm {
			// recurse copy map if value is map
			var tmp any
			copyAny(v, &tmp)
			(*pmap).(map[interface{}]interface{})[k] = tmp
		}
	case map[int]any:
		*pmap = make(map[int]any, len(tm))
		for k, v := range tm {
			// recurse copy map if value is map
			var tmp any
			copyAny(v, &tmp)
			(*pmap).(map[int]any)[k] = tmp
		}
	case []interface{}:
		*pmap = make([]any, len(tm))
		for i, v := range tm {
			// recurse copy map if value is map
			var tmp any
			copyAny(v, &tmp)
			(*pmap).([]any)[i] = tmp
		}
	default:
		*pmap = tmap
	}
}
