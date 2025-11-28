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

// FetchAny fetches the value of the field described by desc from any based on go reflect.
func FetchAny(desc *Descriptor, any interface{}, opts ...FetchOptions) (interface{}, error) {
	if any == nil || desc == nil {
		return nil, nil
	}

	var opt FetchOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	v := reflect.ValueOf(any)
	return fetchValue(desc, v, &opt)
}

// ErrNotFound is returned when a field/index/key is not found and DisallowNotFound is enabled
type ErrNotFound struct {
	Parent *Descriptor
	Field  *Field // the field that is not found
	Msg    string // additional message
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("not found %v at %v: %s", e.Field.Name, e.Parent.Name, e.Msg)
}

// FetchOptions contains options for FetchAny
type FetchOptions struct {
	// DisallowNotFound if true, returns ErrNotFound when a field/index/key is not found
	DisallowNotFound bool
}

// structFieldInfo caches field mapping information for a struct type
type structFieldInfo struct {
	fieldIndexMap      map[int]int // thrift field ID -> struct field index
	unknownFieldsIndex int         // index of _unknownFields field, -1 if not present
}

// fieldCache caches the struct field info for each type
var fieldCache sync.Map // map[reflect.Type]*structFieldInfo

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
		if field.Name == "_unknownFields" {
			info.unknownFieldsIndex = i
			continue
		}

		tag := field.Tag.Get("thrift")
		if tag == "" {
			continue
		}

		// Parse thrift tag: "FieldName,ID" - use IndexByte for better performance
		idx := strings.IndexByte(tag, ',')
		if idx < 0 {
			continue
		}

		fieldID, err := strconv.Atoi(tag[idx+1:])
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
func fetchValue(desc *Descriptor, v reflect.Value, opt *FetchOptions) (interface{}, error) {
	// Dereference pointers
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
	}

	switch desc.Kind {
	case TypeKind_Struct:
		return fetchStruct(desc, v, opt)

	case TypeKind_List:
		return fetchList(desc, v, opt)

	case TypeKind_StrMap:
		return fetchStrMap(desc, v, opt)

	default:
		return v.Interface(), nil
	}
}

// fetchStruct handles TypeKind_Struct
func fetchStruct(desc *Descriptor, v reflect.Value, opt *FetchOptions) (interface{}, error) {
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
					return nil, ErrNotFound{Parent: desc, Field: field, Msg: fmt.Sprintf("field ID=%d is nil", field.ID)}
				}
				continue
			}

			// If field has a child descriptor, recursively fetch
			if field.Desc != nil {
				fetched, err := fetchValue(field.Desc, fieldValue, opt)
				if err != nil {
					return nil, err
				}
				result[field.Name] = fetched
			} else {
				// Otherwise, use the value directly
				result[field.Name] = fieldValue.Interface()
			}
		} else if unknownFieldsMap != nil {
			// Try to get field from unknownFields
			if val, ok := unknownFieldsMap[thrift.FieldID(field.ID)]; ok {
				// Convert the value based on the field's Descriptor
				// (e.g., map[FieldID]interface{} -> map[string]interface{} for nested structs)
				result[field.Name] = fetchUnknownValue(val, field.Desc)
			} else if opt.DisallowNotFound {
				return nil, ErrNotFound{Parent: desc, Field: field, Msg: fmt.Sprintf("field ID=%d not found in struct or unknownFields", field.ID)}
			}
		} else if opt.DisallowNotFound {
			return nil, ErrNotFound{Parent: desc, Field: field, Msg: fmt.Sprintf("field ID=%d not found in struct", field.ID)}
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
		idToField := make(map[int]*Field, len(desc.Children))
		for i := range desc.Children {
			idToField[desc.Children[i].ID] = &desc.Children[i]
		}

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

	case TypeKind_List:
		// ReadAny returns []interface{} for LIST type
		listVal, ok := val.([]interface{})
		if !ok {
			return val
		}

		// Find wildcard or indexed descriptors
		var wildcardDesc *Descriptor
		indexDescMap := make(map[int]*Descriptor, len(desc.Children))
		for i := range desc.Children {
			child := &desc.Children[i]
			if child.Name == "*" {
				wildcardDesc = child.Desc
			} else {
				indexDescMap[child.ID] = child.Desc
			}
		}

		result := make([]interface{}, len(listVal))
		for i, elem := range listVal {
			if childDesc, ok := indexDescMap[i]; ok {
				result[i] = fetchUnknownValue(elem, childDesc)
			} else if wildcardDesc != nil {
				result[i] = fetchUnknownValue(elem, wildcardDesc)
			} else {
				result[i] = elem
			}
		}
		return result

	case TypeKind_StrMap:
		// ReadAny returns map[string]interface{} for string-keyed MAP type
		strMap, ok := val.(map[string]interface{})
		if !ok {
			return val
		}

		// Find wildcard or keyed descriptors
		var wildcardDesc *Descriptor
		keyDescMap := make(map[string]*Descriptor, len(desc.Children))
		for i := range desc.Children {
			child := &desc.Children[i]
			if child.Name == "*" {
				wildcardDesc = child.Desc
			} else {
				keyDescMap[child.Name] = child.Desc
			}
		}

		result := make(map[string]interface{}, len(strMap))
		for key, elem := range strMap {
			if childDesc, ok := keyDescMap[key]; ok {
				result[key] = fetchUnknownValue(elem, childDesc)
			} else if wildcardDesc != nil {
				result[key] = fetchUnknownValue(elem, wildcardDesc)
			} else {
				result[key] = elem
			}
		}
		return result

	default:
		return val
	}
}

