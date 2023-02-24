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

package t2j

import (
	"context"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
	_ "github.com/cloudwego/dynamicgo/thrift/annotation"
)

// BinaryConv is a converter from thrift binary to json
type BinaryConv struct {
	opts conv.Options
}

// NewBinaryConv returns a new BinaryConv
func NewBinaryConv(opts conv.Options) BinaryConv {
	return BinaryConv{opts: opts}
}

// SetOptions sets options
func (self *BinaryConv) SetOptions(opts conv.Options) {
	self.opts = opts
}

// Do converts thrift binary (tbytes) to json bytes (jbytes)
//
// desc is the thrift type descriptor of the thrift binary, usually it is a response STRUCT type
// ctx is the context, which can be used to pass arguments as below:
//     - conv.CtxKeyHTTPResponse: http.ResponseSetter as http request
//     - conv.CtxKeyThriftRespBase: thrift.Base as base metadata of thrift response
func (self *BinaryConv) Do(ctx context.Context, desc *thrift.TypeDescriptor, tbytes []byte) (json []byte, err error) {
	buf := conv.NewBytes()

	var resp http.ResponseSetter
	if self.opts.EnableHttpMapping {
		respi := ctx.Value(conv.CtxKeyHTTPResponse)
		if respi != nil {
			respi, ok := respi.(http.ResponseSetter)
			if !ok {
				return nil, wrapError(meta.ErrInvalidParam, "invalid http.Response", nil)
			}
			resp = respi
		} else {
			return nil, wrapError(meta.ErrInvalidParam, "no http response in context", nil)
		}
	}

	err = self.do(ctx, tbytes, desc, buf, resp)
	if err == nil && len(*buf) > 0 {
		json = make([]byte, len(*buf))
		copy(json, *buf)
	}

	conv.FreeBytes(buf)
	return
}

// DoInto behaves like Do, but it writes the result to buffer directly instead of returning a new buffer
func (self *BinaryConv) DoInto(ctx context.Context, desc *thrift.TypeDescriptor, tbytes []byte, buf *[]byte) (err error) {
	var resp http.ResponseSetter
	if self.opts.EnableHttpMapping {
		respi := ctx.Value(conv.CtxKeyHTTPResponse)
		if respi != nil {
			respi, ok := respi.(http.ResponseSetter)
			if !ok {
				return wrapError(meta.ErrInvalidParam, "invalid http.Response", nil)
			}
			resp = respi
		} else {
			return wrapError(meta.ErrInvalidParam, "no http response in context", nil)
		}
	}

	return self.do(ctx, tbytes, desc, buf, resp)
}
