package j2p

import (
	"context"
	"fmt"

	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
)

const (
	_GUARD_SLICE_FACTOR = 1
)

func (self *BinaryConv) do(ctx context.Context, src []byte, desc *proto.Descriptor, buf *[]byte, req http.RequestGetter) error {
	//NOTICE: output buffer must be larger than src buffer
	rt.GuardSlice(buf, len(src)*_GUARD_SLICE_FACTOR)
	if err := self.Unmarshal(src, buf, desc); err != nil {
		return meta.NewError(meta.ErrConvert, fmt.Sprintf("json convert to protobuf failed, MessageDescriptor: %v", (*desc).Name()), err)
	}
	return nil
}

func (self *BinaryConv) Unmarshal(src []byte, out *[]byte, desc *proto.Descriptor) error {
	// use sonic to decode json bytes, get visitorUserNode
	vu := NewVisitorUserNodeBuffer()
	// use Visitor onxxx() to decode json2pb
	data, err := vu.Decode(src, desc)
	FreeVisitorUserNodePool(vu)
	if err != nil {
		return newError(meta.ErrConvert, "sonic decode json bytes failed", err)
	}
	*out = data
	return nil
}
