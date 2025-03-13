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
	"errors"
	"fmt"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/json"
	"github.com/cloudwego/dynamicgo/internal/primitive"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/cloudwego/dynamicgo/thrift/base"
)

const (
	_GUARD_SLICE_FACTOR = 2
)

//go:noinline
func wrapError(code meta.ErrCode, msg string, err error) error {
	// panic(meta.NewError(meta.NewErrorCode(code, meta.THRIFT2JSON), msg, err))
	return meta.NewError(meta.NewErrorCode(code, meta.THRIFT2JSON), msg, err)
}

//go:noinline
func unwrapError(msg string, err error) error {
	if v, ok := err.(meta.Error); ok {
		return wrapError(v.Code, msg, err)
	} else {
		return wrapError(meta.ErrConvert, msg, err)
	}
}

func (self *BinaryConv) readResponseBase(ctx context.Context, p *thrift.BinaryProtocol) (bool, error) {
	obj := ctx.Value(conv.CtxKeyThriftRespBase)
	if obj == nil {
		return false, nil
	}
	base, ok := obj.(*base.BaseResp)
	if !ok || base == nil {
		return false, wrapError(meta.ErrInvalidParam, "invalid response base", nil)
	}
	s := p.Read
	if err := p.Skip(thrift.STRUCT, self.opts.UseNativeSkip); err != nil {
		return false, wrapError(meta.ErrRead, "", err)
	}
	e := p.Read
	if _, err := base.FastRead(p.Buf[s:e]); err != nil {
		return false, wrapError(meta.ErrRead, "", err)
	}
	return true, nil
}

func (self *BinaryConv) do(ctx context.Context, src []byte, desc *thrift.TypeDescriptor, out *[]byte, resp http.ResponseSetter) (err error) {
	//NOTICE: output buffer must be larger than src buffer
	rt.GuardSlice(out, len(src)*_GUARD_SLICE_FACTOR)

	var p = thrift.BinaryProtocol{
		Buf: src,
	}

	if desc.Type() != thrift.STRUCT {
		return self.doRecurse(ctx, desc, out, resp, &p)
	}

	_, e := p.ReadStructBegin()
	if e != nil {
		return wrapError(meta.ErrRead, "", e)
	}
	*out = json.EncodeObjectBegin(*out)

	r := thrift.NewRequiresBitmap()
	desc.Struct().Requires().CopyTo(r)
	comma := false

	existExceptionField := false

	for {
		_, typeId, id, e := p.ReadFieldBegin()
		if e != nil {
			return wrapError(meta.ErrRead, "", e)
		}
		if typeId == thrift.STOP {
			break
		}

		field := desc.Struct().FieldById(thrift.FieldID(id))
		if field == nil {
			if self.opts.DisallowUnknownField {
				return wrapError(meta.ErrUnknownField, fmt.Sprintf("unknown field %d", id), nil)
			}
			if e := p.Skip(typeId, self.opts.UseNativeSkip); e != nil {
				return wrapError(meta.ErrRead, "", e)
			}
			continue
		}

		r.Set(field.ID(), thrift.OptionalRequireness)

		if self.opts.EnableThriftBase && field.IsResponseBase() {
			skip, err := self.readResponseBase(ctx, &p)
			if err != nil {
				return err
			}
			if skip {
				continue
			}
		}

		restart := p.Read

		if resp != nil && self.opts.EnableHttpMapping && field.HTTPMappings() != nil {
			ok, err := self.writeHttpValue(ctx, resp, &p, field)
			if err != nil {
				return unwrapError(fmt.Sprintf("mapping field %s of STRUCT %s failed", field.Name(), desc.Name()), err)
			}
			// NOTICE: if option HttpMappingAsExtra is false and http-mapping failed,
			// continue to write to json body
			if !self.opts.WriteHttpValueFallback || ok {
				continue
			}
		}

		p.Read = restart
		if comma {
			*out = json.EncodeObjectComma(*out)
		} else {
			comma = true
		}
		// NOTICE: always use field.Alias() here, because alias equals to name by default
		*out = json.EncodeString(*out, field.Alias())
		*out = json.EncodeObjectColon(*out)

		// handle a thrift exception field. the id of a thrift exception field is non-zero
		if self.opts.ConvertException && id != 0 {
			existExceptionField = true
			// reset out to get only exception field data
			*out = (*out)[:0]
		}

		if self.opts.EnableValueMapping && field.ValueMapping() != nil {
			err = field.ValueMapping().Read(ctx, &p, field, out)
			if err != nil {
				return unwrapError(fmt.Sprintf("mapping field %s of STRUCT %s failed", field.Name(), desc.Type()), err)
			}
		} else {
			err = self.doRecurse(ctx, field.Type(), out, resp, &p)
			if err != nil {
				return unwrapError(fmt.Sprintf("converting field %s of STRUCT %s failed", field.Name(), desc.Type()), err)
			}
		}

		if existExceptionField {
			break
		}
	}

	if err = self.handleUnsets(r, desc.Struct(), out, comma, ctx, resp); err != nil {
		return err
	}

	thrift.FreeRequiresBitmap(r)
	if existExceptionField && err == nil {
		err = errors.New(string(*out))
	} else {
		*out = json.EncodeObjectEnd(*out)
	}
	return err
}

