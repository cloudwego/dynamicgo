package proto

import (
	"context"
	"fmt"

	"github.com/cloudwego/dynamicgo/internal/util_test"
)

// GetFnDescFromFile get a fucntion descriptor from idl path (relative to your git root) and
// the function name
func GetFnDescFromFile(filePath, fnName string, opts Options, includeDirs ...string) *MethodDescriptor {
	svc, err := opts.NewDescriptorFromPath(context.Background(), util_test.MustGitPath(filePath), includeDirs...)
	if err != nil {
		panic(fmt.Errorf("%s:%s", util_test.MustGitPath(filePath), err))
	}
	fn := svc.LookupMethodByName(fnName)
	if fn == nil {
		return nil
	}
	return fn
}

// FnRequest get the normal requestDescriptor
func FnRequest(fn *MethodDescriptor) *TypeDescriptor {
	request := fn.Input()
	if request == nil {
		return nil
	}
	return request
}

// FnResponse get hte normal responseDescriptor
func FnResponse(fn *MethodDescriptor) *TypeDescriptor {
	response := fn.Output()
	if response == nil {
		return nil
	}
	return response
}
