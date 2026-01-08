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
	"fmt"
	"unsafe"

	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/util"
	"github.com/cloudwego/thriftgo/parser"
)

const (
	VERSION_MASK = 0xffff0000
	VERSION_1    = 0x80010000
)

// Type constants in the Thrift protocol
type Type byte

// built-in Types
const (
	STOP   Type = 0
	VOID   Type = 1
	BOOL   Type = 2
	BYTE   Type = 3
	I08    Type = 3
	DOUBLE Type = 4
	I16    Type = 6
	I32    Type = 8
	I64    Type = 10
	STRING Type = 11
	// UTF7   Type = 11
	STRUCT Type = 12
	MAP    Type = 13
	SET    Type = 14
	LIST   Type = 15
	UTF8   Type = 16
	UTF16  Type = 17
	// BINARY Type = 18   wrong and unusued

	ERROR Type = 255
)

func (p Type) Valid() bool {
	switch p {
	case STOP, VOID, BOOL, BYTE, DOUBLE, I16, I32, I64, STRING, STRUCT, MAP, SET, LIST, UTF8, UTF16:
		return true
	default:
		return false
	}
}

// String for format and print
func (p Type) String() string {
	switch p {
	case STOP:
		return "STOP"
	case VOID:
		return "VOID"
	case BOOL:
		return "BOOL"
	case BYTE:
		return "BYTE"
	case DOUBLE:
		return "DOUBLE"
	case I16:
		return "I16"
	case I32:
		return "I32"
	case I64:
		return "I64"
	case STRING:
		return "STRING"
	case STRUCT:
		return "STRUCT"
	case MAP:
		return "MAP"
	case SET:
		return "SET"
	case LIST:
		return "LIST"
	case UTF8:
		return "UTF8"
	case UTF16:
		return "UTF16"
	case ERROR:
		return "ERROR"
	}
	return "Unknown"
}

// IsInt tells if the type is one of I08, I16, I32, I64
func (p Type) IsInt() bool {
	return p == I08 || p == I16 || p == I32 || p == I64
}

// IsComplex tells if the type is one of STRUCT, MAP, SET, LIST
func (p Type) IsComplex() bool {
	return p == STRUCT || p == MAP || p == SET || p == LIST
}

// TypeDescriptor is the runtime descriptor of a thrift type
type TypeDescriptor struct {
	typ   Type
	name  string
	key   *TypeDescriptor   // for map key
	elem  *TypeDescriptor   // for slice or map element
	struc *StructDescriptor // for struct
}

const (
	nameBinary = "binary"
)

// IsBinary tells if the type is binary type ([]byte)
func (d TypeDescriptor) IsBinary() bool {
	if d.name == nameBinary {
		return true
	}
	return false
}

// Type returns the build-in type of the descriptor
func (d TypeDescriptor) Type() Type {
	return d.typ
}

// Name returns the name of the descriptor
//
//	for struct, it is the struct name;
//	for build-in type, it is the type name.
func (d TypeDescriptor) Name() string {
	return d.name
}

// Key returns the key type descriptor of a MAP type
func (d TypeDescriptor) Key() *TypeDescriptor {
	return d.key
}

// Elem returns the element type descriptor of a LIST, SET or MAP type
func (d TypeDescriptor) Elem() *TypeDescriptor {
	return d.elem
}

// Struct returns the struct type descriptor of a STRUCT type
func (d TypeDescriptor) Struct() *StructDescriptor {
	return d.struc
}

// StructDescriptor is the runtime descriptor of a STRUCT type
type StructDescriptor struct {
	baseID      FieldID
	name        string
	ids         util.FieldIDMap
	names       util.FieldNameMap
	requires    RequiresBitmap
	hmFields    []*FieldDescriptor
	annotations []parser.Annotation
}

// GetRequestBase returns the base.Base field of a STRUCT
func (s StructDescriptor) GetRequestBase() *FieldDescriptor {
	field := s.FieldById(s.baseID)
	if field != nil && field.IsRequestBase() {
		return field
	}
	return nil
}

// GetResponseBase returns the base.BaseResp field of a STRUCT
func (s StructDescriptor) GetResponseBase() *FieldDescriptor {
	field := s.FieldById(s.baseID)
	if field != nil && field.IsResponseBase() {
		return field
	}
	return nil
}

// GetHttpMappingFields returns the fields with http-mapping annotations
func (s StructDescriptor) HttpMappingFields() []*FieldDescriptor {
	return s.hmFields
}

func (s *StructDescriptor) addHttpMappingField(f *FieldDescriptor) {
	for _, fd := range s.hmFields {
		if fd.id == f.id {
			return
		}
	}
	s.hmFields = append(s.hmFields, f)
}

// Name returns the name of the struct
func (s StructDescriptor) Name() string {
	return s.name
}

// Len returns the number of fields in the struct
func (s StructDescriptor) Len() int {
	return len(s.ids.All())
}

// Fields returns all fields in the struct
func (s StructDescriptor) Fields() []*FieldDescriptor {
	ret := s.ids.All()
	return *(*[]*FieldDescriptor)(unsafe.Pointer(&ret))
}

// Fields returns requireness bitmap in the struct.
// By default, only Requred and Default fields are marked.
func (s StructDescriptor) Requires() RequiresBitmap {
	return s.requires
}

func (s StructDescriptor) Annotations() []parser.Annotation {
	return s.annotations
}

