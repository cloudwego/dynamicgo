package j2p

import "github.com/cloudwego/dynamicgo/meta"

func newError(code meta.ErrCode, msg string, err error) error {
	return meta.NewError(meta.NewErrorCode(code, meta.JSON2PROTOBUF), msg, err)
}
