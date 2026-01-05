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

type Assigner struct {
	AssignOptions
}

// AssignAny assigns values from src (map[string]interface{}) to dest (protobuf struct) according to desc.
// For fields that exist in src but not in dest's struct definition (unknown fields):
//   - They will be encoded to XXX_unrecognized field using protobuf binary encoding
//   - Their raw values will also be stored in XXX_NoUnkeyedLiteral field (if present) as map[string]interface{}
//     with field names as keys
//
// Warning: desc must be normalized before calling this method.
func (a Assigner) AssignAny(desc *Descriptor, src interface{}, dest interface{}) error {
	if src == nil || dest == nil || desc == nil {
		return nil
	}

	// desc.Normalize()

	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer to struct")
	}

	// Initialize path stack from pool
	stack := getStackFrames()
	defer putStackFrames(stack)

	return assignValue(desc, src, destValue.Elem(), &a.AssignOptions, stack)
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
	// noUnkeyedLiteralIndex is the index of XXX_NoUnkeyedLiteral field, -1 if not present
	noUnkeyedLiteralIndex int
}

// pbFieldCache caches the struct field info for each type
var pbFieldCache sync.Map // map[reflect.Type]*pbStructFieldInfo

var (
	pbUnknownFieldName            = "XXX_unrecognized"
	pbUnknownFieldNameOnce        sync.Once
	NoUnkeyedLiteralFieldName     = "XXX_NoUnkeyedLiteral"
	NoUnkeyedLiteralFieldNameOnce sync.Once
)

// SetPBUnknownFieldName sets the name of the field used to store unknown fields in protobuf structs
func SetPBUnknownFieldName(name string) {
	pbUnknownFieldNameOnce.Do(func() {
		pbUnknownFieldName = name
	})
}

// SetPBNoUnkeyedLiteralFieldName sets the name of the field used to store unknown fields as raw values in protobuf structs
func SetPBNoUnkeyedLiteralFieldName(name string) {
	NoUnkeyedLiteralFieldNameOnce.Do(func() {
		NoUnkeyedLiteralFieldName = name
	})
}

