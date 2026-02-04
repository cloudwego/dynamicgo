package trim

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/cloudwego/thriftgo/generator/golang/extension/unknown"
	"github.com/stretchr/testify/require"
)

func BenchmarkFetchAndAssign(b *testing.B) {
	src := makeSampleFetch(3, 3)
	desc := makeDesc(3, 3, true)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m, err := fetchAny(desc, src)
		if err != nil {
			b.Fatalf("FetchAny failed: %v", err)
		}
		dest := &sampleAssign{}
		err = assignAny(desc, m, dest)
		if err != nil {
			b.Fatalf("AssignAny failed: %v", err)
		}
	}
}

func TestFetchAndAssign(t *testing.T) {
	src := makeSampleFetch(3, 3)
	srcjson, err := json.Marshal(src)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	desc := makeDesc(3, 3, true)

	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			m, err := fetchAny(desc, src)
			if err != nil {
				t.Fatalf("FetchAny failed: %v", err)
			}

			mjson, err := json.Marshal(m)
			if err != nil {
				t.Fatalf("json.Marshal failed: %v", err)
			}
			require.Equal(t, string(srcjson), string(mjson))

			dest := makeSampleAssign(3, 3)
			err = assignAny(desc, m, dest)
			if err != nil {
				t.Fatalf("AssignAny failed: %v", err)
			}

			destjson, err := json.Marshal(dest)
			if err != nil {
				t.Fatalf("json.Marshal failed: %v", err)
			}
			var srcAny interface{}
			if err := json.Unmarshal(srcjson, &srcAny); err != nil {
				t.Fail()
			}
			var destAny interface{}
			if err := json.Unmarshal(destjson, &destAny); err != nil {
				t.Fail()
			}
			require.Equal(t, srcAny, destAny)
		}()
	}

	wg.Wait()
}

func fetchAny(desc *Descriptor, any interface{}) (interface{}, error) {
	if desc != nil {
		desc.Normalize()
	}
	fetcher := &Fetcher{}
	return fetcher.FetchAny(desc, any)
}

// ===================== UnknownFields FetchAndAssign Tests =====================
// These tests verify the mapping between thrift _unknownFields and protobuf XXX_unrecognized,
// using the Descriptor as the field mapping source (NOT treating thrift ID as protobuf ID directly).

// thriftFetchStruct simulates a thrift struct with some fields in _unknownFields
type thriftFetchStruct struct {
	FieldA         int            `thrift:"FieldA,1" json:"field_a,omitempty"`
	_unknownFields unknown.Fields `json:"-"`
}

// protoAssignStruct simulates a protobuf struct that has different field layout
// Note: protobuf field IDs are different from thrift field IDs!
type protoAssignStruct struct {
	// field_a maps to protobuf field ID 10 (different from thrift ID 1)
	FieldA int `protobuf:"varint,10,req,name=field_a" json:"field_a,omitempty"`
	// field_b maps to protobuf field ID 20
	FieldB string `protobuf:"bytes,20,opt,name=field_b" json:"field_b,omitempty"`
	// field_c maps to protobuf field ID 30
	FieldC           int64  `protobuf:"varint,30,opt,name=field_c" json:"field_c,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

// protoAssignStructSmall has only field_a, other fields will go to XXX_unrecognized
type protoAssignStructSmall struct {
	FieldA           int    `protobuf:"varint,10,req,name=field_a" json:"field_a,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

// TestFetchAndAssign_UnknownToUnrecognized tests:
// fetch._unknownFields -> assign.XXX_unrecognized
// Thrift struct has field in _unknownFields, which should be encoded to protobuf XXX_unrecognized
// using the descriptor's ID mapping (descriptor ID -> protobuf field ID).
func TestFetchAndAssign_UnknownToUnrecognized(t *testing.T) {
	// Create thrift struct with field_a as known and field_b in _unknownFields
	src := &thriftFetchStruct{
		FieldA: 42,
	}

	// Encode field_b (thrift ID=2) and field_c (thrift ID=3) into _unknownFields
	p := thrift.BinaryProtocol{}
	p.WriteFieldBegin("", thrift.STRING, 2) // field_b, thrift ID=2
	p.WriteString("hello from unknown")
	p.WriteFieldBegin("", thrift.I64, 3) // field_c, thrift ID=3
	p.WriteI64(12345)
	src._unknownFields = p.Buf

	// Descriptor provides the mapping:
	// - field_a: Thrift ID=1, will be fetched as "field_a"
	// - field_b: Thrift ID=2, will be fetched as "field_b"
	// - field_c: Thrift ID=3, will be fetched as "field_c"
	// Note: these descriptor IDs are Thrift IDs, NOT protobuf IDs!
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "TestStruct",
		Children: []Field{
			{Name: "field_a", ID: 1}, // Thrift ID 1
			{Name: "field_b", ID: 2}, // Thrift ID 2 -> will become protobuf field 20
			{Name: "field_c", ID: 3}, // Thrift ID 3 -> will become protobuf field 30
		},
	}

	// Fetch from thrift struct
	fetched, err := fetchAny(desc, src)
	require.NoError(t, err)

	fetchedMap, ok := fetched.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, 42, fetchedMap["field_a"])
	require.Equal(t, "hello from unknown", fetchedMap["field_b"])
	require.Equal(t, int64(12345), fetchedMap["field_c"])

	// Assign to protobuf struct that only has field_a defined
	// field_b and field_c should go to XXX_unrecognized with protobuf IDs from descriptor
	dest := &protoAssignStructSmall{}

	// Create a new descriptor for assign that maps names to protobuf field IDs
	// The key point: descriptor.ID here represents the TARGET protobuf field ID
	assignDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "TestStruct",
		Children: []Field{
			{Name: "field_a", ID: 10}, // Protobuf ID 10
			{Name: "field_b", ID: 20}, // Protobuf ID 20 -> will go to XXX_unrecognized
			{Name: "field_c", ID: 30}, // Protobuf ID 30 -> will go to XXX_unrecognized
		},
	}

	err = assignAny(assignDesc, fetched, dest)
	require.NoError(t, err)

	// Verify field_a is assigned correctly
	require.Equal(t, 42, dest.FieldA)

	// Verify field_b and field_c are in XXX_unrecognized with correct protobuf IDs
	require.NotEmpty(t, dest.XXX_unrecognized)

	// Decode XXX_unrecognized to verify field IDs
	bp := binary.NewBinaryProtol(dest.XXX_unrecognized)
	defer binary.FreeBinaryProtocol(bp)

	foundFieldB := false
	foundFieldC := false

	for bp.Read < len(bp.Buf) {
		fieldNum, wireType, _, err := bp.ConsumeTag()
		require.NoError(t, err)

		switch fieldNum {
		case 20: // field_b with protobuf ID 20
			require.Equal(t, proto.BytesType, wireType) // length-delimited for string
			val, err := bp.ReadString(true)
			require.NoError(t, err)
			require.Equal(t, "hello from unknown", val)
			foundFieldB = true
		case 30: // field_c with protobuf ID 30
			require.Equal(t, proto.VarintType, wireType) // varint for int64
			val, err := bp.ReadInt64()
			require.NoError(t, err)
			require.Equal(t, int64(12345), val)
			foundFieldC = true
		default:
			t.Errorf("unexpected field number in XXX_unrecognized: %d", fieldNum)
		}
	}

	require.True(t, foundFieldB, "field_b not found in XXX_unrecognized")
	require.True(t, foundFieldC, "field_c not found in XXX_unrecognized")
}

