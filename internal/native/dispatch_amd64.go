/*
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

package native

import (
	"unsafe"

	"github.com/bytedance/sonic/loader"
	"github.com/cloudwego/dynamicgo/internal/cpu"
	"github.com/cloudwego/dynamicgo/internal/native/avx"
	"github.com/cloudwego/dynamicgo/internal/native/avx2"
	"github.com/cloudwego/dynamicgo/internal/native/sse"
	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/cloudwego/dynamicgo/internal/rt"
)

const MaxFrameSize uintptr = 1024

var (
	__j2t_fsm_exec func(fsm unsafe.Pointer, buf unsafe.Pointer, src unsafe.Pointer, flag uint64) (ret uint64) 

	__tb_skip func(st unsafe.Pointer, s unsafe.Pointer, n int, t uint8) (ret int)
)

//go:nosplit
//go:linkname Quote github.com/bytedance/sonic/internal/native.Quote
func Quote(s unsafe.Pointer, nb int, dp unsafe.Pointer, dn *int, flags uint64) int

//go:nosplit
//go:linkname Unquote github.com/bytedance/sonic/internal/native.Unquote
func Unquote(s unsafe.Pointer, nb int, dp unsafe.Pointer, ep *int, flags uint64) int

//go:nosplit
//go:linkname Value github.com/bytedance/sonic/internal/native.Value
func Value(s unsafe.Pointer, n int, p int, v *types.JsonState, allow_control int) int

//go:nosplit
//go:linkname SkipOne github.com/bytedance/sonic/internal/native.SkipOne
func SkipOne(s *string, p *int, m *types.StateMachine) int

//go:nosplit
//go:linkname I64toa github.com/bytedance/sonic/internal/native.I64toa
func I64toa(out *byte, val int64) (ret int)

//go:nosplit
//go:linkname F64toa github.com/bytedance/sonic/internal/native.F64toa
func F64toa(out *byte, val float64) (ret int)

//go:nosplit
func TBSkip(st *types.TStateMachine, s *byte, n int, t uint8) (ret int) {
	return __tb_skip(rt.NoEscape(unsafe.Pointer(st)), rt.NoEscape(unsafe.Pointer(s)), n, t)
}

//go:nosplit
func J2T_FSM(fsm *types.J2TStateMachine, buf *[]byte, src *string, flag uint64) (ret uint64) {
    return __j2t_fsm_exec(rt.NoEscape(unsafe.Pointer(fsm)), rt.NoEscape(unsafe.Pointer(buf)), rt.NoEscape(unsafe.Pointer(src)), flag)
}

var stubs = []loader.GoC{
    {"_j2t_fsm_exec", nil, &__j2t_fsm_exec},
    {"_tb_skip", nil, &__tb_skip},
}

func useAVX() {
    loader.WrapGoC(avx.Text__native_entry__, avx.Funcs, stubs, "avx", "avx/native.c")
}

func useAVX2() {
    loader.WrapGoC(avx2.Text__native_entry__, avx2.Funcs, stubs, "avx2", "avx2/native.c")
}

func useSSE() {
    loader.WrapGoC(sse.Text__native_entry__, sse.Funcs, stubs, "sse", "sse/native.c")
}

func init() {
    if cpu.HasAVX2 {
        useAVX2()
    } else if cpu.HasAVX {
        useAVX()
    } else if cpu.HasSSE {
        useSSE()
    } else {
        panic("Unsupported CPU, maybe it's too old to run Sonic.")
    }
}