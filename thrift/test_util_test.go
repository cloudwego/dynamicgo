package thrift

import (
	"testing"
)

func TestMakeFuzzData_MapInt64Key(t *testing.T) {
	// Build a MAP descriptor: key=I64, elem=STRING
	keyDesc := &TypeDescriptor{typ: I64}
	elemDesc := &TypeDescriptor{typ: STRING}
	desc := &TypeDescriptor{typ: MAP, key: keyDesc, elem: elemDesc}

	opts := FuzzDataOptions{MaxDepth: 2, MaxWidth: 2}

	data, err := MakeFuzzData(desc, opts)
	if err != nil {
		t.Fatalf("MakeFuzzData error: %v", err)
	}

	// Expect map[int64]interface{}
	if _, ok := data.(map[int64]interface{}); !ok {
		t.Fatalf("expected map[int64]interface{}, got %T", data)
	}
}

// TestMakeFuzzData_ComplexIDL builds a descriptor from an IDL string containing
// struct/list/map/int/string/float/bool, then validates the generated fuzz data types.
func TestMakeFuzzData_ComplexIDL(t *testing.T) {
	content := `
namespace go test

struct Inner {
  1: required i32 a
}

struct Complex {
  1: required bool b
  2: required i32 i
  3: required double f
  4: required string s
  5: required list<i64> l
  6: required map<i32, string> m
  7: required Inner inner
}

service S {
  void ExampleMethod(1: Complex req)
}
`

	fn, err := GetDescFromContent(content, "ExampleMethod", &Options{})
	if err != nil {
		t.Fatalf("failed to build descriptor from content: %v", err)
	}

	reqDesc := FnRequest(fn)
	data, err := MakeFuzzData(reqDesc, FuzzDataOptions{MaxDepth: 5, MaxWidth: 3, DefaultFieldsRatio: 1.0, OptionalFieldsRatio: 1.0})
	if err != nil {
		t.Fatalf("MakeFuzzData error: %v", err)
	}

	root, ok := data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected root map[string]interface{}, got %T", data)
	}

	// Validate presence and types of each required field
	if _, ok := root["b"].(bool); !ok {
		t.Fatalf("field b: expected bool, got %T", root["b"])
	}
	if _, ok := root["i"].(int32); !ok {
		t.Fatalf("field i: expected int32, got %T", root["i"])
	}
	if _, ok := root["f"].(float64); !ok {
		t.Fatalf("field f: expected float64, got %T", root["f"])
	}
	if _, ok := root["s"].(string); !ok {
		t.Fatalf("field s: expected string, got %T", root["s"])
	}
	if _, ok := root["l"].([]interface{}); !ok {
		t.Fatalf("field l: expected []interface{}, got %T", root["l"])
	}
	if _, ok := root["m"].(map[int32]interface{}); !ok {
		t.Fatalf("field m: expected map[int32]interface{}, got %T", root["m"])
	}
	if _, ok := root["inner"].(map[string]interface{}); !ok {
		t.Fatalf("field inner: expected map[string]interface{}, got %T", root["inner"])
	}
}
