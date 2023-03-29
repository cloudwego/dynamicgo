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
	"runtime"
	"unsafe"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/json"
	"github.com/cloudwego/dynamicgo/internal/native"
	"github.com/cloudwego/dynamicgo/internal/native/types"
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
			if err := writeRequestBaseToThrift(ctx, buf, f); err != nil {
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
				if err := self.writeHttpRequestToThrift(ctx, req, st, *reqs, buf, true, true); err != nil {
					return err
				}
			}
			// check if any required field exists, if have, traceback on http
			// since this case it always for top-level fields,
			// we should only check opts.BackTraceRequireOrTopField to decide whether to traceback
			err := reqs.HandleRequires(st, self.opts.ReadHttpValueFallback, self.opts.ReadHttpValueFallback, self.opts.ReadHttpValueFallback, func(f *thrift.FieldDescriptor) error {
				val, _ := tryGetValueFromHttp(req, f.Alias())
				if err := self.writeStringValue(ctx, buf, f, val, meta.EncodingJSON, req); err != nil {
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

	return self.doNative(ctx, src, desc, buf, req, true)
}

func (self *BinaryConv) doNative(ctx context.Context, src []byte, desc *thrift.TypeDescriptor, buf *[]byte, req http.RequestGetter, top bool) (err error) {
	jp := rt.Mem2Str(src)
	fsm := types.NewJ2TStateMachine()
	fsm.Init(0, unsafe.Pointer(desc))

exec:
	ret := native.J2T_FSM(fsm, buf, &jp, self.flags)
	if ret != 0 {
		cont, e := self.handleError(ctx, fsm, buf, src, req, ret, top)
		if cont && e == nil {
			goto exec
		}
		err = e
		goto ret
	}

ret:
	types.FreeJ2TStateMachine(fsm)
	runtime.KeepAlive(desc)
	return
}

func isJsonString(val string) bool {
	if len(val) < 2 {
		return false
	}
	s := val[0]
	e := val[len(val)-1]
	return (s == '{' && e == '}') || (s == '[' && e == ']')
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
		} else if enc != meta.EncodingJSON {
			return newError(meta.ErrUnsupportedType, fmt.Sprintf("unsupported encoding type %d of field '%s'", enc, f.Name()), nil)
		}
		if ft := f.Type(); ft.Type().IsComplex() && isJsonString(val) {
			// for nested type, we regard it as a json string and convert it directly
			if err := self.doNative(ctx, rt.Str2Mem(val), ft, &p.Buf, req, false); err != nil {
				return newError(meta.ErrConvert, fmt.Sprintf("failed to convert value of field '%s'", f.Name()), err)
			}
		} else if err := p.WriteStringWithDesc(val, ft, self.opts.DisallowUnknownField, !self.opts.NoBase64Binary); err != nil {
			return newError(meta.ErrConvert, fmt.Sprintf("failed to write field '%s' value", f.Name()), err)
		}
	}
BACK:
	*buf = p.Buf
	return nil
}

func writeRequestBaseToThrift(ctx context.Context, buf *[]byte, field *thrift.FieldDescriptor) error {
	bobj := ctx.Value(conv.CtxKeyThriftReqBase)
	if bobj != nil {
		if b, ok := bobj.(*base.Base); ok && b != nil {
			l := len(*buf)
			n := b.BLength()
			rt.GuardSlice(buf, 3+n)
			*buf = (*buf)[:l+3]
			thrift.BinaryEncoding{}.EncodeFieldBegin((*buf)[l:l+3], field.Type().Type(), field.ID())
			l = len(*buf)
			*buf = (*buf)[:l+n]
			b.FastWrite((*buf)[l : l+n])
		}
	}
	return nil
}

func (self *BinaryConv) writeHttpRequestToThrift(ctx context.Context, req http.RequestGetter, desc *thrift.StructDescriptor, reqs thrift.RequiresBitmap, buf *[]byte, nobody bool, top bool) (err error) {
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

func (self *BinaryConv) handleUnmatchedFields(ctx context.Context, fsm *types.J2TStateMachine, desc *thrift.StructDescriptor, buf *[]byte, pos int, req http.RequestGetter, top bool) (bool, error) {
	if req == nil {
		return false, newError(meta.ErrInvalidParam, "http request is nil", nil)
	}

	// write unmatched fields
	for _, id := range fsm.FieldCache {
		f := desc.FieldById(thrift.FieldID(id))
		if f == nil {
			if self.opts.DisallowUnknownField {
				return false, newError(meta.ErrConvert, fmt.Sprintf("unknown field id %d", id), nil)
			}
			continue
		}
		var val string
		if self.opts.TracebackRequredOrRootFields && (top || f.Required() == thrift.RequiredRequireness) {
			// try get value from http
			val, _ = tryGetValueFromHttp(req, f.Alias())
		}
		if err := self.writeStringValue(ctx, buf, f, val, meta.EncodingJSON, req); err != nil {
			return false, err
		}
	}

	// write STRUCT end
	*buf = append(*buf, byte(thrift.STOP))

	// clear field cache
	fsm.FieldCache = fsm.FieldCache[:0]
	// drop current J2T state
	fsm.SP--
	if fsm.SP > 0 {
		// NOTICE: if j2t_exec haven't finished, we should set current position to next json
		fsm.SetPos(pos)
		return true, nil
	} else {
		return false, nil
	}
}

// searching sequence: url -> [post] -> query -> header -> [body root]
func tryGetValueFromHttp(req http.RequestGetter, key string) (string, bool) {
	if req == nil {
		return "", false
	}
	if v := req.GetParam(key); v != "" {
		return v, true
	}
	if v := req.GetQuery(key); v != "" {
		return v, true
	}
	if v := req.GetHeader(key); v != "" {
		return v, true
	}
	if v := req.GetCookie(key); v != "" {
		return v, true
	}
	if v := req.GetMapBody(key); v != "" {
		return v, true
	}
	return "", false
}

func (self *BinaryConv) handleValueMapping(ctx context.Context, fsm *types.J2TStateMachine, desc *thrift.StructDescriptor, buf *[]byte, pos int, src []byte) (bool, error) {
	p := thrift.BinaryProtocol{}
	p.Buf = *buf

	// write unmatched fields
	fv := fsm.FieldValueCache
	f := desc.FieldById(thrift.FieldID(fv.FieldID))
	if f == nil {
		return false, newError(meta.ErrConvert, fmt.Sprintf("unknown field id %d for value-mapping", fv.FieldID), nil)
	}
	// write field tag
	if err := p.WriteFieldBegin(f.Name(), f.Type().Type(), f.ID()); err != nil {
		return false, newError(meta.ErrWrite, fmt.Sprintf("failed to write field '%s' tag", f.Name()), err)
	}
	// truncate src to value
	if int(fv.ValEnd) >= len(src) || int(fv.ValBegin) < 0 {
		return false, newError(meta.ErrConvert, "invalid value-mapping position", nil)
	}
	jdata := src[fv.ValBegin:fv.ValEnd]
	// call ValueMapping interface
	if err := f.ValueMapping().Write(ctx, &p, f, jdata); err != nil {
		return false, newError(meta.ErrConvert, fmt.Sprintf("failed to convert field '%s' value", f.Name()), err)
	}

	*buf = p.Buf
	// clear field cache
	// fsm.FieldValueCache = fsm.FieldValueCache[:0]

	// drop current J2T state
	fsm.SP--
	if fsm.SP > 0 {
		// NOTICE: if j2t_exec haven't finished, we should set current position to next json
		fsm.SetPos(pos)
		return true, nil
	} else {
		return false, nil
	}
}
