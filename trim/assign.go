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

// errorCollector collects errors during assignment in try-best mode
type errorCollector struct {
	errors []error
}

func (ec *errorCollector) add(err error) {
	if err != nil {
		ec.errors = append(ec.errors, err)
	}
}

func (ec *errorCollector) toError() error {
	if len(ec.errors) == 0 {
		return nil
	}
	if len(ec.errors) == 1 {
		return ec.errors[0]
	}
	return MultiErrors{Errors: ec.errors}
}

// AssignAny assigns values from src (map[string]interface{}) to dest (protobuf struct) according to desc.
// For fields that exist in src but not in dest's struct definition (unknown fields):
//   - If the field exists in descriptor: it will be encoded to XXX_unrecognized field using protobuf binary encoding
//   - All unknown fields (with or without descriptor) will be stored in XXX_NoUnkeyedLiteral field (if present)
//     as map[string]interface{} with field names as keys
//
// Warning: desc must be normalized before calling this method.
// This method uses try-best mode: it will continue processing even if some fields fail,
// collecting all errors and returning them at the end.
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

	// Initialize error collector
	ec := &errorCollector{}

	assignValue(desc, src, destValue.Elem(), &a.AssignOptions, stack, ec)
	return ec.toError()
}

// unassignedFieldEntry represents a field that is not assigned to a struct field
type unassignedFieldEntry struct {
	value   interface{}
	hasDesc bool // true if field exists in descriptor
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
func assignValue(desc *Descriptor, src interface{}, destValue reflect.Value, opt *AssignOptions, stack *pathStack, ec *errorCollector) {
	if src == nil {
		return
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
		assignStruct(desc, src, destValue, opt, stack, ec)
	case TypeKind_StrMap:
		assignStrMap(desc, src, destValue, opt, stack, ec)
	case TypeKind_List:
		assignList(desc, src, destValue, opt, stack, ec)
	default:
		ec.add(assignLeaf(reflect.ValueOf(src), destValue))
	}
}

// assignStruct handles TypeKind_Struct assignment
func assignStruct(desc *Descriptor, src interface{}, destValue reflect.Value, opt *AssignOptions, stack *pathStack, ec *errorCollector) {
	srcMap, ok := src.(map[string]interface{})
	if !ok {
		ec.add(fmt.Errorf("expected map[string]interface{} for struct at %s, got %T", stack.buildPath(), src))
		return
	}

	if destValue.Kind() != reflect.Struct {
		ec.add(fmt.Errorf("expected struct destination at %s, got %v", stack.buildPath(), destValue.Kind()))
		return
	}

	// Get cached field info for this type
	fieldInfo := getPBStructFieldInfo(destValue.Type())

	// Track which fields in srcMap are not assigned to struct fields
	unassignedFields := make(map[string]unassignedFieldEntry, len(srcMap))

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
			if hasDescField && descField.Desc != nil {
				assignValueToField(descField.Desc, value, fieldValue, opt, stack, ec)
			} else {
				// Otherwise, assign the value directly
				ec.add(assignLeaf(reflect.ValueOf(value), fieldValue))
			}

			// Pop field from stack
			stack.pop()
		} else if hasDescField {
			// Field exists in descriptor but not in struct
			unassignedFields[key] = unassignedFieldEntry{value: value, hasDesc: true}
		} else {
			// Field does not exist in descriptor
			if opt.DisallowNotDefined {
				// Include path in error message
				stack.push(TypeKind_Struct, key, fieldID)
				path := stack.buildPath()
				stack.pop()
				ec.add(ErrNotFound{Parent: desc, Field: Field{Name: key}, Msg: fmt.Sprintf("field '%s' not found in struct at path %s", key, path)})
			} else {
				// Store field without descriptor
				unassignedFields[key] = unassignedFieldEntry{value: value, hasDesc: false}
			}
		}
	}

	// Encode unassigned fields with descriptor to XXX_unrecognized
	if len(unassignedFields) > 0 && fieldInfo.unrecognizedIndex >= 0 {
		unrecognizedValue := destValue.Field(fieldInfo.unrecognizedIndex)
		if unrecognizedValue.CanSet() {
			bp := binary.NewBinaryProtocolBuffer()
			defer binary.FreeBinaryProtocol(bp)

			for key, entry := range unassignedFields {
				// Only encode fields with descriptor to XXX_unrecognized
				if !entry.hasDesc {
					continue
				}
				// Encode this field to XXX_unrecognized
				field := descFieldMap[key]
				if err := encodeUnknownField(bp, field.ID, entry.value, field.Desc); err != nil {
					ec.add(fmt.Errorf("failed to encode unknown field '%s': %w", key, err))
					continue
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

	// Store all unassigned fields to XXX_NoUnkeyedLiteral (with or without descriptor)
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
				for key, entry := range unassignedFields {
					noUnkeyedLiteralValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(entry.value))
				}
			}
		}
	}
}