// getPBStructFieldInfo returns cached struct field info for the given protobuf type
func getPBStructFieldInfo(t reflect.Type) *pbStructFieldInfo {
	if cached, ok := pbFieldCache.Load(t); ok {
		return cached.(*pbStructFieldInfo)
	}

	// Build the field info
	numField := t.NumField()
	info := &pbStructFieldInfo{
		nameToFieldIndex:      make(map[string]int, numField),
		nameToFieldID:         make(map[string]int, numField),
		idToFieldIndex:        make(map[int]int, numField),
		unrecognizedIndex:     -1,
		noUnkeyedLiteralIndex: -1,
	}

	for i := 0; i < numField; i++ {
		field := t.Field(i)

		// Check for XXX_unrecognized field
		if field.Name == pbUnknownFieldName {
			info.unrecognizedIndex = i
			continue
		}

		// Check for XXX_NoUnkeyedLiteral field
		if field.Name == NoUnkeyedLiteralFieldName {
			info.noUnkeyedLiteralIndex = i
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

// assignValue is the internal implementation that works with reflect.Value directly
func assignValue(desc *Descriptor, src interface{}, destValue reflect.Value, opt *AssignOptions, stack *pathStack) error {
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
		return assignStruct(desc, src, destValue, opt, stack)
	case TypeKind_StrMap:
		return assignStrMap(desc, src, destValue, opt, stack)
	case TypeKind_List:
		return assignList(desc, src, destValue, opt, stack)
	default:
		return assignLeaf(reflect.ValueOf(src), destValue)
	}
}

// assignStruct handles TypeKind_Struct assignment
func assignStruct(desc *Descriptor, src interface{}, destValue reflect.Value, opt *AssignOptions, stack *pathStack) error {
	srcMap, ok := src.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map[string]interface{} for struct at %s, got %T", stack.buildPath(), src)
	}

	if destValue.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct destination at %s, got %v", stack.buildPath(), destValue.Kind())
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
		fieldID := 0
		if hasDescField {
			fieldID = descField.ID
		}

		// Find struct field by name using cached index map
		fieldIdx, found := fieldInfo.nameToFieldIndex[key]
		if found {
			fieldValue := destValue.Field(fieldIdx)

			// Make sure field is settable
			if !fieldValue.CanSet() {
				continue
			}

			// Push field onto stack
			stack.push(TypeKind_Struct, key, fieldID)

			// If field has a child descriptor, recursively assign
			var err error
			if hasDescField && descField.Desc != nil {
				err = assignValueToField(descField.Desc, value, fieldValue, opt, stack)
			} else {
				// Otherwise, assign the value directly
				err = assignLeaf(reflect.ValueOf(value), fieldValue)
			}

			// Pop field from stack
			stack.pop()

			if err != nil {
				return err
			}
		} else if hasDescField {
			// Field exists in descriptor but not in struct - encode to XXX_unrecognized
			// This will be handled below
			unassignedFields[key] = value
		} else if opt.DisallowNotDefined {
			// Include path in error message
			stack.push(TypeKind_Struct, key, fieldID)
			path := stack.buildPath()
			stack.pop()
			return ErrNotFound{Parent: desc, Field: Field{Name: key}, Msg: fmt.Sprintf("field '%s' not found in struct at path %s", key, path)}
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
				field := descFieldMap[key]
				if err := encodeUnknownField(bp, field.ID, val, field.Desc); err != nil {
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

	// Store unassigned fields as raw values to XXX_NoUnkeyedLiteral
	if len(unassignedFields) > 0 && fieldInfo.noUnkeyedLiteralIndex >= 0 {
		noUnkeyedLiteralValue := destValue.Field(fieldInfo.noUnkeyedLiteralIndex)
		if noUnkeyedLiteralValue.CanSet() {
			// Check if it's a map[string]interface{} or compatible type
			if noUnkeyedLiteralValue.Kind() == reflect.Map {
				// Initialize map if nil
				if noUnkeyedLiteralValue.IsNil() {
					noUnkeyedLiteralValue.Set(reflect.MakeMap(noUnkeyedLiteralValue.Type()))
				}

				// Set each unassigned field to the map
				for key, value := range unassignedFields {
					noUnkeyedLiteralValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
				}
			}
		}
	}

	return nil
}

// assignValueToField assigns a value to a field, handling pointer allocation
func assignValueToField(desc *Descriptor, src interface{}, fieldValue reflect.Value, opt *AssignOptions, stack *pathStack) error {
	// Handle pointer fields - allocate if needed
	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			// Skip allocating for empty non-leaf sources to avoid clobbering existing data
			if isEmptyNonLeaf(desc, src) {
				return nil
			}

			fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
		}
		return assignValue(desc, src, fieldValue.Elem(), opt, stack)
	}
	return assignValue(desc, src, fieldValue, opt, stack)
}

// isEmptyNonLeaf returns true if src is an empty composite value for the given descriptor.
// Used to avoid allocating new nested structs or slices when there is nothing to assign.
func isEmptyNonLeaf(desc *Descriptor, src interface{}) bool {
	if desc == nil || src == nil {
		return false
	}

	switch desc.Kind {
	case TypeKind_Struct, TypeKind_StrMap:
		if m, ok := src.(map[string]interface{}); ok {
			return len(m) == 0
		}
	case TypeKind_List:
		if s, ok := src.([]interface{}); ok {
			return len(s) == 0
		}
	}

	return false
}

// assignStrMap handles TypeKind_StrMap assignment
func assignStrMap(desc *Descriptor, src interface{}, destValue reflect.Value, opt *AssignOptions, stack *pathStack) error {
	srcMap, ok := src.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map[string]interface{} for strmap at %s, got %T", stack.buildPath(), src)
	}

	if destValue.Kind() != reflect.Map {
		return fmt.Errorf("expected map destination at %s, got %v", stack.buildPath(), destValue.Kind())
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
		keyValue := reflect.ValueOf(key)
		existing := destValue.MapIndex(keyValue)
		elemValue := reflect.New(elemType).Elem()
		if existing.IsValid() {
			elemValue.Set(existing)
		}

		// Find the appropriate descriptor for this entry
		var childDesc *Descriptor
		if wildcardDesc != nil {
			childDesc = wildcardDesc
		} else if child, ok := keyDescMap[key]; ok {
			childDesc = child.Desc
		}

		// Skip assigning empty non-leaf nodes to avoid overwriting existing values
		if childDesc != nil && isEmptyNonLeaf(childDesc, value) {
			if !existing.IsValid() {
				continue
			}
			// keep existing value untouched
			continue
		}

		if value == nil {
			// Preserve existing entry if present; otherwise set zero value
			if existing.IsValid() {
				destValue.SetMapIndex(keyValue, existing)
			}
			continue
		}

		// Push map key onto stack
		stack.push(TypeKind_StrMap, key, 0)

		var err error
		if childDesc != nil {
			if elemType.Kind() == reflect.Ptr {
				ptrValue := elemValue
				if elemValue.IsNil() {
					ptrValue = reflect.New(elemType.Elem())
				}
				err = assignValue(childDesc, value, ptrValue.Elem(), opt, stack)
				if err == nil {
					elemValue = ptrValue
				}
			} else {
				err = assignValue(childDesc, value, elemValue, opt, stack)
			}
		} else {
			err = assignLeaf(reflect.ValueOf(value), elemValue)
		}

		// Pop map key from stack
		stack.pop()

		if err != nil {
			return err
		}

		destValue.SetMapIndex(keyValue, elemValue)
	}

	return nil
}