// TestFetchAndAssign_UnknownToKnown tests:
// fetch._unknownFields -> assign.已知字段
// Thrift struct has field in _unknownFields, which should be assigned to protobuf known field
func TestFetchAndAssign_UnknownToKnown(t *testing.T) {
	// Create thrift struct with only field_a as known
	src := &thriftFetchStruct{
		FieldA: 42,
	}

	// Encode field_b (thrift ID=2) into _unknownFields
	p := thrift.BinaryProtocol{}
	p.WriteFieldBegin("", thrift.STRING, 2)
	p.WriteString("from thrift unknown")
	src._unknownFields = p.Buf

	// Descriptor for fetching (Thrift IDs)
	fetchDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "TestStruct",
		Children: []Field{
			{Name: "field_a", ID: 1}, // Thrift ID 1
			{Name: "field_b", ID: 2}, // Thrift ID 2, in _unknownFields
		},
	}

	// Fetch from thrift struct
	fetched, err := fetchAny(fetchDesc, src)
	require.NoError(t, err)

	fetchedMap, ok := fetched.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, 42, fetchedMap["field_a"])
	require.Equal(t, "from thrift unknown", fetchedMap["field_b"])

	// Assign to protobuf struct that has field_b as a known field
	dest := &protoAssignStruct{}

	// Descriptor for assigning (Protobuf IDs)
	assignDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "TestStruct",
		Children: []Field{
			{Name: "field_a", ID: 10}, // Protobuf ID 10
			{Name: "field_b", ID: 20}, // Protobuf ID 20, is a known field in dest
		},
	}

	err = assignAny(assignDesc, fetched, dest)
	require.NoError(t, err)

	// Verify both fields are assigned correctly to known fields
	require.Equal(t, 42, dest.FieldA)
	require.Equal(t, "from thrift unknown", dest.FieldB)

	// XXX_unrecognized should be empty since all fields are known
	require.Empty(t, dest.XXX_unrecognized)
}

