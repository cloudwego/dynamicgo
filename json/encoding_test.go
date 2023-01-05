/**
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

package json

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/cloudwego/dynamicgo/internal/native"
	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/stretchr/testify/require"
)

func TestUnquote(t *testing.T) {
	var src = `"\ud83d\ude00"`
	var out string
	err := json.Unmarshal([]byte(src), &out)
	require.Nil(t, err)
	require.Equal(t, "😀", out)
}

func TestDecodeBool(t *testing.T) {
	var src = "true,"
	ret, v := DecodeBool(src, 0)
	require.Equal(t, 4, ret)
	require.Equal(t, true, v)

	src = "false,"
	ret, v = DecodeBool(src, 0)
	require.Equal(t, 5, ret)
	require.Equal(t, false, v)

	src = "tru"
	ret, v = DecodeBool(src, 0)
	require.Equal(t, -1, ret)
	require.Equal(t, false, v)

	src = "trua"
	ret, v = DecodeBool(src, 0)
	require.Equal(t, -1, ret)
	require.Equal(t, false, v)

	src = "fals"
	ret, v = DecodeBool(src, 0)
	require.Equal(t, -1, ret)
	require.Equal(t, false, v)

	src = "falss"
	ret, v = DecodeBool(src, 0)
	require.Equal(t, -2, ret)
	require.Equal(t, false, v)
}

func TestSkipSpace(t *testing.T) {
	var src = "\ra\nb c\td "
	ret := SkipBlank(src, 0)
	require.Equal(t, 1, ret)
	ret = SkipBlank(src, 1)
	require.Equal(t, 1, ret)
	ret = SkipBlank(src, 2)
	require.Equal(t, 3, ret)
	ret = SkipBlank(src, 3)
	require.Equal(t, 3, ret)
	ret = SkipBlank(src, 4)
	require.Equal(t, 5, ret)
	ret = SkipBlank(src, 5)
	require.Equal(t, 5, ret)
	ret = SkipBlank(src, 6)
	require.Equal(t, 7, ret)
	ret = SkipBlank(src, 7)
	require.Equal(t, 7, ret)
	ret = SkipBlank(src, 8)
	require.Equal(t, -1, ret)
	ret = SkipBlank(src, 9)
	require.Equal(t, -1, ret)
	// ret = skipSpace(src, -1)
	// require.Equal(t, -1, ret)
}

func TestSkipNumber(t *testing.T) {
	type test struct {
		src string
		pos int
		ret int
	}
	var tests = []test{
		{"123,", 0, 3},
		{"-123,", 0, 4},
		{"123.456,", 0, 7},
		{"123.456E789,", 0, 11},
		{"123.456E+789,", 0, 12},
		{"123.456E-789,", 0, 12},
		{"-", 0, -1},
		{"+", 0, -2},
		{".1", 0, -2},
		{"1.", 0, -1},
		{"1.e", 0, -2},
		{"1.1.", 0, -2},
		{"e", 0, -2},
		{"1e1e", 0, -2},
		{"1e", 0, -1},
		{"1+", 0, -2},
		{"1e+", 0, -1},
	}
	for _, test := range tests {
		t.Log(test.src)
		ret := skipNumber(test.src, test.pos)
		require.Equal(t, test.ret, ret)
	}
}

func TestDecodeString(t *testing.T) {
	var src = `"\u666f\\1\\\"\bc\"d"`
	exp, err := strconv.Unquote(src)
	require.NoError(t, err)
	ret, v := decodeString((src), 0)
	require.Equal(t, len(src), ret)
	require.Equal(t, exp, (v))

	src = `"abcd"`
	exp, err = strconv.Unquote(src)
	require.NoError(t, err)
	ret, v = decodeString((src), 0)
	require.Equal(t, len(src), ret)
	require.Equal(t, exp, (v))

	src = `"abcd",`
	ret, v = decodeString((src), 0)
	require.Equal(t, len(src)-1, ret)
	require.Equal(t, exp, (v))
}

func TestDecodeBinary(t *testing.T) {
	var src = `"\naGVsbG8s\nIHdvcmxk\r"`
	var exp []byte
	require.NoError(t, json.Unmarshal([]byte(src), &exp))
	ret, v := DecodeBinary((src), 0)
	require.Equal(t, len(src), ret)
	require.Equal(t, exp, (v))

	src = `"\naGVsbG8s\nIHdvcmxk\r",`
	ret, v = DecodeBinary((src), 0)
	require.Equal(t, len(src)-1, ret)
	require.Equal(t, exp, (v))
}

func TestDecodeInt64(t *testing.T) {
	var src = "-12345678000009999999999"
	ret, v, e := decodeInt64(src, 0)
	require.Equal(t, len(src), ret)
	require.Equal(t, int64(0), v)
	require.NotNil(t, e)

	src = "-123456789"
	ret, v, _ = decodeInt64(src, 0)
	exp, err := strconv.ParseInt(src, 10, 64)
	require.NoError(t, err)
	require.Equal(t, len(src), ret)
	require.Equal(t, exp, v)
}

func TestDecodeFloat64(t *testing.T) {
	var src = "-123.45678E+999999"
	ret, v, e := decodeFloat64(src, 0)
	require.Equal(t, len(src), ret)
	require.Equal(t, float64(0), v)
	require.NotNil(t, e)

	src = "-123.45678E+9"
	exp, err := strconv.ParseFloat(src, 64)
	require.NoError(t, err)
	ret, v, _ = decodeFloat64(src, 0)
	require.Equal(t, len(src), ret)
	require.Equal(t, exp, v)
}

func TestDecodeValue(t *testing.T) {
	type testCase struct {
		src string
		ret int
		exp types.JsonState
		err error
	}
	var cases = []testCase{
		{src: ` null, `, ret: 5, exp: types.JsonState{Vt: types.V_NULL}},
		{src: ` true, `, ret: 5, exp: types.JsonState{Vt: types.V_TRUE}},
		{src: ` false, `, ret: 6, exp: types.JsonState{Vt: types.V_FALSE}},
		{src: ` 123, `, ret: 4, exp: types.JsonState{Vt: types.V_INTEGER, Iv: 123, Ep: 1}},
		{src: ` 123.456, `, ret: 8, exp: types.JsonState{Vt: types.V_DOUBLE, Dv: 123.456, Ep: 1}},
		{src: ` 123.456E+9, `, ret: 11, exp: types.JsonState{Vt: types.V_DOUBLE, Dv: 123.456e+9, Ep: 1}},
		{src: ` "abcd", `, ret: 7, exp: types.JsonState{Vt: types.V_STRING, Iv: 2, Ep: -1}},
		{src: ` "ab\"cd", `, ret: 9, exp: types.JsonState{Vt: types.V_STRING, Iv: 2, Ep: 4}},
		{src: ` "aGVsbG8sIHdvcmxk", `, ret: 19, exp: types.JsonState{Vt: types.V_STRING, Iv: 2, Ep: -1}},
		{src: ` "a\nGVsbG8sIHdvcmxk", `, ret: 21, exp: types.JsonState{Vt: types.V_STRING, Iv: 2, Ep: 3}},
		{src: ` [1,2,3], `, ret: 2, exp: types.JsonState{Vt: types.V_ARRAY}},
		{src: ` {"a":1,"b":2}, `, ret: 2, exp: types.JsonState{Vt: types.V_OBJECT}},
	}

	for _, c := range cases {
		t.Run(c.src, func(t *testing.T) {
			ret, v := DecodeValue(c.src, 0)
			if c.err != nil {
				require.Equal(t, c.err.Error(), types.ParsingError(-ret).Error())
			}
			require.Equal(t, c.ret, ret)
			require.Equal(t, c.exp, v)
		})
	}
}

func TestSkipValue(t *testing.T) {
	type testCase struct {
		src   string
		ret   int
		start int
	}
	var cases = []testCase{
		{src: ` null, `, ret: 5, start: 1},
		{src: ` true, `, ret: 5, start: 1},
		{src: ` false, `, ret: 6, start: 1},
		{src: ` 123, `, ret: 4, start: 1},
		{src: ` 123.456, `, ret: 8, start: 1},
		{src: ` 123.456E+9, `, ret: 11, start: 1},
		{src: ` "abcd", `, ret: 7, start: 1},
		{src: ` "ab\"cd", `, ret: 9, start: 1},
		{src: ` "aGVsbG8sIHdvcmxk", `, ret: 19, start: 1},
		{src: ` "a\nGVsbG8sIHdvcmxk", `, ret: 21, start: 1},
		{src: ` [1,2,3], `, ret: 8, start: 1},
		{src: ` {"a":1,"b":2}, `, ret: 14, start: 1},
	}

	for _, c := range cases {
		t.Run(c.src, func(t *testing.T) {
			ret, start := SkipValue(c.src, 0)
			require.Equal(t, c.ret, ret)
			require.Equal(t, c.start, start)
		})
	}
}

func TestSkipPair(t *testing.T) {
	var src = ` [1,"[", 3,"\" \\[", [[ ],{} ] ] `
	ret := skipPair(src, 1, '[', ']')
	require.Equal(t, len(src)-1, ret)

	src = ` {1,"{", 3,"\" \\}", [[ ],{} ] } `
	ret = skipPair(src, 1, '{', '}')
	require.Equal(t, len(src)-1, ret)
}

func BenchmarkSkipValue(b *testing.B) {
	var src = ` [ 1, "[", 3, "\" \\[", [ [ "", { } ], { "a": [ ] } ] ] `
	b.Run("dyanmicgo", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = SkipValue(src, 0)
		}
	})
	b.Run("native", func(b *testing.B) {
		var fsm = types.NewStateMachine()
		var p int
		for i := 0; i < b.N; i++ {
			fsm.Sp = 0
			p = 0
			_ = native.SkipOne(&src, &p, fsm)
		}
	})
}

func BenchmarkSkipNumber(b *testing.B) {
	var src = "123.456E789,"
	b.Run("skipNumber", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			skipNumber(src, 0)
		}
	})
	b.Run("strconv", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			strconv.ParseFloat(src, 64)
		}
	})
}

func BenchmarkDecodeBool(b *testing.B) {
	b.Run("dyanmicgo", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = DecodeBool("false", 0)
		}
	})
	b.Run("std", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = strconv.ParseBool("false")
		}
	})
}

func BenchmarkSkipSpace(b *testing.B) {
	b.Run("4 blanks", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = SkipBlank("    b", 0)
		}
	})
	b.Run("1 blanks", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = SkipBlank(" b", 0)
		}
	})
	b.Run("0 blanks", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = SkipBlank("b", 0)
		}
	})
}

func BenchmarkDecodeString(b *testing.B) {
	b.Run("dyanmicgo", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = decodeString(`"abcd"`, 0)
		}
	})
	b.Run("std", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = strconv.Unquote(`"abcd"`)
		}
	})
}

func BenchmarkDecodeBinary(b *testing.B) {
	b.Run("dyanmicgo", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = DecodeBinary(`"aGVsbG8sIHdvcmxk"`, 0)
		}
	})
	b.Run("std", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			vv, _ := strconv.Unquote(`"aGVsbG8sIHdvcmxk"`)
			_, _ = base64.StdEncoding.DecodeString(vv)
		}
	})
}