// assignList handles TypeKind_List assignment
func assignList(desc *Descriptor, src interface{}, destValue reflect.Value, opt *AssignOptions, stack *pathStack) error {
	srcSlice, ok := src.([]interface{})
	if !ok {
		return fmt.Errorf("expected []interface{} for list at %s, got %T", stack.buildPath(), src)
	}

	if destValue.Kind() != reflect.Slice && destValue.Kind() != reflect.Array {
		return fmt.Errorf("expected slice or array destination at %s, got %v", stack.buildPath(), destValue.Kind())
	}

	childrenLen := len(desc.Children)

	// Fast path: only wildcard descriptor ("*" means all elements)
	if childrenLen == 1 && desc.Children[0].Name == "*" {
		wildcardDesc := desc.Children[0].Desc
		srcLen := len(srcSlice)

		// Handle array (fixed size)
		if destValue.Kind() == reflect.Array {
			arrayLen := destValue.Len()
			if srcLen != arrayLen {
				return fmt.Errorf("array length mismatch at %s: expected %d, got %d", stack.buildPath(), arrayLen, srcLen)
			}
			for i := 0; i < srcLen; i++ {
				elemValue := destValue.Index(i)
				if srcSlice[i] == nil {
					continue
				}

				// Push list index onto stack
				stack.push(TypeKind_List, "*", i)

				var err error
				if wildcardDesc != nil {
					err = assignValueToField(wildcardDesc, srcSlice[i], elemValue, opt, stack)
				} else {
					err = assignLeaf(reflect.ValueOf(srcSlice[i]), elemValue)
				}

				// Pop list index from stack
				stack.pop()

				if err != nil {
					return err
				}
			}
			return nil
		}

		// Handle slice (dynamic size). Reuse existing slice when possible; only allocate when growing.
		elemType := destValue.Type().Elem()
		destLen := destValue.Len()
		useSlice := destValue
		if destLen < srcLen {
			useSlice = reflect.MakeSlice(destValue.Type(), srcLen, srcLen)
			if destLen > 0 {
				reflect.Copy(useSlice, destValue)
			}
		}

		for i := 0; i < srcLen; i++ {
			elemValue := useSlice.Index(i)

			// Skip empty non-leaf sources to avoid overwriting existing elements
			if wildcardDesc != nil && isEmptyNonLeaf(wildcardDesc, srcSlice[i]) {
				continue
			}

			if srcSlice[i] == nil {
				// Preserve existing value
				continue
			}

			// Push list index onto stack
			stack.push(TypeKind_List, "*", i)

			var err error
			if wildcardDesc != nil {
				// Handle pointer element type without dropping existing value
				if elemType.Kind() == reflect.Ptr {
					targetPtr := elemValue
					if elemValue.IsNil() {
						targetPtr = reflect.New(elemType.Elem())
					}
					err = assignValue(wildcardDesc, srcSlice[i], targetPtr.Elem(), opt, stack)
					if err == nil {
						elemValue.Set(targetPtr)
					}
				} else {
					err = assignValue(wildcardDesc, srcSlice[i], elemValue, opt, stack)
				}
			} else {
				err = assignLeaf(reflect.ValueOf(srcSlice[i]), elemValue)
			}

			// Pop list index from stack
			stack.pop()

			if err != nil {
				return err
			}
		}

		if useSlice.Pointer() != destValue.Pointer() || destLen != useSlice.Len() {
			destValue.Set(useSlice)
		}
		return nil
	}

	// Specific indices requested
	// This is a sparse assignment - we need to ensure the slice is large enough
	maxIdx := -1
	for _, child := range desc.Children {
		idx := child.ID
		if idx > maxIdx {
			maxIdx = idx
		}
	}

	if maxIdx < 0 {
		// No valid indices
		return nil
	}

	// Handle array (fixed size)
	if destValue.Kind() == reflect.Array {
		arrayLen := destValue.Len()
		if maxIdx >= arrayLen {
			return fmt.Errorf("index %d out of bounds for array of length %d at %s", maxIdx, arrayLen, stack.buildPath())
		}

		// Find corresponding source elements by index
		srcMap := make(map[int]interface{})
		for i, elem := range srcSlice {
			if i < len(desc.Children) {
				idx := desc.Children[i].ID
				srcMap[idx] = elem
			}
		}

		for _, child := range desc.Children {
			idx := child.ID
			if idx < 0 || idx >= arrayLen {
				if opt.DisallowNotDefined {
					stack.push(TypeKind_List, "", idx)
					path := stack.buildPath()
					stack.pop()
					return ErrNotFound{Parent: desc, Field: child, Msg: fmt.Sprintf("index %d out of bounds for array at path %s", idx, path)}
				}
				continue
			}

			srcVal, hasSrc := srcMap[idx]
			if !hasSrc {
				if opt.DisallowNotDefined {
					stack.push(TypeKind_List, "", idx)
					path := stack.buildPath()
					stack.pop()
					return ErrNotFound{Parent: desc, Field: child, Msg: fmt.Sprintf("index %d not found in source at path %s", idx, path)}
				}
				continue
			}

			elemValue := destValue.Index(idx)

			// Push list index onto stack
			stack.push(TypeKind_List, "", idx)

			var err2 error
			if child.Desc != nil {
				err2 = assignValueToField(child.Desc, srcVal, elemValue, opt, stack)
			} else {
				err2 = assignLeaf(reflect.ValueOf(srcVal), elemValue)
			}

			// Pop list index from stack
			stack.pop()

			if err2 != nil {
				return err2
			}
		}
		return nil
	}

	// Handle slice (dynamic size)
	// Create a slice large enough to hold all specified indices
	requiredLen := maxIdx + 1
	elemType := destValue.Type().Elem()
	newSlice := reflect.MakeSlice(destValue.Type(), requiredLen, requiredLen)

	// Copy existing elements if dest slice already has values
	if !destValue.IsNil() && destValue.Len() > 0 {
		reflect.Copy(newSlice, destValue)
	}

	// Find corresponding source elements by index
	srcMap := make(map[int]interface{})
	for i, elem := range srcSlice {
		if i < len(desc.Children) {
			idx := desc.Children[i].ID
			srcMap[idx] = elem
		}
	}

	for _, child := range desc.Children {
		idx := child.ID

		srcVal, hasSrc := srcMap[idx]
		if !hasSrc {
			if opt.DisallowNotDefined {
				stack.push(TypeKind_List, "", idx)
				path := stack.buildPath()
				stack.pop()
				return ErrNotFound{Parent: desc, Field: child, Msg: fmt.Sprintf("index %d not found in source at path %s", idx, path)}
			}
			continue
		}

		elemValue := newSlice.Index(idx)

		// Push list index onto stack
		stack.push(TypeKind_List, "", idx)

		var err2 error
		if child.Desc != nil {
			// Handle pointer element type
			if elemType.Kind() == reflect.Ptr {
				newElem := reflect.New(elemType.Elem())
				err2 = assignValue(child.Desc, srcVal, newElem.Elem(), opt, stack)
				if err2 == nil {
					elemValue.Set(newElem)
				}
			} else {
				err2 = assignValue(child.Desc, srcVal, elemValue, opt, stack)
			}
		} else {
			err2 = assignLeaf(reflect.ValueOf(srcVal), elemValue)
		}

		// Pop list index from stack
		stack.pop()

		if err2 != nil {
			return err2
		}
	}

	destValue.Set(newSlice)
	return nil
}

