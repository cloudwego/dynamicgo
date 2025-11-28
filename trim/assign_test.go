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
	"reflect"
	"testing"

	"github.com/cloudwego/dynamicgo/proto/binary"
)

type SampleAssign struct {
	FieldA           *int                     `protobuf:"varint,1,req,name=field_a"`
	FieldB           []*SampleAssign          `protobuf:"bytes,2,opt,name=field_b"`
	FieldC           map[string]*SampleAssign `protobuf:"bytes,3,opt,name=field_c"`
	FieldD           *SampleAssign            `protobuf:"bytes,4,opt,name=field_d"`
	FieldE           string                   `protobuf:"bytes,5,opt,name=field_e"`
	FieldList        []int                    `protobuf:"bytes,6,opt,name=field_list"`
	FieldMap         map[string]int           `protobuf:"bytes,7,opt,name=field_map"`
	XXX_unrecognized []byte                   `json:"-"`
}

// SampleAssignSmall is a struct with fewer fields than SampleAssign
// Used to test XXX_unrecognized field encoding
type SampleAssignSmall struct {
	FieldA           *int   `protobuf:"varint,1,req,name=field_a"`
	FieldE           string `protobuf:"bytes,5,opt,name=field_e"`
	XXX_unrecognized []byte `json:"-"`
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
		Name: "SampleAssign",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{Name: "field_e", ID: 5},
		},
	}

	dest := &SampleAssign{}
	err := AssignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	if dest.FieldA == nil || *dest.FieldA != 42 {
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
		Name: "SampleAssign",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{
				Name: "field_d",
				ID:   4,
				Desc: &Descriptor{
					Kind: TypeKind_Struct,
					Name: "SampleAssign",
					Children: []Field{
						{Name: "field_a", ID: 1},
						{Name: "field_e", ID: 5},
					},
				},
			},
		},
	}

	dest := &SampleAssign{}
	err := AssignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	if dest.FieldA == nil || *dest.FieldA != 1 {
		t.Errorf("field_a: expected 1, got %v", dest.FieldA)
	}
	if dest.FieldD == nil {
		t.Fatalf("field_d: expected non-nil")
	}
	if dest.FieldD.FieldA == nil || *dest.FieldD.FieldA != 2 {
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
		Name: "SampleAssign",
		Children: []Field{
			{
				Name: "field_list",
				ID:   6,
				Desc: &Descriptor{
					Kind: TypeKind_Scalar,
					Name: "LIST",
				},
			},
		},
	}

	dest := &SampleAssign{}
	err := AssignAny(desc, src, dest)
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
		Name: "SampleAssign",
		Children: []Field{
			{
				Name: "field_map",
				ID:   7,
				Desc: &Descriptor{
					Kind: TypeKind_StrMap,
					Name: "MAP",
					Children: []Field{
						{Name: "*"},
					},
				},
			},
		},
	}

	dest := &SampleAssign{}
	err := AssignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	expected := map[string]int{"key1": 100, "key2": 200}
	if !reflect.DeepEqual(dest.FieldMap, expected) {
		t.Errorf("field_map: expected %v, got %v", expected, dest.FieldMap)
	}
}

func TestAssignAny_UnknownFields(t *testing.T) {
	// Source has fields that don't exist in destination struct
	src := map[string]interface{}{
		"field_a":     42,
		"unknown_int": 100,      // Field ID 10
		"unknown_str": "secret", // Field ID 11
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Name: "SampleAssignSmall",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{Name: "unknown_int", ID: 10},
			{Name: "unknown_str", ID: 11},
		},
	}

	dest := &SampleAssignSmall{}
	err := AssignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	// Check known field
	if dest.FieldA == nil || *dest.FieldA != 42 {
		t.Errorf("field_a: expected 42, got %v", dest.FieldA)
	}

	// Check that XXX_unrecognized has data
	if len(dest.XXX_unrecognized) == 0 {
		t.Fatalf("XXX_unrecognized: expected non-empty bytes")
	}

	// Decode the unknown fields
	bp := binary.NewBinaryProtol(dest.XXX_unrecognized)
	defer binary.FreeBinaryProtocol(bp)

	foundInt := false
	foundStr := false

	for bp.Read < len(bp.Buf) {
		fieldNum, wireType, _, err := bp.ConsumeTag()
		if err != nil {
			t.Fatalf("failed to consume tag: %v", err)
		}

		switch fieldNum {
		case 10:
			// unknown_int
			val, err := bp.ReadInt64()
			if err != nil {
				t.Fatalf("failed to read int64: %v", err)
			}
			if val != 100 {
				t.Errorf("unknown_int: expected 100, got %v", val)
			}
			foundInt = true
		case 11:
			// unknown_str
			if wireType != 2 { // length-delimited
				t.Errorf("unknown_str: expected wire type 2, got %v", wireType)
			}
			val, err := bp.ReadString(true)
			if err != nil {
				t.Fatalf("failed to read string: %v", err)
			}
			if val != "secret" {
				t.Errorf("unknown_str: expected 'secret', got %v", val)
			}
			foundStr = true
		default:
			t.Errorf("unexpected field number: %v", fieldNum)
		}
	}

	if !foundInt {
		t.Errorf("unknown_int field not found in XXX_unrecognized")
	}
	if !foundStr {
		t.Errorf("unknown_str field not found in XXX_unrecognized")
	}
}