// assignValueToField assigns a value to a field, handling pointer allocation
func assignValueToField(desc *Descriptor, src interface{}, fieldValue reflect.Value, opt *AssignOptions, stack *pathStack, ec *errorCollector) {
	// If source is an empty non-leaf, only initialize nil destinations and skip assignment.
	if isEmptyNonLeaf(desc, src) {
		initEmptyNonLeafValue(desc, fieldValue)
		return
	}

	// Handle pointer fields - allocate if needed
	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
		}
		assignValue(desc, src, fieldValue.Elem(), opt, stack, ec)
		return
	}
	assignValue(desc, src, fieldValue, opt, stack, ec)
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

// initEmptyNonLeafValue initializes nil destinations for empty non-leaf sources.
// It avoids overwriting existing non-nil values.
func initEmptyNonLeafValue(desc *Descriptor, destValue reflect.Value) {
	if desc == nil || !destValue.IsValid() {
		return
	}

	// Unwrap pointers but only allocate when nil
	if destValue.Kind() == reflect.Ptr {
		if destValue.IsNil() {
			destValue.Set(reflect.New(destValue.Type().Elem()))
		}
		destValue = destValue.Elem()
	}

	switch desc.Kind {
	case TypeKind_StrMap:
		if destValue.Kind() == reflect.Map && destValue.IsNil() && destValue.CanSet() {
			destValue.Set(reflect.MakeMap(destValue.Type()))
		}
	case TypeKind_List:
		if destValue.Kind() == reflect.Slice && destValue.IsNil() && destValue.CanSet() {
			destValue.Set(reflect.MakeSlice(destValue.Type(), 0, 0))
		}
	case TypeKind_Struct:
		// Struct zero value is fine once pointer is allocated.
		return
	}
}

// assignStrMap handles TypeKind_StrMap assignment
func assignStrMap(desc *Descriptor, src interface{}, destValue reflect.Value, opt *AssignOptions, stack *pathStack, ec *errorCollector) {
	srcMap, ok := src.(map[string]interface{})
	if !ok {
		ec.add(fmt.Errorf("expected map[string]interface{} for strmap at %s, got %T", stack.buildPath(), src))
		return
	}

	if destValue.Kind() != reflect.Map {
		ec.add(fmt.Errorf("expected map destination at %s, got %v", stack.buildPath(), destValue.Kind()))
		return
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
			if existing.IsValid() {
				// keep existing value untouched
				continue
			}
			// initialize zero value when destination is nil
			if elemType.Kind() == reflect.Ptr {
				newElem := reflect.New(elemType.Elem())
				initEmptyNonLeafValue(childDesc, newElem)
				destValue.SetMapIndex(keyValue, newElem)
			} else {
				zeroElem := reflect.New(elemType).Elem()
				initEmptyNonLeafValue(childDesc, zeroElem)
				destValue.SetMapIndex(keyValue, zeroElem)
			}
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

		if childDesc != nil {
			if elemType.Kind() == reflect.Ptr {
				ptrValue := elemValue
				if elemValue.IsNil() {
					ptrValue = reflect.New(elemType.Elem())
				}
				assignValue(childDesc, value, ptrValue.Elem(), opt, stack, ec)
				elemValue = ptrValue
			} else {
				assignValue(childDesc, value, elemValue, opt, stack, ec)
			}
		} else {
			ec.add(assignLeaf(reflect.ValueOf(value), elemValue))
		}

		// Pop map key from stack
		stack.pop()

		destValue.SetMapIndex(keyValue, elemValue)
	}
}

