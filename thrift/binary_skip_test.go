/**
 * Copyright 2023 ByteDance Inc.
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

package thrift

import (
	"strconv"
	"testing"

	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/skip"
	"github.com/stretchr/testify/require"
)

func TestSkip(t *testing.T) {
	r := require.New(t)

	obj := skip.NewTestListMap()
	obj.ListMap = make([]map[string]string, 10)
	for i := 0; i < 10; i++ {
		obj.ListMap[i] = map[string]string{strconv.Itoa(i): strconv.Itoa(i)}
	}
	data := make([]byte, obj.BLength())
	_ = obj.FastWriteNocopy(data, nil)

	p := BinaryProtocol{
		Buf: data,
	}
	e1 := p.SkipType(STRUCT)
	r.NoError(e1)
	r.Equal(len(data), p.Read)
}

func BenchmarkSkipNoCheck(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()

	p := NewBinaryProtocol(data)
	err := p.SkipType(desc.Type())
	require.Nil(b, err)
	require.Equal(b, len(data), p.Read)

	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Read = 0
		_ = p.SkipType(desc.Type())
	}
}

func BenchmarkSkipAndCheck(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()

	p := NewBinaryProtocol(data)
	err := p.SkipType(desc.Type())
	require.Nil(b, err)
	require.Equal(b, len(data), p.Read)

	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Read = 0
		err := p.SkipType(desc.Type())
		require.Nil(b, err)
		require.Equal(b, len(data), p.Read)
	}
}
