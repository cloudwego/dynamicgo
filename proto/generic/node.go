package generic

import (
	"unsafe"

	"github.com/cloudwego/dynamicgo/proto"
)

type Node struct {
	t  proto.Type
	et proto.Type
	kt proto.Type
	v  unsafe.Pointer
	l  int
}

