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
	"sync"
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
	sync.Once
	ids   map[int]Field
	names map[string]Field
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

// Normalize cache all the fields in the descriptor for speeding up search
func (d *Descriptor) Normalize() {
	d.Once.Do(func() {
		d.ids = make(map[int]Field, len(d.Children))
		d.names = make(map[string]Field, len(d.Children))
		for _, f := range d.Children {
			d.ids[f.ID] = f
			d.names[f.Name] = f
			if f.Desc != nil {
				f.Desc.Normalize()
			}
		}
	})
}

// String returns the string representation of the descriptor.
func (d *Descriptor) String() string {
	sb := strings.Builder{}
	var printer func(*Descriptor)
	printer = func(pbtr *Descriptor) {
		sb.WriteString("<" + pbtr.Name + ">")
		for _, f := range pbtr.Children {
			sb.WriteString("--" + f.Name)
			if f.Desc == nil {
				sb.WriteString("\n")
				continue
			}
			sb.WriteString("->")
			printer(f.Desc)
		}
	}
	printer(d)
	return sb.String()
}
