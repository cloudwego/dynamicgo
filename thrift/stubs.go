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

package thrift

import (
	"reflect"
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/rt"
)

var (
	byteType = rt.UnpackType(reflect.TypeOf(byte(0)))
)

const (
	byteTypeSize   = unsafe.Sizeof(byte(0))
	ptrTypeSize    = unsafe.Sizeof(uintptr(0))
	uint64TypeSize = unsafe.Sizeof(uint64(0))
)

//go:noescape
//go:linkname memmove runtime.memmove
//goland:noinspection GoUnusedParameter
func memmove(to unsafe.Pointer, from unsafe.Pointer, n uintptr)