// TestFetchAndAssign_KnownToUnrecognized tests:
// fetch.已知字段 -> assign.XXX_unrecognized
// Thrift struct has field as known, which should be encoded to protobuf XXX_unrecognized
// because protobuf struct doesn't have that field defined
func TestFetchAndAssign_KnownToUnrecognized(t *testing.T) {
	// thriftFetchStructFull has more known fields than protoAssignStructSmall
	type thriftFetchStructFull struct {
		FieldA int    `thrift:"FieldA,1" json:"field_a,omitempty"`
		FieldB string `thrift:"FieldB,2" json:"field_b,omitempty"`
		FieldC int64  `thrift:"FieldC,3" json:"field_c,omitempty"`
	}

	// Create thrift struct with all fields as known
	src := &thriftFetchStructFull{
		FieldA: 42,
		FieldB: "known field in thrift",
		FieldC: 98765,
	}

	// Descriptor for fetching (Thrift IDs)
	fetchDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "TestStruct",
		Children: []Field{
			{Name: "field_a", ID: 1}, // Thrift ID 1
			{Name: "field_b", ID: 2}, // Thrift ID 2
			{Name: "field_c", ID: 3}, // Thrift ID 3
		},
	}

	// Fetch from thrift struct
	fetched, err := fetchAny(fetchDesc, src)
	require.NoError(t, err)

	fetchedMap, ok := fetched.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, 42, fetchedMap["field_a"])
	require.Equal(t, "known field in thrift", fetchedMap["field_b"])
	require.Equal(t, int64(98765), fetchedMap["field_c"])

	// Assign to protobuf struct that only has field_a
	// field_b and field_c should go to XXX_unrecognized
	dest := &protoAssignStructSmall{}

	// Descriptor for assigning (Protobuf IDs)
	// Note: IDs here are protobuf field IDs, different from thrift IDs!
	assignDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "TestStruct",
		Children: []Field{
			{Name: "field_a", ID: 10}, // Protobuf ID 10, is a known field in dest
			{Name: "field_b", ID: 20}, // Protobuf ID 20, NOT in dest -> XXX_unrecognized
			{Name: "field_c", ID: 30}, // Protobuf ID 30, NOT in dest -> XXX_unrecognized
		},
	}

	err = assignAny(assignDesc, fetched, dest)
	require.NoError(t, err)

	// Verify field_a is assigned correctly
	require.Equal(t, 42, dest.FieldA)

	// Verify field_b and field_c are in XXX_unrecognized with protobuf IDs
	require.NotEmpty(t, dest.XXX_unrecognized)

	// Decode XXX_unrecognized to verify
	bp := binary.NewBinaryProtol(dest.XXX_unrecognized)
	defer binary.FreeBinaryProtocol(bp)

	foundFieldB := false
	foundFieldC := false

	for bp.Read < len(bp.Buf) {
		fieldNum, wireType, _, err := bp.ConsumeTag()
		require.NoError(t, err)

		switch fieldNum {
		case 20: // field_b with protobuf ID 20
			require.Equal(t, proto.BytesType, wireType) // length-delimited for string
			val, err := bp.ReadString(true)
			require.NoError(t, err)
			require.Equal(t, "known field in thrift", val)
			foundFieldB = true
		case 30: // field_c with protobuf ID 30
			require.Equal(t, proto.VarintType, wireType) // varint for int64
			val, err := bp.ReadInt64()
			require.NoError(t, err)
			require.Equal(t, int64(98765), val)
			foundFieldC = true
		default:
			t.Errorf("unexpected field number in XXX_unrecognized: %d", fieldNum)
		}
	}

	require.True(t, foundFieldB, "field_b not found in XXX_unrecognized")
	require.True(t, foundFieldC, "field_c not found in XXX_unrecognized")
}

// TestFetchAndAssign_MixedScenario tests a complex scenario with all three cases:
// 1. fetch._unknownFields -> assign.XXX_unrecognized
// 2. fetch._unknownFields -> assign.known field
// 3. fetch.known field -> assign.XXX_unrecognized
func TestFetchAndAssign_MixedScenario(t *testing.T) {
	// thriftMixedStruct has some known fields, some in _unknownFields
	type thriftMixedStruct struct {
		FieldA         int            `thrift:"FieldA,1" json:"field_a,omitempty"`
		FieldD         string         `thrift:"FieldD,4" json:"field_d,omitempty"` // known in thrift, will go to XXX_unrecognized
		_unknownFields unknown.Fields `json:"-"`
	}

	// protoMixedStruct has different field layout
	type protoMixedStruct struct {
		FieldA           int    `protobuf:"varint,10,req,name=field_a" json:"field_a,omitempty"`
		FieldB           string `protobuf:"bytes,20,opt,name=field_b" json:"field_b,omitempty"` // will receive from thrift _unknownFields
		XXX_unrecognized []byte `json:"-"`
	}

	// Create thrift struct
	src := &thriftMixedStruct{
		FieldA: 100,
		FieldD: "thrift known -> pb unknown",
	}

	// Encode field_b (thrift ID=2) and field_c (thrift ID=3) into _unknownFields
	p := thrift.BinaryProtocol{}
	p.WriteFieldBegin("", thrift.STRING, 2) // field_b -> will go to pb known field
	p.WriteString("thrift unknown -> pb known")
	p.WriteFieldBegin("", thrift.I32, 3) // field_c -> will go to pb XXX_unrecognized
	p.WriteI32(999)
	src._unknownFields = p.Buf

	// Fetch descriptor (Thrift IDs)
	fetchDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "MixedStruct",
		Children: []Field{
			{Name: "field_a", ID: 1}, // known in thrift
			{Name: "field_b", ID: 2}, // in thrift _unknownFields
			{Name: "field_c", ID: 3}, // in thrift _unknownFields
			{Name: "field_d", ID: 4}, // known in thrift
		},
	}

	// Fetch from thrift struct
	fetched, err := fetchAny(fetchDesc, src)
	require.NoError(t, err)

	fetchedMap, ok := fetched.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, 100, fetchedMap["field_a"])
	require.Equal(t, "thrift unknown -> pb known", fetchedMap["field_b"])
	require.Equal(t, int32(999), fetchedMap["field_c"])
	require.Equal(t, "thrift known -> pb unknown", fetchedMap["field_d"])

	// Assign descriptor (Protobuf IDs)
	assignDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "MixedStruct",
		Children: []Field{
			{Name: "field_a", ID: 10}, // known in pb
			{Name: "field_b", ID: 20}, // known in pb (from thrift _unknownFields)
			{Name: "field_c", ID: 30}, // NOT in pb -> XXX_unrecognized
			{Name: "field_d", ID: 40}, // NOT in pb -> XXX_unrecognized
		},
	}

	dest := &protoMixedStruct{}
	err = assignAny(assignDesc, fetched, dest)
	require.NoError(t, err)

	// Verify known fields
	require.Equal(t, 100, dest.FieldA)
	require.Equal(t, "thrift unknown -> pb known", dest.FieldB) // from thrift _unknownFields to pb known

	// Verify XXX_unrecognized contains field_c and field_d
	require.NotEmpty(t, dest.XXX_unrecognized)

	bp := binary.NewBinaryProtol(dest.XXX_unrecognized)
	defer binary.FreeBinaryProtocol(bp)

	foundFieldC := false
	foundFieldD := false

	for bp.Read < len(bp.Buf) {
		fieldNum, wireType, _, err := bp.ConsumeTag()
		require.NoError(t, err)

		switch fieldNum {
		case 30: // field_c with protobuf ID 30
			require.Equal(t, proto.VarintType, wireType) // varint for int32
			val, err := bp.ReadInt64()                   // stored as varint (int64)
			require.NoError(t, err)
			require.Equal(t, int64(999), val)
			foundFieldC = true
		case 40: // field_d with protobuf ID 40
			require.Equal(t, proto.BytesType, wireType) // length-delimited for string
			val, err := bp.ReadString(true)
			require.NoError(t, err)
			require.Equal(t, "thrift known -> pb unknown", val)
			foundFieldD = true
		default:
			t.Errorf("unexpected field number in XXX_unrecognized: %d", fieldNum)
		}
	}

	require.True(t, foundFieldC, "field_c (from thrift _unknownFields) not found in pb XXX_unrecognized")
	require.True(t, foundFieldD, "field_d (from thrift known) not found in pb XXX_unrecognized")
}

