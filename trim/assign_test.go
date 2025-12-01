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
	"testing"

	"github.com/cloudwego/dynamicgo/proto/binary"
)

type sampleAssign struct {
	FieldA           int                      `protobuf:"varint,1,req,name=field_a" json:"field_a,omitempty"`
	FieldB           []*sampleAssign          `protobuf:"bytes,2,opt,name=field_b" json:"field_b,omitempty"`
	FieldC           map[string]*sampleAssign `protobuf:"bytes,3,opt,name=field_c" json:"field_c,omitempty"`
	FieldD           *sampleAssign            `protobuf:"bytes,4,opt,name=field_d" json:"field_d,omitempty"`
	FieldE           string                   `protobuf:"bytes,5,opt,name=field_e" json:"field_e,omitempty"`
	FieldList        []int                    `protobuf:"bytes,6,opt,name=field_list" json:"field_list,omitempty"`
	FieldMap         map[string]int           `protobuf:"bytes,7,opt,name=field_map" json:"field_map,omitempty"`
	XXX_unrecognized []byte                   `json:"-"`
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

	dest := &sampleAssign{}
	err := AssignAny(desc, src, dest)
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

	dest := &sampleAssign{}
	err := AssignAny(desc, src, dest)
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

	dest := &sampleAssign{}
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

	dest := &sampleAssign{}
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

	dest := &sampleAssignSmall{}
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

	dest := &sampleAssign{}
	err := AssignAny(desc, src, dest)
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

	dest := &sampleAssign{}
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
	if dest.FieldC["key1"].FieldA != 10 {
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
	dest := &sampleAssign{}

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

	dest := &sampleAssign{}
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
		dest := &sampleAssign{}
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
		dest := &sampleAssignSmall{}
		_ = AssignAny(desc, src, dest)
	}
}

// SourceStruct is used for struct-to-struct assignment tests via json tag matching
type SourceStruct struct {
	Name   string  `json:"name"`
	Age    int     `json:"age"`
	Score  float64 `json:"score"`
	Active bool    `json:"active"`
}

// DestStruct has same json tags but different Go field names
type DestStruct struct {
	UserName  string  `json:"name"`
	UserAge   int     `json:"age"`
	UserScore float64 `json:"score"`
	IsActive  bool    `json:"active"`
}

// NestedSourceStruct contains nested struct
type NestedSourceStruct struct {
	ID   int           `json:"id"`
	Data *SourceStruct `json:"data"`
}

// NestedDestStruct contains nested struct with different types
type NestedDestStruct struct {
	ID   int         `json:"id"`
	Data *DestStruct `json:"data"`
}

// ListSourceStruct contains a list of structs
type ListSourceStruct struct {
	Items []*SourceStruct `json:"items"`
}

// ListDestStruct contains a list of different struct types
type ListDestStruct struct {
	Items []*DestStruct `json:"items"`
}

// MapSourceStruct contains a map of structs
type MapSourceStruct struct {
	Data map[string]*SourceStruct `json:"data"`
}

// MapDestStruct contains a map of different struct types
type MapDestStruct struct {
	Data map[string]*DestStruct `json:"data"`
}

