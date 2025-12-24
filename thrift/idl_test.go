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
	"math"
	"path/filepath"
	"testing"

	"github.com/cloudwego/thriftgo/parser"
	"github.com/stretchr/testify/require"
)

func TestThriftContentWithAbsIncludePath(t *testing.T) {
	path := "a/b/main.thrift"
	content := `
	namespace go kitex.test.server
	include "x.thrift"
	include "../y.thrift" 

	service InboxService {
		void Echo(1: x.A req)
	}
	`
	includes := map[string]string{
		path: content,
		"a/b/x.thrift": `namespace go kitex.test1.server
struct A {
	1: string A
}
`,
		"x.thrift": `namespace go kitex.test2.server
struct A {
	2: i64 B
}		
`,
		"a/y.thrift": `
		namespace go kitex.test.server
		include "z.thrift"
		`,
		"a/z.thrift": "namespace go kitex.test.server",
	}
	p, err := NewDescritorFromContent(context.Background(), path, content, includes, true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", p)
	fieldA := p.functions["Echo"].request.Struct().FieldByKey("req").Type().Struct().FieldByKey("A")
	require.NotNil(t, fieldA)
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
	p, err := NewDescritorFromContent(context.Background(), path, content, includes, true)
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
	p, err := NewDescritorFromPath(context.Background(), "../testdata/idl/example3.thrift")
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
		string ExampleMethod(1: string req) (anno1 = "", anno.test = "中文1", anno.test = "中文2")
	} (anno1 = "", anno.test = "中文1", anno.test = "中文2")
	`

	includes := map[string]string{
		path: content,
	}
	p, err := NewDescritorFromContent(context.Background(), path, content, includes, true)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, []parser.Annotation{{Key: "anno1", Values: []string{""}}, {Key: "anno.test", Values: []string{"中文1", "中文2"}}}, p.Annotations())
	require.Equal(t, []parser.Annotation{{Key: "anno1", Values: []string{""}}, {Key: "anno.test", Values: []string{"中文1", "中文2"}}}, p.Functions()["ExampleMethod"].Annotations())
}

func GetDescFromContent(content string, method string, opts *Options) (*FunctionDescriptor, error) {
	path := "a/b/main.thrift"
	includes := map[string]string{
		path: content,
	}
	p, err := opts.NewDescritorFromContent(context.Background(), path, content, includes, true)
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
		}.NewDescritorFromPath(context.Background(), "../testdata/idl/example.thrift")
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
		}.NewDescritorFromPath(context.Background(), "../testdata/idl/example.thrift")
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

func TestOptionPutNameSpaceToAnnotation(t *testing.T) {
	content := `
	namespace py py.base
	namespace go go.base

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
		PutNameSpaceToAnnotation: true,
	})
	require.NoError(t, err)
	req := p.Request().Struct().Fields()[0].Type()
	annos := req.Struct().Annotations()
	var ns *parser.Annotation
	for i, a := range annos {
		if a.Key == NameSpaceAnnotationKey {
			ns = &annos[i]
			break
		}
	}
	require.NotNil(t, ns)
	require.Equal(t, ns.Values, []string{"py", "py.base", "go", "go.base"})
}

func TestOptionPutThriftFilenameToAnnotation(t *testing.T) {
	path := filepath.Join("..", "testdata", "idl", "example.thrift")
	opt := Options{PutThriftFilenameToAnnotation: true}
	descriptor, err := opt.NewDescritorFromPath(context.Background(), path)
	require.NoError(t, err)
	method, err := descriptor.LookupFunctionByMethod("ExampleMethod")
	require.NoError(t, err)
	req := method.Request().Struct().Fields()[0].Type()
	annos := req.Struct().Annotations()
	var filename *parser.Annotation
	for i, a := range annos {
		if a.Key == FilenameAnnotationKey {
			filename = &annos[i]
			break
		}
	}
	require.Equal(t, filename.Values, []string{path})
}

func TestNewFunctionDescriptorFromContent_absPath(t *testing.T) {
	content := `
	include "/a/b/main.thrift"
	include "/ref.thrift"

	namespace go kitex.test.server

	struct Base {
		1: string DefaultField = ref.ConstString,
		2: optional string OptionalField,
		3: required string RequiredField,
	}

	service InboxService {
		Base Method1(1: Base req)
		Base Method2(1: Base req)
	}
	`
	ref := `
	include "/a/b/main.thrift"

	namespace go ref

	const string ConstString = "const string"
	`
	path := "/a/b/main.thrift"
	includes := map[string]string{
		path:          content,
		"/ref.thrift": ref,
	}

	p, err := Options{}.NewDescriptorFromContentWithMethod(context.Background(), path, content, includes, false, "Method1")
	require.NoError(t, err)
	require.NotNil(t, p.Functions()["Method1"])
	require.Nil(t, p.Functions()["Method2"])
}

