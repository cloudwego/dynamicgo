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
    __Quote func(s unsafe.Pointer, nb int, dp unsafe.Pointer, dn unsafe.Pointer, flags uint64) int

    __Unquote func(s unsafe.Pointer, nb int, dp unsafe.Pointer, ep unsafe.Pointer, flags uint64) int

    __HTMLEscape func(s unsafe.Pointer, nb int, dp unsafe.Pointer, dn unsafe.Pointer) int

    __Value func(s unsafe.Pointer, n int, p int, v unsafe.Pointer, flags uint64) int

    __SkipOne func(s unsafe.Pointer, p unsafe.Pointer, m unsafe.Pointer) int

    __GetByPath func(s unsafe.Pointer, p unsafe.Pointer, path unsafe.Pointer) int

    __ValidateOne func(s unsafe.Pointer, p unsafe.Pointer, m unsafe.Pointer) int

    __I64toa func(out unsafe.Pointer, val int64) (ret int)

    __U64toa func(out unsafe.Pointer, val uint64) (ret int)

    __F64toa func(out unsafe.Pointer, val float64) (ret int)

    __ValidateUTF8 func(s unsafe.Pointer, p unsafe.Pointer, m unsafe.Pointer) (ret int)

    __ValidateUTF8Fast func(s unsafe.Pointer) (ret int)

    __J2T_FSM func(fsm unsafe.Pointer, buf unsafe.Pointer, src unsafe.Pointer, flag uint64) (ret uint64) 

    __TB_Skip func(st unsafe.Pointer, s unsafe.Pointer, n int, t uint8) (ret int)
)

//go:nosplit
func Quote(s unsafe.Pointer, nb int, dp unsafe.Pointer, dn *int, flags uint64) int {
    return __Quote(rt.NoEscape(unsafe.Pointer(s)), nb, rt.NoEscape(unsafe.Pointer(dp)), rt.NoEscape(unsafe.Pointer(dn)), flags)
}

//go:nosplit
func Unquote(s unsafe.Pointer, nb int, dp unsafe.Pointer, ep *int, flags uint64) int {
    return __Unquote(rt.NoEscape(unsafe.Pointer(s)), nb, rt.NoEscape(unsafe.Pointer(dp)), rt.NoEscape(unsafe.Pointer(ep)), flags)
}

//go:nosplit
func HTMLEscape(s unsafe.Pointer, nb int, dp unsafe.Pointer, dn *int) int {
    return __HTMLEscape(rt.NoEscape(unsafe.Pointer(s)), nb, rt.NoEscape(unsafe.Pointer(dp)), rt.NoEscape(unsafe.Pointer(dn)))
}

//go:nosplit
func Value(s unsafe.Pointer, n int, p int, v *types.JsonState, flags uint64) int {
    return __Value(rt.NoEscape(unsafe.Pointer(s)), n, p, rt.NoEscape(unsafe.Pointer(v)), flags)
}

//go:nosplit
func SkipOne(s *string, p *int, m *types.StateMachine) int {
    return __SkipOne(rt.NoEscape(unsafe.Pointer(s)), rt.NoEscape(unsafe.Pointer(p)), rt.NoEscape(unsafe.Pointer(m)))
}

//go:nosplit
func ValidateOne(s *string, p *int, m *types.StateMachine) int {
    return __ValidateOne(rt.NoEscape(unsafe.Pointer(s)), rt.NoEscape(unsafe.Pointer(p)), rt.NoEscape(unsafe.Pointer(m)))
}

//go:nosplit
func I64toa(out *byte, val int64) (ret int) {
    return __I64toa(rt.NoEscape(unsafe.Pointer(out)), val)
}

//go:nosplit
func U64toa(out *byte, val uint64) (ret int) {
    return __U64toa(rt.NoEscape(unsafe.Pointer(out)), val)
}

//go:nosplit
func F64toa(out *byte, val float64) (ret int) {
    return __F64toa(rt.NoEscape(unsafe.Pointer(out)), val)
}

//go:nosplit
func J2T_FSM(fsm *types.J2TStateMachine, buf *[]byte, src *string, flag uint64) (ret uint64) {
	return __J2T_FSM(rt.NoEscape(unsafe.Pointer(fsm)), rt.NoEscape(unsafe.Pointer(buf)), rt.NoEscape(unsafe.Pointer(src)), flag)
}

//go:nosplit
func TBSkip(st *types.TStateMachine, s *byte, n int, t uint8) (ret int) {
	return __TB_Skip(rt.NoEscape(unsafe.Pointer(st)), rt.NoEscape(unsafe.Pointer(s)), n, t)
}

var stubs = []loader.GoC{
    {"_f64toa", nil, &__F64toa},
    {"_i64toa", nil, &__I64toa},
    {"_u64toa", nil, &__U64toa},
    {"_quote", nil, &__Quote},
    {"_unquote", nil, &__Unquote},
    {"_html_escape", nil, &__HTMLEscape},
    {"_value", nil, &__Value},
    {"_skip_one", nil, &__SkipOne},
    {"_validate_one", nil, &__ValidateOne},
    {"_j2t_fsm_exec", nil, &__J2T_FSM},
    {"_tb_skip", nil, &__TB_Skip},
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
