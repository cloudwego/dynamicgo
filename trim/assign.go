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

	"github.com/cloudwego/dynamicgo/proto/binary"
	"github.com/cloudwego/dynamicgo/proto/protowire"
)

// AssignOptions contains options for AssignAny
type AssignOptions struct {
	// DisallowNotDefined if true, returns ErrNotFound when a field/index/key is not found
	DisallowNotDefined bool
}

type AssignOption func(*AssignOptions)

// WithDisallowNotDefined sets the DisallowNotFound option
func WithDisallowNotDefined(disallow bool) AssignOption {
	return func(o *AssignOptions) {
		o.DisallowNotDefined = disallow
	}
}

// pbStructFieldInfo caches field mapping information for a protobuf struct type
type pbStructFieldInfo struct {
	// nameToFieldIndex maps field name (from protobuf tag) to struct field index
	nameToFieldIndex map[string]int
	// nameToFieldID maps field name to protobuf field ID
	nameToFieldID map[string]int
	// idToFieldIndex maps protobuf field ID to struct field index
	idToFieldIndex map[int]int
	// unrecognizedIndex is the index of XXX_unrecognized field, -1 if not present
	unrecognizedIndex int
}

// pbFieldCache caches the struct field info for each type
var pbFieldCache sync.Map // map[reflect.Type]*pbStructFieldInfo

// getPBStructFieldInfo returns cached struct field info for the given protobuf type
func getPBStructFieldInfo(t reflect.Type) *pbStructFieldInfo {
	if cached, ok := pbFieldCache.Load(t); ok {
		return cached.(*pbStructFieldInfo)
	}

	// Build the field info
	numField := t.NumField()
	info := &pbStructFieldInfo{
		nameToFieldIndex:  make(map[string]int, numField),
		nameToFieldID:     make(map[string]int, numField),
		idToFieldIndex:    make(map[int]int, numField),
		unrecognizedIndex: -1,
	}

	for i := 0; i < numField; i++ {
		field := t.Field(i)

		// Check for XXX_unrecognized field
		if field.Name == "XXX_unrecognized" {
			info.unrecognizedIndex = i
			continue
		}

		tag := field.Tag.Get("protobuf")
		if tag == "" {
			continue
		}

		// Parse protobuf tag: "varint,1,req,name=field_a" or "bytes,2,opt,name=field_b"
		// Format: wireType,fieldID,cardinality,name=fieldName,...
		parts := strings.Split(tag, ",")
		if len(parts) < 4 {
			continue
		}

		// Parse field ID (second part)
		fieldID, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		// Parse field name (look for name=xxx)
		var fieldName string
		for _, part := range parts[3:] {
			if strings.HasPrefix(part, "name=") {
				fieldName = strings.TrimPrefix(part, "name=")
				break
			}
		}
		if fieldName == "" {
			continue
		}

		info.nameToFieldIndex[fieldName] = i
		info.nameToFieldID[fieldName] = fieldID
		info.idToFieldIndex[fieldID] = i
	}

	// Store in cache (use LoadOrStore to handle concurrent initialization)
	actual, _ := pbFieldCache.LoadOrStore(t, info)
	return actual.(*pbStructFieldInfo)
}

// AssignAny assigns values from src (map[string]interface{}) to dest (protobuf struct) according to desc.
// For fields that exist in src but not in dest's struct definition, they will be encoded
// to XXX_unrecognized field using protobuf binary encoding.
func AssignAny(desc *Descriptor, src interface{}, dest interface{}, opts ...AssignOption) error {
	if src == nil || dest == nil || desc == nil {
		return nil
	}

	desc.Normalize()

	var opt AssignOptions
	for _, o := range opts {
		o(&opt)
	}

	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer to struct")
	}

	return assignValue(desc, src, destValue.Elem(), &opt)
}

