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

	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/cloudwego/thriftgo/generator/golang/extension/unknown"
)

type sampleFetch struct {
	FieldA         int                     `thrift:"FieldA,1" json:"field_a,omitempty"`
	FieldB         []*sampleFetch          `thrift:"FieldB,2" json:"field_b,omitempty"`
	FieldC         map[string]*sampleFetch `thrift:"FieldC,3" json:"field_c,omitempty"`
	FieldD         *sampleFetch            `thrift:"FieldD,4" json:"field_d,omitempty"`
	FieldE         string                  `thrift:"FieldE,5" json:"field_e,omitempty"`
	FieldList      []int                   `thrift:"FieldList,6" json:"field_list,omitempty"`
	FieldMap       map[string]int          `thrift:"FieldMap,7" json:"field_map,omitempty"`
	_unknownFields unknown.Fields          `json:"-"`
}

func makeSampleFetch(width int, depth int) *sampleFetch {
	if depth <= 0 {
		return nil
	}
	ret := &sampleFetch{
		FieldA:    1,
		FieldE:    "1",
		FieldC:    make(map[string]*sampleFetch),
		FieldList: []int{1, 2, 3},
		FieldMap: map[string]int{
			"1": 1,
			"2": 2,
			"3": 3,
		},
	}
	for i := 0; i < width; i++ {
		ret.FieldB = append(ret.FieldB, makeSampleFetch(width, depth-1))
		ret.FieldC[fmt.Sprintf("%d", i)] = makeSampleFetch(width, depth-1)
	}
	ret.FieldD = makeSampleFetch(width, depth-1)
	return ret
}

// makeDesc generates a descriptor for fetching SampleFetch struct.
// NOTICE: it ignores FieldE.
func makeDesc(width int, depth int, withE bool) *Descriptor {
	if depth <= 0 {
		return nil
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Name: "SampleFetch",
		Children: []Field{
			{
				Name: "field_a",
				ID:   1,
			},
			{
				Name: "field_b",
				ID:   2,
			},
			{
				Name: "field_c",
				ID:   3,
			},
			{
				Name: "field_d",
				ID:   4,
			},
			{
				Name: "field_list",
				ID:   6,
			},
			{
				Name: "field_map",
				ID:   7,
			},
		},
	}

	if withE {
		desc.Children = append(desc.Children, Field{
			Name: "field_e",
			ID:   5,
		})
	}

	nd := makeDesc(width, depth-1, withE)
	desc.Children[2].Desc = &Descriptor{
		Kind: TypeKind_StrMap,
		Name: "MAP",
		Children: []Field{
			{
				Name: "*",
				Desc: nd,
			},
		},
	}
	desc.Children[3].Desc = nd

	return desc
}

// makeSampleAny generates a map[string]interface{} for fetching SampleFetch struct.
// NOTICE: it ignores FieldE and nil field_d at leaf level.
func makeSampleAny(width int, depth int) interface{} {
	if depth <= 0 {
		return nil
	}
	ret := map[string]interface{}{
		"field_a":    int(1),
		"field_b":    []*sampleFetch{},
		"field_c":    map[string]interface{}{},
		"field_list": []int{1, 2, 3},
		"field_map": map[string]int{
			"1": 1,
			"2": 2,
			"3": 3,
		},
	}
	for i := 0; i < width; i++ {
		ret["field_b"] = append(ret["field_b"].([]*sampleFetch), makeSampleFetch(width, depth-1))
		ret["field_c"].(map[string]interface{})[fmt.Sprintf("%d", i)] = makeSampleAny(width, depth-1)
	}
	// Only include field_d if it's not nil (depth > 1 means child will not be nil)
	childD := makeSampleAny(width, depth-1)
	if childD != nil {
		ret["field_d"] = childD
	}
	return ret
}

func TestFetchAny(t *testing.T) {
	width := 2
	depth := 2
	obj := makeSampleFetch(width, depth)
	desc := makeDesc(width, depth, false)
	ret, err := FetchAny(desc, obj)
	if err != nil {
		t.Fatalf("FetchAny failed: %v", err)
	}
	exp := makeSampleAny(width, depth)
	if !reflect.DeepEqual(ret, exp) {
		t.Fatalf("FetchAny failed: %v != %v", ret, exp)
	}
}

