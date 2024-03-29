package binary

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example"
	goproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type TestMode int8

const (
	PartialTest TestMode = iota
	ScalarsTest
	MessageTest
	NestedTest
	ListTest
	MapTest
	OneofTest
)

type testCase struct {
	file        string
	service     string
	isInput     bool
	fieldNumber int
	mode        TestMode
}

var testGroup = map[string]testCase{
	"Partial test": testCase{
		file:        "../../testdata/idl/example.proto",
		service:     "PartialMethodTest",
		isInput:     false,
		fieldNumber: 3,
		mode:        PartialTest,
	},
	"Scalars test": testCase{
		file:        "../../testdata/idl/example.proto",
		service:     "ScalarsMethodTest",
		isInput:     false,
		fieldNumber: 7,
		mode:        ScalarsTest,
	},
	"Message test": testCase{
		file:        "../../testdata/idl/example.proto",
		service:     "MessageMethodTest",
		isInput:     false,
		fieldNumber: 1,
		mode:        MessageTest,
	},
	"NestedMessage test": testCase{
		file:        "../../testdata/idl/example.proto",
		service:     "NestedMethodTest",
		isInput:     false,
		fieldNumber: 1,
		mode:        NestedTest,
	},
	"List test": testCase{
		file:        "../../testdata/idl/example.proto",
		service:     "ListMethodTest",
		isInput:     false,
		fieldNumber: 1,
		mode:        ListTest,
	},
	"Map test": testCase{
		file:        "../../testdata/idl/example.proto",
		service:     "MapMethodTest",
		isInput:     false,
		fieldNumber: 1,
		mode:        MapTest,
	},
	"Oneof test": testCase{
		file:        "../../testdata/idl/example.proto",
		service:     "OneofMethodTest",
		isInput:     false,
		fieldNumber: 1,
		mode:        OneofTest,
	},
}