// assignValue is the internal implementation that works with reflect.Value directly
func assignValue(desc *Descriptor, src interface{}, destValue reflect.Value, opt *AssignOptions) error {
	if src == nil {
		return nil
	}

	// Dereference pointers on dest
	for destValue.Kind() == reflect.Ptr {
		if destValue.IsNil() {
			// Allocate new value
			destValue.Set(reflect.New(destValue.Type().Elem()))
		}
		destValue = destValue.Elem()
	}

	switch desc.Kind {
	case TypeKind_Struct:
		return assignStruct(desc, src, destValue, opt)
	case TypeKind_StrMap:
		return assignStrMap(desc, src, destValue, opt)
	default:
		return assignScalar(src, destValue)
	}
}

// assignStruct handles TypeKind_Struct assignment
func assignStruct(desc *Descriptor, src interface{}, destValue reflect.Value, opt *AssignOptions) error {
	srcMap, ok := src.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map[string]interface{} for struct, got %T", src)
	}

	if destValue.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct destination, got %v", destValue.Kind())
	}

	// Get cached field info for this type
	fieldInfo := getPBStructFieldInfo(destValue.Type())

	// Track which fields in srcMap are assigned to struct fields
	unassignedFields := make(map[string]interface{}, len(srcMap))

	// Iterate through srcMap and assign values
	descFieldMap := desc.names
	for key, value := range srcMap {
		if value == nil {
			continue
		}

		// Find the descriptor field for this key
		descField, hasDescField := descFieldMap[key]

		// Find struct field by name using cached index map
		fieldIdx, found := fieldInfo.nameToFieldIndex[key]
		if found {
			fieldValue := destValue.Field(fieldIdx)

			// Make sure field is settable
			if !fieldValue.CanSet() {
				continue
			}

			// If field has a child descriptor, recursively assign
			if hasDescField && descField.Desc != nil {
				if err := assignValueToField(descField.Desc, value, fieldValue, opt); err != nil {
					return err
				}
			} else {
				// Otherwise, assign the value directly
				if err := assignScalar(value, fieldValue); err != nil {
					return err
				}
			}
		} else if hasDescField {
			// Field exists in descriptor but not in struct - encode to XXX_unrecognized
			// This will be handled below
			unassignedFields[key] = value
		} else if opt.DisallowNotDefined {
			return ErrNotFound{Parent: desc, Field: Field{Name: key}, Msg: fmt.Sprintf("field '%s' not found in struct", key)}
		}
	}

	// Encode unassigned fields (from descriptor) to XXX_unrecognized
	if len(unassignedFields) > 0 && fieldInfo.unrecognizedIndex >= 0 {
		unrecognizedValue := destValue.Field(fieldInfo.unrecognizedIndex)
		if unrecognizedValue.CanSet() {
			bp := binary.NewBinaryProtocolBuffer()
			defer binary.FreeBinaryProtocol(bp)

			for key, val := range unassignedFields {
				// Encode this field to XXX_unrecognized
				if err := encodeUnknownField(bp, descFieldMap[key].ID, val); err != nil {
					return fmt.Errorf("failed to encode unknown field '%s': %w", key, err)
				}
			}

			if len(bp.Buf) > 0 {
				// Append to existing XXX_unrecognized if any
				existingBytes := unrecognizedValue.Bytes()
				newBytes := make([]byte, len(existingBytes)+len(bp.Buf))
				copy(newBytes, existingBytes)
				copy(newBytes[len(existingBytes):], bp.Buf)
				unrecognizedValue.SetBytes(newBytes)
			}
		}
	}

	return nil
}

// assignValueToField assigns a value to a field, handling pointer allocation
func assignValueToField(desc *Descriptor, src interface{}, fieldValue reflect.Value, opt *AssignOptions) error {
	// Handle pointer fields - allocate if needed
	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
		}
		return assignValue(desc, src, fieldValue.Elem(), opt)
	}
	return assignValue(desc, src, fieldValue, opt)
}

