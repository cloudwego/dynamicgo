//go:build amd64 && !go1.25

// Copyright 2023 CloudWeGo Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package json

import (
	"runtime"
	"unsafe"

	"github.com/cloudwego/base64x"
	"github.com/cloudwego/dynamicgo/internal/native"
	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/cloudwego/dynamicgo/internal/rt"
)

var typeByte = rt.UnpackEface(byte(0)).Type

// NoQuote only escape inner '\' and '"' of one string, but it does add quotes besides string.
//
//go:nocheckptr
func NoQuote(buf *[]byte, val string) {
	sp := rt.IndexChar(val, 0)
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
		*b = rt.Growslice(typeByte, *b, b.Cap*2)
		// ret is the complement of consumed input
		ret = ^ret
		// update input buffer
		nb -= ret
		sp = unsafe.Pointer(uintptr(sp) + uintptr(ret))
	}

	runtime.KeepAlive(buf)
	runtime.KeepAlive(sp)
}

func i64toa(buf *[]byte, val int64) int {
	rt.GuardSlice(buf, types.MaxInt64StringLen)
	s := len(*buf)
	ret := native.I64toa((*byte)(rt.IndexPtr((*rt.GoSlice)(unsafe.Pointer(buf)).Ptr, typeByte.Size, s)), val)
	if ret < 0 {
		*buf = append((*buf)[s:], '0')
		return 1
	}
	*buf = (*buf)[:s+ret]
	return ret
}

func f64toa(buf *[]byte, val float64) int {
	rt.GuardSlice(buf, types.MaxFloat64StringLen)
	s := len(*buf)
	ret := native.F64toa((*byte)(rt.IndexPtr((*rt.GoSlice)(unsafe.Pointer(buf)).Ptr, typeByte.Size, s)), val)
	if ret < 0 {
		*buf = append((*buf)[s:], '0')
		return 1
	}
	*buf = (*buf)[:s+ret]
	return ret
}

func encodeBase64(src []byte) string {
	return base64x.StdEncoding.EncodeToString(src)
}
