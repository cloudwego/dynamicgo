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

/*
 * Copyright 2021 CloudWeGo Authors
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

	"github.com/cloudwego/dynamicgo/internal/json"
	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/cloudwego/thriftgo/parser"
)

const (
	JSConv                      = thrift.AnnoType(types.VM_JSCONV)
	BodyDynamic thrift.AnnoType = thrift.AnnoType(257)
)

type valueMappingAnnotation struct {
	typ thrift.AnnoID
}

func newValueMappingAnnotation(typ thrift.AnnoID) valueMappingAnnotation {
	return valueMappingAnnotation{
		typ: typ,
	}
}

func (self valueMappingAnnotation) ID() thrift.AnnoID {
	return self.typ
}

func (self valueMappingAnnotation) Make(ctx context.Context, values []parser.Annotation, ast interface{}) (interface{}, error) {
	if len(values) != 1 {
		return nil, fmt.Errorf("valueMappingAnnotation only support one value")
	}
	switch t := self.typ.Type(); t {
	case JSConv:
		return apiJSConv{}, nil
	case BodyDynamic:
		return agwBodyDynamic{}, nil
	default:
		return nil, errNotImplemented(fmt.Sprintf("unsupported type %v of valueMappingAnnotation", t))
	}
}

type agwBodyDynamic struct{}

func (m agwBodyDynamic) Read(ctx context.Context, p *thrift.BinaryProtocol, field *thrift.FieldDescriptor, out *[]byte) error {
	if field.Type().Type() != thrift.STRING {
		return errors.New("body_dynamic only support STRING type")
	}
	b, err := p.ReadBinary(false)
	if err != nil {
		return err
	}
	//FIXME: must validate the json string
	// ok, _ := encoder.Valid(b)
	// if !ok {
	// 	return fmt.Errorf("invalid json: %s", string(b))
	// }
	*out = append(*out, b...)
	return nil
}

func (m agwBodyDynamic) Write(ctx context.Context, p *thrift.BinaryProtocol, field *thrift.FieldDescriptor, in []byte) error {
	if field.Type().Type() != thrift.STRING {
		return errors.New("body_dynamic only support STRING type")
	}
	return p.WriteBinary(in)
}

type apiJSConv struct{}

func (m apiJSConv) Write(ctx context.Context, p *thrift.BinaryProtocol, field *thrift.FieldDescriptor, in []byte) (err error) {
	var val = rt.Mem2Str(in)
	t := field.Type().Type()
	if len(in) >= 2 && in[0] == '"' && in[len(in)-1] == '"' {
		val, err = strconv.Unquote(val)
		if err != nil {
			return err
		}
		if t != thrift.STRING && val == "" {
			val = "0"
		}
	}

	switch t {
	case thrift.I08, thrift.I16, thrift.I32, thrift.I64:
		iv, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		return p.WriteInt(t, int(iv))
	case thrift.DOUBLE:
		dv, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		return p.WriteDouble(dv)
	case thrift.STRING:
		return p.WriteString(val)
	default:
		return fmt.Errorf("unsupported type: %v", t)
	}
}

func (m apiJSConv) Read(ctx context.Context, p *thrift.BinaryProtocol, field *thrift.FieldDescriptor, out *[]byte) error {
	switch field.Type().Type() {
	case thrift.LIST:
		*out = append(*out, '[')
		et, n, err := p.ReadListBegin()
		if err != nil {
			return err
		}
		for i := 0; i < n; i++ {
			err := appendInt(p, thrift.Type(et), out)
			if err != nil {
				return err
			}
			if i != n-1 {
				*out = append(*out, ',')
			}
		}
		*out = append(*out, ']')
	default:
		return appendInt(p, field.Type().Type(), out)
	}
	return nil
}

func appendInt(p *thrift.BinaryProtocol, typ thrift.Type, out *[]byte) error {
	*out = append(*out, '"')
	l := len(*out)
	if cap(*out)-l < types.MaxInt64StringLen {
		rt.GuardSlice(out, types.MaxInt64StringLen)
	}

	switch typ {
	case thrift.BYTE:
		i, err := p.ReadByte()
		if err != nil {
			return err
		}
		*out = json.EncodeInt64(*out, int64(i))
	case thrift.I16:
		i, err := p.ReadI16()
		if err != nil {
			return err
		}
		*out = json.EncodeInt64(*out, int64(i))
	case thrift.I32:
		i, err := p.ReadI32()
		if err != nil {
			return err
		}
		*out = json.EncodeInt64(*out, int64(i))
	case thrift.I64:
		i, err := p.ReadI64()
		if err != nil {
			return err
		}
		*out = json.EncodeInt64(*out, int64(i))
	case thrift.DOUBLE:
		i, err := p.ReadDouble()
		if err != nil {
			return err
		}
		*out = json.EncodeFloat64(*out, float64(i))
	case thrift.STRING:
		s, err := p.ReadString(false)
		if err != nil {
			return err
		}
		*out = append(*out, s...)
	default:
		return meta.NewError(meta.ErrUnsupportedType, fmt.Sprintf("unsupported type: %v", typ), nil)
	}

	*out = append(*out, '"')
	return nil
}
