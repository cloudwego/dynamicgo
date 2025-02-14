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
	"errors"
	"fmt"
	"strconv"
	"unsafe"

	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/json"
	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
)

var errNull = errors.New("error null")

func errSyntax(s string, r int) error {
	if r >= len(s) {
		return newError(meta.ErrRead, "EOF at "+strconv.Itoa(r)+"", nil)
	} else if r < 0 {
		return newError(meta.ErrRead, "", types.ParsingError(-r))
	} else {
		return newError(meta.ErrRead, "invalid char '"+string(s[r])+"'", nil)
	}
}

func unwrapError(msg string, err error) error {
	if v, ok := err.(meta.Error); ok {
		return newError(v.Code, msg, err)
	} else {
		return newError(meta.ErrConvert, msg, err)
	}
}

//go:noinline
func newError(code meta.ErrCode, msg string, err error) error {
	return meta.NewError(meta.NewErrorCode(code, meta.JSON2THRIFT), msg, err)
	// panic(meta.NewError(meta.NewErrorCode(code, meta.JSON2THRIFT), msg, err).Error())
}

type _J2TExtra_STRUCT struct {
	desc unsafe.Pointer
	reqs string
}

func expectType(typ thrift.Type, act thrift.Type) error {
	if typ != act {
		return newError(meta.ErrDismatchType, "expect "+typ.String()+" but got "+act.String(), nil)
	}
	return nil
}

func expectType2(typ1 thrift.Type, typ2 thrift.Type, act thrift.Type) error {
	if typ1 != act && typ2 != act {
		return newError(meta.ErrDismatchType, "expect "+typ1.String()+" or "+typ2.String()+" but got "+act.String(), nil)
	}
	return nil
}

func locateInput(in []byte, ip int) string {
	je := json.SyntaxError{
		Pos: (ip),
		Src: string(in),
	}
	return je.Locate()
}

func makePanicMsg(msg interface{}, src []byte, desc *thrift.TypeDescriptor, buf *[]byte, req http.RequestGetter, flags uint64, fsm *types.J2TStateMachine, ret uint64) string {
	var re string
	if rt, ok := req.(*http.HTTPRequest); ok {
		re = fmt.Sprintf("%v", rt)
	}
	return fmt.Sprintf(`%v
Flags: %b
JSON: %s
Buf: %v
<FSM>
%s
</FSM>
Desc: %s
<Request>
%s
</Request>
Ret: %x`, msg, flags, string(src), *buf, fsm.String(), desc.Name(), re, ret)
}
