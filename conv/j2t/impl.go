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
	"fmt"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/json"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/cloudwego/dynamicgo/thrift/base"
)

const (
	_GUARD_SLICE_FACTOR = 1
)

func (self *BinaryConv) do(ctx context.Context, src []byte, desc *thrift.TypeDescriptor, buf *[]byte, req http.RequestGetter) error {
	//NOTICE: output buffer must be larger than src buffer
	rt.GuardSlice(buf, len(src)*_GUARD_SLICE_FACTOR)

	if self.opts.EnableThriftBase {
		if f := desc.Struct().GetRequestBase(); f != nil {
			if err := self.writeRequestBaseToThrift(ctx, buf, f); err != nil {
				return err
			}
		}
	}

	if len(src) == 0 {
		// empty body
		if self.opts.EnableHttpMapping && req != nil {
			st := desc.Struct()
			var reqs = thrift.NewRequiresBitmap()
			st.Requires().CopyTo(reqs)
			// check if any http-mapping exists
			if desc.Struct().HttpMappingFields() != nil {
				if err := self.handleHttpMappings(ctx, req, st, *reqs, buf, true, true); err != nil {
					return err
				}
			}
			// check if any required field exists, if have, traceback on http
			// since this case it always for top-level fields,
			// we should only check opts.BackTraceRequireOrTopField to decide whether to traceback
			err := reqs.HandleRequires(st, self.opts.ReadHttpValueFallback, self.opts.ReadHttpValueFallback, self.opts.ReadHttpValueFallback, func(f *thrift.FieldDescriptor) error {
				val, _, enc := tryGetValueFromHttp(req, f.Alias())
				if err := self.writeStringValue(ctx, buf, f, val, enc, req); err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return newError(meta.ErrWrite, "failed to write required field", err)
			}
			thrift.FreeRequiresBitmap(reqs)
		}
		// since there is no json data, we should add struct end into buf and return
		*buf = append(*buf, byte(thrift.STOP))
		return nil
	}

	// special case for unquoted json string
	if desc.Type() == thrift.STRING && src[0] != '"' {
		buf := make([]byte, 0, len(src)+2)
		src = json.EncodeString(buf, rt.Mem2Str(src))
	}

	return self.doImpl(ctx, src, desc, buf, req, true)
}

func isJsonString(val string) bool {
	if len(val) < 2 {
		return false
	}

	c := json.SkipBlank(val, 0)
	if c < 0 {
		return false
	}
	s := val[c]
	e := val[len(val)-1] //FIXME: may need exist blank
	return (s == '{' && e == '}') || (s == '[' && e == ']') || (s == '"' && e == '"')
}

func (self *BinaryConv) writeStringValue(ctx context.Context, buf *[]byte, f *thrift.FieldDescriptor, val string, enc meta.Encoding, req http.RequestGetter) error {
	p := thrift.BinaryProtocol{Buf: *buf}
	if val == "" {
		if !self.opts.WriteRequireField && f.Required() == thrift.RequiredRequireness {
			// requred field not found, return error
			return newError(meta.ErrMissRequiredField, fmt.Sprintf("required field '%s' not found", f.Name()), nil)
		}
		if !self.opts.WriteOptionalField && f.Required() == thrift.OptionalRequireness {
			return nil
		}
		if !self.opts.WriteDefaultField && f.Required() == thrift.DefaultRequireness {
			return nil
		}
		if err := p.WriteFieldBegin(f.Name(), f.Type().Type(), f.ID()); err != nil {
			return newError(meta.ErrWrite, fmt.Sprintf("failed to write field '%s' tag", f.Name()), err)
		}
		if err := p.WriteDefaultOrEmpty(f); err != nil {
			return newError(meta.ErrWrite, fmt.Sprintf("failed to write empty value of field '%s'", f.Name()), err)
		}
	} else {
		if err := p.WriteFieldBegin(f.Name(), f.Type().Type(), f.ID()); err != nil {
			return newError(meta.ErrWrite, fmt.Sprintf("failed to write field '%s' tag", f.Name()), err)
		}
		// not http-encoded value, write directly into buf
		if enc == meta.EncodingThriftBinary {
			p.Buf = append(p.Buf, val...)
			goto BACK
		} else if enc == meta.EncodingText || !f.Type().Type().IsComplex() || !isJsonString(val) {
			if err := p.WriteStringWithDesc(val, f.Type(), self.opts.DisallowUnknownField, !self.opts.NoBase64Binary); err != nil {
				return newError(meta.ErrConvert, fmt.Sprintf("failed to write field '%s' value", f.Name()), err)
			}
		} else if enc == meta.EncodingJSON {
			// for nested type, we regard it as a json string and convert it directly
			if err := self.doImpl(ctx, rt.Str2Mem(val), f.Type(), &p.Buf, req, false); err != nil {
				return newError(meta.ErrConvert, fmt.Sprintf("failed to convert value of field '%s'", f.Name()), err)
			}
			// try text encoding, see thrift.EncodeText
		} else {
			return newError(meta.ErrConvert, fmt.Sprintf("unsupported http-mapping encoding %v for '%s'", enc, f.Name()), nil)
		}
	}
BACK:
	*buf = p.Buf
	return nil
}

func (self *BinaryConv) writeRequestBaseToThrift(ctx context.Context, buf *[]byte, field *thrift.FieldDescriptor) error {
	var b *base.Base
	if bobj := ctx.Value(conv.CtxKeyThriftReqBase); bobj != nil {
		if v, ok := bobj.(*base.Base); ok && v != nil {
			b = v
		}
	}
	if b == nil && self.opts.WriteRequireField && field.Required() == thrift.RequiredRequireness {
		b = &base.Base{}
	}
	if b != nil {
		l := len(*buf)
		n := b.BLength()
		rt.GuardSlice(buf, 3+n)
		*buf = (*buf)[:l+3]
		thrift.BinaryEncoding{}.EncodeFieldBegin((*buf)[l:l+3], field.Type().Type(), field.ID())
		l = len(*buf)
		*buf = (*buf)[:l+n]
		b.FastWrite((*buf)[l : l+n])
	}
	return nil
}

// searching sequence: url -> [post] -> query -> header -> [body root]
func tryGetValueFromHttp(req http.RequestGetter, key string) (string, bool, meta.Encoding) {
	if req == nil {
		return "", false, meta.EncodingText
	}
	if v := req.GetParam(key); v != "" {
		return v, true, meta.EncodingJSON
	}
	if v := req.GetQuery(key); v != "" {
		return v, true, meta.EncodingJSON
	}
	if v := req.GetHeader(key); v != "" {
		return v, true, meta.EncodingJSON
	}
	if v := req.GetCookie(key); v != "" {
		return v, true, meta.EncodingJSON
	}
	if v := req.GetMapBody(key); v != "" {
		return v, true, meta.EncodingJSON
	}
	return "", false, meta.EncodingText
}

func (self *BinaryConv) tryWriteValueRecursive(ctx context.Context, req http.RequestGetter, field *thrift.FieldDescriptor, p *thrift.BinaryProtocol, tryHttp bool) error {
	typ := field.Type()
	if typ.Type() == thrift.STRUCT && len(field.HTTPMappings()) == 0 {
		// recursively http-mapping struct fields, if it has no http-mapping
		for _, f := range typ.Struct().Fields() {
			if err := self.tryWriteValueRecursive(ctx, req, f, p, tryHttp); err != nil {
				return newError(meta.ErrWrite, fmt.Sprintf("failed to write field '%s' value", f.Name()), err)
			}
		}
		if err := p.WriteStructEnd(); err != nil {
			return newError(meta.ErrWrite, fmt.Sprintf("failed to write struct end of field '%s'", field.Name()), err)
		}
	} else {
		var val string
		var enc = meta.EncodingText
		if tryHttp {
			val, _, enc = tryGetValueFromHttp(req, field.Alias())
		}
		if err := self.writeStringValue(ctx, &p.Buf, field, val, enc, req); err != nil {
			return newError(meta.ErrWrite, fmt.Sprintf("failed to write field '%s' http value", field.Name()), err)
		}
	}
	return nil
}

func (self *BinaryConv) handleHttpMappings(ctx context.Context, req http.RequestGetter, desc *thrift.StructDescriptor, reqs thrift.RequiresBitmap, buf *[]byte, nobody bool, top bool) (err error) {
	if req == nil {
		return newError(meta.ErrInvalidParam, "http request is nil", nil)
	}
	fs := desc.HttpMappingFields()
	for _, f := range fs {
		var ok bool
		var val string
		var httpEnc meta.Encoding
		// loop http mapping until first non-null value
		for _, hm := range f.HTTPMappings() {
			v, err := hm.Request(ctx, req, f)
			if err == nil {
				httpEnc = hm.Encoding()
				ok = true
				val = v
				break
			}
		}
		if !ok {
			// no json body, check if return error
			if nobody {
				if f.Required() == thrift.RequiredRequireness && !self.opts.WriteRequireField {
					return newError(meta.ErrNotFound, fmt.Sprintf("not found http value of field %d:'%s'", f.ID(), f.Name()), nil)
				}
				if !self.opts.WriteDefaultField && f.Required() == thrift.DefaultRequireness {
					continue
				}
				if !self.opts.WriteOptionalField && f.Required() == thrift.OptionalRequireness {
					continue
				}
			} else {
				// NOTICE: if no value found, tracebak on current json layeer to find value
				// it must be a top level field or required field
				if self.opts.ReadHttpValueFallback {
					reqs.Set(f.ID(), thrift.RequiredRequireness)
					continue
				}
			}
		}

		reqs.Set(f.ID(), thrift.OptionalRequireness)
		if err := self.writeStringValue(ctx, buf, f, val, httpEnc, req); err != nil {
			return err
		}
	}

	// p.Recycle()
	return
}
