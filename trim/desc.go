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
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// TypeKind is the kind of type.
type TypeKind int

const (
	// TypeKind_Leaf indicates Descriptor is a leaf node, its underlying type can be anything (event go struct/map/list)
	TypeKind_Leaf TypeKind = iota
	// TypeKind_Struct indicates Descriptor.Field is struct field
	TypeKind_Struct
	// TypeKind_StrMap indicates Descriptor.Field is map key
	TypeKind_StrMap
	// TypeKind_List indicates Descriptor.Field is list index
	TypeKind_List
)

// Descriptor describes the entire a DSL-pruning scheme for a type.
// base on this, we can fetch the type's object data on demands
type Descriptor struct {
	// the kind of corresponding type
	// Based on this, we can decide how to manipulate the data (e.g. mapKey or strucField)
	Kind TypeKind

	// Type of the type
	Type string

	// children for TypeKind_Struct|TypeKind_StrMap|TypeKind_List
	// - For TypeKind_StrMap, either each Field is a key-value pair or one field with Name "*"
	// - For TypeKind_Struct, each Field is a field with both Name and ID
	// - For TypeKind_List, either each Field.ID is the element index or one field with Name "*" (means all elements)
	Children []Field

	// for speed-up search
	normalized int32 // atomic flag: 0=not started, 1=in progress/done
	ids        map[int]Field
	names      map[string]Field
}

// Field represents a mapping selection
type Field struct {
	// Name of the field path for TypeKind_Struct
	// Or the selection key for TypeKind_StrMap
	Name string

	// FieldID in IDL
	ID int

	// the child of the field
	Desc *Descriptor
}

// Normalize cache all the fields in the descriptor for speeding up search.
// It handles circular references by using an atomic flag to detect re-entry.
func (d *Descriptor) Normalize() {
	// Use atomic to detect if we're already normalizing this descriptor
	// This prevents infinite recursion in circular references
	if !atomic.CompareAndSwapInt32(&d.normalized, 0, 1) {
		// Already being normalized or already done
		return
	}

	d.ids = make(map[int]Field, len(d.Children))
	d.names = make(map[string]Field, len(d.Children))
	for _, f := range d.Children {
		d.ids[f.ID] = f
		d.names[f.Name] = f
		if f.Desc != nil {
			f.Desc.Normalize()
		}
	}
}

// String returns the string representation of the descriptor in JSON-like format.
// It handles circular references by tracking visited descriptors.
// Format: <TypeName>{...} for struct/map, "-" for scalar types.
func (d *Descriptor) String() string {
	sb := strings.Builder{}
	visited := make(map[*Descriptor]bool)

	var printer func(desc *Descriptor, indent string)
	printer = func(desc *Descriptor, indent string) {
		// Handle circular references
		if visited[desc] {
			sb.WriteString("<" + desc.Type + ">")
			return
		}
		visited[desc] = true

		// Get type prefix based on Kind
		var typePrefix string
		switch desc.Kind {
		case TypeKind_Leaf:
			sb.WriteString("-")
			return
		case TypeKind_StrMap:
			typePrefix = "<MAP>"
		default: // TypeKind_Struct
			typePrefix = "<" + desc.Type + ">"
		}

		sb.WriteString(typePrefix)

		if len(desc.Children) == 0 {
			sb.WriteString("{}")
			return
		}

		sb.WriteString("{\n")
		nextIndent := indent + "\t"

		for i, f := range desc.Children {
			sb.WriteString(nextIndent)

			// For MAP with "*" key, just use "*"
			if desc.Kind == TypeKind_StrMap && f.Name == "*" {
				sb.WriteString("\"*\": ")
			} else {
				sb.WriteString("\"" + f.Name + "\": ")
			}

			if f.Desc == nil {
				sb.WriteString("-")
			} else {
				printer(f.Desc, nextIndent)
			}

			if i < len(desc.Children)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}

		sb.WriteString(indent + "}")
	}

	printer(d, "")
	return sb.String()
}

// descriptorJSON is the JSON representation of Descriptor
type descriptorJSON struct {
	Kind     TypeKind    `json:"kind"`
	Name     string      `json:"name"`
	Children []fieldJSON `json:"children,omitempty"`
}

// fieldJSON is the JSON representation of Field
type fieldJSON struct {
	Name string          `json:"name"`
	ID   int             `json:"id"`
	Desc *descriptorJSON `json:"desc,omitempty"`
	Ref  string          `json:"$ref,omitempty"` // reference to another descriptor by path
}

// MarshalJSON implements json.Marshaler interface for Descriptor
// It handles circular references by using $ref to reference already visited descriptors
func (d *Descriptor) MarshalJSON() ([]byte, error) {
	visited := make(map[*Descriptor]string) // maps pointer to path
	result := d.marshalWithPath("$", visited)
	return json.Marshal(result)
}

