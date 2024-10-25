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
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
)

// RequiresBitmap is a bitmap to mark fields
type RequiresBitmap []uint64

const (
	defaultMaxFieldID = 256
	int64BitSize      = 64
	int64ByteSize     = 8
)

var bitmapPool = sync.Pool{
	New: func() interface{} {
		ret := RequiresBitmap(make([]uint64, 0, defaultMaxFieldID/int64BitSize+1))
		return &ret
	},
}

// Set mark the bit corresponding the given id, with the given requireness
//   - RequiredRequireness|DefaultRequireness mark the bit as 1
//   - OptionalRequireness mark the bit as 0
func (b *RequiresBitmap) Set(id FieldID, val Requireness) {
	i := int(id) / int64BitSize
	if len(*b) <= i {
		b.malloc(int32(id))
	}
	p := unsafe.Pointer(uintptr((*rt.GoSlice)(unsafe.Pointer(b)).Ptr) + uintptr(i)*int64ByteSize)
	switch val {
	case RequiredRequireness, DefaultRequireness:
		*(*uint64)(p) |= (0b1 << (id % int64BitSize))
	case OptionalRequireness:
		*(*uint64)(p) &= ^(0b1 << (id % int64BitSize))
	default:
		panic("invalid requireness")
	}
}

// IsSet tells if the bit corresponding the given id is marked
func (b RequiresBitmap) IsSet(id FieldID) bool {
	i := int(id) / int64BitSize
	if i >= len(b) {
		panic("bitmap id out of range")
	}
	return (b[i] & (0b1 << (id % int64BitSize))) != 0
}

func (b *RequiresBitmap) malloc(id int32) {
	if n := int32(id / int64BitSize); int(n) >= len(*b) {
		buf := make([]uint64, n+1, int32((n+1)*2))
		copy(buf, *b)
		*b = buf
	}
}

// CopyTo copy the bitmap to a given bitmap
func (b RequiresBitmap) CopyTo(to *RequiresBitmap) {
	c := cap(*to)
	l := len(b)
	if l > c {
		*to = make([]uint64, l)
	}
	*to = (*to)[:l]
	copy(*to, b)
}

// NewRequiresBitmap get bitmap from pool, if pool is empty, create a new one
// WARN: memory from pool maybe dirty!
func NewRequiresBitmap() *RequiresBitmap {
	return bitmapPool.Get().(*RequiresBitmap)
}

// FreeRequiresBitmap free the bitmap, but not clear its memory
func FreeRequiresBitmap(b *RequiresBitmap) {
	// memclrNoHeapPointers(*(*unsafe.Pointer)(unsafe.Pointer(b)), uintptr(len(*b))*uint64TypeSize)
	*b = (*b)[:0]
	bitmapPool.Put(b)
}

// CheckRequires scan every bit of the bitmap. When a bit is marked, it will:
//   - if the corresponding field is required-requireness, it reports error
//   - if the corresponding is not required-requireness but writeDefault is true, it will call handler to handle this field
//
//go:nocheckptr
func (b RequiresBitmap) CheckRequires(desc *StructDescriptor, writeDefault bool, handler func(field *FieldDescriptor) error) error {
	// handle bitmap first
	n := len(b)
	s := (*rt.GoSlice)(unsafe.Pointer(&b)).Ptr

	// test 64 bits once
	for i := 0; i < n; i++ {
		v := *(*uint64)(s)
		for j := 0; v != 0 && j < int64BitSize; j++ {
			if v%2 == 1 {
				id := FieldID(i*int64BitSize + j)
				f := desc.FieldById(id)
				if f == nil {
					return errInvalidBitmapId(id, desc)
				}
				if f.Required() == RequiredRequireness {
					return errMissRequiredField(f, desc)
				} else if !writeDefault {
					v >>= 1
					continue
				}
				if err := handler(f); err != nil {
					return err
				}
			}
			v >>= 1
		}
		s = unsafe.Pointer(uintptr(s) + int64ByteSize)
	}
	runtime.KeepAlive(s)
	return nil
}

// HandleRequires scan every bit of the bitmap. When a bit is marked, it will:
//   - if the corresponding field is required-requireness and writeRquired is true, it will call handler to handle this field, otherwise report error
//   - if the corresponding is default-requireness and writeDefault is true, it will call handler to handle this field
//   - if the corresponding is optional-requireness and writeOptional is true, it will call handler to handle this field
//
//go:nocheckptr
func (b RequiresBitmap) HandleRequires(desc *StructDescriptor, writeRequired bool, writeDefault bool, writeOptional bool, handler func(field *FieldDescriptor) error) error {
	// handle bitmap first
	n := len(b)
	s := (*rt.GoSlice)(unsafe.Pointer(&b)).Ptr
	// test 64 bits once
	for i := 0; i < n; i++ {
		v := *(*uint64)(s)
		for j := 0; v != 0 && j < int64BitSize; j++ {
			if v%2 == 1 {
				f := desc.FieldById(FieldID(i*int64BitSize + j))
				if f.Required() == RequiredRequireness && !writeRequired {
					return errMissRequiredField(f, desc)
				}
				if (f.Required() == DefaultRequireness && !writeDefault) || (f.Required() == OptionalRequireness && !writeOptional && f.DefaultValue() == nil) {
					v >>= 1
					continue
				}
				if err := handler(f); err != nil {
					return err
				}
			}
			v >>= 1
		}
		s = unsafe.Pointer(uintptr(s) + int64ByteSize)
	}
	runtime.KeepAlive(s)
	return nil
}

func errMissRequiredField(field *FieldDescriptor, st *StructDescriptor) error {
	return meta.NewError(meta.ErrMissRequiredField, fmt.Sprintf("miss required field '%s' of struct '%s'", field.Name(), st.Name()), nil)
}

func errInvalidBitmapId(id FieldID, st *StructDescriptor) error {
	return meta.NewError(meta.ErrInvalidParam, fmt.Sprintf("invalid field id %d of struct '%s'", id, st.Name()), nil)
}

// DefaultValue is the default value of a field
type DefaultValue struct {
	goValue      interface{}
	jsonValue    string
	thriftBinary string
}

// GoValue return the go runtime representation of the default value
func (d DefaultValue) GoValue() interface{} {
	return d.goValue
}

// JSONValue return the json-encoded representation of the default value
func (d DefaultValue) JSONValue() string {
	return d.jsonValue
}

// ThriftBinary return the thrift-binary-encoded representation of the default value
func (d DefaultValue) ThriftBinary() string {
	return d.thriftBinary
}