// jsonStructFieldInfo caches field mapping information for a struct type based on json tags
type jsonStructFieldInfo struct {
	// jsonNameToFieldIndex maps json tag name to struct field index
	jsonNameToFieldIndex map[string]int
	// noUnkeyedLiteralIndex is the index of XXX_NoUnkeyedLiteral field, -1 if not present
	noUnkeyedLiteralIndex int
}

// jsonFieldCache caches the struct field info for each type
var jsonFieldCache sync.Map // map[reflect.Type]*jsonStructFieldInfo

// getJSONStructFieldInfo returns cached struct field info for the given type based on json tags
func getJSONStructFieldInfo(t reflect.Type) *jsonStructFieldInfo {
	if cached, ok := jsonFieldCache.Load(t); ok {
		return cached.(*jsonStructFieldInfo)
	}

	// Build the field info
	numField := t.NumField()
	info := &jsonStructFieldInfo{
		jsonNameToFieldIndex: make(map[string]int, numField),
	}

	for i := 0; i < numField; i++ {
		field := t.Field(i)

		if field.Name == NoUnkeyedLiteralFieldName {
			info.noUnkeyedLiteralIndex = i
			continue
		}

		tag := field.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		// Parse json tag: "field_name" or "field_name,omitempty"
		jsonName := tag
		if idx := strings.IndexByte(tag, ','); idx >= 0 {
			jsonName = tag[:idx]
		}
		if jsonName == "" {
			continue
		}

		info.jsonNameToFieldIndex[jsonName] = i
	}

	// Store in cache (use LoadOrStore to handle concurrent initialization)
	actual, _ := jsonFieldCache.LoadOrStore(t, info)
	return actual.(*jsonStructFieldInfo)
}