func BenchmarkFetchAny(b *testing.B) {
	benchmarks := []struct {
		name  string
		width int
		depth int
	}{
		{"small_2x2", 2, 2},
		{"medium_3x3", 3, 3},
		{"large_4x4", 4, 4},
		{"wide_5x2", 5, 2},
		{"deep_2x5", 2, 5},
	}

	for _, bm := range benchmarks {
		obj := makeSampleFetch(bm.width, bm.depth)
		desc := makeDesc(bm.width, bm.depth, false)

		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = FetchAny(desc, obj)
			}
		})
	}
}

func BenchmarkFetchAny_CacheHit(b *testing.B) {
	// This benchmark specifically tests the performance with cache hit
	// by running FetchAny once before benchmark to warm up the cache
	width := 3
	depth := 3
	obj := makeSampleFetch(width, depth)
	desc := makeDesc(width, depth, false)

	// Warm up the cache
	_, _ = FetchAny(desc, obj)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = FetchAny(desc, obj)
	}
}

// sampleWithUnknown is a struct that has a subset of fields compared to SampleFetch
// It simulates a scenario where some fields are stored in _unknownFields
type sampleWithUnknown struct {
	FieldA int `thrift:"FieldA,1"`
	// FieldB, FieldC, FieldD, FieldE are not declared here, they will be in _unknownFields
	_unknownFields unknown.Fields
}

func TestFetchAnyWithUnknownFields(t *testing.T) {
	// Create a struct with some fields in _unknownFields
	obj := &sampleWithUnknown{
		FieldA: 42,
	}

	// Encode various types into _unknownFields using BinaryProtocol
	p := thrift.BinaryProtocol{}

	// Encode FieldE (id=5, string type)
	p.WriteFieldBegin("", thrift.STRING, 5)
	p.WriteString("hello")

	// Encode custom field (id=6, i32 type)
	p.WriteFieldBegin("", thrift.I32, 6)
	p.WriteI32(100)

	// Encode list<i32> field (id=20)
	p.WriteFieldBegin("", thrift.LIST, 20)
	p.WriteListBegin(thrift.I32, 3)
	p.WriteI32(10)
	p.WriteI32(20)
	p.WriteI32(30)
	p.WriteListEnd()

	// Encode map<string, i32> field (id=21)
	p.WriteFieldBegin("", thrift.MAP, 21)
	p.WriteMapBegin(thrift.STRING, thrift.I32, 2)
	p.WriteString("key1")
	p.WriteI32(100)
	p.WriteString("key2")
	p.WriteI32(200)
	p.WriteMapEnd()

	obj._unknownFields = p.Buf

	// Create a descriptor that asks for all fields
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Name: "SampleWithUnknown",
		Children: []Field{
			{Name: "field_a", ID: 1},      // Static field
			{Name: "field_e", ID: 5},      // From unknownFields (string)
			{Name: "custom_field", ID: 6}, // From unknownFields (i32)
			{Name: "list_field", ID: 20},  // From unknownFields (list<i32>)
			{Name: "map_field", ID: 21},   // From unknownFields (map<string,i32>)
		},
	}

	ret, err := FetchAny(desc, obj)
	if err != nil {
		t.Fatalf("FetchAny failed: %v", err)
	}

	result, ok := ret.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", ret)
	}

	// Check field_a (static field)
	if result["field_a"] != 42 {
		t.Errorf("field_a: expected 42, got %v", result["field_a"])
	}

	// Check field_e (from unknownFields)
	if result["field_e"] != "hello" {
		t.Errorf("field_e: expected 'hello', got %v", result["field_e"])
	}

	// Check custom_field (from unknownFields)
	if result["custom_field"] != int32(100) {
		t.Errorf("custom_field: expected 100, got %v", result["custom_field"])
	}

	// Check list_field
	listVal, ok := result["list_field"].([]interface{})
	if !ok {
		t.Fatalf("list_field: expected []interface{}, got %T", result["list_field"])
	}
	if len(listVal) != 3 {
		t.Errorf("list_field: expected length 3, got %d", len(listVal))
	}
	expectedList := []int32{10, 20, 30}
	for i, v := range listVal {
		if v != expectedList[i] {
			t.Errorf("list_field[%d]: expected %d, got %v", i, expectedList[i], v)
		}
	}

	// Check map_field
	mapVal, ok := result["map_field"].(map[string]interface{})
	if !ok {
		t.Fatalf("map_field: expected map[string]interface{}, got %T", result["map_field"])
	}
	if mapVal["key1"] != int32(100) {
		t.Errorf("map_field['key1']: expected 100, got %v", mapVal["key1"])
	}
	if mapVal["key2"] != int32(200) {
		t.Errorf("map_field['key2']: expected 200, got %v", mapVal["key2"])
	}
}

