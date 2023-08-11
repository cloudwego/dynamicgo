package p2j

import "github.com/cloudwego/dynamicgo/meta"

func wrapError(code meta.ErrCode, msg string, err error) error {
	return meta.NewError(code, msg, err)
}
