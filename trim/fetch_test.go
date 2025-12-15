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
	ret, err := fetchAny(desc, obj)
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
				_, _ = fetchAny(desc, obj)
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
	_, _ = fetchAny(desc, obj)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = fetchAny(desc, obj)
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

	ret, err := fetchAny(desc, obj)
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

	ret, err := fetchAny(desc, obj)
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

		f := Fetcher{FetchOptions: FetchOptions{DisallowNotFound: true}}
		_, err := f.FetchAny(desc, obj)
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

		f := Fetcher{FetchOptions: FetchOptions{DisallowNotFound: true}}
		_, err := f.FetchAny(desc, obj)
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

		f := Fetcher{FetchOptions: FetchOptions{DisallowNotFound: true}}
		_, err := f.FetchAny(desc, obj)
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

		f := Fetcher{FetchOptions: FetchOptions{DisallowNotFound: true}}
		ret, err := f.FetchAny(desc, obj)
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
// ===================== Circular Reference Tests =====================
// These tests verify that FetchAny can handle circular reference type descriptions.
// The key principle is: recursively process data until data is nil (any == nil).

// circularNode represents a node that can reference itself (like a linked list or tree)
type circularNode struct {
	Value int           `thrift:"Value,1" json:"value,omitempty"`
	Next  *circularNode `thrift:"Next,2" json:"next,omitempty"`
}

// circularTree represents a tree node that can reference itself
type circularTree struct {
	Value    int             `thrift:"Value,1" json:"value,omitempty"`
	Left     *circularTree   `thrift:"Left,2" json:"left,omitempty"`
	Right    *circularTree   `thrift:"Right,3" json:"right,omitempty"`
	Children []*circularTree `thrift:"Children,4" json:"children,omitempty"`
}

// makeCircularDesc creates a descriptor that references itself (circular reference)
func makeCircularDesc() *Descriptor {
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

// makeCircularTreeDesc creates a tree descriptor that references itself
func makeCircularTreeDesc() *Descriptor {
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Name: "CircularTree",
		Children: []Field{
			{Name: "value", ID: 1},
			{Name: "left", ID: 2},
			{Name: "right", ID: 3},
			{Name: "children", ID: 4},
		},
	}
	// Make it circular: left, right fields' Desc point back to the same descriptor
	desc.Children[1].Desc = desc
	desc.Children[2].Desc = desc
	// children is a list, no further Desc needed (will be handled as raw value)
	return desc
}

func TestFetchAny_CircularDescriptor_LinkedList(t *testing.T) {
	// Create a linked list: 1 -> 2 -> 3 -> nil
	list := &circularNode{
		Value: 1,
		Next: &circularNode{
			Value: 2,
			Next: &circularNode{
				Value: 3,
				Next:  nil, // Termination point
			},
		},
	}

	desc := makeCircularDesc()

	// Fetch should work correctly, recursing until Next is nil
	fetched, err := fetchAny(desc, list)
	if err != nil {
		t.Fatalf("FetchAny failed: %v", err)
	}

	// Verify the fetched structure
	result, ok := fetched.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", fetched)
	}

	if result["value"] != 1 {
		t.Errorf("value: expected 1, got %v", result["value"])
	}

	next1, ok := result["next"].(map[string]interface{})
	if !ok {
		t.Fatalf("next: expected map[string]interface{}, got %T", result["next"])
	}
	if next1["value"] != 2 {
		t.Errorf("next.value: expected 2, got %v", next1["value"])
	}

	next2, ok := next1["next"].(map[string]interface{})
	if !ok {
		t.Fatalf("next.next: expected map[string]interface{}, got %T", next1["next"])
	}
	if next2["value"] != 3 {
		t.Errorf("next.next.value: expected 3, got %v", next2["value"])
	}

	// The last node's next should not be present (nil)
	if _, exists := next2["next"]; exists {
		t.Errorf("next.next.next should not exist (nil)")
	}
}