func TestAssignScalar_StructToStruct(t *testing.T) {
	t.Run("basic struct to struct via json tag", func(t *testing.T) {
		src := &SourceStruct{
			Name:   "Alice",
			Age:    30,
			Score:  95.5,
			Active: true,
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "Wrapper",
			Children: []Field{
				{Name: "data", ID: 1},
			},
		}

		type Wrapper struct {
			Data *DestStruct `protobuf:"bytes,1,opt,name=data" json:"data"`
		}

		srcMap := map[string]interface{}{
			"data": src,
		}

		dest := &Wrapper{}
		err := AssignAny(desc, srcMap, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		if dest.Data == nil {
			t.Fatalf("dest.Data should not be nil")
		}
		if dest.Data.UserName != "Alice" {
			t.Errorf("UserName: expected 'Alice', got '%s'", dest.Data.UserName)
		}
		if dest.Data.UserAge != 30 {
			t.Errorf("UserAge: expected 30, got %d", dest.Data.UserAge)
		}
		if dest.Data.UserScore != 95.5 {
			t.Errorf("UserScore: expected 95.5, got %f", dest.Data.UserScore)
		}
		if dest.Data.IsActive != true {
			t.Errorf("IsActive: expected true, got %v", dest.Data.IsActive)
		}
	})

	t.Run("nested struct to struct via json tag", func(t *testing.T) {
		src := &NestedSourceStruct{
			ID: 100,
			Data: &SourceStruct{
				Name:   "Bob",
				Age:    25,
				Score:  88.0,
				Active: false,
			},
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "Wrapper",
			Children: []Field{
				{Name: "nested", ID: 1},
			},
		}

		type Wrapper struct {
			Nested *NestedDestStruct `protobuf:"bytes,1,opt,name=nested" json:"nested"`
		}

		srcMap := map[string]interface{}{
			"nested": src,
		}

		dest := &Wrapper{}
		err := AssignAny(desc, srcMap, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		if dest.Nested == nil {
			t.Fatalf("dest.Nested should not be nil")
		}
		if dest.Nested.ID != 100 {
			t.Errorf("ID: expected 100, got %d", dest.Nested.ID)
		}
		if dest.Nested.Data == nil {
			t.Fatalf("dest.Nested.Data should not be nil")
		}
		if dest.Nested.Data.UserName != "Bob" {
			t.Errorf("UserName: expected 'Bob', got '%s'", dest.Nested.Data.UserName)
		}
		if dest.Nested.Data.UserAge != 25 {
			t.Errorf("UserAge: expected 25, got %d", dest.Nested.Data.UserAge)
		}
	})
}

func TestAssignScalar_SliceToSlice(t *testing.T) {
	t.Run("slice of structs via json tag", func(t *testing.T) {
		src := &ListSourceStruct{
			Items: []*SourceStruct{
				{Name: "Alice", Age: 30, Score: 95.5, Active: true},
				{Name: "Bob", Age: 25, Score: 88.0, Active: false},
			},
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "Wrapper",
			Children: []Field{
				{Name: "list", ID: 1},
			},
		}

		type Wrapper struct {
			List *ListDestStruct `protobuf:"bytes,1,opt,name=list" json:"list"`
		}

		srcMap := map[string]interface{}{
			"list": src,
		}

		dest := &Wrapper{}
		err := AssignAny(desc, srcMap, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		if dest.List == nil {
			t.Fatalf("dest.List should not be nil")
		}
		if len(dest.List.Items) != 2 {
			t.Fatalf("Items length: expected 2, got %d", len(dest.List.Items))
		}
		if dest.List.Items[0].UserName != "Alice" {
			t.Errorf("Items[0].UserName: expected 'Alice', got '%s'", dest.List.Items[0].UserName)
		}
		if dest.List.Items[1].UserName != "Bob" {
			t.Errorf("Items[1].UserName: expected 'Bob', got '%s'", dest.List.Items[1].UserName)
		}
	})

	t.Run("slice of primitives", func(t *testing.T) {
		src := []int32{1, 2, 3, 4, 5}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "Wrapper",
			Children: []Field{
				{Name: "nums", ID: 1},
			},
		}

		type Wrapper struct {
			Nums []int64 `protobuf:"bytes,1,opt,name=nums" json:"nums"`
		}

		srcMap := map[string]interface{}{
			"nums": src,
		}

		dest := &Wrapper{}
		err := AssignAny(desc, srcMap, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		expected := []int64{1, 2, 3, 4, 5}
		if !reflect.DeepEqual(dest.Nums, expected) {
			t.Errorf("Nums: expected %v, got %v", expected, dest.Nums)
		}
	})
}

func TestAssignScalar_MapToMap(t *testing.T) {
	t.Run("map of structs via json tag", func(t *testing.T) {
		src := &MapSourceStruct{
			Data: map[string]*SourceStruct{
				"user1": {Name: "Alice", Age: 30, Score: 95.5, Active: true},
				"user2": {Name: "Bob", Age: 25, Score: 88.0, Active: false},
			},
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "Wrapper",
			Children: []Field{
				{Name: "map_data", ID: 1},
			},
		}

		type Wrapper struct {
			MapData *MapDestStruct `protobuf:"bytes,1,opt,name=map_data" json:"map_data"`
		}

		srcMap := map[string]interface{}{
			"map_data": src,
		}

		dest := &Wrapper{}
		err := AssignAny(desc, srcMap, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		if dest.MapData == nil {
			t.Fatalf("dest.MapData should not be nil")
		}
		if len(dest.MapData.Data) != 2 {
			t.Fatalf("Data length: expected 2, got %d", len(dest.MapData.Data))
		}
		if dest.MapData.Data["user1"].UserName != "Alice" {
			t.Errorf("Data['user1'].UserName: expected 'Alice', got '%s'", dest.MapData.Data["user1"].UserName)
		}
		if dest.MapData.Data["user2"].UserName != "Bob" {
			t.Errorf("Data['user2'].UserName: expected 'Bob', got '%s'", dest.MapData.Data["user2"].UserName)
		}
	})

	t.Run("map of primitives with type conversion", func(t *testing.T) {
		src := map[string]int32{"a": 1, "b": 2, "c": 3}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "Wrapper",
			Children: []Field{
				{Name: "data", ID: 1},
			},
		}

		type Wrapper struct {
			Data map[string]int64 `protobuf:"bytes,1,opt,name=data" json:"data"`
		}

		srcMap := map[string]interface{}{
			"data": src,
		}

		dest := &Wrapper{}
		err := AssignAny(desc, srcMap, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		expected := map[string]int64{"a": 1, "b": 2, "c": 3}
		if !reflect.DeepEqual(dest.Data, expected) {
			t.Errorf("Data: expected %v, got %v", expected, dest.Data)
		}
	})
}

func TestAssignScalar_ComplexNested(t *testing.T) {
	// Test complex nested structure similar to TestFetchAndAssign scenario
	t.Run("complex nested with lists and maps", func(t *testing.T) {
		type InnerSource struct {
			Value int    `json:"value"`
			Label string `json:"label"`
		}

		type OuterSource struct {
			ID       int                     `json:"id"`
			Children []*InnerSource          `json:"children"`
			Mapping  map[string]*InnerSource `json:"mapping"`
		}

		type InnerDest struct {
			Value int    `json:"value"`
			Label string `json:"label"`
		}

		type OuterDest struct {
			ID       int                   `json:"id"`
			Children []*InnerDest          `json:"children"`
			Mapping  map[string]*InnerDest `json:"mapping"`
		}

		src := &OuterSource{
			ID: 1,
			Children: []*InnerSource{
				{Value: 10, Label: "first"},
				{Value: 20, Label: "second"},
			},
			Mapping: map[string]*InnerSource{
				"key1": {Value: 100, Label: "mapped1"},
				"key2": {Value: 200, Label: "mapped2"},
			},
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "Wrapper",
			Children: []Field{
				{Name: "data", ID: 1},
			},
		}

		type Wrapper struct {
			Data *OuterDest `protobuf:"bytes,1,opt,name=data" json:"data"`
		}

		srcMap := map[string]interface{}{
			"data": src,
		}

		dest := &Wrapper{}
		err := AssignAny(desc, srcMap, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		if dest.Data == nil {
			t.Fatalf("dest.Data should not be nil")
		}
		if dest.Data.ID != 1 {
			t.Errorf("ID: expected 1, got %d", dest.Data.ID)
		}
		if len(dest.Data.Children) != 2 {
			t.Fatalf("Children length: expected 2, got %d", len(dest.Data.Children))
		}
		if dest.Data.Children[0].Value != 10 {
			t.Errorf("Children[0].Value: expected 10, got %d", dest.Data.Children[0].Value)
		}
		if dest.Data.Children[0].Label != "first" {
			t.Errorf("Children[0].Label: expected 'first', got '%s'", dest.Data.Children[0].Label)
		}
		if len(dest.Data.Mapping) != 2 {
			t.Fatalf("Mapping length: expected 2, got %d", len(dest.Data.Mapping))
		}
		if dest.Data.Mapping["key1"].Value != 100 {
			t.Errorf("Mapping['key1'].Value: expected 100, got %d", dest.Data.Mapping["key1"].Value)
		}
	})
}

func TestAssignScalar_NilHandling(t *testing.T) {
	t.Run("nil pointer in source struct", func(t *testing.T) {
		src := &NestedSourceStruct{
			ID:   100,
			Data: nil, // nil pointer
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "Wrapper",
			Children: []Field{
				{Name: "nested", ID: 1},
			},
		}

		type Wrapper struct {
			Nested *NestedDestStruct `protobuf:"bytes,1,opt,name=nested" json:"nested"`
		}

		srcMap := map[string]interface{}{
			"nested": src,
		}

		dest := &Wrapper{}
		err := AssignAny(desc, srcMap, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		if dest.Nested == nil {
			t.Fatalf("dest.Nested should not be nil")
		}
		if dest.Nested.ID != 100 {
			t.Errorf("ID: expected 100, got %d", dest.Nested.ID)
		}
		if dest.Nested.Data != nil {
			t.Errorf("Data: expected nil, got %v", dest.Nested.Data)
		}
	})

	t.Run("nil slice in source struct", func(t *testing.T) {
		src := &ListSourceStruct{
			Items: nil,
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "Wrapper",
			Children: []Field{
				{Name: "list", ID: 1},
			},
		}

		type Wrapper struct {
			List *ListDestStruct `protobuf:"bytes,1,opt,name=list" json:"list"`
		}

		srcMap := map[string]interface{}{
			"list": src,
		}

		dest := &Wrapper{}
		err := AssignAny(desc, srcMap, dest)
		if err != nil {
			t.Fatalf("AssignAny failed: %v", err)
		}

		if dest.List == nil {
			t.Fatalf("dest.List should not be nil")
		}
		if dest.List.Items != nil {
			t.Errorf("Items: expected nil, got %v", dest.List.Items)
		}
	})
}

func BenchmarkAssignScalar_StructToStruct(b *testing.B) {
	src := &SourceStruct{
		Name:   "Alice",
		Age:    30,
		Score:  95.5,
		Active: true,
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Name: "Wrapper",
		Children: []Field{
			{Name: "data", ID: 1},
		},
	}

	type Wrapper struct {
		Data *DestStruct `protobuf:"bytes,1,opt,name=data" json:"data"`
	}

	srcMap := map[string]interface{}{
		"data": src,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dest := &Wrapper{}
		_ = AssignAny(desc, srcMap, dest)
	}
}

func BenchmarkAssignScalar_SliceOfStructs(b *testing.B) {
	src := &ListSourceStruct{
		Items: []*SourceStruct{
			{Name: "Alice", Age: 30, Score: 95.5, Active: true},
			{Name: "Bob", Age: 25, Score: 88.0, Active: false},
			{Name: "Charlie", Age: 35, Score: 92.0, Active: true},
		},
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Name: "Wrapper",
		Children: []Field{
			{Name: "list", ID: 1},
		},
	}

	type Wrapper struct {
		List *ListDestStruct `protobuf:"bytes,1,opt,name=list" json:"list"`
	}

	srcMap := map[string]interface{}{
		"list": src,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dest := &Wrapper{}
		_ = AssignAny(desc, srcMap, dest)
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
		Name: "CircularNode",
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
		Name: "CircularTree",
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
	err := AssignAny(desc, src, dest)
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
	err := AssignAny(desc, src, dest)
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
	err := AssignAny(desc, src, dest)
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
	err := AssignAny(desc, nil, dest)
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
	err := AssignAny(desc, src, dest)
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
		Name: "CircularMapNode",
		Children: []Field{
			{Name: "name", ID: 1},
			{Name: "children", ID: 2},
		},
	}
	// Make children field circular: it's a map with values of the same type
	desc.Children[1].Desc = &Descriptor{
		Kind: TypeKind_StrMap,
		Name: "ChildrenMap",
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
	err := AssignAny(desc, src, dest)
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

func BenchmarkAssignAny_CircularDescriptor(b *testing.B) {
	// Create a linked list of depth 10 as source
	depth := 10
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

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dest := &circularAssignNode{}
		_ = AssignAny(desc, src, dest)
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
		Name: "CircularNode",
		Children: []Field{
			{Name: "value", ID: 1},
			{Name: "next", ID: 2},
		},
	}
	fetchDesc.Children[1].Desc = fetchDesc

	// Fetch
	fetched, err := FetchAny(fetchDesc, srcList)
	if err != nil {
		t.Fatalf("FetchAny failed: %v", err)
	}

	// Create circular descriptor for assign (with different IDs if needed)
	assignDesc := &Descriptor{
		Kind: TypeKind_Struct,
		Name: "CircularNode",
		Children: []Field{
			{Name: "value", ID: 1},
			{Name: "next", ID: 2},
		},
	}
	assignDesc.Children[1].Desc = assignDesc

	// Assign
	dest := &circularAssignNode{}
	err = AssignAny(assignDesc, fetched, dest)
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
