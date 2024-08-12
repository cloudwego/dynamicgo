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
	"fmt"
	"strings"

	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/thriftgo/parser"
)

//--------------------------------- Annotation Interface ----------------------------------

// Annotation idl annotation interface
type Annotation interface {
	// unique id of the Annotation
	ID() AnnoID

	// Make makes the handler function under specific values and idl description
	//
	// desc is specific type to its registered AnnoScope:
	//   AnnoScopeService: desc is *parser.Service
	//   AnnoScopeFunction: desc is *parser.Function
	//   AnnoScopeStruct: desc is *parser.StructLike
	//   AnnoScopeField: desc is *parser.Field
	//
	// returned handler SHOULD BE one interface according to its AnnoKind:
	//   AnnoKindHttpMappping: HTTPMapping interface
	//   AnnoKindKeyMapping: KeyMapping interface
	//   AnnoKindKeyMapping: ValueMapping interface
	//   AnnoKindOptionMapping: OptionMapping interface
	Make(ctx context.Context, values []parser.Annotation, desc interface{}) (handler interface{}, err error)
}

// AnnoID is the unique id of an annotation, which is composed of kind, scope and type:
//
//	0xff000000: AnnoKind
//	0x00ff0000: AnnoScope
//	0x0000ffff: AnnoType
type AnnoID uint32

func MakeAnnoID(kind AnnoKind, scope AnnoScope, typ AnnoType) AnnoID {
	return AnnoID((uint32(kind) << 24) | (uint32(scope) << 16) | uint32(typ))
}

// AnnoKind is the kind of annotation,
// which defines the result handler of Annotation.Make()
type AnnoKind uint8

const (
	// AnnoKindHttpMappping is the kind of http mapping annotation
	// These annotations Make() will return HTTPMapping
	AnnoKindHttpMappping AnnoKind = iota + 1

	// AnnotationKindKeyMapping is the kind of key mapping annotation
	// These annotations Make() will return KeyMapping
	AnnoKindValueMapping

	// AnnotationKindValueMapping is the kind of value mapping annotation
	// These annotations Make() will return ValueMapping
	AnnoKindOptionMapping

	// AnnotationKindOptionMapping is the kind of option mapping annotation
	// These annotations Make() will return OptionMapping
	AnnoKindKeyMapping
)

// Kind returns the kind of the annotation
func (t AnnoID) Kind() AnnoKind {
	return AnnoKind(t >> 24)
}

// AnnoScope is effective scope of annotation
type AnnoScope uint8

const (
	// AnnoScopeService works on service description
	AnnoScopeService AnnoScope = iota + 1

	// AnnoScopeFunction works on function description
	AnnoScopeFunction

	// AnnoScopeStruct works on struct description
	AnnoScopeStruct

	// AnnoScopeField works on field description
	AnnoScopeField
)

// Scope returns the scope of the annotation
func (t AnnoID) Scope() AnnoScope {
	return AnnoScope((t >> 16) & (0x00ff))
}

// AnnoType is the specific type of an annotation
type AnnoType uint16

// Type returns the type of the annotation
func (t AnnoID) Type() AnnoType {
	return AnnoType(t & 0xffff)
}

// OptionMapping is used to convert thrift.Options while parsing idl.
// See also: thrift/annotation/option_mapping.go
type OptionMapping interface {
	// Map options to new options
	Map(ctx context.Context, opts Options) Options
}

// ValueMapping is used to convert thrift value while running convertion.
// See also: thrift/annotation/value_mapping.go
type ValueMapping interface {
	// Read thrift value from p and convert it into out
	Read(ctx context.Context, p *BinaryProtocol, field *FieldDescriptor, out *[]byte) error

	// Write thrift value into p, which is converted from in
	Write(ctx context.Context, p *BinaryProtocol, field *FieldDescriptor, in []byte) error
}

// HTTPMapping is used to convert http value while running convertion.
// See also: thrift/annotation/http_mapping.go
type HttpMapping interface {
	// Request get a http value from req
	Request(ctx context.Context, req http.RequestGetter, field *FieldDescriptor) (string, error)

	// Response set a http value into resp
	Response(ctx context.Context, resp http.ResponseSetter, field *FieldDescriptor, val string) error

	// RawEncoding indicates the encoding of the value, it should be meta.EncodingText by default
	Encoding() meta.Encoding
}

// KeyMapping is used to convert field key while parsing idl.
// See also: thrift/annotation/key_mapping.go
type KeyMapping interface {
	// Map key to new key
	Map(ctx context.Context, key string) string
}

//--------------------------------- Thrid-party Register ----------------------------------

var annotations = map[string]map[AnnoScope]Annotation{}

// RegisterAnnotation register an annotation on specific AnnoScope
func RegisterAnnotation(an Annotation, keys ...string) {
	for _, key := range keys {
		key = strings.ToLower(key)
		m := annotations[key]
		if m == nil {
			m = make(map[AnnoScope]Annotation)
			annotations[key] = m
		}
		m[an.ID().Scope()] = an
	}
}