func TestFetchAny_CircularDescriptor_SingleNode(t *testing.T) {
	// Single node with nil next
	node := &circularNode{
		Value: 42,
		Next:  nil,
	}

	desc := makeCircularDesc()

	fetched, err := fetchAny(desc, node)
	if err != nil {
		t.Fatalf("FetchAny failed: %v", err)
	}

	result, ok := fetched.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", fetched)
	}

	if result["value"] != 42 {
		t.Errorf("value: expected 42, got %v", result["value"])
	}

	// next should not be present
	if _, exists := result["next"]; exists {
		t.Errorf("next should not exist for nil pointer")
	}
}

func TestFetchAny_CircularDescriptor_Tree(t *testing.T) {
	// Create a binary tree:
	//       1
	//      / \
	//     2   3
	//    /
	//   4
	tree := &circularTree{
		Value: 1,
		Left: &circularTree{
			Value: 2,
			Left: &circularTree{
				Value: 4,
			},
		},
		Right: &circularTree{
			Value: 3,
		},
	}

	desc := makeCircularTreeDesc()

	fetched, err := fetchAny(desc, tree)
	if err != nil {
		t.Fatalf("FetchAny failed: %v", err)
	}

	result, ok := fetched.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", fetched)
	}

	// Verify root
	if result["value"] != 1 {
		t.Errorf("value: expected 1, got %v", result["value"])
	}

	// Verify left subtree
	left, ok := result["left"].(map[string]interface{})
	if !ok {
		t.Fatalf("left: expected map[string]interface{}, got %T", result["left"])
	}
	if left["value"] != 2 {
		t.Errorf("left.value: expected 2, got %v", left["value"])
	}

	leftLeft, ok := left["left"].(map[string]interface{})
	if !ok {
		t.Fatalf("left.left: expected map[string]interface{}, got %T", left["left"])
	}
	if leftLeft["value"] != 4 {
		t.Errorf("left.left.value: expected 4, got %v", leftLeft["value"])
	}

	// Verify right subtree
	right, ok := result["right"].(map[string]interface{})
	if !ok {
		t.Fatalf("right: expected map[string]interface{}, got %T", result["right"])
	}
	if right["value"] != 3 {
		t.Errorf("right.value: expected 3, got %v", right["value"])
	}
}

func TestFetchAny_CircularDescriptor_NilRoot(t *testing.T) {
	desc := makeCircularDesc()

	// Fetch with nil input should return nil
	fetched, err := fetchAny(desc, nil)
	if err != nil {
		t.Fatalf("FetchAny failed: %v", err)
	}

	if fetched != nil {
		t.Errorf("expected nil, got %v", fetched)
	}
}

func TestFetchAny_CircularDescriptor_DeepList(t *testing.T) {
	// Create a deep linked list (depth=100) to stress test
	depth := 100
	var head *circularNode
	for i := depth; i > 0; i-- {
		head = &circularNode{
			Value: i,
			Next:  head,
		}
	}

	desc := makeCircularDesc()

	fetched, err := fetchAny(desc, head)
	if err != nil {
		t.Fatalf("FetchAny failed: %v", err)
	}

	// Verify the structure by traversing
	current := fetched
	for i := 1; i <= depth; i++ {
		m, ok := current.(map[string]interface{})
		if !ok {
			t.Fatalf("at depth %d: expected map[string]interface{}, got %T", i, current)
		}
		if m["value"] != i {
			t.Errorf("at depth %d: expected value %d, got %v", i, i, m["value"])
		}
		if i < depth {
			current = m["next"]
		} else {
			// Last node should not have next
			if _, exists := m["next"]; exists {
				t.Errorf("last node should not have next")
			}
		}
	}
}

// circularMapNode represents a node with a map that can contain circular references
type circularMapNode struct {
	Name     string                      `thrift:"Name,1" json:"name,omitempty"`
	Children map[string]*circularMapNode `thrift:"Children,2" json:"children,omitempty"`
}

