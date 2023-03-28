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

package thrift

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/caching"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
)

const (
	defaultMaxBucketSize     float64 = 10
	defaultMapSize           int     = 4
	defaultHashMapLoadFactor int     = 4
	defaultMaxFieldID                = 256
	defaultMaxNestedDepth            = 1024
)

// FieldNameMap is a map for field name and field descriptor
type FieldNameMap struct {
	maxKeyLength int
	all          []caching.Pair
	trie         *caching.TrieTree
	hash         *caching.HashMap
}

// Set sets the field descriptor for the given key
func (ft *FieldNameMap) Set(key string, field *FieldDescriptor) (exist bool) {
	if len(key) > ft.maxKeyLength {
		ft.maxKeyLength = len(key)
	}
	for i, v := range ft.all {
		if v.Key == key {
			exist = true
			ft.all[i].Val = unsafe.Pointer(field)
			return
		}
	}
	ft.all = append(ft.all, caching.Pair{Val: unsafe.Pointer(field), Key: key})
	return
}

// Get gets the field descriptor for the given key
func (ft FieldNameMap) Get(k string) *FieldDescriptor {
	if ft.trie != nil {
		return (*FieldDescriptor)(ft.trie.Get(k))
	} else if ft.hash != nil {
		return (*FieldDescriptor)(ft.hash.Get(k))
	}
	return nil
}

// All returns all field descriptors
func (ft FieldNameMap) All() []*FieldDescriptor {
	return *(*[]*FieldDescriptor)(unsafe.Pointer(&ft.all))
}

// Size returns the size of the map
func (ft FieldNameMap) Size() int {
	if ft.hash != nil {
		return ft.hash.Size()
	} else {
		return ft.trie.Size()
	}
}

// Build builds the map.
// It will try to build a trie tree if the dispersion of keys is higher enough (min).
func (ft *FieldNameMap) Build() {
	var empty unsafe.Pointer

	// statistics the distrubution for each position:
	//   - primary slice store the position as its index
	//   - secondary map used to merge values with same char at the same position
	var positionDispersion = make([]map[byte][]int, ft.maxKeyLength)

	for i, v := range ft.all {
		for j := ft.maxKeyLength - 1; j >= 0; j-- {
			if v.Key == "" {
				// empty key, especially store
				empty = v.Val
			}
			// get the char at the position, defualt (position beyonds key range) is ASCII 0
			var c = byte(0)
			if j < len(v.Key) {
				c = v.Key[j]
			}

			if positionDispersion[j] == nil {
				positionDispersion[j] = make(map[byte][]int, 16)
			}
			// recoder the index i of the value with same char c at the same position j
			positionDispersion[j][c] = append(positionDispersion[j][c], i)
		}
	}

	// calculate the best position which has the highest dispersion
	var idealPos = -1
	var min = defaultMaxBucketSize
	var count = len(ft.all)

	for i := ft.maxKeyLength - 1; i >= 0; i-- {
		cd := positionDispersion[i]
		l := len(cd)
		// calculate the dispersion (average bucket size)
		f := float64(count) / float64(l)
		if f < min {
			min = f
			idealPos = i
		}
		// 1 means all the value store in different bucket, no need to continue calulating
		if min == 1 {
			break
		}
	}

	if idealPos != -1 {
		// find the best position, build a trie tree
		ft.hash = nil
		ft.trie = &caching.TrieTree{}
		// NOTICE: we only use a two-layer tree here, for better performance
		ft.trie.Positions = append(ft.trie.Positions, idealPos)
		// set all key-values to the trie tree
		for _, v := range ft.all {
			ft.trie.Set(v.Key, v.Val)
		}
		if empty != nil {
			ft.trie.Empty = empty
		}

	} else {
		// no ideal position, build a hash map
		ft.trie = nil
		ft.hash = caching.NewHashMap(len(ft.all), defaultHashMapLoadFactor)
		// set all key-values to the trie tree
		for _, v := range ft.all {
			// caching.HashMap does not support duplicate key, so must check if the key exists before set
			// WARN: if the key exists, the value WON'T be replaced
			o := ft.hash.Get(v.Key)
			if o == nil {
				ft.hash.Set(v.Key, v.Val)
			}
		}
		if empty != nil {
			ft.hash.Set("", empty)
		}
	}
}

