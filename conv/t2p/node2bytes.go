package t2p

import (
	"fmt"
	"unicode/utf8"

	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
	"github.com/cloudwego/dynamicgo/thrift"
	tgeneric "github.com/cloudwego/dynamicgo/thrift/generic"
)

type PathNodeToBytesConv struct{}

// Do converts a thrift DOM to a protobuf message.
func (c PathNodeToBytesConv) Do(in *tgeneric.PathNode, out *[]byte) error {
	if in == nil {
		return fmt.Errorf("input PathNode is nil")
	}

	// Use caller-provided buffer as the initial backing to avoid final copy
	initial := *out
	p := binary.NewBinaryProtol(initial[:0])
	defer binary.FreeBinaryProtocol(p)

	// Convert the thrift PathNode to protobuf binary format
	if err := c.convertPathNodeToProto(in, p, true); err != nil {
		return err
	}

	// Assign the resulting buffer directly to out
	*out = p.Buf
	return nil
}

// convertPathNodeToProto recursively converts thrift PathNode to protobuf binary format
// root indicates whether the current node is the root message (top-level)
func (c PathNodeToBytesConv) convertPathNodeToProto(node *tgeneric.PathNode, p *binary.BinaryProtocol, root bool) error {
	if node.IsError() {
		return fmt.Errorf("node error: %v", node.Node)
	}

	// Handle complex types based on the node type
	switch node.Node.Type() {
	case thrift.STRUCT:
		if len(node.Next) == 0 {
			// Pre-marshaled struct or empty struct
			if root {
				// top-level: write fields directly
				return c.convertThriftStructToProto(node, p)
			}
			// nested field: write length-delimited directly into parent buffer
			var pos int
			p.Buf, pos = binary.AppendSpeculativeLength(p.Buf)
			if err := c.convertThriftStructToProto(node, p); err != nil {
				return err
			}
			p.Buf = binary.FinishSpeculativeLength(p.Buf, pos)
			return nil
		}
		return c.writeStruct(node, p, root)
	case thrift.LIST, thrift.SET:
		if len(node.Next) == 0 {
			// Pre-marshaled list or empty list: write as unpacked elements with field tag
			return c.convertThriftListToProto(node, p)
		}
		return c.writeList(node, p)
	case thrift.MAP:
		if len(node.Next) == 0 {
			// Pre-marshaled map or empty map: write entries with field tag
			return c.convertThriftMapToProto(node, p)
		}
		return c.writeMap(node, p)
	default:
		// For primitive types, we need to convert thrift encoding to protobuf encoding
		return c.writePrimitiveValue(node, p)
	}
}

// writePrimitiveValue converts thrift primitive value to protobuf encoding
func (c PathNodeToBytesConv) writePrimitiveValue(node *tgeneric.PathNode, p *binary.BinaryProtocol) error {
	// Get the raw thrift data
	rawData := node.Node.Raw()
	if len(rawData) == 0 {
		return nil
	}

	switch node.Node.Type() {
	case thrift.BOOL:
		// Thrift BOOL is 1 byte, protobuf BOOL is varint
		if len(rawData) != 1 {
			return fmt.Errorf("invalid bool data length: %d", len(rawData))
		}
		value := rawData[0] != 0
		return p.WriteBool(value)

	case thrift.BYTE:
		// Thrift BYTE is 1 byte, can be mapped to protobuf INT32 or UINT32
		if len(rawData) != 1 {
			return fmt.Errorf("invalid byte data length: %d", len(rawData))
		}
		value := int32(rawData[0])
		return p.WriteInt32(value)

	case thrift.I16:
		// Thrift I16 is 2 bytes big-endian, protobuf INT32 is varint
		if len(rawData) != 2 {
			return fmt.Errorf("invalid i16 data length: %d", len(rawData))
		}
		value := int32(thrift.BinaryEncoding{}.DecodeInt16(rawData))
		return p.WriteInt32(value)

	case thrift.I32:
		// Thrift I32 is 4 bytes big-endian, protobuf INT32 is varint
		if len(rawData) != 4 {
			return fmt.Errorf("invalid i32 data length: %d", len(rawData))
		}
		value := thrift.BinaryEncoding{}.DecodeInt32(rawData)
		return p.WriteInt32(value)

	case thrift.I64:
		// Thrift I64 is 8 bytes big-endian, protobuf INT64 is varint
		if len(rawData) != 8 {
			return fmt.Errorf("invalid i64 data length: %d", len(rawData))
		}
		value := thrift.BinaryEncoding{}.DecodeInt64(rawData)
		return p.WriteInt64(value)

	case thrift.DOUBLE:
		// Thrift DOUBLE is 8 bytes big-endian, protobuf DOUBLE is fixed64
		if len(rawData) != 8 {
			return fmt.Errorf("invalid double data length: %d", len(rawData))
		}
		value := thrift.BinaryEncoding{}.DecodeDouble(rawData)
		return p.WriteDouble(value)

	case thrift.STRING:
		// Thrift STRING (including binary) is length-prefixed. For protobuf, treat it as bytes to avoid UTF-8 checks.
		if len(rawData) < 4 {
			return fmt.Errorf("invalid string/binary data length: %d", len(rawData))
		}
		// skip 4-byte length prefix and write the raw bytes
		return p.WriteBytes(rawData[4:])

	default:
		return fmt.Errorf("unsupported primitive thrift type: %v", node.Node.Type())
	}
}

