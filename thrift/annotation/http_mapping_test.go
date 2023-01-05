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
	"testing"

	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/stretchr/testify/require"
)

func GetDescFromContent(content string, method string) (*thrift.FunctionDescriptor, error) {
	path := "a/b/main.thrift"
	includes := map[string]string{
		path: content,
	}
	p, err := thrift.NewDescritorFromContent(path, content, includes, true)
	if err != nil {
		return nil, err
	}
	return p.Functions()[method], nil
}

func TestAPIQuery(t *testing.T) {
	fn, err := GetDescFromContent(`
	namespace go kitex.test.server
	struct Base {
		1: required string required_field
		999: required string test1 (api.query="test2")
		3: string pass 
	}

	service InboxService {
		string ExampleMethod(1: Base req)
	}
	`, "ExampleMethod")
	if err != nil {
		t.Fatal(err)
	}
	req := fn.Request().Struct().FieldByKey("req").Type()
	field := req.Struct().FieldByKey("test1")
	require.Equal(t, field, req.Struct().HttpMappingFields()[0])
	hm := field.HTTPMappings()
	require.Equal(t, []thrift.HttpMapping{apiQuery{"test2"}}, hm)
}

func TestAPIRawUri(t *testing.T) {
	fn, err := GetDescFromContent(`
	namespace go kitex.test.server
	struct Base {
		999: required string test1 (api.raw_uri="")
	}

	service InboxService {
		string ExampleMethod(1: Base req)
	}
	`, "ExampleMethod")
	if err != nil {
		t.Fatal(err)
	}
	req := fn.Request().Struct().Fields()[0].Type()
	field := req.Struct().FieldByKey("test1")
	require.Equal(t, field, req.Struct().HttpMappingFields()[0])
	hm := field.HTTPMappings()
	require.Equal(t, []thrift.HttpMapping{apiRawUri{}}, hm)
}
