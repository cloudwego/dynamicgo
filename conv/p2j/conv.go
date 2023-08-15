package p2j

import (
	"context"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
)

type BinaryConv2 struct {
	opts conv.Options
}

// NewBinaryConv returns a new BinaryConv
func NewBinaryConv(opts conv.Options) BinaryConv2 {
	return BinaryConv2{opts: opts}
}

// SetOptions sets options
func (self *BinaryConv2) SetOptions(opts conv.Options) {
	self.opts = opts
}

// Do converts protobuf binary (tbytes) to json bytes (jbytes)
//
// desc is the thrift type descriptor of the protobuf binary, usually it is a response STRUCT type
// ctx is the context, which can be used to pass arguments as below:
//   - conv.CtxKeyHTTPResponse: http.ResponseSetter as http request
//   - conv.CtxKeyThriftRespBase: thrift.Base as base metadata of thrift response
func (self *BinaryConv2) Do(ctx context.Context, desc *proto.MessageDescriptor, tbytes []byte) (json []byte, err error) {
	buf := conv.NewBytes()

	var resp http.ResponseSetter
	if self.opts.EnableHttpMapping {
		respi := ctx.Value(conv.CtxKeyHTTPResponse)
		if respi != nil {
			respi, ok := respi.(http.ResponseSetter)
			if !ok {
				return nil, wrapError(meta.ErrInvalidParam, "invalid http.Response", nil)
			}
			resp = respi
		} else {
			return nil, wrapError(meta.ErrInvalidParam, "no http response in context", nil)
		}
	}

	err = self.do(ctx, tbytes, desc, buf, resp)
	if err == nil && len(*buf) > 0 {
		json = make([]byte, len(*buf))
		copy(json, *buf)
	}

	conv.FreeBytes(buf)
	return
}
