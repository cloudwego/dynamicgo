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

// /*
//  * Copyright 2021 CloudWeGo Authors
//  *
//  * Licensed under the Apache License, Version 2.0 (the "License");
//  * you may not use this file except in compliance with the License.
//  * You may obtain a copy of the License at
//  *
//  *     http://www.apache.org/licenses/LICENSE-2.0
//  *
//  * Unless required by applicable law or agreed to in writing, software
//  * distributed under the License is distributed on an "AS IS" BASIS,
//  * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  * See the License for the specific language governing permissions and
//  * limitations under the License.
//  */

package annotation

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/cloudwego/thriftgo/parser"
)

const (
	APIQuery thrift.AnnoType = 1 + iota
	APIPath
	APIHeader
	APICookie
	APIBody
	APIHTTPCode
	APIRawBody
	APIPostForm
	APIRawUri
	APINoBodyStruct
)

type httpMappingAnnotation struct {
	typ thrift.AnnoID
}

func newHttpMappingAnnotation(typ thrift.AnnoID) httpMappingAnnotation {
	return httpMappingAnnotation{
		typ: typ,
	}
}

func (self httpMappingAnnotation) ID() thrift.AnnoID {
	return self.typ
}

func (self httpMappingAnnotation) Make(ctx context.Context, values []parser.Annotation, ast interface{}) (interface{}, error) {
	if len(values) != 1 || len(values[0].Values) < 1 {
		return nil, errors.New("httpMappingAnnotation only accept single key and value")
	}
	value := values[0].Values[0]
	switch t := self.typ.Type(); t {
	case APIQuery:
		return apiQuery{value: value}, nil
	case APIPath:
		return apiPath{value: value}, nil
	case APIHeader:
		return apiHeader{value: value}, nil
	case APICookie:
		return apiCookie{value: value}, nil
	case APIBody:
		return apiBody{value: value}, nil
	case APIHTTPCode:
		return apiHTTPCode{}, nil
	case APIRawBody:
		return apiRawBody{}, nil
	case APIPostForm:
		return apiPostForm{value: value}, nil
	case APIRawUri:
		return apiRawUri{}, nil
	case APINoBodyStruct:
		return apiNoBodyStruct{}, nil
	default:
		return nil, errNotImplemented(fmt.Sprintf("unsupported type %d of http-mapping annotation", t))
	}
}

type apiPostForm struct {
	value string // == xx
}

func (m apiPostForm) Request(ctx context.Context, req http.RequestGetter, field *thrift.FieldDescriptor) (string, error) {
	v := req.GetPostForm(m.value)
	if v == "" {
		return "", errNotFound(m.value, "postform")
	}
	return v, nil
}

func (apiPostForm) Response(ctx context.Context, resp http.ResponseSetter, field *thrift.FieldDescriptor, val string) error {
	return errNotImplemented("apiPostForm not support http response!")
}

func (apiPostForm) Encoding() meta.Encoding {
	return meta.EncodingJSON
}

type apiQuery struct {
	value string // == xx
}

func (m apiQuery) Request(ctx context.Context, req http.RequestGetter, field *thrift.FieldDescriptor) (string, error) {
	v := req.GetQuery(m.value)
	if v == "" {
		return "", errNotFound(m.value, "query")
	}
	return v, nil
}

func (apiQuery) Response(ctx context.Context, resp http.ResponseSetter, field *thrift.FieldDescriptor, val string) error {
	return errNotImplemented("apiQuery not support http response!")
}

func (apiQuery) Encoding() meta.Encoding {
	return meta.EncodingJSON
}

// api.path = xx
type apiPath struct {
	value string
}

func (m apiPath) Request(ctx context.Context, req http.RequestGetter, field *thrift.FieldDescriptor) (string, error) {
	v := req.GetParam(m.value)
	if v == "" {
		return "", errNotFound(m.value, "url path")
	}
	return v, nil
}

func (apiPath) Response(ctx context.Context, resp http.ResponseSetter, field *thrift.FieldDescriptor, val string) error {
	return errNotImplemented("apiPath not support http response!")
}

func (apiPath) Encoding() meta.Encoding {
	return meta.EncodingJSON
}

// api.header = xx
type apiHeader struct {
	value string
}

func (m apiHeader) Request(ctx context.Context, req http.RequestGetter, field *thrift.FieldDescriptor) (string, error) {
	v := req.GetHeader(m.value)
	if v == "" {
		return "", errNotFound(m.value, "request header")
	}
	return v, nil
}

func (m apiHeader) Response(ctx context.Context, resp http.ResponseSetter, field *thrift.FieldDescriptor, val string) error {
	return resp.SetHeader(m.value, val)
}

func (apiHeader) Encoding() meta.Encoding {
	return meta.EncodingJSON
}

// api.cookie = xx
type apiCookie struct {
	value string
}

func (m apiCookie) Request(ctx context.Context, req http.RequestGetter, field *thrift.FieldDescriptor) (string, error) {
	v := req.GetCookie(m.value)
	if v == "" {
		return "", errNotFound(m.value, "request cookie")
	}
	return v, nil
}