// FieldById finds the field by field id
func (s StructDescriptor) FieldById(id FieldID) *FieldDescriptor {
	return (*FieldDescriptor)(s.ids.Get(int32(id)))
}

// FieldByName finds the field by key
//
// NOTICE: Options.MapFieldWay can influence the behavior of this method.
// ep: if Options.MapFieldWay is MapFieldWayName, then field names should be used as key.
func (s StructDescriptor) FieldByKey(k string) (field *FieldDescriptor) {
	return (*FieldDescriptor)(s.names.Get(k))
}

// FieldID is used to identify a field in a struct
type FieldID uint16

// FieldDescriptor is the runtime descriptor of a field in a struct
type FieldDescriptor struct {
	// isException      bool
	isRequestBase    bool
	isResponseBase   bool
	required         Requireness
	valueMappingType AnnoType
	id               FieldID
	typ              *TypeDescriptor
	defaultValue     *DefaultValue
	name             string // field name
	alias            string // alias name
	valueMapping     ValueMapping
	httpMappings     []HttpMapping
	annotations      []parser.Annotation
}

// func (f FieldDescriptor) IsException() bool {
// 	return f.isException
// }

// IsRequestBase tells if the field is base.Base
func (f FieldDescriptor) IsRequestBase() bool {
	return f.isRequestBase
}

// IsResponseBase tells if the field is base.BaseResp
func (f FieldDescriptor) IsResponseBase() bool {
	return f.isResponseBase
}

// ValueMappingType returns the value-mapping annotation's type of a field
func (f FieldDescriptor) ValueMappingType() AnnoType {
	return f.valueMappingType
}

// Required return the requiredness of a field
func (f FieldDescriptor) Required() Requireness {
	return f.required
}

// ID returns the id of a field
func (f FieldDescriptor) ID() FieldID {
	return f.id
}

// Name returns the name of a field
func (f FieldDescriptor) Name() string {
	return f.name
}

// Alias returns the alias of a field
func (f FieldDescriptor) Alias() string {
	return f.alias
}

// Type returns the type descriptor of a field
func (f FieldDescriptor) Type() *TypeDescriptor {
	return f.typ
}

// HTTPMappings returns the http-mapping annotations of a field
func (f FieldDescriptor) HTTPMappings() []HttpMapping {
	return f.httpMappings
}

// ValueMapping returns the value-mapping annotation of a field
func (f FieldDescriptor) ValueMapping() ValueMapping {
	return f.valueMapping
}

func (f FieldDescriptor) Annotations() []parser.Annotation {
	return f.annotations
}

// DefaultValue returns the default value of a field
func (f FieldDescriptor) DefaultValue() *DefaultValue {
	return f.defaultValue
}

// FunctionDescriptor idl function descriptor
type FunctionDescriptor struct {
	oneway            bool
	hasRequestBase    bool
	request           *TypeDescriptor
	response          *TypeDescriptor
	name              string
	endpoints         []http.Endpoint
	annotations       []parser.Annotation
	isWithoutWrapping bool
}

// Name returns the name of the function
func (f FunctionDescriptor) Name() string {
	return f.name
}

// Oneway tells if the function is oneway type
func (f FunctionDescriptor) Oneway() bool {
	return f.oneway
}

// HasRequestBase tells if the function has a base.Base field
func (f FunctionDescriptor) HasRequestBase() bool {
	return f.hasRequestBase
}

// Request returns the request type descriptor of the function
// The request arguements is mapped with arguement id and name
func (f FunctionDescriptor) Request() *TypeDescriptor {
	return f.request
}

// Response returns the response type descriptor of the function
// The response arguements is mapped with arguement id
func (f FunctionDescriptor) Response() *TypeDescriptor {
	return f.response
}

// Endpoints returns the http endpoints of the function
func (f FunctionDescriptor) Endpoints() []http.Endpoint {
	return f.endpoints
}

// Annotations returns the annotations of the function
func (f FunctionDescriptor) Annotations() []parser.Annotation {
	return f.annotations
}

// IsWithoutWrapping returns if the request and response are not wrapped in struct
func (f FunctionDescriptor) IsWithoutWrapping() bool {
	return f.isWithoutWrapping
}

// ServiceDescriptor is the runtime descriptor of a service
type ServiceDescriptor struct {
	name        string
	functions   map[string]*FunctionDescriptor
	annotations []parser.Annotation
}

// Name returns the name of the service
func (s ServiceDescriptor) Name() string {
	return s.name
}

// Functions returns all functions in the service
func (s ServiceDescriptor) Functions() map[string]*FunctionDescriptor {
	return s.functions
}

// LookupFunctionByMethod lookup function by method name
func (s *ServiceDescriptor) LookupFunctionByMethod(method string) (*FunctionDescriptor, error) {
	fnSvc, ok := s.functions[method]
	if !ok {
		return nil, fmt.Errorf("missing method: %s in service: %s", method, s.name)
	}
	return fnSvc, nil
}

// Annotations returns the annotations of a service
func (s ServiceDescriptor) Annotations() []parser.Annotation {
	return s.annotations
}

// Requireness is the requireness of a field.
// See also https://thrift.apache.org/docs/idl.html
type Requireness uint8

const (
	// OptionalRequireness means the field is optional
	OptionalRequireness Requireness = 0
	// DefaultRequireness means the field is default-requireness
	DefaultRequireness Requireness = 1
	// RequiredRequireness means the field is required
	RequiredRequireness Requireness = 2
)
