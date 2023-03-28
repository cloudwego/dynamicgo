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
	"encoding/base64"
	"fmt"
	"unsafe"

	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/json"
	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
)

//go:noinline
func newError(code meta.ErrCode, msg string, err error) error {
	return meta.NewError(meta.NewErrorCode(code, meta.JSON2THRIFT), msg, err)
	// panic(meta.NewError(meta.NewErrorCode(code, meta.JSON2THRIFT), msg, err).Error())
}

type _J2TExtra_STRUCT struct {
	desc unsafe.Pointer
	reqs string
}

//go:nocheckptr
func getJ2TExtraStruct(fsm *types.J2TStateMachine, offset int) (td *thrift.TypeDescriptor, reqs thrift.RequiresBitmap) {
	state := fsm.At(offset - 1)
	if state == nil {
		return nil, thrift.RequiresBitmap{}
	}
	td = (*thrift.TypeDescriptor)(unsafe.Pointer(state.TdPointer()))
	je := (*_J2TExtra_STRUCT)(unsafe.Pointer(&state.Extra))
	v := rt.Str2Mem(je.reqs)
	reqs = *(*thrift.RequiresBitmap)(unsafe.Pointer(&v))
	return
}

func (self BinaryConv) handleError(ctx context.Context, fsm *types.J2TStateMachine, buf *[]byte, src []byte, req http.RequestGetter, ret uint64, top bool) (cont bool, err error) {
	e := types.ParsingError(ret & ((1 << types.ERR_WRAP_SHIFT_CODE) - 1))
	p := int(ret >> types.ERR_WRAP_SHIFT_CODE)

	switch e {
	case types.ERR_HTTP_MAPPING:
		{
			desc, reqs := getJ2TExtraStruct(fsm, fsm.SP)
			if desc == nil {
				return false, newError(meta.ErrConvert, "invalid json input", nil)
			}
			if desc.Type() != thrift.STRUCT {
				return false, newError(meta.ErrConvert, "invalid descriptor while http mapping", nil)
			}
			return true, self.writeHttpRequestToThrift(ctx, req, desc.Struct(), reqs, buf, false, top)
		}
	case types.ERR_HTTP_MAPPING_END:
		{
			desc, _ := getJ2TExtraStruct(fsm, fsm.SP)
			if desc == nil {
				return false, newError(meta.ErrConvert, "invalid json input", nil)
			}
			if desc.Type() != thrift.STRUCT {
				return false, newError(meta.ErrConvert, "invalid descriptor while http mapping", nil)
			}
			if len(fsm.FieldCache) == 0 {
				return false, newError(meta.ErrConvert, "invalid FSM field-cache length", nil)
			}
			return self.handleUnmatchedFields(ctx, fsm, desc.Struct(), buf, p, req, top)
		}
	case types.ERR_OOM_BM:
		{
			fsm.GrowReqCache(p)
			return true, nil
		}
	case types.ERR_OOM_KEY:
		{
			fsm.GrowKeyCache(p)
			return true, nil
		}
	case types.ERR_OOM_BUF:
		{
			c := cap(*buf)
			c += c >> 1
			if c < cap(*buf)+p {
				c = cap(*buf) + p*2
			}
			tmp := make([]byte, len(*buf), c)
			copy(tmp, *buf)
			*buf = tmp
			return true, nil
		}
	case types.ERR_OOM_FIELD:
		{
			fsm.GrowFieldCache(types.J2T_FIELD_CACHE_SIZE)
			fsm.SetPos(p)
			return true, nil
		}
	// case types.ERR_OOM_FVAL:
	// 	{
	// 		fsm.GrowFieldValueCache(types.J2T_FIELD_CACHE_SIZE)
	// 		fsm.SetPos(p)
	// 		return true, nil
	// 	}
	case types.ERR_VALUE_MAPPING_END:
		{
			desc, _ := getJ2TExtraStruct(fsm, fsm.SP-1)
			if desc == nil {
				return false, newError(meta.ErrConvert, "invalid json input", nil)
			}
			if desc.Type() != thrift.STRUCT {
				return false, newError(meta.ErrConvert, "invalid descriptor while value mapping", nil)
			}
			return self.handleValueMapping(ctx, fsm, desc.Struct(), buf, p, src)
		}
	}

	return false, explainNativeError(e, src, p)
}

func explainNativeError(e types.ParsingError, in []byte, v int) error {
	ip := v & ((1 << types.ERR_WRAP_SHIFT_POS) - 1)
	v = v >> types.ERR_WRAP_SHIFT_POS
	switch e {
	case types.ERR_INVALID_CHAR:
		ch, st := v>>types.ERR_WRAP_SHIFT_CODE, v&((1<<types.ERR_WRAP_SHIFT_CODE)-1)
		return newError(meta.ErrRead, fmt.Sprintf("invalid char '%c' for state %s, near %d of %s", byte(ch), types.J2T_STATE(st), ip, locateInput(in, ip)), e)
	case types.ERR_INVALID_NUMBER_FMT:
		return newError(meta.ErrConvert, fmt.Sprintf("unexpected number type %d, near %d of %s", types.ValueType(v), ip, locateInput(in, ip)), e)
	case types.ERR_UNSUPPORT_THRIFT_TYPE:
		t := thrift.Type(v)
		return newError(meta.ErrUnsupportedType, fmt.Sprintf("unsupported thrift type %s, near %d of %s", t, ip, locateInput(in, ip)), nil)
	case types.ERR_UNSUPPORT_VM_TYPE:
		t := thrift.AnnoID(v)
		return newError(meta.ErrUnsupportedType, fmt.Sprintf("unsupported value-mapping type %d, near %d of %q", t, ip, locateInput(in, ip)), nil)
	case types.ERR_DISMATCH_TYPE:
		exp, act := v>>types.ERR_WRAP_SHIFT_CODE, v&((1<<types.ERR_WRAP_SHIFT_CODE)-1)
		return newError(meta.ErrDismatchType, fmt.Sprintf("expect type %s but got type %d, near %d of %s", thrift.Type(exp), act, ip, locateInput(in, ip)), nil)
	case types.ERR_NULL_REQUIRED:
		id := thrift.FieldID(v)
		return newError(meta.ErrMissRequiredField, fmt.Sprintf("missing required field %d, near %d of %s", id, ip, locateInput(in, ip)), nil)
	case types.ERR_UNKNOWN_FIELD:
		n := ip - v - 1
		if n < 0 {
			n = 0
		}
		key := in[n : ip-1]
		return newError(meta.ErrUnknownField, fmt.Sprintf("unknown field '%s', near %d of %s", string(key), ip, locateInput(in, ip)), nil)
	case types.ERR_RECURSE_EXCEED_MAX:
		return newError(meta.ErrStackOverflow, fmt.Sprintf("stack %d overflow, near %d of %s", v, ip, locateInput(in, ip)), nil)
	case types.ERR_DECODE_BASE64:
		berr := base64.CorruptInputError(v)
		return newError(meta.ErrRead, fmt.Sprintf("decode base64 error: %v, near %d of %s", berr, ip, locateInput(in, ip)), nil)
	default:
		return newError(meta.ErrConvert, fmt.Sprintf("native error %q, value %d, near %d of %s", types.ParsingError(e).Message(), v, ip, locateInput(in, ip)), nil)
	}
}

func locateInput(in []byte, ip int) string {
	je := json.SyntaxError{
		Pos: (ip),
		Src: string(in),
	}
	return je.Locate()
}
