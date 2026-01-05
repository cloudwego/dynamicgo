/**
 * Copyright 2025 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package trim

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/cloudwego/dynamicgo/thrift"
)

type Fetcher struct {
	FetchOptions
}

// FetchOptions contains options for FetchAny
type FetchOptions struct {
	// DisallowNotFound if true, returns ErrNotFound when a field/index/key is not found
	DisallowNotFound bool
}

// FetchAny fetches the value of the field described by desc from any based on go reflect.
//
// Warning: desc must be normalized before calling this method.
func (f Fetcher) FetchAny(desc *Descriptor, any interface{}) (interface{}, error) {
	if any == nil || desc == nil {
		return nil, nil
	}

	// desc.Normalize()

	// Initialize path stack from pool
	stack := getStackFrames()
	defer putStackFrames(stack)

	v := reflect.ValueOf(any)
	return fetchValue(desc, v, &f.FetchOptions, stack)
}

// ErrNotFound is returned when a field/index/key is not found and disallowNotFound is enabled
type ErrNotFound struct {
	Parent *Descriptor
	Field  Field  // the field that is not found
	Msg    string // additional message
}

func (e ErrNotFound) Error() string {
	if e.Msg != "" {
		return fmt.Sprintf("not found %v at %v: %s", e.Field.Name, e.Parent.Type, e.Msg)
	}
	return fmt.Sprintf("not found %v at %v", e.Field.Name, e.Parent.Type)
}

// structFieldInfo caches field mapping information for a struct type
type structFieldInfo struct {
	fieldIndexMap      map[int]int // thrift field ID -> struct field index
	unknownFieldsIndex int         // index of _unknownFields field, -1 if not present
}

// fieldCache caches the struct field info for each type
var fieldCache sync.Map // map[reflect.Type]*structFieldInfo

var (
	thriftUnknownFieldName     = "_unknownFields"
	thriftUnknownFieldNameOnce sync.Once
)

func SetThriftUnknownFieldName(name string) {
	thriftUnknownFieldNameOnce.Do(func() {
		thriftUnknownFieldName = name
	})
}

// getStructFieldInfo returns cached struct field info for the given type
func getStructFieldInfo(t reflect.Type) *structFieldInfo {
	if cached, ok := fieldCache.Load(t); ok {
		return cached.(*structFieldInfo)
	}

	// Build the field info
	numField := t.NumField()
	info := &structFieldInfo{
		fieldIndexMap:      make(map[int]int, numField),
		unknownFieldsIndex: -1,
	}

	for i := 0; i < numField; i++ {
		field := t.Field(i)

		// Check for _unknownFields field
		if field.Name == thriftUnknownFieldName {
			info.unknownFieldsIndex = i
			continue
		}

		tag := field.Tag.Get("thrift")
		if tag == "" {
			continue
		}

		// Parse thrift tag: "FieldName,ID" - use IndexByte for better performance
		idx := strings.Split(tag, ",")
		if len(idx) < 2 {
			continue
		}

		fieldID, err := strconv.Atoi(idx[1])
		if err != nil {
			continue
		}

		info.fieldIndexMap[fieldID] = i
	}

	// Store in cache (use LoadOrStore to handle concurrent initialization)
	actual, _ := fieldCache.LoadOrStore(t, info)
	return actual.(*structFieldInfo)
}

// fetchValue is the internal implementation that works with reflect.Value directly
// to avoid repeated interface{} boxing/unboxing overhead
func fetchValue(desc *Descriptor, v reflect.Value, opt *FetchOptions, stack *pathStack) (interface{}, error) {
	// Dereference pointers
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
	}

	switch desc.Kind {
	case TypeKind_Struct:
		return fetchStruct(desc, v, opt, stack)

	case TypeKind_StrMap:
		return fetchStrMap(desc, v, opt, stack)

	case TypeKind_List:
		return fetchList(desc, v, opt, stack)

	default:
		return v.Interface(), nil
	}
}

// fetchStruct handles TypeKind_Struct
func fetchStruct(desc *Descriptor, v reflect.Value, opt *FetchOptions, stack *pathStack) (interface{}, error) {
	if v.Kind() != reflect.Struct {
		return nil, nil
	}

	result := make(map[string]interface{}, len(desc.Children))

	// Get cached field info for this type
	fieldInfo := getStructFieldInfo(v.Type())

	// Parse unknownFields if present
	var unknownFieldsMap map[thrift.FieldID]interface{}
	if fieldInfo.unknownFieldsIndex >= 0 {
		unknownFieldsValue := v.Field(fieldInfo.unknownFieldsIndex)
		if unknownFieldsValue.Len() > 0 {
			// unknownFields is []byte (unknown.Fields)
			unknownBytes := unknownFieldsValue.Bytes()
			var err error
			unknownFieldsMap, err = parseUnknownFields(unknownBytes)
			if err != nil {
				return nil, err
			}
		}
	}

	// Iterate through descriptor fields
	for i := range desc.Children {
		field := &desc.Children[i]

		// Find struct field by ID (thrift id) using cached index map
		fieldIdx, found := fieldInfo.fieldIndexMap[field.ID]
		if found {
			fieldValue := v.Field(fieldIdx)
			if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
				if opt.DisallowNotFound {
					stack.push(TypeKind_Struct, field.Name, field.ID)
					path := stack.buildPath()
					stack.pop()
					return nil, ErrNotFound{Parent: desc, Field: *field, Msg: fmt.Sprintf("field ID=%d is nil at path %s", field.ID, path)}
				}
				continue
			}

			// Push field onto stack
			stack.push(TypeKind_Struct, field.Name, field.ID)

			// If field has a child descriptor, recursively fetch
			var err error
			if field.Desc != nil {
				var fetched interface{}
				fetched, err = fetchValue(field.Desc, fieldValue, opt, stack)
				if err == nil {
					result[field.Name] = fetched
				}
			} else {
				// Otherwise, use the value directly
				result[field.Name] = fieldValue.Interface()
			}

			// Pop field from stack
			stack.pop()

			if err != nil {
				return nil, err
			}
		} else if unknownFieldsMap != nil {
			// Try to get field from unknownFields
			if val, ok := unknownFieldsMap[thrift.FieldID(field.ID)]; ok {
				// Convert the value based on the field's Descriptor
				// (e.g., map[FieldID]interface{} -> map[string]interface{} for nested structs)
				result[field.Name] = fetchUnknownValue(val, field.Desc)
			} else if opt.DisallowNotFound {
				stack.push(TypeKind_Struct, field.Name, field.ID)
				path := stack.buildPath()
				stack.pop()
				return nil, ErrNotFound{Parent: desc, Field: *field, Msg: fmt.Sprintf("field ID=%d not found in struct or unknownFields at path %s", field.ID, path)}
			}
		} else if opt.DisallowNotFound {
			stack.push(TypeKind_Struct, field.Name, field.ID)
			path := stack.buildPath()
			stack.pop()
			return nil, ErrNotFound{Parent: desc, Field: *field, Msg: fmt.Sprintf("field ID=%d not found in struct at path %s", field.ID, path)}
		}
	}
	return result, nil
}

// parseUnknownFields parses thrift binary encoded unknown fields and returns a map of field ID to value
func parseUnknownFields(data []byte) (map[thrift.FieldID]interface{}, error) {
	if len(data) == 0 {
		return nil, nil
	}

	result := make(map[thrift.FieldID]interface{})
	p := thrift.BinaryProtocol{Buf: data}

	for p.Read < len(p.Buf) {
		// Read field header
		_, fieldType, fieldID, err := p.ReadFieldBegin()
		if err != nil {
			return nil, err
		}
		if fieldType == thrift.STOP {
			break
		}

		// Read field value using ReadAny
		val, err := p.ReadAny(fieldType, false, false)
		if err != nil {
			return nil, err
		}

		result[fieldID] = val
	}

	return result, nil
}

// fetchUnknownValue converts the value from unknownFields based on the field's Descriptor.
// For struct types, ReadAny returns map[FieldID]interface{}, which needs to be converted
// to map[string]interface{} using the Descriptor's Children field names.
func fetchUnknownValue(val interface{}, desc *Descriptor) interface{} {
	if desc == nil {
		return val
	}

	switch desc.Kind {
	case TypeKind_Struct:
		// ReadAny returns map[FieldID]interface{} for STRUCT type
		fieldIDMap, ok := val.(map[thrift.FieldID]interface{})
		if !ok {
			return val
		}

		// Build a map from field ID to Field for quick lookup
		idToField := desc.ids

		// Convert map[FieldID]interface{} to map[string]interface{}
		result := make(map[string]interface{}, len(fieldIDMap))
		for fieldID, fieldVal := range fieldIDMap {
			if field, ok := idToField[int(fieldID)]; ok {
				// Recursively convert nested values
				result[field.Name] = fetchUnknownValue(fieldVal, field.Desc)
			}
			// Fields not in descriptor are ignored
		}
		return result

	case TypeKind_StrMap:
		// ReadAny returns map[string]interface{} for string-keyed MAP type
		strMap, ok := val.(map[string]interface{})
		if !ok {
			return val
		}

		// Find wildcard or keyed descriptors
		keyDescMap := desc.names

		result := make(map[string]interface{}, len(strMap))
		for key, elem := range strMap {
			if childDesc, ok := keyDescMap[key]; ok {
				result[key] = fetchUnknownValue(elem, childDesc.Desc)
			} else if len(desc.Children) == 1 && desc.Children[0].Name == "*" {
				result[key] = fetchUnknownValue(elem, desc.Children[0].Desc)
			} else {
				result[key] = elem
			}
		}
		return result

	case TypeKind_List:
		// ReadAny returns []interface{} for LIST type
		list, ok := val.([]interface{})
		if !ok {
			return val
		}

		// Check if wildcard descriptor ("*" means all elements)
		if len(desc.Children) == 1 && desc.Children[0].Name == "*" {
			childDesc := desc.Children[0].Desc
			result := make([]interface{}, len(list))
			for i, elem := range list {
				result[i] = fetchUnknownValue(elem, childDesc)
			}
			return result
		}

		// Specific indices requested
		result := make([]interface{}, 0, len(desc.Children))
		for _, child := range desc.Children {
			idx := child.ID
			if idx < 0 || idx >= len(list) {
				continue
			}
			result = append(result, fetchUnknownValue(list[idx], child.Desc))
		}
		return result

	default:
		return val
	}
}

// fetchStrMap handles TypeKind_StrMap
func fetchStrMap(desc *Descriptor, v reflect.Value, opt *FetchOptions, stack *pathStack) (interface{}, error) {
	if v.Kind() != reflect.Map || v.Type().Key().Kind() != reflect.String {
		return nil, nil
	}

	childrenLen := len(desc.Children)

	// Fast path: only wildcard descriptor
	if childrenLen == 1 && desc.Children[0].Name == "*" {
		wildcardDesc := desc.Children[0].Desc
		result := make(map[string]interface{}, v.Len())
		iter := v.MapRange()
		for iter.Next() {
			keyStr := iter.Key().String()
			elemValue := iter.Value()

			if elemValue.Kind() == reflect.Ptr && elemValue.IsNil() {
				result[keyStr] = nil
				continue
			}

			// Push map key onto stack
			stack.push(TypeKind_StrMap, keyStr, 0)

			var err error
			if wildcardDesc != nil {
				var fetched interface{}
				fetched, err = fetchValue(wildcardDesc, elemValue, opt, stack)
				if err == nil {
					result[keyStr] = fetched
				}
			} else {
				result[keyStr] = elemValue.Interface()
			}

			// Pop map key from stack
			stack.pop()

			if err != nil {
				return nil, err
			}
		}
		return result, nil
	}

	// range over children
	keyDescMap := desc.names
	result := make(map[string]interface{}, childrenLen)
	for key, child := range keyDescMap {
		val := v.MapIndex(reflect.ValueOf(key))
		// Check if specific keys are requested but not available in the map
		if !val.IsValid() {
			if opt.DisallowNotFound {
				stack.push(TypeKind_StrMap, key, 0)
				path := stack.buildPath()
				stack.pop()
				return nil, ErrNotFound{Parent: desc, Field: keyDescMap[key], Msg: fmt.Sprintf("key '%s' not found in map at path %s", key, path)}
			} else {
				continue
			}
		}
		if val.Kind() == reflect.Ptr && val.IsNil() {
			result[key] = nil
			continue
		}

		// Push map key onto stack
		stack.push(TypeKind_StrMap, key, 0)

		var err error
		if child.Desc != nil {
			var fetched interface{}
			fetched, err = fetchValue(child.Desc, val, opt, stack)
			if err == nil {
				result[key] = fetched
			}
		} else {
			result[key] = val.Interface()
		}

		// Pop map key from stack
		stack.pop()

		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// fetchList handles TypeKind_List
func fetchList(desc *Descriptor, v reflect.Value, opt *FetchOptions, stack *pathStack) (interface{}, error) {
	kind := v.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return nil, nil
	}

	childrenLen := len(desc.Children)

	// Fast path: only wildcard descriptor ("*" means all elements)
	if childrenLen == 1 && desc.Children[0].Name == "*" {
		wildcardDesc := desc.Children[0].Desc
		listLen := v.Len()
		result := make([]interface{}, 0, listLen)

		for i := 0; i < listLen; i++ {
			elemValue := v.Index(i)

			if elemValue.Kind() == reflect.Ptr && elemValue.IsNil() {
				result = append(result, nil)
				continue
			}

			// Push list index onto stack
			stack.push(TypeKind_List, "*", i)

			var err error
			if wildcardDesc != nil {
				var fetched interface{}
				fetched, err = fetchValue(wildcardDesc, elemValue, opt, stack)
				if err == nil {
					result = append(result, fetched)
				}
			} else {
				result = append(result, elemValue.Interface())
			}

			// Pop list index from stack
			stack.pop()

			if err != nil {
				return nil, err
			}
		}
		return result, nil
	}

	// Specific indices requested
	result := make([]interface{}, 0, childrenLen)
	for _, child := range desc.Children {
		// Use Field.ID as the index
		idx := child.ID

		// Check if index is out of bounds
		if idx < 0 || idx >= v.Len() {
			if opt.DisallowNotFound {
				stack.push(TypeKind_List, "", idx)
				path := stack.buildPath()
				stack.pop()
				return nil, ErrNotFound{Parent: desc, Field: child, Msg: fmt.Sprintf("index %d out of bounds (len=%d) at path %s", idx, v.Len(), path)}
			}
			continue
		}

		elemValue := v.Index(idx)
		if elemValue.Kind() == reflect.Ptr && elemValue.IsNil() {
			result = append(result, nil)
			continue
		}

		// Push list index onto stack
		stack.push(TypeKind_List, "", idx)

		var err error
		if child.Desc != nil {
			var fetched interface{}
			fetched, err = fetchValue(child.Desc, elemValue, opt, stack)
			if err == nil {
				result = append(result, fetched)
			}
		} else {
			result = append(result, elemValue.Interface())
		}

		// Pop list index from stack
		stack.pop()

		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
