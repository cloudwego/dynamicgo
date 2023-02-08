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
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/stretchr/testify/require"
)

func TestAPIJSConv(t *testing.T) {
	p, err := thrift.NewDescritorFromPath(context.Background(), "../../testdata/idl/example3.thrift")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#+v\n", p)
	req := p.Functions()["ExampleMethod"].Request().Struct().FieldByKey("req")
	h := req.Type().Struct().FieldById(1)
	hm := h.HTTPMappings()
	require.Equal(t, []thrift.HttpMapping(nil), hm)
	vm := req.Type().Struct().FieldById(6).ValueMappingType()
	require.Equal(t, JSConv, vm)
}

func init() {
	thrift.RegisterAnnotation(NewValueMappingAnnotation2(thrift.MakeAnnoID(thrift.AnnoKindValueMapping, thrift.AnnoScopeField, APIJSConv2)), "test.js_conv2")
}

const APIJSConv2 = thrift.AnnoType(999)

type ValueMappingAnnotation2 struct {
	typ thrift.AnnoID
}

func NewValueMappingAnnotation2(typ thrift.AnnoID) ValueMappingAnnotation2 {
	return ValueMappingAnnotation2{
		typ: typ,
	}
}

func (self ValueMappingAnnotation2) ID() thrift.AnnoID {
	return self.typ
}

func (self ValueMappingAnnotation2) Make(ctx context.Context, values []parser.Annotation, ast interface{}) (interface{}, error) {
	if len(values) != 1 {
		return nil, errors.New("ValueMappingAnnotation2 must have one value")
	}
	switch self.typ.Type() {
	case APIJSConv2:
		return apiJSConv2{}, nil
	default:
		return nil, ErrNotImplemented
	}
}

type apiJSConv2 struct {
	apiJSConv
}

func (m apiJSConv2) Write(ctx context.Context, p *thrift.BinaryProtocol, field *thrift.FieldDescriptor, in []byte) error {
	if len(in) == 0 {
		return errors.New("empty value")
	}
	if in[0] == '"' {
		in = in[1 : len(in)-1]
	}
	switch t := field.Type().Type(); t {
	case thrift.BYTE, thrift.I16, thrift.I32, thrift.I64:
		i, err := strconv.ParseInt(rt.Mem2Str(in), 10, 64)
		if err != nil {
			return err
		}
		if err := p.WriteInt(t, int(i)); err != nil {
			return err
		}
		return nil
	case thrift.DOUBLE:
		f, err := strconv.ParseFloat(rt.Mem2Str(in), 64)
		if err != nil {
			return err
		}
		if err := p.WriteDouble(f); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported type %s", t)
	}
}

func TestJSConv2(t *testing.T) {
	path := "a/b/main.thrift"
	content := `
	namespace go kitex.test.server
	struct Base {
		1: required string required_field
		2: optional string test1 (test.js_conv2="")
		3: string pass 
	}

	service ExampleService {
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
	require.Equal(t, apiJSConv2{}, req.Struct().FieldById(2).ValueMapping())
	require.Equal(t, APIJSConv2, req.Struct().FieldById(2).ValueMappingType())
}
