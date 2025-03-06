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

package json

import (
	"runtime"
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/cloudwego/dynamicgo/internal/rt"
)

const (
	bytesNull   = "null"
	bytesTrue   = "true"
	bytesFalse  = "false"
	bytesObject = "{}"
	bytesArray  = "[]"
)

func EncodeNull(buf []byte) []byte {
	return append(buf, bytesNull...)
}

func EncodeBool(buf []byte, val bool) []byte {
	if val {
		return append(buf, bytesTrue...)
	}
	return append(buf, bytesFalse...)
}

func EncodeBaniry(buf []byte, val []byte) []byte {
	out := append(buf, '"')
	nb := len(val)
	if nb == 0 {
		out = append(out, '"')
		return out
	}

	out = append(out, encodeBase64(val)...)
	return append(out, '"')
}

func EncodeBase64(buf []byte, val []byte) []byte {
	return append(buf, encodeBase64(val)...)
}

func EncodeString(buf []byte, val string) []byte {
	buf = append(buf, '"')
	if len(val) == 0 {
		buf = append(buf, '"')
		return buf
	}
	NoQuote(&buf, val)
	buf = append(buf, '"')
	return buf
}

func EncodeInt64(buf []byte, val int64) []byte {
	i64toa(&buf, val)
	return buf
}

func EncodeFloat64(buf []byte, val float64) []byte {
	f64toa(&buf, val)
	return buf
}

func EncodeArrayBegin(buf []byte) []byte {
	return append(buf, '[')
}

func EncodeArrayComma(buf []byte) []byte {
	return append(buf, ',')
}

func EncodeArrayEnd(buf []byte) []byte {
	return append(buf, ']')
}

func EncodeEmptyArray(buf []byte) []byte {
	return append(buf, '[', ']')
}

func EncodeObjectBegin(buf []byte) []byte {
	return append(buf, '{')
}

func EncodeObjectColon(buf []byte) []byte {
	return append(buf, ':')
}

func EncodeObjectComma(buf []byte) []byte {
	return append(buf, ',')
}

func EncodeObjectEnd(buf []byte) []byte {
	return append(buf, '}')
}

func EncodeEmptyObject(buf []byte) []byte {
	return append(buf, '{', '}')
}

func IsSpace(c byte) bool {
	return (int(1<<c) & _blankCharsMask) != 0
}

const _blankCharsMask = (1 << ' ') | (1 << '\t') | (1 << '\r') | (1 << '\n')

//go:nocheckptr
func SkipBlank(src string, pos int) int {
	se := rt.StrBoundary(src)
	sp := rt.IndexCharUint(src, pos)

	for sp < se {
		if !IsSpace(*(*byte)(unsafe.Pointer(sp))) {
			break
		}
		sp += 1
	}
	if sp >= se {
		return -int(types.ERR_EOF)
	}
	runtime.KeepAlive(src)
	return int(sp - rt.IndexCharUint(src, 0))
}