// writeStruct converts thrift struct to protobuf message
func (c PathNodeToBytesConv) writeStruct(node *tgeneric.PathNode, p *binary.BinaryProtocol, root bool) error {
	// For nested messages, write length-delimited content
	pos := -1
	if !root {
		var okPos int
		p.Buf, okPos = binary.AppendSpeculativeLength(p.Buf)
		if okPos < 0 {
			return fmt.Errorf("append speculative length failed")
		}
		pos = okPos
	}

	// Write each field
	for i := range node.Next {
		child := &node.Next[i]
		if child.IsEmpty() {
			continue
		}

		// Skip empty complex types with zero raw data
		t := child.Node.Type()
		if (t == thrift.STRUCT || t == thrift.LIST || t == thrift.SET || t == thrift.MAP) && len(child.Node.Raw()) == 0 && len(child.Next) == 0 {
			continue
		}

		// When child is not LIST/SET/MAP, parent must append the field tag
		if t != thrift.LIST && t != thrift.SET && t != thrift.MAP {
			fieldID := proto.FieldNumber(child.Path.Id())
			wireType := c.getWireType(t)
			if err := p.AppendTag(fieldID, wireType); err != nil {
				return fmt.Errorf("failed to append tag for field %d: %v", fieldID, err)
			}
		}

		// Write the field value; nested structs are treated as length-delimited here
		if err := c.convertPathNodeToProto(child, p, false); err != nil {
			return fmt.Errorf("failed to convert field %d: %v", child.Path.Id(), err)
		}
	}

	if !root {
		p.Buf = binary.FinishSpeculativeLength(p.Buf, pos)
	}
	return nil
}

// writeList converts thrift list/set to protobuf repeated field
func (c PathNodeToBytesConv) writeList(node *tgeneric.PathNode, p *binary.BinaryProtocol) error {
    fieldID := proto.FieldNumber(node.Path.Id())
    if fieldID == 0 {
        fieldID = 1
    }
    et := node.Node.ElemType()
    pack := c.isPackableThriftType(et)
    if pack {
        if err := p.AppendTag(fieldID, proto.BytesType); err != nil {
            return err
        }
        var pos int
        p.Buf, pos = binary.AppendSpeculativeLength(p.Buf)
        for i := range node.Next {
            child := &node.Next[i]
            if child.IsEmpty() {
                continue
            }
            if err := c.convertPathNodeToProto(child, p, false); err != nil {
                return err
            }
        }
        p.Buf = binary.FinishSpeculativeLength(p.Buf, pos)
        return nil
    }
    for i := range node.Next {
        child := &node.Next[i]
        if child.IsEmpty() {
            continue
        }
        wireType := c.getWireType(child.Node.Type())
        if err := p.AppendTag(fieldID, wireType); err != nil {
            return err
        }
        if err := c.convertPathNodeToProto(child, p, false); err != nil {
            return err
        }
    }
    return nil
}