// AssignValue assigns values from src to dest by matching reflect-type or map-key or json-tag
func (Assigner) AssignValue(src interface{}, dest interface{}) error {
	if src == nil || dest == nil {
		return nil
	}

	destValue := reflect.ValueOf(dest)

	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	return assignLeaf(reflect.ValueOf(src), destValue)
}

// assignStructToStruct assigns a struct to another struct by matching json tags
func assignStructToStruct(srcValue, destValue reflect.Value) error {
	srcType := srcValue.Type()
	destType := destValue.Type()

	// Get json field info for both types
	srcInfo := getJSONStructFieldInfo(srcType)
	destInfo := getJSONStructFieldInfo(destType)

	// Iterate through dest fields and find matching src fields by json tag
	for jsonName, destIdx := range destInfo.jsonNameToFieldIndex {
		srcIdx, found := srcInfo.jsonNameToFieldIndex[jsonName]
		if !found {
			continue
		}

		srcField := srcValue.Field(srcIdx)
		destField := destValue.Field(destIdx)

		if !destField.CanSet() {
			continue
		}

		// nil pointer in source struct: set dest to nil if pointer
		if srcField.Kind() == reflect.Ptr && srcField.IsNil() {
			if destField.Kind() == reflect.Ptr {
				destField.Set(reflect.Zero(destField.Type()))
			}
			continue
		}

		// Recursively assign the field value
		if err := assignLeaf(srcField, destField); err != nil {
			return err
		}
	}

	if srcInfo.noUnkeyedLiteralIndex >= 0 && destInfo.noUnkeyedLiteralIndex >= 0 {
		// Special handling for XXX_NoUnkeyedLiteral field
		// Copy the map if both src and dest have this field
		srcField := srcValue.Field(srcInfo.noUnkeyedLiteralIndex)
		destField := destValue.Field(destInfo.noUnkeyedLiteralIndex)
		if destField.CanSet() && srcField.Kind() == reflect.Map && destField.Kind() == reflect.Map {
			if !srcField.IsNil() && srcField.Len() > 0 {
				// Initialize dest map if nil
				if destField.IsNil() {
					destField.Set(reflect.MakeMap(destField.Type()))
				}

				// Copy all entries from src to dest
				iter := srcField.MapRange()
				for iter.Next() {
					destField.SetMapIndex(iter.Key(), iter.Value())
				}
			}
		}
	}

	return nil
}

// assignMapToStruct assigns a map[string]interface{} to a struct by matching json tags
func assignMapToStruct(srcMap map[string]interface{}, destValue reflect.Value) error {
	destType := destValue.Type()

	// Get json field info for dest type
	destInfo := getJSONStructFieldInfo(destType)

	// Iterate through source map and find matching dest fields by json tag
	for jsonName, srcVal := range srcMap {
		destIdx, found := destInfo.jsonNameToFieldIndex[jsonName]
		if !found {
			continue
		}

		destField := destValue.Field(destIdx)
		if !destField.CanSet() {
			continue
		}

		if srcVal == nil {
			// Set to zero value if source is nil
			if destField.Kind() == reflect.Ptr || destField.Kind() == reflect.Slice || destField.Kind() == reflect.Map {
				destField.Set(reflect.Zero(destField.Type()))
			}
			continue
		}

		// Recursively assign the field value
		if err := assignLeaf(reflect.ValueOf(srcVal), destField); err != nil {
			return err
		}
	}

	return nil
}