func TestAssignAny_ListOfStructs(t *testing.T) {

	src := map[string]interface{}{
		"field_b": []*SampleAssign{
			{
				FieldA: intPtr(1),
				FieldE: "first",
			},
			{
				FieldA: intPtr(2),
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
		Name: "SampleAssign",
		Children: []Field{
			{
				Name: "field_b",
				ID:   2,
				Desc: &Descriptor{
					Kind: TypeKind_Scalar,
					Name: "LIST",
				},
			},
		},
	}

	dest := &SampleAssign{}
	err := AssignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	if len(dest.FieldB) != 2 {
		t.Fatalf("field_b: expected length 2, got %d", len(dest.FieldB))
	}

	if dest.FieldB[0].FieldA == nil || *dest.FieldB[0].FieldA != 1 {
		t.Errorf("field_b[0].field_a: expected 1, got %v", dest.FieldB[0].FieldA)
	}
	if dest.FieldB[0].FieldE != "first" {
		t.Errorf("field_b[0].field_e: expected 'first', got %v", dest.FieldB[0].FieldE)
	}
	if dest.FieldB[1].FieldA == nil || *dest.FieldB[1].FieldA != 2 {
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
		Name: "SampleAssign",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{Name: "field_e", ID: 5},
		},
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Name: "SampleAssign",
		Children: []Field{
			{
				Name: "field_c",
				ID:   3,
				Desc: &Descriptor{
					Kind: TypeKind_StrMap,
					Name: "MAP",
					Children: []Field{
						{Name: "*", Desc: nestedDesc},
					},
				},
			},
		},
	}

	dest := &SampleAssign{}
	err := AssignAny(desc, src, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}

	if len(dest.FieldC) != 2 {
		t.Fatalf("field_c: expected length 2, got %d", len(dest.FieldC))
	}

	if dest.FieldC["key1"] == nil {
		t.Fatalf("field_c['key1']: expected non-nil")
	}
	if dest.FieldC["key1"].FieldA == nil || *dest.FieldC["key1"].FieldA != 10 {
		t.Errorf("field_c['key1'].field_a: expected 10, got %v", dest.FieldC["key1"].FieldA)
	}
	if dest.FieldC["key1"].FieldE != "value1" {
		t.Errorf("field_c['key1'].field_e: expected 'value1', got %v", dest.FieldC["key1"].FieldE)
	}
}

func TestAssignAny_NilValues(t *testing.T) {
	err := AssignAny(nil, nil, nil)
	if err != nil {
		t.Errorf("expected nil error for nil inputs, got %v", err)
	}

	desc := &Descriptor{Kind: TypeKind_Struct, Name: "Test"}
	dest := &SampleAssign{}

	err = AssignAny(desc, nil, dest)
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
		Name: "SampleAssign",
		Children: []Field{
			{Name: "field_a", ID: 1},
			// nonexistent is not in descriptor
		},
	}

	dest := &SampleAssign{}
	err := AssignAny(desc, src, dest, WithDisallowNotDefined(true))
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

func BenchmarkAssignAny(b *testing.B) {
	src := map[string]interface{}{
		"field_a":    42,
		"field_e":    "hello",
		"field_list": []interface{}{1, 2, 3, 4, 5},
		"field_map": map[string]interface{}{
			"key1": 100,
			"key2": 200,
		},
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Name: "SampleAssign",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{Name: "field_e", ID: 5},
			{
				Name: "field_list",
				ID:   6,
				Desc: &Descriptor{
					Kind: TypeKind_Scalar,
					Name: "LIST",
				},
			},
			{
				Name: "field_map",
				ID:   7,
				Desc: &Descriptor{
					Kind:     TypeKind_StrMap,
					Name:     "MAP",
					Children: []Field{{Name: "*"}},
				},
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dest := &SampleAssign{}
		_ = AssignAny(desc, src, dest)
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
		Name: "SampleAssignSmall",
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
		dest := &SampleAssignSmall{}
		_ = AssignAny(desc, src, dest)
	}
}
