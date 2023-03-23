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
	"context"

	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/cloudwego/thriftgo/parser"
)

const (
	APIKey thrift.AnnoType = 501
	// NameCase thrift.AnnoType = 502

	APIKeyName string = "api.key"
)

var (
	NameCaseKeys = []string{"agw.to_snake", "janus.to_snake", "agw.to_lower_camel_case", "janus.to_lower_camel_case"}
)

type keyMappingAnnotation struct {
	typ thrift.AnnoID
}

func newKeyMappingAnnotation(typ thrift.AnnoID) keyMappingAnnotation {
	return keyMappingAnnotation{
		typ: typ,
	}
}

func (self keyMappingAnnotation) ID() thrift.AnnoID {
	return self.typ
}

func (self keyMappingAnnotation) Make(ctx context.Context, values []parser.Annotation, ast interface{}) (interface{}, error) {
	if len(values) == 0 {
		return nil, nil
	}
	for _, v := range values {
		// NOTICE: we only handle the first value here
		if len(v.Values) == 1 {
			switch self.typ.Type() {
			case APIKey:
				return &apiKey{v.Values[0]}, nil
			default:
				return nil, errNotImplemented("keyMappingAnnotation must have APIKey type")
			}
		}
	}
	return nil, nil
}

type apiKey struct {
	Value string
}

func (m apiKey) Map(ctx context.Context, key string) string {
	return m.Value
}