// assignList handles TypeKind_List assignment
func assignList(desc *Descriptor, src interface{}, destValue reflect.Value, opt *AssignOptions, stack *pathStack, ec *errorCollector) {
	srcSlice, ok := src.([]interface{})
	if !ok {
		ec.add(fmt.Errorf("expected []interface{} for list at %s, got %T", stack.buildPath(), src))
		return
	}

	if destValue.Kind() != reflect.Slice && destValue.Kind() != reflect.Array {
		ec.add(fmt.Errorf("expected slice or array destination at %s, got %v", stack.buildPath(), destValue.Kind()))
		return
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
				ec.add(fmt.Errorf("array length mismatch at %s: expected %d, got %d", stack.buildPath(), arrayLen, srcLen))
				return
			}
			for i := 0; i < srcLen; i++ {
				elemValue := destValue.Index(i)
				if srcSlice[i] == nil {
					continue
				}

				// Push list index onto stack
				stack.push(TypeKind_List, "*", i)

				if wildcardDesc != nil {
					assignValueToField(wildcardDesc, srcSlice[i], elemValue, opt, stack, ec)
				} else {
					ec.add(assignLeaf(reflect.ValueOf(srcSlice[i]), elemValue))
				}

				// Pop list index from stack
				stack.pop()
			}
			return
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
				if elemType.Kind() == reflect.Ptr {
					if elemValue.IsNil() {
						newElem := reflect.New(elemType.Elem())
						initEmptyNonLeafValue(wildcardDesc, newElem)
						elemValue.Set(newElem)
					}
				} else {
					initEmptyNonLeafValue(wildcardDesc, elemValue)
				}
				continue
			}

			if srcSlice[i] == nil {
				// Preserve existing value
				continue
			}

			// Push list index onto stack
			stack.push(TypeKind_List, "*", i)

			if wildcardDesc != nil {
				// Handle pointer element type without dropping existing value
				if elemType.Kind() == reflect.Ptr {
					targetPtr := elemValue
					if elemValue.IsNil() {
						targetPtr = reflect.New(elemType.Elem())
					}
					assignValue(wildcardDesc, srcSlice[i], targetPtr.Elem(), opt, stack, ec)
					elemValue.Set(targetPtr)
				} else {
					assignValue(wildcardDesc, srcSlice[i], elemValue, opt, stack, ec)
				}
			} else {
				ec.add(assignLeaf(reflect.ValueOf(srcSlice[i]), elemValue))
			}

			// Pop list index from stack
			stack.pop()
		}

		if useSlice.Pointer() != destValue.Pointer() || destLen != useSlice.Len() {
			destValue.Set(useSlice)
		}
		return
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
		return
	}

	// Handle array (fixed size)
	if destValue.Kind() == reflect.Array {
		arrayLen := destValue.Len()
		if maxIdx >= arrayLen {
			ec.add(fmt.Errorf("index %d out of bounds for array of length %d at %s", maxIdx, arrayLen, stack.buildPath()))
			return
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
					ec.add(ErrNotFound{Parent: desc, Field: child, Msg: fmt.Sprintf("index %d out of bounds for array at path %s", idx, path)})
				}
				continue
			}

			srcVal, hasSrc := srcMap[idx]
			if !hasSrc {
				if opt.DisallowNotDefined {
					stack.push(TypeKind_List, "", idx)
					path := stack.buildPath()
					stack.pop()
					ec.add(ErrNotFound{Parent: desc, Field: child, Msg: fmt.Sprintf("index %d not found in source at path %s", idx, path)})
				}
				continue
			}

			elemValue := destValue.Index(idx)

			// Push list index onto stack
			stack.push(TypeKind_List, "", idx)

			if child.Desc != nil {
				assignValueToField(child.Desc, srcVal, elemValue, opt, stack, ec)
			} else {
				ec.add(assignLeaf(reflect.ValueOf(srcVal), elemValue))
			}

			// Pop list index from stack
			stack.pop()
		}
		return
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
				ec.add(ErrNotFound{Parent: desc, Field: child, Msg: fmt.Sprintf("index %d not found in source at path %s", idx, path)})
			}
			continue
		}

		elemValue := newSlice.Index(idx)

		// Push list index onto stack
		stack.push(TypeKind_List, "", idx)

		if child.Desc != nil {
			// Handle pointer element type
			if elemType.Kind() == reflect.Ptr {
				newElem := reflect.New(elemType.Elem())
				assignValue(child.Desc, srcVal, newElem.Elem(), opt, stack, ec)
				elemValue.Set(newElem)
			} else {
				assignValue(child.Desc, srcVal, elemValue, opt, stack, ec)
			}
		} else {
			ec.add(assignLeaf(reflect.ValueOf(srcVal), elemValue))
		}

		// Pop list index from stack
		stack.pop()
	}

	destValue.Set(newSlice)
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
// This method uses try-best mode: it will continue processing even if some fields fail,
// collecting all errors and returning them at the end.
func (Assigner) AssignValue(src interface{}, dest interface{}) error {
	if src == nil || dest == nil {
		return nil
	}

	destValue := reflect.ValueOf(dest)

	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	// Initialize error collector
	ec := &errorCollector{}

	assignLeafTryBest(reflect.ValueOf(src), destValue, ec)
	return ec.toError()
}