// assignLeaf assigns a reflect.Value to another reflect.Value
func assignLeaf(srcValue, destValue reflect.Value) error {
	// Dereference pointer source
	for srcValue.Kind() == reflect.Ptr {
		if srcValue.IsNil() {
			return nil
		}
		srcValue = srcValue.Elem()
	}

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

	// Handle interface{} source - extract the underlying value
	if srcValue.Kind() == reflect.Interface {
		if srcValue.IsNil() {
			return nil
		}
		// Get the underlying value and use assignScalar instead
		// because the underlying value could be map[string]interface{}
		return assignLeaf(srcValue.Elem(), destValue)
	}

	// Handle struct to struct mapping via json tags
	if srcValue.Kind() == reflect.Struct && destValue.Kind() == reflect.Struct {
		return assignStructToStruct(srcValue, destValue)
	}

	// Handle slice to slice mapping
	if srcValue.Kind() == reflect.Slice && destValue.Kind() == reflect.Slice {
		return assignSliceToSlice(srcValue, destValue)
	}

	// Handle map to map mapping
	if srcValue.Kind() == reflect.Map && destValue.Kind() == reflect.Map {
		return assignMapToMap(srcValue, destValue)
	}

	// Handle map[string]interface{} to struct mapping
	if srcValue.Kind() == reflect.Map && destValue.Kind() == reflect.Struct {
		srcMapIface, ok := srcValue.Interface().(map[string]interface{})
		if ok {
			return assignMapToStruct(srcMapIface, destValue)
		}
	}

	return fmt.Errorf("cannot assign %v to %v", srcValue.Type(), destValue.Type())
}

// assignSliceToSlice assigns a slice to another slice, converting elements as needed
func assignSliceToSlice(srcValue, destValue reflect.Value) error {
	if srcValue.IsNil() {
		destValue.Set(reflect.Zero(destValue.Type()))
		return nil
	}
	srcLen := srcValue.Len()
	destElemType := destValue.Type().Elem()

	// Create a new slice with the same length
	newSlice := reflect.MakeSlice(destValue.Type(), srcLen, srcLen)

	for i := 0; i < srcLen; i++ {
		srcElem := srcValue.Index(i)
		destElem := newSlice.Index(i)

		// Handle pointer element type
		if destElemType.Kind() == reflect.Ptr {
			// Check if source element is nil pointer
			if srcElem.Kind() == reflect.Ptr && srcElem.IsNil() {
				// Keep destination as nil (zero value for pointer)
				continue
			}
			// Check if source element is interface{} containing nil
			if srcElem.Kind() == reflect.Interface && srcElem.IsNil() {
				// Keep destination as nil (zero value for pointer)
				continue
			}
			newElem := reflect.New(destElemType.Elem())
			if err := assignLeaf(srcElem, newElem.Elem()); err != nil {
				return err
			}
			destElem.Set(newElem)
		} else {
			if err := assignLeaf(srcElem, destElem); err != nil {
				return err
			}
		}
	}

	destValue.Set(newSlice)
	return nil
}

