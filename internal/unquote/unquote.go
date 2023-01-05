/*
 * Copyright 2022 CloudWeGo Authors.
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

package unquote

import (
	"reflect"
	"runtime"
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/native"
	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/cloudwego/dynamicgo/internal/rt"
)

func String(s string) (ret string, err types.ParsingError) {
	mm := make([]byte, 0, len(s))
	err = intoBytesUnsafe(s, &mm)
	ret = rt.Mem2Str(mm)
	return
}

func IntoBytes(s string, m *[]byte) types.ParsingError {
	if cap(*m) < len(s) {
		return types.ERR_EOF
	} else {
		return intoBytesUnsafe(s, m)
	}
}

func intoBytesUnsafe(s string, m *[]byte) types.ParsingError {
	pos := -1
	slv := (*rt.GoSlice)(unsafe.Pointer(m))
	str := (*rt.GoString)(unsafe.Pointer(&s))
	ret := native.Unquote(str.Ptr, str.Len, slv.Ptr, &pos, 0)

	/* check for errors */
	if ret < 0 {
		return types.ParsingError(-ret)
	}

	/* update the length */
	slv.Len = ret
	return 0
}

var typeByte = rt.UnpackType(reflect.TypeOf(byte(0)))

//go:linkname growslice runtime.growslice
//goland:noinspection GoUnusedParameter
func growslice(et *rt.GoType, old rt.GoSlice, cap int) rt.GoSlice

func QuoteIntoBytes(val string, buf *[]byte) {
	sp := (*rt.GoString)(unsafe.Pointer(&val)).Ptr
	nb := len(val)
	b := (*rt.GoSlice)(unsafe.Pointer(buf))
	// input buffer
	for nb > 0 {
		// output buffer
		dp := unsafe.Pointer(uintptr(b.Ptr) + uintptr(b.Len))
		dn := b.Cap - b.Len
		// call native.Quote, dn is byte count it outputs
		ret := native.Quote(sp, nb, dp, &dn, 0)
		// update *buf length
		b.Len += dn

		// no need more output
		if ret >= 0 {
			break
		}

		// double buf size
		*b = growslice(typeByte, *b, b.Cap*2)
		// ret is the complement of consumed input
		ret = ^ret
		// update input buffer
		nb -= ret
		sp = unsafe.Pointer(uintptr(sp) + uintptr(ret))
	}
	runtime.KeepAlive(buf)
	runtime.KeepAlive(sp)
}

func QuoteString(val string) string {
	buf := make([]byte, 0, len(val)+2)
	QuoteIntoBytes(val, &buf)
	return rt.Mem2Str(buf)
}