func (self *BinaryConv) doRecurse(ctx context.Context, desc *thrift.TypeDescriptor, out *[]byte, resp http.ResponseSetter, p *thrift.BinaryProtocol) (err error) {
	tt := desc.Type()
	switch tt {
	case thrift.BOOL:
		v, e := p.ReadBool()
		if e != nil {
			return wrapError(meta.ErrRead, "", e)
		}
		*out = json.EncodeBool(*out, v)
	case thrift.BYTE:
		v, e := p.ReadByte()
		if e != nil {
			return wrapError(meta.ErrWrite, "", e)
		}
		if self.opts.ByteAsUint8 {
			*out = json.EncodeInt64(*out, int64(uint8(v)))
		} else {
			*out = json.EncodeInt64(*out, int64(int8(v)))
		}
	case thrift.I16:
		v, e := p.ReadI16()
		if e != nil {
			return wrapError(meta.ErrWrite, "", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case thrift.I32:
		v, e := p.ReadI32()
		if e != nil {
			return wrapError(meta.ErrWrite, "", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case thrift.I64:
		v, e := p.ReadI64()
		if e != nil {
			return wrapError(meta.ErrWrite, "", e)
		}
		if self.opts.Int642String {
			*out = append(*out, '"')
			*out = json.EncodeInt64(*out, int64(v))
			*out = append(*out, '"')
		} else {
			*out = json.EncodeInt64(*out, int64(v))
		}
	case thrift.DOUBLE:
		v, e := p.ReadDouble()
		if e != nil {
			return wrapError(meta.ErrWrite, "", e)
		}
		*out = json.EncodeFloat64(*out, float64(v))
	case thrift.STRING:
		if desc.IsBinary() && !self.opts.NoBase64Binary {
			v, e := p.ReadBinary(false)
			if e != nil {
				return wrapError(meta.ErrRead, "", e)
			}
			*out = json.EncodeBaniry(*out, v)
		} else {
			v, e := p.ReadString(false)
			if e != nil {
				return wrapError(meta.ErrRead, "", e)
			}
			*out = json.EncodeString(*out, v)
		}
	case thrift.STRUCT:
		_, e := p.ReadStructBegin()
		if e != nil {
			return wrapError(meta.ErrRead, "", e)
		}
		*out = json.EncodeObjectBegin(*out)

		r := thrift.NewRequiresBitmap()
		desc.Struct().Requires().CopyTo(r)
		comma := false

		for {
			_, typeId, id, e := p.ReadFieldBegin()
			if e != nil {
				return wrapError(meta.ErrRead, "", e)
			}
			if typeId == 0 {
				break
			}

			field := desc.Struct().FieldById(thrift.FieldID(id))
			if field == nil {
				if self.opts.DisallowUnknownField {
					return wrapError(meta.ErrUnknownField, fmt.Sprintf("unknown field %d", id), nil)
				}
				if e := p.Skip(typeId, self.opts.UseNativeSkip); e != nil {
					return wrapError(meta.ErrRead, "", e)
				}
				continue
			}

			r.Set(field.ID(), thrift.OptionalRequireness)
			restart := p.Read

			if resp != nil && self.opts.EnableHttpMapping && field.HTTPMappings() != nil {
				ok, err := self.writeHttpValue(ctx, resp, p, field)
				if err != nil {
					return unwrapError(fmt.Sprintf("mapping field %s of STRUCT %s failed", field.Name(), desc.Name()), err)
				}
				// NOTICE: if option HttpMappingAsExtra is false and http-mapping failed,
				// continue to write to json body
				if !self.opts.WriteHttpValueFallback || ok {
					continue
				}
			}

			// HttpMappingAsExtra is true, return to begining and write json body
			p.Read = restart
			if comma {
				*out = json.EncodeObjectComma(*out)
				if err != nil {
					return wrapError(meta.ErrWrite, "", err)
				}
			} else {
				comma = true
			}
			// NOTICE: always use field.Alias() here, because alias equals to name by default
			*out = json.EncodeString(*out, field.Alias())
			*out = json.EncodeObjectColon(*out)

			if self.opts.EnableValueMapping && field.ValueMapping() != nil {
				if err = field.ValueMapping().Read(ctx, p, field, out); err != nil {
					return unwrapError(fmt.Sprintf("mapping field %s of STRUCT %s failed", field.Name(), desc.Type()), err)
				}
			} else {
				err = self.doRecurse(ctx, field.Type(), out, nil, p)
				if err != nil {
					return unwrapError(fmt.Sprintf("converting field %s of STRUCT %s failed", field.Name(), desc.Type()), err)
				}
			}
		}

		if err = self.handleUnsets(r, desc.Struct(), out, comma, ctx, resp); err != nil {
			return err
		}

		thrift.FreeRequiresBitmap(r)
		*out = json.EncodeObjectEnd(*out)
	case thrift.MAP:
		keyType, valueType, size, e := p.ReadMapBegin()
		if e != nil {
			return wrapError(meta.ErrRead, "", e)
		}
		if keyType != desc.Key().Type() {
			return wrapError(meta.ErrDismatchType, fmt.Sprintf("expect type %s but got type %s", desc.Key().Type(), keyType), nil)
		} else if valueType != desc.Elem().Type() {
			return wrapError(meta.ErrDismatchType, fmt.Sprintf("expect type %s but got type %s", desc.Elem().Type(), valueType), nil)
		}
		*out = json.EncodeObjectBegin(*out)
		for i := 0; i < size; i++ {
			if i != 0 {
				*out = json.EncodeObjectComma(*out)
			}
			err = self.buildinTypeToKey(p, desc.Key(), out)
			if err != nil {
				return wrapError(meta.ErrConvert, "", err)
			}
			*out = json.EncodeObjectColon(*out)
			err = self.doRecurse(ctx, desc.Elem(), out, nil, p)
			if err != nil {
				return unwrapError(fmt.Sprintf("converting %dth element of MAP failed", i), err)
			}
		}
		e = p.ReadMapEnd()
		if e != nil {
			return wrapError(meta.ErrRead, "", e)
		}
		*out = json.EncodeObjectEnd(*out)
	case thrift.SET, thrift.LIST:
		elemType, size, e := p.ReadSetBegin()
		if e != nil {
			return wrapError(meta.ErrRead, "", e)
		}
		if elemType != desc.Elem().Type() {
			return wrapError(meta.ErrDismatchType, fmt.Sprintf("expect type %s but got type %s", desc.Elem().Type(), elemType), nil)
		}
		*out = json.EncodeArrayBegin(*out)
		for i := 0; i < size; i++ {
			if i != 0 {
				*out = json.EncodeArrayComma(*out)
			}
			err = self.doRecurse(ctx, desc.Elem(), out, nil, p)
			if err != nil {
				return unwrapError(fmt.Sprintf("converting %dth element of SET failed", i), err)
			}
		}
		e = p.ReadSetEnd()
		if e != nil {
			return wrapError(meta.ErrRead, "", e)
		}
		*out = json.EncodeArrayEnd(*out)

	default:
		return wrapError(meta.ErrUnsupportedType, fmt.Sprintf("unknown descriptor type %s", tt), nil)
	}

	return
}

func (self *BinaryConv) handleUnsets(b *thrift.RequiresBitmap, desc *thrift.StructDescriptor, out *[]byte, comma bool, ctx context.Context, resp http.ResponseSetter) error {
	return b.HandleRequires(desc, self.opts.WriteRequireField, self.opts.WriteDefaultField, self.opts.WriteOptionalField, func(field *thrift.FieldDescriptor) error {
		// check if field has http mapping
		var ok = false
		if hms := field.HTTPMappings(); self.opts.EnableHttpMapping && hms != nil {
			// make a default thrift value
			p := thrift.BinaryProtocol{Buf: make([]byte, 0, conv.DefaulHttpValueBufferSizeForJSON)}
			if err := p.WriteDefaultOrEmpty(field); err != nil {
				return wrapError(meta.ErrWrite, fmt.Sprintf("encoding field '%s' default value failed", field.Name()), err)
			}
			// convert it into http
			var err error
			ok, err = self.writeHttpValue(ctx, resp, &p, field)
			if err != nil {
				return err
			}
		}
		if ok {
			return nil
		}
		if !comma {
			comma = true
		} else {
			*out = json.EncodeArrayComma(*out)
		}
		*out = json.EncodeString(*out, field.Name())
		*out = json.EncodeObjectColon(*out)
		return writeDefaultOrEmpty(field, out)
	})
}

func writeDefaultOrEmpty(field *thrift.FieldDescriptor, out *[]byte) (err error) {
	if dv := field.DefaultValue(); dv != nil {
		*out = append(*out, dv.JSONValue()...)
		return nil
	}
	switch t := field.Type().Type(); t {
	case thrift.BOOL:
		*out = json.EncodeBool(*out, false)
	case thrift.BYTE, thrift.I16, thrift.I32, thrift.I64:
		*out = json.EncodeInt64(*out, 0)
	case thrift.DOUBLE:
		*out = json.EncodeFloat64(*out, 0)
	case thrift.STRING:
		*out = json.EncodeString(*out, "")
	case thrift.LIST, thrift.SET:
		*out = json.EncodeEmptyArray(*out)
	case thrift.MAP:
		*out = json.EncodeEmptyObject(*out)
	case thrift.STRUCT:
		*out = json.EncodeObjectBegin(*out)
		*out = json.EncodeObjectEnd(*out)
	default:
		return wrapError(meta.ErrUnsupportedType, fmt.Sprintf("unknown descriptor type %s", t), nil)
	}
	return err
}

func (self *BinaryConv) buildinTypeToKey(p *thrift.BinaryProtocol, dest *thrift.TypeDescriptor, out *[]byte) error {
	*out = append(*out, '"')
	switch t := dest.Type(); t {
	// case thrift.BOOL:
	// 	v, err := p.ReadBool()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	*out = json.EncodeBool(*out, v)
	case thrift.I08:
		v, err := p.ReadByte()
		if err != nil {
			return err
		}
		if self.opts.ByteAsUint8 {
			*out = json.EncodeInt64(*out, int64(uint8(v)))
		} else {
			*out = json.EncodeInt64(*out, int64(int8(v)))
		}
	case thrift.I16:
		v, err := p.ReadI16()
		if err != nil {
			return err
		}
		*out = json.EncodeInt64(*out, int64(v))
	case thrift.I32:
		v, err := p.ReadI32()
		if err != nil {
			return err
		}
		*out = json.EncodeInt64(*out, int64(v))
	case thrift.I64:
		v, err := p.ReadI64()
		if err != nil {
			return err
		}
		*out = json.EncodeInt64(*out, int64(v))
	// case thrift.DOUBLE:
	// 	v, err := p.ReadDouble()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	*out, err = json.EncodeFloat64(*out, v)
	case thrift.STRING:
		v, err := p.ReadString(false)
		if err != nil {
			return err
		}
		json.NoQuote(out, v)
	default:
		return wrapError(meta.ErrUnsupportedType, fmt.Sprintf("unsupported descriptor type %s as MAP key", t), nil)
	}
	*out = append(*out, '"')
	return nil
}

func (self *BinaryConv) writeHttpValue(ctx context.Context, resp http.ResponseSetter, p *thrift.BinaryProtocol, field *thrift.FieldDescriptor) (ok bool, err error) {
	var thriftVal []byte
	var jsonVal []byte
	var textVal []byte

	for _, hm := range field.HTTPMappings() {
		var val []byte
		enc := hm.Encoding()

		if enc == meta.EncodingThriftBinary {
			// raw encoding, check if raw value is set
			if thriftVal == nil {
				var start = p.Read
				if err := p.Skip(field.Type().Type(), self.opts.UseNativeSkip); err != nil {
					return false, wrapError(meta.ErrRead, "", err)
				}
				val = p.Buf[start:p.Read]
				thriftVal = val
			} else {
				val = thriftVal
			}

		} else if enc == meta.EncodingText || !field.Type().Type().IsComplex() { // not complex value, must use text encoding
			if textVal == nil {
				tmp := make([]byte, 0, conv.DefaulHttpValueBufferSizeForJSON)
				err = p.ReadStringWithDesc(field.Type(), &tmp, self.opts.ByteAsUint8, self.opts.DisallowUnknownField, !self.opts.NoBase64Binary)
				if err != nil {
					return false, unwrapError(fmt.Sprintf("reading thrift value of '%s' failed, thrift pos:%d", field.Name(), p.Read), err)
				}
				val = tmp
				textVal = val
			} else {
				val = textVal
			}
		} else if self.opts.UseKitexHttpEncoding {
			// kitex http encoding fallback
			if textVal == nil {
				obj, err := p.ReadAnyWithDesc(field.Type(), self.opts.ByteAsUint8, !self.opts.NoCopyString, self.opts.DisallowUnknownField, true)
				if err != nil {
					return false, unwrapError(fmt.Sprintf("reading thrift value of '%s' failed, thrift pos:%d", field.Name(), p.Read), err)
				}
				textVal = rt.Str2Mem(primitive.KitexToString(obj))
				val = textVal
			} else {
				val = textVal
			}
		} else if enc == meta.EncodingJSON {
			// for nested type, convert it to a new JSON string
			if jsonVal == nil {
				tmp := make([]byte, 0, conv.DefaulHttpValueBufferSizeForJSON)
				err := self.doRecurse(ctx, field.Type(), &tmp, resp, p)
				if err != nil {
					return false, unwrapError(fmt.Sprintf("mapping field %s failed, thrift pos:%d", field.Name(), p.Read), err)
				}
				val = tmp
				jsonVal = val
			} else {
				val = jsonVal
			}

		} else {
			return false, wrapError(meta.ErrConvert, fmt.Sprintf("unsuported http-value encoding %v of field '%s'", enc, field.Name()), nil)
		}

		// NOTICE: ignore error if the value is not set
		if e := hm.Response(ctx, resp, field, rt.Mem2Str(val)); e == nil {
			ok = true
			break
		} else if !self.opts.OmitHttpMappingErrors {
			return false, e
		}
	}
	return
}

// func setValueToHttp(resp http.ResponseSetter, typ thrift.AnnoType, key string, val string) error {
// 	switch typ {
// 	case annotation.APIHeader:
// 		return resp.SetHeader(key, val)
// 	case annotation.APIRawBody:
// 		return resp.SetRawBody(rt.Str2Mem(val))
// 	case annotation.APICookie:
// 		return resp.SetCookie(key, val)
// 	case annotation.APIHTTPCode:
// 		code, err := strconv.Atoi(val)
// 		if err != nil {
// 			return err
// 		}
// 		return resp.SetStatusCode(code)
// 	default:
// 		return wrapError(meta.ErrUnsupportedType, fmt.Sprintf("unsupported annotation type %d", typ), nil)
// 	}
// }
