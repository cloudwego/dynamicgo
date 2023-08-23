package generic

import (
	"fmt"

	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/protowire"
)

func (self Node) should(api string, t1 proto.Type) string {
	if self.t == proto.ERROR {
		return self.Error()
	}
	if self.t == t1 {
		return ""
	}
	return fmt.Sprintf("API `%s` only supports %+v type", api, t1)
}

func (self Node) should2(api string, t1 proto.Type, t2 proto.Type) string {
	if self.t == proto.ERROR {
		return self.Error()
	}
	if self.t == t1 || self.t == t2 {
		return ""
	}
	return fmt.Sprintf("API `%s` only supports %s or %s type", api, t1, t2)
}

// Len returns the element count of container-kind type (LIST/SET/MAP)
func (self Node) Len() (int, error) {
	if self.IsError() {
		return 0, self
	}
	return self.len()
}

// TODO: len 
func (self Node) len() (int, error) {
	switch self.t {
	case proto.LIST:
		return -1, errNode(meta.ErrUnsupportedType, "", nil)
		// b := rt.BytesFrom(unsafe.Pointer(uintptr(self.v)+uintptr(1)), 4, 4)
		// return int(thrift.BinaryEncoding{}.DecodeInt32(b)), nil
	case proto.MAP:
		return -1, errNode(meta.ErrUnsupportedType, "", nil)
		// b := rt.BytesFrom(unsafe.Pointer(uintptr(self.v)+uintptr(2)), 4, 4)
		// return int(thrift.BinaryEncoding{}.DecodeInt32(b)), nil
	default:
		return -1, errNode(meta.ErrUnsupportedType, "", nil)
	}
}

func (self Node) raw() []byte {
	return rt.BytesFrom(self.v, self.l, self.l)
}

// Return its underlying raw data
func (self Node) Raw() []byte {
	if self.Error() != "" {
		return nil
	}
	return self.raw()
}

// Byte returns the byte value contained by a I8/BYTE node
func (self Node) Byte() (byte, error) {
	if self.IsError() {
		return 0, self
	}
	return self.byte()
}

func (self Node) byte() (byte, error) {
	switch self.t {
	case proto.BYTE:
		return byte(protowire.BinaryDecoder{}.DecodeByte(rt.BytesFrom(self.v, int(self.l), int(self.l)))), nil
	default:
		return 0, errNode(meta.ErrUnsupportedType, "", nil)
	}
}

// Bool returns the bool value contained by a BOOL node
func (self Node) Bool() (bool, error) {
	if self.IsError() {
		return false, self
	}
	return self.bool()
}

func (self Node) bool() (bool, error) {
	switch self.t {
	case proto.BOOL:
		v, _ := protowire.BinaryDecoder{}.DecodeBool(rt.BytesFrom(self.v, int(self.l), int(self.l)))
		return v, nil
	default:
		return false, errNode(meta.ErrUnsupportedType, "", nil)
	}
}

// Int returns the int value contaned by a I8/I16/I32/I64 node
func (self Node) Uint() (uint, error) {
	if self.IsError() {
		return 0, self
	}
	return self.uint()
}

func (self Node) uint() (uint, error) {
	buf := rt.BytesFrom(self.v, int(self.l), int(self.l))
	switch self.t {
	case proto.UINT32:
		v, _ := protowire.BinaryDecoder{}.DecodeUint32(buf)
		return uint(v), nil
	case proto.FIX32:
		v, _ := protowire.BinaryDecoder{}.DecodeFixed32(buf)
		return uint(v), nil
	case proto.UINT64:
		v, _ := protowire.BinaryDecoder{}.DecodeUint64(buf)
		return uint(v), nil
	case proto.FIX64:
		v, _ := protowire.BinaryDecoder{}.DecodeFixed64(buf)
		return uint(v), nil
	default:
		return 0, errNode(meta.ErrUnsupportedType, "", nil)
	}
}

// Int returns the int value contaned by a I8/I16/I32/I64 node
func (self Node) Int() (int, error) {
	if self.IsError() {
		return 0, self
	}
	return self.int()
}

