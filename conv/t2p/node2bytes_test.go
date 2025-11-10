package t2p

import (
	"testing"

	"github.com/cloudwego/dynamicgo/thrift"
	tgeneric "github.com/cloudwego/dynamicgo/thrift/generic"
	"github.com/stretchr/testify/require"
)

func TestPathNodeToBytesConv_Do(t *testing.T) {
	// Create a simple thrift PathNode for testing
	conv := &PathNodeToBytesConv{}

	t.Run("primitive bool", func(t *testing.T) {
		// Create a bool node
		boolNode := tgeneric.NewNodeBool(true)
		pathNode := &tgeneric.PathNode{
			Node: boolNode,
		}

		var out []byte
		err := conv.Do(pathNode, &out)
		require.NoError(t, err)
		require.NotEmpty(t, out)
		// Protobuf bool true should be encoded as varint 1
		require.Equal(t, []byte{0x01}, out)
	})

	t.Run("primitive int32", func(t *testing.T) {
		// Create an int32 node
		intNode := tgeneric.NewNodeInt32(42)
		pathNode := &tgeneric.PathNode{
			Node: intNode,
		}

		var out []byte
		err := conv.Do(pathNode, &out)
		require.NoError(t, err)
		require.NotEmpty(t, out)
		// Protobuf int32 42 should be encoded as varint 42
		require.Equal(t, []byte{0x2A}, out) // 42 in varint
	})

	t.Run("primitive string", func(t *testing.T) {
		// Create a string node
		strNode := tgeneric.NewNodeString("hello")
		pathNode := &tgeneric.PathNode{
			Node: strNode,
		}

		var out []byte
		err := conv.Do(pathNode, &out)
		require.NoError(t, err)
		require.NotEmpty(t, out)
		// Protobuf string should be length-prefixed
		require.Equal(t, []byte{0x05, 'h', 'e', 'l', 'l', 'o'}, out)
	})

	t.Run("struct with fields", func(t *testing.T) {
		// Create a struct-like PathNode with fields
		opts := &tgeneric.Options{}
		structNode := tgeneric.NewNodeStruct(map[thrift.FieldID]interface{}{
			1: true,
			2: int32(123),
		}, opts)
		pathNode := &tgeneric.PathNode{
			Node: structNode,
		}

		var out []byte
		err := conv.Do(pathNode, &out)
		require.NoError(t, err)
		require.NotEmpty(t, out)
		// Should contain protobuf encoded fields
		require.Greater(t, len(out), 2)
	})

	t.Run("list", func(t *testing.T) {
		// Create a list node
		listNode := tgeneric.NewNodeList([]interface{}{int32(1), int32(2), int32(3)})
		pathNode := &tgeneric.PathNode{
			Node: listNode,
		}

		var out []byte
		err := conv.Do(pathNode, &out)
		require.NoError(t, err)
		require.NotEmpty(t, out)
	})

	t.Run("map", func(t *testing.T) {
		// Create a map node
		opts := &tgeneric.Options{}
		mapNode := tgeneric.NewNodeMap(map[interface{}]interface{}{
			"key1": "value1",
			"key2": "value2",
		}, opts)
		pathNode := &tgeneric.PathNode{
			Node: mapNode,
		}

		var out []byte
		err := conv.Do(pathNode, &out)
		require.NoError(t, err)
		require.NotEmpty(t, out)
	})

	t.Run("nil input", func(t *testing.T) {
		var out []byte
		err := conv.Do(nil, &out)
		require.Error(t, err)
		require.Contains(t, err.Error(), "input PathNode is nil")
	})
}

func TestPathNodeToBytesConv_ComplexStruct(t *testing.T) {
	conv := &PathNodeToBytesConv{}

	// Create a more complex struct with nested data
	// This simulates a thrift message with multiple fields
	
	// Field 1: bool
	boolField := tgeneric.PathNode{
		Path: tgeneric.NewPathFieldId(1),
		Node: tgeneric.NewNodeBool(true),
	}
	
	// Field 2: int32  
	intField := tgeneric.PathNode{
		Path: tgeneric.NewPathFieldId(2),
		Node: tgeneric.NewNodeInt32(42),
	}
	
	// Field 3: string
	strField := tgeneric.PathNode{
		Path: tgeneric.NewPathFieldId(3),
		Node: tgeneric.NewNodeString("test"),
	}

	// Create the root struct node
	// Use NewNodeStruct to create a proper struct node
	structData := map[thrift.FieldID]interface{}{
		1: true,
		2: int32(42),
		3: "test",
	}
	opts := &tgeneric.Options{}
	structNode := tgeneric.NewNodeStruct(structData, opts)
	
	rootNode := &tgeneric.PathNode{
		Node: structNode,
		Next: []tgeneric.PathNode{boolField, intField, strField},
	}

	var out []byte
	err := conv.Do(rootNode, &out)
	require.NoError(t, err)
	require.NotEmpty(t, out)
	
	// Verify that we get some protobuf-encoded data
	// The exact format will depend on the field numbers and types
	require.Greater(t, len(out), 3)
}

func TestPathNodeToBytesConv_EmptyNode(t *testing.T) {
	conv := &PathNodeToBytesConv{}

	// Test with empty struct
	emptyStructNode := tgeneric.NewNodeStruct(map[thrift.FieldID]interface{}{}, &tgeneric.Options{})
	emptyStruct := &tgeneric.PathNode{
		Node: emptyStructNode,
		Next: []tgeneric.PathNode{},
	}

	var out []byte
	err := conv.Do(emptyStruct, &out)
	require.NoError(t, err)
	require.Empty(t, out) // Empty struct should produce empty output
}