func makeCircularMapDesc() *Descriptor {
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

func TestFetchAny_CircularDescriptor_MapOfNodes(t *testing.T) {
	// Create a tree-like structure using maps:
	// root
	// ├── child1
	// │   └── grandchild1
	// └── child2
	node := &circularMapNode{
		Name: "root",
		Children: map[string]*circularMapNode{
			"child1": {
				Name: "child1",
				Children: map[string]*circularMapNode{
					"grandchild1": {
						Name:     "grandchild1",
						Children: nil, // Termination
					},
				},
			},
			"child2": {
				Name:     "child2",
				Children: nil, // Termination
			},
		},
	}

	desc := makeCircularMapDesc()

	fetched, err := fetchAny(desc, node)
	if err != nil {
		t.Fatalf("FetchAny failed: %v", err)
	}

	result, ok := fetched.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", fetched)
	}

	if result["name"] != "root" {
		t.Errorf("name: expected 'root', got %v", result["name"])
	}

	children, ok := result["children"].(map[string]interface{})
	if !ok {
		t.Fatalf("children: expected map[string]interface{}, got %T", result["children"])
	}

	child1, ok := children["child1"].(map[string]interface{})
	if !ok {
		t.Fatalf("child1: expected map[string]interface{}, got %T", children["child1"])
	}
	if child1["name"] != "child1" {
		t.Errorf("child1.name: expected 'child1', got %v", child1["name"])
	}

	child1Children, ok := child1["children"].(map[string]interface{})
	if !ok {
		t.Fatalf("child1.children: expected map[string]interface{}, got %T", child1["children"])
	}

	grandchild1, ok := child1Children["grandchild1"].(map[string]interface{})
	if !ok {
		t.Fatalf("grandchild1: expected map[string]interface{}, got %T", child1Children["grandchild1"])
	}
	if grandchild1["name"] != "grandchild1" {
		t.Errorf("grandchild1.name: expected 'grandchild1', got %v", grandchild1["name"])
	}

	child2, ok := children["child2"].(map[string]interface{})
	if !ok {
		t.Fatalf("child2: expected map[string]interface{}, got %T", children["child2"])
	}
	if child2["name"] != "child2" {
		t.Errorf("child2.name: expected 'child2', got %v", child2["name"])
	}
}

func BenchmarkFetchAny_CircularDescriptor(b *testing.B) {
	// Create a linked list of depth 10
	depth := 10
	var head *circularNode
	for i := depth; i > 0; i-- {
		head = &circularNode{
			Value: i,
			Next:  head,
		}
	}

	desc := makeCircularDesc()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = fetchAny(desc, head)
	}
}

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

		ret, err := fetchAny(desc, obj)
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

		ret, err := fetchAny(desc, obj)
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

		ret, err := fetchAny(desc, obj)
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

		ret, err := fetchAny(desc, obj)
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

