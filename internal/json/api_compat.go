//go:build !amd64 || !go1.16
// +build !amd64 !go1.16

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
	"strconv"
	"encoding/base64"
	"encoding/json"

	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/cloudwego/dynamicgo/internal/rt"
)

func quote(buf *[]byte, val string) {
	*buf = append(*buf, '"')
	quoteString(buf, val)
	*buf = append(*buf, '"')
}

func NoQuote(buf *[]byte, val string) {
	quoteString(buf, val)
}

func unquote(src string) (string, types.ParsingError) {
	sp := rt.IndexChar(src, -1)
	out, ok := unquoteBytes(rt.BytesFrom(sp, len(src)+2, len(src)+2))
	if !ok {
		return "", types.ERR_INVALID_ESCAPE
	}
	return rt.Mem2Str(out), 0
}

func decodeBase64(src string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(src)
}

func encodeBase64(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func (self *Parser) decodeValue() (val types.JsonState) {
	e, v := DecodeValue(self.s, self.p)
	if e < 0 {
		return v
	}
	self.p = e
	return v
}

func (self *Parser) skip() (int, types.ParsingError) {
	e, s := SkipValue(self.s, self.p)
	if e < 0 {
		return self.p, types.ParsingError(-e)
	}
	self.p = e
	return s, 0
}

func (self *Node) encodeInterface(buf *[]byte) error {
	out, err := json.Marshal(self.packAny())
	if err != nil {
		return err
	}
	*buf = append(*buf, out...)
	return nil
}

func i64toa(buf *[]byte, v int64) {
	*buf = strconv.AppendInt(*buf, v, 10)
}

func f64toa(buf *[]byte, v float64) {
	*buf = strconv.AppendFloat(*buf, v, 'g', -1, 64)
}