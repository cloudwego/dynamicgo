//go:build amd64 && go1.16
// +build amd64,go1.16

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
	"testing"

	"github.com/bytedance/sonic/encoder"
	"github.com/cloudwego/dynamicgo/internal/native"
	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/stretchr/testify/assert"
)

func TestSortNodeTwitter(t *testing.T) {
	root, err := NewSearcher(_TwitterJson).GetByPath()
	if err != nil {
		t.Fatal(err)
	}
	obj, err := root.MapUseNumber()
	if err != nil {
		t.Fatal(err)
	}
	exp, err := encoder.Encode(obj, encoder.SortMapKeys)
	if err != nil {
		t.Fatal(err)
	}
	if err := root.SortKeys(true); err != nil {
		t.Fatal(err)
	}
	act, err := root.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(exp), len(act))
	assert.Equal(t, string(exp), string(act))
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