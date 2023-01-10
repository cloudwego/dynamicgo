/**
 * Copyright 2022 CloudWeGo Authors.
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
	thrift.RegisterAnnotation(newValueMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindValueMapping, thrift.AnnoScopeField, JSConv)), "api.js_conv", "agw.js_conv")
	thrift.RegisterAnnotation(newValueMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindValueMapping, thrift.AnnoScopeField, BodyDynamic)), "agw.body_dynamic")

	// KeyMapping
	thrift.RegisterAnnotation(newKeyMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindKeyMapping, thrift.AnnoScopeField, APIKey)), "agw.key", APIKeyName)
	// thrift.RegisterAnnotation(newKeyMappingAnnotation(thrift.MakeAnnoID(thrift.AnnoKindKeyMapping, thrift.AnnoScopeField, NameCase)), "agw.to_snake", "janus.to_snake", "agw.to_lower_camel_case", "janus.to_lower_camel_case")

	// AnnotationMapper
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeField, goTagMapper{}, "go.tag")
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeField, apiBodyMapper{}, "api.body")
	// make raw.body not failed, expected caller to implement this anno
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeField, apiBodyMapper{}, "raw.body")
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeField, sourceMapper{}, "janus.source", "agw.source")
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeField, targetMapper{}, "agw.target", "janus.target")
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeService, nameCaseMapper{}, NameCaseKeys...)
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeFunction, nameCaseMapper{}, NameCaseKeys...)
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeStruct, nameCaseMapper{}, NameCaseKeys...)
	thrift.RegisterAnnotationMapper(thrift.AnnoScopeField, nameCaseMapper{}, NameCaseKeys...)
}

var ErrNotImplemented = meta.NewError(meta.ErrUnsupportedType, "not implemented annotation", nil)