func (self Node) int() (int, error) {
	buf := rt.BytesFrom(self.v, int(self.l), int(self.l))
	switch self.t {
	case proto.INT32:
		return int(protowire.BinaryDecoder{}.DecodeByte(buf)), nil
	case proto.SINT32:
		v, _ := protowire.BinaryDecoder{}.DecodeSint32(buf)
		return int(v), nil
	case proto.SFIX32:
		v, _ := protowire.BinaryDecoder{}.DecodeSfixed32(buf)
		return int(v), nil
	case proto.INT64:
		v, _ := protowire.BinaryDecoder{}.DecodeInt64(buf)
		return int(v), nil
	case proto.SINT64:
		v, _ := protowire.BinaryDecoder{}.DecodeSint64(buf)
		return int(v), nil
	case proto.SFIX64:
		v, _ := protowire.BinaryDecoder{}.DecodeSfixed64(buf)
		return int(v), nil
	default:
		return 0, errNode(meta.ErrUnsupportedType, "", nil)
	}
}

// Float64 returns the float64 value contained by a DOUBLE node
func (self Node) Float64() (float64, error) {
	if self.IsError() {
		return 0, self
	}
	return self.float64()
}

func (self Node) float64() (float64, error) {
	switch self.t {
	case proto.DOUBLE:
		v, _ := protowire.BinaryDecoder{}.DecodeDouble(rt.BytesFrom(self.v, int(self.l), int(self.l)))
		return v, nil
	default:
		return 0, errNode(meta.ErrUnsupportedType, "", nil)
	}
}

// String returns the string value contianed by a STRING node
func (self Node) String() (string, error) {
	if self.IsError() {
		return "", self
	}
	return self.string()
}

func (self Node) string() (string, error) {
	switch self.t {
	case proto.STRING:
		str, _ := protowire.BinaryDecoder{}.DecodeString(rt.BytesFrom(self.v, int(self.l), int(self.l)))
		// if self.d.IsBinary() {
		// 	if !utf8.Valid(rt.Str2Mem(str)) {
		// 		return "", errNode(meta.ErrInvalidParam, "invalid utf8 string", nil)
		// 	}
		// }
		return str, nil
	default:
		return "", errNode(meta.ErrUnsupportedType, "", nil)
	}
}

// Binary returns the bytes value contained by a BINARY node
func (self Node) Binary() ([]byte, error) {
	if self.IsError() {
		return nil, self
	}
	return self.binary()
}
// BYTE?
func (self Node) binary() ([]byte, error) {
	switch self.t {
	case proto.STRING:
		v, _ := protowire.BinaryDecoder{}.DecodeBytes(rt.BytesFrom(self.v, int(self.l), int(self.l)))
		return v, nil
	default:
		return nil, errNode(meta.ErrUnsupportedType, "", nil)
	}
}

// List returns interface elements contained by a LIST/SET node
func (self Node) List(opts *Options) ([]interface{}, error) {
	return nil,nil
}

// StrMap returns the string keys and interface elements contained by a MAP<STRING,XX> node
func (self Node) StrMap(opts *Options) (map[string]interface{}, error) {
	return nil,nil
}

// StrMap returns the integer keys and interface elements contained by a MAP<I8|I16|I32|I64,XX> node
func (self Node) IntMap(opts *Options) (map[int]interface{}, error) {
	return nil,nil
}

// InterfaceMap returns the interface keys and interface elements contained by a MAP node.
// If the key type is complex (LIST/SET/MAP/STRUCT),
// it will be stored using its pointer since its value are not supported by Go
func (self Node) InterfaceMap(opts *Options) (map[interface{}]interface{}, error) {
	return nil,nil
}

func (self Node) Struct(opts *Options) (map[interface{}]interface{}, error) {
	return nil,nil
}

// Interface returns the go interface value contained by a node.
// If the node is a STRUCT, it will return a map[thrift.FieldID]interface{}
// If it is a map, it will return map[int|string|interface{}]interface{}, which depends on the key type
func (self Node) Interface(opts *Options) (interface{}, error) {
	switch self.t {
	case proto.ERROR:
		return nil, self
	case proto.BOOL:
		return self.bool()
	case proto.INT32, proto.SINT32, proto.SFIX32, proto.INT64, proto.SINT64, proto.SFIX64:
		return self.int()
	case proto.UINT32, proto.UINT64, proto.FIX32, proto.FIX64:
		return self.uint()
	case proto.DOUBLE:
		return self.float64()
	case proto.STRING:
		if opts.CastStringAsBinary {
			return self.binary()
		}
		return self.string()
	case proto.LIST:
		return self.List(opts)
	case proto.MAP:
		if kt := self.kt; kt == proto.STRING {
			return self.StrMap(opts)
		} else if kt.IsInt() {
			return self.IntMap(opts)
		} else {
			return self.InterfaceMap(opts)
		}
	case proto.MESSAGE:
		return self.Struct(opts)
	default:
		return 0, errNode(meta.ErrUnsupportedType, "", nil)
	}
}
