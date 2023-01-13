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

package thrift

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	stdh "net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/meta"
)

func TestThriftContentWithAbsIncludePath(t *testing.T) {
	path := "a/b/main.thrift"
	content := `
	namespace go kitex.test.server
	include "x.thrift"
	include "../y.thrift" 

	service InboxService {}
	`
	includes := map[string]string{
		path:           content,
		"a/b/x.thrift": "namespace go kitex.test.server",
		"a/y.thrift": `
		namespace go kitex.test.server
		include "z.thrift"
		`,
		"a/z.thrift": "namespace go kitex.test.server",
	}
	p, err := NewDescritorFromContent(path, content, includes, true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", p)
}

func TestBitmap(t *testing.T) {
	content := `
	namespace go kitex.test.server
	struct Base {
		1: string A
		2: required string B
		3: optional string C
		32767: required string D
	}
	service InboxService {
		Base ExampleMethod(1: Base req)
	}`
	p, err := GetDescFromContent(content, "ExampleMethod", &Options{})
	require.NoError(t, err)
	testID := math.MaxInt16
	println(testID/int64BitSize, testID%int64BitSize)
	x := p.Request().Struct().FieldByKey("req").Type().Struct().Requires()
	exp := []uint64{0x6}
	for i := 0; i < int(testID)/int64BitSize-1; i++ {
		exp = append(exp, 0)
	}
	exp = append(exp, 0x8000000000000000)
	require.Equal(t, RequiresBitmap(exp), (x))
}

func TestDynamicgoDeprecated(t *testing.T) {
	path := "a/b/main.thrift"
	content := `
	namespace go kitex.test.server
	struct Base {
		1: required string required_field
		999: required string ignored (` + AnnoKeyDynamicGoDeprecated + `="")
		3: string pass 
	}

	service InboxService {
		string ExampleMethod(1: Base req)
	}
	`
	includes := map[string]string{
		path: content,
	}
	p, err := NewDescritorFromContent(path, content, includes, true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", p)
	x := p.Functions()["ExampleMethod"].Request().Struct().FieldByKey("req").Type().Struct()
	require.Equal(t, (*FieldDescriptor)(nil), x.FieldById(999))
	require.NotNil(t, x.FieldById(3))
	require.Equal(t, (*FieldDescriptor)(nil), x.FieldByKey("ignored"))
	require.Equal(t, true, x.requires.IsSet(FieldID(1)))
	require.Equal(t, true, x.requires.IsSet(FieldID(3)))
}

func TestHttpEndPoints(t *testing.T) {
	p, err := NewDescritorFromPath("../testdata/idl/example3.thrift")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", p)
	ep1 := p.Functions()["ExampleMethod"].endpoints
	require.Equal(t, 1, len(ep1))
	require.Equal(t, "POST", ep1[0].Method)
	require.Equal(t, "/example/set", ep1[0].Path)
	ep2 := p.Functions()["ErrorMethod"].endpoints
	require.Equal(t, 1, len(ep2))
	require.Equal(t, "GET", ep2[0].Method)
	require.Equal(t, "/example/get", ep2[0].Path)
}

func TestBypassAnnotatio(t *testing.T) {
	path := "a/b/main.thrift"
	content := `
	namespace go kitex.test.server

	service InboxService {
		string ExampleMethod(1: string req) (anno.test = "中文1", anno.test = "中文2")
	} (anno.test = "中文1", anno.test = "中文2")
	`

	includes := map[string]string{
		path: content,
	}
	p, err := NewDescritorFromContent(path, content, includes, true)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, []string{"中文1", "中文2"}, p.Annotations()["anno.test"])
	require.Equal(t, []string{"中文1", "中文2"}, p.Functions()["ExampleMethod"].Annotations()["anno.test"])
}

func TestRouterLookup(t *testing.T) {
	opts := Options{}
	svc, err := opts.NewDescritorFromPath("../testdata/idl/example3.thrift")
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadFile("../testdata/data/example3req.json")
	if err != nil {
		panic(err)
	}
	hr, err := stdh.NewRequest("POST", "http://localhost:8888/example/set?query=1,2,3", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	req := &http.HTTPRequest{
		Request: hr,
	}
	fnDsc, err := svc.Router.Lookup(req)
	require.Equal(t, err, nil)
	require.Equal(t, fnDsc.name, "ExampleMethod")
}

func GetDescFromContent(content string, method string, opts *Options) (*FunctionDescriptor, error) {
	path := "a/b/main.thrift"
	includes := map[string]string{
		path: content,
	}
	p, err := opts.NewDescritorFromContent(path, content, includes, true)
	if err != nil {
		return nil, err
	}
	return p.Functions()[method], nil
}

// func TestRequireBitmap(t *testing.T) {
// 	reqs := NewRequiresBitmap()
// 	reqs.Set(256, DefaultRequireness)
// 	b := (*reqs)[:5]
// 	require.Equal(t, b[4], uint64(0x1))
// 	FreeRequiresBitmap(reqs)
// 	reqs2 := NewRequiresBitmap()
// 	if cap(*reqs2) >= 5 {
// 		b := (*reqs2)[:5]
// 		require.Equal(t, b[4], uint64(0x1))
// 	}
// }

func TestDefalutValue(t *testing.T) {
	t.Run("use", func(t *testing.T) {
		p, err := Options{
			UseDefaultValue: true,
		}.NewDescritorFromPath("../testdata/idl/example.thrift")
		require.NoError(t, err)
		desc := p.Functions()["ExampleDefaultValue"].Request().Struct().Fields()[0].Type()
		buf := make([]byte, 11)
		BinaryEncoding{}.EncodeString(buf, "default")
		dv := &DefaultValue{
			goValue:      "default",
			jsonValue:    `"default"`,
			thriftBinary: string(buf),
		}
		require.Equal(t, dv, desc.Struct().FieldById(1).DefaultValue())
		buf = make([]byte, 4)
		BinaryEncoding{}.EncodeInt32(buf, int32(1))
		dv = &DefaultValue{
			goValue:      int64(1),
			jsonValue:    `1`,
			thriftBinary: string(buf),
		}
		require.Equal(t, dv, desc.Struct().FieldById(2).DefaultValue())
		buf = make([]byte, 8)
		BinaryEncoding{}.EncodeDouble(buf, float64(1.1))
		dv = &DefaultValue{
			goValue:      float64(1.1),
			jsonValue:    `1.1`,
			thriftBinary: string(buf),
		}
		require.Equal(t, dv, desc.Struct().FieldById(3).DefaultValue())
		dv = &DefaultValue{
			goValue:      true,
			jsonValue:    `true`,
			thriftBinary: string([]byte{1}),
		}
		require.Equal(t, dv, desc.Struct().FieldById(4).DefaultValue())
		require.Equal(t, (*DefaultValue)(nil), desc.Struct().FieldById(5).DefaultValue())
		require.Equal(t, (*DefaultValue)(nil), desc.Struct().FieldById(6).DefaultValue())
		require.Equal(t, (*DefaultValue)(nil), desc.Struct().FieldById(7).DefaultValue())
		buf = make([]byte, 16)
		BinaryEncoding{}.EncodeString(buf, "const string")
		dv = &DefaultValue{
			goValue:      "const string",
			jsonValue:    `"const string"`,
			thriftBinary: string(buf),
		}
		require.Equal(t, dv, desc.Struct().FieldById(8).DefaultValue())
		buf = make([]byte, 4)
		BinaryEncoding{}.EncodeInt32(buf, int32(1))
		dv = &DefaultValue{
			goValue:      int64(1),
			jsonValue:    `1`,
			thriftBinary: string(buf),
		}
		require.Equal(t, dv, desc.Struct().FieldById(9).DefaultValue())
	})
	t.Run("not use", func(t *testing.T) {
		p, err := Options{
			UseDefaultValue: false,
		}.NewDescritorFromPath("../testdata/idl/example.thrift")
		require.NoError(t, err)
		desc := p.Functions()["ExampleDefaultValue"].Request().Struct().Fields()[0].Type()
		require.Equal(t, (*DefaultValue)(nil), desc.Struct().FieldById(1).DefaultValue())
		require.Equal(t, (*DefaultValue)(nil), desc.Struct().FieldById(2).DefaultValue())
		require.Equal(t, (*DefaultValue)(nil), desc.Struct().FieldById(3).DefaultValue())
		require.Equal(t, (*DefaultValue)(nil), desc.Struct().FieldById(4).DefaultValue())
		require.Equal(t, (*DefaultValue)(nil), desc.Struct().FieldById(5).DefaultValue())
		require.Equal(t, (*DefaultValue)(nil), desc.Struct().FieldById(6).DefaultValue())
		require.Equal(t, (*DefaultValue)(nil), desc.Struct().FieldById(7).DefaultValue())
		require.Equal(t, (*DefaultValue)(nil), desc.Struct().FieldById(8).DefaultValue())
		require.Equal(t, (*DefaultValue)(nil), desc.Struct().FieldById(9).DefaultValue())
	})
}

func TestOptionSetOptionalBitmap(t *testing.T) {
	content := `
	namespace go kitex.test.server

	struct Base {
		1: string DefaultField,
		2: optional string OptionalField,
		3: required string RequiredField,
	}

	service InboxService {
		Base ExampleMethod(1: Base req)
	}
	`
	p, err := GetDescFromContent(content, "ExampleMethod", &Options{
		SetOptionalBitmap: true,
	})
	require.NoError(t, err)
	req := p.Request().Struct().Fields()[0].Type()

	require.Equal(t, DefaultRequireness, req.Struct().FieldById(1).Required())
	require.Equal(t, OptionalRequireness, req.Struct().FieldById(2).Required())
	require.Equal(t, RequiredRequireness, req.Struct().FieldById(3).Required())
	require.Equal(t, true, req.Struct().Requires().IsSet(1))
	require.Equal(t, true, req.Struct().Requires().IsSet(2))
	require.Equal(t, true, req.Struct().Requires().IsSet(3))
}

func TestOptionHalfParsing(t *testing.T) {
	content := `
	namespace go kitex.test.server

	struct Base {
		1: string DefaultField,
		2: optional string OptionalField,
		3: required string RequiredField,
	}

	service InboxService {
		Base ExampleMethod(1: Base req)
	}
	`
	p, err := GetDescFromContent(content, "ExampleMethod", &Options{
		ParseFunctionMode: meta.ParseResponseOnly,
	})
	require.NoError(t, err)
	require.Nil(t, p.Request())
	require.NotNil(t, p.Response())

	p, err = GetDescFromContent(content, "ExampleMethod", &Options{
		ParseFunctionMode: meta.ParseRequestOnly,
	})
	require.NoError(t, err)
	require.NotNil(t, p.Request())
	require.Nil(t, p.Response())
}

func TestNewFunctionDescriptorFromContent(t *testing.T) {
	content := `
	namespace go kitex.test.server

	struct Base {
		1: string DefaultField,
		2: optional string OptionalField,
		3: required string RequiredField,
	}

	service InboxService {
		Base Method1(1: Base req)
		Base Method2(1: Base req)
	}
	`
	path := "a/b/main.thrift"
	includes := map[string]string{
		path: content,
	}
	p, err := Options{}.NewDescriptorFromContentWithMethod(path, content, includes, false, "Method1")
	require.NoError(t, err)
	require.NotNil(t, p.Functions()["Method1"])
	require.Nil(t, p.Functions()["Method2"])
}

func TestNewFunctionDescriptorFromPath(t *testing.T) {
	p, err := Options{}.NewDescriptorFromPathWithMethod("../testdata/idl/example.thrift", nil, "ExampleMethod")
	require.NoError(t, err)
	require.NotNil(t, p.Functions()["ExampleMethod"])
	require.Nil(t, p.Functions()["Ping"])
}
