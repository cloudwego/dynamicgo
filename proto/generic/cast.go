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

// Len returns the element count of container-kind type (LIST/MAP)
func (self Node) Len() (int, error) {
	if self.IsError() {
		return 0, self
	}
	return self.len()
}

func (self Node) len() (int, error) {
	switch self.t {
	case proto.LIST, proto.MAP:
		return self.size, nil
	default:
		return -1, errNode(meta.ErrUnsupportedType, "Node.len: len() only used in LIST/MAP", nil)
	}
}

// Return its underlying raw data
func (self Node) Raw() []byte {
	if self.Error() != "" {
		return nil
	}
	return self.raw()
}

func (self Node) raw() []byte {
	return rt.BytesFrom(self.v, self.l, self.l)
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
		return false, errNode(meta.ErrUnsupportedType, "Node.bool: the Node type is not BOOL", nil)
	}
}

// Uint returns the uint value contained by a UINT32/UINT64/FIX32/FIX64 node
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
		return 0, errNode(meta.ErrUnsupportedType, "Node.uint: the Node type is not UINT32/UINT64/FIX32/FIX64", nil)
	}
}

// Int returns the int value contained by a INT32/SINT32/SFIX32/INT64/SINT64/SFIX64 node
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
		v, _ := protowire.BinaryDecoder{}.DecodeInt32(buf)
		return int(v), nil
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
		return 0, errNode(meta.ErrUnsupportedType, "Node.int: the Node type is not INT32/SINT32/SFIX32/INT64/SINT64/SFIX64", nil)
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
		return 0, errNode(meta.ErrUnsupportedType, "Node.float64: the Node type is not DOUBLE", nil)
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
		str, _, _ := protowire.BinaryDecoder{}.DecodeString(rt.BytesFrom(self.v, int(self.l), int(self.l)))
		return str, nil
	default:
		return "", errNode(meta.ErrUnsupportedType, "Node.float64: the Node type is not STRING", nil)
	}
}

// Binary returns the bytes value contained by a BYTE node
func (self Node) Binary() ([]byte, error) {
	if self.IsError() {
		return nil, self
	}
	return self.binary()
}

func (self Node) binary() ([]byte, error) {
	switch self.t {
	case proto.BYTE:
		v, _, _ := protowire.BinaryDecoder{}.DecodeBytes(rt.BytesFrom(self.v, int(self.l), int(self.l)))
		return v, nil
	default:
		return nil, errNode(meta.ErrUnsupportedType, "Node.binary: the Node type is not BYTE", nil)
	}
}

func (self Value) List(opts *Options) ([]interface{}, error) {
	if self.IsError() {
		return nil, self
	}
	return self.list(opts)
}

// List returns interface elements contained by a LIST node
func (self Value) list(opts *Options) ([]interface{}, error) {
	if self.IsError() {
		return nil, self
	}

	if err := self.should("List", proto.LIST); err != "" {
		return nil, errValue(meta.ErrUnsupportedType, err, nil)
	}

	it := self.iterElems()
	if it.Err != nil {
		return nil, it.Err
	}
	ret := make([]interface{}, 0, it.Size())
	isPacked := (*self.Desc).IsPacked()

	// read packed list tag and bytelen
	if isPacked {
		if _, _, _, err := it.p.ConsumeTag(); err != nil {
			return nil, errValue(meta.ErrRead, "", err)
		}
		if _, err := it.p.ReadLength(); err != nil {
			return nil, errValue(meta.ErrRead, "", err)
		}
	}

	for it.HasNext() {
		s, e := it.Next(UseNativeSkipForGet)
		if it.Err != nil {
			return nil, it.Err
		}

		v := wrapValue(self.slice(s, e, self.et), self.Desc)
		vv, err := v.Interface(opts)
		if err != nil {
			return ret, err
		}
		ret = append(ret, vv)
	}

	if it.Err != nil {
		return nil, errValue(meta.ErrRead, "", it.Err)
	}
	return ret, nil
}

func (self Value) IntMap(opts *Options) (map[int]interface{}, error) {
	if self.IsError() {
		return nil, self
	}
	return self.intMap(opts)
}