// writeMap converts thrift map to protobuf map field
func (c PathNodeToBytesConv) writeMap(node *tgeneric.PathNode, p *binary.BinaryProtocol) error {
	// For protobuf maps, each key-value pair is written as an entry sub-message
	fieldID := proto.FieldNumber(node.Path.Id())
	if fieldID == 0 {
		fieldID = 1 // fallback for map without explicit field number
	}
	for i := range node.Next {
		child := &node.Next[i]
		if child.IsEmpty() {
			continue
		}
		// Write the entry tag (length-delimited)
		if err := p.AppendTag(fieldID, proto.BytesType); err != nil {
			return fmt.Errorf("failed to append tag for map entry: %v", err)
		}

		// Inline entry content using speculative length
		var pos int
		p.Buf, pos = binary.AppendSpeculativeLength(p.Buf)
		// key: field 1
		keyWireType := c.getWireTypeFromPath(child.Path)
		if err := p.AppendTag(1, keyWireType); err != nil {
			return fmt.Errorf("failed to append key tag: %v", err)
		}
		if err := c.writePathValue(child.Path, p); err != nil {
			return fmt.Errorf("failed to write key value: %v", err)
		}
		// value: field 2
		valWireType := c.getWireType(child.Node.Type())
		if err := p.AppendTag(2, valWireType); err != nil {
			return fmt.Errorf("failed to append value tag: %v", err)
		}
		if err := c.convertPathNodeToProto(child, p, false); err != nil {
			return fmt.Errorf("failed to write value: %v", err)
		}
		// finish entry
		p.Buf = binary.FinishSpeculativeLength(p.Buf, pos)
	}
	return nil
}

// getWireType maps thrift type to protobuf wire type
func (c PathNodeToBytesConv) getWireType(thriftType thrift.Type) proto.WireType {
	switch thriftType {
	case thrift.BOOL:
		return proto.VarintType
	case thrift.BYTE, thrift.I16, thrift.I32, thrift.I64:
		return proto.VarintType
	case thrift.DOUBLE:
		return proto.Fixed64Type
	case thrift.STRING:
		return proto.BytesType
	case thrift.STRUCT:
		return proto.BytesType
	case thrift.LIST, thrift.SET:
		return proto.BytesType // Repeated fields are length-delimited
	case thrift.MAP:
		return proto.BytesType // Map fields are length-delimited
	default:
		return proto.BytesType // Default to bytes for unknown types
	}
}

// getWireTypeFromPath maps path type to protobuf wire type for map keys
func (c PathNodeToBytesConv) getWireTypeFromPath(path tgeneric.Path) proto.WireType {
	switch path.Type() {
	case tgeneric.PathStrKey:
		return proto.BytesType
	case tgeneric.PathIntKey:
		return proto.VarintType
	case tgeneric.PathBinKey:
		return proto.BytesType
	default:
		return proto.VarintType // Default to varint
	}
}

func (c PathNodeToBytesConv) isPackableThriftType(t thrift.Type) bool {
    switch t {
    case thrift.STRING, thrift.STRUCT, thrift.MAP, thrift.LIST, thrift.SET:
        return false
    default:
        return true
    }
}

// writePathValue writes the value from a Path (used for map keys)
func (c PathNodeToBytesConv) writePathValue(path tgeneric.Path, p *binary.BinaryProtocol) error {
	switch path.Type() {
	case tgeneric.PathStrKey:
		return p.WriteString(path.Str())
	case tgeneric.PathIntKey:
		return p.WriteInt64(int64(path.Int()))
	case tgeneric.PathBinKey:
		return p.WriteBytes(path.Bin())
	default:
		return fmt.Errorf("unsupported path type: %v", path.Type())
	}
}