func TestFetchAnyWithEmptyUnknownFields(t *testing.T) {
	// Create a struct with empty _unknownFields
	obj := &sampleWithUnknown{
		FieldA: 42,
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Name: "SampleWithUnknown",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{Name: "field_e", ID: 5}, // Not present in unknownFields
		},
	}

	ret, err := FetchAny(desc, obj)
	if err != nil {
		t.Fatalf("FetchAny failed: %v", err)
	}

	result, ok := ret.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", ret)
	}

	// Check field_a (static field)
	if result["field_a"] != 42 {
		t.Errorf("field_a: expected 42, got %v", result["field_a"])
	}

	// field_e should not be present since unknownFields is empty
	if _, exists := result["field_e"]; exists {
		t.Errorf("field_e should not exist when unknownFields is empty")
	}
}

func TestFetchAnyWithDisallowNotFound(t *testing.T) {
	t.Run("struct field not found", func(t *testing.T) {
		obj := &sampleWithUnknown{FieldA: 42}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "SampleWithUnknown",
			Children: []Field{
				{Name: "field_a", ID: 1},
				{Name: "field_e", ID: 5}, // Not present
			},
		}

		_, err := FetchAny(desc, obj, FetchOptions{DisallowNotFound: true})
		if err == nil {
			t.Fatalf("expected ErrNotFound, got nil")
		}

		notFoundErr, ok := err.(ErrNotFound)
		if !ok {
			t.Fatalf("expected ErrNotFound, got %T: %v", err, err)
		}
		if notFoundErr.Parent.Name != "SampleWithUnknown" {
			t.Errorf("expected parent name 'SampleWithUnknown', got '%s'", notFoundErr.Parent.Name)
		}
		if notFoundErr.Field.Name != "field_e" {
			t.Errorf("expected field name 'field_e', got '%s'", notFoundErr.Field.Name)
		}
	})

	t.Run("map key not found", func(t *testing.T) {
		obj := &struct {
			Data map[string]int `thrift:"Data,1"`
		}{
			Data: map[string]int{"key1": 100},
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "Test",
			Children: []Field{
				{
					Name: "data",
					ID:   1,
					Desc: &Descriptor{
						Kind: TypeKind_StrMap,
						Name: "MAP",
						Children: []Field{
							{Name: "key1"}, // exists
							{Name: "key2"}, // not exists
						},
					},
				},
			},
		}

		_, err := FetchAny(desc, obj, FetchOptions{DisallowNotFound: true})
		if err == nil {
			t.Fatalf("expected ErrNotFound, got nil")
		}

		notFoundErr, ok := err.(ErrNotFound)
		if !ok {
			t.Fatalf("expected ErrNotFound, got %T: %v", err, err)
		}
		if notFoundErr.Parent.Name != "MAP" {
			t.Errorf("expected parent name 'MAP', got '%s'", notFoundErr.Parent.Name)
		}
		if notFoundErr.Field.Name != "key2" {
			t.Errorf("expected field name 'key2', got '%s'", notFoundErr.Field.Name)
		}
	})

	t.Run("nested struct field not found", func(t *testing.T) {
		obj := &struct {
			Inner *struct {
				Value int `thrift:"Value,1"`
			} `thrift:"Inner,1"`
		}{
			Inner: &struct {
				Value int `thrift:"Value,1"`
			}{Value: 100},
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "Outer",
			Children: []Field{
				{
					Name: "inner",
					ID:   1,
					Desc: &Descriptor{
						Kind: TypeKind_Struct,
						Name: "Inner",
						Children: []Field{
							{Name: "value", ID: 1},    // exists
							{Name: "missing", ID: 99}, // not exists
						},
					},
				},
			},
		}

		_, err := FetchAny(desc, obj, FetchOptions{DisallowNotFound: true})
		if err == nil {
			t.Fatalf("expected ErrNotFound, got nil")
		}

		notFoundErr, ok := err.(ErrNotFound)
		if !ok {
			t.Fatalf("expected ErrNotFound, got %T: %v", err, err)
		}
		if notFoundErr.Parent.Name != "Inner" {
			t.Errorf("expected parent name 'Inner', got '%s'", notFoundErr.Parent.Name)
		}
		if notFoundErr.Field.Name != "missing" {
			t.Errorf("expected field name 'missing', got '%s'", notFoundErr.Field.Name)
		}
	})

	t.Run("no error when all fields found", func(t *testing.T) {
		obj := &sampleWithUnknown{FieldA: 42}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "SampleWithUnknown",
			Children: []Field{
				{Name: "field_a", ID: 1}, // exists
			},
		}

		ret, err := FetchAny(desc, obj, FetchOptions{DisallowNotFound: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result, ok := ret.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map[string]interface{}, got %T", ret)
		}
		if result["field_a"] != 42 {
			t.Errorf("field_a: expected 42, got %v", result["field_a"])
		}
	})
}

