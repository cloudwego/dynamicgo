//go:build !amd64 || go1.25

/**
 * Copyright 2025 ByteDance Inc.
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
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/json"
	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/cloudwego/dynamicgo/internal/rt"
	_ "github.com/cloudwego/dynamicgo/internal/warning"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
)

func (self *BinaryConv) doImpl(ctx context.Context, src []byte, desc *thrift.TypeDescriptor, buf *[]byte, req http.RequestGetter, top bool) (err error) {
	return self.doGo(ctx, rt.Mem2Str(src), desc, buf, req, top)
}

func (self *BinaryConv) doGo(ctx context.Context, src string, desc *thrift.TypeDescriptor, buf *[]byte, req http.RequestGetter, top bool) (err error) {
	p := thrift.BinaryProtocol{Buf: *buf}
	depth := 0
	if !top {
		depth = 1
	}
	ret, err := self.doRecurse(ctx, src, 0, desc, &p, req, depth)
	*buf = p.Buf
	if err != nil {
		return unwrapError(locateInput([]byte(src), ret), err)
	}
	return nil
}

func (self *BinaryConv) doRecurse(ctx context.Context, s string, jp int, desc *thrift.TypeDescriptor, p *thrift.BinaryProtocol, req http.RequestGetter, depth int) (ret int, err error) {
	n := len(s)
	ret = jp

	for ret < n && ret >= 0 {
		jp, v := json.DecodeValue(s, ret)
		if jp < 0 {
			return jp, errSyntax(s, jp)
		}
		ret = jp

		switch v.Vt {

		case types.V_NULL:
			// go back to upper layer to handle nil values
			return ret, errNull

		case types.V_TRUE, types.V_FALSE:
			if err = expectType(thrift.BOOL, desc.Type()); err != nil {
				return
			}
			return ret, p.WriteBool(v.Vt == types.V_TRUE)

		case types.V_INTEGER:
			if t := desc.Type(); t.IsInt() {
				return ret, p.WriteInt(t, int(v.Iv))
			} else if t == thrift.DOUBLE {
				return ret, p.WriteDouble(float64(v.Iv))
			} else {
				return ret, newError(meta.ErrDismatchType, "expected thrift INT or DOUBLE type", nil)
			}

		case types.V_DOUBLE:
			if t := desc.Type(); t.IsInt() {
				return ret, p.WriteInt(t, int(v.Dv))
			} else if t == thrift.DOUBLE {
				return ret, p.WriteDouble(float64(v.Dv))
			} else {
				return ret, newError(meta.ErrDismatchType, "expected thrift INT or DOUBLE type", nil)
			}

		case types.V_STRING:
			var str string
			if v.Ep >= 0 && v.Ep < int64(ret) {
				str, err = strconv.Unquote(s[v.Iv-1 : ret])
				if err != nil {
					return
				}
			} else {
				str = s[v.Iv : ret-1]
			}

			if desc.IsBinary() && !self.opts.NoBase64Binary {
				bs, err := base64.StdEncoding.DecodeString(str)
				if err != nil {
					return ret, err
				}
				return ret, p.WriteBinary(bs)

			} else if desc.Type() == thrift.STRING {
				return ret, p.WriteString(str)

			} else if self.opts.String2Int64 && desc.Type().IsInt() {
				if str == "" {
					str = "0"
				}
				iv, err := strconv.ParseInt(str, 10, 64)
				if err != nil {
					return ret, err
				}
				return ret, p.WriteInt(desc.Type(), int(iv))

			} else if self.opts.String2Int64 && desc.Type() == thrift.DOUBLE {
				if str == "" {
					str = "0"
				}
				dv, err := strconv.ParseFloat(str, 64)
				if err != nil {
					return ret, err
				}
				return ret, p.WriteDouble(dv)
			}

		case types.V_ARRAY:
			if err = expectType2(thrift.LIST, thrift.SET, desc.Type()); err != nil {
				return
			}

			et := desc.Elem()
			back, _ := p.WriteListBeginWithSizePos(et.Type(), 0)
			size := int32(0)

			jp, nt := json.Peek(s, ret)
			if nt == json.EndArr {
				ret = jp + 1
				return ret, nil
			}

			nt = json.Comma
			for nt == json.Comma {
				jp, err := self.doRecurse(ctx, s, ret, et, p, req, depth+1)
				if err != nil && err != errNull {
					return jp, err
				}
				ret = jp
				if err != errNull {
					size += 1
				}
				jp, nt = json.Peek(s, ret)
				ret = jp + 1
			}

			if nt == json.EndArr {
				return ret, p.ModifyI32(back, int32(size))
			} else {
				return ret, errSyntax(s, ret)
			}

		case types.V_OBJECT:
			if t := desc.Type(); t == thrift.MAP {
				size := 0
				kt := desc.Key()
				et := desc.Elem()
				back, _ := p.WriteMapBeginWithSizePos(kt.Type(), et.Type(), 0)

				jp, nt := json.Peek(s, ret)
				if nt == json.EndObj {
					ret = jp + 1
					return ret, nil
				}

				nt = json.Comma
				for nt == json.Comma {
					jp, v = json.DecodeValue(s, ret)
					if v.Vt != types.V_STRING {
						return ret, errSyntax(s, ret)
					}
					ret = jp

					var key string
					if v.Ep >= 0 && v.Ep < int64(ret) {
						key, err = strconv.Unquote(s[v.Iv-1 : ret])
						if err != nil {
							return
						}
					} else {
						key = s[v.Iv : ret-1]
					}

					ks := len(p.Buf)

					if kt.Type() == thrift.STRING {
						p.WriteString(key)

					} else if kt.Type().IsInt() {
						//todo: use native
						i, err := strconv.ParseInt(key, 10, 64)
						if err != nil {
							return ret, err
						}
						p.WriteInt(kt.Type(), int(i))

					} else {
						return ret, newError(meta.ErrUnsupportedType, "thrift MAP key type must be STRING", nil)
					}

					if jp, nt = json.Peek(s, ret); nt != json.Colon {
						return ret, errSyntax(s, ret)
					} else {
						ret = jp + 1
					}

					jp, err := self.doRecurse(ctx, s, ret, et, p, req, depth+1)
					if err != nil && err != errNull {
						return jp, err
					}
					ret = jp

					if err != errNull {
						size += 1
					} else {
						// unwind written key
						p.Buf = p.Buf[:ks]
					}

					jp, nt = json.Peek(s, ret)
					ret = jp + 1
				}

				if nt == json.EndObj {
					return ret, p.ModifyI32(back, int32(size))
				} else {
					return ret, errSyntax(s, ret)
				}

			} else if t == thrift.STRUCT {
				bm := thrift.NewRequiresBitmap()
				desc.Struct().Requires().CopyTo(bm)

				// do http-mapping if any
				if self.opts.EnableHttpMapping {
					if err := self.handleHttpMappings(ctx, req, desc.Struct(), *bm, &p.Buf, false, depth == 0); err != nil {
						return ret, err
					}
				}

				jp, nt := json.Peek(s, ret)
				if nt == json.EndObj {
					ret = jp + 1
					goto STRUCT_END
				}

				nt = json.Comma
				for nt == json.Comma {
					jp, v = json.DecodeValue(s, ret)
					if v.Vt != types.V_STRING {
						return ret, errSyntax(s, ret)
					}
					ret = jp

					var key string
					if v.Ep >= 0 && v.Ep < int64(ret) {
						key, err = strconv.Unquote(s[v.Iv-1 : ret])
						if err != nil {
							return
						}
					} else {
						key = s[v.Iv : ret-1]
					}

					if jp, nt = json.Peek(s, ret); nt != json.Colon {
						return ret, errSyntax(s, ret)
					} else {
						ret = jp + 1
					}

					ft := desc.Struct().FieldByKey(key)
					if ft == nil {
						if self.opts.DisallowUnknownField {
							return ret, newError(meta.ErrUnknownField, key, nil)
						}
						jp, _ = json.SkipValue(s, ret)
						if jp < 0 {
							return ret, errSyntax(s, jp)
						}
						ret = jp
						goto OBJECT_NEXT
					}

					if self.opts.EnableHttpMapping && len(ft.HTTPMappings()) > 0 && !bm.IsSet(ft.ID()) {
						// skip http-mapped value
						jp, _ = json.SkipValue(s, ret)
						if jp < 0 {
							return ret, errSyntax(s, jp)
						}
						ret = jp

					} else if self.opts.EnableValueMapping && ft.ValueMapping() != nil {
						// handle value-mapped value
						jp, start := json.SkipValue(s, ret)
						if jp < 0 {
							return ret, errSyntax(s, jp)
						}
						ret = jp
						if err = self.handleValueMapping(ctx, s, start, ft, p, jp); err != nil {
							return ret, err
						}
						bm.Set(ft.ID(), thrift.OptionalRequireness)

					} else {
						// normal json mapping
						ks := len(p.Buf)
						p.WriteFieldBegin(key, ft.Type().Type(), ft.ID())

						jp, err := self.doRecurse(ctx, s, ret, ft.Type(), p, req, depth+1)
						if err != nil && err != errNull {
							return jp, err
						}
						ret = jp

						if err == errNull {
							// unwind written field tag
							p.Buf = p.Buf[:ks]
						}

						bm.Set(ft.ID(), thrift.OptionalRequireness)
					}

				OBJECT_NEXT:
					jp, nt = json.Peek(s, ret)
					ret = jp + 1
				}

			STRUCT_END:
				if nt == json.EndObj {
					// notice: when option TracebackRequredOrRootFields enabled, we should always WriteXXField
					traceback := self.opts.TracebackRequredOrRootFields && depth == 0
					if err := bm.HandleRequires(desc.Struct(), self.opts.WriteRequireField || traceback, self.opts.WriteDefaultField || traceback, self.opts.WriteOptionalField || traceback, func(f *thrift.FieldDescriptor) error {
						// special case: traceback http values for root of required fields
						var val string
						var enc meta.Encoding
						if self.opts.TracebackRequredOrRootFields && (depth == 0 || f.Required() == thrift.RequiredRequireness) {
							val, _, enc = tryGetValueFromHttp(req, f.Alias())
						}
						return self.writeStringValue(ctx, &p.Buf, f, val, enc, req)
					}); err != nil {
						return ret, err
					}

					p.WriteStructEnd()
				} else {

					return ret, errSyntax(s, ret)
				}

				thrift.FreeRequiresBitmap(bm)
				return

			} else {
				return ret, newError(meta.ErrDismatchType, "json object can only convert to thrift MAP or STRUCT", nil)
			}
		}
	}

	return
}

func (self *BinaryConv) handleValueMapping(ctx context.Context, src string, start int, f *thrift.FieldDescriptor, p *thrift.BinaryProtocol, end int) error {
	// write field tag
	if err := p.WriteFieldBegin(f.Name(), f.Type().Type(), f.ID()); err != nil {
		return newError(meta.ErrWrite, fmt.Sprintf("failed to write field '%s' tag", f.Name()), err)
	}
	// truncate src to value
	if int(start) >= len(src) || int(end) < 0 {
		return newError(meta.ErrConvert, "invalid value-mapping position", nil)
	}
	jdata := rt.Str2Mem(src[start:end])
	// call ValueMapping interface
	if err := f.ValueMapping().Write(ctx, p, f, jdata); err != nil {
		return newError(meta.ErrConvert, fmt.Sprintf("failed to convert field '%s' value", f.Name()), err)
	}
	return nil
}