// assignStructToStruct assigns a struct to another struct by matching json tags
func assignStructToStruct(srcValue, destValue reflect.Value) error {
	ec := &errorCollector{}
	assignStructToStructTryBest(srcValue, destValue, ec)
	return ec.toError()
}

// assignStructToStructTryBest assigns a struct to another struct by matching json tags in try-best mode
func assignStructToStructTryBest(srcValue, destValue reflect.Value, ec *errorCollector) {
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
		assignLeafTryBest(srcField, destField, ec)
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
}

// assignMapToStruct assigns a map[string]interface{} to a struct by matching json tags
func assignMapToStruct(srcMap map[string]interface{}, destValue reflect.Value) error {
	ec := &errorCollector{}
	assignMapToStructTryBest(srcMap, destValue, ec)
	return ec.toError()
}

// assignMapToStructTryBest assigns a map[string]interface{} to a struct by matching json tags in try-best mode
func assignMapToStructTryBest(srcMap map[string]interface{}, destValue reflect.Value, ec *errorCollector) {
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
		assignLeafTryBest(reflect.ValueOf(srcVal), destField, ec)
	}
}

// assignLeafTryBest assigns a reflect.Value to another reflect.Value in try-best mode
func assignLeafTryBest(srcValue, destValue reflect.Value, ec *errorCollector) {
	// Dereference pointer source
	for srcValue.Kind() == reflect.Ptr {
		if srcValue.IsNil() {
			return
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
		return
	}

	// Try conversion
	if srcValue.Type().ConvertibleTo(destValue.Type()) {
		destValue.Set(srcValue.Convert(destValue.Type()))
		return
	}

	// Handle interface{} source - extract the underlying value
	if srcValue.Kind() == reflect.Interface {
		if srcValue.IsNil() {
			return
		}
		// Get the underlying value and use assignScalar instead
		// because the underlying value could be map[string]interface{}
		assignLeafTryBest(srcValue.Elem(), destValue, ec)
		return
	}

	// Handle struct to struct mapping via json tags
	if srcValue.Kind() == reflect.Struct && destValue.Kind() == reflect.Struct {
		assignStructToStructTryBest(srcValue, destValue, ec)
		return
	}

	// Handle slice to slice mapping
	if srcValue.Kind() == reflect.Slice && destValue.Kind() == reflect.Slice {
		assignSliceToSliceTryBest(srcValue, destValue, ec)
		return
	}

	// Handle map to map mapping
	if srcValue.Kind() == reflect.Map && destValue.Kind() == reflect.Map {
		assignMapToMapTryBest(srcValue, destValue, ec)
		return
	}

	// Handle map[string]interface{} to struct mapping
	if srcValue.Kind() == reflect.Map && destValue.Kind() == reflect.Struct {
		srcMapIface, ok := srcValue.Interface().(map[string]interface{})
		if ok {
			assignMapToStructTryBest(srcMapIface, destValue, ec)
			return
		}
	}

	ec.add(fmt.Errorf("cannot assign %v to %v", srcValue.Type(), destValue.Type()))
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
	ec := &errorCollector{}
	assignSliceToSliceTryBest(srcValue, destValue, ec)
	return ec.toError()
}

// assignSliceToSliceTryBest assigns a slice to another slice, converting elements as needed in try-best mode
func assignSliceToSliceTryBest(srcValue, destValue reflect.Value, ec *errorCollector) {
	if srcValue.IsNil() {
		destValue.Set(reflect.Zero(destValue.Type()))
		return
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
			assignLeafTryBest(srcElem, newElem.Elem(), ec)
			destElem.Set(newElem)
		} else {
			assignLeafTryBest(srcElem, destElem, ec)
		}
	}

	destValue.Set(newSlice)
}

// assignMapToMap assigns a map to another map, converting elements as needed
func assignMapToMap(srcValue, destValue reflect.Value) error {
	ec := &errorCollector{}
	assignMapToMapTryBest(srcValue, destValue, ec)
	return ec.toError()
}

