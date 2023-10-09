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

// SetOptions sets options
func (self *BinaryConv) SetOptions(opts conv.Options) {
	self.opts = opts
}

// Do converts json bytes (jbytes) to protobuf binary (tbytes)
//
// desc is the protobuf type descriptor of the protobuf binary, usually it the request STRUCT type
// ctx is the context, which can be used to pass arguments as below:
//   - conv.CtxKeyHTTPRequest: http.RequestGetter as http request
//   - conv.CtxKeyThriftRespBase: protobuf.Base as base metadata of protobuf response
func (self *BinaryConv) Do(ctx context.Context, desc *proto.MessageDescriptor, jbytes []byte) (tbytes []byte, err error) {
	buf := conv.NewBytes()

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

func (self *BinaryConv) DoInto(ctx context.Context, desc *proto.MessageDescriptor, jbytes []byte, buf *[]byte) error {
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
