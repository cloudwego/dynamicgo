package j2p

import (
	"context"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
)

// BinaryConv is a converter from json to protobuf binary
type BinaryConv struct {
	opts conv.Options
}

// NewBinaryConv returns a new BinaryConv
func NewBinaryConv(opts conv.Options) BinaryConv {
	return BinaryConv{
		opts: opts,
	}
}

// Do converts json bytes (jbytes) to protobuf binary (tbytes)
// desc is the protobuf type descriptor of the protobuf binary, usually it the request Message type
func (self *BinaryConv) Do(ctx context.Context, desc *proto.TypeDescriptor, jbytes []byte) (tbytes []byte, err error) {
	buf := conv.NewBytes()

	// req alloc but not use
	var req http.RequestGetter
	if self.opts.EnableHttpMapping {
		reqi := ctx.Value(conv.CtxKeyHTTPRequest)
		if reqi != nil {
			reqi, ok := reqi.(http.RequestGetter)
			if !ok {
				return nil, newError(meta.ErrInvalidParam, "invalid http.RequestGetter", nil)
			}
			req = reqi
		} else {
			return nil, newError(meta.ErrInvalidParam, "EnableHttpMapping but no http response in context", nil)
		}
	}

	err = self.do(ctx, jbytes, desc, buf, req)
	if err == nil && len(*buf) > 0 {
		tbytes = make([]byte, len(*buf))
		copy(tbytes, *buf)
	}

	conv.FreeBytes(buf)
	return
}

// DoInto behaves like Do, but it writes the result to buffer directly instead of returning a new buffer
func (self *BinaryConv) DoInto(ctx context.Context, desc *proto.TypeDescriptor, jbytes []byte, buf *[]byte) error {
	// req alloc but not use
	var req http.RequestGetter
	if self.opts.EnableHttpMapping {
		reqi := ctx.Value(conv.CtxKeyHTTPRequest)
		if reqi != nil {
			reqi, ok := reqi.(http.RequestGetter)
			if !ok {
				return newError(meta.ErrInvalidParam, "invalid http.RequestGetter", nil)
			}
			req = reqi
		} else {
			return newError(meta.ErrInvalidParam, "EnableHttpMapping but no http response in context", nil)
		}
	}
	return self.do(ctx, jbytes, desc, buf, req)
}
