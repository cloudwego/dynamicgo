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

	MapFieldWay meta.MapFieldWay // not implemented.

	ParseFieldRandomRate float64 // not implemented.

	ParseEnumAsInt64 bool // not implemented.

	SetOptionalBitmap bool // not implemented.

	UseDefaultValue bool // not implemented.

	ParseFunctionMode meta.ParseFunctionMode // not implemented.

	EnableProtoBase bool // not implemented.
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
	importDirs = append(importDirs, path)
	fileDescriptors, err := protoparse.Parser{}.ParseFiles(importDirs...)
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

	return sDsc, nil
}