// FieldIDMap is a map from field id to field descriptor
type FieldIDMap struct {
	m   []*FieldDescriptor
	all []*FieldDescriptor
}

// All returns all field descriptors
func (fd FieldIDMap) All() (ret []*FieldDescriptor) {
	return fd.all
}

// Size returns the size of the map
func (fd FieldIDMap) Size() int {
	return len(fd.m)
}

// Get gets the field descriptor for the given id
func (fd FieldIDMap) Get(id FieldID) *FieldDescriptor {
	if int(id) >= len(fd.m) {
		return nil
	}
	return fd.m[id]
}

// Set sets the field descriptor for the given id
func (fd *FieldIDMap) Set(id FieldID, f *FieldDescriptor) {
	if int(id) >= len(fd.m) {
		len := int(id) + 1
		tmp := make([]*FieldDescriptor, len)
		copy(tmp, fd.m)
		fd.m = tmp
	}
	o := (fd.m)[id]
	if o == nil {
		fd.all = append(fd.all, f)
	} else {
		for i, v := range fd.all {
			if v == o {
				fd.all[i] = f
				break
			}
		}
	}
	fd.m[id] = f
}

// RequiresBitmap is a bitmap to mark fields
type RequiresBitmap []uint64

const (
	int64BitSize  = 64
	int64ByteSize = 8
)

var bitmapPool = sync.Pool{
	New: func() interface{} {
		ret := RequiresBitmap(make([]uint64, 0, defaultMaxFieldID/int64BitSize+1))
		return &ret
	},
}

// Set mark the bit corresponding the given id, with the given requireness
//   - RequiredRequireness|DefaultRequireness mark the bit as 1
//   - OptionalRequireness mark the bit as 0
func (b *RequiresBitmap) Set(id FieldID, val Requireness) {
	i := int(id) / int64BitSize
	if len(*b) <= i {
		b.malloc(int32(id))
	}
	p := unsafe.Pointer(uintptr((*rt.GoSlice)(unsafe.Pointer(b)).Ptr) + uintptr(i)*int64ByteSize)
	switch val {
	case RequiredRequireness, DefaultRequireness:
		*(*uint64)(p) |= (0b1 << (id % int64BitSize))
	case OptionalRequireness:
		*(*uint64)(p) &= ^(0b1 << (id % int64BitSize))
	default:
		panic("invalid requireness")
	}
}

// IsSet tells if the bit corresponding the given id is marked
func (b RequiresBitmap) IsSet(id FieldID) bool {
	i := int(id) / int64BitSize
	if i >= len(b) {
		panic("bitmap id out of range")
	}
	return (b[i] & (0b1 << (id % int64BitSize))) != 0
}

func (b *RequiresBitmap) malloc(id int32) {
	if n := int32(id / int64BitSize); int(n) >= len(*b) {
		buf := make([]uint64, n+1, int32((n+1)*2))
		copy(buf, *b)
		*b = buf
	}
}

// CopyTo copy the bitmap to a given bitmap
func (b RequiresBitmap) CopyTo(to *RequiresBitmap) {
	c := cap(*to)
	l := len(b)
	if l > c {
		*to = make([]uint64, l)
	}
	*to = (*to)[:l]
	copy(*to, b)
}

// NewRequiresBitmap get bitmap from pool, if pool is empty, create a new one
// WARN: memory from pool maybe dirty!
func NewRequiresBitmap() *RequiresBitmap {
	return bitmapPool.Get().(*RequiresBitmap)
}

