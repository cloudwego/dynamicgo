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
	sizePathNode = unsafe.Sizeof(PathNode{})
)

var (
	// UseNativeSkipForGet indicates to use native.Skip (instead of go.Skip) method to skip thrift value
	// This only works for single-value searching API like GetByInt()/GetByRaw()/GetByStr()/Field()/Index()/GetByPath() methods.
	// WARN: this will promote performance when thrift value to be skipped is large, but may decrease preformance when thrift value is small.
	UseNativeSkipForGet = false

	// DefaultNodeSliceCap is the default capacity of a Node or NodePath slice
	// Usually, a Node or NodePath slice is used to store intermediate or consequential elements of a generic API like Children()|Interface()|SetMany()
	DefaultNodeSliceCap = 16
)

// Opions for generic.Node
type Options struct {
	// DisallowUnknow indicates to report error when read unknown fields.
	DisallowUnknow bool

	// WriteDefault indicates to write value if a DEFAULT requireness field is not set.
	WriteDefault bool

	// NoCheckRequireNess indicates not to check requiredness when writing.
	NotCheckRequireNess bool

	// UseNativeSkip indicates to use native.Skip (instead of go.Skip) method to skip thrift value
	//  WARNING: this will promote performance when thrift value to be skipped is large, but may decrease preformance when thrift value is small.
	UseNativeSkip bool

	// MapStructById indicates to use FieldId instead of int as map key instead of when call Node.Interface() on STRUCT type.
	MapStructById bool

	// CastStringAsBinary indicates to cast STRING type to []byte when call Node.Interface()/Map().
	CastStringAsBinary bool

	// NotScanParentNode indicates to only assign children node when PathNode.Load()/Node.Children.
	// Thies will promote performance but may be misued when handle PathNode.
	NotScanParentNode bool

	// ClearDirtyValues indicates one multi-query (includeing
	// Fields()/GetMany()/Gets()/Indexies()) to clear out all nodes
	// in passed []PathNode first
	ClearDirtyValues bool

	// StoreChildrenById indicates to store children node by id when call Node.Children() or PathNode.Load().
	// When field id exceeds StoreChildrenByIdShreshold, children node will be stored sequentially after the threshold.
	StoreChildrenById bool

	// StoreChildrenByHash indicates to store children node by str hash (mod parent's size) when call Node.Children() or PathNode.Load().
	StoreChildrenByHash bool

	// IterateStructByName indicates `Value.Foreach()` API to pass PathFieldName instead of PathFieldId to handler.
	IterateStructByName bool
}

var (
	// StoreChildrenByIdShreshold is the maximum id to store children node by id.
	StoreChildrenByIdShreshold  = 256

	// StoreChildrenByIdShreshold is the minimum id to store children node by hash.
	StoreChildrenByIntHashShreshold = DefaultNodeSliceCap
)