// assignMapToMap assigns a map to another map, converting elements as needed
func assignMapToMap(srcValue, destValue reflect.Value) error {
	if srcValue.IsNil() {
		return nil
	}

	destType := destValue.Type()
	destKeyType := destType.Key()
	destElemType := destType.Elem()

	// Create a new map
	newMap := reflect.MakeMap(destType)

	iter := srcValue.MapRange()
	for iter.Next() {
		srcKey := iter.Key()
		srcVal := iter.Value()

		// Convert key
		var destKey reflect.Value
		if srcKey.Type().AssignableTo(destKeyType) {
			destKey = srcKey
		} else if srcKey.Type().ConvertibleTo(destKeyType) {
			destKey = srcKey.Convert(destKeyType)
		} else {
			return fmt.Errorf("cannot convert map key %v to %v", srcKey.Type(), destKeyType)
		}

		// Convert value
		destVal := reflect.New(destElemType).Elem()
		if destElemType.Kind() == reflect.Ptr {
			// Check if source value is nil pointer
			if srcVal.Kind() == reflect.Ptr && srcVal.IsNil() {
				// Keep destination as nil (zero value for pointer)
				newMap.SetMapIndex(destKey, destVal)
				continue
			}
			// Check if source value is interface{} containing nil
			if srcVal.Kind() == reflect.Interface && srcVal.IsNil() {
				// Keep destination as nil (zero value for pointer)
				newMap.SetMapIndex(destKey, destVal)
				continue
			}
			newElem := reflect.New(destElemType.Elem())
			if err := assignLeaf(srcVal, newElem.Elem()); err != nil {
				return err
			}
			destVal.Set(newElem)
		} else {
			if err := assignLeaf(srcVal, destVal); err != nil {
				return err
			}
		}

		newMap.SetMapIndex(destKey, destVal)
	}

	destValue.Set(newMap)
	return nil
}

// encodeUnknownField encodes a field value to protobuf binary format
// desc: field descriptor for this value, can be nil for basic types
func encodeUnknownField(bp *binary.BinaryProtocol, fieldID int, value interface{}, desc *Descriptor) error {
	if value == nil {
		return nil
	}

	// Handle nested structures with descriptor
	if desc != nil {
		switch desc.Kind {
		case TypeKind_Struct:
			// Encode as embedded message
			subBp := binary.NewBinaryProtocolBuffer()
			defer binary.FreeBinaryProtocol(subBp)

			if m, ok := value.(map[string]interface{}); ok {
				for key, val := range m {
					if childField, ok := desc.names[key]; ok {
						if err := encodeUnknownField(subBp, childField.ID, val, childField.Desc); err != nil {
							return err
						}
					}
				}
			}

			bp.Buf = appendTag(bp.Buf, fieldID, 2) // length-delimited wire type
			bp.WriteBytes(subBp.Buf)
			return nil

		case TypeKind_List:
			// Encode as repeated field
			if arr, ok := value.([]interface{}); ok {
				// Get the element descriptor (usually wildcard "*")
				var elemDesc *Descriptor
				if len(desc.Children) > 0 {
					if desc.Children[0].Name == "*" {
						elemDesc = desc.Children[0].Desc
					}
				}

				for _, elem := range arr {
					if err := encodeUnknownField(bp, fieldID, elem, elemDesc); err != nil {
						return err
					}
				}
			}
			return nil

		case TypeKind_StrMap:
			// For map types, we need to encode each entry as a nested message
			// with field 1 = key, field 2 = value
			if m, ok := value.(map[string]interface{}); ok {
				// Get the value descriptor (usually wildcard "*")
				var valueDesc *Descriptor
				if len(desc.Children) > 0 {
					if desc.Children[0].Name == "*" {
						valueDesc = desc.Children[0].Desc
					}
				}

				for key, val := range m {
					// Each map entry is encoded as a nested message
					subBp := binary.NewBinaryProtocolBuffer()
					defer binary.FreeBinaryProtocol(subBp)

					// Field 1: key (string)
					subBp.Buf = appendTag(subBp.Buf, 1, 2)
					subBp.WriteString(key)

					// Field 2: value
					if err := encodeUnknownField(subBp, 2, val, valueDesc); err != nil {
						return err
					}

					// Write the map entry
					bp.Buf = appendTag(bp.Buf, fieldID, 2)
					bp.WriteBytes(subBp.Buf)
				}
			}
			return nil
		}
	}

	// Fallback to type-based encoding for basic types or when no descriptor
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
		// Encode list as repeated field (without descriptor)
		for _, elem := range v {
			if err := encodeUnknownField(bp, fieldID, elem, nil); err != nil {
				return err
			}
		}

	case map[string]interface{}:
		// Encode as embedded message (without descriptor)
		subBp := binary.NewBinaryProtocolBuffer()
		defer binary.FreeBinaryProtocol(subBp)

		for key, val := range v {
			// For unknown map without descriptor, use hash-based field ID
			if err := encodeUnknownField(subBp, hashFieldName(key), val, nil); err != nil {
				return err
			}
		}

		bp.Buf = appendTag(bp.Buf, fieldID, 2) // length-delimited wire type
		bp.WriteBytes(subBp.Buf)

	default:
		return fmt.Errorf("unsupported type for unknown field encoding: %T", value)
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
