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

package j2t

import (
	"context"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
	_ "github.com/cloudwego/dynamicgo/thrift/annotation"
)

// CompactConv is a converter from json to thrift compact
type CompactConv struct {
	opts  conv.Options
	flags uint64
}

// NewCompactConv returns a new CompactConv
func NewCompactConv(opts conv.Options) CompactConv {
	return CompactConv{
		opts:  opts,
		flags: toFlags(opts),
	}
}

// SetOptions sets options
func (self *CompactConv) SetOptions(opts conv.Options) {
	self.opts = opts
	self.flags = toFlags(self.opts)
}

// Do converts json bytes (jbytes) to thrift compact (tbytes)
//
// desc is the thrift type descriptor of the thrift compact, usually it the request STRUCT type
// ctx is the context, which can be used to pass arguments as below:
//   - conv.CtxKeyHTTPRequest: http.RequestGetter as http request
//   - conv.CtxKeyThriftRespBase: thrift.Base as base metadata of thrift response
func (self *CompactConv) Do(ctx context.Context, desc *thrift.TypeDescriptor, jbytes []byte) (tbytes []byte, err error) {
	buf := conv.NewBytes()

	var req http.RequestGetter
	if self.opts.EnableHttpMapping {
		reqi := ctx.Value(conv.CtxKeyHTTPRequest)
		if reqi != nil {
			reqi, ok := reqi.(http.RequestGetter)
			if !ok {
				return nil, newError(meta.ErrInvalidParam, "invalid http.RequestGetter", nil)
			}
			req = reqi
		} else {
			return nil, newError(meta.ErrInvalidParam, "EnableHttpMapping but no http response in context", nil)
		}
	}
	fsm := types.NewJ2TStateMachine()
	err = self.do(ctx, fsm, jbytes, desc, buf, req)
	if err == nil && len(*buf) > 0 {
		tbytes = make([]byte, len(*buf))
		copy(tbytes, *buf)
	}
	types.FreeJ2TStateMachine(fsm)
	conv.FreeBytes(buf)
	return
}

// DoInto behaves like Do, but it writes the result to buffer directly instead of returning a new buffer
func (self *CompactConv) DoInto(ctx context.Context, desc *thrift.TypeDescriptor, jbytes []byte, buf *[]byte) (err error) {
	var req http.RequestGetter
	if self.opts.EnableHttpMapping {
		reqi := ctx.Value(conv.CtxKeyHTTPRequest)
		if reqi != nil {
			reqi, ok := reqi.(http.RequestGetter)
			if !ok {
				return newError(meta.ErrInvalidParam, "invalid http.RequestGetter", nil)
			}
			req = reqi
		} else {
			return newError(meta.ErrInvalidParam, "EnableHttpMapping but no http response in context", nil)
		}
	}
	fsm := types.NewJ2TStateMachine()
	err = self.do(ctx, fsm, jbytes, desc, buf, req)
	types.FreeJ2TStateMachine(fsm)
	return
}