// TestFetchAndAssign_NestedStructWithUnknownFields tests nested struct handling
func TestFetchAndAssign_NestedStructWithUnknownFields(t *testing.T) {
	// Thrift struct with nested struct in _unknownFields
	type thriftOuter struct {
		ID             int            `thrift:"ID,1" json:"id,omitempty"`
		_unknownFields unknown.Fields `json:"-"`
	}

	// Proto struct with nested struct as known field
	type protoInner struct {
		Name  string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
		Value int    `protobuf:"varint,2,opt,name=value" json:"value,omitempty"`
	}
	type protoOuter struct {
		ID               int         `protobuf:"varint,10,req,name=id" json:"id,omitempty"`
		Inner            *protoInner `protobuf:"bytes,20,opt,name=inner" json:"inner,omitempty"`
		XXX_unrecognized []byte      `json:"-"`
	}

	// Create thrift struct with nested struct in _unknownFields
	src := &thriftOuter{
		ID: 1,
	}

	// Encode nested struct (thrift ID=2) into _unknownFields
	p := thrift.BinaryProtocol{}
	p.WriteFieldBegin("", thrift.STRUCT, 2) // inner struct, thrift ID=2
	p.WriteFieldBegin("", thrift.STRING, 1) // inner.name
	p.WriteString("nested name")
	p.WriteFieldBegin("", thrift.I32, 2) // inner.value
	p.WriteI32(123)
	p.WriteFieldStop() // end inner struct
	src._unknownFields = p.Buf

	// Fetch descriptor
	fetchDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "Outer",
		Children: []Field{
			{Name: "id", ID: 1},
			{
				Name: "inner",
				ID:   2,
				Desc: &Descriptor{
					Kind: TypeKind_Struct,
					Type: "Inner",
					Children: []Field{
						{Name: "name", ID: 1},
						{Name: "value", ID: 2},
					},
				},
			},
		},
	}

	// Fetch from thrift struct
	fetched, err := fetchAny(fetchDesc, src)
	require.NoError(t, err)

	fetchedMap, ok := fetched.(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, 1, fetchedMap["id"])

	innerMap, ok := fetchedMap["inner"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "nested name", innerMap["name"])
	require.Equal(t, int32(123), innerMap["value"])

	// Assign descriptor (Protobuf IDs)
	assignDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "Outer",
		Children: []Field{
			{Name: "id", ID: 10},
			{
				Name: "inner",
				ID:   20,
				Desc: &Descriptor{
					Kind: TypeKind_Struct,
					Type: "Inner",
					Children: []Field{
						{Name: "name", ID: 1},
						{Name: "value", ID: 2},
					},
				},
			},
		},
	}

	dest := &protoOuter{}
	err = assignAny(assignDesc, fetched, dest)
	require.NoError(t, err)

	// Verify
	require.Equal(t, 1, dest.ID)
	require.NotNil(t, dest.Inner)
	require.Equal(t, "nested name", dest.Inner.Name)
	require.Equal(t, 123, dest.Inner.Value)
}

// ===================== Shallow Descriptor Tests =====================
// These tests verify FetchAndAssign correctness when Descriptor is shallower than the actual objects
// or when Descriptor is missing some fields.