func (m apiCookie) Response(ctx context.Context, resp http.ResponseSetter, field *thrift.FieldDescriptor, val string) error {
	return resp.SetCookie(m.value, val)
}

func (apiCookie) Encoding() meta.Encoding {
	return meta.EncodingJSON
}

// api.body = xx
type apiBody struct {
	value string
}

func (m apiBody) Request(ctx context.Context, req http.RequestGetter, field *thrift.FieldDescriptor) (string, error) {
	v := req.GetMapBody(m.value)
	if v == "" {
		return "", errNotFound(m.value, "body")
	}
	return v, nil
}

func (m apiBody) Response(ctx context.Context, resp http.ResponseSetter, field *thrift.FieldDescriptor, val string) error {
	// v, err := convertAny(val, field.Type().Type())
	// if err != nil {
	// 	return err
	// }
	// resp.Body[m.value] = val
	return errNotImplemented("apiBody not support http response!")
}

func (apiBody) Encoding() meta.Encoding {
	return meta.EncodingJSON
}

type apiHTTPCode struct{}

func (m apiHTTPCode) Request(ctx context.Context, req http.RequestGetter, field *thrift.FieldDescriptor) (string, error) {
	return "", errNotImplemented("apiBody not support http request!")
}

func (m apiHTTPCode) Response(ctx context.Context, resp http.ResponseSetter, field *thrift.FieldDescriptor, val string) error {
	i, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	return resp.SetStatusCode(i)
}

func (apiHTTPCode) Encoding() meta.Encoding {
	return meta.EncodingJSON
}

type apiNone struct{}

func (m apiNone) Request(ctx context.Context, req http.RequestGetter, field *thrift.FieldDescriptor) (string, error) {
	return "", nil
}

func (m apiNone) Response(ctx context.Context, resp http.ResponseSetter, field *thrift.FieldDescriptor, val string) error {
	return nil
}

func (apiNone) Encoding() meta.Encoding {
	return meta.EncodingJSON
}

type apiRawBody struct{}

func (apiRawBody) Request(ctx context.Context, req http.RequestGetter, field *thrift.FieldDescriptor) (string, error) {
	// OPT: pass unsafe buffer to conv.DoNative may cause panic
	return string(req.GetBody()), nil
}

func (apiRawBody) Response(ctx context.Context, resp http.ResponseSetter, field *thrift.FieldDescriptor, val string) error {
	return resp.SetRawBody(rt.Str2Mem(val))
}

func (apiRawBody) Encoding() meta.Encoding {
	return meta.EncodingJSON
}

type apiRawUri struct{}

func (m apiRawUri) Request(ctx context.Context, req http.RequestGetter, field *thrift.FieldDescriptor) (string, error) {
	return req.GetUri(), nil
}

func (m apiRawUri) Response(ctx context.Context, resp http.ResponseSetter, field *thrift.FieldDescriptor, val string) error {
	return nil
}

func (apiRawUri) Encoding() meta.Encoding {
	return meta.EncodingJSON
}

type apiNoBodyStruct struct{}

func (m apiNoBodyStruct) Request(ctx context.Context, req http.RequestGetter, field *thrift.FieldDescriptor) (string, error) {
	if field.Type().Type() != thrift.STRUCT {
		return "", fmt.Errorf("apiNoBodyStruct only support STRUCT type, but got %v", field.Type().Type())
	}

	var opts conv.Options
	if v := ctx.Value(conv.CtxKeyConvOptions); v != nil {
		op1, ok := v.(conv.Options)
		if !ok {
			op2, ok := v.(*conv.Options)
			if !ok {
				return "", fmt.Errorf("invalid conv options type: %T", v)
			}
			opts = *op2
		}
		opts = op1
	}

	p := thrift.NewBinaryProtocolBuffer()
	// p.WriteStructBegin(field.Name())
	for _, f := range field.Type().Struct().HttpMappingFields() {
		var ok bool
		var val string
		for _, hm := range f.HTTPMappings() {
			v, err := hm.Request(ctx, req, f)
			if err == nil {
				val = v
				ok = true
				break
			}
		}
		p.WriteFieldBegin(f.Name(), f.Type().Type(), f.ID())
		if !ok || val == "" {
			p.WriteDefaultOrEmpty(f)
		} else {
			// TODO: pass conv options to decide
			p.WriteStringWithDesc(val, f.Type(), opts.DisallowUnknownField, !opts.NoBase64Binary)
		}
		// p.WriteFieldEnd()
	}
	p.WriteStructEnd()
	out := make([]byte, len(p.Buf))
	copy(out, p.Buf)
	p.Recycle()
	return rt.Mem2Str(out), nil
}

func (m apiNoBodyStruct) Response(ctx context.Context, resp http.ResponseSetter, field *thrift.FieldDescriptor, val string) error {
	return errNotImplemented("apiNoBodyStruct not support http Response!")
}

func (apiNoBodyStruct) Encoding() meta.Encoding {
	return meta.EncodingThriftBinary
}