func TestNewFunctionDescriptorFromContent_relativePath(t *testing.T) {
	content := `
	include "main.thrift"
	include "ref.thrift"

	namespace go kitex.test.server
	
	struct Base {
		1: string DefaultField = ref.ConstString,
		2: optional string OptionalField,
		3: required string RequiredField,
	}

	service InboxService {
		Base Method1(1: Base req)
		Base Method2(1: Base req)
	}
	`
	ref := `
	include "/a/b/main.thrift"

	namespace go ref

	const string ConstString = "const string"
	`
	path := "/a/b/main.thrift"
	includes := map[string]string{
		path:              content,
		"/a/b/ref.thrift": ref,
	}
	p, err := Options{}.NewDescriptorFromContentWithMethod(context.Background(), path, content, includes, true, "Method1")
	require.NoError(t, err)
	require.NotNil(t, p.Functions()["Method1"])
	require.Nil(t, p.Functions()["Method2"])
}

func TestNewFunctionDescriptorFromPath(t *testing.T) {
	p, err := Options{}.NewDescriptorFromPathWithMethod(context.Background(), "../testdata/idl/example.thrift", nil, "ExampleMethod")
	require.NoError(t, err)
	require.NotNil(t, p.Functions()["ExampleMethod"])
	require.Nil(t, p.Functions()["Ping"])
}

func TestStreamingFunctionDescriptorFromContent(t *testing.T) {
	path := "a/b/main.thrift"
	content := `
	namespace go thrift
	
	struct Request {
		1: required string message,
	}
	
	struct Response {
		1: required string message,
	}
	
	service TestService {
		Response Echo (1: Request req) (streaming.mode="bidirectional"),
		Response EchoClient (1: Request req) (streaming.mode="client"),
		Response EchoServer (1: Request req) (streaming.mode="server"),
		Response EchoUnary (1: Request req) (streaming.mode="unary"), // not recommended
		Response EchoBizException (1: Request req) (streaming.mode="client"),
	
		Response EchoPingPong (1: Request req), // KitexThrift, non-streaming
	}
	`
	dsc, err := NewDescritorFromContent(context.Background(), path, content, nil, false)
	require.Nil(t, err)
	require.Equal(t, true, dsc.Functions()["Echo"].IsWithoutWrapping())
	require.Equal(t, true, dsc.Functions()["EchoClient"].IsWithoutWrapping())
	require.Equal(t, true, dsc.Functions()["EchoServer"].IsWithoutWrapping())
	require.Equal(t, false, dsc.Functions()["EchoUnary"].IsWithoutWrapping())
	require.Equal(t, true, dsc.Functions()["EchoBizException"].IsWithoutWrapping())
	require.Equal(t, false, dsc.Functions()["EchoPingPong"].IsWithoutWrapping())
	require.Equal(t, "Request", dsc.Functions()["EchoClient"].Request().Struct().Name())
	require.Equal(t, "", dsc.Functions()["EchoUnary"].Request().Struct().Name())
}

func TestParseWithServiceName(t *testing.T) {
	path := "a/b/main.thrift"
	content := `
	namespace go thrift
	
	struct Request {
		1: required string message,
	}
	
	struct Response {
		1: required string message,
	}
	
	service Service1 {
		Response Test(1: Request req)
	}

	service Service2 {
		Response Test(1: Request req)
	}

	service Service3 {
		Response Test(1: Request req)
	}
	`

	opts := Options{ServiceName: "Service2"}
	p, err := opts.NewDescritorFromContent(context.Background(), path, content, nil, false)
	require.Nil(t, err)
	require.Equal(t, p.Name(), "Service2")

	opts = Options{ServiceName: "UnknownService"}
	p, err = opts.NewDescritorFromContent(context.Background(), path, content, nil, false)
	require.NotNil(t, err)
	require.Nil(t, p)
	require.Equal(t, err.Error(), "the idl service name UnknownService is not in the idl. Please check your idl")
}

