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

package caching

import (
	"math/rand"
	"testing"
	"time"
	"unsafe"
)

var (
	charTable       = []byte(`._-0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`)
	charTableLength = len(charTable)
)

func getSample(minLength, maxLength, count int) (ret []string) {
	rand.Seed(time.Now().Unix())
	for i := 0; i < count; i++ {
		l := rand.Intn(maxLength + 1)
		if l < minLength {
			l = minLength
		}
		buf := make([]byte, l)
		for j := 0; j < l; j++ {
			buf[j] = charTable[rand.Intn(charTableLength)]
		}
		ret = append(ret, string(buf))
	}
	return
}

var (
	maxLength = 20
	minLength = 1
	count     = 20
	samples   = getSample(minLength, maxLength, count)
)

func TestStdMapGet(t *testing.T) {
	var i int
	var p = unsafe.Pointer(&i)
	m := map[string]unsafe.Pointer{}
	for _, v := range samples {
		m[v] = p
	}
	for _, v := range samples {
		if m[v] != p {
			t.Fatal(v)
		}
	}
}

func TestHashMapGet(t *testing.T) {
	var i int
	var p = unsafe.Pointer(&i)
	m := NewHashMap(count, 4)
	for _, v := range samples {
		m.Set(v, p)
	}
	m.Set("", p)
	m.Set("", p)
	if m.Size() != len(samples)+2 {
		t.Fatal(m.Size())
	}
	for _, v := range samples {
		if m.Get(v) != p {
			t.Fatal(v)
		}
	}
}

func TestTrieTreeGet(t *testing.T) {
	var i int
	var p = unsafe.Pointer(&i)
	m := TrieTree{
		Positions: []int{10},
		TrieNode: TrieNode{
			Leaves: &[]Pair{},
		},
	}
	for _, v := range samples {
		m.Set(v, p)
	}
	m.Set("", p)
	m.Set("", p)
	// if m.Size() != len(samples)+1 {
	// 	t.Fatal(m.Size())
	// }
	for _, v := range samples {
		if m.Get(v) != p {
			t.Fatal(v)
		}
	}
}

func BenchmarkStdMapGet(b *testing.B) {
	var i int
	var p = unsafe.Pointer(&i)
	m := map[string]unsafe.Pointer{}
	for _, v := range samples {
		m[v] = p
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range samples {
			_ = m[v]
		}
	}
}

func BenchmarkHashMapGet(b *testing.B) {
	var i int
	var p = unsafe.Pointer(&i)
	m := NewHashMap(count, 2)
	for _, v := range samples {
		m.Set(v, p)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range samples {
			_ = m.Get(v)
		}
	}
}

func BenchmarkTrieTreeGet(b *testing.B) {
	var i int
	var p = unsafe.Pointer(&i)
	m := TrieTree{
		Positions: []int{10},
	}
	for _, v := range samples {
		m.Set(v, p)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range samples {
			_ = m.Get(v)
		}
	}
}

func BenchmarkStdMapGet_Parallel(b *testing.B) {
	var i int
	var p = unsafe.Pointer(&i)
	m := map[string]unsafe.Pointer{}
	for _, v := range samples {
		m[v] = p
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, v := range samples {
				_ = m[v]
			}
		}
	})
}

func BenchmarkHashMapGet_Parallel(b *testing.B) {
	var i int
	var p = unsafe.Pointer(&i)
	m := NewHashMap(count, 2)
	for _, v := range samples {
		m.Set(v, p)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, v := range samples {
				_ = m.Get(v)
			}
		}
	})
}

func BenchmarkTrieTreeGet_Parallel(b *testing.B) {
	var i int
	var p = unsafe.Pointer(&i)
	m := TrieTree{
		Positions: []int{10},
	}
	for _, v := range samples {
		m.Set(v, p)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, v := range samples {
				_ = m.Get(v)
			}
		}
	})
}
