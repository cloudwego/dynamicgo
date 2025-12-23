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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDescriptorMarshalJSON(t *testing.T) {
	// Create a simple descriptor without circular reference
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "Root",
		Children: []Field{
			{
				Name: "field1",
				ID:   1,
				Desc: &Descriptor{
					Kind: TypeKind_Leaf,
					Type: "Leaf1",
				},
			},
			{
				Name: "field2",
				ID:   2,
				Desc: &Descriptor{
					Kind: TypeKind_StrMap,
					Type: "Map1",
					Children: []Field{
						{
							Name: "key1",
							ID:   0,
							Desc: &Descriptor{
								Kind: TypeKind_Leaf,
								Type: "Leaf2",
							},
						},
					},
				},
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(desc)
	require.NoError(t, err)
	t.Logf("JSON: %s", string(data))

	// Unmarshal back
	var desc2 Descriptor
	err = json.Unmarshal(data, &desc2)
	require.NoError(t, err)

	// Verify structure
	require.Equal(t, desc.Kind, desc2.Kind)
	require.Equal(t, desc.Type, desc2.Type)
	require.Len(t, desc2.Children, 2)
	require.Equal(t, desc.Children[0].Name, desc2.Children[0].Name)
	require.Equal(t, desc.Children[0].ID, desc2.Children[0].ID)
	require.NotNil(t, desc2.Children[0].Desc)
	require.Equal(t, desc.Children[0].Desc.Type, desc2.Children[0].Desc.Type)
}

func TestDescriptorMarshalJSONWithCircularReference(t *testing.T) {
	// Create a descriptor with circular reference
	root := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "Root",
	}

	child := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "Child",
	}

	// Root -> Child -> Root (circular reference)
	root.Children = []Field{
		{
			Name: "child",
			ID:   1,
			Desc: child,
		},
	}

	child.Children = []Field{
		{
			Name: "parent",
			ID:   1,
			Desc: root, // circular reference back to root
		},
	}

	// Marshal to JSON - should not panic or infinite loop
	data, err := json.Marshal(root)
	require.NoError(t, err)
	t.Logf("JSON with circular ref: %s", string(data))

	// Unmarshal back
	var root2 Descriptor
	err = json.Unmarshal(data, &root2)
	require.NoError(t, err)

	// Verify structure
	require.Equal(t, root.Kind, root2.Kind)
	require.Equal(t, root.Type, root2.Type)
	require.Len(t, root2.Children, 1)
	require.NotNil(t, root2.Children[0].Desc)
	require.Equal(t, "Child", root2.Children[0].Desc.Type)

	// Verify circular reference is resolved
	require.NotNil(t, root2.Children[0].Desc.Children)
	require.Len(t, root2.Children[0].Desc.Children, 1)
	require.NotNil(t, root2.Children[0].Desc.Children[0].Desc)
	// The circular reference should point back to root2
	require.Equal(t, &root2, root2.Children[0].Desc.Children[0].Desc)
}

func TestDescriptorMarshalJSONSelfReference(t *testing.T) {
	// Create a descriptor that references itself
	self := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "Self",
	}

	self.Children = []Field{
		{
			Name: "self",
			ID:   1,
			Desc: self, // self reference
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(self)
	require.NoError(t, err)
	t.Logf("JSON with self ref: %s", string(data))

	// Unmarshal back
	var self2 Descriptor
	err = json.Unmarshal(data, &self2)
	require.NoError(t, err)

	// Verify self reference is resolved
	require.Equal(t, &self2, self2.Children[0].Desc)
}

func TestDescriptorMarshalJSONNil(t *testing.T) {
	// Descriptor with nil child
	desc := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "Root",
		Children: []Field{
			{
				Name: "field1",
				ID:   1,
				Desc: nil, // nil descriptor
			},
		},
	}

	data, err := json.Marshal(desc)
	require.NoError(t, err)
	t.Logf("JSON with nil: %s", string(data))

	var desc2 Descriptor
	err = json.Unmarshal(data, &desc2)
	require.NoError(t, err)

	require.Nil(t, desc2.Children[0].Desc)
}

func TestDescriptorMarshalJSONMultipleReferences(t *testing.T) {
	// Create a shared descriptor referenced by multiple fields
	shared := &Descriptor{
		Kind: TypeKind_Leaf,
		Type: "Shared",
	}

	root := &Descriptor{
		Kind: TypeKind_Struct,
		Type: "Root",
		Children: []Field{
			{
				Name: "ref1",
				ID:   1,
				Desc: shared,
			},
			{
				Name: "ref2",
				ID:   2,
				Desc: shared, // same descriptor referenced again
			},
		},
	}

	data, err := json.Marshal(root)
	require.NoError(t, err)
	t.Logf("JSON with multiple refs: %s", string(data))

	var root2 Descriptor
	err = json.Unmarshal(data, &root2)
	require.NoError(t, err)

	// Both refs should point to the same descriptor after unmarshaling
	require.NotNil(t, root2.Children[0].Desc)
	require.NotNil(t, root2.Children[1].Desc)
	require.Equal(t, root2.Children[0].Desc, root2.Children[1].Desc)
}
