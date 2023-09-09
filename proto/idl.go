package proto

import (
	"context"
	"errors"

	"github.com/cloudwego/dynamicgo/meta"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
)

const (
	Request ParseTarget = iota
	Response
	Exception
)

// ParseTarget indicates the target to parse
type ParseTarget uint8

// Options is options for parsing thrift IDL.
type Options struct {
	// ParseServiceMode indicates how to parse service.
	ParseServiceMode meta.ParseServiceMode

	// MapFieldWay indicates StructDescriptor.FieldByKey() uses alias to map field.
	// By default, we use alias to map, and alias always equals to field name if not given.
	MapFieldWay meta.MapFieldWay

	// ParseFieldRandomRate indicates whether to parse partial fields and is only used for mock test.
	// The value means the possibility of randomly parse and embed one field into StructDescriptor.
	// It must be within (0, 1], and 0 means always parse all fields.
	ParseFieldRandomRate float64

	// ParseEnumAsInt64 indicates whether to parse enum as I64 (default I32).
	ParseEnumAsInt64 bool

	// SetOptionalBitmap indicates to set bitmap for optional fields
	SetOptionalBitmap bool

	// UseDefaultValue indicates to parse and store default value defined on IDL fields.
	UseDefaultValue bool

	// ParseFunctionMode indicates to parse only response or request for a function
	ParseFunctionMode meta.ParseFunctionMode

	// EnableThriftBase indicates to explictly handle thrift/base (see README.md) fields.
	// One field is identified as a thrift base if it satisfies **BOTH** of the following conditions:
	//   1. Its type is 'base.Base' (for request base) or 'base.BaseResp' (for response base);
	//   2. it is on the top layer of the root struct of one function.
	EnableThriftBase bool
}

// NewDefaultOptions creates a default Options.
func NewDefaultOptions() Options {
	return Options{}
}

// NewDescritorFromPath behaviors like NewDescritorFromPath, besides it uses DefaultOptions.
func NewDescritorFromPath(ctx context.Context, path string, importDirs ...string) (*ServiceDescriptor, error) {
	return NewDefaultOptions().NewDescriptorFromPath(ctx, path, importDirs...)
}

// NewDescritorFromContent creates a ServiceDescriptor from a proto path and its imports, which uses the given options.
// The importDirs is used to find the include files.
func (opts Options) NewDescriptorFromPath(ctx context.Context, path string, importDirs ...string) (*ServiceDescriptor, error) {
	fileDescriptors, err := protoparse.Parser{}.ParseFiles(path)
	if err != nil {
		return nil, err
	}
	svc, err := parse(ctx, fileDescriptors[0], opts.ParseServiceMode, opts)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

// Parse descriptor from fileDescriptor
func parse(ctx context.Context, fileDesc *desc.FileDescriptor, mode meta.ParseServiceMode, opts Options, methods ...string) (*ServiceDescriptor, error) {
	var sDsc *ServiceDescriptor
	svcs := fileDesc.GetServices()
	if len(svcs) == 0 {
		return nil, errors.New("empty service from idls")
	}

	// support one service
	switch mode {
	case meta.LastServiceOnly:
		svcs = svcs[len(svcs)-1:]
		svcsPtr, _ := (svcs[0].Unwrap()).(ServiceDescriptor)
		sDsc = &svcsPtr
	case meta.FirstServiceOnly:
		svcs = svcs[:1]
		svcsPtr, _ := (svcs[0].Unwrap()).(ServiceDescriptor)
		sDsc = &svcsPtr
	}

	// cope with annotations
	// ....

	return sDsc, nil
}
