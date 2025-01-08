package proto

import (
	"context"
	"errors"
	"math"
	"unsafe"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"

	"github.com/cloudwego/dynamicgo/internal/util"
	"github.com/cloudwego/dynamicgo/meta"
)

type compilingInstance struct {
	desc        *TypeDescriptor
	opts        Options
	parseTarget ParseTarget
}

type compilingCache map[string]*compilingInstance

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
	var pbParser protoparse.Parser
	ImportPaths := []string{""} // default import "" when path is absolute path, no need to join with importDirs
	// append importDirs to ImportPaths
	ImportPaths = append(ImportPaths, importDirs...)
	pbParser.ImportPaths = ImportPaths
	fds, err := pbParser.ParseFiles(path)
	if err != nil {
		return nil, err
	}
	fd := fds[0]
	svc, err := parse(ctx, fd, opts.ParseServiceMode, opts)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

// NewDescritorFromContent behaviors like NewDescritorFromPath, besides it uses DefaultOptions.
func NewDescritorFromContent(ctx context.Context, path, content string, includes map[string]string, importDirs ...string) (*ServiceDescriptor, error) {
	return NewDefaultOptions().NewDesccriptorFromContent(ctx, path, content, includes, importDirs...)
}

func (opts Options) NewDesccriptorFromContent(ctx context.Context, path, content string, includes map[string]string, importDirs ...string) (*ServiceDescriptor, error) {

	var pbParser protoparse.Parser
	// add main proto to includes
	includes[path] = content

	ImportPaths := []string{""} // default import "" when path is absolute path, no need to join with importDirs
	// append importDirs to ImportPaths
	ImportPaths = append(ImportPaths, importDirs...)

	pbParser.ImportPaths = ImportPaths
	pbParser.Accessor = protoparse.FileContentsFromMap(includes)
	fds, err := pbParser.ParseFiles(path)
	if err != nil {
		return nil, err
	}

	fd := fds[0]
	sdsc, err := parse(ctx, fd, opts.ParseServiceMode, opts)
	if err != nil {
		return nil, err
	}

	return sdsc, nil
}

// Parse descriptor from fileDescriptor
func parse(ctx context.Context, fileDesc *desc.FileDescriptor, mode meta.ParseServiceMode, opts Options, methods ...string) (*ServiceDescriptor, error) {
	svcs := fileDesc.GetServices()
	if len(svcs) == 0 {
		return nil, errors.New("empty service from idls")
	}

	sDsc := &ServiceDescriptor{
		methods:     map[string]*MethodDescriptor{},
		packageName: fileDesc.GetPackage(),
	}

	structsCache := compilingCache{}
	// support one service
	switch mode {
	case meta.LastServiceOnly:
		svcs = svcs[len(svcs)-1:]
		sDsc.serviceName = svcs[len(svcs)-1].GetName()
	case meta.FirstServiceOnly:
		svcs = svcs[:1]
		sDsc.serviceName = svcs[0].GetName()
	case meta.CombineServices:
		sDsc.serviceName = "CombinedService"
		sDsc.isCombinedServices = true
	}

	for _, svc := range svcs {
		for _, mtd := range svc.GetMethods() {
			var req *TypeDescriptor
			var resp *TypeDescriptor
			var err error

			req, err = parseMessage(ctx, mtd.GetInputType(), structsCache, 0, opts, Request)
			if err != nil {
				return nil, err
			}

			resp, err = parseMessage(ctx, mtd.GetOutputType(), structsCache, 0, opts, Response)
			if err != nil {
				return nil, err
			}

			sDsc.methods[mtd.GetName()] = &MethodDescriptor{
				name:              mtd.GetName(),
				input:             req,
				output:            resp,
				isClientStreaming: mtd.IsClientStreaming(),
				isServerStreaming: mtd.IsServerStreaming(),
			}

		}
	}
	return sDsc, nil
}

func parseMessage(ctx context.Context, msgDesc *desc.MessageDescriptor, cache compilingCache, recursionDepth int, opts Options, parseTarget ParseTarget) (*TypeDescriptor, error) {
	if tycache, ok := cache[msgDesc.GetName()]; ok && tycache.parseTarget == parseTarget {
		return tycache.desc, nil
	}

	var ty *TypeDescriptor
	var err error
	fields := msgDesc.GetFields()
	md := &MessageDescriptor{
		baseId: FieldNumber(math.MaxInt32),
		ids:    util.FieldIDMap{},
		names:  util.FieldNameMap{},
	}

	ty = &TypeDescriptor{
		typ:  MESSAGE,
		name: msgDesc.GetName(),
		msg:  md,
	}

	cache[ty.name] = &compilingInstance{
		desc:        ty,
		opts:        opts,
		parseTarget: parseTarget,
	}

	for _, field := range fields {
		descpbType := field.GetType()
		id := field.GetNumber()
		name := field.GetName()
		jsonName := field.GetJSONName()
		fieldDesc := &FieldDescriptor{
			id:       FieldNumber(id),
			name:     name,
			jsonName: jsonName,
		}

		// MAP TypeDescriptor
		if field.IsMap() {
			kt := builtinTypes[field.GetMapKeyType().GetType()]
			vt := builtinTypes[field.GetMapValueType().GetType()]
			if vt.Type() == MESSAGE {
				vt, err = parseMessage(ctx, field.GetMapValueType().GetMessageType(), cache, recursionDepth+1, opts, parseTarget)
				if err != nil {
					return nil, err
				}
			}

			mapt, err := parseMessage(ctx, field.GetMessageType(), cache, recursionDepth+1, opts, parseTarget)
			if err != nil {
				return nil, err
			}

			fieldDesc.typ = &TypeDescriptor{
				typ:    MAP,
				name:   name,
				key:    kt,
				elem:   vt,
				baseId: FieldNumber(id),
				msg:    mapt.msg,
			}
			fieldDesc.kind = MessageKind
		} else {
			// basic type or message TypeDescriptor
			t := builtinTypes[descpbType]
			if t.Type() == MESSAGE {
				t, err = parseMessage(ctx, field.GetMessageType(), cache, recursionDepth+1, opts, parseTarget)
				if err != nil {
					return nil, err
				}
			}
			fieldDesc.kind = t.typ.TypeToKind()

			// LIST TypeDescriptor
			if field.IsRepeated() {
				t = &TypeDescriptor{
					typ:    LIST,
					name:   name,
					elem:   t,
					baseId: FieldNumber(id),
					msg:    t.msg,
				}
			}
			fieldDesc.typ = t
		}

		// add fieldDescriptor to MessageDescriptor
		// md.ids[FieldNumber(id)] = fieldDesc
		md.ids.Set(int32(id), unsafe.Pointer(fieldDesc))
		md.names.Set(name, unsafe.Pointer(fieldDesc))
		md.names.Set(jsonName, unsafe.Pointer(fieldDesc))
	}
	md.names.Build()

	return ty, nil
}
