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

package generic

import (
	"fmt"
	"unsafe"

	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
)

var (
	errNotFound = errNode(meta.ErrNotFound, "", nil)
)

//go:noinline
func wrapError(code meta.ErrCode, msg string, err error) error {
	return meta.NewError(meta.NewErrorCode(code, meta.PROTOBUF), msg, err)
}

//go:noinline
func unwrapError(msg string, err error) error {
	if v, ok := err.(meta.Error); ok {
		return wrapError(v.Code, msg, err)
	} else if v, ok := err.(Value); ok {
		return wrapError(v.ErrCode(), msg, err)
	} else if v, ok := err.(Node); ok {
		return wrapError(v.ErrCode(), msg, err)
	} else {
		return wrapError(0, msg, err)
	}
}

//go:noinline
func wrapValue(n Node, desc *thrift.TypeDescriptor) Value {
	return Value{
		Node: n,
		Desc: desc,
	}
}

//go:noinline
func errNode(code meta.ErrCode, msg string, err error) Node {
	// panic(code.Behavior())
	e := meta.NewError(meta.NewErrorCode(code, meta.THRIFT), msg, err).(meta.Error)
	return Node{
		t: thrift.ERROR,
		l: int(code),
		v: unsafe.Pointer(&e),
	}
}

const (
	lastErrNotFoud thrift.Type = 1
)

//go:noinline
func errNotFoundLast(ptr unsafe.Pointer, parent thrift.Type) Node {
	return Node{
		t:  thrift.ERROR,
		et: lastErrNotFoud,
		kt: parent,
		l:  int(meta.ErrNotFound),
		v:  ptr,
	}
}

//go:noinline
func errValue(code meta.ErrCode, msg string, err error) Value {
	e := meta.NewError(meta.NewErrorCode(code, meta.THRIFT), msg, err).(meta.Error)
	return Value{
		Node: Node{
			t: thrift.ERROR,
			l: int(code),
			v: unsafe.Pointer(&e),
		},
	}
}

//go:noinline
func errPathNode(code meta.ErrCode, msg string, err error) *PathNode {
	// panic(code.Behavior())
	e := meta.NewError(meta.NewErrorCode(code, meta.THRIFT), msg, err).(meta.Error)
	return &PathNode{
			Node: Node{
			t: thrift.ERROR,
			l: int(code),
			v: unsafe.Pointer(&e),
		},
	}
}

// IsEmpty tells if the node is thrift.STOP
func (self Node) IsEmpty() bool {
	return self.t == thrift.STOP
}

// IsEmtpy tells if the node is thrift.ERROR
func (self Node) IsError() bool {
	return self.t == thrift.ERROR
}

// IsErrorNotFound tells if the node is not-found-data error
func (self Node) IsErrNotFound() bool {
	return self.t == thrift.ERROR && self.l == int(meta.ErrNotFound)
}

func (self Node) isErrNotFoundLast() bool {
	return self.IsErrNotFound() && self.et == lastErrNotFoud
}

// ErrCode return the meta.ErrCode of a ERROR node
func (self Node) ErrCode() meta.ErrCode {
	if self.t != thrift.ERROR {
		return 0
	}
	return meta.ErrCode(self.l)
}

// Check checks if it is a ERROR node and returns corresponding error
func (self *Node) Check() error {
	if self == nil {
		return fmt.Errorf("nil node")
	}
	if err := self.Error(); err != "" {
		return self
	}
	return nil
}

// Error return error message if it is a ERROR node
func (self Node) Error() string {
	switch self.t {
	case thrift.ERROR:
		if self.v != nil && self.et != lastErrNotFoud {
			return (*meta.Error)(self.v).Error()
		}
		return fmt.Sprintf("%s", meta.ErrCode(self.l))
	// case thrift.STOP:
	// 	return "error empty node"
	default:
		return ""
	}
}
