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
	"strings"
	"sync/atomic"
)

// TypeKind is the kind of type.
type TypeKind int

const (
	// TypeKind_Scalar indicates Descriptor is a leaf node, its underlying type can be anything (event go struct/map/list)
	TypeKind_Scalar TypeKind = iota
	// TypeKind_Struct indicates Descriptor.Field is struct field
	TypeKind_Struct
	// TypeKind_StrMap indicates Descriptor.Field is map key
	TypeKind_StrMap
)

// Descriptor describes the entire a DSL-pruning scheme for a type.
// base on this, we can fetch the type's object data on demands
type Descriptor struct {
	// the kind of corresponding type
	// Based on this, we can decide how to manipulate the data (e.g. mapKey or strucField)
	Kind TypeKind

	// Name of the type
	Name string

	// children for TypeKind_Struct|TypeKind_StrMap|TypeKind_List
	// - For TypeKind_StrMap, either each Field is a key-value pair or one field with Name "*"
	// - For TypeKind_Struct, each Field is a field with both Name and ID
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
			sb.WriteString("<" + desc.Name + ">")
			return
		}
		visited[desc] = true

		// Get type prefix based on Kind
		var typePrefix string
		switch desc.Kind {
		case TypeKind_Scalar:
			sb.WriteString("-")
			return
		case TypeKind_StrMap:
			typePrefix = "<MAP>"
		default: // TypeKind_Struct
			typePrefix = "<" + desc.Name + ">"
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