// assignMapToMapTryBest assigns a map to another map, converting elements as needed in try-best mode
func assignMapToMapTryBest(srcValue, destValue reflect.Value, ec *errorCollector) {
	if srcValue.IsNil() {
		return
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
			ec.add(fmt.Errorf("cannot convert map key %v to %v", srcKey.Type(), destKeyType))
			continue
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
			assignLeafTryBest(srcVal, newElem.Elem(), ec)
			destVal.Set(newElem)
		} else {
			assignLeafTryBest(srcVal, destVal, ec)
		}

		newMap.SetMapIndex(destKey, destVal)
	}

	destValue.Set(newMap)
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
				// Handle map[string]interface{}
				for key, val := range m {
					if childField, ok := desc.names[key]; ok {
						if err := encodeUnknownField(subBp, childField.ID, val, childField.Desc); err != nil {
							return err
						}
					}
				}
			} else {
				// Handle struct type via reflection
				rv := reflect.ValueOf(value)
				for rv.Kind() == reflect.Ptr {
					rv = rv.Elem()
				}

				if rv.Kind() == reflect.Struct {
					// Use cached field info to avoid repeated json tag parsing
					structType := rv.Type()
					fieldInfo := getJSONStructFieldInfo(structType)

					// Iterate through cached json field mappings
					for jsonName, fieldIdx := range fieldInfo.jsonNameToFieldIndex {
						fieldValue := rv.Field(fieldIdx)

						// Find corresponding field in descriptor
						if childField, ok := desc.names[jsonName]; ok {
							if err := encodeUnknownField(subBp, childField.ID, fieldValue.Interface(), childField.Desc); err != nil {
								return err
							}
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

				// Try packed encoding for primitive numeric types
				if len(arr) > 0 && canUsePacked(arr[0]) {
					// Use packed encoding: tag + length + data
					packedBuf := binary.NewBinaryProtocolBuffer()
					defer binary.FreeBinaryProtocol(packedBuf)

					for _, elem := range arr {
						if err := encodePackedElement(packedBuf, elem); err != nil {
							// Fallback to non-packed if any element fails
							goto nonPacked
						}
					}

					bp.Buf = appendTag(bp.Buf, fieldID, 2) // length-delimited wire type
					bp.WriteBytes(packedBuf.Buf)
					return nil
				}

			nonPacked:
				// Non-packed encoding: each element with its own tag
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
		// Handle typedef types (e.g., type Int int32) using reflection
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Bool:
			bp.Buf = appendTag(bp.Buf, fieldID, 0)
			bp.WriteBool(rv.Bool())

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			bp.Buf = appendTag(bp.Buf, fieldID, 0)
			bp.WriteInt64(rv.Int())

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			bp.Buf = appendTag(bp.Buf, fieldID, 0)
			bp.WriteUint64(rv.Uint())

		case reflect.Float32:
			bp.Buf = appendTag(bp.Buf, fieldID, 5) // fixed32 wire type
			bp.WriteFloat(float32(rv.Float()))

		case reflect.Float64:
			bp.Buf = appendTag(bp.Buf, fieldID, 1) // fixed64 wire type
			bp.WriteDouble(rv.Float())

		case reflect.String:
			bp.Buf = appendTag(bp.Buf, fieldID, 2)
			bp.WriteString(rv.String())

		case reflect.Slice:
			// Handle []byte typedef
			if rv.Type().Elem().Kind() == reflect.Uint8 {
				bp.Buf = appendTag(bp.Buf, fieldID, 2)
				bp.WriteBytes(rv.Bytes())
			} else {
				return fmt.Errorf("unsupported slice type for unknown field encoding: %T", value)
			}

		default:
			return fmt.Errorf("unsupported type for unknown field encoding: %T (kind: %v)", value, rv.Kind())
		}
	}

	return nil
}

// canUsePacked returns true if the value type can be encoded using packed encoding
func canUsePacked(value interface{}) bool {
	if value == nil {
		return false
	}

	switch value.(type) {
	case bool, int, int32, int64, uint32, uint64, float32, float64:
		return true
	default:
		// Check typedef using reflection
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return true
		}
	}
	return false
}

// encodePackedElement encodes a single element for packed encoding (without tag)
func encodePackedElement(bp *binary.BinaryProtocol, value interface{}) error {
	switch v := value.(type) {
	case bool:
		bp.WriteBool(v)
	case int:
		bp.WriteInt64(int64(v))
	case int32:
		bp.WriteInt32(v)
	case int64:
		bp.WriteInt64(v)
	case uint32:
		bp.WriteUint32(v)
	case uint64:
		bp.WriteUint64(v)
	case float32:
		bp.WriteFloat(v)
	case float64:
		bp.WriteDouble(v)
	default:
		// Handle typedef types using reflection
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Bool:
			bp.WriteBool(rv.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			bp.WriteInt64(rv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			bp.WriteUint64(rv.Uint())
		case reflect.Float32:
			bp.WriteFloat(float32(rv.Float()))
		case reflect.Float64:
			bp.WriteDouble(rv.Float())
		default:
			return fmt.Errorf("unsupported type for packed encoding: %T", value)
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
