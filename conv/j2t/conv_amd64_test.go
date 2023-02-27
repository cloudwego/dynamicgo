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
	"encoding/json"
	"runtime"
	"testing"
	"unsafe"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/native"
	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example3"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/stretchr/testify/require"
)

func TestConvJSON2Thrift(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()
	cv := NewBinaryConv(conv.Options{})
	ctx := context.Background()
	out, err := cv.Do(ctx, desc, data)
	require.Nil(t, err)
	exp := example3.NewExampleReq()
	err = json.Unmarshal(data, exp)
	require.Nil(t, err)
	act := example3.NewExampleReq()
	_, err = act.FastRead(out)
	require.Nil(t, err)
	require.Equal(t, exp, act)
}

func (mock MockConv) do(self *BinaryConv, ctx context.Context, src []byte, desc *thrift.TypeDescriptor, buf *[]byte, req http.RequestGetter, top bool) (err error) {
	flags := toFlags(self.opts)
	jp := rt.Mem2Str(src)
	tmp := make([]byte, 0, mock.dcap)
	fsm := &types.J2TStateMachine{
		SP: mock.sp,
		JT: types.JsonState{
			Dbuf: *(**byte)(unsafe.Pointer(&tmp)),
			Dcap: mock.dcap,
		},
		ReqsCache:  make([]byte, 0, mock.reqsCache),
		KeyCache:   make([]byte, 0, mock.keyCache),
		SM:         types.StateMachine{},
		VT:         [types.MAX_RECURSE]types.J2TState{},
		FieldCache: make([]int32, 0, mock.fc),
	}
	fsm.Init(0, unsafe.Pointer(desc))
	if mock.sp != 0 {
		fsm.SP = mock.sp
	}

exec:
	ret := native.J2T_FSM(fsm, buf, &jp, flags)
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