// FreeRequiresBitmap free the bitmap, but not clear its memory
func FreeRequiresBitmap(b *RequiresBitmap) {
	// memclrNoHeapPointers(*(*unsafe.Pointer)(unsafe.Pointer(b)), uintptr(len(*b))*uint64TypeSize)
	*b = (*b)[:0]
	bitmapPool.Put(b)
}

//go:nocheckptr
// CheckRequires scan every bit of the bitmap. When a bit is marked, it will:
//   - if the corresponding field is required-requireness, it reports error
//   - if the corresponding is not required-requireness but writeDefault is true, it will call handler to handle this field
func (b RequiresBitmap) CheckRequires(desc *StructDescriptor, writeDefault bool, handler func(field *FieldDescriptor) error) error {
	// handle bitmap first
	n := len(b)
	s := (*rt.GoSlice)(unsafe.Pointer(&b)).Ptr

	// test 64 bits once
	for i := 0; i < n; i++ {
		v := *(*uint64)(s)
		for j := 0; v != 0 && j < int64BitSize; j++ {
			if v%2 == 1 {
				id := FieldID(i*int64BitSize + j)
				f := desc.FieldById(id)
				if f == nil {
					return errInvalidBitmapId(id, desc)
				}
				if f.Required() == RequiredRequireness {
					return errMissRequiredField(f, desc)
				} else if !writeDefault {
					v >>= 1
					continue
				}
				if err := handler(f); err != nil {
					return err
				}
			}
			v >>= 1
		}
		s = unsafe.Pointer(uintptr(s) + int64ByteSize)
	}
	runtime.KeepAlive(s)
	return nil
}

//go:nocheckptr
// CheckRequires scan every bit of the bitmap. When a bit is marked, it will:
//   - if the corresponding field is required-requireness and writeRquired is true, it will call handler to handle this field, otherwise report error
//   - if the corresponding is default-requireness and writeDefault is true, it will call handler to handle this field
//   - if the corresponding is optional-requireness and writeOptional is true, it will call handler to handle this field
func (b RequiresBitmap) HandleRequires(desc *StructDescriptor, writeRquired bool, writeDefault bool, writeOptional bool, handler func(field *FieldDescriptor) error) error {
	// handle bitmap first
	n := len(b)
	s := (*rt.GoSlice)(unsafe.Pointer(&b)).Ptr
	// test 64 bits once
	for i := 0; i < n; i++ {
		v := *(*uint64)(s)
		for j := 0; v != 0 && j < int64BitSize; j++ {
			if v%2 == 1 {
				f := desc.FieldById(FieldID(i*int64BitSize + j))
				if f.Required() == RequiredRequireness && !writeRquired {
					return errMissRequiredField(f, desc)
				}
				if (f.Required() == DefaultRequireness && !writeDefault) || (f.Required() == OptionalRequireness && !writeOptional) {
					v >>= 1
					continue
				}
				if err := handler(f); err != nil {
					return err
				}
			}
			v >>= 1
		}
		s = unsafe.Pointer(uintptr(s) + int64ByteSize)
	}
	runtime.KeepAlive(s)
	return nil
}

func errMissRequiredField(field *FieldDescriptor, st *StructDescriptor) error {
	return meta.NewError(meta.ErrMissRequiredField, fmt.Sprintf("miss required field '%s' of struct '%s'", field.Name(), st.Name()), nil)
}

func errInvalidBitmapId(id FieldID, st *StructDescriptor) error {
	return meta.NewError(meta.ErrInvalidParam, fmt.Sprintf("invalid field id %d of struct '%s'", id, st.Name()), nil)
}

// DefaultValue is the default value of a field
type DefaultValue struct {
	goValue      interface{}
	jsonValue    string
	thriftBinary string
}

// GoValue return the go runtime representation of the default value
func (d DefaultValue) GoValue() interface{} {
	return d.goValue
}

// JSONValue return the json-encoded representation of the default value
func (d DefaultValue) JSONValue() string {
	return d.jsonValue
}

// ThriftBinary return the thrift-binary-encoded representation of the default value
func (d DefaultValue) ThriftBinary() string {
	return d.thriftBinary
}