func binaryDataBuild(mode TestMode) []byte {
	var req protoreflect.ProtoMessage
	switch mode {
	case PartialTest:
		data := &example.ExamplePartialResp{}
		// data.ShortEnglishMsg = "first"
		// data.ChineseMsg = "哈哈哈哈但是"
		data.LongEnglishMsg = "noboiubibipbpsdakonobnuondfap123141adfasdf"
		req = data
	case MessageTest:
		data := &example.ExampleMessageResp{}
		data.Base = &example.InnerBase{}
		data.Base.SBool = true
		data.Base.SInt32 = 12
		data.Base.SInt64 = 52
		data.Base.SUint32 = uint32(22)
		data.Base.SUint64 = uint64(62)
		data.Base.SFixed32 = uint32(120)
		data.Base.SFixed64 = uint64(240)
		data.Base.SSfixed32 = int32(120)
		data.Base.SSfixed64 = int64(240)
		data.Base.SFloat = 26.4
		data.Base.SDouble = 55.2
		data.Base.SBytes = []byte{1, 2, 7, 12, 64}
		data.Base.SString = "npdabigdbas dsaf@232#~32adgna;sbf;"
		req = data
	case NestedTest:
		data := &example.ExampleNestedResp{}
		data.TestNested = &example.Nested{}
		data.TestNested.SString = "aaaa"
		data.TestNested.Base = &example.InnerBase{}
		data.TestNested.Base.SBool = true
		data.TestNested.Base.SInt32 = 12
		data.TestNested.Base.SInt64 = 52
		data.TestNested.Base.SUint32 = uint32(22)
		data.TestNested.Base.SUint64 = uint64(62)
		data.TestNested.Base.SFixed32 = uint32(120)
		data.TestNested.Base.SFixed64 = uint64(240)
		data.TestNested.Base.SSfixed32 = int32(120)
		data.TestNested.Base.SSfixed64 = int64(240)
		data.TestNested.Base.SFloat = 26.4
		data.TestNested.Base.SDouble = 55.2
		data.TestNested.Base.SBytes = []byte{1, 2, 7, 12, 64}
		data.TestNested.Base.SString = "npdabigdbas dsaf@232#~32adgna;sbf;"
		req = data
	case ScalarsTest:
		data := &example.ExampleScalarsResp{}
		data.Scalars = &example.Scalars{}
		data.Scalars.SBool = true
		data.Scalars.SInt32 = 12
		data.Scalars.SInt64 = 52
		data.Scalars.SUint32 = uint32(22)
		data.Scalars.SUint64 = uint64(62)
		data.Scalars.SFixed32 = uint32(120)
		data.Scalars.SFixed64 = uint64(240)
		data.Scalars.SSfixed32 = int32(120)
		data.Scalars.SSfixed64 = int64(240)
		data.Scalars.SFloat = 26.4
		data.Scalars.SDouble = 55.2
		data.Scalars.SBytes = []byte{1, 2, 7, 12, 64}
		data.Scalars.SString = "npdabigdbas dsaf@232#~32adgna;sbf;"
		req = data
	case ListTest:
		data := &example.ExampleListResp{}
		data.TestList = &example.Repeats{}
		data.TestList.RptBool = []bool{true, false, true, true, false, true}
		data.TestList.RptInt32 = []int32{int32(12), int32(52), int32(123), int32(205)}
		data.TestList.RptInt64 = []int64{int64(21), int64(56), int64(210), int64(650)}
		data.TestList.RptUint32 = []uint32{uint32(22), uint32(78), uint32(110), uint32(430)}
		data.TestList.RptUint64 = []uint64{uint64(24), uint64(88), uint64(250), uint64(400)}
		data.TestList.RptFloat = []float32{float32(33), float32(50), float32(88), float32(130)}
		data.TestList.RptDouble = []float64{float64(33), float64(50), float64(88), float64(130)}
		data.TestList.RptString = []string{string("aaaa"), string("12sdfa"), string("2165cxvznpnbhpbnda"), string("bpibpbpi2b3541341")}
		data.TestList.RptBytes = [][]byte{[]byte{97, 98, 99, 100, 101, 102, 103}, []byte{104, 105, 106, 107, 108, 109, 110}, []byte{111, 112, 113, 114, 115, 116, 117}, []byte{118, 119, 120, 121, 122, 123, 124}}
		req = data
	case MapTest:
		data := &example.ExampleMapResp{}
		data.TestMap = &example.Maps{}
		data.TestMap.Int32ToStr = make(map[int32]string)
		data.TestMap.Int32ToStr[1] = "aaa"
		data.TestMap.Int32ToStr[2] = "bbb"
		data.TestMap.Int32ToStr[3] = "ccc"

		data.TestMap.BoolToUint32 = make(map[bool]uint32)
		data.TestMap.BoolToUint32[true] = uint32(10)
		data.TestMap.BoolToUint32[false] = uint32(12)
		data.TestMap.BoolToUint32[true] = uint32(14)

		data.TestMap.Uint64ToEnum = make(map[uint64]example.Enum)
		data.TestMap.Uint64ToEnum[uint64(1)] = example.Enum_ONE
		data.TestMap.Uint64ToEnum[uint64(2)] = example.Enum_TWO
		data.TestMap.Uint64ToEnum[uint64(3)] = example.Enum_TEN

		data.TestMap.StrToNested = make(map[string]*example.Nested)
		nestedObj1 := example.Nested{}
		nestedObj1.SString = "111"
		nestedObj1.Base = &example.InnerBase{}
		nestedObj1.Base.SBool = true
		nestedObj1.Base.SString = "nested111"
		nestedObj2 := example.Nested{}
		nestedObj2.SString = "222"
		nestedObj2.Base = &example.InnerBase{}
		nestedObj2.Base.SBool = false
		nestedObj2.Base.SString = "nested222"
		nestedObj3 := example.Nested{}
		nestedObj3.SString = "333"
		nestedObj3.Base = &example.InnerBase{}
		nestedObj3.Base.SBool = true
		nestedObj3.Base.SString = "nested333"
		data.TestMap.StrToNested["aaa"] = &nestedObj1
		data.TestMap.StrToNested["bbb"] = &nestedObj2
		data.TestMap.StrToNested["ccc"] = &nestedObj3

		data.TestMap.StrToOneofs = make(map[string]*example.Oneofs)
		oneofObj1 := example.Oneofs{}
		enumUnion := &example.Oneofs_OneofEnum{}
		enumUnion.OneofEnum = example.Enum_ONE
		oneofObj1.Union = enumUnion
		data.TestMap.StrToOneofs["aaa"] = &oneofObj1

		oneofObj2 := example.Oneofs{}
		stringUnion := &example.Oneofs_OneofString{}
		stringUnion.OneofString = "stringUnion"
		oneofObj2.Union = stringUnion
		data.TestMap.StrToOneofs["bbb"] = &oneofObj2

		oneofObj3 := example.Oneofs{}
		nestedUnion := &example.Oneofs_OneofNested{}
		nestedUnion.OneofNested = &example.Nested{}
		nestedUnion.OneofNested.SString = "nestedStringUnion"
		nestedUnion.OneofNested.Base = &example.InnerBase{}
		nestedUnion.OneofNested.Base.SBool = true
		nestedUnion.OneofNested.Base.SString = "bbb"
		oneofObj3.Union = nestedUnion
		data.TestMap.StrToOneofs["ccc"] = &oneofObj3
		req = data
	case OneofTest:
		data := &example.ExampleOneofResp{}
		data.TestOneof = &example.Oneofs{}
		// enum inner of union
		// enumUnion := &example.Oneofs_OneofEnum{}
		// enumUnion.OneofEnum = example.Enum_ONE
		// data.TestOneof.Union = enumUnion
		// string inner of union
		// stringUnion := &example.Oneofs_OneofString{}
		// stringUnion.OneofString = "aaaa"
		// data.TestOneof.Union = stringUnion
		// Nested inner of union
		nestedUnion := example.Oneofs_OneofNested{}
		nestedUnion.OneofNested = &example.Nested{}
		nestedUnion.OneofNested.SString = "aaa"
		nestedUnion.OneofNested.Base = &example.InnerBase{}
		nestedUnion.OneofNested.Base.SBool = true
		nestedUnion.OneofNested.Base.SString = "bbb"
		data.TestOneof.Union = &nestedUnion
		req = data
	}
	res, err := goproto.Marshal(req)
	if err != nil {
		panic("proto Marshal failed: PartialTest")
	}
	return res
}

