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

	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
)

func (self CompactConv) handleError(ctx context.Context, fsm *types.J2TStateMachine, buf *[]byte, src []byte, req http.RequestGetter, ret uint64, top bool) (cont bool, err error) {
	e := types.ParsingError(ret & ((1 << types.ERR_WRAP_SHIFT_CODE) - 1))
	p := int(ret >> types.ERR_WRAP_SHIFT_CODE)

	switch e {
	case types.ERR_HTTP_MAPPING:
		{
			desc, ext := getJ2TExtraStruct(fsm, fsm.SP)
			if desc == nil || ext == nil {
				return false, newError(meta.ErrConvert, "invalid json input", nil)
			}
			if desc.Type() != thrift.STRUCT {
				return false, newError(meta.ErrConvert, "invalid descriptor while http mapping", nil)
			}
			return true, self.writeHttpRequestToThrift(ctx, fsm, req, desc.Struct(), &ext.reqs, buf, false, top)
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
	case types.ERR_OOM_CWBS:
		{
			fsm.GrowContainerWriteBack(p)
			return true, nil
		}
	case types.ERR_OOM_LFID:
		{
			fsm.GrowLastFieldID(p)
			return true, nil
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
