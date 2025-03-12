/**
 * Copyright 2023 CloudWeGo Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package thrift

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/cloudwego/thriftgo/generator/golang/streaming"

	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/json"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/internal/util"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/semantic"
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

	// ServiceName indicates which idl service to be parsed.
	ServiceName string

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

	// PutNameSpaceToAnnotation indicates to extract the name-space of one type
	// and put it on the type's annotation. The annotion format is:
	//   - Key: "thrift.name_space" (== NameSpaceAnnotationKey)
	//   - Values: pairs of Language and Name. for example:
	//      `namespace go base` will got ["go", "base"]
	// NOTICE: at present, only StructDescriptor.Annotations() can get this
	PutNameSpaceToAnnotation bool

	// PutThriftFilenameToAnnotation indicates to extract the filename of one type
	// and put it on the type's annotation. The annotion format is:
	//   - Key: "thrift.filename" (== FilenameAnnotationKey)
	//   - Values: pairs of Language and Name. for example:
	//      `// path := /a/b/c.thrift` will got ["/a/b/c.thrift"]
	// NOTICE: at present, only StructDescriptor.Annotations() can get this
	PutThriftFilenameToAnnotation bool

	// ApiBodyFastPath indicates `api.body` will change alias-name of root field, which can avoid search http-body on them
	ApiBodyFastPath bool
}

// NewDefaultOptions creates a default Options.
func NewDefaultOptions() Options {
	return Options{}
}

// path := /a/b/c.thrift
// includePath := ../d.thrift
// result := /a/d.thrift
func absPath(path, includePath string) string {
	if filepath.IsAbs(includePath) {
		return includePath
	}
	return filepath.Join(filepath.Dir(path), includePath)
}

// NewDescritorFromPath behaviors like NewDescritorFromPath, besides it uses DefaultOptions.
func NewDescritorFromPath(ctx context.Context, path string, includeDirs ...string) (*ServiceDescriptor, error) {
	return NewDefaultOptions().NewDescritorFromPath(ctx, path, includeDirs...)
}

// NewDescritorFromContent creates a ServiceDescriptor from a thrift path and its includes, which uses the given options.
// The includeDirs is used to find the include files.
func (opts Options) NewDescritorFromPath(ctx context.Context, path string, includeDirs ...string) (*ServiceDescriptor, error) {
	tree, err := parser.ParseFile(path, includeDirs, true)
	if err != nil {
		return nil, err
	}
	svc, err := parse(ctx, tree, opts.ParseServiceMode, opts)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

// NewDescritorFromContent creates a ServiceDescriptor from a thrift path and its includes, with specific methods.
// If methods is empty, all methods will be parsed.
// The includeDirs is used to find the include files.
func (opts Options) NewDescriptorFromPathWithMethod(ctx context.Context, path string, includeDirs []string, methods ...string) (*ServiceDescriptor, error) {
	tree, err := parser.ParseFile(path, includeDirs, true)
	if err != nil {
		return nil, err
	}
	svc, err := parse(ctx, tree, opts.ParseServiceMode, opts, methods...)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

// NewDescritorFromContent behaviors like NewDescritorFromPath, besides it uses DefaultOptions.
func NewDescritorFromContent(ctx context.Context, path, content string, includes map[string]string, isAbsIncludePath bool) (*ServiceDescriptor, error) {
	return NewDefaultOptions().NewDescritorFromContent(ctx, path, content, includes, isAbsIncludePath)
}

// NewDescritorFromContent creates a ServiceDescriptor from a thrift content and its includes, which uses the default options.
// path is the main thrift file path, content is the main thrift file content.
// includes is the thrift file content map, and its keys are specific including thrift file path.
// isAbsIncludePath argument has become obsolete. Regardless of whether its value is true or false, both absolute path and relative path will be searched.
func (opts Options) NewDescritorFromContent(ctx context.Context, path, content string, includes map[string]string, isAbsIncludePath bool) (*ServiceDescriptor, error) {
	tree, err := parseIDLContent(path, content, includes)
	if err != nil {
		return nil, err
	}
	svc, err := parse(ctx, tree, opts.ParseServiceMode, opts)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

// NewDescritorFromContentWithMethod creates a ServiceDescriptor from a thrift content and its includes, but only parse specific methods.
func (opts Options) NewDescriptorFromContentWithMethod(ctx context.Context, path, content string, includes map[string]string, isAbsIncludePath bool, methods ...string) (*ServiceDescriptor, error) {
	tree, err := parseIDLContent(path, content, includes)
	if err != nil {
		return nil, err
	}
	svc, err := parse(ctx, tree, opts.ParseServiceMode, opts, methods...)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func parseIDLContent(path, content string, includes map[string]string) (*parser.Thrift, error) {
	tree, err := parser.ParseString(path, content)
	if err != nil {
		return nil, err
	}
	_includes := make(map[string]*parser.Thrift, len(includes))
	for k, v := range includes {
		t, err := parser.ParseString(k, v)
		if err != nil {
			return nil, err
		}
		_includes[k] = t
	}

	done := map[string]*parser.Thrift{path: tree}
	if err := refIncludes(tree, path, done, _includes); err != nil {
		return nil, err
	}
	return tree, nil
}

func refIncludes(tree *parser.Thrift, path string, done map[string]*parser.Thrift, includes map[string]*parser.Thrift) error {
	done[path] = tree
	for _, i := range tree.Includes {
		paths := []string{absPath(tree.Filename, i.Path), i.Path}

		for _, p := range paths {
			// check cycle reference
			if t := done[p]; t != nil {
				i.Reference = t
				break
			}

			if ref, ok := includes[p]; ok {
				if err := refIncludes(ref, p, done, includes); err != nil {
					return err
				}
				i.Reference = ref
			}
		}

		if i.Reference == nil {
			return fmt.Errorf("miss include path: %s for file: %s", i.Path, tree.Filename)
		}
	}
	return nil
}

// Parse descriptor from parser.Thrift
func parse(ctx context.Context, tree *parser.Thrift, mode meta.ParseServiceMode, opts Options, methods ...string) (*ServiceDescriptor, error) {
	if len(tree.Services) == 0 {
		return nil, errors.New("empty serverce from idls")
	}
	if err := semantic.ResolveSymbols(tree); err != nil {
		return nil, err
	}

	sDsc := &ServiceDescriptor{
		functions:   map[string]*FunctionDescriptor{},
		annotations: []parser.Annotation{},
	}

	structsCache := compilingCache{}

	// support one service
	svcs := tree.Services

	// if an idl service name is specified, it takes precedence over parse mode
	if opts.ServiceName != "" {
		var err error
		svcs, err = getTargetService(svcs, opts.ServiceName)
		if err != nil {
			return nil, err
		}
		sDsc.name = opts.ServiceName
	} else {
		switch mode {
		case meta.LastServiceOnly:
			svcs = svcs[len(svcs)-1:]
			sDsc.name = svcs[len(svcs)-1].Name
		case meta.FirstServiceOnly:
			svcs = svcs[:1]
			sDsc.name = svcs[0].Name
		case meta.CombineServices:
			sDsc.name = "CombinedServices"
		}
	}

	for _, svc := range svcs {
		sopts := opts
		// pass origin annotations
		sDsc.annotations = copyAnnotationValues(svc.Annotations)
		// handle thrid-party annotations
		anns, _, next, err := mapAnnotations(ctx, svc.Annotations, AnnoScopeService, svc, sopts)
		if err != nil {
			return nil, err
		}
		for _, p := range anns {
			if err := handleAnnotation(ctx, AnnoScopeService, p.inter, p.cont, &sopts, svcs); err != nil {
				return nil, err
			}
		}

		funcs := make([]funcTreePair, 0, len(svc.Functions))
		getAllFuncs(svc, tree, &funcs)
		if len(methods) > 0 {
			funcs = findFuncs(funcs, methods)
		}
		for _, p := range funcs {
			injectAnnotations((*[]*parser.Annotation)(&p.fn.Annotations), next)
			if err := addFunction(ctx, p.fn, p.tree, sDsc, structsCache, sopts); err != nil {
				return nil, err
			}
		}

	}
	return sDsc, nil
}

func getTargetService(svcs []*parser.Service, serviceName string) ([]*parser.Service, error) {
	for _, svc := range svcs {
		if svc.Name == serviceName {
			return []*parser.Service{svc}, nil
		}
	}
	return nil, fmt.Errorf("the idl service name %s is not in the idl. Please check your idl", serviceName)
}

type funcTreePair struct {
	tree *parser.Thrift
	fn   *parser.Function
}

func findFuncs(funcs []funcTreePair, methods []string) (ret []funcTreePair) {
	for _, m := range methods {
		for _, p := range funcs {
			if p.fn.Name == m {
				ret = append(ret, p)
			}
		}
	}
	return
}

func getAllFuncs(svc *parser.Service, tree *parser.Thrift, ret *[]funcTreePair) {
	funcs := *ret
	n := len(svc.Functions)
	l := len(funcs)
	if cap(funcs) < l+n {
		tmp := make([]funcTreePair, l, l+n)
		copy(tmp, funcs)
		funcs = tmp
	}

	for _, fn := range svc.Functions {
		funcs = append(funcs, funcTreePair{
			tree: tree,
			fn:   fn,
		})
	}

	if svc.Extends != "" {
		ref := svc.GetReference()
		if ref != nil {
			idx := ref.GetIndex()
			name := ref.GetName()
			subTree := tree.Includes[idx].Reference
			sub, _ := subTree.GetService(name)
			if sub != nil {
				getAllFuncs(sub, subTree, &funcs)
			}
		}
	}
	*ret = funcs
	return
}

func addFunction(ctx context.Context, fn *parser.Function, tree *parser.Thrift, sDsc *ServiceDescriptor, structsCache compilingCache, opts Options) error {
	// for fuzzing test
	if opts.ParseFieldRandomRate > 0 {
		rand.Seed(time.Now().UnixNano())
	}

	if sDsc.functions[fn.Name] != nil {
		return fmt.Errorf("duplicate method name: %s", fn.Name)
	}
	if len(fn.Arguments) == 0 {
		return fmt.Errorf("empty arguments in function: %s", fn.Name)
	}

	// find all endpoints of this function
	enpdoints := []http.Endpoint{}
	// for http router
	for _, val := range fn.Annotations {
		method := http.AnnoToMethod(val.Key)
		if method != "" && len(val.Values) != 0 {
			// record the pair{method,path}
			enpdoints = append(enpdoints, http.Endpoint{Method: method, Path: val.Values[0]})
		}
	}

	annos := copyAnnotationValues(fn.Annotations)
	// handle thrid-party annotations
	anns, _, nextAnns, err := mapAnnotations(ctx, fn.Annotations, AnnoScopeFunction, fn, opts)
	if err != nil {
		return err
	}
	for _, p := range anns {
		// handle thrid-party annotations
		if err := handleAnnotation(ctx, AnnoScopeFunction, p.inter, p.cont, &opts, fn); err != nil {
			return err
		}

	}

	st, err := streaming.ParseStreaming(fn)
	if err != nil {
		return err
	}
	isStreaming := st.ClientStreaming || st.ServerStreaming

	var hasRequestBase bool
	var req *TypeDescriptor
	var resp *TypeDescriptor

	// parse request field
	if opts.ParseFunctionMode != meta.ParseResponseOnly {
		req, hasRequestBase, err = parseRequest(ctx, isStreaming, fn, tree, structsCache, nextAnns, opts)
		if err != nil {
			return err
		}
	}

	// parse response filed
	if opts.ParseFunctionMode != meta.ParseRequestOnly {
		resp, err = parseResponse(ctx, isStreaming, fn, tree, structsCache, nextAnns, opts)
		if err != nil {
			return err
		}
	}

	fnDsc := &FunctionDescriptor{
		name:              fn.Name,
		oneway:            fn.Oneway,
		request:           req,
		response:          resp,
		hasRequestBase:    hasRequestBase,
		endpoints:         enpdoints,
		annotations:       annos,
		isWithoutWrapping: isStreaming,
	}
	sDsc.functions[fn.Name] = fnDsc
	return nil
}

func parseRequest(ctx context.Context, isStreaming bool, fn *parser.Function, tree *parser.Thrift, structsCache compilingCache, nextAnns []parser.Annotation, opts Options) (req *TypeDescriptor, hasRequestBase bool, err error) {
	// WARN: only support single argument
	reqAst := fn.Arguments[0]
	reqType, err := parseType(ctx, reqAst.Type, tree, structsCache, 0, opts, nextAnns, Request)
	if err != nil {
		return nil, hasRequestBase, err
	}
	if reqType.Type() == STRUCT {
		for _, f := range reqType.Struct().names.All() {
			x := (*FieldDescriptor)(f.Val)
			if x.isRequestBase {
				hasRequestBase = true
				break
			}
		}
	}

	if isStreaming {
		return reqType, hasRequestBase, nil
	}

	// wrap with a struct
	wrappedTyDsc := &TypeDescriptor{
		typ: STRUCT,
		struc: &StructDescriptor{
			baseID:   FieldID(math.MaxUint16),
			ids:      util.FieldIDMap{},
			names:    util.FieldNameMap{},
			requires: make(RequiresBitmap, 1),
		},
	}
	reqField := &FieldDescriptor{
		name: reqAst.Name,
		id:   FieldID(reqAst.ID),
		typ:  reqType,
	}
	wrappedTyDsc.Struct().ids.Set(int32(reqAst.ID), unsafe.Pointer(reqField))
	wrappedTyDsc.Struct().names.Set(reqAst.Name, unsafe.Pointer(reqField))
	wrappedTyDsc.Struct().names.Build()
	return wrappedTyDsc, hasRequestBase, nil
}

func parseResponse(ctx context.Context, isStreaming bool, fn *parser.Function, tree *parser.Thrift, structsCache compilingCache, nextAnns []parser.Annotation, opts Options) (resp *TypeDescriptor, err error) {
	respAst := fn.FunctionType
	respType, err := parseType(ctx, respAst, tree, structsCache, 0, opts, nextAnns, Response)
	if err != nil {
		return nil, err
	}

	if isStreaming {
		return respType, nil
	}

	wrappedResp := &TypeDescriptor{
		typ: STRUCT,
		struc: &StructDescriptor{
			baseID:   FieldID(math.MaxUint16),
			ids:      util.FieldIDMap{},
			names:    util.FieldNameMap{},
			requires: make(RequiresBitmap, 1),
		},
	}
	respField := &FieldDescriptor{
		typ: respType,
	}
	wrappedResp.Struct().ids.Set(0, unsafe.Pointer(respField))
	// response has no name or id
	wrappedResp.Struct().names.Set("", unsafe.Pointer(respField))

	// parse exceptions
	if len(fn.Throws) > 0 {
		// only support single exception
		exp := fn.Throws[0]
		exceptionType, err := parseType(ctx, exp.Type, tree, structsCache, 0, opts, nextAnns, Exception)
		if err != nil {
			return nil, err
		}
		exceptionField := &FieldDescriptor{
			name:  exp.Name,
			alias: exp.Name,
			id:    FieldID(exp.ID),
			// isException: true,
			typ: exceptionType,
		}
		wrappedResp.Struct().ids.Set(int32(exp.ID), unsafe.Pointer(exceptionField))
		wrappedResp.Struct().names.Set(exp.Name, unsafe.Pointer(exceptionField))
	}
	wrappedResp.Struct().names.Build()
	return wrappedResp, nil
}

// reuse builtin types
var builtinTypes = map[string]*TypeDescriptor{
	"void":   {name: "void", typ: VOID, struc: new(StructDescriptor)},
	"bool":   {name: "bool", typ: BOOL},
	"byte":   {name: "byte", typ: BYTE},
	"i8":     {name: "i8", typ: I08},
	"i16":    {name: "i16", typ: I16},
	"i32":    {name: "i32", typ: I32},
	"i64":    {name: "i64", typ: I64},
	"double": {name: "double", typ: DOUBLE},
	"string": {name: "string", typ: STRING},
	"binary": {name: "binary", typ: STRING},
}

type compilingInstance struct {
	desc        *TypeDescriptor
	opts        Options
	parseTarget ParseTarget
}

type compilingCache map[string]*compilingInstance

// arg cache:
// only support self reference on the same file
// cross file self reference complicate matters
func parseType(ctx context.Context, t *parser.Type, tree *parser.Thrift, cache compilingCache, recursionDepth int, opts Options, nextAnns []parser.Annotation, parseTarget ParseTarget) (*TypeDescriptor, error) {
	if ty, ok := builtinTypes[t.Name]; ok {
		return ty, nil
	}

	nextRecursionDepth := recursionDepth + 1

	var err error
	switch t.Name {
	case "list":
		ty := &TypeDescriptor{name: t.Name}
		ty.typ = LIST
		ty.elem, err = parseType(ctx, t.ValueType, tree, cache, nextRecursionDepth, opts, nextAnns, parseTarget)
		return ty, err
	case "set":
		ty := &TypeDescriptor{name: t.Name}
		ty.typ = SET
		ty.elem, err = parseType(ctx, t.ValueType, tree, cache, nextRecursionDepth, opts, nextAnns, parseTarget)
		return ty, err
	case "map":
		ty := &TypeDescriptor{name: t.Name}
		ty.typ = MAP
		if ty.key, err = parseType(ctx, t.KeyType, tree, cache, nextRecursionDepth, opts, nextAnns, parseTarget); err != nil {
			return nil, err
		}
		ty.elem, err = parseType(ctx, t.ValueType, tree, cache, nextRecursionDepth, opts, nextAnns, parseTarget)
		return ty, err
	default:
		// check the cache
		if ty, ok := cache[t.Name]; ok && ty.parseTarget == parseTarget {
			return ty.desc, nil
		}

		// get type from AST tree
		typePkg, typeName := util.SplitSubfix(t.Name)
		if typePkg != "" {
			ref, ok := tree.GetReference(typePkg)
			if !ok {
				return nil, fmt.Errorf("miss reference: %s", typePkg)
			}
			tree = ref
			// cross file reference need empty cache
			cache = compilingCache{}
		}
		if typDef, ok := tree.GetTypedef(typeName); ok {
			return parseType(ctx, typDef.Type, tree, cache, nextRecursionDepth, opts, nextAnns, parseTarget)
		}
		if _, ok := tree.GetEnum(typeName); ok {
			if opts.ParseEnumAsInt64 {
				return builtinTypes["i64"], nil
			}
			return builtinTypes["i32"], nil
		}

		// handle STRUCT type
		var st *parser.StructLike
		var ok bool
		st, ok = tree.GetUnion(typeName)
		if !ok {
			st, ok = tree.GetStruct(typeName)
		}
		if !ok {
			st, ok = tree.GetException(typeName)
		}
		if !ok {
			return nil, fmt.Errorf("missing type: %s", typeName)
		}

		// copy original annotations
		oannos := copyAnnotationValues(st.Annotations)
		if opts.PutNameSpaceToAnnotation {
			oannos = append(oannos, extractNameSpaceToAnnos(tree))
		}

		if opts.PutThriftFilenameToAnnotation {
			oannos = append(oannos, extractThriftFilePathToAnnos(tree))
		}

		// inject previous annotations
		injectAnnotations((*[]*parser.Annotation)(&st.Annotations), nextAnns)

		// make struct descriptor
		ty := &TypeDescriptor{
			name: t.Name,
			typ:  STRUCT,
			struc: &StructDescriptor{
				baseID:      FieldID(math.MaxUint16),
				name:        typeName,
				ids:         util.FieldIDMap{},
				names:       util.FieldNameMap{},
				requires:    make(RequiresBitmap, len(st.Fields)),
				annotations: oannos,
			},
		}

		if recursionDepth == 0 {
			ctx = context.WithValue(ctx, CtxKeyIsBodyRoot, true)
		} else {
			ctx = context.WithValue(ctx, CtxKeyIsBodyRoot, false)
		}

		// handle thrid-party annotations for struct itself
		anns, _, nextAnns2, err := mapAnnotations(ctx, st.Annotations, AnnoScopeStruct, st, opts)
		if err != nil {
			return nil, err
		}
		for _, p := range anns {
			if err := handleAnnotation(ctx, AnnoScopeStruct, p.inter, p.cont, &opts, st); err != nil {
				return nil, err
			}
		}
		if st := ty.Struct(); st != nil {
			cache[t.Name] = &compilingInstance{parseTarget: parseTarget, desc: ty}
		}

		// parse fields
		for _, field := range st.Fields {
			// fork field-specific options
			fopts := opts
			if fopts.ParseFieldRandomRate > 0 && rand.Float64() > fopts.ParseFieldRandomRate {
				continue
			}
			var isRequestBase, isResponseBase bool
			if fopts.EnableThriftBase {
				isRequestBase = field.Type.Name == "base.Base" && recursionDepth == 0
				isResponseBase = field.Type.Name == "base.BaseResp" && recursionDepth == 0
			}
			// cannot cache the request base
			if isRequestBase {
				delete(cache, t.Name)
			}
			if isRequestBase || isResponseBase {
				ty.struc.baseID = FieldID(field.ID)
			}
			_f := &FieldDescriptor{
				isRequestBase:  isRequestBase,
				isResponseBase: isResponseBase,
				id:             FieldID(field.ID),
				name:           field.Name,
				alias:          field.Name,
				annotations:    copyAnnotationValues(field.Annotations),
			}

			// inject previous annotations
			injectAnnotations((*[]*parser.Annotation)(&field.Annotations), nextAnns2)
			anns, left, _, err := mapAnnotations(ctx, field.Annotations, AnnoScopeField, field, fopts)
			if err != nil {
				return nil, err
			}

			// handle annotations
			ignore := false
			for _, val := range left {
				skip, err := handleNativeFieldAnnotation(val, _f, parseTarget)
				if err != nil {
					return nil, err
				}
				if skip {
					ignore = true
					break
				}
			}
			if ignore {
				continue
			}
			for _, p := range anns {
				if err := handleFieldAnnotation(ctx, p.inter, p.cont, &fopts, _f, ty.struc, field); err != nil {
					return nil, err
				}
			}
			// copyAnnos(_f.annotations, anns)

			// recursively parse field type
			// WARN: options and annotations on field SHOULD NOT override these on their type definition
			if _f.typ, err = parseType(ctx, field.Type, tree, cache, nextRecursionDepth, opts, nil, parseTarget); err != nil {
				return nil, err
			}

			// make default value
			// WARN: ignore errors here
			if fopts.UseDefaultValue {
				dv, _ := makeDefaultValue(_f.typ, field.Default, tree)
				_f.defaultValue = dv
			}
			fp := unsafe.Pointer(_f)
			// set field id
			ty.Struct().ids.Set(int32(field.ID), fp)
			// set field requireness
			convertRequireness(field.Requiredness, ty.struc, _f, fopts)
			// set field key
			if fopts.MapFieldWay == meta.MapFieldUseAlias {
				ty.Struct().names.Set(_f.alias, fp)
			} else if fopts.MapFieldWay == meta.MapFieldUseFieldName {
				ty.Struct().names.Set(_f.name, fp)
			} else {
				ty.Struct().names.Set(_f.alias, fp)
				ty.Struct().names.Set(_f.name, fp)
			}

		}
		// buidl field name map
		ty.Struct().names.Build()
		return ty, nil
	}
}

func convertRequireness(r parser.FieldType, st *StructDescriptor, f *FieldDescriptor, opts Options) {
	var req Requireness
	switch r {
	case parser.FieldType_Default:
		f.required = DefaultRequireness
		if opts.SetOptionalBitmap {
			req = RequiredRequireness
		} else {
			req = DefaultRequireness
		}
	case parser.FieldType_Optional:
		f.required = OptionalRequireness
		if opts.SetOptionalBitmap {
			req = DefaultRequireness
		} else {
			req = OptionalRequireness
		}
	case parser.FieldType_Required:
		f.required = RequiredRequireness
		req = RequiredRequireness
	default:
		panic("invalid requireness type")
	}

	if f.isRequestBase || f.isResponseBase {
		// hence users who set EnableThriftBase can pass thrift base through conversion context, the field is always set optional
		req = OptionalRequireness
	}

	st.requires.Set(f.id, req)
}

func assertType(expected, but Type) error {
	if expected == but {
		return nil
	}
	return fmt.Errorf("need %s type, but got: %s", expected, but)
}

func makeDefaultValue(typ *TypeDescriptor, val *parser.ConstValue, tree *parser.Thrift) (*DefaultValue, error) {
	if val == nil {
		return nil, nil
	}
	switch val.Type {
	case parser.ConstType_ConstInt:
		if !typ.typ.IsInt() {
			return nil, fmt.Errorf("mismatched int default value with type %s", typ.name)
		}
		if x := val.TypedValue.Int; x != nil {
			v := int64(*x)
			p := BinaryProtocol{Buf: make([]byte, 0, 4)}
			p.WriteInt(typ.typ, int(v))
			jbuf := json.EncodeInt64(make([]byte, 0, 8), v)
			return &DefaultValue{
				goValue:      v,
				jsonValue:    rt.Mem2Str(jbuf),
				thriftBinary: rt.Mem2Str(p.Buf),
			}, nil
		}
	case parser.ConstType_ConstDouble:
		if typ.typ != DOUBLE {
			return nil, fmt.Errorf("mismatched double default value with type %s", typ.name)
		}
		if x := val.TypedValue.Double; x != nil {
			v := float64(*x)
			tbuf := make([]byte, 8)
			BinaryEncoding{}.EncodeDouble(tbuf, v)
			jbuf := json.EncodeFloat64(make([]byte, 0, 8), v)
			return &DefaultValue{
				goValue:      v,
				jsonValue:    rt.Mem2Str(jbuf),
				thriftBinary: rt.Mem2Str(tbuf),
			}, nil
		}
	case parser.ConstType_ConstLiteral:
		if typ.typ != STRING {
			return nil, fmt.Errorf("mismatched string default value with type %s", typ.name)
		}
		if x := val.TypedValue.Literal; x != nil {
			v := string(*x)
			tbuf := make([]byte, len(v)+4)
			BinaryEncoding{}.EncodeString(tbuf, v)
			jbuf := json.EncodeString(make([]byte, 0, len(v)+2), v)
			return &DefaultValue{
				goValue:      v,
				jsonValue:    rt.Mem2Str(jbuf),
				thriftBinary: rt.Mem2Str(tbuf),
			}, nil
		}
	case parser.ConstType_ConstIdentifier:
		x := *val.TypedValue.Identifier

		// try to handle it as bool
		if typ.typ == BOOL {
			if v := strings.ToLower(x); v == "true" {
				return &DefaultValue{
					goValue:      true,
					jsonValue:    "true",
					thriftBinary: string([]byte{0x01}),
				}, nil
			} else if v == "false" {
				return &DefaultValue{
					goValue:      false,
					jsonValue:    "false",
					thriftBinary: string([]byte{0x00}),
				}, nil
			}
		}

		// try to handle as const value
		var ctree = tree
		pkg, name := util.SplitSubfix(x)
		if pkg != "" {
			if nt, ok := tree.GetReference(pkg); ok && nt != nil {
				ctree = nt
			}
		}
		y, ok := ctree.GetConstant(name)
		if ok && y != nil {
			return makeDefaultValue(typ, y.Value, ctree)
		}

		// try to handle as enum value
		if pkg == "" {
			return nil, fmt.Errorf("not found identifier: %s", x)
		}
		emv := name
		emp, emt := util.SplitSubfix(pkg)
		if emp != "" {
			if nt, ok := tree.GetReference(emp); ok && nt != nil {
				tree = nt
			}
		}
		z, ok := tree.GetEnum(emt)
		if !ok || z == nil {
			return nil, fmt.Errorf("not found enum type: %s", emt)
		}
		if !typ.typ.IsInt() {
			return nil, fmt.Errorf("mismatched int default value with type %s", typ.name)
		}
		for _, vv := range z.Values {
			if vv.Name == emv {
				v := int64(vv.Value)
				p := BinaryProtocol{Buf: make([]byte, 0, 4)}
				p.WriteInt(typ.typ, int(v))
				jbuf := json.EncodeInt64(make([]byte, 0, 8), v)
				return &DefaultValue{
					goValue:      v,
					jsonValue:    rt.Mem2Str(jbuf),
					thriftBinary: rt.Mem2Str(p.Buf),
				}, nil
			}
		}

	}
	return nil, nil
}