// TestFetchAndAssign_ShallowDescriptor tests when Descriptor depth < object depth
// Only the fields specified in the Descriptor should be fetched/assigned
func TestFetchAndAssign_ShallowDescriptor(t *testing.T) {
	// Use makeSampleFetch to create a deep nested structure (depth=3)
	// But use a shallow descriptor (depth=1) that only fetches top-level fields
	src := makeSampleFetch(2, 3) // width=2, depth=3

	// Create a shallow descriptor (depth=1) - no nested Desc for Children
	shallowDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleFetch",
		Children: []Field{
			{Name: "field_a", ID: 1}, // scalar field, no Desc needed
			{Name: "field_e", ID: 5}, // scalar field, no Desc needed
			// field_b, field_c, field_d are complex types but no Desc -> fetch as-is
			{Name: "field_b", ID: 2}, // list field, no Desc -> fetch raw value
			{Name: "field_d", ID: 4}, // pointer to struct, no Desc -> fetch raw value
		},
	}

	// Fetch with shallow descriptor
	fetched, err := fetchAny(shallowDesc, src)
	require.NoError(t, err)

	fetchedMap, ok := fetched.(map[string]interface{})
	require.True(t, ok)

	// Verify scalar fields
	require.Equal(t, 1, fetchedMap["field_a"])
	require.Equal(t, "1", fetchedMap["field_e"])

	// field_b should be the raw slice (not recursively processed)
	fieldB, ok := fetchedMap["field_b"]
	require.True(t, ok)
	require.NotNil(t, fieldB)

	// field_d should be the raw pointer value (not recursively processed)
	fieldD, ok := fetchedMap["field_d"]
	require.True(t, ok)
	require.NotNil(t, fieldD)

	// Now assign with the same shallow descriptor
	dest := makeSampleAssign(2, 3)
	err = assignAny(shallowDesc, fetched, dest)
	require.NoError(t, err)

	// Verify scalar fields are correctly assigned
	require.Equal(t, 1, dest.FieldA)
	require.Equal(t, "1", dest.FieldE)

	// Verify complex fields are assigned (even though descriptor is shallow)
	require.NotNil(t, dest.FieldB)
	require.NotNil(t, dest.FieldD)
}

// TestFetchAndAssign_MissingFieldsInDescriptor tests when Descriptor is missing some fields
// Only the fields specified in the Descriptor should be fetched/assigned
func TestFetchAndAssign_MissingFieldsInDescriptor(t *testing.T) {
	// Create a thrift struct with many fields
	type thriftFullStruct struct {
		FieldA int    `thrift:"FieldA,1" json:"field_a,omitempty"`
		FieldB string `thrift:"FieldB,2" json:"field_b,omitempty"`
		FieldC int64  `thrift:"FieldC,3" json:"field_c,omitempty"`
		FieldD string `thrift:"FieldD,4" json:"field_d,omitempty"`
		FieldE int32  `thrift:"FieldE,5" json:"field_e,omitempty"`
	}

	// Create a protobuf struct with same fields
	type protoFullStruct struct {
		FieldA           int    `protobuf:"varint,1,req,name=field_a" json:"field_a,omitempty"`
		FieldB           string `protobuf:"bytes,2,opt,name=field_b" json:"field_b,omitempty"`
		FieldC           int64  `protobuf:"varint,3,opt,name=field_c" json:"field_c,omitempty"`
		FieldD           string `protobuf:"bytes,4,opt,name=field_d" json:"field_d,omitempty"`
		FieldE           int32  `protobuf:"varint,5,opt,name=field_e" json:"field_e,omitempty"`
		XXX_unrecognized []byte `json:"-"`
	}

	src := &thriftFullStruct{
		FieldA: 100,
		FieldB: "hello",
		FieldC: 12345,
		FieldD: "world",
		FieldE: 999,
	}

	// Descriptor only includes field_a, field_c, field_e (missing field_b, field_d)
	partialDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "PartialStruct",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{Name: "field_c", ID: 3},
			{Name: "field_e", ID: 5},
		},
	}

	// Fetch with partial descriptor
	fetched, err := fetchAny(partialDesc, src)
	require.NoError(t, err)

	fetchedMap, ok := fetched.(map[string]interface{})
	require.True(t, ok)

	// Only specified fields should be fetched
	require.Equal(t, 100, fetchedMap["field_a"])
	require.Equal(t, int64(12345), fetchedMap["field_c"])
	require.Equal(t, int32(999), fetchedMap["field_e"])

	// Missing fields should NOT be in the fetched map
	_, hasFieldB := fetchedMap["field_b"]
	_, hasFieldD := fetchedMap["field_d"]
	require.False(t, hasFieldB, "field_b should not be fetched")
	require.False(t, hasFieldD, "field_d should not be fetched")

	// Assign with partial descriptor
	dest := &protoFullStruct{
		FieldB: "original_b", // pre-set values
		FieldD: "original_d",
	}

	err = assignAny(partialDesc, fetched, dest)
	require.NoError(t, err)

	// Verify only specified fields are assigned
	require.Equal(t, 100, dest.FieldA)
	require.Equal(t, int64(12345), dest.FieldC)
	require.Equal(t, int32(999), dest.FieldE)

	// Pre-existing values for missing fields should be preserved
	require.Equal(t, "original_b", dest.FieldB)
	require.Equal(t, "original_d", dest.FieldD)
}