// fetchList handles TypeKind_List
func fetchList(desc *Descriptor, v reflect.Value, opt *FetchOptions) (interface{}, error) {
	kind := v.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return nil, nil
	}

	childrenLen := len(desc.Children)
	vLen := v.Len()

	// Fast path: only wildcard descriptor
	if childrenLen == 1 && desc.Children[0].Name == "*" {
		wildcardDesc := desc.Children[0].Desc
		result := make([]interface{}, 0, vLen)
		for i := 0; i < vLen; i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Ptr && elem.IsNil() {
				result = append(result, nil)
				continue
			}
			if wildcardDesc != nil {
				fetched, err := fetchValue(wildcardDesc, elem, opt)
				if err != nil {
					return nil, err
				}
				result = append(result, fetched)
			} else {
				result = append(result, elem.Interface())
			}
		}
		return result, nil
	}

	// Build a map of index -> descriptor for quick lookup
	indexDescMap := make(map[int]*Field, childrenLen)
	var wildcardDesc *Field
	for i := range desc.Children {
		child := &desc.Children[i]
		if child.Name == "*" {
			wildcardDesc = child
		} else {
			// Field.ID represents the slice index
			indexDescMap[child.ID] = child
		}
	}

	// Check if specific indices are requested but not available
	if opt.DisallowNotFound {
		for idx := range indexDescMap {
			if idx >= vLen {
				return nil, ErrNotFound{Parent: desc, Field: indexDescMap[idx], Msg: fmt.Sprintf("index %d out of range (len=%d)", idx, vLen)}
			}
		}
	}

	result := make([]interface{}, 0, vLen)
	for i := 0; i < vLen; i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Ptr && elem.IsNil() {
			result = append(result, nil)
			continue
		}

		// First try to find descriptor by index (Field.ID)
		var childDesc *Descriptor
		if field, ok := indexDescMap[i]; ok && field.Desc != nil {
			childDesc = field.Desc
		} else if wildcardDesc != nil && wildcardDesc.Desc != nil {
			// Fallback to wildcard descriptor
			childDesc = wildcardDesc.Desc
		}

		if childDesc != nil {
			fetched, err := fetchValue(childDesc, elem, opt)
			if err != nil {
				return nil, err
			}
			result = append(result, fetched)
		} else {
			result = append(result, elem.Interface())
		}
	}
	return result, nil
}

// fetchStrMap handles TypeKind_StrMap
func fetchStrMap(desc *Descriptor, v reflect.Value, opt *FetchOptions) (interface{}, error) {
	if v.Kind() != reflect.Map || v.Type().Key().Kind() != reflect.String {
		return nil, nil
	}

	childrenLen := len(desc.Children)
	mapLen := v.Len()

	// Fast path: only wildcard descriptor
	if childrenLen == 1 && desc.Children[0].Name == "*" {
		wildcardDesc := desc.Children[0].Desc
		result := make(map[string]interface{}, mapLen)
		iter := v.MapRange()
		for iter.Next() {
			keyStr := iter.Key().String()
			elemValue := iter.Value()

			if elemValue.Kind() == reflect.Ptr && elemValue.IsNil() {
				result[keyStr] = nil
				continue
			}
			if wildcardDesc != nil {
				fetched, err := fetchValue(wildcardDesc, elemValue, opt)
				if err != nil {
					return nil, err
				}
				result[keyStr] = fetched
			} else {
				result[keyStr] = elemValue.Interface()
			}
		}
		return result, nil
	}

	// Build a map of key -> descriptor for quick lookup
	keyDescMap := make(map[string]*Field, childrenLen)
	var wildcardDesc *Field
	for i := range desc.Children {
		child := &desc.Children[i]
		if child.Name == "*" {
			wildcardDesc = child
		} else {
			keyDescMap[child.Name] = child
		}
	}

	// Check if specific keys are requested but not available in the map
	if opt.DisallowNotFound && wildcardDesc == nil {
		for key := range keyDescMap {
			if !v.MapIndex(reflect.ValueOf(key)).IsValid() {
				return nil, ErrNotFound{Parent: desc, Field: keyDescMap[key], Msg: fmt.Sprintf("key '%s' not found in map", key)}
			}
		}
	}

	result := make(map[string]interface{}, mapLen)
	iter := v.MapRange()
	for iter.Next() {
		keyStr := iter.Key().String()
		elemValue := iter.Value()

		if elemValue.Kind() == reflect.Ptr && elemValue.IsNil() {
			result[keyStr] = nil
			continue
		}

		// First try to find descriptor by key, then fallback to wildcard
		var childDesc *Descriptor
		if field, ok := keyDescMap[keyStr]; ok && field.Desc != nil {
			childDesc = field.Desc
		} else if wildcardDesc != nil && wildcardDesc.Desc != nil {
			childDesc = wildcardDesc.Desc
		}

		if childDesc != nil {
			fetched, err := fetchValue(childDesc, elemValue, opt)
			if err != nil {
				return nil, err
			}
			result[keyStr] = fetched
		} else {
			result[keyStr] = elemValue.Interface()
		}
	}
	return result, nil
}