// assignStrMap handles TypeKind_StrMap assignment
func assignStrMap(desc *Descriptor, src interface{}, destValue reflect.Value, opt *AssignOptions) error {
	srcMap, ok := src.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map[string]interface{} for strmap, got %T", src)
	}

	if destValue.Kind() != reflect.Map {
		return fmt.Errorf("expected map destination, got %v", destValue.Kind())
	}

	// Find wildcard or keyed descriptors
	var wildcardDesc *Descriptor
	if len(desc.Children) == 1 && desc.Children[0].Name == "*" {
		wildcardDesc = desc.Children[0].Desc
	}
	keyDescMap := desc.names

	// Create a new map if nil
	if destValue.IsNil() {
		destValue.Set(reflect.MakeMap(destValue.Type()))
	}

	elemType := destValue.Type().Elem()

	for key, value := range srcMap {
		if value == nil {
			continue
		}

		// Create a new element
		elemValue := reflect.New(elemType).Elem()

		// Find the appropriate descriptor
		if wildcardDesc != nil {
			if err := assignValueToField(wildcardDesc, value, elemValue, opt); err != nil {
				return err
			}
		} else if child, ok := keyDescMap[key]; ok && child.Desc != nil {
			if err := assignValueToField(child.Desc, value, elemValue, opt); err != nil {
				return err
			}
		} else {
			if err := assignScalar(value, elemValue); err != nil {
				return err
			}
		}

		destValue.SetMapIndex(reflect.ValueOf(key), elemValue)
	}

	return nil
}

// assignScalar assigns a scalar value to destValue
func assignScalar(src interface{}, destValue reflect.Value) error {
	if src == nil {
		return nil
	}

	srcValue := reflect.ValueOf(src)

	// Handle pointer destination
	if destValue.Kind() == reflect.Ptr {
		if destValue.IsNil() {
			destValue.Set(reflect.New(destValue.Type().Elem()))
		}
		destValue = destValue.Elem()
	}

	// Try direct assignment first
	if srcValue.Type().AssignableTo(destValue.Type()) {
		destValue.Set(srcValue)
		return nil
	}

	// Try conversion
	if srcValue.Type().ConvertibleTo(destValue.Type()) {
		destValue.Set(srcValue.Convert(destValue.Type()))
		return nil
	}

	// Handle special cases for numeric type conversions
	switch destValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v, ok := toInt64(src); ok {
			destValue.SetInt(v)
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v, ok := toUint64(src); ok {
			destValue.SetUint(v)
			return nil
		}
	case reflect.Float32, reflect.Float64:
		if v, ok := toFloat64(src); ok {
			destValue.SetFloat(v)
			return nil
		}
	case reflect.String:
		if v, ok := src.(string); ok {
			destValue.SetString(v)
			return nil
		}
	case reflect.Bool:
		if v, ok := src.(bool); ok {
			destValue.SetBool(v)
			return nil
		}
	}

	return fmt.Errorf("cannot assign %T to %v", src, destValue.Type())
}

// toInt64 converts various numeric types to int64
func toInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case uint:
		return int64(n), true
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		return int64(n), true
	case float32:
		return int64(n), true
	case float64:
		return int64(n), true
	default:
		return 0, false
	}
}

// toUint64 converts various numeric types to uint64
func toUint64(v interface{}) (uint64, bool) {
	switch n := v.(type) {
	case int:
		return uint64(n), true
	case int8:
		return uint64(n), true
	case int16:
		return uint64(n), true
	case int32:
		return uint64(n), true
	case int64:
		return uint64(n), true
	case uint:
		return uint64(n), true
	case uint8:
		return uint64(n), true
	case uint16:
		return uint64(n), true
	case uint32:
		return uint64(n), true
	case uint64:
		return n, true
	case float32:
		return uint64(n), true
	case float64:
		return uint64(n), true
	default:
		return 0, false
	}
}

// toFloat64 converts various numeric types to float64
func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}

