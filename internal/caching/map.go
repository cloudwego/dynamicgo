/*
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

package caching

import (
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/rt"
)

type HashMap struct {
	b unsafe.Pointer
	c uint32
	N uint32
}

type Entry struct {
	Hash uint32
	Key  string
	Val  unsafe.Pointer
}

const (
	FieldMap_N     = int64(unsafe.Offsetof(HashMap{}.N))
	FieldMap_b     = int64(unsafe.Offsetof(HashMap{}.b))
	FieldEntrySize = int64(unsafe.Sizeof(Entry{}))
)

func newBucket(n int) unsafe.Pointer {
	v := make([]Entry, n)
	return (*rt.GoSlice)(unsafe.Pointer(&v)).Ptr
}

func NewHashMap(n int, loadFactor int) *HashMap {
	return &HashMap{
		N: uint32(n * loadFactor),
		c: 0,
		b: newBucket(n * loadFactor),
	}
}

func (self *HashMap) At(p uint32) *Entry {
	return (*Entry)(unsafe.Pointer(uintptr(self.b) + uintptr(p)*uintptr(FieldEntrySize)))
}

// Get searches FieldMap by name. JIT generated assembly does NOT call this
// function, rather it implements its own version directly in assembly. So
// we must ensure this function stays in sync with the JIT generated one.
func (self *HashMap) Get(name string) unsafe.Pointer {
	h := DJBHash32(name)
	p := h % self.N
	s := self.At(p)

	/* find the element;
	 * the hash map is never full, so the loop will always terminate */
	for s.Hash != 0 {
		if s.Hash == h && s.Key == name {
			return s.Val
		} else {
			p = (p + 1) % self.N
			s = self.At(p)
		}
	}

	/* not found */
	return nil
}

// WARN: this function is not idempotent, set same key twice will cause incorrect result.
func (self *HashMap) Set(name string, val unsafe.Pointer) {
	h := DJBHash32(name)
	p := h % self.N
	s := self.At(p)

	/* searching for an empty slot;
	 * the hash map is never full, so the loop will always terminate */
	for s.Hash != 0 {
		p = (p + 1) % self.N
		s = self.At(p)
	}

	/* set the value */
	s.Val = val
	s.Hash = h
	s.Key = name
	self.c++
}

func (self *HashMap) Size() int {
	return int(self.c)
}
