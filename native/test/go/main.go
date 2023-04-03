package main

// /*
// #cgo CFLAGS: -I../..
// #cgo LDFLAGS: -L../../../output -lnative

// #include "thrift.h"
// #include "thrift_skip.c"
// */
// import "C"
// import (
// 	"fmt"
// 	"unsafe"

// 	"github.com/cloudwego/dynamicgo/internal/native/types"
// 	"github.com/cloudwego/dynamicgo/thrift"
// )

// func main() {
// 	var (
// 		j2t_fsm_exec = C.j2t_fsm_exec
// 		tb_skip      = C.tb_skip
// 	)
// 	var buf = [512]byte{}
// 	off := 0
// 	fsm := types.NewTStateMachine()
// 	C.tb_skip(
// 		(*C.skipbuf_t)(unsafe.Pointer(fsm)),
// 		(*C.char)(unsafe.Pointer(&buf[0])),
// 		C.long(off),
// 		C.uchar(thrift.STRING),
// 	)

// 	fmt.Printf("%+#v %+#v\n", j2t_fsm_exec, tb_skip)
// }

func main() {}
