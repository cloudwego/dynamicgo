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

package rt

import (
	"unsafe"
)

//go:nosplit
func Get16(v []byte) int16 {
	return *(*int16)((*GoSlice)(unsafe.Pointer(&v)).Ptr)
}

//go:nosplit
func Get32(v []byte) int32 {
	return *(*int32)((*GoSlice)(unsafe.Pointer(&v)).Ptr)
}

//go:nosplit
func Get64(v []byte) int64 {
	return *(*int64)((*GoSlice)(unsafe.Pointer(&v)).Ptr)
}

//go:nosplit
func Mem2Str(v []byte) (s string) {
	(*GoString)(unsafe.Pointer(&s)).Len = (*GoSlice)(unsafe.Pointer(&v)).Len
	(*GoString)(unsafe.Pointer(&s)).Ptr = (*GoSlice)(unsafe.Pointer(&v)).Ptr
	return
}

//go:nosplit
func Str2Mem(s string) (v []byte) {
	(*GoSlice)(unsafe.Pointer(&v)).Cap = (*GoString)(unsafe.Pointer(&s)).Len
	(*GoSlice)(unsafe.Pointer(&v)).Len = (*GoString)(unsafe.Pointer(&s)).Len
	(*GoSlice)(unsafe.Pointer(&v)).Ptr = (*GoString)(unsafe.Pointer(&s)).Ptr
	return
}

//go:nosplit
func BytesFrom(p unsafe.Pointer, n int, c int) (r []byte) {
	(*GoSlice)(unsafe.Pointer(&r)).Ptr = p
	(*GoSlice)(unsafe.Pointer(&r)).Len = n
	(*GoSlice)(unsafe.Pointer(&r)).Cap = c
	return
}

//go:nosplit
func StringFrom(p unsafe.Pointer, n int) (r string) {
	(*GoString)(unsafe.Pointer(&r)).Ptr = p
	(*GoString)(unsafe.Pointer(&r)).Len = n
	return
}

func IndexSlice(slice unsafe.Pointer, size uintptr, index int) unsafe.Pointer {
	// if slice.Ptr == nil || slice.Cap == 0 {
	// 	return nil
	// }
	return unsafe.Pointer(uintptr(*(*unsafe.Pointer)(slice)) + uintptr(index)*size)
}

func IndexPtr(ptr unsafe.Pointer, size uintptr, index int) unsafe.Pointer {
	// if slice.Ptr == nil || slice.Cap == 0 {
	// 	return nil
	// }
	return unsafe.Pointer(uintptr(ptr) + uintptr(index)*size)
}

func IndexByte(ptr *[]byte, index int) unsafe.Pointer {
	// if slice.Ptr == nil || slice.Cap == 0 {
	// 	return nil
	// }
	return unsafe.Pointer(uintptr((*GoSlice)(unsafe.Pointer(ptr)).Ptr) + uintptr(index))
}

func IndexChar(src string, index int) unsafe.Pointer {
	// if slice.Ptr == nil || slice.Cap == 0 {
	// 	return nil
	// }
	return unsafe.Pointer(uintptr((*GoString)(unsafe.Pointer(&src)).Ptr) + uintptr(index))
}

func IndexCharUint(src string, index int) uintptr {
	// if slice.Ptr == nil || slice.Cap == 0 {
	// 	return nil
	// }
	return uintptr((*GoString)(unsafe.Pointer(&src)).Ptr) + uintptr(index)
}

func StrBoundary(src string) uintptr {
	return uintptr((*GoString)(unsafe.Pointer(&src)).Ptr) + uintptr(len(src))
}

func GuardSlice(buf *[]byte, n int) {
	c := cap(*buf)
	l := len(*buf)
	if c-l < n {
		c = c>>1 + n + l
		if c < 1024 {
			c = 1024
		}
		tmp := make([]byte, l, c)
		copy(tmp, *buf)
		*buf = tmp
	}
}

func AddPtr(a unsafe.Pointer, b uintptr) unsafe.Pointer {
	return unsafe.Add(a, b)
}

func SubPtr(a unsafe.Pointer, b uintptr) unsafe.Pointer {
	return unsafe.Add(a, -b)
}

func PtrOffset(a uintptr, b uintptr) int {
	return int(uintptr(a)) - int(uintptr(b))
}

func GetBytePtr(b []byte) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&b))
}

func Growslice(et *GoType, old GoSlice, cap int) GoSlice {
	return growslice(et, old, cap)
}

// NoEscape hides a pointer from escape analysis. NoEscape is
// the identity function but escape analysis doesn't think the
// output depends on the input. NoEscape is inlined and currently
// compiles down to zero instructions.
// USE CAREFULLY!
//
//go:nosplit
//goland:noinspection GoVetUnsafePointer
func NoEscape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}
