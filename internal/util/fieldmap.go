/**
 * Copyright 2024 ByteDance Inc.
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

package util

import (
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/caching"
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
func (ft *FieldNameMap) Set(key string, field unsafe.Pointer) (exist bool) {
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
func (ft FieldNameMap) Get(k string) unsafe.Pointer {
	if ft.trie != nil {
		return (unsafe.Pointer)(ft.trie.Get(k))
	} else if ft.hash != nil {
		return (unsafe.Pointer)(ft.hash.Get(k))
	}
	return nil
}

// All returns all field descriptors
func (ft FieldNameMap) All() []caching.Pair {
	return ft.all
}

// Size returns the size of the map
func (ft FieldNameMap) Size() int {
	if ft.hash != nil {
		return ft.hash.Size()
	} else if ft.trie != nil {
		return ft.trie.Size()
	}
	return 0
}

// Build builds the map.
// It will try to build a trie tree if the dispersion of keys is higher enough (min).
func (ft *FieldNameMap) Build() {
	if len(ft.all) == 0 {
		return
	}

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
	m   []unsafe.Pointer
	all []unsafe.Pointer
}

// All returns all field descriptors
func (fd FieldIDMap) All() (ret []unsafe.Pointer) {
	return fd.all
}

// Size returns the size of the map
func (fd FieldIDMap) Size() int {
	return len(fd.m)
}

// Get gets the field descriptor for the given id
func (fd FieldIDMap) Get(id int32) unsafe.Pointer {
	if int(id) >= len(fd.m) {
		return nil
	}
	return fd.m[id]
}

// Set sets the field descriptor for the given id
func (fd *FieldIDMap) Set(id int32, f unsafe.Pointer) {
	if int(id) >= len(fd.m) {
		len := int(id) + 1
		tmp := make([]unsafe.Pointer, len)
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