func FindAnnotation(key string, scope AnnoScope) Annotation {
	key = strings.ToLower(key)
	m := annotations[key]
	if m == nil {
		return nil
	}
	return m[scope]
}

// makeAnnotation search an annotation by given key+scope and make its handler
func makeAnnotation(ctx context.Context, anns []parser.Annotation, scope AnnoScope, desc interface{}) (interface{}, AnnoID, error) {
	if len(anns) == 0 {
		return nil, 0, nil
	}
	ann := FindAnnotation(anns[0].Key, scope)
	if ann == nil {
		return nil, 0, nil
	}
	if handler, err := ann.Make(ctx, anns, desc); err != nil {
		return nil, 0, err
	} else {
		return handler, ann.ID(), nil
	}
}

// AnnotationMapper is used to convert a annotation to equivalent annotations
// desc is specific to its registered AnnoScope:
//
//	AnnoScopeService: desc is *parser.Service
//	AnnoScopeFunction: desc is *parser.Function
//	AnnoScopeStruct: desc is *parser.StructLike
//	AnnoScopeField: desc is *parser.Field
type AnnotationMapper interface {
	// Map map a annotation to equivalent annotations
	Map(ctx context.Context, ann []parser.Annotation, desc interface{}, opt Options) (cur []parser.Annotation, next []parser.Annotation, err error)
}

var annotationMapper = map[string]map[AnnoScope]AnnotationMapper{}

// RegisterAnnotationMapper register a annotation mapper on specific scope
func RegisterAnnotationMapper(scope AnnoScope, mapper AnnotationMapper, keys ...string) {
	for _, key := range keys {
		m := annotationMapper[key]
		if m == nil {
			m = make(map[AnnoScope]AnnotationMapper)
			annotationMapper[key] = m
		}
		m[scope] = mapper
	}
}

func FindAnnotationMapper(key string, scope AnnoScope) AnnotationMapper {
	m := annotationMapper[key]
	if m == nil {
		return nil
	}
	return m[scope]
}

func RemoveAnnotationMapper(scope AnnoScope, keys ...string) {
	for _, key := range keys {
		m := annotationMapper[key]
		if m != nil {
			if _, ok := m[scope]; ok {
				delete(m, scope)
			}
		}
	}
}

//------------------------------- IDL processing logic -------------------------------

const (
	// AnnoKeyDynamicGoDeprecated is used to mark a description as deprecated
	AnnoKeyDynamicGoDeprecated = "dynamicgo.deprecated"
	// AnnoKeyDynamicGoApiNone is used to deal with http response field with api.none annotation
	AnnoKeyDynamicGoApiNone = "api.none"
)

type amPair struct {
	cont  []parser.Annotation
	inter AnnotationMapper
}

type amPairs []amPair

func (p *amPairs) Add(con parser.Annotation, inter AnnotationMapper) {
	for i := range *p {
		if (*p)[i].inter == inter {
			(*p)[i].cont = append((*p)[i].cont, con)
			return
		}
	}
	*p = append(*p, amPair{
		cont:  []parser.Annotation{con},
		inter: inter,
	})
}

func mapAnnotations(ctx context.Context, as parser.Annotations, scope AnnoScope, desc interface{}, opt Options) (ret []annoPair, left []parser.Annotation, next []parser.Annotation, err error) {
	con := make(amPairs, 0, len(as))
	cur := make([]parser.Annotation, 0, len(as))
	// try find mapper
	for _, a := range as {
		if mapper := FindAnnotationMapper(a.Key, scope); mapper != nil {
			con.Add(*a, mapper)
		} else {
			// no mapper found, just append it to the result
			cur = append(cur, *a)
		}
	}
	// process all the annotations under the mapper
	for _, a := range con {
		if c, n, err := a.inter.Map(ctx, a.cont, desc, opt); err != nil {
			return nil, nil, nil, err
		} else {
			cur = append(cur, c...)
			next = n
		}
	}
	m, left := mergeAnnotations(cur, scope)
	return m, left, next, nil
}

type annoPair struct {
	cont  []parser.Annotation
	inter Annotation
}

type annoPairs []annoPair

func (p *annoPairs) Add(con parser.Annotation, inter Annotation) {
	for i := range *p {
		if (*p)[i].inter == inter {
			(*p)[i].cont = append((*p)[i].cont, con)
			return
		}
	}
	*p = append(*p, annoPair{
		cont:  []parser.Annotation{con},
		inter: inter,
	})
}

func mergeAnnotations(in []parser.Annotation, scope AnnoScope) ([]annoPair, []parser.Annotation) {
	m := make(annoPairs, 0, len(in))
	left := make([]parser.Annotation, 0, len(in))
	for _, a := range in {
		if ann := FindAnnotation(a.Key, scope); ann != nil {
			m.Add(a, ann)
		} else {
			left = append(left, a)
		}
	}
	return m, left
}

