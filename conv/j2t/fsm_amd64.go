//go:build amd64
// +build amd64

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
	"runtime"
	"unsafe"

	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/native"
	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/thrift"
)

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