// TestFetchAnyWithUnknownFieldsStruct tests fetching struct type from unknownFields
// This covers two cases:
// 1. No further fetch: Descriptor is nil, the struct is returned as-is (map[FieldID]interface{})
// 2. With further fetch: Descriptor is provided, the struct is converted to map[string]interface{}
func TestFetchAnyWithUnknownFieldsStruct(t *testing.T) {
	t.Run("nested struct without descriptor (no further fetch)", func(t *testing.T) {
		// Create a struct with a nested struct in _unknownFields
		obj := &sampleWithUnknown{
			FieldA: 42,
		}

		// Encode a nested struct into _unknownFields (id=10)
		p := thrift.BinaryProtocol{}
		p.WriteFieldBegin("", thrift.STRUCT, 10)
		// Write nested struct fields
		p.WriteFieldBegin("", thrift.STRING, 1) // field 1: string
		p.WriteString("nested_value")
		p.WriteFieldBegin("", thrift.I32, 2) // field 2: i32
		p.WriteI32(999)
		p.WriteFieldStop() // end of nested struct

		obj._unknownFields = p.Buf

		// Create a descriptor that asks for the nested struct but WITHOUT a Desc
		// This means we don't want to further fetch into the struct
		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "SampleWithUnknown",
			Children: []Field{
				{Name: "field_a", ID: 1},
				{Name: "nested_struct", ID: 10}, // No Desc, so no further fetch
			},
		}

		ret, err := FetchAny(desc, obj)
		if err != nil {
			t.Fatalf("FetchAny failed: %v", err)
		}

		result, ok := ret.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map[string]interface{}, got %T", ret)
		}

		// Check field_a
		if result["field_a"] != 42 {
			t.Errorf("field_a: expected 42, got %v", result["field_a"])
		}

		// Check nested_struct: without Desc, it should be map[FieldID]interface{}
		nestedVal, exists := result["nested_struct"]
		if !exists {
			t.Fatalf("nested_struct should exist")
		}

		nestedMap, ok := nestedVal.(map[thrift.FieldID]interface{})
		if !ok {
			t.Fatalf("nested_struct: expected map[thrift.FieldID]interface{}, got %T", nestedVal)
		}

		// Verify the nested struct content
		if nestedMap[1] != "nested_value" {
			t.Errorf("nested_struct[1]: expected 'nested_value', got %v", nestedMap[1])
		}
		if nestedMap[2] != int32(999) {
			t.Errorf("nested_struct[2]: expected 999, got %v", nestedMap[2])
		}
	})

	t.Run("nested struct with descriptor (with further fetch)", func(t *testing.T) {
		// Create a struct with a nested struct in _unknownFields
		obj := &sampleWithUnknown{
			FieldA: 42,
		}

		// Encode a nested struct into _unknownFields (id=10)
		p := thrift.BinaryProtocol{}
		p.WriteFieldBegin("", thrift.STRUCT, 10)
		// Write nested struct fields
		p.WriteFieldBegin("", thrift.STRING, 1) // field 1: string
		p.WriteString("nested_value")
		p.WriteFieldBegin("", thrift.I32, 2) // field 2: i32
		p.WriteI32(999)
		p.WriteFieldBegin("", thrift.STRING, 3) // field 3: string (will be ignored)
		p.WriteString("ignored_value")
		p.WriteFieldStop() // end of nested struct

		obj._unknownFields = p.Buf

		// Create a descriptor that asks for the nested struct WITH a Desc
		// This means we want to further fetch into the struct and convert it
		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "SampleWithUnknown",
			Children: []Field{
				{Name: "field_a", ID: 1},
				{
					Name: "nested_struct",
					ID:   10,
					Desc: &Descriptor{
						Kind: TypeKind_Struct,
						Name: "NestedStruct",
						Children: []Field{
							{Name: "name", ID: 1},  // maps to field 1
							{Name: "count", ID: 2}, // maps to field 2
							// field 3 is not in the descriptor, so it will be ignored
						},
					},
				},
			},
		}

		ret, err := FetchAny(desc, obj)
		if err != nil {
			t.Fatalf("FetchAny failed: %v", err)
		}

		result, ok := ret.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map[string]interface{}, got %T", ret)
		}

		// Check field_a
		if result["field_a"] != 42 {
			t.Errorf("field_a: expected 42, got %v", result["field_a"])
		}

		// Check nested_struct: with Desc, it should be map[string]interface{}
		nestedVal, exists := result["nested_struct"]
		if !exists {
			t.Fatalf("nested_struct should exist")
		}

		nestedMap, ok := nestedVal.(map[string]interface{})
		if !ok {
			t.Fatalf("nested_struct: expected map[string]interface{}, got %T", nestedVal)
		}

		// Verify the nested struct content with field names
		if nestedMap["name"] != "nested_value" {
			t.Errorf("nested_struct['name']: expected 'nested_value', got %v", nestedMap["name"])
		}
		if nestedMap["count"] != int32(999) {
			t.Errorf("nested_struct['count']: expected 999, got %v", nestedMap["count"])
		}

		// field 3 should not be present (not in descriptor)
		if _, exists := nestedMap["ignored"]; exists {
			t.Errorf("nested_struct should not have 'ignored' field")
		}
		// Also verify that only 2 keys are present
		if len(nestedMap) != 2 {
			t.Errorf("nested_struct: expected 2 fields, got %d", len(nestedMap))
		}
	})

	t.Run("deeply nested struct with descriptor", func(t *testing.T) {
		// Create a struct with a deeply nested struct in _unknownFields
		obj := &sampleWithUnknown{
			FieldA: 42,
		}

		// Encode a nested struct with another nested struct inside (id=10)
		p := thrift.BinaryProtocol{}
		p.WriteFieldBegin("", thrift.STRUCT, 10)
		// Write level 1 nested struct fields
		p.WriteFieldBegin("", thrift.STRING, 1) // field 1: string
		p.WriteString("level1")
		p.WriteFieldBegin("", thrift.STRUCT, 2) // field 2: nested struct
		// Write level 2 nested struct fields
		p.WriteFieldBegin("", thrift.STRING, 1) // field 1: string
		p.WriteString("level2")
		p.WriteFieldBegin("", thrift.I64, 2) // field 2: i64
		p.WriteI64(12345)
		p.WriteFieldStop() // end of level 2 struct
		p.WriteFieldStop() // end of level 1 struct

		obj._unknownFields = p.Buf

		// Create a descriptor for deeply nested fetch
		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "SampleWithUnknown",
			Children: []Field{
				{Name: "field_a", ID: 1},
				{
					Name: "level1_struct",
					ID:   10,
					Desc: &Descriptor{
						Kind: TypeKind_Struct,
						Name: "Level1Struct",
						Children: []Field{
							{Name: "level1_name", ID: 1},
							{
								Name: "level2_struct",
								ID:   2,
								Desc: &Descriptor{
									Kind: TypeKind_Struct,
									Name: "Level2Struct",
									Children: []Field{
										{Name: "level2_name", ID: 1},
										{Name: "level2_value", ID: 2},
									},
								},
							},
						},
					},
				},
			},
		}

		ret, err := FetchAny(desc, obj)
		if err != nil {
			t.Fatalf("FetchAny failed: %v", err)
		}

		result, ok := ret.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map[string]interface{}, got %T", ret)
		}

		// Check field_a
		if result["field_a"] != 42 {
			t.Errorf("field_a: expected 42, got %v", result["field_a"])
		}

		// Check level1_struct
		level1, ok := result["level1_struct"].(map[string]interface{})
		if !ok {
			t.Fatalf("level1_struct: expected map[string]interface{}, got %T", result["level1_struct"])
		}

		if level1["level1_name"] != "level1" {
			t.Errorf("level1_struct['level1_name']: expected 'level1', got %v", level1["level1_name"])
		}

		// Check level2_struct
		level2, ok := level1["level2_struct"].(map[string]interface{})
		if !ok {
			t.Fatalf("level2_struct: expected map[string]interface{}, got %T", level1["level2_struct"])
		}

		if level2["level2_name"] != "level2" {
			t.Errorf("level2_struct['level2_name']: expected 'level2', got %v", level2["level2_name"])
		}
		if level2["level2_value"] != int64(12345) {
			t.Errorf("level2_struct['level2_value']: expected 12345, got %v", level2["level2_value"])
		}
	})

	t.Run("struct in map in unknownFields", func(t *testing.T) {
		// Create a struct with a map<string, struct> in _unknownFields
		obj := &sampleWithUnknown{
			FieldA: 42,
		}

		// Encode a map<string, struct> into _unknownFields (id=10)
		p := thrift.BinaryProtocol{}
		p.WriteFieldBegin("", thrift.MAP, 10)
		p.WriteMapBegin(thrift.STRING, thrift.STRUCT, 2)
		// First entry
		p.WriteString("key1")
		p.WriteFieldBegin("", thrift.STRING, 1)
		p.WriteString("value1")
		p.WriteFieldBegin("", thrift.I32, 2)
		p.WriteI32(100)
		p.WriteFieldStop()
		// Second entry
		p.WriteString("key2")
		p.WriteFieldBegin("", thrift.STRING, 1)
		p.WriteString("value2")
		p.WriteFieldBegin("", thrift.I32, 2)
		p.WriteI32(200)
		p.WriteFieldStop()
		p.WriteMapEnd()

		obj._unknownFields = p.Buf

		// Create a descriptor with map containing struct descriptor
		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "SampleWithUnknown",
			Children: []Field{
				{Name: "field_a", ID: 1},
				{
					Name: "data_map",
					ID:   10,
					Desc: &Descriptor{
						Kind: TypeKind_StrMap,
						Name: "DataMap",
						Children: []Field{
							{
								Name: "*",
								Desc: &Descriptor{
									Kind: TypeKind_Struct,
									Name: "Data",
									Children: []Field{
										{Name: "name", ID: 1},
										{Name: "count", ID: 2},
									},
								},
							},
						},
					},
				},
			},
		}

		ret, err := FetchAny(desc, obj)
		if err != nil {
			t.Fatalf("FetchAny failed: %v", err)
		}

		result, ok := ret.(map[string]interface{})
		if !ok {
			t.Fatalf("expected map[string]interface{}, got %T", ret)
		}

		// Check data_map
		dataMap, ok := result["data_map"].(map[string]interface{})
		if !ok {
			t.Fatalf("data_map: expected map[string]interface{}, got %T", result["data_map"])
		}

		if len(dataMap) != 2 {
			t.Fatalf("data_map: expected 2 entries, got %d", len(dataMap))
		}

		// Check key1
		data1, ok := dataMap["key1"].(map[string]interface{})
		if !ok {
			t.Fatalf("data_map['key1']: expected map[string]interface{}, got %T", dataMap["key1"])
		}
		if data1["name"] != "value1" {
			t.Errorf("data_map['key1']['name']: expected 'value1', got %v", data1["name"])
		}
		if data1["count"] != int32(100) {
			t.Errorf("data_map['key1']['count']: expected 100, got %v", data1["count"])
		}

		// Check key2
		data2, ok := dataMap["key2"].(map[string]interface{})
		if !ok {
			t.Fatalf("data_map['key2']: expected map[string]interface{}, got %T", dataMap["key2"])
		}
		if data2["name"] != "value2" {
			t.Errorf("data_map['key2']['name']: expected 'value2', got %v", data2["name"])
		}
		if data2["count"] != int32(200) {
			t.Errorf("data_map['key2']['count']: expected 200, got %v", data2["count"])
		}
	})
}
