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
)

const (
	ToSnake    thrift.AnnoType = 401
	ToLowCamel thrift.AnnoType = 402
)

type optionMappingAnnotation struct {
	typ thrift.AnnoID
}

func newOptionMapping(typ thrift.AnnoID) optionMappingAnnotation {
	return optionMappingAnnotation{
		typ: typ,
	}
}

func (self optionMappingAnnotation) ID() thrift.AnnoID {
	return self.typ
}

func (self optionMappingAnnotation) Make(ctx context.Context, key string, value []string, ast interface{}) (interface{}, error) {
	return nil, nil
}
