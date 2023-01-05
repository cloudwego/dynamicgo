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

var (
	V_strhash    = rt.UnpackEface(strhash)
	S_strhash    = *(*uintptr)(V_strhash.Value)
	byteTypeSize = unsafe.Sizeof(byte(0))
)

//go:noescape
//go:linkname strhash runtime.strhash
func strhash(_ unsafe.Pointer, _ uintptr) uintptr

func StrHash(s string) uint64 {
	if v := strhash(unsafe.Pointer(&s), 0); v == 0 {
		return 1
	} else {
		return uint64(v)
	}
}

func DJBHash32(k string) uint32 {
	var hash uint32 = 5381
	var ks = *(*unsafe.Pointer)(unsafe.Pointer(&k))
	for i := 0; i < len(k); i++ {
		c := *(*byte)(rt.IndexPtr(ks, byteTypeSize, i))
		hash = ((hash << 5) + hash + uint32(c))
	}
	return hash
}

func DJBHash64(k string) uint64 {
	var hash uint64 = 5381
	var ks = *(*unsafe.Pointer)(unsafe.Pointer(&k))
	for i := 0; i < len(k); i++ {
		c := *(*byte)(rt.IndexPtr(ks, byteTypeSize, i))
		hash = ((hash << 5) + hash + uint64(c))
	}
	return hash
}
