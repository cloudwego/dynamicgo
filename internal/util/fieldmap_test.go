/**
 * Copyright 2024 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"testing"
	"unsafe"
)

func TestFieldMap(t *testing.T) {
	ids := NewFieldNameMap()
	v1 := "a"
	ids.Set("1", unsafe.Pointer(&v1))
	v2 := "b"
	ids.Set("2", unsafe.Pointer(&v2))
	ids.Set("1", unsafe.Pointer(&v2))
	ids.Set("", unsafe.Pointer(&v1))

	ids.Build(false)

	if ids.Get("1") != unsafe.Pointer(&v2) {
		t.Fatalf("expect 1")
	}
	if ids.Get("2") != unsafe.Pointer(&v2) {
		t.Fatalf("expect 1")
	}
	if ids.Get("") != unsafe.Pointer(&v1) {
		t.Fatalf("expect 1")
	}

	ids = NewFieldNameMap()
	ids.Set("", unsafe.Pointer(&v1))
	ids.Set("1", unsafe.Pointer(&v2))
	ids.Build(true)

	if ids.Get("") != unsafe.Pointer(&v1) {
		t.Fatalf("expect 1")
	}
	if ids.Get("1") != unsafe.Pointer(&v2) {
		t.Fatalf("expect 2")
	}

}
