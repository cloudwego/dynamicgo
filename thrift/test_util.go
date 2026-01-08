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
	"math/rand"
	"time"

	"github.com/cloudwego/dynamicgo/internal/util_test"
)

// GetFnDescFromFile get a fucntion descriptor from idl path (relative to your git root) and
// the function name
func GetFnDescFromFile(filePath, fnName string, opts Options) *FunctionDescriptor {
	svc, err := opts.NewDescritorFromPath(context.Background(), util_test.MustGitPath(filePath))
	if err != nil {
		panic(fmt.Errorf("%s:%s", util_test.MustGitPath(filePath), err))
	}
	fn, err := svc.LookupFunctionByMethod(fnName)
	if err != nil {
		panic(err)
	}
	return fn
}

// FnResponse get the normal response type
func FnResponse(fn *FunctionDescriptor) *TypeDescriptor {
	// let-it-fail: it panic when something is nil
	return fn.Response().Struct().FieldById(0).Type()
}

// FnWholeResponse get the normal response type
func FnWholeResponse(fn *FunctionDescriptor) *TypeDescriptor {
	// let-it-fail: it panic when something is nil
	return fn.Response()
}

// FnRequest
// We assume the request only have one argument and the only argument it the type we want.
func FnRequest(fn *FunctionDescriptor) *TypeDescriptor {
	// let-it-fail: it panic when something is nil
	return fn.Request().Struct().Fields()[0].Type()
}

type FuzzDataOptions struct {
	// MaxDepth is the max depth of the generated data
	MaxDepth int
	// MaxWidth is the max width (map/list/string length) of the generated data
	MaxWidth int
	// DefaultFieldsRatio is the default ratio of fields to be filled in a struct
	DefaultFieldsRatio float64
	// OptionalFieldsRatio is the ratio of optional fields to be filled in a struct
	OptionalFieldsRatio float64
}

