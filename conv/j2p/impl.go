package j2p

import (
	"context"
	"fmt"

	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
)

const (
	_GUARD_SLICE_FACTOR = 1
)

func (self *BinaryConv) do(ctx context.Context, src []byte, desc *proto.TypeDescriptor, buf *[]byte, req http.RequestGetter) error {
	//NOTICE: output buffer must be larger than src buffer
	rt.GuardSlice(buf, len(src)*_GUARD_SLICE_FACTOR)
	if err := self.unmarshal(src, buf, desc); err != nil {
		return meta.NewError(meta.ErrConvert, fmt.Sprintf("json convert to protobuf failed, MessageDescriptor: %v", desc.Message().Name()), err)
	}
	return nil
}

func (self *BinaryConv) unmarshal(src []byte, out *[]byte, desc *proto.TypeDescriptor) error {
	// use sonic to decode json bytes, get visitorUserNode
	vu := newVisitorUserNodeBuffer()
	vu.opts = &self.opts
	vu.p = binary.NewBinaryProtocolBuffer()
	// use Visitor onxxx() to decode json2pb
	data, err := vu.decode(src, desc)
	freeVisitorUserNodePool(vu)
	if err != nil {
		return err
	}
	*out = data
	return nil
}
