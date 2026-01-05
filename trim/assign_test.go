/**
 * Copyright 2025 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package trim

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/dynamicgo/proto/binary"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

func assignAny(desc *Descriptor, src interface{}, dest interface{}) error {
	if desc != nil {
		desc.Normalize()
	}
	assigner := &Assigner{}
	return assigner.AssignAny(desc, src, dest)
}

type sampleAssign struct {
	XXX_NoUnkeyedLiteral map[string]interface{}   `json:"-"`
	FieldA               int                      `protobuf:"varint,1,req,name=field_a" json:"field_a,omitempty"`
	FieldB               []*sampleAssign          `protobuf:"bytes,2,opt,name=field_b" json:"field_b,omitempty"`
	FieldC               map[string]*sampleAssign `protobuf:"bytes,3,opt,name=field_c" json:"field_c,omitempty"`
	FieldD               *sampleAssign            `protobuf:"bytes,4,opt,name=field_d" json:"field_d,omitempty"`
	FieldE               string                   `protobuf:"bytes,5,opt,name=field_e" json:"field_e,omitempty"`
	FieldList            []int                    `protobuf:"bytes,6,opt,name=field_list" json:"field_list,omitempty"`
	FieldMap             map[string]int           `protobuf:"bytes,7,opt,name=field_map" json:"field_map,omitempty"`
	XXX_unrecognized     []byte                   `json:"-"`
}

func makeSampleAssign(width, depth int) *sampleAssign {
	if width <= 0 || depth <= 0 {
		return nil
	}
	ret := &sampleAssign{
		FieldA:    2,
		FieldE:    "2",
		FieldC:    make(map[string]*sampleAssign),
		FieldList: []int{4, 5, 6},
		FieldMap: map[string]int{
			"4": 4,
			"5": 5,
			"6": 6,
		},
	}
	for i := 0; i < width; i++ {
		ret.FieldB = append(ret.FieldB, makeSampleAssign(width, depth-1))
		ret.FieldC[fmt.Sprintf("%d", i)] = makeSampleAssign(width, depth-1)
	}
	ret.FieldD = makeSampleAssign(width, depth-1)
	return ret
}

// sampleAssignSmall is a struct with fewer fields than SampleAssign
// Used to test XXX_unrecognized field encoding
type sampleAssignSmall struct {
	XXX_NoUnkeyedLiteral map[string]interface{} `json:"-"`
	FieldA               *int                   `protobuf:"varint,1,req,name=field_a"`
	FieldE               string                 `protobuf:"bytes,5,opt,name=field_e"`
	XXX_unrecognized     []byte                 `json:"-"`
}

func intPtr(i int) *int {
	return &i
}

func TestAssignAny_Basic(t *testing.T) {
	src := map[string]interface{}{
		"field_a": 42,
		"field_e": "hello",
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssign",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{Name: "field_e", ID: 5},
		},
	}

	dest := &sampleAssign{}
	err := assignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	if dest.FieldA != 42 {
		t.Errorf("field_a: expected 42, got %v", dest.FieldA)
	}
	if dest.FieldE != "hello" {
		t.Errorf("field_e: expected 'hello', got %v", dest.FieldE)
	}
}

func TestAssignAny_NestedStruct(t *testing.T) {
	src := map[string]interface{}{
		"field_a": 1,
		"field_d": map[string]interface{}{
			"field_a": 2,
			"field_e": "nested",
		},
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssign",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{
				Name: "field_d",
				ID:   4,
				Desc: &Descriptor{
					Kind: TypeKind_Struct,
					Type: "SampleAssign",
					Children: []Field{
						{Name: "field_a", ID: 1},
						{Name: "field_e", ID: 5},
					},
				},
			},
		},
	}

	dest := &sampleAssign{}
	err := assignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	if dest.FieldA != 1 {
		t.Errorf("field_a: expected 1, got %v", dest.FieldA)
	}
	if dest.FieldD == nil {
		t.Fatalf("field_d: expected non-nil")
	}
	if dest.FieldD.FieldA != 2 {
		t.Errorf("field_d.field_a: expected 2, got %v", dest.FieldD.FieldA)
	}
	if dest.FieldD.FieldE != "nested" {
		t.Errorf("field_d.field_e: expected 'nested', got %v", dest.FieldD.FieldE)
	}
}

func TestAssignAny_List(t *testing.T) {
	src := map[string]interface{}{
		"field_list": []int{1, 2, 3},
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssign",
		Children: []Field{
			{
				Name: "field_list",
				ID:   6,
				Desc: &Descriptor{
					Kind: TypeKind_Leaf,
					Type: "LIST",
				},
			},
		},
	}

	dest := &sampleAssign{}
	err := assignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(dest.FieldList, expected) {
		t.Errorf("field_list: expected %v, got %v", expected, dest.FieldList)
	}
}

func TestAssignAny_Map(t *testing.T) {
	src := map[string]interface{}{
		"field_map": map[string]interface{}{
			"key1": 100,
			"key2": 200,
		},
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssign",
		Children: []Field{
			{
				Name: "field_map",
				ID:   7,
				Desc: &Descriptor{
					Kind: TypeKind_StrMap,
					Type: "MAP",
					Children: []Field{
						{Name: "*"},
					},
				},
			},
		},
	}

	dest := &sampleAssign{}
	err := assignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	expected := map[string]int{"key1": 100, "key2": 200}
	if !reflect.DeepEqual(dest.FieldMap, expected) {
		t.Errorf("field_map: expected %v, got %v", expected, dest.FieldMap)
	}
}

func TestAssignAny_OnlyAssignLeafNodes(t *testing.T) {
	src := map[string]interface{}{
		"field_b": []interface{}{
			map[string]interface{}{},
			map[string]interface{}{
				"field_a": 10,
				"field_b": []interface{}{},
			},
			map[string]interface{}{
				"field_a": 10,
			},
		},
		"field_c": map[string]interface{}{
			"empty": map[string]interface{}{},
			"non_empty": map[string]interface{}{
				"field_a": 10,
				"field_c": map[string]interface{}{},
			},
			"add": map[string]interface{}{
				"field_a": 10,
			},
		},
		"field_d": map[string]interface{}{},
	}
	var desc = new(Descriptor)
	*desc = Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssign",
		Children: []Field{
			{Name: "field_a", ID: 1, Desc: &Descriptor{Kind: TypeKind_Leaf, Type: "INT32"}},
			{Name: "field_b", ID: 2, Desc: &Descriptor{
				Kind:     TypeKind_List,
				Type:     "LIST",
				Children: []Field{{Name: "*", Desc: desc}},
			}},
			{Name: "field_c", ID: 3, Desc: &Descriptor{
				Kind:     TypeKind_StrMap,
				Type:     "MAP",
				Children: []Field{{Name: "*", Desc: desc}},
			}},
			{
				Name: "field_d",
				ID:   4,
				Desc: desc,
			},
		},
	}

	dest := &sampleAssign{
		FieldB: []*sampleAssign{
			{FieldA: 1, FieldE: "should not be cleared"},
			{FieldA: 1, FieldB: []*sampleAssign{{FieldA: 1}}, FieldE: "should not be cleared"},
		},
		FieldC: map[string]*sampleAssign{
			"empty":     {FieldA: 1, FieldE: "should not be cleared"},
			"non_empty": {FieldA: 1, FieldC: map[string]*sampleAssign{"a": {FieldA: 1}}, FieldE: "should not be cleared"},
		},
		FieldD: &sampleAssign{FieldA: 1, FieldE: "should not be cleared"},
	}

	assigner := &Assigner{}
	desc.Normalize()
	err := assigner.AssignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	expected := &sampleAssign{
		FieldB: []*sampleAssign{
			{FieldA: 1, FieldE: "should not be cleared"},
			{FieldA: 10, FieldB: []*sampleAssign{{FieldA: 1}}, FieldE: "should not be cleared"},
			{FieldA: 10},
		},
		FieldC: map[string]*sampleAssign{
			"empty":     {FieldA: 1, FieldE: "should not be cleared"},
			"non_empty": {FieldA: 10, FieldC: map[string]*sampleAssign{"a": {FieldA: 1}}, FieldE: "should not be cleared"},
			"add":       {FieldA: 10},
		},
		FieldD: &sampleAssign{FieldA: 1, FieldE: "should not be cleared"},
	}

	require.Equal(t, expected, dest)
}

// TestAssignAny_UnknownFields tests that the converted sample
// with XXX_unrecognized can be correctly serialized and deserialized
func TestAssignAny_UnknownFields(t *testing.T) {
	// Step 1: Create a sampleAssignSmall with unknown fields
	src := map[string]interface{}{
		"field_a":     42,
		"field_e":     "hello",
		"unknown_int": 100,      // Field ID 10
		"unknown_str": "secret", // Field ID 11
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssignSmall",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{Name: "field_e", ID: 5},
			{Name: "unknown_int", ID: 10},
			{Name: "unknown_str", ID: 11},
		},
	}

	dest := &sampleAssignSmall{}
	err := assignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	// Step 2: Serialize sampleAssignSmall to protobuf binary
	// We need to manually serialize this since sampleAssignSmall is not a generated proto message
	bp := binary.NewBinaryProtocolBuffer()
	defer binary.FreeBinaryProtocol(bp)

	// Write field_a (field ID 1, varint)
	if dest.FieldA != nil {
		bp.AppendTag(1, 0) // 0 = varint wire type
		bp.WriteInt32(int32(*dest.FieldA))
	}

	// Write field_e (field ID 5, string/bytes)
	if dest.FieldE != "" {
		bp.AppendTag(5, 2) // 2 = length-delimited wire type
		bp.WriteString(dest.FieldE)
	}

	// Write XXX_unrecognized fields (these contain unknown_int and unknown_str)
	if len(dest.XXX_unrecognized) > 0 {
		bp.Buf = append(bp.Buf, dest.XXX_unrecognized...)
	}

	serializedData := bp.Buf

	// Step 3: Create a dynamic proto message descriptor using official protobuf reflect API
	// This defines the full structure including all fields (known and unknown)
	messageDesc := &descriptorpb.DescriptorProto{
		Name: proto.String("SampleAssignFull"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   proto.String("field_a"),
				Number: proto.Int32(1),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum(),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			},
			{
				Name:   proto.String("field_e"),
				Number: proto.Int32(5),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			},
			{
				Name:   proto.String("unknown_int"),
				Number: proto.Int32(10),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum(),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			},
			{
				Name:   proto.String("unknown_str"),
				Number: proto.Int32(11),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			},
		},
	}

	// Create file descriptor
	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:        proto.String("test.proto"),
		Syntax:      proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{messageDesc},
	}

	// Build the descriptor using protodesc
	fd, err := protodesc.NewFile(fileDesc, nil)
	if err != nil {
		t.Fatalf("failed to create file descriptor: %v", err)
	}

	msgDesc := fd.Messages().Get(0)

	// Step 4: Create a dynamic message and unmarshal using official proto.Unmarshal
	dynamicMsg := dynamicpb.NewMessage(msgDesc)
	err = proto.Unmarshal(serializedData, dynamicMsg)
	if err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}

	// Step 5: Verify that all fields are correctly deserialized
	fields := dynamicMsg.Descriptor().Fields()

	fieldA := dynamicMsg.Get(fields.ByNumber(1)).Int()
	if fieldA != 42 {
		t.Errorf("field_a: expected 42, got %v", fieldA)
	}

	fieldE := dynamicMsg.Get(fields.ByNumber(5)).String()
	if fieldE != "hello" {
		t.Errorf("field_e: expected 'hello', got %v", fieldE)
	}

	unknownInt := dynamicMsg.Get(fields.ByNumber(10)).Int()
	if unknownInt != 100 {
		t.Errorf("unknown_int: expected 100, got %v", unknownInt)
	}

	unknownStr := dynamicMsg.Get(fields.ByNumber(11)).String()
	if unknownStr != "secret" {
		t.Errorf("unknown_str: expected 'secret', got %v", unknownStr)
	}

	t.Logf("Successfully verified protobuf serialization with unknown fields using official proto")
	t.Logf("  field_a: %v", fieldA)
	t.Logf("  field_e: %v", fieldE)
	t.Logf("  unknown_int: %v", unknownInt)
	t.Logf("  unknown_str: %v", unknownStr)
}

// Test struct for nested unknown fields
type sampleNestedUnknown struct {
	FieldA           int    `protobuf:"varint,1,opt,name=field_a" json:"field_a,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func TestEncodeUnknownField_NestedStruct(t *testing.T) {
	// Test nested struct encoding
	src := map[string]interface{}{
		"field_a": 42,
		"nested_struct": map[string]interface{}{
			"inner_field1": "hello",
			"inner_field2": int64(123),
		},
	}

	// Create descriptor for the struct with nested struct field
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleNestedUnknown",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{
				Name: "nested_struct",
				ID:   2,
				Desc: &Descriptor{
					Kind: TypeKind_Struct,
					Type: "NestedStruct",
					Children: []Field{
						{Name: "inner_field1", ID: 1},
						{Name: "inner_field2", ID: 2},
					},
				},
			},
		},
	}

	dest := &sampleNestedUnknown{}
	err := assignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	// Verify field_a is assigned correctly
	if dest.FieldA != 42 {
		t.Errorf("field_a: expected 42, got %v", dest.FieldA)
	}

	// Verify XXX_unrecognized contains the nested struct
	if len(dest.XXX_unrecognized) == 0 {
		t.Fatal("XXX_unrecognized should not be empty")
	}

	// Create protobuf descriptor for verification
	messageDesc := &descriptorpb.DescriptorProto{
		Name: proto.String("SampleNestedUnknown"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   proto.String("field_a"),
				Number: proto.Int32(1),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum(),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			},
			{
				Name:     proto.String("nested_struct"),
				Number:   proto.Int32(2),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				TypeName: proto.String(".NestedStruct"),
			},
		},
	}

	nestedMessageDesc := &descriptorpb.DescriptorProto{
		Name: proto.String("NestedStruct"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   proto.String("inner_field1"),
				Number: proto.Int32(1),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			},
			{
				Name:   proto.String("inner_field2"),
				Number: proto.Int32(2),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum(),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			},
		},
	}

	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:        proto.String("test.proto"),
		Syntax:      proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{messageDesc, nestedMessageDesc},
	}

	fd, err := protodesc.NewFile(fileDesc, nil)
	if err != nil {
		t.Fatalf("failed to create file descriptor: %v", err)
	}

	msgDesc := fd.Messages().Get(0)

	// Serialize the complete message
	bp := binary.NewBinaryProtocolBuffer()
	defer binary.FreeBinaryProtocol(bp)

	bp.AppendTag(1, 0)
	bp.WriteInt32(int32(dest.FieldA))
	bp.Buf = append(bp.Buf, dest.XXX_unrecognized...)

	// Unmarshal and verify
	dynamicMsg := dynamicpb.NewMessage(msgDesc)
	err = proto.Unmarshal(bp.Buf, dynamicMsg)
	if err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}

	fields := dynamicMsg.Descriptor().Fields()
	fieldA := dynamicMsg.Get(fields.ByNumber(1)).Int()
	if fieldA != 42 {
		t.Errorf("field_a: expected 42, got %v", fieldA)
	}

	nestedMsg := dynamicMsg.Get(fields.ByNumber(2)).Message()
	nestedFields := nestedMsg.Descriptor().Fields()
	innerField1 := nestedMsg.Get(nestedFields.ByNumber(1)).String()
	innerField2 := nestedMsg.Get(nestedFields.ByNumber(2)).Int()

	if innerField1 != "hello" {
		t.Errorf("nested_struct.inner_field1: expected 'hello', got %v", innerField1)
	}
	if innerField2 != 123 {
		t.Errorf("nested_struct.inner_field2: expected 123, got %v", innerField2)
	}

	t.Logf("Successfully verified nested struct encoding")
	t.Logf("  field_a: %v", fieldA)
	t.Logf("  nested_struct.inner_field1: %v", innerField1)
	t.Logf("  nested_struct.inner_field2: %v", innerField2)
}

func TestEncodeUnknownField_NestedList(t *testing.T) {
	// Test nested list encoding
	src := map[string]interface{}{
		"field_a": 42,
		"nested_list": []interface{}{
			map[string]interface{}{
				"item_field": "item1",
			},
			map[string]interface{}{
				"item_field": "item2",
			},
		},
	}

	// Create descriptor for the struct with nested list field
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleNestedUnknown",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{
				Name: "nested_list",
				ID:   3,
				Desc: &Descriptor{
					Kind: TypeKind_List,
					Type: "NestedList",
					Children: []Field{
						{
							Name: "*",
							ID:   0,
							Desc: &Descriptor{
								Kind: TypeKind_Struct,
								Type: "ListItem",
								Children: []Field{
									{Name: "item_field", ID: 1},
								},
							},
						},
					},
				},
			},
		},
	}

	dest := &sampleNestedUnknown{}
	err := assignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	// Verify field_a is assigned correctly
	if dest.FieldA != 42 {
		t.Errorf("field_a: expected 42, got %v", dest.FieldA)
	}

	// Verify XXX_unrecognized contains the nested list
	if len(dest.XXX_unrecognized) == 0 {
		t.Fatal("XXX_unrecognized should not be empty")
	}

	// Create protobuf descriptor for verification
	listItemDesc := &descriptorpb.DescriptorProto{
		Name: proto.String("ListItem"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   proto.String("item_field"),
				Number: proto.Int32(1),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			},
		},
	}

	messageDesc := &descriptorpb.DescriptorProto{
		Name: proto.String("SampleNestedUnknown"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   proto.String("field_a"),
				Number: proto.Int32(1),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum(),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			},
			{
				Name:     proto.String("nested_list"),
				Number:   proto.Int32(3),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
				TypeName: proto.String(".ListItem"),
			},
		},
	}

	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:        proto.String("test.proto"),
		Syntax:      proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{messageDesc, listItemDesc},
	}

	fd, err := protodesc.NewFile(fileDesc, nil)
	if err != nil {
		t.Fatalf("failed to create file descriptor: %v", err)
	}

	msgDesc := fd.Messages().Get(0)

	// Serialize the complete message
	bp := binary.NewBinaryProtocolBuffer()
	defer binary.FreeBinaryProtocol(bp)

	bp.AppendTag(1, 0)
	bp.WriteInt32(int32(dest.FieldA))
	bp.Buf = append(bp.Buf, dest.XXX_unrecognized...)

	// Unmarshal and verify
	dynamicMsg := dynamicpb.NewMessage(msgDesc)
	err = proto.Unmarshal(bp.Buf, dynamicMsg)
	if err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}

	fields := dynamicMsg.Descriptor().Fields()
	fieldA := dynamicMsg.Get(fields.ByNumber(1)).Int()
	if fieldA != 42 {
		t.Errorf("field_a: expected 42, got %v", fieldA)
	}

	nestedList := dynamicMsg.Get(fields.ByNumber(3)).List()
	if nestedList.Len() != 2 {
		t.Fatalf("nested_list: expected 2 items, got %v", nestedList.Len())
	}

	for i := 0; i < nestedList.Len(); i++ {
		itemMsg := nestedList.Get(i).Message()
		itemFields := itemMsg.Descriptor().Fields()
		itemField := itemMsg.Get(itemFields.ByNumber(1)).String()
		expectedValue := "item" + string(rune('1'+i))
		if itemField != expectedValue {
			t.Errorf("nested_list[%d].item_field: expected '%s', got %v", i, expectedValue, itemField)
		}
		t.Logf("  nested_list[%d].item_field: %v", i, itemField)
	}

	t.Logf("Successfully verified nested list encoding")
}

func TestEncodeUnknownField_NestedMap(t *testing.T) {
	// Test nested map encoding
	src := map[string]interface{}{
		"field_a": 42,
		"nested_map": map[string]interface{}{
			"key1": map[string]interface{}{
				"value_field": "value1",
			},
			"key2": map[string]interface{}{
				"value_field": "value2",
			},
		},
	}

	// Create descriptor for the struct with nested map field
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleNestedUnknown",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{
				Name: "nested_map",
				ID:   4,
				Desc: &Descriptor{
					Kind: TypeKind_StrMap,
					Type: "NestedMap",
					Children: []Field{
						{
							Name: "*",
							ID:   0,
							Desc: &Descriptor{
								Kind: TypeKind_Struct,
								Type: "MapValue",
								Children: []Field{
									{Name: "value_field", ID: 1},
								},
							},
						},
					},
				},
			},
		},
	}

	dest := &sampleNestedUnknown{}
	err := assignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	// Verify field_a is assigned correctly
	if dest.FieldA != 42 {
		t.Errorf("field_a: expected 42, got %v", dest.FieldA)
	}

	// Verify XXX_unrecognized contains the nested map
	if len(dest.XXX_unrecognized) == 0 {
		t.Fatal("XXX_unrecognized should not be empty")
	}

	// Create protobuf descriptor for verification
	mapValueDesc := &descriptorpb.DescriptorProto{
		Name: proto.String("MapValue"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   proto.String("value_field"),
				Number: proto.Int32(1),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			},
		},
	}

	// Protobuf maps are represented as repeated message with key/value fields
	mapEntryDesc := &descriptorpb.DescriptorProto{
		Name: proto.String("NestedMapEntry"),
		Options: &descriptorpb.MessageOptions{
			MapEntry: proto.Bool(true),
		},
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   proto.String("key"),
				Number: proto.Int32(1),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			},
			{
				Name:     proto.String("value"),
				Number:   proto.Int32(2),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
				TypeName: proto.String(".MapValue"),
			},
		},
	}

	messageDesc := &descriptorpb.DescriptorProto{
		Name: proto.String("SampleNestedUnknown"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:   proto.String("field_a"),
				Number: proto.Int32(1),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_INT32.Enum(),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			},
			{
				Name:     proto.String("nested_map"),
				Number:   proto.Int32(4),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
				TypeName: proto.String(".SampleNestedUnknown.NestedMapEntry"),
			},
		},
	}

	// Add the map entry as a nested message within the parent message
	messageDesc.NestedType = []*descriptorpb.DescriptorProto{mapEntryDesc}

	fileDesc := &descriptorpb.FileDescriptorProto{
		Name:        proto.String("test.proto"),
		Syntax:      proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{messageDesc, mapValueDesc},
	}

	fd, err := protodesc.NewFile(fileDesc, nil)
	if err != nil {
		t.Fatalf("failed to create file descriptor: %v", err)
	}

	msgDesc := fd.Messages().Get(0)

	// Serialize the complete message
	bp := binary.NewBinaryProtocolBuffer()
	defer binary.FreeBinaryProtocol(bp)

	bp.AppendTag(1, 0)
	bp.WriteInt32(int32(dest.FieldA))
	bp.Buf = append(bp.Buf, dest.XXX_unrecognized...)

	// Unmarshal and verify
	dynamicMsg := dynamicpb.NewMessage(msgDesc)
	err = proto.Unmarshal(bp.Buf, dynamicMsg)
	if err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}

	fields := dynamicMsg.Descriptor().Fields()
	fieldA := dynamicMsg.Get(fields.ByNumber(1)).Int()
	if fieldA != 42 {
		t.Errorf("field_a: expected 42, got %v", fieldA)
	}

	nestedMap := dynamicMsg.Get(fields.ByNumber(4)).Map()
	if nestedMap.Len() != 2 {
		t.Fatalf("nested_map: expected 2 entries, got %v", nestedMap.Len())
	}

	// Collect map entries
	mapEntries := make(map[string]string)
	nestedMap.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		key := k.String()
		valueMsg := v.Message()
		valueFields := valueMsg.Descriptor().Fields()
		value := valueMsg.Get(valueFields.ByNumber(1)).String()
		mapEntries[key] = value
		t.Logf("  nested_map[%s].value_field: %v", key, value)
		return true
	})

	if mapEntries["key1"] != "value1" {
		t.Errorf("nested_map[key1].value_field: expected 'value1', got %v", mapEntries["key1"])
	}
	if mapEntries["key2"] != "value2" {
		t.Errorf("nested_map[key2].value_field: expected 'value2', got %v", mapEntries["key2"])
	}

	t.Logf("Successfully verified nested map encoding")
}

// TestAssignAny_NoUnkeyedLiteral tests that unknown fields are also stored in XXX_NoUnkeyedLiteral
func TestAssignAny_NoUnkeyedLiteral(t *testing.T) {
	src := map[string]interface{}{
		"field_a":     42,
		"field_e":     "hello",
		"unknown_int": 100,
		"unknown_str": "secret",
		"unknown_map": map[string]interface{}{
			"nested_key": "nested_value",
		},
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssignSmall",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{Name: "field_e", ID: 5},
			{Name: "unknown_int", ID: 10},
			{Name: "unknown_str", ID: 11},
			{Name: "unknown_map", ID: 12},
		},
	}

	dest := &sampleAssignSmall{}
	err := assignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	// Verify known fields
	if dest.FieldA == nil || *dest.FieldA != 42 {
		t.Errorf("field_a: expected 42, got %v", dest.FieldA)
	}
	if dest.FieldE != "hello" {
		t.Errorf("field_e: expected 'hello', got %v", dest.FieldE)
	}

	// Verify XXX_unrecognized is set (protobuf binary encoded)
	if len(dest.XXX_unrecognized) == 0 {
		t.Errorf("XXX_unrecognized: expected non-empty, got empty")
	}

	// Verify XXX_NoUnkeyedLiteral contains the raw unknown field values
	if dest.XXX_NoUnkeyedLiteral == nil {
		t.Fatalf("XXX_NoUnkeyedLiteral: expected non-nil map")
	}

	// Check unknown_int
	if val, ok := dest.XXX_NoUnkeyedLiteral["unknown_int"]; !ok {
		t.Errorf("XXX_NoUnkeyedLiteral: expected 'unknown_int' key")
	} else if intVal, ok := val.(int); !ok || intVal != 100 {
		t.Errorf("XXX_NoUnkeyedLiteral['unknown_int']: expected 100, got %v (type %T)", val, val)
	}

	// Check unknown_str
	if val, ok := dest.XXX_NoUnkeyedLiteral["unknown_str"]; !ok {
		t.Errorf("XXX_NoUnkeyedLiteral: expected 'unknown_str' key")
	} else if strVal, ok := val.(string); !ok || strVal != "secret" {
		t.Errorf("XXX_NoUnkeyedLiteral['unknown_str']: expected 'secret', got %v (type %T)", val, val)
	}

	// Check unknown_map
	if val, ok := dest.XXX_NoUnkeyedLiteral["unknown_map"]; !ok {
		t.Errorf("XXX_NoUnkeyedLiteral: expected 'unknown_map' key")
	} else if mapVal, ok := val.(map[string]interface{}); !ok {
		t.Errorf("XXX_NoUnkeyedLiteral['unknown_map']: expected map[string]interface{}, got %T", val)
	} else if nestedVal, ok := mapVal["nested_key"]; !ok || nestedVal != "nested_value" {
		t.Errorf("XXX_NoUnkeyedLiteral['unknown_map']['nested_key']: expected 'nested_value', got %v", nestedVal)
	}

	t.Logf("Successfully verified XXX_NoUnkeyedLiteral and XXX_unrecognized for unknown fields")
	t.Logf("  field_a: %v", *dest.FieldA)
	t.Logf("  field_e: %v", dest.FieldE)
	t.Logf("  XXX_unrecognized length: %d", len(dest.XXX_unrecognized))
	t.Logf("  XXX_NoUnkeyedLiteral: %+v", dest.XXX_NoUnkeyedLiteral)
}

func TestAssignAny_ListOfStructs(t *testing.T) {

	src := map[string]interface{}{
		"field_b": []*sampleAssign{
			{
				FieldA: 2,
				FieldE: "first",
			},
			{
				FieldA: 2,
				FieldE: "second",
			},
		},
	}

	// nestedDesc := &Descriptor{
	// 	Kind: TypeKind_Struct,
	// 	Name: "SampleAssign",
	// 	Children: []Field{
	// 		{Name: "field_a", ID: 1},
	// 		{Name: "field_e", ID: 5},
	// 	},
	// }

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssign",
		Children: []Field{
			{
				Name: "field_b",
				ID:   2,
				Desc: &Descriptor{
					Kind: TypeKind_Leaf,
					Type: "LIST",
				},
			},
		},
	}

	dest := &sampleAssign{}
	err := assignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	if len(dest.FieldB) != 2 {
		t.Fatalf("field_b: expected length 2, got %d", len(dest.FieldB))
	}

	if dest.FieldB[0].FieldA != 2 {
		t.Errorf("field_b[0].field_a: expected 1, got %v", dest.FieldB[0].FieldA)
	}
	if dest.FieldB[0].FieldE != "first" {
		t.Errorf("field_b[0].field_e: expected 'first', got %v", dest.FieldB[0].FieldE)
	}
	if dest.FieldB[1].FieldA != 2 {
		t.Errorf("field_b[1].field_a: expected 2, got %v", dest.FieldB[1].FieldA)
	}
	if dest.FieldB[1].FieldE != "second" {
		t.Errorf("field_b[1].field_e: expected 'second', got %v", dest.FieldB[1].FieldE)
	}
}

func TestAssignAny_MapOfStructs(t *testing.T) {
	src := map[string]interface{}{
		"field_c": map[string]interface{}{
			"key1": map[string]interface{}{
				"field_a": 10,
				"field_e": "value1",
			},
			"key2": map[string]interface{}{
				"field_a": 20,
				"field_e": "value2",
			},
		},
	}

	nestedDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssign",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{Name: "field_e", ID: 5},
		},
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssign",
		Children: []Field{
			{
				Name: "field_c",
				ID:   3,
				Desc: &Descriptor{
					Kind: TypeKind_StrMap,
					Type: "MAP",
					Children: []Field{
						{Name: "*", Desc: nestedDesc},
					},
				},
			},
		},
	}

	dest := &sampleAssign{}
	err := assignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	if len(dest.FieldC) != 2 {
		t.Fatalf("field_c: expected length 2, got %d", len(dest.FieldC))
	}

	if dest.FieldC["key1"] == nil {
		t.Fatalf("field_c['key1']: expected non-nil")
	}
	if dest.FieldC["key1"].FieldA != 10 {
		t.Errorf("field_c['key1'].field_a: expected 10, got %v", dest.FieldC["key1"].FieldA)
	}
	if dest.FieldC["key1"].FieldE != "value1" {
		t.Errorf("field_c['key1'].field_e: expected 'value1', got %v", dest.FieldC["key1"].FieldE)
	}
}

// TestAssignAny_ListWithSpecificIndices tests assigning list elements by specific indices or wildcard
func TestAssignAny_ListWithSpecificIndices(t *testing.T) {
	// Test case 1: Wildcard - assign all elements
	t.Run("wildcard_all_elements", func(t *testing.T) {
		src := map[string]interface{}{
			"field_list": []interface{}{10, 20, 30, 40, 50},
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Type: "SampleAssign",
			Children: []Field{
				{
					Name: "field_list",
					ID:   6,
					Desc: &Descriptor{
						Kind: TypeKind_List,
						Type: "LIST",
						Children: []Field{
							{Name: "*"}, // wildcard - all elements
						},
					},
				},
			},
		}

		dest := &sampleAssign{}
		err := assignAny(desc, src, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		expected := []int{10, 20, 30, 40, 50}
		if !reflect.DeepEqual(dest.FieldList, expected) {
			t.Errorf("field_list: expected %v, got %v", expected, dest.FieldList)
		}
	})

	// Test case 2: Specific indices (0, 2, 4)
	// The source array's elements are mapped to destination indices specified by Field.ID
	// src[0] (10) -> dest[0], src[1] (20) -> dest[2], src[2] (30) -> dest[4]
	t.Run("specific_indices", func(t *testing.T) {
		src := map[string]interface{}{
			"field_list": []interface{}{10, 20, 30}, // 3 elements
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Type: "SampleAssign",
			Children: []Field{
				{
					Name: "field_list",
					ID:   6,
					Desc: &Descriptor{
						Kind: TypeKind_List,
						Type: "LIST",
						Children: []Field{
							{Name: "0", ID: 0}, // src[0] -> dest[0]
							{Name: "2", ID: 2}, // src[1] -> dest[2]
							{Name: "4", ID: 4}, // src[2] -> dest[4]
						},
					},
				},
			},
		}

		dest := &sampleAssign{}
		err := assignAny(desc, src, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		// Should create a slice with length maxIdx+1 = 5
		if len(dest.FieldList) != 5 {
			t.Fatalf("field_list: expected length 5, got %d", len(dest.FieldList))
		}

		// Check that specific indices are assigned
		if dest.FieldList[0] != 10 {
			t.Errorf("field_list[0]: expected 10, got %v", dest.FieldList[0])
		}
		if dest.FieldList[2] != 20 {
			t.Errorf("field_list[2]: expected 20, got %v", dest.FieldList[2])
		}
		if dest.FieldList[4] != 30 {
			t.Errorf("field_list[4]: expected 30, got %v", dest.FieldList[4])
		}

		// Indices 1 and 3 should be zero values
		if dest.FieldList[1] != 0 {
			t.Errorf("field_list[1]: expected 0 (zero value), got %v", dest.FieldList[1])
		}
		if dest.FieldList[3] != 0 {
			t.Errorf("field_list[3]: expected 0 (zero value), got %v", dest.FieldList[3])
		}
	})

	// Test case 3: Mapping to non-contiguous indices
	// src[0] -> dest[0], src[1] -> dest[1], src[2] -> dest[10]
	t.Run("non_contiguous_indices", func(t *testing.T) {
		src := map[string]interface{}{
			"field_list": []interface{}{10, 20, 30},
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Type: "SampleAssign",
			Children: []Field{
				{
					Name: "field_list",
					ID:   6,
					Desc: &Descriptor{
						Kind: TypeKind_List,
						Type: "LIST",
						Children: []Field{
							{Name: "0", ID: 0},   // src[0] -> dest[0]
							{Name: "1", ID: 1},   // src[1] -> dest[1]
							{Name: "10", ID: 10}, // src[2] -> dest[10]
						},
					},
				},
			},
		}

		dest := &sampleAssign{}
		err := assignAny(desc, src, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		// Should create a slice with length maxIdx+1 = 11
		if len(dest.FieldList) != 11 {
			t.Fatalf("field_list: expected length 11, got %d", len(dest.FieldList))
		}

		// Check assigned values
		if dest.FieldList[0] != 10 {
			t.Errorf("field_list[0]: expected 10, got %v", dest.FieldList[0])
		}
		if dest.FieldList[1] != 20 {
			t.Errorf("field_list[1]: expected 20, got %v", dest.FieldList[1])
		}
		if dest.FieldList[10] != 30 {
			t.Errorf("field_list[10]: expected 30, got %v", dest.FieldList[10])
		}

		// Indices 2-9 should be zero values
		for i := 2; i < 10; i++ {
			if dest.FieldList[i] != 0 {
				t.Errorf("field_list[%d]: expected 0 (zero value), got %v", i, dest.FieldList[i])
			}
		}
	})

	// Test case 4: DisallowNotDefined with insufficient source elements
	// Descriptor requires 3 elements but source only has 2
	t.Run("disallow_not_defined_insufficient_source", func(t *testing.T) {
		src := map[string]interface{}{
			"field_list": []interface{}{10, 20}, // only 2 elements
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Type: "SampleAssign",
			Children: []Field{
				{
					Name: "field_list",
					ID:   6,
					Desc: &Descriptor{
						Kind: TypeKind_List,
						Type: "LIST",
						Children: []Field{
							{Name: "0", ID: 0}, // src[0] -> dest[0]
							{Name: "1", ID: 1}, // src[1] -> dest[1]
							{Name: "2", ID: 2}, // src[2] doesn't exist!
						},
					},
				},
			},
		}

		assigner := Assigner{AssignOptions: AssignOptions{DisallowNotDefined: true}}
		dest := &sampleAssign{}
		desc.Normalize()
		err := assigner.AssignAny(desc, src, dest)
		if err == nil {
			t.Fatalf("expected ErrNotFound, got nil")
		}

		notFoundErr, ok := err.(ErrNotFound)
		if !ok {
			t.Fatalf("expected ErrNotFound, got %T: %v", err, err)
		}
		if notFoundErr.Parent.Type != "LIST" {
			t.Errorf("expected parent name 'LIST', got '%s'", notFoundErr.Parent.Type)
		}
	})

	// Test case 5: List with nested structures
	t.Run("list_with_nested_structs_wildcard", func(t *testing.T) {
		src := map[string]interface{}{
			"field_b": []interface{}{
				map[string]interface{}{
					"field_a": 1,
					"field_e": "first",
				},
				map[string]interface{}{
					"field_a": 2,
					"field_e": "second",
				},
				map[string]interface{}{
					"field_a": 3,
					"field_e": "third",
				},
			},
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Type: "SampleAssign",
			Children: []Field{
				{
					Name: "field_b",
					ID:   2,
					Desc: &Descriptor{
						Kind: TypeKind_List,
						Type: "LIST",
						Children: []Field{
							{
								Name: "*",
								Desc: &Descriptor{
									Kind: TypeKind_Struct,
									Type: "SampleAssign",
									Children: []Field{
										{Name: "field_a", ID: 1},
										{Name: "field_e", ID: 5},
									},
								},
							},
						},
					},
				},
			},
		}

		dest := &sampleAssign{}
		err := assignAny(desc, src, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		if len(dest.FieldB) != 3 {
			t.Fatalf("field_b: expected length 3, got %d", len(dest.FieldB))
		}

		if dest.FieldB[0].FieldA != 1 {
			t.Errorf("field_b[0].field_a: expected 1, got %v", dest.FieldB[0].FieldA)
		}
		if dest.FieldB[0].FieldE != "first" {
			t.Errorf("field_b[0].field_e: expected 'first', got %v", dest.FieldB[0].FieldE)
		}

		if dest.FieldB[2].FieldA != 3 {
			t.Errorf("field_b[2].field_a: expected 3, got %v", dest.FieldB[2].FieldA)
		}
		if dest.FieldB[2].FieldE != "third" {
			t.Errorf("field_b[2].field_e: expected 'third', got %v", dest.FieldB[2].FieldE)
		}
	})

	// Test case 6: List with nested structures and specific indices
	// src[0] -> dest[0], src[1] -> dest[2] (Note: assign copies ALL available fields from source)
	t.Run("list_with_nested_structs_specific_indices", func(t *testing.T) {
		src := map[string]interface{}{
			"field_b": []interface{}{
				map[string]interface{}{
					"field_a": 1,
					"field_e": "first",
				},
				map[string]interface{}{
					"field_a": 2,
					"field_e": "second",
				},
			},
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Type: "SampleAssign",
			Children: []Field{
				{
					Name: "field_b",
					ID:   2,
					Desc: &Descriptor{
						Kind: TypeKind_List,
						Type: "LIST",
						Children: []Field{
							{
								Name: "0",
								ID:   0, // src[0] -> dest[0]
								Desc: &Descriptor{
									Kind: TypeKind_Struct,
									Type: "SampleAssign",
									Children: []Field{
										{Name: "field_a", ID: 1},
										{Name: "field_e", ID: 5},
									},
								},
							},
							{
								Name: "2",
								ID:   2, // src[1] -> dest[2]
								Desc: &Descriptor{
									Kind: TypeKind_Struct,
									Type: "SampleAssign",
									Children: []Field{
										{Name: "field_a", ID: 1},
										{Name: "field_e", ID: 5},
									},
								},
							},
						},
					},
				},
			},
		}

		dest := &sampleAssign{}
		err := assignAny(desc, src, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		// Should create a slice with length maxIdx+1 = 3
		if len(dest.FieldB) != 3 {
			t.Fatalf("field_b: expected length 3, got %d", len(dest.FieldB))
		}

		// Check first element (src[0] -> dest[0])
		if dest.FieldB[0] == nil {
			t.Fatalf("field_b[0]: expected non-nil")
		}
		if dest.FieldB[0].FieldA != 1 {
			t.Errorf("field_b[0].field_a: expected 1, got %v", dest.FieldB[0].FieldA)
		}
		if dest.FieldB[0].FieldE != "first" {
			t.Errorf("field_b[0].field_e: expected 'first', got %v", dest.FieldB[0].FieldE)
		}

		// Index 1 should be nil (not assigned)
		if dest.FieldB[1] != nil {
			t.Errorf("field_b[1]: expected nil, got %v", dest.FieldB[1])
		}

		// Check element at index 2 (src[1] -> dest[2])
		if dest.FieldB[2] == nil {
			t.Fatalf("field_b[2]: expected non-nil")
		}
		if dest.FieldB[2].FieldA != 2 {
			t.Errorf("field_b[2].field_a: expected 2, got %v", dest.FieldB[2].FieldA)
		}
		if dest.FieldB[2].FieldE != "second" {
			t.Errorf("field_b[2].field_e: expected 'second', got %v", dest.FieldB[2].FieldE)
		}
	})

	// Test case 7: Preserve existing slice elements not accessed by descriptor
	// dest already has [elem0, elem1, elem2, elem3], descriptor only modifies indices 1 and 3
	t.Run("preserve_unmodified_elements", func(t *testing.T) {
		src := map[string]interface{}{
			"field_b": []interface{}{
				map[string]interface{}{
					"field_a": 100,
					"field_e": "new_first",
				},
				map[string]interface{}{
					"field_a": 200,
					"field_e": "new_second",
				},
			},
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Type: "SampleAssign",
			Children: []Field{
				{
					Name: "field_b",
					ID:   2,
					Desc: &Descriptor{
						Kind: TypeKind_List,
						Type: "LIST",
						Children: []Field{
							{
								Name: "1",
								ID:   1, // src[0] -> dest[1]
								Desc: &Descriptor{
									Kind: TypeKind_Struct,
									Type: "SampleAssign",
									Children: []Field{
										{Name: "field_a", ID: 1},
										{Name: "field_e", ID: 5},
									},
								},
							},
							{
								Name: "3",
								ID:   3, // src[1] -> dest[3]
								Desc: &Descriptor{
									Kind: TypeKind_Struct,
									Type: "SampleAssign",
									Children: []Field{
										{Name: "field_a", ID: 1},
										{Name: "field_e", ID: 5},
									},
								},
							},
						},
					},
				},
			},
		}

		// Pre-populate dest with existing elements
		dest := &sampleAssign{
			FieldB: []*sampleAssign{
				{FieldA: 1, FieldE: "original_0"},
				{FieldA: 2, FieldE: "original_1"},
				{FieldA: 3, FieldE: "original_2"},
				{FieldA: 4, FieldE: "original_3"},
			},
		}

		err := assignAny(desc, src, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		// Should have 4 elements (maxIdx+1 = 4)
		if len(dest.FieldB) != 4 {
			t.Fatalf("field_b: expected length 4, got %d", len(dest.FieldB))
		}

		// Index 0 should be preserved (not modified)
		if dest.FieldB[0] == nil {
			t.Fatalf("field_b[0]: expected non-nil")
		}
		if dest.FieldB[0].FieldA != 1 {
			t.Errorf("field_b[0].field_a: expected 1 (preserved), got %v", dest.FieldB[0].FieldA)
		}
		if dest.FieldB[0].FieldE != "original_0" {
			t.Errorf("field_b[0].field_e: expected 'original_0' (preserved), got %v", dest.FieldB[0].FieldE)
		}

		// Index 1 should be overwritten by src[0]
		if dest.FieldB[1] == nil {
			t.Fatalf("field_b[1]: expected non-nil")
		}
		if dest.FieldB[1].FieldA != 100 {
			t.Errorf("field_b[1].field_a: expected 100 (overwritten), got %v", dest.FieldB[1].FieldA)
		}
		if dest.FieldB[1].FieldE != "new_first" {
			t.Errorf("field_b[1].field_e: expected 'new_first' (overwritten), got %v", dest.FieldB[1].FieldE)
		}

		// Index 2 should be preserved (not modified)
		if dest.FieldB[2] == nil {
			t.Fatalf("field_b[2]: expected non-nil")
		}
		if dest.FieldB[2].FieldA != 3 {
			t.Errorf("field_b[2].field_a: expected 3 (preserved), got %v", dest.FieldB[2].FieldA)
		}
		if dest.FieldB[2].FieldE != "original_2" {
			t.Errorf("field_b[2].field_e: expected 'original_2' (preserved), got %v", dest.FieldB[2].FieldE)
		}

		// Index 3 should be overwritten by src[1]
		if dest.FieldB[3] == nil {
			t.Fatalf("field_b[3]: expected non-nil")
		}
		if dest.FieldB[3].FieldA != 200 {
			t.Errorf("field_b[3].field_a: expected 200 (overwritten), got %v", dest.FieldB[3].FieldA)
		}
		if dest.FieldB[3].FieldE != "new_second" {
			t.Errorf("field_b[3].field_e: expected 'new_second' (overwritten), got %v", dest.FieldB[3].FieldE)
		}
	})

	// Test case 8: Expand existing slice when descriptor requires larger size
	t.Run("expand_existing_slice", func(t *testing.T) {
		src := map[string]interface{}{
			"field_b": []interface{}{
				map[string]interface{}{
					"field_a": 999,
				},
			},
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Type: "SampleAssign",
			Children: []Field{
				{
					Name: "field_b",
					ID:   2,
					Desc: &Descriptor{
						Kind: TypeKind_List,
						Type: "LIST",
						Children: []Field{
							{
								Name: "5",
								ID:   5, // src[0] -> dest[5]
								Desc: &Descriptor{
									Kind: TypeKind_Struct,
									Type: "SampleAssign",
									Children: []Field{
										{Name: "field_a", ID: 1},
									},
								},
							},
						},
					},
				},
			},
		}

		// Pre-populate dest with smaller slice
		dest := &sampleAssign{
			FieldB: []*sampleAssign{
				{FieldA: 1, FieldE: "keep_0"},
				{FieldA: 2, FieldE: "keep_1"},
			},
		}

		err := assignAny(desc, src, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		// Should expand to length 6 (maxIdx+1 = 6)
		if len(dest.FieldB) != 6 {
			t.Fatalf("field_b: expected length 6, got %d", len(dest.FieldB))
		}

		// Original elements should be preserved
		if dest.FieldB[0].FieldA != 1 || dest.FieldB[0].FieldE != "keep_0" {
			t.Errorf("field_b[0]: expected {1, keep_0}, got {%v, %v}", dest.FieldB[0].FieldA, dest.FieldB[0].FieldE)
		}
		if dest.FieldB[1].FieldA != 2 || dest.FieldB[1].FieldE != "keep_1" {
			t.Errorf("field_b[1]: expected {2, keep_1}, got {%v, %v}", dest.FieldB[1].FieldA, dest.FieldB[1].FieldE)
		}

		// Indices 2-4 should be nil (newly created, not assigned)
		for i := 2; i <= 4; i++ {
			if dest.FieldB[i] != nil {
				t.Errorf("field_b[%d]: expected nil, got %v", i, dest.FieldB[i])
			}
		}

		// Index 5 should have the new value
		if dest.FieldB[5] == nil {
			t.Fatalf("field_b[5]: expected non-nil")
		}
		if dest.FieldB[5].FieldA != 999 {
			t.Errorf("field_b[5].field_a: expected 999, got %v", dest.FieldB[5].FieldA)
		}
	})

	// Test case 9: Wildcard overwrites all existing elements
	t.Run("wildcard_overwrites_all", func(t *testing.T) {
		src := map[string]interface{}{
			"field_b": []interface{}{
				map[string]interface{}{"field_a": 10},
				map[string]interface{}{"field_a": 20},
			},
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Type: "SampleAssign",
			Children: []Field{
				{
					Name: "field_b",
					ID:   2,
					Desc: &Descriptor{
						Kind: TypeKind_List,
						Type: "LIST",
						Children: []Field{
							{
								Name: "*",
								Desc: &Descriptor{
									Kind: TypeKind_Struct,
									Type: "SampleAssign",
									Children: []Field{
										{Name: "field_a", ID: 1},
									},
								},
							},
						},
					},
				},
			},
		}

		// Pre-populate dest with different size slice
		dest := &sampleAssign{
			FieldB: []*sampleAssign{
				{FieldA: 100, FieldE: "old_0"},
				{FieldA: 200, FieldE: "old_1"},
				{FieldA: 300, FieldE: "old_2"},
			},
		}

		err := assignAny(desc, src, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		// Wildcard should completely replace with source length
		if len(dest.FieldB) != 3 {
			t.Fatalf("field_b: expected length 2 (from source), got %d", len(dest.FieldB))
		}

		// All elements should be new values from source
		if dest.FieldB[0].FieldA != 10 {
			t.Errorf("field_b[0].field_a: expected 10, got %v", dest.FieldB[0].FieldA)
		}
		if dest.FieldB[1].FieldA != 20 {
			t.Errorf("field_b[1].field_a: expected 20, got %v", dest.FieldB[1].FieldA)
		}
	})
}

func TestAssignAny_NilValues(t *testing.T) {
	err := assignAny(nil, nil, nil)
	if err != nil {
		t.Errorf("expected nil error for nil inputs, got %v", err)
	}

	desc := &Descriptor{Kind: TypeKind_Struct, Type: "Test"}
	dest := &sampleAssign{}

	err = assignAny(desc, nil, dest)
	if err != nil {
		t.Errorf("expected nil error for nil src, got %v", err)
	}
}

func TestAssignAny_DisallowNotFound(t *testing.T) {
	src := map[string]interface{}{
		"field_a":     42,
		"nonexistent": 100,
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssign",
		Children: []Field{
			{Name: "field_a", ID: 1},
			// nonexistent is not in descriptor
		},
	}

	dest := &sampleAssign{}
	as := Assigner{AssignOptions{DisallowNotDefined: true}}
	desc.Normalize()
	err := as.AssignAny(desc, src, dest)
	if err == nil {
		t.Fatalf("expected error for nonexistent field with DisallowNotFound")
	}

	notFoundErr, ok := err.(ErrNotFound)
	if !ok {
		t.Fatalf("expected ErrNotFound, got %T", err)
	}
	if notFoundErr.Field.Name != "nonexistent" {
		t.Errorf("expected field name 'nonexistent', got %v", notFoundErr.Field.Name)
	}
}

// ===================== Circular Reference Tests =====================
// These tests verify that AssignAny can handle circular reference type descriptions.
// The key principle is: recursively process data until data is nil (src == nil).

// circularAssignNode represents a node that can reference itself (like a linked list)
type circularAssignNode struct {
	Value            int                 `protobuf:"varint,1,req,name=value" json:"value,omitempty"`
	Next             *circularAssignNode `protobuf:"bytes,2,opt,name=next" json:"next,omitempty"`
	XXX_unrecognized []byte              `json:"-"`
}

// circularAssignTree represents a tree node that can reference itself
type circularAssignTree struct {
	Value            int                 `protobuf:"varint,1,req,name=value" json:"value,omitempty"`
	Left             *circularAssignTree `protobuf:"bytes,2,opt,name=left" json:"left,omitempty"`
	Right            *circularAssignTree `protobuf:"bytes,3,opt,name=right" json:"right,omitempty"`
	XXX_unrecognized []byte              `json:"-"`
}

// makeCircularAssignDesc creates a descriptor that references itself (circular reference)
func makeCircularAssignDesc() *Descriptor {
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "CircularNode",
		Children: []Field{
			{Name: "value", ID: 1},
			{Name: "next", ID: 2},
		},
	}
	// Make it circular: next field's Desc points back to the same descriptor
	desc.Children[1].Desc = desc
	return desc
}

// makeCircularAssignTreeDesc creates a tree descriptor that references itself
func makeCircularAssignTreeDesc() *Descriptor {
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "CircularTree",
		Children: []Field{
			{Name: "value", ID: 1},
			{Name: "left", ID: 2},
			{Name: "right", ID: 3},
		},
	}
	// Make it circular: left, right fields' Desc point back to the same descriptor
	desc.Children[1].Desc = desc
	desc.Children[2].Desc = desc
	return desc
}

func TestAssignAny_CircularDescriptor_LinkedList(t *testing.T) {
	// Create source data: linked list 1 -> 2 -> 3 -> nil
	src := map[string]interface{}{
		"value": 1,
		"next": map[string]interface{}{
			"value": 2,
			"next": map[string]interface{}{
				"value": 3,
				// next is nil (absent)
			},
		},
	}

	desc := makeCircularAssignDesc()

	dest := &circularAssignNode{}
	err := assignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	// Verify the assigned structure
	if dest.Value != 1 {
		t.Errorf("value: expected 1, got %v", dest.Value)
	}

	if dest.Next == nil {
		t.Fatalf("next: expected non-nil")
	}
	if dest.Next.Value != 2 {
		t.Errorf("next.value: expected 2, got %v", dest.Next.Value)
	}

	if dest.Next.Next == nil {
		t.Fatalf("next.next: expected non-nil")
	}
	if dest.Next.Next.Value != 3 {
		t.Errorf("next.next.value: expected 3, got %v", dest.Next.Next.Value)
	}

	// The last node's next should be nil
	if dest.Next.Next.Next != nil {
		t.Errorf("next.next.next: expected nil, got %v", dest.Next.Next.Next)
	}
}

func TestAssignAny_CircularDescriptor_SingleNode(t *testing.T) {
	// Single node with nil next
	src := map[string]interface{}{
		"value": 42,
		// next is not present (nil)
	}

	desc := makeCircularAssignDesc()

	dest := &circularAssignNode{}
	err := assignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	if dest.Value != 42 {
		t.Errorf("value: expected 42, got %v", dest.Value)
	}

	if dest.Next != nil {
		t.Errorf("next: expected nil, got %v", dest.Next)
	}
}

func TestAssignAny_CircularDescriptor_Tree(t *testing.T) {
	// Create source data: binary tree
	//       1
	//      / \
	//     2   3
	//    /
	//   4
	src := map[string]interface{}{
		"value": 1,
		"left": map[string]interface{}{
			"value": 2,
			"left": map[string]interface{}{
				"value": 4,
			},
		},
		"right": map[string]interface{}{
			"value": 3,
		},
	}

	desc := makeCircularAssignTreeDesc()

	dest := &circularAssignTree{}
	err := assignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	// Verify root
	if dest.Value != 1 {
		t.Errorf("value: expected 1, got %v", dest.Value)
	}

	// Verify left subtree
	if dest.Left == nil {
		t.Fatalf("left: expected non-nil")
	}
	if dest.Left.Value != 2 {
		t.Errorf("left.value: expected 2, got %v", dest.Left.Value)
	}

	if dest.Left.Left == nil {
		t.Fatalf("left.left: expected non-nil")
	}
	if dest.Left.Left.Value != 4 {
		t.Errorf("left.left.value: expected 4, got %v", dest.Left.Left.Value)
	}

	// Verify right subtree
	if dest.Right == nil {
		t.Fatalf("right: expected non-nil")
	}
	if dest.Right.Value != 3 {
		t.Errorf("right.value: expected 3, got %v", dest.Right.Value)
	}
}

func TestAssignAny_CircularDescriptor_NilSrc(t *testing.T) {
	desc := makeCircularAssignDesc()

	dest := &circularAssignNode{Value: 999}

	// Assign with nil src should not modify dest
	err := assignAny(desc, nil, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	// Original value should be preserved
	if dest.Value != 999 {
		t.Errorf("value should be preserved, expected 999, got %v", dest.Value)
	}
}

func TestAssignAny_CircularDescriptor_DeepList(t *testing.T) {
	// Create a deep linked list (depth=100) as source
	depth := 100
	var src interface{}
	for i := depth; i > 0; i-- {
		node := map[string]interface{}{
			"value": i,
		}
		if src != nil {
			node["next"] = src
		}
		src = node
	}

	desc := makeCircularAssignDesc()

	dest := &circularAssignNode{}
	err := assignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	// Verify the structure by traversing
	current := dest
	for i := 1; i <= depth; i++ {
		if current.Value != i {
			t.Errorf("at depth %d: expected value %d, got %v", i, i, current.Value)
		}
		if i < depth {
			if current.Next == nil {
				t.Fatalf("at depth %d: expected non-nil next", i)
			}
			current = current.Next
		} else {
			// Last node should have nil next
			if current.Next != nil {
				t.Errorf("last node should have nil next")
			}
		}
	}
}

// circularAssignMapNode represents a node with a map that can contain circular references
type circularAssignMapNode struct {
	Name             string                            `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Children         map[string]*circularAssignMapNode `protobuf:"bytes,2,opt,name=children" json:"children,omitempty"`
	XXX_unrecognized []byte                            `json:"-"`
}

func makeCircularAssignMapDesc() *Descriptor {
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "CircularMapNode",
		Children: []Field{
			{Name: "name", ID: 1},
			{Name: "children", ID: 2},
		},
	}
	// Make children field circular: it's a map with values of the same type
	desc.Children[1].Desc = &Descriptor{
		Kind: TypeKind_StrMap,
		Type: "ChildrenMap",
		Children: []Field{
			{Name: "*", Desc: desc}, // Wildcard with circular reference
		},
	}
	return desc
}

func TestAssignAny_CircularDescriptor_MapOfNodes(t *testing.T) {
	// Create source data: tree-like structure using maps
	// root
	// ├── child1
	// │   └── grandchild1
	// └── child2
	src := map[string]interface{}{
		"name": "root",
		"children": map[string]interface{}{
			"child1": map[string]interface{}{
				"name": "child1",
				"children": map[string]interface{}{
					"grandchild1": map[string]interface{}{
						"name": "grandchild1",
						// children is nil
					},
				},
			},
			"child2": map[string]interface{}{
				"name": "child2",
				// children is nil
			},
		},
	}

	desc := makeCircularAssignMapDesc()

	dest := &circularAssignMapNode{}
	err := assignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	// Verify root
	if dest.Name != "root" {
		t.Errorf("name: expected 'root', got %v", dest.Name)
	}

	if dest.Children == nil {
		t.Fatalf("children: expected non-nil")
	}

	child1 := dest.Children["child1"]
	if child1 == nil {
		t.Fatalf("child1: expected non-nil")
	}
	if child1.Name != "child1" {
		t.Errorf("child1.name: expected 'child1', got %v", child1.Name)
	}

	if child1.Children == nil {
		t.Fatalf("child1.children: expected non-nil")
	}

	grandchild1 := child1.Children["grandchild1"]
	if grandchild1 == nil {
		t.Fatalf("grandchild1: expected non-nil")
	}
	if grandchild1.Name != "grandchild1" {
		t.Errorf("grandchild1.name: expected 'grandchild1', got %v", grandchild1.Name)
	}

	child2 := dest.Children["child2"]
	if child2 == nil {
		t.Fatalf("child2: expected non-nil")
	}
	if child2.Name != "child2" {
		t.Errorf("child2.name: expected 'child2', got %v", child2.Name)
	}
}

// TestAssignAny_CircularDescriptor_FetchThenAssign tests the full round-trip:
// fetch from a circular structure, then assign to another circular structure
func TestAssignAny_CircularDescriptor_FetchThenAssign(t *testing.T) {
	// This uses types from fetch_test.go
	// Create a linked list: 1 -> 2 -> 3 -> nil
	type circularFetchNode struct {
		Value int                `thrift:"Value,1" json:"value,omitempty"`
		Next  *circularFetchNode `thrift:"Next,2" json:"next,omitempty"`
	}

	srcList := &circularFetchNode{
		Value: 1,
		Next: &circularFetchNode{
			Value: 2,
			Next: &circularFetchNode{
				Value: 3,
				Next:  nil,
			},
		},
	}

	// Create circular descriptor for fetch
	fetchDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "CircularNode",
		Children: []Field{
			{Name: "value", ID: 1},
			{Name: "next", ID: 2},
		},
	}
	fetchDesc.Children[1].Desc = fetchDesc

	// Fetch
	fetched, err := fetchAny(fetchDesc, srcList)
	if err != nil {
		t.Fatalf("FetchAny failed: %v", err)
	}

	// Create circular descriptor for assign (with different IDs if needed)
	assignDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "CircularNode",
		Children: []Field{
			{Name: "value", ID: 1},
			{Name: "next", ID: 2},
		},
	}
	assignDesc.Children[1].Desc = assignDesc

	// Assign
	dest := &circularAssignNode{}
	err = assignAny(assignDesc, fetched, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	// Verify
	if dest.Value != 1 {
		t.Errorf("value: expected 1, got %v", dest.Value)
	}
	if dest.Next == nil || dest.Next.Value != 2 {
		t.Errorf("next.value: expected 2")
	}
	if dest.Next.Next == nil || dest.Next.Next.Value != 3 {
		t.Errorf("next.next.value: expected 3")
	}
	if dest.Next.Next.Next != nil {
		t.Errorf("next.next.next: expected nil")
	}
}

// TestAssignAny_PathTracking tests that error messages include the correct DSL path
func TestAssignAny_PathTracking(t *testing.T) {
	tests := []struct {
		name        string
		src         interface{}
		desc        *Descriptor
		expectedErr string
	}{
		{
			name: "field not found in root",
			src: map[string]interface{}{
				"unknown_field": 42,
			},
			desc: &Descriptor{
				Kind: TypeKind_Struct,
				Type: "SampleAssign",
				Children: []Field{
					{Name: "field_a", ID: 1},
				},
			},
			expectedErr: "not found unknown_field at SampleAssign: field 'unknown_field' not found in struct at path $.unknown_field",
		},
		{
			name: "field not found in nested struct",
			src: map[string]interface{}{
				"field_d": map[string]interface{}{
					"unknown_nested": 123,
				},
			},
			desc: &Descriptor{
				Kind: TypeKind_Struct,
				Type: "SampleAssign",
				Children: []Field{
					{
						Name: "field_d",
						ID:   4,
						Desc: &Descriptor{
							Kind: TypeKind_Struct,
							Type: "SampleAssign",
							Children: []Field{
								{Name: "field_a", ID: 1},
							},
						},
					},
				},
			},
			expectedErr: "not found unknown_nested at SampleAssign: field 'unknown_nested' not found in struct at path $.field_d.unknown_nested",
		},
		{
			name: "field not found in deeply nested struct",
			src: map[string]interface{}{
				"field_d": map[string]interface{}{
					"field_a": 1,
					"field_d": map[string]interface{}{
						"missing_field": "test",
					},
				},
			},
			desc: &Descriptor{
				Kind: TypeKind_Struct,
				Type: "SampleAssign",
				Children: []Field{
					{
						Name: "field_d",
						ID:   4,
						Desc: &Descriptor{
							Kind: TypeKind_Struct,
							Type: "SampleAssign",
							Children: []Field{
								{Name: "field_a", ID: 1},
								{
									Name: "field_d",
									ID:   4,
									Desc: &Descriptor{
										Kind: TypeKind_Struct,
										Type: "SampleAssign",
										Children: []Field{
											{Name: "field_a", ID: 1},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedErr: "not found missing_field at SampleAssign: field 'missing_field' not found in struct at path $.field_d.field_d.missing_field",
		},
		{
			name: "field not found in map",
			src: map[string]interface{}{
				"field_c": map[string]interface{}{
					"key1": map[string]interface{}{
						"bad_field": 999,
					},
				},
			},
			desc: &Descriptor{
				Kind: TypeKind_Struct,
				Type: "SampleAssign",
				Children: []Field{
					{
						Name: "field_c",
						ID:   3,
						Desc: &Descriptor{
							Kind: TypeKind_StrMap,
							Type: "MAP",
							Children: []Field{
								{
									Name: "*",
									Desc: &Descriptor{
										Kind: TypeKind_Struct,
										Type: "SampleAssign",
										Children: []Field{
											{Name: "field_a", ID: 1},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedErr: "not found bad_field at SampleAssign: field 'bad_field' not found in struct at path $.field_c[key1].bad_field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dest := &sampleAssign{}
			as := Assigner{AssignOptions{DisallowNotDefined: true}}
			tt.desc.Normalize()
			err := as.AssignAny(tt.desc, tt.src, dest)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("expected error:\n%s\ngot:\n%s", tt.expectedErr, err.Error())
			}
		})
	}
}

// TestAssignAny_PathTracking_TypeErrors tests path tracking for type mismatch errors
func TestAssignAny_PathTracking_TypeErrors(t *testing.T) {
	tests := []struct {
		name        string
		src         interface{}
		desc        *Descriptor
		errorSubstr string // Substring that should be in the error
	}{
		{
			name: "type error at root",
			src:  "not a map", // Should be map[string]interface{}
			desc: &Descriptor{
				Kind: TypeKind_Struct,
				Type: "SampleAssign",
				Children: []Field{
					{Name: "field_a", ID: 1},
				},
			},
			errorSubstr: "expected map[string]interface{} for struct at $",
		},
		{
			name: "type error in nested struct",
			src: map[string]interface{}{
				"field_d": "not a map", // Should be map[string]interface{}
			},
			desc: &Descriptor{
				Kind: TypeKind_Struct,
				Type: "SampleAssign",
				Children: []Field{
					{
						Name: "field_d",
						ID:   4,
						Desc: &Descriptor{
							Kind: TypeKind_Struct,
							Type: "SampleAssign",
							Children: []Field{
								{Name: "field_a", ID: 1},
							},
						},
					},
				},
			},
			errorSubstr: "expected map[string]interface{} for struct at $.field_d",
		},
		{
			name: "type error in map value",
			src: map[string]interface{}{
				"field_c": map[string]interface{}{
					"key1": []int{1, 2, 3}, // Should be a struct map
				},
			},
			desc: &Descriptor{
				Kind: TypeKind_Struct,
				Type: "SampleAssign",
				Children: []Field{
					{
						Name: "field_c",
						ID:   3,
						Desc: &Descriptor{
							Kind: TypeKind_StrMap,
							Type: "MAP",
							Children: []Field{
								{
									Name: "*",
									Desc: &Descriptor{
										Kind: TypeKind_Struct,
										Type: "SampleAssign",
										Children: []Field{
											{Name: "field_a", ID: 1},
										},
									},
								},
							},
						},
					},
				},
			},
			errorSubstr: "expected map[string]interface{} for struct at $.field_c[key1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dest := &sampleAssign{}
			err := assignAny(tt.desc, tt.src, dest)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !contains(err.Error(), tt.errorSubstr) {
				t.Errorf("expected error to contain:\n%s\ngot:\n%s", tt.errorSubstr, err.Error())
			}
		})
	}
}

// TestPathStack tests the pathStack implementation directly
func TestPathStack(t *testing.T) {
	tests := []struct {
		name     string
		ops      func(*pathStack)
		expected string
	}{
		{
			name: "empty stack",
			ops: func(s *pathStack) {
			},
			expected: "$",
		},
		{
			name: "single field",
			ops: func(s *pathStack) {
				s.push(TypeKind_Struct, "field_a", 1)
			},
			expected: "$.field_a",
		},
		{
			name: "nested fields",
			ops: func(s *pathStack) {
				s.push(TypeKind_Struct, "field_d", 4)
				s.push(TypeKind_Struct, "field_a", 1)
			},
			expected: "$.field_d.field_a",
		},
		{
			name: "map key",
			ops: func(s *pathStack) {
				s.push(TypeKind_Struct, "field_c", 3)
				s.push(TypeKind_StrMap, "my_key", 0)
			},
			expected: "$.field_c[my_key]",
		},
		{
			name: "array index",
			ops: func(s *pathStack) {
				s.push(TypeKind_Struct, "field_b", 2)
				s.push(TypeKind_List, "", 0)
			},
			expected: "$.field_b[0]",
		},
		{
			name: "complex path",
			ops: func(s *pathStack) {
				s.push(TypeKind_Struct, "root_field", 1)
				s.push(TypeKind_Struct, "map_field", 3)
				s.push(TypeKind_StrMap, "key1", 0)
				s.push(TypeKind_Struct, "nested", 4)
				s.push(TypeKind_Struct, "array", 2)
				s.push(TypeKind_List, "", 2)
			},
			expected: "$.root_field.map_field[key1].nested.array[2]",
		},
		{
			name: "push and pop",
			ops: func(s *pathStack) {
				s.push(TypeKind_Struct, "field_a", 1)
				s.push(TypeKind_Struct, "field_b", 2)
				s.pop()
			},
			expected: "$.field_a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stack := getStackFrames()
			defer putStackFrames(stack)

			tt.ops(stack)
			result := stack.buildPath()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestStackFramePool tests that the stack frame pool works correctly
func TestStackFramePool(t *testing.T) {
	// Get a stack from the pool
	frames1 := getStackFrames()
	if len(*frames1) != 0 {
		t.Errorf("expected empty frames, got length %d", len(*frames1))
	}
	if cap(*frames1) < 16 {
		t.Errorf("expected capacity >= 16, got %d", cap(*frames1))
	}

	// Use it and return it
	*frames1 = append(*frames1, stackFrame{name: "test", id: 1})
	putStackFrames(frames1)

	// Get another one - should be reused
	frames2 := getStackFrames()
	if len(*frames2) != 0 {
		t.Errorf("expected frames to be reset, got length %d", len(*frames2))
	}

	// Return it
	putStackFrames(frames2)
}

// TestAssignAny_PathTracking_Integration tests path tracking in a complex nested scenario
func TestAssignAny_PathTracking_Integration(t *testing.T) {
	// Create a complex nested structure
	src := map[string]interface{}{
		"field_a": 1,
		"field_c": map[string]interface{}{
			"item1": map[string]interface{}{
				"field_a": 10,
				"field_d": map[string]interface{}{
					"field_a":   20,
					"bad_field": "should fail", // This will cause error
				},
			},
		},
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssign",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{
				Name: "field_c",
				ID:   3,
				Desc: &Descriptor{
					Kind: TypeKind_StrMap,
					Type: "MAP",
					Children: []Field{
						{
							Name: "*",
							Desc: &Descriptor{
								Kind: TypeKind_Struct,
								Type: "SampleAssign",
								Children: []Field{
									{Name: "field_a", ID: 1},
									{
										Name: "field_d",
										ID:   4,
										Desc: &Descriptor{
											Kind: TypeKind_Struct,
											Type: "SampleAssign",
											Children: []Field{
												{Name: "field_a", ID: 1},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	dest := &sampleAssign{}
	as := Assigner{AssignOptions{DisallowNotDefined: true}}
	desc.Normalize()
	err := as.AssignAny(desc, src, dest)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	expectedPath := "$.field_c[item1].field_d.bad_field"
	if !contains(err.Error(), expectedPath) {
		t.Errorf("expected error to contain path %s, got: %s", expectedPath, err.Error())
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			anyIndex(s, substr)))
}

func anyIndex(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests for path tracking overhead

// BenchmarkAssignAny_NestedStruct tests performance with nested structures
func BenchmarkAssignAny_NestedStruct(b *testing.B) {
	src := map[string]interface{}{
		"field_a": 1,
		"field_d": map[string]interface{}{
			"field_a": 2,
			"field_d": map[string]interface{}{
				"field_a": 3,
				"field_e": "nested",
			},
		},
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssign",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{
				Name: "field_d",
				ID:   4,
				Desc: &Descriptor{
					Kind: TypeKind_Struct,
					Type: "SampleAssign",
					Children: []Field{
						{Name: "field_a", ID: 1},
						{
							Name: "field_d",
							ID:   4,
							Desc: &Descriptor{
								Kind: TypeKind_Struct,
								Type: "SampleAssign",
								Children: []Field{
									{Name: "field_a", ID: 1},
									{Name: "field_e", ID: 5},
								},
							},
						},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dest := &sampleAssign{}
		_ = assignAny(desc, src, dest)
	}
}

func BenchmarkAssignAny_WithUnknownFields(b *testing.B) {
	src := map[string]interface{}{
		"field_a":  42,
		"field_e":  "hello",
		"unknown1": 100,
		"unknown2": "secret",
		"unknown3": 3.14,
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssignSmall",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{Name: "field_e", ID: 5},
			{Name: "unknown1", ID: 10},
			{Name: "unknown2", ID: 11},
			{Name: "unknown3", ID: 12},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dest := &sampleAssignSmall{}
		_ = assignAny(desc, src, dest)
	}
}

// BenchmarkAssignAny_WithMap tests performance with map structures
func BenchmarkAssignAny_WithMap(b *testing.B) {
	src := map[string]interface{}{
		"field_a": 1,
		"field_c": map[string]interface{}{
			"key1": map[string]interface{}{
				"field_a": 10,
				"field_e": "value1",
			},
			"key2": map[string]interface{}{
				"field_a": 20,
				"field_e": "value2",
			},
			"key3": map[string]interface{}{
				"field_a": 30,
				"field_e": "value3",
			},
		},
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssign",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{
				Name: "field_c",
				ID:   3,
				Desc: &Descriptor{
					Kind: TypeKind_StrMap,
					Type: "MAP",
					Children: []Field{
						{
							Name: "*",
							Desc: &Descriptor{
								Kind: TypeKind_Struct,
								Type: "SampleAssign",
								Children: []Field{
									{Name: "field_a", ID: 1},
									{Name: "field_e", ID: 5},
								},
							},
						},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dest := &sampleAssign{}
		_ = assignAny(desc, src, dest)
	}
}

// BenchmarkAssignAny_ErrorPath tests performance when error occurs (path building)
func BenchmarkAssignAny_ErrorPath(b *testing.B) {
	src := map[string]interface{}{
		"field_a": 1,
		"field_d": map[string]interface{}{
			"field_a": 2,
			"unknown": 999, // This will cause error
		},
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssign",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{
				Name: "field_d",
				ID:   4,
				Desc: &Descriptor{
					Kind: TypeKind_Struct,
					Type: "SampleAssign",
					Children: []Field{
						{Name: "field_a", ID: 1},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dest := &sampleAssign{}
		as := Assigner{AssignOptions{DisallowNotDefined: true}}
		desc.Normalize()
		_ = as.AssignAny(desc, src, dest)
	}
}

// BenchmarkPathStack_Operations tests the performance of stack operations
func BenchmarkPathStack_Operations(b *testing.B) {
	b.Run("push_and_pop", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			stack := getStackFrames()
			for j := 0; j < 10; j++ {
				stack.push(TypeKind_Struct, "field", j)
			}
			for j := 0; j < 10; j++ {
				stack.pop()
			}
			putStackFrames(stack)
		}
	})

	b.Run("build_path_shallow", func(b *testing.B) {
		stack := getStackFrames()
		stack.push(TypeKind_Struct, "field_a", 1)
		stack.push(TypeKind_Struct, "field_b", 2)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = stack.buildPath()
		}
	})

	b.Run("build_path_deep", func(b *testing.B) {
		stack := getStackFrames()
		for i := 0; i < 10; i++ {
			stack.push(TypeKind_Struct, "field", i)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = stack.buildPath()
		}
	})

	b.Run("build_path_mixed", func(b *testing.B) {
		stack := getStackFrames()
		stack.push(TypeKind_Struct, "root", 1)
		stack.push(TypeKind_StrMap, "map_field", 3)
		stack.push(TypeKind_StrMap, "key1", 0)
		stack.push(TypeKind_Struct, "nested", 4)
		stack.push(TypeKind_List, "", 5)
		stack.push(TypeKind_Struct, "deep", 6)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = stack.buildPath()
		}
	})
}

// BenchmarkPathTracking_Overhead compares the overhead of path tracking
func BenchmarkPathTracking_Overhead(b *testing.B) {
	src := map[string]interface{}{
		"field_a": 1,
		"field_e": "test",
		"field_c": map[string]interface{}{
			"k1": map[string]interface{}{
				"field_a": 10,
			},
			"k2": map[string]interface{}{
				"field_a": 20,
			},
		},
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "SampleAssign",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{Name: "field_e", ID: 5},
			{
				Name: "field_c",
				ID:   3,
				Desc: &Descriptor{
					Kind: TypeKind_StrMap,
					Type: "MAP",
					Children: []Field{
						{
							Name: "*",
							Desc: &Descriptor{
								Kind: TypeKind_Struct,
								Type: "SampleAssign",
								Children: []Field{
									{Name: "field_a", ID: 1},
								},
							},
						},
					},
				},
			},
		},
	}

	b.Run("success_case", func(b *testing.B) {
		// Normal case - path tracking is active but never used for errors
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dest := &sampleAssign{}
			_ = assignAny(desc, src, dest)
		}
	})

	b.Run("error_case", func(b *testing.B) {
		// Error case - path is built when error occurs
		srcWithError := map[string]interface{}{
			"field_a": 1,
			"unknown": 999,
		}
		b.ResetTimer()
		desc.Normalize()
		for i := 0; i < b.N; i++ {
			dest := &sampleAssign{}
			as := Assigner{AssignOptions{DisallowNotDefined: true}}
			_ = as.AssignAny(desc, srcWithError, dest)
		}
	})
}

// Test types for AssignValue API
type SimpleStruct struct {
	Name  string  `json:"name"`
	Age   int     `json:"age"`
	Score float64 `json:"score"`
}

type NestedStruct struct {
	XXX_NoUnkeyedLiteral map[string]interface{} `json:"-"`
	ID                   int                    `json:"id"`
	Simple               SimpleStruct           `json:"simple"`
	PtrSimple            *SimpleStruct          `json:"ptr_simple"`
}

type ComplexStruct struct {
	Numbers   []int                   `json:"numbers"`
	Strings   []string                `json:"strings"`
	PtrSlice  []*SimpleStruct         `json:"ptr_slice"`
	StrMap    map[string]int          `json:"str_map"`
	StructMap map[string]SimpleStruct `json:"struct_map"`
	Nested    NestedStruct            `json:"nested"`
}

func TestAssignValue_BasicTypes(t *testing.T) {
	assigner := Assigner{}

	t.Run("int_to_int", func(t *testing.T) {
		var dest int
		err := assigner.AssignValue(42, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dest != 42 {
			t.Errorf("expected 42, got %d", dest)
		}
	})

	t.Run("int_to_int64", func(t *testing.T) {
		var dest int64
		err := assigner.AssignValue(42, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dest != 42 {
			t.Errorf("expected 42, got %d", dest)
		}
	})

	t.Run("int32_to_int64", func(t *testing.T) {
		var dest int64
		err := assigner.AssignValue(int32(100), &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dest != 100 {
			t.Errorf("expected 100, got %d", dest)
		}
	})

	t.Run("uint_to_int", func(t *testing.T) {
		var dest int
		err := assigner.AssignValue(uint(50), &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dest != 50 {
			t.Errorf("expected 50, got %d", dest)
		}
	})

	t.Run("float64_to_float32", func(t *testing.T) {
		var dest float32
		err := assigner.AssignValue(3.14, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dest < 3.13 || dest > 3.15 {
			t.Errorf("expected ~3.14, got %f", dest)
		}
	})

	t.Run("int_to_float64", func(t *testing.T) {
		var dest float64
		err := assigner.AssignValue(42, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dest != 42.0 {
			t.Errorf("expected 42.0, got %f", dest)
		}
	})

	t.Run("float64_to_int", func(t *testing.T) {
		var dest int
		err := assigner.AssignValue(42.9, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dest != 42 {
			t.Errorf("expected 42, got %d", dest)
		}
	})

	t.Run("string_to_string", func(t *testing.T) {
		var dest string
		err := assigner.AssignValue("hello", &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dest != "hello" {
			t.Errorf("expected 'hello', got %s", dest)
		}
	})

	t.Run("bool_to_bool", func(t *testing.T) {
		var dest bool
		err := assigner.AssignValue(true, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !dest {
			t.Errorf("expected true, got false")
		}
	})

	t.Run("nil_source", func(t *testing.T) {
		var dest int
		err := assigner.AssignValue(nil, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("non_pointer_dest", func(t *testing.T) {
		var dest int
		err := assigner.AssignValue(42, dest)
		if err == nil {
			t.Fatal("expected error for non-pointer dest")
		}
	})
}

func TestAssignValue_StructToStruct(t *testing.T) {
	assigner := Assigner{}

	t.Run("simple_struct_copy", func(t *testing.T) {
		src := SimpleStruct{Name: "Alice", Age: 30, Score: 95.5}
		dest := SimpleStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest.Name != "Alice" || dest.Age != 30 || dest.Score != 95.5 {
			t.Errorf("struct not copied correctly: %+v", dest)
		}
	})

	t.Run("partial_field_match", func(t *testing.T) {
		type SourceStruct struct {
			Name  string `json:"name"`
			Age   int    `json:"age"`
			Extra string `json:"extra"`
		}

		src := SourceStruct{Name: "Bob", Age: 25, Extra: "ignored"}
		dest := SimpleStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest.Name != "Bob" || dest.Age != 25 {
			t.Errorf("fields not matched correctly: %+v", dest)
		}
	})

	t.Run("pointer_source", func(t *testing.T) {
		src := &SimpleStruct{Name: "Charlie", Age: 35, Score: 88.0}
		dest := SimpleStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest.Name != "Charlie" || dest.Age != 35 {
			t.Errorf("pointer struct not copied correctly: %+v", dest)
		}
	})

	t.Run("nil_pointer_field", func(t *testing.T) {
		type StructWithPtr struct {
			Value *int `json:"value"`
		}

		src := StructWithPtr{Value: nil}
		dest := StructWithPtr{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest.Value != nil {
			t.Errorf("expected nil pointer, got %v", dest.Value)
		}
	})

	t.Run("nested_struct_with_XXX_NoUnkeyedLiteral", func(t *testing.T) {
		src := NestedStruct{
			XXX_NoUnkeyedLiteral: map[string]interface{}{
				"unknown_field1": 100,
				"unknown_field2": "secret_data",
				"unknown_nested": map[string]interface{}{
					"inner_key": "inner_value",
				},
			},
			ID: 42,
			Simple: SimpleStruct{
				Name:  "Alice",
				Age:   30,
				Score: 95.5,
			},
			PtrSimple: &SimpleStruct{
				Name:  "Bob",
				Age:   25,
				Score: 88.0,
			},
		}

		type NestedStruct2 struct {
			XXX_NoUnkeyedLiteral map[string]interface{} `json:"-"`
			ID                   int                    `json:"id"`
			Simple               SimpleStruct           `json:"simple"`
			PtrSimple            *SimpleStruct          `json:"ptr_simple"`
		}
		dest := NestedStruct2{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify regular fields
		if dest.ID != 42 {
			t.Errorf("ID: expected 42, got %d", dest.ID)
		}
		if dest.Simple.Name != "Alice" {
			t.Errorf("Simple.Name: expected 'Alice', got '%s'", dest.Simple.Name)
		}
		if dest.PtrSimple == nil {
			t.Fatalf("PtrSimple should not be nil")
		}
		if dest.PtrSimple.Name != "Bob" {
			t.Errorf("PtrSimple.Name: expected 'Bob', got '%s'", dest.PtrSimple.Name)
		}

		// Verify XXX_NoUnkeyedLiteral was copied
		if dest.XXX_NoUnkeyedLiteral == nil {
			t.Fatalf("XXX_NoUnkeyedLiteral should not be nil")
		}

		if len(dest.XXX_NoUnkeyedLiteral) != 3 {
			t.Errorf("XXX_NoUnkeyedLiteral: expected 3 entries, got %d", len(dest.XXX_NoUnkeyedLiteral))
		}

		// Verify unknown_field1
		if val, ok := dest.XXX_NoUnkeyedLiteral["unknown_field1"]; !ok {
			t.Errorf("XXX_NoUnkeyedLiteral: expected 'unknown_field1' key")
		} else if intVal, ok := val.(int); !ok || intVal != 100 {
			t.Errorf("XXX_NoUnkeyedLiteral['unknown_field1']: expected 100, got %v (type %T)", val, val)
		}

		// Verify unknown_field2
		if val, ok := dest.XXX_NoUnkeyedLiteral["unknown_field2"]; !ok {
			t.Errorf("XXX_NoUnkeyedLiteral: expected 'unknown_field2' key")
		} else if strVal, ok := val.(string); !ok || strVal != "secret_data" {
			t.Errorf("XXX_NoUnkeyedLiteral['unknown_field2']: expected 'secret_data', got %v (type %T)", val, val)
		}

		// Verify unknown_nested
		if val, ok := dest.XXX_NoUnkeyedLiteral["unknown_nested"]; !ok {
			t.Errorf("XXX_NoUnkeyedLiteral: expected 'unknown_nested' key")
		} else if mapVal, ok := val.(map[string]interface{}); !ok {
			t.Errorf("XXX_NoUnkeyedLiteral['unknown_nested']: expected map[string]interface{}, got %T", val)
		} else if innerVal, ok := mapVal["inner_key"]; !ok || innerVal != "inner_value" {
			t.Errorf("XXX_NoUnkeyedLiteral['unknown_nested']['inner_key']: expected 'inner_value', got %v", innerVal)
		}

		t.Logf("Successfully verified XXX_NoUnkeyedLiteral copy in NestedStruct")
		t.Logf("  XXX_NoUnkeyedLiteral: %+v", dest.XXX_NoUnkeyedLiteral)
	})

	t.Run("nested_struct_with_nil_XXX_NoUnkeyedLiteral", func(t *testing.T) {
		src := NestedStruct{
			XXX_NoUnkeyedLiteral: nil,
			ID:                   100,
			Simple: SimpleStruct{
				Name:  "Test",
				Age:   40,
				Score: 92.0,
			},
		}

		dest := NestedStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Regular fields should be copied
		if dest.ID != 100 {
			t.Errorf("ID: expected 100, got %d", dest.ID)
		}

		// XXX_NoUnkeyedLiteral should remain nil or empty
		if len(dest.XXX_NoUnkeyedLiteral) > 0 {
			t.Errorf("XXX_NoUnkeyedLiteral: expected nil or empty, got %v", dest.XXX_NoUnkeyedLiteral)
		}
	})

	t.Run("nested_struct_with_empty_XXX_NoUnkeyedLiteral", func(t *testing.T) {
		src := NestedStruct{
			XXX_NoUnkeyedLiteral: map[string]interface{}{},
			ID:                   200,
			Simple: SimpleStruct{
				Name:  "Empty",
				Age:   50,
				Score: 85.0,
			},
		}

		dest := NestedStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Regular fields should be copied
		if dest.ID != 200 {
			t.Errorf("ID: expected 200, got %d", dest.ID)
		}

		// Empty map should not cause issues
		if dest.Simple.Name != "Empty" {
			t.Errorf("Simple.Name: expected 'Empty', got '%s'", dest.Simple.Name)
		}
	})
}

func TestAssignValue_Slices(t *testing.T) {
	assigner := Assigner{}

	t.Run("int_slice", func(t *testing.T) {
		src := []int{1, 2, 3, 4, 5}
		var dest []int

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !reflect.DeepEqual(dest, src) {
			t.Errorf("slices not equal: got %v, want %v", dest, src)
		}
	})

	t.Run("string_slice", func(t *testing.T) {
		src := []string{"a", "b", "c"}
		var dest []string

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !reflect.DeepEqual(dest, src) {
			t.Errorf("slices not equal: got %v, want %v", dest, src)
		}
	})

	t.Run("int_to_int64_slice", func(t *testing.T) {
		src := []int{10, 20, 30}
		var dest []int64

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := []int64{10, 20, 30}
		if !reflect.DeepEqual(dest, expected) {
			t.Errorf("slices not equal: got %v, want %v", dest, expected)
		}
	})

	t.Run("struct_slice", func(t *testing.T) {
		src := []SimpleStruct{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
		}
		var dest []SimpleStruct

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(dest) != 2 {
			t.Fatalf("expected 2 elements, got %d", len(dest))
		}
		if dest[0].Name != "Alice" || dest[1].Name != "Bob" {
			t.Errorf("struct slice not copied correctly: %+v", dest)
		}
	})

	t.Run("pointer_slice", func(t *testing.T) {
		src := []*SimpleStruct{
			{Name: "Alice", Age: 30},
			nil,
			{Name: "Bob", Age: 25},
		}
		var dest []*SimpleStruct

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(dest) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(dest))
		}
		if dest[0] == nil || dest[0].Name != "Alice" {
			t.Errorf("first element incorrect: %+v", dest[0])
		}
		if dest[1] != nil {
			t.Errorf("expected nil at index 1, got %+v", dest[1])
		}
		if dest[2] == nil || dest[2].Name != "Bob" {
			t.Errorf("third element incorrect: %+v", dest[2])
		}
	})

	t.Run("nil_slice", func(t *testing.T) {
		var src []int = nil
		var dest []int

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest != nil {
			t.Errorf("expected nil slice, got %v", dest)
		}
	})

	t.Run("empty_slice", func(t *testing.T) {
		src := []int{}
		var dest []int

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(dest) != 0 {
			t.Errorf("expected empty slice, got %v", dest)
		}
	})
}

func TestAssignValue_Maps(t *testing.T) {
	assigner := Assigner{}

	t.Run("string_int_map", func(t *testing.T) {
		src := map[string]int{"a": 1, "b": 2, "c": 3}
		dest := make(map[string]int)

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !reflect.DeepEqual(dest, src) {
			t.Errorf("maps not equal: got %v, want %v", dest, src)
		}
	})

	t.Run("int_conversion_map", func(t *testing.T) {
		src := map[string]int32{"x": 10, "y": 20}
		dest := make(map[string]int64)

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest["x"] != 10 || dest["y"] != 20 {
			t.Errorf("map values not converted correctly: %v", dest)
		}
	})

	t.Run("struct_value_map", func(t *testing.T) {
		src := map[string]SimpleStruct{
			"alice": {Name: "Alice", Age: 30},
			"bob":   {Name: "Bob", Age: 25},
		}
		dest := make(map[string]SimpleStruct)

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(dest) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(dest))
		}
		if dest["alice"].Name != "Alice" || dest["bob"].Name != "Bob" {
			t.Errorf("struct map not copied correctly: %+v", dest)
		}
	})

	t.Run("pointer_value_map", func(t *testing.T) {
		src := map[string]*SimpleStruct{
			"alice": {Name: "Alice", Age: 30},
			"nil":   nil,
		}
		dest := make(map[string]*SimpleStruct)

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest["alice"] == nil || dest["alice"].Name != "Alice" {
			t.Errorf("map pointer incorrect: %+v", dest["alice"])
		}
		if dest["nil"] != nil {
			t.Errorf("expected nil pointer in map, got %+v", dest["nil"])
		}
	})

	t.Run("nil_map", func(t *testing.T) {
		var src map[string]int = nil
		var dest map[string]int

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAssignValue_MapToStruct(t *testing.T) {
	assigner := Assigner{}

	t.Run("basic_map_to_struct", func(t *testing.T) {
		src := map[string]interface{}{
			"name":  "Alice",
			"age":   30,
			"score": 95.5,
		}
		dest := SimpleStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest.Name != "Alice" || dest.Age != 30 || dest.Score != 95.5 {
			t.Errorf("map to struct failed: %+v", dest)
		}
	})

	t.Run("partial_fields", func(t *testing.T) {
		src := map[string]interface{}{
			"name": "Bob",
		}
		dest := SimpleStruct{Age: 100, Score: 50.0}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest.Name != "Bob" {
			t.Errorf("name not set: %s", dest.Name)
		}
		// Age and Score should remain unchanged
		if dest.Age != 100 || dest.Score != 50.0 {
			t.Errorf("other fields were modified: %+v", dest)
		}
	})

	t.Run("type_conversion_in_map", func(t *testing.T) {
		src := map[string]interface{}{
			"name":  "Charlie",
			"age":   int32(25),
			"score": float32(88.5),
		}
		dest := SimpleStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest.Name != "Charlie" || dest.Age != 25 {
			t.Errorf("type conversion failed: %+v", dest)
		}
	})

	t.Run("nil_value_in_map", func(t *testing.T) {
		src := map[string]interface{}{
			"name": "David",
			"age":  nil,
		}
		dest := SimpleStruct{Age: 50}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest.Name != "David" {
			t.Errorf("name not set: %s", dest.Name)
		}
		// Age should remain as is since nil won't overwrite
		if dest.Age != 50 {
			t.Errorf("age was modified: %d", dest.Age)
		}
	})

	t.Run("unknown_fields_ignored", func(t *testing.T) {
		src := map[string]interface{}{
			"name":    "Eve",
			"unknown": "should be ignored",
		}
		dest := SimpleStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest.Name != "Eve" {
			t.Errorf("name not set: %s", dest.Name)
		}
	})
}

func TestAssignValue_NestedStructs(t *testing.T) {
	assigner := Assigner{}

	t.Run("nested_struct", func(t *testing.T) {
		src := NestedStruct{
			ID: 1,
			Simple: SimpleStruct{
				Name:  "Alice",
				Age:   30,
				Score: 95.5,
			},
		}
		dest := NestedStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest.ID != 1 {
			t.Errorf("ID not set: %d", dest.ID)
		}
		if dest.Simple.Name != "Alice" || dest.Simple.Age != 30 {
			t.Errorf("nested struct not copied: %+v", dest.Simple)
		}
	})

	t.Run("nested_pointer_struct", func(t *testing.T) {
		src := NestedStruct{
			ID: 2,
			PtrSimple: &SimpleStruct{
				Name: "Bob",
				Age:  25,
			},
		}
		dest := NestedStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest.PtrSimple == nil {
			t.Fatal("PtrSimple is nil")
		}
		if dest.PtrSimple.Name != "Bob" || dest.PtrSimple.Age != 25 {
			t.Errorf("nested pointer struct not copied: %+v", dest.PtrSimple)
		}
	})

	t.Run("map_to_nested_struct", func(t *testing.T) {
		src := map[string]interface{}{
			"id": 3,
			"simple": map[string]interface{}{
				"name":  "Charlie",
				"age":   35,
				"score": 88.0,
			},
		}
		dest := NestedStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest.ID != 3 {
			t.Errorf("ID not set: %d", dest.ID)
		}
		if dest.Simple.Name != "Charlie" || dest.Simple.Age != 35 {
			t.Errorf("nested struct from map failed: %+v", dest.Simple)
		}
	})

	t.Run("map_to_nested_pointer_struct", func(t *testing.T) {
		src := map[string]interface{}{
			"id": 4,
			"ptr_simple": map[string]interface{}{
				"name": "David",
				"age":  40,
			},
		}
		dest := NestedStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest.PtrSimple == nil {
			t.Fatal("PtrSimple is nil")
		}
		if dest.PtrSimple.Name != "David" || dest.PtrSimple.Age != 40 {
			t.Errorf("nested pointer struct from map failed: %+v", dest.PtrSimple)
		}
	})
}

func TestAssignValue_ComplexNested(t *testing.T) {
	assigner := Assigner{}

	t.Run("struct_with_slices_and_maps", func(t *testing.T) {
		src := ComplexStruct{
			Numbers: []int{1, 2, 3},
			Strings: []string{"a", "b", "c"},
			StrMap:  map[string]int{"x": 10, "y": 20},
		}
		dest := ComplexStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !reflect.DeepEqual(dest.Numbers, src.Numbers) {
			t.Errorf("numbers not copied: %v", dest.Numbers)
		}
		if !reflect.DeepEqual(dest.Strings, src.Strings) {
			t.Errorf("strings not copied: %v", dest.Strings)
		}
		if !reflect.DeepEqual(dest.StrMap, src.StrMap) {
			t.Errorf("map not copied: %v", dest.StrMap)
		}
	})

	t.Run("pointer_slice_in_struct", func(t *testing.T) {
		src := ComplexStruct{
			PtrSlice: []*SimpleStruct{
				{Name: "Alice", Age: 30},
				{Name: "Bob", Age: 25},
			},
		}
		dest := ComplexStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(dest.PtrSlice) != 2 {
			t.Fatalf("expected 2 elements, got %d", len(dest.PtrSlice))
		}
		if dest.PtrSlice[0].Name != "Alice" || dest.PtrSlice[1].Name != "Bob" {
			t.Errorf("pointer slice not copied correctly: %+v", dest.PtrSlice)
		}
	})

	t.Run("struct_map_in_struct", func(t *testing.T) {
		src := ComplexStruct{
			StructMap: map[string]SimpleStruct{
				"alice": {Name: "Alice", Age: 30},
				"bob":   {Name: "Bob", Age: 25},
			},
		}
		dest := ComplexStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(dest.StructMap) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(dest.StructMap))
		}
		if dest.StructMap["alice"].Name != "Alice" {
			t.Errorf("struct map not copied correctly: %+v", dest.StructMap)
		}
	})

	t.Run("deeply_nested", func(t *testing.T) {
		src := ComplexStruct{
			Nested: NestedStruct{
				ID: 100,
				Simple: SimpleStruct{
					Name:  "DeepAlice",
					Age:   30,
					Score: 95.5,
				},
				PtrSimple: &SimpleStruct{
					Name:  "DeepBob",
					Age:   25,
					Score: 80.0,
				},
			},
		}
		dest := ComplexStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest.Nested.ID != 100 {
			t.Errorf("nested ID not copied: %d", dest.Nested.ID)
		}
		if dest.Nested.Simple.Name != "DeepAlice" {
			t.Errorf("deeply nested struct not copied: %+v", dest.Nested.Simple)
		}
		if dest.Nested.PtrSimple == nil || dest.Nested.PtrSimple.Name != "DeepBob" {
			t.Errorf("deeply nested pointer struct not copied: %+v", dest.Nested.PtrSimple)
		}
	})

	t.Run("map_to_complex_struct", func(t *testing.T) {
		src := map[string]interface{}{
			"numbers": []interface{}{1, 2, 3},
			"strings": []interface{}{"x", "y", "z"},
			"str_map": map[string]interface{}{"key1": 100, "key2": 200},
			"ptr_slice": []interface{}{
				map[string]interface{}{"name": "MapAlice", "age": 30},
				map[string]interface{}{"name": "MapBob", "age": 25},
			},
			"nested": map[string]interface{}{
				"id": 500,
				"simple": map[string]interface{}{
					"name": "NestedFromMap",
					"age":  45,
				},
			},
		}
		dest := ComplexStruct{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(dest.Numbers) != 3 || dest.Numbers[0] != 1 {
			t.Errorf("numbers not set from map: %v", dest.Numbers)
		}
		if len(dest.Strings) != 3 || dest.Strings[0] != "x" {
			t.Errorf("strings not set from map: %v", dest.Strings)
		}
		if dest.StrMap["key1"] != 100 {
			t.Errorf("str_map not set from map: %v", dest.StrMap)
		}
		if len(dest.PtrSlice) != 2 || dest.PtrSlice[0].Name != "MapAlice" {
			t.Errorf("ptr_slice not set from map: %+v", dest.PtrSlice)
		}
		if dest.Nested.ID != 500 || dest.Nested.Simple.Name != "NestedFromMap" {
			t.Errorf("nested not set from map: %+v", dest.Nested)
		}
	})
}

func TestAssignValue_EdgeCases(t *testing.T) {
	assigner := Assigner{}

	t.Run("both_nil", func(t *testing.T) {
		err := assigner.AssignValue(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("interface_wrapping", func(t *testing.T) {
		var src interface{} = 42
		var dest int

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if dest != 42 {
			t.Errorf("expected 42, got %d", dest)
		}
	})

	t.Run("interface_to_interface", func(t *testing.T) {
		var src interface{} = "hello"
		var dest interface{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if str, ok := dest.(string); !ok || str != "hello" {
			t.Errorf("expected 'hello', got %v", dest)
		}
	})

	t.Run("slice_with_interface_elements", func(t *testing.T) {
		src := []interface{}{1, "two", 3.0}
		var dest []interface{}

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(dest) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(dest))
		}
	})

	t.Run("map_with_interface_values", func(t *testing.T) {
		src := map[string]interface{}{
			"int":    42,
			"string": "hello",
			"float":  3.14,
		}
		dest := make(map[string]interface{})

		err := assigner.AssignValue(src, &dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(dest) != 3 {
			t.Fatalf("expected 3 entries, got %d", len(dest))
		}
		if dest["int"] != 42 {
			t.Errorf("int value incorrect: %v", dest["int"])
		}
	})

	t.Run("zero_values", func(t *testing.T) {
		var srcInt int = 0
		var destInt int = 100

		err := assigner.AssignValue(srcInt, &destInt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if destInt != 0 {
			t.Errorf("expected 0, got %d", destInt)
		}
	})
}

// BenchmarkComplexStruct_Comparison compares AssignValue vs JSON marshal/unmarshal
func BenchmarkAssignValue(b *testing.B) {
	// Create a complex source struct with nested data (shared by both sub-benchmarks)
	src := ComplexStruct{
		Numbers: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Strings: []string{"a", "b", "c", "d", "e"},
		PtrSlice: []*SimpleStruct{
			{Name: "Alice", Age: 30, Score: 95.5},
			{Name: "Bob", Age: 25, Score: 88.0},
			{Name: "Charlie", Age: 35, Score: 92.0},
		},
		StrMap: map[string]int{
			"key1": 100,
			"key2": 200,
			"key3": 300,
		},
		StructMap: map[string]SimpleStruct{
			"user1": {Name: "Dave", Age: 40, Score: 87.5},
			"user2": {Name: "Eve", Age: 28, Score: 91.0},
		},
		Nested: NestedStruct{
			XXX_NoUnkeyedLiteral: map[string]interface{}{
				"extra1": 999,
				"extra2": "hidden",
			},
			ID: 123,
			Simple: SimpleStruct{
				Name:  "NestedUser",
				Age:   45,
				Score: 89.5,
			},
			PtrSimple: &SimpleStruct{
				Name:  "NestedPtr",
				Age:   50,
				Score: 86.0,
			},
		},
	}

	b.Run("AssignValue", func(b *testing.B) {
		assigner := Assigner{}
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			dest := ComplexStruct{}
			_ = assigner.AssignValue(src, &dest)
		}
	})

	b.Run("JSON", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Marshal to JSON
			jsonData, err := sonic.Marshal(src)
			if err != nil {
				b.Fatalf("marshal failed: %v", err)
			}

			// Unmarshal back to struct
			dest := ComplexStruct{}
			err = sonic.Unmarshal(jsonData, &dest)
			if err != nil {
				b.Fatalf("unmarshal failed: %v", err)
			}
		}
	})
}

type sampleAssignArray struct {
	FieldArray [3]int `protobuf:"varint,1,opt,name=field_array" json:"field_array"`
}

func TestAssignAny_Array(t *testing.T) {
	t.Run("Array_SpecificIndices_Success", func(t *testing.T) {
		src := map[string]interface{}{
			"field_array": []interface{}{10, 30},
		}
		// Assign 10 to index 0, 30 to index 2
		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Type: "SampleAssignArray",
			Children: []Field{
				{
					Name: "field_array",
					ID:   1,
					Desc: &Descriptor{
						Kind: TypeKind_List,
						Children: []Field{
							{ID: 0}, // Maps src[0] (10) to dest[0]
							{ID: 2}, // Maps src[1] (30) to dest[2]
						},
					},
				},
			},
		}

		dest := &sampleAssignArray{}
		assigner := &Assigner{}
		desc.Normalize()
		err := assigner.AssignAny(desc, src, dest)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dest.FieldArray[0] != 10 {
			t.Errorf("expected dest[0] to be 10, got %d", dest.FieldArray[0])
		}
		if dest.FieldArray[1] != 0 {
			t.Errorf("expected dest[1] to be 0, got %d", dest.FieldArray[1])
		}
		if dest.FieldArray[2] != 30 {
			t.Errorf("expected dest[2] to be 30, got %d", dest.FieldArray[2])
		}
	})

	t.Run("Array_OutOfBounds", func(t *testing.T) {
		src := map[string]interface{}{
			"field_array": []interface{}{10},
		}
		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Children: []Field{
				{
					Name: "field_array",
					ID:   1,
					Desc: &Descriptor{
						Kind: TypeKind_List,
						Children: []Field{
							{ID: 3}, // Index 3 is out of bounds for [3]int
						},
					},
				},
			},
		}
		dest := &sampleAssignArray{}
		assigner := &Assigner{}
		desc.Normalize()
		err := assigner.AssignAny(desc, src, dest)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "out of bounds") {
			t.Errorf("expected out of bounds error, got: %v", err)
		}
	})

	t.Run("Array_MissingSource_DisallowNotDefined", func(t *testing.T) {
		src := map[string]interface{}{
			"field_array": []interface{}{10},
		}
		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Children: []Field{
				{
					Name: "field_array",
					ID:   1,
					Desc: &Descriptor{
						Kind: TypeKind_List,
						Children: []Field{
							{ID: 0},
							{ID: 2}, // Missing in source (src has len 1)
						},
					},
				},
			},
		}
		dest := &sampleAssignArray{}
		assigner := &Assigner{AssignOptions: AssignOptions{DisallowNotDefined: true}}
		desc.Normalize()
		err := assigner.AssignAny(desc, src, dest)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "not found in source") {
			t.Errorf("expected not found error, got: %v", err)
		}
	})

	t.Run("Array_NegativeIndex", func(t *testing.T) {
		src := map[string]interface{}{
			"field_array": []interface{}{10},
		}
		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Children: []Field{
				{
					Name: "field_array",
					ID:   1,
					Desc: &Descriptor{
						Kind: TypeKind_List,
						Children: []Field{
							{ID: 0},
							{ID: -1},
						},
					},
				},
			},
		}
		dest := &sampleAssignArray{}
		assigner := &Assigner{AssignOptions: AssignOptions{DisallowNotDefined: true}}
		desc.Normalize()
		err := assigner.AssignAny(desc, src, dest)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "out of bounds") {
			t.Errorf("expected out of bounds error, got: %v", err)
		}
	})
}