// encodeUnknownField encodes a field value to protobuf binary format
func encodeUnknownField(bp *binary.BinaryProtocol, fieldID int, value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case bool:
		// varint type for bool
		bp.Buf = appendTag(bp.Buf, fieldID, 0) // varint wire type
		bp.WriteBool(v)

	case int:
		bp.Buf = appendTag(bp.Buf, fieldID, 0)
		bp.WriteInt64(int64(v))

	case int32:
		bp.Buf = appendTag(bp.Buf, fieldID, 0)
		bp.WriteInt32(v)

	case int64:
		bp.Buf = appendTag(bp.Buf, fieldID, 0)
		bp.WriteInt64(v)

	case uint32:
		bp.Buf = appendTag(bp.Buf, fieldID, 0)
		bp.WriteUint32(v)

	case uint64:
		bp.Buf = appendTag(bp.Buf, fieldID, 0)
		bp.WriteUint64(v)

	case float32:
		bp.Buf = appendTag(bp.Buf, fieldID, 5) // fixed32 wire type
		bp.WriteFloat(v)

	case float64:
		bp.Buf = appendTag(bp.Buf, fieldID, 1) // fixed64 wire type
		bp.WriteDouble(v)

	case string:
		bp.Buf = appendTag(bp.Buf, fieldID, 2) // length-delimited wire type
		bp.WriteString(v)

	case []byte:
		bp.Buf = appendTag(bp.Buf, fieldID, 2)
		bp.WriteBytes(v)

	case []interface{}:
		// Encode list as repeated field
		for _, elem := range v {
			if err := encodeUnknownField(bp, fieldID, elem); err != nil {
				return err
			}
		}

	case map[string]interface{}:
		// Encode as embedded message
		// First encode the message content
		subBp := binary.NewBinaryProtocolBuffer()
		defer binary.FreeBinaryProtocol(subBp)

		for key, val := range v {
			// For unknown map, we assume string keys with field ID based on some hash
			// This is a simplified approach - in practice, you'd need proper field descriptors
			if err := encodeUnknownField(subBp, hashFieldName(key), val); err != nil {
				return err
			}
		}

		bp.Buf = appendTag(bp.Buf, fieldID, 2) // length-delimited wire type
		bp.WriteBytes(subBp.Buf)

	default:
		// Try to use reflection for other types
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			bp.Buf = appendTag(bp.Buf, fieldID, 0)
			bp.WriteInt64(rv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			bp.Buf = appendTag(bp.Buf, fieldID, 0)
			bp.WriteUint64(rv.Uint())
		case reflect.Float32:
			bp.Buf = appendTag(bp.Buf, fieldID, 5)
			bp.WriteFloat(float32(rv.Float()))
		case reflect.Float64:
			bp.Buf = appendTag(bp.Buf, fieldID, 1)
			bp.WriteDouble(rv.Float())
		case reflect.String:
			bp.Buf = appendTag(bp.Buf, fieldID, 2)
			bp.WriteString(rv.String())
		case reflect.Slice:
			if rv.Type().Elem().Kind() == reflect.Uint8 {
				bp.Buf = appendTag(bp.Buf, fieldID, 2)
				bp.WriteBytes(rv.Bytes())
			} else {
				// Encode as repeated field
				for i := 0; i < rv.Len(); i++ {
					if err := encodeUnknownField(bp, fieldID, rv.Index(i).Interface()); err != nil {
						return err
					}
				}
			}
		default:
			return fmt.Errorf("unsupported type for unknown field encoding: %T", value)
		}
	}

	return nil
}

// appendTag appends a protobuf tag to the buffer
// wireType: 0=varint, 1=fixed64, 2=length-delimited, 5=fixed32
func appendTag(buf []byte, fieldNumber int, wireType int) []byte {
	tag := uint64(fieldNumber)<<3 | uint64(wireType&7)
	return protowire.AppendVarint(buf, tag)
}

// hashFieldName generates a simple field ID from a field name (for unknown maps)
func hashFieldName(name string) int {
	// Simple hash function - in practice you'd use a proper mapping
	h := 0
	for _, c := range name {
		h = h*31 + int(c)
	}
	if h < 0 {
		h = -h
	}
	// Keep within valid protobuf field number range
	return (h % 536870911) + 1 // Max valid field number is 2^29 - 1
}
