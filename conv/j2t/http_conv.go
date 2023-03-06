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
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
)

// HTTPConv is a converter from http request to thrift message
type HTTPConv struct {
	proto       meta.Encoding
	st          *thrift.TypeDescriptor
	top, bottom []byte
}

// NewHTTPConv returns a new HTTPConv, which contains the thrift message header and footer
//
// proto is specified thrift encoding protocol (meta.EncodingThriftBinary|meta.EncodingThriftCompact)
// fnDesc is the thrift method descriptor corresponding to the http request url
func NewHTTPConv(proto meta.Encoding, fnDesc *thrift.FunctionDescriptor) *HTTPConv {
	firstField := fnDesc.Request().Struct().Fields()[0] // TODO: some idl may have more than one request
	if firstField.Type().Type() != thrift.STRUCT {
		panic("first request field doesn't have struct kind")
	}

	if proto != meta.EncodingThriftBinary {
		panic("now only support binary protocol")
	}

	header, footer, err := thrift.GetBinaryMessageHeaderAndFooter(fnDesc.Name(), thrift.CALL, firstField.ID(), 0)
	if err != nil {
		panic(err)
	}

	return &HTTPConv{
		proto: proto,
		// if here is not exist, should panic, we don't have better solution
		st:     firstField.Type(),
		top:    header,
		bottom: footer,
	}
}

// Do converts http request into thrift message.
// req body must be one of following:
//   - json (application/json)
//   - url encoded form (application/x-www-form-urlencoded)
//   - empty
func (h HTTPConv) Do(ctx context.Context, req http.RequestGetter, opt conv.Options) (tbytes []byte, err error) {
	if h.proto != meta.EncodingThriftBinary {
		panic("now only support binary protocol")
	}
	cv := NewBinaryConv(opt)
	cv.opts.EnableHttpMapping = true
	// dealing with http request
	jbytes := req.GetBody()
	// manage buffer
	buf := conv.NewBytes()
	// do translation
	err = cv.do(ctx, jbytes, h.st, buf, req)
	if err != nil {
		return nil, err
	}
	// wrap the thrift message header and first fields' message
	topLen := len(h.top)
	bottomLen := len(h.bottom)
	bufLen := len(*buf)
	tbytes = make([]byte, topLen+bufLen+bottomLen)
	copy(tbytes[:topLen], h.top)
	copy(tbytes[topLen:topLen+bufLen], *buf)
	copy(tbytes[topLen+bufLen:], h.bottom)
	// reuse the buf buffer
	conv.FreeBytes(buf)
	return tbytes, nil
}

func (h HTTPConv) DoInto(ctx context.Context, req http.RequestGetter, buf *[]byte, opt conv.Options) (err error) {
	if h.proto != meta.EncodingThriftBinary {
		panic("now only support binary protocol")
	}
	cv := NewBinaryConv(opt)
	cv.opts.EnableHttpMapping = true
	// dealing with http request
	jbytes := req.GetBody()
	// write message header
	*buf = append(*buf, h.top...)
	// do translation
	err = cv.do(ctx, jbytes, h.st, buf, req)
	if err != nil {
		return err
	}
	// write thrift message end and first fields' message
	*buf = append(*buf, h.bottom...)
	return nil
}
