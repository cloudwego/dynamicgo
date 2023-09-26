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

func (self *BinaryConv) do(ctx context.Context, src []byte, desc *proto.MessageDescriptor, buf *[]byte, req http.RequestGetter) error {
	// output buffer must be larger than src buffer
	rt.GuardSlice(buf, len(src)*_GUARD_SLICE_FACTOR)
	if len(src) == 0 {
		// empty obj
	}
	p := binary.NewBinaryProtocolBuffer()
	if err := self.Unmarshal(src, p, desc); err != nil {
		return meta.NewError(meta.ErrConvert, fmt.Sprintf("json convert to protobuf failed, MessageDescriptor: %v", (*desc).Name()), err)
	}
	*buf = p.RawBuf()
	return nil
}

func (self *BinaryConv) Unmarshal(src []byte, p *binary.BinaryProtocol, desc *proto.MessageDescriptor) error {
	// use sonic to decode json bytes, get visitorUserNode
	var d visitorUserNodeVisitorDecoder

	d.Reset()
	root, err := d.Decode(string(src))
	if err != nil {
		return newError(meta.ErrConvert, "sonic decode json bytes failed", err)
	}
	rootDesc := (*desc).(proto.Descriptor)
	parseUserNodeRecursive(root, &rootDesc, p, self, 0)
	return nil
}