// convertThriftStructToProto converts pre-marshaled thrift struct to protobuf
func (c PathNodeToBytesConv) convertThriftStructToProto(node *tgeneric.PathNode, p *binary.BinaryProtocol) error {
	// Parse the thrift struct data and convert to protobuf format
	rawData := node.Node.Raw()
	if len(rawData) == 0 {
		return nil // Empty struct
	}

	// Create a thrift protocol reader to parse the struct
	thriftProto := thrift.NewBinaryProtocol(rawData)
	defer thrift.FreeBinaryProtocolBuffer(thriftProto)

	// Read struct begin
	if _, err := thriftProto.ReadStructBegin(); err != nil {
		return fmt.Errorf("failed to read struct begin: %v", err)
	}

	// Read fields until STOP
	for {
		_, fieldType, fieldID, err := thriftProto.ReadFieldBegin()
		if err != nil {
			return fmt.Errorf("failed to read field begin: %v", err)
		}
		if fieldType == thrift.STOP {
			break
		}

		// Convert field to protobuf
		if err := c.convertThriftFieldToProto(int16(fieldID), fieldType, thriftProto, p); err != nil {
			return fmt.Errorf("failed to convert field %d: %v", fieldID, err)
		}

		if err := thriftProto.ReadFieldEnd(); err != nil {
			return fmt.Errorf("failed to read field end: %v", err)
		}
	}

	if err := thriftProto.ReadStructEnd(); err != nil {
		return fmt.Errorf("failed to read struct end: %v", err)
	}

	return nil
}

// convertThriftListToProto converts pre-marshaled thrift list to protobuf
func (c PathNodeToBytesConv) convertThriftListToProto(node *tgeneric.PathNode, p *binary.BinaryProtocol) error {
	rawData := node.Node.Raw()
	if len(rawData) == 0 {
		return nil // Empty list
	}

	// Create a thrift protocol reader to parse the list
	thriftProto := thrift.NewBinaryProtocol(rawData)
	defer thrift.FreeBinaryProtocolBuffer(thriftProto)

	// Read list begin
	elemType, size, err := thriftProto.ReadListBegin()
	if err != nil {
		return fmt.Errorf("failed to read list begin: %v", err)
	}

    fieldID := proto.FieldNumber(node.Path.Id())
    if fieldID == 0 {
        fieldID = 1
    }
    pack := c.isPackableThriftType(thrift.Type(elemType))
    if pack {
        if err := p.AppendTag(fieldID, proto.BytesType); err != nil {
            return err
        }
        var pos int
        p.Buf, pos = binary.AppendSpeculativeLength(p.Buf)
        for i := 0; i < size; i++ {
            if err := c.convertThriftValueToProto(elemType, thriftProto, p); err != nil {
                return fmt.Errorf("failed to convert list element %d: %v", i, err)
            }
        }
        p.Buf = binary.FinishSpeculativeLength(p.Buf, pos)
    } else {
        for i := 0; i < size; i++ {
            wireType := c.getThriftWireType(elemType)
            if err := p.AppendTag(fieldID, wireType); err != nil {
                return err
            }
            if thrift.Type(elemType) == thrift.STRING {
                s, err := thriftProto.ReadString(false)
                if err != nil {
                    return err
                }
                if err := p.WriteString(s); err != nil {
                    return err
                }
            } else {
                if err := c.convertThriftValueToProto(elemType, thriftProto, p); err != nil {
                    return fmt.Errorf("failed to convert list element %d: %v", i, err)
                }
            }
        }
    }

	if err := thriftProto.ReadListEnd(); err != nil {
		return fmt.Errorf("failed to read list end: %v", err)
	}

	return nil
}

