package test

import (
	"testing"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/conv/t2p"
	tbase "github.com/cloudwego/dynamicgo/testdata/kitex_gen/base"
	texample5 "github.com/cloudwego/dynamicgo/testdata/kitex_gen/example5"
	pbbase "github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/base"
	pbexample5 "github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example5"
	"github.com/cloudwego/dynamicgo/testdata/sample"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/cloudwego/dynamicgo/thrift/generic"
	"google.golang.org/protobuf/proto"
)

// Benchmark comparing manual struct-based Thrift -> PB conversion by field assignment,
// followed by protobuf marshaling to bytes.
func BenchmarkThrift2Proto_Struct(b *testing.B) {
	// get sample thrift object, same as DOM benchmark
	obj := sample.GetThriftInnerBase5(2, 1)

	// marshal thrift object to bytes
	tbytes := make([]byte, obj.BLength())
	if n := obj.FastWriteNocopy(tbytes, nil); n != obj.BLength() {
		b.Fatalf("marshal thrift object failed: expect %d, got %d", obj.BLength(), n)
	}

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
		CopyStruct(&obj, &pbMsg)

		// serialize to protobuf bytes
		_, err := proto.Marshal(&pbMsg)
		if err != nil {
			b.Fatalf("marshal pb message failed: %v", err)
		}
	}
}

// CopyStruct performs a deep field-by-field assignment from thrift InnerBase
// (kitex_gen/example5.InnerBase) into protobuf InnerBase (kitex_gen/pb/example5.InnerBase).
func CopyStruct(th *texample5.InnerBase, pb *pbexample5.InnerBase) {
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
			CopyStruct(th.ListInnerBase[i], tmp)
		}
	}
	// nested map of InnerBase
	if th.MapStringInnerBase != nil {
		pb.MapStringInnerBase = make(map[string]*pbexample5.InnerBase, len(th.MapStringInnerBase))
		for k, v := range th.MapStringInnerBase {
			tmp := &pbexample5.InnerBase{}
			pb.MapStringInnerBase[k] = tmp
			CopyStruct(v, tmp)
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
	obj := sample.GetThriftInnerBase5(2, 1)
	// marshal thrift object to bytes
	tbytes := make([]byte, obj.BLength())
	if n := obj.FastWriteNocopy(tbytes, nil); n != obj.BLength() {
		b.Fatalf("marshal thrift object failed: expect %d, got %d", obj.BLength(), n)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// unmashal thrift bytes to thrift DOM
		doc := generic.NewPathNode()
		defer generic.FreePathNode(doc)
		doc.Node = generic.NewNode(thrift.STRUCT, tbytes)
		if err := doc.Load(true, opts); err != nil {
			b.Fatalf("load thrift doc failed: %v", err)
		}

		// simulate DSL on thrift DOM
		var doc2 = generic.NewPathNode()
		defer generic.FreePathNode(doc2)
		if err := CopyDOM(doc, doc2); err != nil {
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

// CopyDOM deep copy a thrift DOM to another thrift DOM.
func CopyDOM(in *generic.PathNode, out *generic.PathNode) error {
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