// TestFetchAny_PathTracking tests that error messages include the correct DSL path
func TestFetchAny_PathTracking(t *testing.T) {
	tests := []struct {
		name        string
		obj         interface{}
		desc        *Descriptor
		expectedErr string
	}{
		{
			name: "field not found at root",
			obj: &sampleFetch{
				FieldA: 42,
			},
			desc: &Descriptor{
				Kind: TypeKind_Struct,
				Name: "SampleFetch",
				Children: []Field{
					{Name: "field_a", ID: 1},
					{Name: "unknown_field", ID: 99},
				},
			},
			expectedErr: "field ID=99 not found in struct at path $.unknown_field",
		},
		{
			name: "field not found in nested struct",
			obj: &sampleFetch{
				FieldA: 1,
				FieldD: &sampleFetch{
					FieldA: 2,
				},
			},
			desc: &Descriptor{
				Kind: TypeKind_Struct,
				Name: "SampleFetch",
				Children: []Field{
					{Name: "field_a", ID: 1},
					{
						Name: "field_d",
						ID:   4,
						Desc: &Descriptor{
							Kind: TypeKind_Struct,
							Name: "SampleFetch",
							Children: []Field{
								{Name: "field_a", ID: 1},
								{Name: "missing_field", ID: 88},
							},
						},
					},
				},
			},
			expectedErr: "field ID=88 not found in struct at path $.field_d.missing_field",
		},
		{
			name: "field not found in deeply nested struct",
			obj: &sampleFetch{
				FieldA: 1,
				FieldD: &sampleFetch{
					FieldA: 2,
					FieldD: &sampleFetch{
						FieldA: 3,
					},
				},
			},
			desc: &Descriptor{
				Kind: TypeKind_Struct,
				Name: "SampleFetch",
				Children: []Field{
					{Name: "field_a", ID: 1},
					{
						Name: "field_d",
						ID:   4,
						Desc: &Descriptor{
							Kind: TypeKind_Struct,
							Name: "SampleFetch",
							Children: []Field{
								{Name: "field_a", ID: 1},
								{
									Name: "field_d",
									ID:   4,
									Desc: &Descriptor{
										Kind: TypeKind_Struct,
										Name: "SampleFetch",
										Children: []Field{
											{Name: "field_a", ID: 1},
											{Name: "bad_field", ID: 77},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedErr: "field ID=77 not found in struct at path $.field_d.field_d.bad_field",
		},
		{
			name: "map key not found",
			obj: &sampleFetch{
				FieldA: 1,
				FieldC: map[string]*sampleFetch{
					"key1": {FieldA: 10},
				},
			},
			desc: &Descriptor{
				Kind: TypeKind_Struct,
				Name: "SampleFetch",
				Children: []Field{
					{Name: "field_a", ID: 1},
					{
						Name: "field_c",
						ID:   3,
						Desc: &Descriptor{
							Kind: TypeKind_StrMap,
							Name: "MAP",
							Children: []Field{
								{Name: "key1"},
								{Name: "missing_key"},
							},
						},
					},
				},
			},
			expectedErr: "key 'missing_key' not found in map at path $.field_c[missing_key]",
		},
		{
			name: "field not found in map value",
			obj: &sampleFetch{
				FieldA: 1,
				FieldC: map[string]*sampleFetch{
					"item1": {FieldA: 10},
				},
			},
			desc: &Descriptor{
				Kind: TypeKind_Struct,
				Name: "SampleFetch",
				Children: []Field{
					{Name: "field_a", ID: 1},
					{
						Name: "field_c",
						ID:   3,
						Desc: &Descriptor{
							Kind: TypeKind_StrMap,
							Name: "MAP",
							Children: []Field{
								{
									Name: "*",
									Desc: &Descriptor{
										Kind: TypeKind_Struct,
										Name: "SampleFetch",
										Children: []Field{
											{Name: "field_a", ID: 1},
											{Name: "nonexistent", ID: 66},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedErr: "field ID=66 not found in struct at path $.field_c[item1].nonexistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Fetcher{FetchOptions: FetchOptions{DisallowNotFound: true}}
			_, err := f.FetchAny(tt.desc, tt.obj)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !contains(err.Error(), tt.expectedErr) {
				t.Errorf("expected error to contain:\n%s\ngot:\n%s", tt.expectedErr, err.Error())
			}
		})
	}
}

// TestFetchAny_PathTracking_Integration tests path tracking in complex nested scenarios
func TestFetchAny_PathTracking_Integration(t *testing.T) {
	obj := &sampleFetch{
		FieldA: 1,
		FieldC: map[string]*sampleFetch{
			"item1": {
				FieldA: 10,
				FieldD: &sampleFetch{
					FieldA: 20,
				},
			},
		},
	}

	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Name: "SampleFetch",
		Children: []Field{
			{Name: "field_a", ID: 1},
			{
				Name: "field_c",
				ID:   3,
				Desc: &Descriptor{
					Kind: TypeKind_StrMap,
					Name: "MAP",
					Children: []Field{
						{
							Name: "*",
							Desc: &Descriptor{
								Kind: TypeKind_Struct,
								Name: "SampleFetch",
								Children: []Field{
									{Name: "field_a", ID: 1},
									{
										Name: "field_d",
										ID:   4,
										Desc: &Descriptor{
											Kind: TypeKind_Struct,
											Name: "SampleFetch",
											Children: []Field{
												{Name: "field_a", ID: 1},
												{Name: "missing", ID: 55},
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

	f := Fetcher{FetchOptions: FetchOptions{DisallowNotFound: true}}
	_, err := f.FetchAny(desc, obj)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	expectedPath := "$.field_c[item1].field_d.missing"
	if !contains(err.Error(), expectedPath) {
		t.Errorf("expected error to contain path %s, got: %s", expectedPath, err.Error())
	}
}

// BenchmarkFetchAny_PathTracking benchmarks the overhead of path tracking in fetch
func BenchmarkFetchAny_PathTracking(b *testing.B) {
	b.Run("simple_struct", func(b *testing.B) {
		obj := &sampleFetch{
			FieldA: 42,
			FieldE: "hello",
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "SampleFetch",
			Children: []Field{
				{Name: "field_a", ID: 1},
				{Name: "field_e", ID: 5},
			},
		}

		f := Fetcher{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = f.FetchAny(desc, obj)
		}
	})

	b.Run("nested_struct", func(b *testing.B) {
		obj := &sampleFetch{
			FieldA: 1,
			FieldD: &sampleFetch{
				FieldA: 2,
				FieldD: &sampleFetch{
					FieldA: 3,
					FieldE: "nested",
				},
			},
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "SampleFetch",
			Children: []Field{
				{Name: "field_a", ID: 1},
				{
					Name: "field_d",
					ID:   4,
					Desc: &Descriptor{
						Kind: TypeKind_Struct,
						Name: "SampleFetch",
						Children: []Field{
							{Name: "field_a", ID: 1},
							{
								Name: "field_d",
								ID:   4,
								Desc: &Descriptor{
									Kind: TypeKind_Struct,
									Name: "SampleFetch",
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

		f := Fetcher{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = f.FetchAny(desc, obj)
		}
	})

	b.Run("with_map", func(b *testing.B) {
		obj := &sampleFetch{
			FieldA: 1,
			FieldC: map[string]*sampleFetch{
				"key1": {FieldA: 10, FieldE: "v1"},
				"key2": {FieldA: 20, FieldE: "v2"},
				"key3": {FieldA: 30, FieldE: "v3"},
			},
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "SampleFetch",
			Children: []Field{
				{Name: "field_a", ID: 1},
				{
					Name: "field_c",
					ID:   3,
					Desc: &Descriptor{
						Kind: TypeKind_StrMap,
						Name: "MAP",
						Children: []Field{
							{
								Name: "*",
								Desc: &Descriptor{
									Kind: TypeKind_Struct,
									Name: "SampleFetch",
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

		f := Fetcher{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = f.FetchAny(desc, obj)
		}
	})

	b.Run("error_case", func(b *testing.B) {
		obj := &sampleFetch{
			FieldA: 1,
			FieldD: &sampleFetch{
				FieldA: 2,
			},
		}

		desc := &Descriptor{
			Kind: TypeKind_Struct,
			Name: "SampleFetch",
			Children: []Field{
				{Name: "field_a", ID: 1},
				{
					Name: "field_d",
					ID:   4,
					Desc: &Descriptor{
						Kind: TypeKind_Struct,
						Name: "SampleFetch",
						Children: []Field{
							{Name: "field_a", ID: 1},
							{Name: "missing", ID: 99},
						},
					},
				},
			},
		}

		f := Fetcher{FetchOptions: FetchOptions{DisallowNotFound: true}}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = f.FetchAny(desc, obj)
		}
	})
}