// StrMap returns the integer keys and interface elements contained by a MAP<Int/Uint,XX> node
func (self Value) intMap(opts *Options) (map[int]interface{}, error) {
	if self.IsError() {
		return nil, self
	}
	if err := self.should("IntMap", proto.MAP); err != "" {
		return nil, errNode(meta.ErrUnsupportedType, err, nil)
	}
	if !self.kt.IsInt() {
		return nil, errNode(meta.ErrUnsupportedType, "key must be INT type", nil)
	}
	it := self.iterPairs()
	if it.Err != nil {
		return nil, it.Err
	}
	ret := make(map[int]interface{}, it.size)
	valueDesc := (*self.Desc).MapValue()
	for it.HasNext() {
		_, ks, s, e := it.NextInt(opts.UseNativeSkip)
		if it.Err != nil {
			return nil, it.Err
		}
		v := self.sliceWithDesc(s, e, &valueDesc)
		vv, err := v.Interface(opts)
		if err != nil {
			return ret, err
		}
		ret[ks] = vv
	}
	return ret, it.Err
}

func (self Value) StrMap(opts *Options) (map[string]interface{}, error) {
	if self.IsError() {
		return nil, self
	}
	return self.strMap(opts)
}

// StrMap returns the string keys and interface elements contained by a MAP<STRING,XX> node
func (self Value) strMap(opts *Options) (map[string]interface{}, error) {
	if self.IsError() {
		return nil, self
	}
	if err := self.should("StrMap", proto.MAP); err != "" {
		return nil, errNode(meta.ErrUnsupportedType, err, nil)
	}
	it := self.iterPairs()
	if it.Err != nil {
		return nil, it.Err
	}
	if self.kt != proto.STRING {
		return nil, errNode(meta.ErrUnsupportedType, "key type must be STRING", nil)
	}
	ret := make(map[string]interface{}, it.size)
	valueDesc := (*self.Desc).MapValue()
	for it.HasNext() {
		_, ks, s, e := it.NextStr(opts.UseNativeSkip)
		if it.Err != nil {
			return nil, it.Err
		}
		v := self.sliceWithDesc(s, e, &valueDesc)
		vv, err := v.Interface(opts)
		if err != nil {
			return ret, err
		}
		ret[ks] = vv
	}
	return ret, it.Err
}

// Interface returns the go interface value contained by a node.
// If the node is a MESSAGE, it will return map[proto.FieldNumber]interface{} or map[int]interface{}.
// If it is a map, it will return map[int|string]interface{}, which depends on the key type
func (self Value) Interface(opts *Options) (interface{}, error) {
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
	case proto.BYTE:
		return self.binary()
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
			return 0, errValue(meta.ErrUnsupportedType, "Value.Interface: not support other Mapkey type", nil)
		}
	case proto.MESSAGE:
		it := self.iterFields()
		var fds proto.FieldDescriptors
		if self.RootDesc != nil {
			fds = (*self.RootDesc).Fields()
		} else {
			fds = (*self.Desc).Message().Fields()
			if _, err := it.p.ReadLength(); err != nil {
				return nil, errValue(meta.ErrRead, "", err)
			}
		}

		if it.Err != nil {
			return nil, it.Err
		}

		var ret1 map[proto.FieldNumber]interface{}
		var ret2 map[int]interface{}
		if opts.MapStructById {
			ret1 = make(map[proto.FieldNumber]interface{}, DefaultNodeSliceCap)
		} else {
			ret2 = make(map[int]interface{}, DefaultNodeSliceCap)
		}

		for it.HasNext() {
			id, _, s, e, tagPos := it.Next(UseNativeSkipForGet)
			f := fds.ByNumber(id)

			if f == nil {
				return nil, errValue(meta.ErrUnknownField, fmt.Sprintf("unknown field id %d", id), nil)
			}

			if id == f.Number() {
				if f.IsMap() || f.IsList() {
					it.p.Read = tagPos
					if _, err := it.p.SkipAllElements(id, f.IsPacked()); err != nil {
						return nil, errValue(meta.ErrRead, "SkipAllElements in LIST/MAP failed", err)
					}
					s = tagPos
					e = it.p.Read
				}
			}

			v := self.sliceWithDesc(s, e, &f)
			vv, err := v.Interface(opts)
			if err != nil {
				return nil, err
			}

			if opts.MapStructById {
				ret1[id] = vv
			} else {
				ret2[int(id)] = vv
			}
		}

		if it.Err != nil {
			return nil, errValue(meta.ErrRead, "", it.Err)
		}

		if opts.MapStructById {
			return ret1, nil
		}
		return ret2, nil
	default:
		return 0, errNode(meta.ErrUnsupportedType, "", nil)
	}
}