// convertThriftMapToProto converts pre-marshaled thrift map to protobuf
func (c PathNodeToBytesConv) convertThriftMapToProto(node *tgeneric.PathNode, p *binary.BinaryProtocol) error {
	rawData := node.Node.Raw()
	if len(rawData) == 0 {
		return nil // Empty map
	}

	// Create a thrift protocol reader to parse the map
	thriftProto := thrift.NewBinaryProtocol(rawData)
	defer thrift.FreeBinaryProtocolBuffer(thriftProto)

	// Reuse entry buffer
	entryP := binary.NewBinaryProtocolBuffer()
	defer binary.FreeBinaryProtocol(entryP)

	// Read map begin
	keyType, valueType, size, err := thriftProto.ReadMapBegin()
	if err != nil {
		return fmt.Errorf("failed to read map begin: %v", err)
	}

	// Convert each key-value pair
	// When PathNode is built from NewNodeMap, Path is empty (zero), so we need a fallback field number.
	fieldID := proto.FieldNumber(node.Path.Id())
	if fieldID == 0 {
		fieldID = 1 // fallback for map without explicit field number
	}
	for i := 0; i < size; i++ {
		// Write entry tag (length-delimited)
		if err := p.AppendTag(fieldID, proto.BytesType); err != nil {
			return fmt.Errorf("failed to append tag for map entry: %v", err)
		}

		// Reset entry buffer and build entry content
		entryP.Reset()
		// Write key (field 1)
		keyWireType := c.getThriftWireType(keyType)
		if err := entryP.AppendTag(1, keyWireType); err != nil {
			return fmt.Errorf("failed to append key tag: %v", err)
		}
		if thrift.Type(keyType) == thrift.STRING {
			s, err := thriftProto.ReadString(false)
			if err != nil {
				return fmt.Errorf("failed to read map key: %v", err)
			}
			if err := entryP.WriteString(s); err != nil {
				return err
			}
		} else {
			if err := c.convertThriftValueToProto(thrift.Type(keyType), thriftProto, entryP); err != nil {
				return fmt.Errorf("failed to convert map key: %v", err)
			}
		}

		// Write value (field 2)
		valueWireType := c.getThriftWireType(valueType)
		if err := entryP.AppendTag(2, valueWireType); err != nil {
			return fmt.Errorf("failed to append value tag: %v", err)
		}
		if thrift.Type(valueType) == thrift.STRUCT {
			startPos := thriftProto.Read
			if err := thriftProto.SkipType(thrift.STRUCT); err != nil {
				return err
			}
			endPos := thriftProto.Read
			structData := thriftProto.Buf[startPos:endPos]
			tempNode := &tgeneric.PathNode{Node: tgeneric.NewNode(thrift.STRUCT, structData)}
			if err := c.convertThriftStructToProto(tempNode, entryP); err != nil {
				return err
			}
		} else if thrift.Type(valueType) == thrift.STRING {
			s, err := thriftProto.ReadString(false)
			if err != nil {
				return err
			}
			if err := entryP.WriteString(s); err != nil {
				return err
			}
		} else {
			if err := c.convertThriftValueToProto(thrift.Type(valueType), thriftProto, entryP); err != nil {
				return err
			}
		}

		// Write the complete entry as bytes into parent
		if err := p.WriteBytes(entryP.Buf); err != nil {
			return fmt.Errorf("failed to write map entry: %v", err)
		}
	}

	if err := thriftProto.ReadMapEnd(); err != nil {
		return fmt.Errorf("failed to read map end: %v", err)
	}

	return nil
}

// convertThriftFieldToProto converts a thrift field to protobuf format
func (c PathNodeToBytesConv) convertThriftFieldToProto(fieldID int16, fieldType thrift.Type, thriftProto *thrift.BinaryProtocol, p *binary.BinaryProtocol) error {
	protoFieldID := proto.FieldNumber(fieldID)
	switch fieldType {
	case thrift.LIST, thrift.SET:
		return c.convertThriftListToProtoWithField(protoFieldID, thriftProto, p)
	case thrift.MAP:
		return c.convertThriftMapToProtoWithField(protoFieldID, thriftProto, p)
	case thrift.STRUCT:
		// build nested struct in a temporary buffer (reused)
		startPos := thriftProto.Read
		if err := thriftProto.SkipType(thrift.STRUCT); err != nil {
			return err
		}
		endPos := thriftProto.Read
		structData := thriftProto.Buf[startPos:endPos]
		tempNode := &tgeneric.PathNode{Node: tgeneric.NewNode(thrift.STRUCT, structData)}
		entryP := binary.NewBinaryProtocolBuffer()
		defer binary.FreeBinaryProtocol(entryP)
		if err := c.convertThriftStructToProto(tempNode, entryP); err != nil {
			return err
		}
		if err := p.AppendTag(protoFieldID, proto.BytesType); err != nil {
			return err
		}
		return p.WriteBytes(entryP.Buf)
	default:
		// simple scalar: append tag then write value
		wireType := c.getThriftWireType(fieldType)
		if err := p.AppendTag(protoFieldID, wireType); err != nil {
			return fmt.Errorf("failed to append tag for field %d: %v", fieldID, err)
		}
		return c.convertThriftValueToProto(fieldType, thriftProto, p)
	}
}

