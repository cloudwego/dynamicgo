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
	"context"
	"testing"

	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/stretchr/testify/require"
)

func TestAgwKey(t *testing.T) {
	path := "a/b/main.thrift"
	content := `
	namespace go kitex.test.server
	struct Base {
		1: required string required_field
		999: required string test1 (agw.key="test2")
		3: string pass 
	}

	service InboxService {
		string ExampleMethod(1: Base req)
	}
	`
	includes := map[string]string{
		path: content,
	}
	p, err := thrift.NewDescritorFromContent(context.Background(), path, content, includes, true)
	if err != nil {
		t.Fatal(err)
	}
	req := p.Functions()["ExampleMethod"].Request().Struct().Fields()[0].Type()
	require.Equal(t, "test2", req.Struct().FieldById(999).Alias())
}

func TestNameCase(t *testing.T) {
	path := "a/b/main.thrift"
	content := `
	namespace go kitex.test.server

	struct Base1 {
		1: string InnerBase (agw.to_snake="false")
		2: string Msg 
		3: string logID (agw.to_lower_camel_case="false")
		4: string req_list 
	}

	struct Base2 {
		1: string InnerBase (agw.to_snake="false")
		2: string Msg 
		3: string logID
		4: string req_list (agw.to_lower_camel_case="true")
	} (agw.to_snake="False")

	struct Base3 {
		1: string InnerBase 
		2: string Msg 
		3: string logID
		4: string req_list
	}

	struct Base4 {
		1: string InnerBase (agw.to_snake="true")
		2: string Msg (agw.to_lower_camel_case="false")
		3: string logID
		4: string req_list 
	} (agw.to_snake="falsE")

	service ExampleService {
		Base2 ExampleMethod1(1: Base1 req) (agw.to_snake="True")
		Base4 ExampleMethod2(1: Base3 req) (agw.to_lower_camel_case="True")
		Base4 ExampleMethod3(1: Base3 req) (agw.to_snake="True")
	} 
	`
	includes := map[string]string{
		path: content,
	}

	p, err := thrift.Options{}.NewDescriptorFromContentWithMethod(context.Background(), path, content, includes, true, "ExampleMethod1", "ExampleMethod2")
	require.NoError(t, err)

	base1 := p.Functions()["ExampleMethod1"].Request().Struct().Fields()[0].Type()
	require.Equal(t, "InnerBase", base1.Struct().FieldById(1).Alias())
	require.Equal(t, "msg", base1.Struct().FieldById(2).Alias())
	require.Equal(t, "log_id", base1.Struct().FieldById(3).Alias())
	require.Equal(t, "req_list", base1.Struct().FieldById(4).Alias())

	base2 := p.Functions()["ExampleMethod1"].Response().Struct().Fields()[0].Type()
	require.Equal(t, "InnerBase", base2.Struct().FieldById(1).Alias())
	require.Equal(t, "Msg", base2.Struct().FieldById(2).Alias())
	require.Equal(t, "logID", base2.Struct().FieldById(3).Alias())
	require.Equal(t, "reqList", base2.Struct().FieldById(4).Alias())

	base3 := p.Functions()["ExampleMethod2"].Request().Struct().Fields()[0].Type()
	require.Equal(t, "innerBase", base3.Struct().FieldById(1).Alias())
	require.Equal(t, "msg", base3.Struct().FieldById(2).Alias())
	require.Equal(t, "logID", base3.Struct().FieldById(3).Alias())
	require.Equal(t, "reqList", base3.Struct().FieldById(4).Alias())

	base4 := p.Functions()["ExampleMethod2"].Response().Struct().Fields()[0].Type()
	require.Equal(t, "inner_base", base4.Struct().FieldById(1).Alias())
	require.Equal(t, "Msg", base4.Struct().FieldById(2).Alias())
	require.Equal(t, "logID", base4.Struct().FieldById(3).Alias())
	require.Equal(t, "reqList", base4.Struct().FieldById(4).Alias())

	p, err = thrift.Options{}.NewDescriptorFromContentWithMethod(context.Background(), path, content, includes, true, "ExampleMethod3")
	require.NoError(t, err)
	base3 = p.Functions()["ExampleMethod3"].Request().Struct().Fields()[0].Type()
	require.Equal(t, "inner_base", base3.Struct().FieldById(1).Alias())
	require.Equal(t, "msg", base3.Struct().FieldById(2).Alias())
	require.Equal(t, "log_id", base3.Struct().FieldById(3).Alias())
	require.Equal(t, "req_list", base3.Struct().FieldById(4).Alias())

	base4 = p.Functions()["ExampleMethod3"].Response().Struct().Fields()[0].Type()
	require.Equal(t, "inner_base", base4.Struct().FieldById(1).Alias())
	require.Equal(t, "Msg", base4.Struct().FieldById(2).Alias())
	require.Equal(t, "logID", base4.Struct().FieldById(3).Alias())
	require.Equal(t, "req_list", base4.Struct().FieldById(4).Alias())
}
