/**
 * Copyright 2023 CloudWeGo Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package generic

import "unsafe"

const (
	sizePathNode = unsafe.Sizeof(PathNode{}) // not used
)

var (
	// UseNativeSkipForGet indicates to use native.Skip (instead of go.Skip) method to skip proto value
	// This only works for single-value searching API like GetByInt()/GetByRaw()/GetByStr()/Field()/Index()/GetByPath() methods.
	UseNativeSkipForGet = false

	// DefaultNodeSliceCap is the default capacity of a Node or NodePath slice
	// Usually, a Node or NodePath slice is used to store intermediate or consequential elements of a generic API like Children()|Interface()|SetMany()
	DefaultNodeSliceCap = 16
	DefaultTagSliceCap  = 8
)

// Opions for generic.Node
type Options struct {
	// DisallowUnknown indicates to report error when read unknown fields.
	DisallowUnknown bool

	// MapStructById indicates to use FieldId instead of int as map key instead of when call Node.Interface() on STRUCT type.
	MapStructById bool

	// ClearDirtyValues indicates one multi-query (includeing
	// Fields()/GetMany()/Gets()/Indexies()) to clear out all nodes
	// in passed []PathNode first
	ClearDirtyValues bool

	// CastStringAsBinary indicates to cast STRING type to []byte when call Node.Interface()/Map().
	CastStringAsBinary bool

	WriteDefault bool // not implemented

	UseNativeSkip bool // not implemented

	NotScanParentNode bool // not implemented

	StoreChildrenById bool // not implemented

	StoreChildrenByHash bool // not implemented

	IterateStructByName bool // not implemented
}

var (
	// StoreChildrenByIdShreshold is the maximum id to store children node by id.
	StoreChildrenByIdShreshold  = 256

	// StoreChildrenByIdShreshold is the minimum id to store children node by hash.
	StoreChildrenByIntHashShreshold = DefaultNodeSliceCap
)