// marshalWithPath recursively marshals the descriptor, tracking visited nodes
func (d *Descriptor) marshalWithPath(path string, visited map[*Descriptor]string) *descriptorJSON {
	if d == nil {
		return nil
	}

	// Check if we've already visited this descriptor (circular reference)
	if existingPath, ok := visited[d]; ok {
		// Return a reference placeholder - this will be handled specially
		return &descriptorJSON{
			Kind: d.Kind,
			Name: fmt.Sprintf("$ref:%s", existingPath),
		}
	}

	// Mark as visited
	visited[d] = path

	result := &descriptorJSON{
		Kind:     d.Kind,
		Name:     d.Type,
		Children: make([]fieldJSON, 0, len(d.Children)),
	}

	for i, f := range d.Children {
		childPath := fmt.Sprintf("%s.children[%d].desc", path, i)
		fj := fieldJSON{
			Name: f.Name,
			ID:   f.ID,
		}

		if f.Desc != nil {
			// Check if child descriptor was already visited
			if existingPath, ok := visited[f.Desc]; ok {
				fj.Ref = existingPath
			} else {
				fj.Desc = f.Desc.marshalWithPath(childPath, visited)
			}
		}

		result.Children = append(result.Children, fj)
	}

	return result
}

// UnmarshalJSON implements json.Unmarshaler interface for Descriptor
// It handles circular references by resolving $ref references after initial parsing
func (d *Descriptor) UnmarshalJSON(data []byte) error {
	var raw descriptorJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// First pass: build all descriptors and collect references
	refs := make(map[string]*Descriptor) // path -> descriptor
	d.unmarshalFromJSON(&raw, "$", refs)

	// Second pass: resolve references
	d.resolveRefs("$", refs)

	return nil
}

// unmarshalFromJSON populates the descriptor from JSON representation
func (d *Descriptor) unmarshalFromJSON(raw *descriptorJSON, path string, refs map[string]*Descriptor) {
	d.Kind = raw.Kind
	d.Type = raw.Name
	d.Children = make([]Field, 0, len(raw.Children))
	d.ids = nil
	d.names = nil

	// Register this descriptor
	refs[path] = d

	for i, fj := range raw.Children {
		childPath := fmt.Sprintf("%s.children[%d].desc", path, i)
		f := Field{
			Name: fj.Name,
			ID:   fj.ID,
		}

		if fj.Ref != "" {
			// This is a reference, will be resolved later
			// Create a placeholder descriptor with special name
			f.Desc = &Descriptor{Type: "$ref:" + fj.Ref}
		} else if fj.Desc != nil {
			f.Desc = &Descriptor{}
			f.Desc.unmarshalFromJSON(fj.Desc, childPath, refs)
		}

		d.Children = append(d.Children, f)
	}
}

// resolveRefs resolves all $ref references in the descriptor tree
func (d *Descriptor) resolveRefs(path string, refs map[string]*Descriptor) {
	for i := range d.Children {
		if d.Children[i].Desc != nil {
			childPath := fmt.Sprintf("%s.children[%d].desc", path, i)

			// Check if this is a reference
			if strings.HasPrefix(d.Children[i].Desc.Type, "$ref:") {
				refPath := strings.TrimPrefix(d.Children[i].Desc.Type, "$ref:")
				if target, ok := refs[refPath]; ok {
					d.Children[i].Desc = target
				}
			} else {
				// Recursively resolve references
				d.Children[i].Desc.resolveRefs(childPath, refs)
			}
		}
	}
}

// stackFrame represents a single frame in the path tracking stack
type stackFrame struct {
	kind TypeKind
	name string
	id   int
}

// stackFramePool is a pool for stackFrame slices to reduce allocations
var stackFramePool = sync.Pool{
	New: func() interface{} {
		ret := make([]stackFrame, 0, 16) // pre-allocate for common depth
		return (*pathStack)(&ret)
	},
}

// getStackFrames gets a stackFrame slice from the pool
func getStackFrames() *pathStack {
	return stackFramePool.Get().(*pathStack)
}

// putStackFrames returns a stackFrame slice to the pool
func putStackFrames(s *pathStack) {
	if s == nil {
		return
	}
	*s = (*s)[:0]         // reset slice
	stackFramePool.Put(s) //nolint:staticcheck // SA6002: storing slice in Pool is intentional
}

// pathStack tracks the path from root to current node
type pathStack []stackFrame

// push adds a new frame to the stack
func (s *pathStack) push(kind TypeKind, name string, id int) {
	*s = append(*s, stackFrame{
		kind: kind,
		name: name,
		id:   id,
	})
}

// pop removes the last frame from the stack
func (s *pathStack) pop() {
	if len(*s) > 0 {
		*s = (*s)[:len(*s)-1]
	}
}

// buildPath constructs a human-readable DSL path from the stack
// This is only called when an error occurs
func (s *pathStack) buildPath() string {
	if len(*s) == 0 {
		return "$"
	}

	var sb strings.Builder
	sb.WriteString("$")

	for _, frame := range *s {
		if frame.kind == TypeKind_StrMap {
			sb.WriteString("[")
			sb.WriteString(frame.name)
			sb.WriteString("]")
		} else if frame.kind == TypeKind_List {
			sb.WriteString("[")
			sb.WriteString(strconv.Itoa(frame.id))
			sb.WriteString("]")
		} else {
			sb.WriteString(".")
			sb.WriteString(frame.name)
		}
	}

	return sb.String()
}