// convertThriftListToProtoWithField writes a thrift list as repeated field with given field number
func (c PathNodeToBytesConv) convertThriftListToProtoWithField(fieldID proto.FieldNumber, thriftProto *thrift.BinaryProtocol, p *binary.BinaryProtocol) error {
	et, size, err := thriftProto.ReadListBegin()
	if err != nil {
		return fmt.Errorf("failed to read list begin: %v", err)
	}
	// reuse entryP for struct elements
	entryP := binary.NewBinaryProtocolBuffer()
	defer binary.FreeBinaryProtocol(entryP)
    pack := c.isPackableThriftType(thrift.Type(et))
    if pack {
        if err := p.AppendTag(fieldID, proto.BytesType); err != nil {
            return err
        }
        var pos int
        p.Buf, pos = binary.AppendSpeculativeLength(p.Buf)
        for i := 0; i < size; i++ {
            if err := c.convertThriftValueToProto(thrift.Type(et), thriftProto, p); err != nil {
                return err
            }
        }
        p.Buf = binary.FinishSpeculativeLength(p.Buf, pos)
    } else {
        for i := 0; i < size; i++ {
            wireType := c.getThriftWireType(et)
            if err := p.AppendTag(fieldID, wireType); err != nil {
                return err
            }
            if thrift.Type(et) == thrift.STRING {
                s, err := thriftProto.ReadString(false)
                if err != nil {
                    return err
                }
                if err := p.WriteString(s); err != nil {
                    return err
                }
            } else if thrift.Type(et) == thrift.STRUCT {
                startPos := thriftProto.Read
                if err := thriftProto.SkipType(thrift.STRUCT); err != nil {
                    return err
                }
                endPos := thriftProto.Read
                structData := thriftProto.Buf[startPos:endPos]
                entryP.Reset()
                tempNode := &tgeneric.PathNode{Node: tgeneric.NewNode(thrift.STRUCT, structData)}
                if err := c.convertThriftStructToProto(tempNode, entryP); err != nil {
                    return err
                }
                if err := p.WriteBytes(entryP.Buf); err != nil {
                    return err
                }
            } else {
                if err := c.convertThriftValueToProto(thrift.Type(et), thriftProto, p); err != nil {
                    return err
                }
            }
        }
    }
    if err := thriftProto.ReadListEnd(); err != nil {
        return err
    }
    return nil
}

// convertThriftMapToProtoWithField writes a thrift map as repeated map entry messages
func (c PathNodeToBytesConv) convertThriftMapToProtoWithField(fieldID proto.FieldNumber, thriftProto *thrift.BinaryProtocol, p *binary.BinaryProtocol) error {
	kt, vt, size, err := thriftProto.ReadMapBegin()
	if err != nil {
		return fmt.Errorf("failed to read map begin: %v", err)
	}
	// reuse entry and tmp buffers
	entryP := binary.NewBinaryProtocolBuffer()
	defer binary.FreeBinaryProtocol(entryP)
	tmp := binary.NewBinaryProtocolBuffer()
	defer binary.FreeBinaryProtocol(tmp)
	for i := 0; i < size; i++ {
		// key tag
		entryP.Reset()
		keyWireType := c.getThriftWireType(kt)
		if err := entryP.AppendTag(1, keyWireType); err != nil {
			return err
		}
		// write key
		if thrift.Type(kt) == thrift.STRING {
			s, err := thriftProto.ReadString(false)
			if err != nil {
				return err
			}
			if err := entryP.WriteString(s); err != nil {
				return err
			}
		} else {
			if err := c.convertThriftValueToProto(thrift.Type(kt), thriftProto, entryP); err != nil {
				return err
			}
		}
		// value tag
		valWireType := c.getThriftWireType(vt)
		if err := entryP.AppendTag(2, valWireType); err != nil {
			return err
		}
		// write value
		if thrift.Type(vt) == thrift.STRUCT {
			startPos := thriftProto.Read
			if err := thriftProto.SkipType(thrift.STRUCT); err != nil {
				return err
			}
			endPos := thriftProto.Read
			structData := thriftProto.Buf[startPos:endPos]
			tmp.Reset()
			tempNode := &tgeneric.PathNode{Node: tgeneric.NewNode(thrift.STRUCT, structData)}
			if err := c.convertThriftStructToProto(tempNode, tmp); err != nil {
				return err
			}
			if err := entryP.WriteBytes(tmp.Buf); err != nil {
				return err
			}
		} else if thrift.Type(vt) == thrift.STRING {
			s, err := thriftProto.ReadString(false)
			if err != nil {
				return err
			}
			if err := entryP.WriteString(s); err != nil {
				return err
			}
		} else {
			if err := c.convertThriftValueToProto(thrift.Type(vt), thriftProto, entryP); err != nil {
				return err
			}
		}
		// append entry to parent message
		if err := p.AppendTag(fieldID, proto.BytesType); err != nil {
			return err
		}
		if err := p.WriteBytes(entryP.Buf); err != nil {
			return err
		}
	}
	if err := thriftProto.ReadMapEnd(); err != nil {
		return err
	}
	return nil
}

