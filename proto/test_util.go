package proto

import (
	"context"
	"fmt"

	"github.com/cloudwego/dynamicgo/internal/util_test"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// GetFnDescFromFile get a fucntion descriptor from idl path (relative to your git root) and
// the function name
func GetFnDescFromFile(filePath, fnName string, opts Options) *MethodDescriptor {
	svc, err := opts.NewDescriptorFromPath(context.Background(), util_test.MustGitPath(filePath))
	if err != nil {
		panic(fmt.Errorf("%s:%s", util_test.MustGitPath(filePath), err))
	}
	fn := (*svc).Methods().ByName(protoreflect.Name(fnName))
	if fn == nil {
		return nil
	}
	return &fn
}

// FnRequest get the normal request type
func FnRequest(fn *MethodDescriptor) *MessageDescriptor {
	request := (*fn).Input()
	if request == nil {
		return nil
	}
	return &request
}