// TestFetchAndAssign_NestedMissingFields tests nested structs with missing fields in Descriptor
func TestFetchAndAssign_NestedMissingFields(t *testing.T) {
	// Thrift nested struct
	type thriftInner struct {
		Name  string `thrift:"Name,1" json:"name,omitempty"`
		Value int    `thrift:"Value,2" json:"value,omitempty"`
		Extra string `thrift:"Extra,3" json:"extra,omitempty"`
	}
	type thriftOuter struct {
		ID    int          `thrift:"ID,1" json:"id,omitempty"`
		Inner *thriftInner `thrift:"Inner,2" json:"inner,omitempty"`
		Tag   string       `thrift:"Tag,3" json:"tag,omitempty"`
	}

	// Proto nested struct
	type protoInner struct {
		Name  string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
		Value int    `protobuf:"varint,2,opt,name=value" json:"value,omitempty"`
		Extra string `protobuf:"bytes,3,opt,name=extra" json:"extra,omitempty"`
	}
	type protoOuter struct {
		ID               int         `protobuf:"varint,1,req,name=id" json:"id,omitempty"`
		Inner            *protoInner `protobuf:"bytes,2,opt,name=inner" json:"inner,omitempty"`
		Tag              string      `protobuf:"bytes,3,opt,name=tag" json:"tag,omitempty"`
		XXX_unrecognized []byte      `json:"-"`
	}

	src := &thriftOuter{
		ID: 1,
		Inner: &thriftInner{
			Name:  "inner_name",
			Value: 42,
			Extra: "extra_data",
		},
		Tag: "outer_tag",
	}

	// Descriptor: outer has id and inner, but inner only has name (missing value and extra)
	// Also missing outer.tag
	partialDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "Outer",
		Children: []Field{
			{Name: "id", ID: 1},
			{
				Name: "inner",
				ID:   2,
				Desc: &Descriptor{
					Kind: TypeKind_Struct,
					Type: "Inner",
					Children: []Field{
						{Name: "name", ID: 1}, // only fetch name, not value or extra
					},
				},
			},
			// tag (ID: 3) is missing from descriptor
		},
	}

	// Fetch with partial descriptor
	fetched, err := fetchAny(partialDesc, src)
	require.NoError(t, err)

	fetchedMap, ok := fetched.(map[string]interface{})
	require.True(t, ok)

	// Verify outer fields
	require.Equal(t, 1, fetchedMap["id"])
	_, hasTag := fetchedMap["tag"]
	require.False(t, hasTag, "tag should not be fetched")

	// Verify inner struct - only name should be present
	innerMap, ok := fetchedMap["inner"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "inner_name", innerMap["name"])

	_, hasValue := innerMap["value"]
	_, hasExtra := innerMap["extra"]
	require.False(t, hasValue, "inner.value should not be fetched")
	require.False(t, hasExtra, "inner.extra should not be fetched")

	// Assign with partial descriptor
	dest := &protoOuter{
		Tag: "original_tag",
		Inner: &protoInner{
			Value: 999,
			Extra: "original_extra",
		},
	}

	err = assignAny(partialDesc, fetched, dest)
	require.NoError(t, err)

	// Verify assigned fields
	require.Equal(t, 1, dest.ID)
	require.Equal(t, "original_tag", dest.Tag) // preserved

	require.NotNil(t, dest.Inner)
	require.Equal(t, "inner_name", dest.Inner.Name)
	// These should be preserved from original dest.Inner
	require.Equal(t, 999, dest.Inner.Value)
	require.Equal(t, "original_extra", dest.Inner.Extra)
}

