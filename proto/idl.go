package proto

import (
	"context"
	"errors"

	"github.com/bufbuild/protocompile"
	"github.com/cloudwego/dynamicgo/meta"
)

const (
	Request ParseTarget = iota
	Response
	Exception
)
var com = protocompile.Compiler{}

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
	var fd FileDescriptor

	ImportPaths := []string{""} // default import "" when path is absolute path, no need to join with importDirs
	// append importDirs to ImportPaths
	ImportPaths = append(ImportPaths, importDirs...)
	compiler := protocompile.Compiler{
		Resolver: &protocompile.SourceResolver{
			ImportPaths: ImportPaths,
		},
		SourceInfoMode: protocompile.SourceInfoStandard,
	}

	results, err := compiler.Compile(ctx, path)
	if err != nil {
		return nil, err
	}
	fd = results[0]
	svc, err := parse(ctx, &fd, opts.ParseServiceMode, opts)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

// Parse descriptor from fileDescriptor
func parse(ctx context.Context, fileDesc *FileDescriptor, mode meta.ParseServiceMode, opts Options, methods ...string) (*ServiceDescriptor, error) {
	var sDsc *ServiceDescriptor
	svcs := (*fileDesc).Services()
	if svcs.Len() == 0 {
		return nil, errors.New("empty service from idls")
	}

	// support one service
	switch mode {
	case meta.LastServiceOnly:
		service := svcs.Get(svcs.Len() - 1)
		sDsc = &service
	case meta.FirstServiceOnly:
		service := svcs.Get(0)
		sDsc = &service
	}

	return sDsc, nil
}

func (opts Options) NewDesccriptorFromContent(ctx context.Context, filename string, includes map[string]string) (*ServiceDescriptor, error) {
	var fd FileDescriptor
	// change includes path to absolute path

	compiler := protocompile.Compiler{
		Resolver: &protocompile.SourceResolver{
			Accessor: protocompile.SourceAccessorFromMap(includes),
		},
	}

	result, err := compiler.Compile(ctx, filename)
	if err != nil {
		return nil, err
	}
	
	fd = result[0]
	sdsc, err := parse(ctx, &fd, opts.ParseServiceMode, opts)
	if err != nil {
		return nil, err
	}

	return sdsc, nil
}