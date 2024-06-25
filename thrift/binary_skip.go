/**
* Copyright 2024 cloudwego Inc.
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
	"encoding/binary"
	"io"
)

const MaxSkipDepth = 1023

var typeSize = [256]int{
	STOP:   -1,
	VOID:   -1,
	BOOL:   1,
	I08:    1,
	I16:    2,
	I32:    4,
	I64:    8,
	DOUBLE: 8,
	STRING: -1,
	STRUCT: -1,
	MAP:    -1,
	SET:    -1,
	LIST:   -1,
	UTF8:   -1,
	UTF16:  -1,
}

// TypeSize returns the size of the given type.
// -1 means variable size (LIST, SET, MAP, STRING)
// 0 means unknown type
func TypeSize(t Type) int {
	return typeSize[t]
}

// Skip skips over the value for the given type.
func (p *BinaryProtocol) Skip(fieldType Type, useNative bool) (err error) {
	if false {
		// TODO: rm the skip native in the future, no longer use.
		// The pure go implementation is better than c
		return p.SkipNative(fieldType, MaxSkipDepth)
	}
	return p.SkipGo(fieldType, MaxSkipDepth)
}

// next_nopanic returns the same as `next` without panic check, coz `n` is always const value.
func (p *BinaryProtocol) next_nopanic(n int) ([]byte, error) {
	readn := p.Read + n
	if readn > len(p.Buf) {
		return nil, io.EOF
	}
	ret := p.Buf[p.Read:readn]
	p.Read = readn
	return ret, nil
}

// skipn updates p.Read like `next_nopanic` without returning the buf
func (p *BinaryProtocol) skipn(n int) error {
	readn := p.Read + n
	if readn > len(p.Buf) {
		return io.EOF
	}
	p.Read = readn
	return nil
}

// XXX: only used in Skip, this func is small enough to inline
func (p *BinaryProtocol) skipstr() error {
	readn := p.Read + 4
	if readn > len(p.Buf) {
		return io.EOF
	}
	sz := int(binary.BigEndian.Uint32(p.Buf[p.Read:]))
	if sz < 0 {
		return errInvalidDataSize
	}
	readn += sz
	if readn > len(p.Buf) {
		return io.EOF
	}
	p.Read = readn
	return nil
}

// SkipGo skips over the value for the given type using Go implementation.
func (p *BinaryProtocol) SkipGo(fieldType Type, maxDepth int) error {
	if maxDepth <= 0 {
		return errExceedDepthLimit
	}
	if n := typeSize[fieldType]; n > 0 {
		return p.skipn(n)
	}
	switch fieldType {
	case STRING:
		return p.skipstr()
	case STRUCT:
		for {
			b, err := p.next_nopanic(1) // Type
			if err != nil {
				return err
			}
			tp := Type(b[0])
			if tp == STOP {
				break
			}
			if err := p.skipn(2); err != nil { // Field ID
				return err
			}
			if n := typeSize[tp]; n > 0 {
				err = p.skipn(n)
			} else {
				err = p.SkipGo(tp, maxDepth-1)
			}
			if err != nil {
				return err
			}
		}
		return nil
	case MAP:
		b, err := p.next_nopanic(6) // 1 byte key TType, 1 byte value TType, 4 bytes Len
		if err != nil {
			return err
		}
		kt, vt, sz := Type(b[0]), Type(b[1]), int32(binary.BigEndian.Uint32(b[2:]))
		if sz < 0 {
			return errInvalidDataSize
		}
		ksz, vsz := typeSize[kt], typeSize[vt]
		if ksz > 0 && vsz > 0 {
			return p.skipn(int(sz) * (ksz + vsz))
		}
		for i := int32(0); i < sz; i++ {
			if ksz > 0 {
				err = p.skipn(ksz)
			} else if kt == STRING {
				err = p.skipstr()
			} else {
				err = p.SkipGo(kt, maxDepth-1)
			}
			if err != nil {
				return err
			}
			if vsz > 0 {
				err = p.skipn(vsz)
			} else if vt == STRING {
				err = p.skipstr()
			} else {
				err = p.SkipGo(vt, maxDepth-1)
			}
			if err != nil {
				return err
			}
		}
		return nil
	case SET, LIST:
		b, err := p.next_nopanic(5) // 1 byte value type, 4 bytes Len
		if err != nil {
			return err
		}
		vt, sz := Type(b[0]), int32(binary.BigEndian.Uint32(b[1:]))
		if sz < 0 {
			return errInvalidDataSize
		}
		if typeSize[vt] > 0 {
			return p.skipn(int(sz) * typeSize[vt])
		}
		for i := int32(0); i < sz; i++ {
			if vt == STRING {
				err = p.skipstr()
			} else {
				err = p.SkipGo(vt, maxDepth-1)
			}
			if err != nil {
				return err
			}
		}
		return nil
	default:
		return errInvalidDataSize
	}
}