func handleAnnotation(ctx context.Context, scope AnnoScope, ann Annotation, values []parser.Annotation, opts *Options, desc interface{}) error {
	handle, err := ann.Make(ctx, values, desc)
	if err != nil {
		return err
	}
	if handle != nil {
		switch ann.ID().Kind() {
		case AnnoKindOptionMapping:
			om, ok := handle.(OptionMapping)
			if !ok {
				return fmt.Errorf("annotation %#v for %d is not OptionMaker", handle, ann.ID())
			}
			*opts = om.Map(ctx, *opts)
			return nil
		default:
			//NOTICE: ignore unsupported annotations
			return fmt.Errorf("unsupported annotation %q", ann.ID())
		}
	}
	return nil
}

func handleFieldAnnotation(ctx context.Context, ann Annotation, values []parser.Annotation, opts *Options, field *FieldDescriptor, st *StructDescriptor, desc *parser.Field) error {
	handle, err := ann.Make(ctx, values, desc)
	if err != nil {
		return err
	}
	if handle != nil {
		switch ann.ID().Kind() {
		case AnnoKindOptionMapping:
			om, ok := handle.(OptionMapping)
			if !ok {
				return fmt.Errorf("annotation %#v for %d is not OptionMaker", handle, ann.ID())
			}
			*opts = om.Map(ctx, *opts)
			return nil
		case AnnoKindHttpMappping:
			hm, ok := handle.(HttpMapping)
			if !ok {
				return fmt.Errorf("annotation %#v for %d is not HTTPMapping", handle, ann.ID())
			}
			field.httpMappings = append(field.httpMappings, hm)
			st.addHttpMappingField(field)
			return nil
		case AnnoKindValueMapping:
			vm, ok := handle.(ValueMapping)
			if !ok {
				return fmt.Errorf("annotation %#v for %d is not ValueMapping", handle, ann.ID())
			}
			if field.valueMapping != nil {
				return fmt.Errorf("one field can only has one ValueMapping!")
			}
			field.valueMapping = vm
			field.valueMappingType = ann.ID().Type()
		case AnnoKindKeyMapping:
			km, ok := handle.(KeyMapping)
			if !ok {
				return fmt.Errorf("annotation %#v for %d is not KeyMapping", handle, ann.ID())
			}
			field.alias = km.Map(ctx, field.alias)
		default:
			//NOTICE: ignore unsupported annotations
			return fmt.Errorf("unsupported annotation %d", ann.ID())
		}
	}
	return nil
}

func handleNativeFieldAnnotation(ann parser.Annotation, f *FieldDescriptor, parseTarget ParseTarget) (skip bool, err error) {
	switch ann.GetKey() {
	case AnnoKeyDynamicGoDeprecated:
		{
			return true, nil
		}
	case AnnoKeyDynamicGoApiNone:
		if parseTarget == Response {
			return true, nil
		}
	}
	return false, nil
}

// copyAnnotations copy annotations from parser to descriptor
// WARN: this MUST be used before passing any annotation from parser to descriptor
func copyAnnotationValues(anns []*parser.Annotation) []parser.Annotation {
	ret := make([]parser.Annotation, len(anns))
	for i, ann := range anns {
		tmp := make([]string, len(ann.Values))
		copy(tmp, ann.Values)
		ret[i].Values = tmp
		ret[i].Key = ann.Key
	}
	return ret
}

// NameSpaceAnnotationKey is used for Option.PutNameSpaceToAnnotation
const NameSpaceAnnotationKey = "thrift.name_space"

func extractNameSpaceToAnnos(ast *parser.Thrift) parser.Annotation {
	ret := parser.Annotation{
		Key: NameSpaceAnnotationKey,
	}
	for _, s := range ast.Namespaces {
		if s == nil {
			continue
		}
		ret.Values = append(ret.Values, s.Language)
		ret.Values = append(ret.Values, s.Name)
	}
	return ret
}

// FilenameAnnotationKey is used for Option.PutThriftFilenameToAnnotation
const FilenameAnnotationKey = "thrift.filename"

func extractThriftFilePathToAnnos(ast *parser.Thrift) parser.Annotation {
	ret := parser.Annotation{
		Key: FilenameAnnotationKey,
	}
	ret.Values = append(ret.Values, ast.GetFilename())
	return ret
}

// injectAnnotation injects next annotation by appending.
// NOTICE: the next annotation will be appended to the end of the current annotation.
func injectAnnotations(origin *[]*parser.Annotation, next []parser.Annotation) error {
	if len(next) > 0 {
		off := len(*origin)
		tmp := make([]*parser.Annotation, off+len(next))
		copy(tmp, *origin)
		for i := off; i < len(tmp); i++ {
			tmp[i] = &next[i-off]
		}
		*origin = tmp
	}
	return nil
}

var (
	ctxIsBodyRoot    = "isBodyRoot"
	CtxKeyIsBodyRoot = &ctxIsBodyRoot
)