// convertThriftValueToProto converts a thrift value to protobuf format
func (c PathNodeToBytesConv) convertThriftValueToProto(valueType thrift.Type, thriftProto *thrift.BinaryProtocol, p *binary.BinaryProtocol) error {
	switch valueType {
	case thrift.BOOL:
		value, err := thriftProto.ReadBool()
		if err != nil {
			return err
		}
		return p.WriteBool(value)

	case thrift.BYTE:
		value, err := thriftProto.ReadByte()
		if err != nil {
			return err
		}
		return p.WriteInt32(int32(value))

	case thrift.I16:
		value, err := thriftProto.ReadI16()
		if err != nil {
			return err
		}
		return p.WriteInt32(int32(value))

	case thrift.I32:
		value, err := thriftProto.ReadI32()
		if err != nil {
			return err
		}
		return p.WriteInt32(value)

	case thrift.I64:
		value, err := thriftProto.ReadI64()
		if err != nil {
			return err
		}
		return p.WriteInt64(value)

	case thrift.DOUBLE:
		value, err := thriftProto.ReadDouble()
		if err != nil {
			return err
		}
		return p.WriteDouble(value)

	case thrift.STRING:
		// Prefer writing as string if valid UTF-8; otherwise write raw bytes (for thrift binary)
		s, err := thriftProto.ReadString(false)
		if err != nil {
			return err
		}
		if utf8.ValidString(s) {
			return p.WriteString(s)
		}
		// invalid UTF-8 -> treat as bytes
		return p.WriteBytes([]byte(s))

	case thrift.STRUCT:
		// For nested structs, we need to read the entire struct data
		startPos := thriftProto.Read
		if err := thriftProto.SkipType(thrift.STRUCT); err != nil {
			return err
		}
		endPos := thriftProto.Read
		structData := thriftProto.Buf[startPos:endPos]

		// Create a temporary PathNode for the nested struct
		tempNode := &tgeneric.PathNode{
			Node: tgeneric.NewNode(thrift.STRUCT, structData),
		}
		return c.convertThriftStructToProto(tempNode, p)

	case thrift.LIST, thrift.SET:
		// For nested lists, we need to read the entire list data
		startPos := thriftProto.Read
		if err := thriftProto.SkipType(valueType); err != nil {
			return err
		}
		endPos := thriftProto.Read
		listData := thriftProto.Buf[startPos:endPos]

		// Create a temporary PathNode for the nested list
		tempNode := &tgeneric.PathNode{
			Node: tgeneric.NewNode(valueType, listData),
		}
		return c.convertThriftListToProto(tempNode, p)

	case thrift.MAP:
		// For nested maps, we need to read the entire map data
		startPos := thriftProto.Read
		if err := thriftProto.SkipType(thrift.MAP); err != nil {
			return err
		}
		endPos := thriftProto.Read
		mapData := thriftProto.Buf[startPos:endPos]

		// Create a temporary PathNode for the nested map
		tempNode := &tgeneric.PathNode{
			Node: tgeneric.NewNode(thrift.MAP, mapData),
		}
		return c.convertThriftMapToProto(tempNode, p)

	default:
		return fmt.Errorf("unsupported thrift value type: %v", valueType)
	}
}

// getThriftWireType maps thrift type to protobuf wire type
func (c PathNodeToBytesConv) getThriftWireType(thriftType thrift.Type) proto.WireType {
	return c.getWireType(thriftType) // Reuse the existing mapping
}
