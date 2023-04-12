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
	"fmt"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
)

// HTTPConv is a converter from thrift message to http response
type HTTPConv struct {
	proto meta.Encoding
	st    *thrift.StructDescriptor
}

// NewHTTPConv returns a new HTTPConv
func NewHTTPConv(proto meta.Encoding, desc *thrift.FunctionDescriptor) *HTTPConv {
	switch proto {
	case meta.EncodingThriftBinary:
		// do nothing
	default:
		panic(fmt.Sprintf("protocol %#v is not supported", proto))
	}

	if desc.Response().Struct().FieldById(0) == nil {
		panic("response field is not found in function")
	}

	return &HTTPConv{
		proto: proto,
		st:    desc.Response().Struct(),
	}
}

// Do converts thrift message (tbytes) into http response.
// resp body is set as json protocol if has
func (h HTTPConv) Do(ctx context.Context, resp http.ResponseSetter, tbytes []byte, opt conv.Options) (err error) {
	if h.proto != meta.EncodingThriftBinary {
		panic("now only support binary protocol")
	}
	var stDef *thrift.TypeDescriptor
	// not a reply struct, cannot be decoded
	_, rTyp, _, respID, tbytes, err := thrift.UnwrapBinaryMessage(tbytes)
	if err != nil {
		return wrapError(meta.ErrRead, "", err)
	}
	if rTyp != thrift.REPLY {
		field := h.st.FieldById(respID)
		if field == nil {
			return wrapError(meta.ErrUnknownField, "exception field is not foud in function", nil)
		}
		stDef = field.Type()
	} else {
		if respID != 0 {
			return wrapError(meta.ErrInvalidParam, fmt.Sprintf("unexpected response field id %d", respID), nil)
		}
		stDef = h.st.FieldById(0).Type()
	}

	opt.EnableHttpMapping = true
	cv := NewBinaryConv(opt)
	// manage buffer
	buf := conv.NewBytes()
	// do translation
	err = cv.do(ctx, tbytes, stDef, buf, resp)
	if err != nil {
		return
	}

	body := make([]byte, len(*buf))
	copy(body, *buf)
	resp.SetRawBody(body)

	conv.FreeBytes(buf) // explicit leak by on panic
	return
}

// DoInto converts the thrift message (tbytes) to into buf in JSON protocol, as well as other http response arguments.
// WARN: This will set buf to resp, thus DONOT reuse the buf afterward.
func (h HTTPConv) DoInto(ctx context.Context, resp http.ResponseSetter, tbytes []byte, buf *[]byte, opt conv.Options) (err error) {
	if h.proto != meta.EncodingThriftBinary {
		panic("now only support binary protocol")
	}
	var stDef *thrift.TypeDescriptor
	// not a reply struct, cannot be decoded
	_, rTyp, _, respID, tbytes, err := thrift.UnwrapBinaryMessage(tbytes)
	if err != nil {
		return wrapError(meta.ErrRead, "", err)
	}
	if rTyp != thrift.REPLY {
		field := h.st.FieldById(respID)
		if field == nil {
			return wrapError(meta.ErrUnknownField, "exception field is not foud in function", nil)
		}
		stDef = field.Type()
	} else {
		if respID != 0 {
			return wrapError(meta.ErrInvalidParam, fmt.Sprintf("unexpected response field id %d", respID), nil)
		}
		stDef = h.st.FieldById(0).Type()
	}

	opt.EnableHttpMapping = true
	cv := NewBinaryConv(opt)
	err = cv.do(ctx, tbytes, stDef, buf, resp)
	if err != nil {
		return
	}
	resp.SetRawBody(*buf)
	return
}
