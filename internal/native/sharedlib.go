//go:build sharedlib
// +build sharedlib

// ! FOR TESTING PURPOSES

package native

import (
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/native/types"
)

/*
#cgo CFLAGS: -I../../native
#cgo LDFLAGS: -L../../output -lnative

#include "native.h"
#include "thrift.h"
#include "thrift_skip.h"
*/
import "C"

//go:nosplit
//go:nocheckptr
//goland:noinspection GoUnusedParameter
func Value(s unsafe.Pointer, n int, p int, v *types.JsonState, allow_control int) int {
	retc := C.value(
		(*C.char)(s),
		C.size_t(n),
		C.long(p),
		(*C.JsonState)(unsafe.Pointer(v)),
		C.int(allow_control),
	)
	return int(retc)
}

//go:nosplit
//go:nocheckptr
//goland:noinspection GoUnusedParameter
func SkipOne(s *string, p *int, m *types.StateMachine) int {
	retc := C.skip_one(
		(*C._GoString)(unsafe.Pointer(s)),
		(*C.long)(unsafe.Pointer(p)),
		(*C.StateMachine)(unsafe.Pointer(m)),
	)
	return int(retc)
}

//go:nosplit
//go:nocheckptr
//goland:noinspection GoUnusedParameter
func F64toa(out *byte, val float64) int {
	retc := C.f64toa(
		(*C.char)(unsafe.Pointer(out)),
		C.double(val),
	)
	return int(retc)
}

//go:nosplit
//go:nocheckptr
//goland:noinspection GoUnusedParameter
func I64toa(out *byte, val int64) int {
	retc := C.i64toa(
		(*C.char)(unsafe.Pointer(out)),
		C.int64_t(val),
	)
	return int(retc)
}

//go:nosplit
//go:nocheckptr
//goland:noinspection GoUnusedParameter
func U64toa(out *byte, val uint64) int {
	retc := C.u64toa(
		(*C.char)(unsafe.Pointer(out)),
		C.uint64_t(val),
	)
	return int(retc)
}

//go:nosplit
//go:nocheckptr
//goland:noinspection GoUnusedParameter
func Quote(s unsafe.Pointer, nb int, dp unsafe.Pointer, dn *int, flags uint64) int {
	retc := C.quote(
		(*C.char)(s),
		C.ssize_t(nb),
		(*C.char)(dp),
		(*C.ssize_t)(unsafe.Pointer(dn)),
		C.uint64_t(flags),
	)
	return int(retc)
}

//go:nosplit
//go:nocheckptr
//goland:noinspection GoUnusedParameter
func Unquote(s unsafe.Pointer, nb int, dp unsafe.Pointer, ep *int, flags uint64) int {
	retc := C.unquote(
		(*C.char)(s),
		C.ssize_t(nb),
		(*C.char)(dp),
		(*C.ssize_t)(unsafe.Pointer(ep)),
		C.uint64_t(flags),
	)
	return int(retc)
}

//go:nosplit
//go:nocheckptr
//goland:noinspection GoUnusedParameter
func J2T_FSM(fsm *types.J2TStateMachine, buf *[]byte, src *string, flag uint64) uint64 {
	retc := C.j2t_fsm_exec(
		(*C.J2TStateMachine)(unsafe.Pointer(fsm)),
		(*C._GoSlice)(unsafe.Pointer(buf)),
		(*C._GoString)(unsafe.Pointer(src)),
		C.uint64_t(flag),
	)
	return uint64(retc)
}

//go:nosplit
//go:nocheckptr
//goland:noinspection GoUnusedParameter
func TBSkip(st *types.TStateMachine, s *byte, n int, t uint8) int {
	retc := C.tb_skip(
		(*C.skipbuf_t)(unsafe.Pointer(st)),
		(*C.char)(unsafe.Pointer(s)),
		C.int64_t(n),
		C.uint8_t(t),
	)
	return int(retc)
}
