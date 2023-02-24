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
	"github.com/cloudwego/dynamicgo/internal/native/types"
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
	if len(val) == 0 {
		buf = append(buf, '"', '"')
		return buf
	}

	quote(&buf, val)
	return buf
}

// func EncodeNumber(buf []byte, val json.Number) ([]byte, error)

func EncodeInt64(buf []byte, val int64) []byte {
	i64toa(&buf, val)
	return buf
}

func EncodeFloat64(buf []byte, val float64) ([]byte) {
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

func DecodeNull(src string, pos int) (ret int) {
	return decodeNull(src, SkipBlank(src, pos))
}

func DecodeBool(src string, pos int) (ret int, v bool) {
	pos = SkipBlank(src, pos)
	ret = pos + 4
	if ret > len(src) {
		return -int(types.ERR_EOF), false
	}
	if src[pos:ret] == bytesTrue {
		return ret, true
	}
	ret += 1
	if ret > len(src) {
		return -int(types.ERR_EOF), false
	}
	if src[pos:ret] == bytesFalse {
		return ret, false
	}
	return -int(types.ERR_INVALID_CHAR), false
}

func DecodeString(src string, pos int) (ret int, v string) {
	return decodeString(src, SkipBlank(src, pos))
}

func DecodeBinary(src string, pos int) (ret int, v []byte) {
	return decodeBinary(src, SkipBlank(src, pos))
}

func DecodeInt64(src string, pos int) (ret int, v int64, err error) {
	return decodeInt64(src, SkipBlank(src, pos))
}

func DecodeFloat64(src string, pos int) (ret int, v float64, err error) {
	return decodeFloat64(src, SkipBlank(src, pos))
}

func SkipNumber(src string, pos int) int {
	return skipNumber(src, SkipBlank(src, pos))
}

func SkipString(src string, pos int) (int, int) {
	return skipString(src, SkipBlank(src, pos))
}

func SkipArray(src string, pos int) int {
	return skipPair(src, SkipBlank(src, pos), '[', ']')
}

func SkipObject(src string, pos int) int {
	return skipPair(src, SkipBlank(src, pos), '{', '}')
}