// MakeFuzzData generates random data based on the given TypeDescriptor
// It returns a map[string]interface{} like structure for complex types
func MakeFuzzData(reqDesc *TypeDescriptor, opts FuzzDataOptions) (interface{}, error) {
	// Set default options
	if opts.MaxDepth <= 0 {
		opts.MaxDepth = 3
	}
	if opts.MaxWidth <= 0 {
		opts.MaxWidth = 1
	}
	if opts.DefaultFieldsRatio <= 0 {
		opts.DefaultFieldsRatio = 0.8
	}
	if opts.OptionalFieldsRatio <= 0 {
		opts.OptionalFieldsRatio = 0.5
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return makeFuzzValue(reqDesc, opts, rng, 0)
}

func makeFuzzValue(desc *TypeDescriptor, opts FuzzDataOptions, rng *rand.Rand, depth int) (interface{}, error) {
	if desc == nil {
		return nil, nil
	}

	// Check depth limit
	if depth >= opts.MaxDepth {
		return getDefaultValue(desc.Type()), nil
	}

	switch desc.Type() {
	case BOOL:
		return rng.Intn(2) == 1, nil

	case BYTE: // I08 is the same as BYTE
		return int8(rng.Intn(256) - 128), nil

	case I16:
		return int16(rng.Intn(65536) - 32768), nil

	case I32:
		return rng.Int31(), nil

	case I64:
		return rng.Int63(), nil

	case DOUBLE:
		return rng.Float64() * 1000, nil

	case STRING:
		return generateRandomString(rng, opts.MaxWidth), nil

	case LIST:
		return makeFuzzList(desc, opts, rng, depth)

	case SET:
		return makeFuzzList(desc, opts, rng, depth)

	case MAP:
		return makeFuzzMap(desc, opts, rng, depth)

	case STRUCT:
		return makeFuzzStruct(desc, opts, rng, depth)

	default:
		return nil, nil
	}
}

func makeFuzzList(desc *TypeDescriptor, opts FuzzDataOptions, rng *rand.Rand, depth int) (interface{}, error) {
	elemDesc := desc.Elem()
	if elemDesc == nil {
		return []interface{}{}, nil
	}

	size := rng.Intn(opts.MaxWidth)
	list := make([]interface{}, size)

	for i := 0; i < size; i++ {
		val, err := makeFuzzValue(elemDesc, opts, rng, depth+1)
		if err != nil {
			return nil, err
		}
		list[i] = val
	}

	return list, nil
}

func makeFuzzMap(desc *TypeDescriptor, opts FuzzDataOptions, rng *rand.Rand, depth int) (interface{}, error) {
	keyDesc := desc.Key()
	elemDesc := desc.Elem()

	if keyDesc == nil || elemDesc == nil {
		return make(map[string]interface{}), nil
	}

	size := rng.Intn(opts.MaxWidth)

	switch keyDesc.Type() {
	case STRING:
		result := make(map[string]interface{}, size)
		for i := 0; i < size; i++ {
			key := generateRandomString(rng, 8)
			val, err := makeFuzzValue(elemDesc, opts, rng, depth+1)
			if err != nil {
				return nil, err
			}
			result[key] = val
		}
		return result, nil
	case I32:
		result := make(map[int32]interface{}, size)
		for i := 0; i < size; i++ {
			key := int32(i)
			val, err := makeFuzzValue(elemDesc, opts, rng, depth+1)
			if err != nil {
				return nil, err
			}
			result[key] = val
		}
		return result, nil
	case I64:
		result := make(map[int64]interface{}, size)
		for i := 0; i < size; i++ {
			key := int64(i)
			val, err := makeFuzzValue(elemDesc, opts, rng, depth+1)
			if err != nil {
				return nil, err
			}
			result[key] = val
		}
		return result, nil
	case I16:
		result := make(map[int16]interface{}, size)
		for i := 0; i < size; i++ {
			key := int16(i)
			val, err := makeFuzzValue(elemDesc, opts, rng, depth+1)
			if err != nil {
				return nil, err
			}
			result[key] = val
		}
		return result, nil
	case BYTE: // I08
		result := make(map[int8]interface{}, size)
		for i := 0; i < size; i++ {
			key := int8(i)
			val, err := makeFuzzValue(elemDesc, opts, rng, depth+1)
			if err != nil {
				return nil, err
			}
			result[key] = val
		}
		return result, nil
	case BOOL:
		result := make(map[bool]interface{}, 2)
		// ensure we can have up to two entries: true/false
		for i := 0; i < size && i < 2; i++ {
			key := (i%2 == 0)
			val, err := makeFuzzValue(elemDesc, opts, rng, depth+1)
			if err != nil {
				return nil, err
			}
			result[key] = val
		}
		return result, nil
	case DOUBLE:
		result := make(map[float64]interface{}, size)
		for i := 0; i < size; i++ {
			key := rng.Float64()
			val, err := makeFuzzValue(elemDesc, opts, rng, depth+1)
			if err != nil {
				return nil, err
			}
			result[key] = val
		}
		return result, nil
	default:
		// Fallback to string key if unsupported key type
		result := make(map[string]interface{}, size)
		for i := 0; i < size; i++ {
			key := fmt.Sprintf("key_%d_%s", i, generateRandomString(rng, 5))
			val, err := makeFuzzValue(elemDesc, opts, rng, depth+1)
			if err != nil {
				return nil, err
			}
			result[key] = val
		}
		return result, nil
	}
}

func makeFuzzStruct(desc *TypeDescriptor, opts FuzzDataOptions, rng *rand.Rand, depth int) (interface{}, error) {
	if desc.Struct() == nil {
		return make(map[string]interface{}), nil
	}

	result := make(map[string]interface{})
	fields := desc.Struct().Fields()

	for _, field := range fields {
		if field == nil {
			continue
		}

		// Decide whether to include this field based on requireness
		shouldInclude := shouldIncludeField(field, opts, rng)
		if !shouldInclude {
			continue
		}

		fieldDesc := field.Type()
		if fieldDesc == nil {
			continue
		}

		val, err := makeFuzzValue(fieldDesc, opts, rng, depth+1)
		if err != nil {
			return nil, err
		}

		result[field.Name()] = val
	}

	return result, nil
}

func shouldIncludeField(field *FieldDescriptor, opts FuzzDataOptions, rng *rand.Rand) bool {
	req := field.Required()

	// Always include required fields
	if req == RequiredRequireness {
		return true
	}

	// Include default fields with high probability
	if req == DefaultRequireness {
		return rng.Float64() < opts.DefaultFieldsRatio
	}

	// Include optional fields with lower probability
	return rng.Float64() < opts.OptionalFieldsRatio
}

func generateRandomString(rng *rand.Rand, maxLen int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := rng.Intn(maxLen) + 1
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[rng.Intn(len(charset))]
	}
	return string(result)
}

func getDefaultValue(typ Type) interface{} {
	switch typ {
	case BOOL:
		return false
	case BYTE: // I08 is the same as BYTE
		return int8(0)
	case I16:
		return int16(0)
	case I32:
		return int32(0)
	case I64:
		return int64(0)
	case DOUBLE:
		return float64(0)
	case STRING:
		return ""
	case LIST, SET:
		return []interface{}{}
	case MAP:
		return make(map[string]interface{})
	case STRUCT:
		return make(map[string]interface{})
	default:
		return nil
	}
}