func TestNewDescriptorByName(t *testing.T) {
	t.Run("WithIncludes", func(t *testing.T) {
		// Test cross-file struct parsing
		mainPath := "main.thrift"
		mainContent := `
	namespace go kitex.test.server
	include "base.thrift"
	
	struct UserRequest {
		1: required string username
		2: base.User user
	}
	`

		basePath := "base.thrift"
		baseContent := `
	namespace go kitex.test.server
	
	struct User {
		1: required i64 id
		2: required string name
		3: optional string email
	}
	`

		includes := map[string]string{
			mainPath: mainContent,
			basePath: baseContent,
		}

		opts := Options{}

		// Test parsing UserRequest which references User from another file
		desc, err := opts.NewDescriptorByName(context.Background(), mainPath, "UserRequest", includes)
		require.NoError(t, err)
		require.NotNil(t, desc)
		require.Equal(t, STRUCT, desc.Type())
		require.Equal(t, "UserRequest", desc.Struct().Name())

		// Verify fields
		usernameField := desc.Struct().FieldByKey("username")
		require.NotNil(t, usernameField)
		require.Equal(t, STRING, usernameField.Type().Type())

		userField := desc.Struct().FieldByKey("user")
		require.NotNil(t, userField)
		require.Equal(t, STRUCT, userField.Type().Type())
		require.Equal(t, "User", userField.Type().Struct().Name())

		// Verify nested User struct fields
		userStruct := userField.Type().Struct()
		idField := userStruct.FieldByKey("id")
		require.NotNil(t, idField)
		require.Equal(t, I64, idField.Type().Type())

		nameField := userStruct.FieldByKey("name")
		require.NotNil(t, nameField)
		require.Equal(t, STRING, nameField.Type().Type())

		emailField := userStruct.FieldByKey("email")
		require.NotNil(t, emailField)
		require.Equal(t, STRING, emailField.Type().Type())
	})

	t.Run("PackageNotation", func(t *testing.T) {
		// Test cross-file reference using package notation (e.g., "base.User")
		mainPath := "main.thrift"
		mainContent := `
	namespace go kitex.test.server
	include "base.thrift"
	`

		basePath := "base.thrift"
		baseContent := `
	namespace go kitex.test.base
	
	struct User {
		1: required i64 id
		2: required string name
	}
	`

		includes := map[string]string{
			mainPath: mainContent,
			basePath: baseContent,
		}

		opts := Options{}

		// Test parsing User struct with cross-file reference notation
		desc, err := opts.NewDescriptorByName(context.Background(), mainPath, "base.User", includes)
		// verify file and namespace correctness
		require.Equal(t, "base.thrift", desc.File())
		require.Equal(t, "kitex.test.base", desc.Namespace())

		require.NoError(t, err)
		require.NotNil(t, desc)
		require.Equal(t, STRUCT, desc.Type())
		require.Equal(t, "User", desc.Struct().Name())

		// Verify fields
		idField := desc.Struct().FieldByKey("id")
		require.NotNil(t, idField)
		require.Equal(t, I64, idField.Type().Type())
	})

	// Test with mixed relative and absolute paths
	t.Run("WithMixedPaths", func(t *testing.T) {
		mainPath := "a/b/main.thrift"
		mainContent := `
	namespace go kitex.test.server
	include "../c/base.thrift"
	include "a/b/ref.thrift"
	
	struct UserRequest {
		1: required string username
		2: base.User user
		3: ref.Placeholder placeholder
	}
	`

		basePath := "a/c/base.thrift"
		baseContent := `
	namespace go kitex.test.server
	
	struct User {
		1: required i64 id
		2: required string name
	}
	`
		refPath := "a/b/ref.thrift"
		refContent := `
	namespace go kitex.test.server
	
	struct Placeholder {
		1: required string info
	}
	`

		includes := map[string]string{
			mainPath: mainContent,
			basePath: baseContent,
			refPath:  refContent,
		}

		opts := Options{}

		// Test parsing UserRequest with relative path that goes up directories
		desc, err := opts.NewDescriptorByName(context.Background(), mainPath, "UserRequest", includes)
		// verify file and namespace correctness
		require.Equal(t, "a/b/main.thrift", desc.File())
		require.Equal(t, "kitex.test.server", desc.Namespace())

		require.NoError(t, err)
		require.NotNil(t, desc)
		require.Equal(t, STRUCT, desc.Type())
		require.Equal(t, "UserRequest", desc.Struct().Name())

		// Verify nested User struct is correctly resolved via relative path
		userField := desc.Struct().FieldByKey("user")
		require.NotNil(t, userField)
		require.Equal(t, STRUCT, userField.Type().Type())
		require.Equal(t, "User", userField.Type().Struct().Name())

		// Verify nested Placeholder struct is correctly resolved via direct relative path
		placeholderField := desc.Struct().FieldByKey("placeholder")
		require.NotNil(t, placeholderField)
		require.Equal(t, STRUCT, placeholderField.Type().Type())
		require.Equal(t, "Placeholder", placeholderField.Type().Struct().Name())
	})
}
