//go:build !amd64 || go1.25

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
	"encoding/base64"
	"strconv"
	"unicode/utf8"
	_ "unsafe"

	"github.com/cloudwego/dynamicgo/internal/rt"
)

func NoQuote(buf *[]byte, val string) {
	quoteString(buf, val)
}

func quoteString(e *[]byte, s string) {
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if rt.SafeSet[b] {
				i++
				continue
			}
			if start < i {
				*e = append(*e, s[start:i]...)
			}
			*e = append(*e, '\\')
			switch b {
			case '\\', '"':
				*e = append(*e, b)
			case '\n':
				*e = append(*e, 'n')
			case '\r':
				*e = append(*e, 'r')
			case '\t':
				*e = append(*e, 't')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				*e = append(*e, `u00`...)
				*e = append(*e, rt.Hex[b>>4])
				*e = append(*e, rt.Hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		// if c == utf8.RuneError && size == 1 {
		//     if start < i {
		//         e.Write(s[start:i])
		//     }
		//     e.WriteString(`\ufffd`)
		//     i += size
		//     start = i
		//     continue
		// }
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				*e = append(*e, s[start:i]...)
			}
			*e = append(*e, `\u202`...)
			*e = append(*e, rt.Hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		*e = append(*e, s[start:]...)
	}
}

func encodeBase64(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func i64toa(buf *[]byte, v int64) {
	*buf = strconv.AppendInt(*buf, v, 10)
}

func f64toa(buf *[]byte, v float64) {
	*buf = strconv.AppendFloat(*buf, v, 'g', -1, 64)
}
