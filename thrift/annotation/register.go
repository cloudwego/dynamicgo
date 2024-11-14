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

package annotation

import (
	"fmt"

	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
)

func init() {
	// HttpMapping
	thrift.RegisterAnnotation(newHttpMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindHttpMappping, thrift.AnnoScopeField, APIQuery)), "api.query")
	thrift.RegisterAnnotation(newHttpMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindHttpMappping, thrift.AnnoScopeField, APIPath)), "api.path")
	thrift.RegisterAnnotation(newHttpMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindHttpMappping, thrift.AnnoScopeField, APIHeader)), "api.header")
	thrift.RegisterAnnotation(newHttpMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindHttpMappping, thrift.AnnoScopeField, APICookie)), "api.cookie")
	thrift.RegisterAnnotation(newHttpMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindHttpMappping, thrift.AnnoScopeField, APIBody)), "api.body")
	thrift.RegisterAnnotation(newHttpMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindHttpMappping, thrift.AnnoScopeField, APIHTTPCode)), "api.http_code")
	thrift.RegisterAnnotation(newHttpMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindHttpMappping, thrift.AnnoScopeField, APIRawBody)), "api.raw_body")
	thrift.RegisterAnnotation(newHttpMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindHttpMappping, thrift.AnnoScopeField, APIPostForm)), "api.form")
	thrift.RegisterAnnotation(newHttpMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindHttpMappping, thrift.AnnoScopeField, APIRawUri)), "api.raw_uri")
	thrift.RegisterAnnotation(newHttpMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindHttpMappping, thrift.AnnoScopeField, APINoBodyStruct)), "api.no_body_struct")

	// OptionMaker

	// ValueMapping
	thrift.RegisterAnnotation(newValueMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindValueMapping, thrift.AnnoScopeField, JSConv)), "api.js_conv")

	// KeyMapping
	thrift.RegisterAnnotation(newKeyMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindKeyMapping, thrift.AnnoScopeField, APIKey)), APIKeyName)

	// AnnotationMapper
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeField, goTagMapper{}, "go.tag")
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeField, apiBodyMapper{}, "api.body")
	// make raw.body not failed, expected caller to implement this anno
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeField, apiBodyMapper{}, "raw.body")
}

// This is only used for internal specifications.
// DO NOT USE IT if you don't know related annotations
func InitAGWAnnos() {
	thrift.RegisterAnnotation(newValueMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindValueMapping, thrift.AnnoScopeField, JSConv)), "agw.js_conv")
	thrift.RegisterAnnotation(newValueMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindValueMapping, thrift.AnnoScopeField, BodyDynamic)), "agw.body_dynamic")
	thrift.RegisterAnnotation(newKeyMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindKeyMapping, thrift.AnnoScopeField, APIKey)), "agw.key")
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeField, sourceMapper{}, "janus.source", "agw.source")
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeField, targetMapper{}, "agw.target", "janus.target")
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeService, nameCaseMapper{}, NameCaseKeys...)
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeFunction, nameCaseMapper{}, NameCaseKeys...)
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeStruct, nameCaseMapper{}, NameCaseKeys...)
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeField, nameCaseMapper{}, NameCaseKeys...)
}

//go:noline
func errNotFound(key string, scope string) error {
	return meta.NewError(meta.ErrNotFound, fmt.Sprintf("not fould %s in %s", key, scope), nil)
}

//go:noline
func errNotImplemented(msg string) error {
	return meta.NewError(meta.ErrUnsupportedType, msg, nil)
}