// TestFetchAndAssign_DescShallowerThanNestedList tests when Descriptor is shallow for list of structs
func TestFetchAndAssign_DescShallowerThanNestedList(t *testing.T) {
	// Thrift struct with list of nested structs
	type thriftItem struct {
		Name  string `thrift:"Name,1" json:"name,omitempty"`
		Value int    `thrift:"Value,2" json:"value,omitempty"`
	}
	type thriftContainer struct {
		Items []*thriftItem `thrift:"Items,1" json:"items,omitempty"`
	}

	// Proto struct
	type protoItem struct {
		Name  string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
		Value int    `protobuf:"varint,2,opt,name=value" json:"value,omitempty"`
	}
	type protoContainer struct {
		Items            []*protoItem `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
		XXX_unrecognized []byte       `json:"-"`
	}

	src := &thriftContainer{
		Items: []*thriftItem{
			{Name: "item1", Value: 100},
			{Name: "item2", Value: 200},
			{Name: "item3", Value: 300},
		},
	}

	// Shallow descriptor - no Desc for items, so items are fetched as-is
	shallowDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "Container",
		Children: []Field{
			{Name: "items", ID: 1}, // No Desc, so list is fetched as raw value
		},
	}

	// Fetch with shallow descriptor
	fetched, err := fetchAny(shallowDesc, src)
	require.NoError(t, err)

	fetchedMap, ok := fetched.(map[string]interface{})
	require.True(t, ok)

	// Items should be fetched as raw slice
	items, ok := fetchedMap["items"]
	require.True(t, ok)
	require.NotNil(t, items)

	// Assign with shallow descriptor
	dest := &protoContainer{}
	err = assignAny(shallowDesc, fetched, dest)
	require.NoError(t, err)

	// Verify items are assigned
	require.Len(t, dest.Items, 3)
	require.Equal(t, "item1", dest.Items[0].Name)
	require.Equal(t, 100, dest.Items[0].Value)
	require.Equal(t, "item2", dest.Items[1].Name)
	require.Equal(t, 200, dest.Items[1].Value)
	require.Equal(t, "item3", dest.Items[2].Name)
	require.Equal(t, 300, dest.Items[2].Value)
}

// TestFetchAndAssign_DescShallowerThanNestedMap tests when Descriptor is shallow for map of structs
func TestFetchAndAssign_DescShallowerThanNestedMap(t *testing.T) {
	// Thrift struct with map of nested structs
	type thriftItem struct {
		Name  string `thrift:"Name,1" json:"name,omitempty"`
		Value int    `thrift:"Value,2" json:"value,omitempty"`
	}
	type thriftContainer struct {
		Data map[string]*thriftItem `thrift:"Data,1" json:"data,omitempty"`
	}

	// Proto struct
	type protoItem struct {
		Name  string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
		Value int    `protobuf:"varint,2,opt,name=value" json:"value,omitempty"`
	}
	type protoContainer struct {
		Data             map[string]*protoItem `protobuf:"bytes,1,rep,name=data" json:"data,omitempty"`
		XXX_unrecognized []byte                `json:"-"`
	}

	src := &thriftContainer{
		Data: map[string]*thriftItem{
			"key1": {Name: "item1", Value: 100},
			"key2": {Name: "item2", Value: 200},
		},
	}

	// Shallow descriptor - no Desc for map values
	shallowDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "Container",
		Children: []Field{
			{
				Name: "data",
				ID:   1,
				Desc: &Descriptor{
					Kind: TypeKind_StrMap,
					Type: "DataMap",
					Children: []Field{
						{Name: "*"}, // Wildcard, but no Desc for values -> fetch as raw
					},
				},
			},
		},
	}

	// Fetch with shallow descriptor
	fetched, err := fetchAny(shallowDesc, src)
	require.NoError(t, err)

	fetchedMap, ok := fetched.(map[string]interface{})
	require.True(t, ok)

	// Data should be fetched as map
	data, ok := fetchedMap["data"].(map[string]interface{})
	require.True(t, ok)
	require.Len(t, data, 2)

	// Assign with shallow descriptor
	dest := &protoContainer{}
	err = assignAny(shallowDesc, fetched, dest)
	require.NoError(t, err)

	// Verify data is assigned
	require.Len(t, dest.Data, 2)
	require.Equal(t, "item1", dest.Data["key1"].Name)
	require.Equal(t, 100, dest.Data["key1"].Value)
	require.Equal(t, "item2", dest.Data["key2"].Name)
	require.Equal(t, 200, dest.Data["key2"].Value)
}

// TestFetchAndAssign_PartialMapKeys tests when Descriptor only specifies some map keys
func TestFetchAndAssign_PartialMapKeys(t *testing.T) {
	// Thrift struct with map
	type thriftContainer struct {
		Data map[string]int `thrift:"Data,1" json:"data,omitempty"`
	}

	// Proto struct
	type protoContainer struct {
		Data             map[string]int `protobuf:"bytes,1,rep,name=data" json:"data,omitempty"`
		XXX_unrecognized []byte         `json:"-"`
	}

	src := &thriftContainer{
		Data: map[string]int{
			"key1": 100,
			"key2": 200,
			"key3": 300,
			"key4": 400,
		},
	}

	// Descriptor only specifies key1 and key3 (not key2, key4)
	partialDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "Container",
		Children: []Field{
			{
				Name: "data",
				ID:   1,
				Desc: &Descriptor{
					Kind: TypeKind_StrMap,
					Type: "DataMap",
					Children: []Field{
						{Name: "key1"},
						{Name: "key3"},
						// key2 and key4 are not specified
					},
				},
			},
		},
	}

	// Fetch with partial descriptor
	fetched, err := fetchAny(partialDesc, src)
	require.NoError(t, err)

	fetchedMap, ok := fetched.(map[string]interface{})
	require.True(t, ok)

	// Only specified keys should be fetched
	data, ok := fetchedMap["data"].(map[string]interface{})
	require.True(t, ok)
	require.Len(t, data, 2)
	require.Equal(t, 100, data["key1"])
	require.Equal(t, 300, data["key3"])

	_, hasKey2 := data["key2"]
	_, hasKey4 := data["key4"]
	require.False(t, hasKey2, "key2 should not be fetched")
	require.False(t, hasKey4, "key4 should not be fetched")

	// Assign with partial descriptor
	dest := &protoContainer{
		Data: map[string]int{
			"key2": 999, // pre-existing key
		},
	}
	err = assignAny(partialDesc, fetched, dest)
	require.NoError(t, err)

	// Verify only specified keys are assigned, pre-existing keys might be overwritten depending on implementation
	// Based on current implementation, the map is replaced
	require.Equal(t, 100, dest.Data["key1"])
	require.Equal(t, 300, dest.Data["key3"])
}

// TestFetchAndAssign_EmptyDescriptor tests when Descriptor has no children
func TestFetchAndAssign_EmptyDescriptor(t *testing.T) {
	type thriftStruct struct {
		FieldA int    `thrift:"FieldA,1" json:"field_a,omitempty"`
		FieldB string `thrift:"FieldB,2" json:"field_b,omitempty"`
	}

	type protoStruct struct {
		FieldA           int    `protobuf:"varint,1,req,name=field_a" json:"field_a,omitempty"`
		FieldB           string `protobuf:"bytes,2,opt,name=field_b" json:"field_b,omitempty"`
		XXX_unrecognized []byte `json:"-"`
	}

	src := &thriftStruct{
		FieldA: 100,
		FieldB: "hello",
	}

	// Empty descriptor - no children
	emptyDesc := &Descriptor{
		Kind:     TypeKind_Struct,
		Type:     "EmptyStruct",
		Children: []Field{},
	}

	// Fetch with empty descriptor
	fetched, err := fetchAny(emptyDesc, src)
	require.NoError(t, err)

	fetchedMap, ok := fetched.(map[string]interface{})
	require.True(t, ok)

	// No fields should be fetched
	require.Len(t, fetchedMap, 0)

	// Assign with empty descriptor
	dest := &protoStruct{
		FieldA: 999,
		FieldB: "original",
	}
	err = assignAny(emptyDesc, fetched, dest)
	require.NoError(t, err)

	// Original values should be preserved
	require.Equal(t, 999, dest.FieldA)
	require.Equal(t, "original", dest.FieldB)
}

// ===================== Descriptor.String() Tests =====================

// TestDescriptor_String_Scalar tests String() for scalar type descriptor
func TestDescriptor_String_Scalar(t *testing.T) {
	desc := &Descriptor{
		Kind: TypeKind_Leaf,
		Type: "ScalarType",
	}

	result := desc.String()
	require.Equal(t, "-", result)
}

// TestDescriptor_String_EmptyStruct tests String() for struct with no children
func TestDescriptor_String_EmptyStruct(t *testing.T) {
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "EmptyStruct",
	}

	result := desc.String()
	require.Equal(t, "<EmptyStruct>{}", result)
}

// TestDescriptor_String_SimpleStruct tests String() for struct with scalar fields
func TestDescriptor_String_SimpleStruct(t *testing.T) {
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SimpleStruct",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{Name: "field_b", ID: 2},
		},
	}

	result := desc.String()
	expected := `<SimpleStruct>{
	"field_a": -,
	"field_b": -
}`
	require.Equal(t, expected, result)
}

// TestDescriptor_String_NestedStruct tests String() for nested struct
func TestDescriptor_String_NestedStruct(t *testing.T) {
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "OuterStruct",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{
				Name: "inner",
				ID:   2,
				Desc: &Descriptor{
					Kind: TypeKind_Struct,
					Type: "InnerStruct",
					Children: []Field{
						{Name: "name", ID: 1},
						{Name: "value", ID: 2},
					},
				},
			},
		},
	}

	result := desc.String()
	expected := `<OuterStruct>{
	"field_a": -,
	"inner": <InnerStruct>{
		"name": -,
		"value": -
	}
}`
	require.Equal(t, expected, result)
}

// TestDescriptor_String_MapWithWildcard tests String() for map with "*" key
func TestDescriptor_String_MapWithWildcard(t *testing.T) {
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "ContainerStruct",
		Children: []Field{
			{Name: "id", ID: 1},
			{
				Name: "data",
				ID:   2,
				Desc: &Descriptor{
					Kind: TypeKind_StrMap,
					Type: "DataMap",
					Children: []Field{
						{
							Name: "*",
							Desc: &Descriptor{
								Kind: TypeKind_Struct,
								Type: "ItemStruct",
								Children: []Field{
									{Name: "name", ID: 1},
								},
							},
						},
					},
				},
			},
		},
	}

	result := desc.String()
	expected := `<ContainerStruct>{
	"id": -,
	"data": <MAP>{
		"*": <ItemStruct>{
			"name": -
		}
	}
}`
	require.Equal(t, expected, result)
}

// TestDescriptor_String_MapWithSpecificKeys tests String() for map with specific keys
func TestDescriptor_String_MapWithSpecificKeys(t *testing.T) {
	desc := &Descriptor{
		Kind: TypeKind_StrMap,
		Type: "SpecificKeyMap",
		Children: []Field{
			{Name: "key1"},
			{Name: "key2"},
			{
				Name: "key3",
				Desc: &Descriptor{
					Kind: TypeKind_Struct,
					Type: "NestedType",
					Children: []Field{
						{Name: "value", ID: 1},
					},
				},
			},
		},
	}

	result := desc.String()
	expected := `<MAP>{
	"key1": -,
	"key2": -,
	"key3": <NestedType>{
		"value": -
	}
}`
	require.Equal(t, expected, result)
}

// TestDescriptor_String_DeeplyNested tests String() for deeply nested structure
func TestDescriptor_String_DeeplyNested(t *testing.T) {
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "Level1",
		Children: []Field{
			{
				Name: "level2",
				ID:   1,
				Desc: &Descriptor{
					Kind: TypeKind_Struct,
					Type: "Level2",
					Children: []Field{
						{
							Name: "level3",
							ID:   1,
							Desc: &Descriptor{
								Kind: TypeKind_Struct,
								Type: "Level3",
								Children: []Field{
									{Name: "value", ID: 1},
								},
							},
						},
					},
				},
			},
		},
	}

	result := desc.String()
	expected := `<Level1>{
	"level2": <Level2>{
		"level3": <Level3>{
			"value": -
		}
	}
}`
	require.Equal(t, expected, result)
}

// TestDescriptor_String_CircularReference tests String() handles circular references
func TestDescriptor_String_CircularReference(t *testing.T) {
	// Create a descriptor that references itself
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SelfRef",
		Children: []Field{
			{Name: "value", ID: 1},
		},
	}
	// Add self-reference
	desc.Children = append(desc.Children, Field{
		Name: "self",
		ID:   2,
		Desc: desc, // circular reference
	})

	result := desc.String()
	expected := `<SelfRef>{
	"value": -,
	"self": <SelfRef>
}`
	require.Equal(t, expected, result)
}

// TestDescriptor_String_MixedTypes tests String() for mixed struct and map types
func TestDescriptor_String_MixedTypes(t *testing.T) {
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "MixedStruct",
		Children: []Field{
			{Name: "scalar_field", ID: 1},
			{
				Name: "struct_field",
				ID:   2,
				Desc: &Descriptor{
					Kind: TypeKind_Struct,
					Type: "NestedStruct",
					Children: []Field{
						{Name: "a", ID: 1},
					},
				},
			},
			{
				Name: "map_field",
				ID:   3,
				Desc: &Descriptor{
					Kind: TypeKind_StrMap,
					Type: "NestedMap",
					Children: []Field{
						{Name: "*"},
					},
				},
			},
			{
				Name: "scalar_desc",
				ID:   4,
				Desc: &Descriptor{
					Kind: TypeKind_Leaf,
					Type: "ScalarType",
				},
			},
		},
	}

	result := desc.String()
	expected := `<MixedStruct>{
	"scalar_field": -,
	"struct_field": <NestedStruct>{
		"a": -
	},
	"map_field": <MAP>{
		"*": -
	},
	"scalar_desc": -
}`
	require.Equal(t, expected, result)
}