func TestBinaryProtocol_ReadAnyWithDesc(t *testing.T) {
	for name, test := range testGroup {
		t.Run(name, func(t *testing.T) {
			p1, err := proto.NewDescritorFromPath(context.Background(), test.file)
			if err != nil {
				panic(err)
			}
			// get Target FieldDescriptor
			var fieldDescriptor *proto.FieldDescriptor
			if test.isInput {
				fieldDescriptor = p1.LookupMethodByName(test.service).Input().Message().ByNumber(proto.FieldNumber(test.fieldNumber))
			} else {
				fieldDescriptor = p1.LookupMethodByName(test.service).Output().Message().ByNumber(proto.FieldNumber(test.fieldNumber))
			}
			// protoc build pbData
			pbData := binaryDataBuild(test.mode)
			p := NewBinaryProtol(pbData)
			// WriteAnyWithDesc write the data into BinaryProtocol
			hasmessageLen := true
			if !fieldDescriptor.IsMap() && !fieldDescriptor.IsList() {
				if _, _, _, err := p.ConsumeTag(); err != nil {
					panic(err)
				}
			}
			v1, err := p.ReadAnyWithDesc(fieldDescriptor.Type(), hasmessageLen, true, false, true)
			fmt.Printf("%#v\n", v1)

			// wirte
			p = NewBinaryProtocolBuffer()
			if !fieldDescriptor.IsMap() && !fieldDescriptor.IsList() {
				p.AppendTagByKind(fieldDescriptor.Number(), fieldDescriptor.Kind())
			}
			needMessageLen := true
			err = p.WriteAnyWithDesc(fieldDescriptor.Type(), v1, needMessageLen, true, false, true)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%x\n", p.RawBuf())

			// Read again by ReadAnyWithDesc
			p = NewBinaryProtol(p.RawBuf())
			hasmessageLen = true
			if !fieldDescriptor.IsMap() && !fieldDescriptor.IsList() {
				if _, _, _, err := p.ConsumeTag(); err != nil {
					panic(err)
				}
			}
			v2, err := p.ReadAnyWithDesc(fieldDescriptor.Type(), hasmessageLen, true, false, true)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%#v\n", v2)
			if !reflect.DeepEqual(v1, v2) {
				panic("test ReadAnyWithDesc error")
			}
		})
	}
}

func TestTag(t *testing.T) {
	src := make([]byte, 0, 1024)
	p := NewBinaryProtol(src)
	// using a for loop to check each case of appendtag and consumeTag
	// the case is in the TestTag2

	testCase := []struct {
		number proto.FieldNumber
		wtyp   proto.WireType
		err    error
	}{
		{0, proto.Fixed32Type, errInvalidTag},
		{1, proto.Fixed32Type, nil},
		{proto.FirstReservedNumber, proto.BytesType, nil},
		{proto.LastReservedNumber, proto.StartGroupType, nil},
		{proto.MaxValidNumber, proto.VarintType, nil},
	}

	for _, c := range testCase {
		p.AppendTag(c.number, c.wtyp)
		num, _, _, err := p.ConsumeTag()
		if err != nil && err != errInvalidTag {
			t.Fatal(err)
		}
		if num != c.number {
			t.Fatal("test failed")
		}
	}

	_, _, _, err := p.ConsumeTag()
	if err != errInvalidTag {
		t.Fatal("test failed")
	}
}
