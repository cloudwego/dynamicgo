package p2j

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/json"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/base"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	_GUARD_SLICE_FACTOR = 2
)

func wrapError(code meta.ErrCode, msg string, err error) error {
	return meta.NewError(code, msg, err)
}

//go:noinline
func unwrapError(msg string, err error) error {
	if v, ok := err.(meta.Error); ok {
		return wrapError(v.Code, msg, err)
	} else {
		return wrapError(meta.ErrConvert, msg, err)
	}
}

func (self *ProtoConv) do(ctx context.Context, src []byte, desc *proto.MessageDescriptor, out *[]byte, resp http.ResponseSetter) (err error) {
	rt.GuardSlice(out, len(src)*_GUARD_SLICE_FACTOR)
	var p = base.BinaryProtocol{
		Buf: src,
	}

	fields := (*desc).Fields()
	comma := false
	existExceptionField := false

	*out = json.EncodeObjectBegin(*out)

	for p.Read <= len(src) {
		fieldId, typeId, _, e := p.ConsumeTag()
		if e != nil {
			return wrapError(meta.ErrRead, "", e)
		}

		fd := fields.ByNumber(protowire.Number(fieldId))
		if fd == nil {
			return wrapError(meta.ErrRead, "invalid field", nil)
		}

		if comma {
			*out = json.EncodeObjectComma(*out)
		} else {
			comma = true
		}

		// serizalize jsonname
		*out = json.EncodeString(*out, fd.JSONName())
		*out = json.EncodeObjectColon(*out)

		if self.opts.EnableValueMapping {

		} else {
			err := self.doRecurse(ctx, &fd, out, resp, &p, typeId)
			if err != nil {
				return unwrapError(fmt.Sprintf("converting field %s of MESSAGE %s failed", fd.Name(), fd.Kind()), err)
			}
		}

		if existExceptionField {
			break
		}
	}

	// if err = self.handleUnsets(r, desc.Struct(), out, comma, ctx, resp); err != nil {
	// 	return err
	// }

	// thrift.FreeRequiresBitmap(r)
	if existExceptionField && err == nil {
		err = errors.New(string(*out))
	} else {
		*out = json.EncodeObjectEnd(*out)
	}
	return err
}

// parse MessageField recursive
func (self *ProtoConv) doRecurse(ctx context.Context, fd *proto.FieldDescriptor, out *[]byte, resp http.ResponseSetter, p *base.BinaryProtocol, typeId proto.WireType) error {
	switch {
	case (*fd).IsList():
		return self.unmarshalList(ctx, resp, p, typeId, out, fd)
	case (*fd).IsMap():
		return nil
	default:
		return self.unmarshalSingular(ctx, resp, p, typeId, out, fd)
	}
	// tt := desc.Type()
	// switch tt {
	// case thrift.BOOL:
	// 	v, e := p.ReadBool()
	// 	if e != nil {
	// 		return wrapError(meta.ErrRead, "", e)
	// 	}
	// 	*out = json.EncodeBool(*out, v)
	// case thrift.BYTE:
	// 	v, e := p.ReadByte()
	// 	if e != nil {
	// 		return wrapError(meta.ErrWrite, "", e)
	// 	}
	// 	if self.opts.ByteAsUint8 {
	// 		*out = json.EncodeInt64(*out, int64(uint8(v)))
	// 	} else {
	// 		*out = json.EncodeInt64(*out, int64(int8(v)))
	// 	}
	// case thrift.I16:
	// 	v, e := p.ReadI16()
	// 	if e != nil {
	// 		return wrapError(meta.ErrWrite, "", e)
	// 	}
	// 	*out = json.EncodeInt64(*out, int64(v))
	// case thrift.I32:
	// 	v, e := p.ReadI32()
	// 	if e != nil {
	// 		return wrapError(meta.ErrWrite, "", e)
	// 	}
	// 	*out = json.EncodeInt64(*out, int64(v))
	// case thrift.I64:
	// 	v, e := p.ReadI64()
	// 	if e != nil {
	// 		return wrapError(meta.ErrWrite, "", e)
	// 	}
	// 	if self.opts.String2Int64 {
	// 		*out = append(*out, '"')
	// 		*out = json.EncodeInt64(*out, int64(v))
	// 		*out = append(*out, '"')
	// 	} else {
	// 		*out = json.EncodeInt64(*out, int64(v))
	// 	}
	// case thrift.DOUBLE:
	// 	v, e := p.ReadDouble()
	// 	if e != nil {
	// 		return wrapError(meta.ErrWrite, "", e)
	// 	}
	// 	*out = json.EncodeFloat64(*out, float64(v))
	// case thrift.STRING:
	// 	if desc.IsBinary() && !self.opts.NoBase64Binary {
	// 		v, e := p.ReadBinary(false)
	// 		if e != nil {
	// 			return wrapError(meta.ErrRead, "", e)
	// 		}
	// 		*out = json.EncodeBaniry(*out, v)
	// 	} else {
	// 		v, e := p.ReadString(false)
	// 		if e != nil {
	// 			return wrapError(meta.ErrRead, "", e)
	// 		}
	// 		*out = json.EncodeString(*out, v)
	// 	}
	// case thrift.STRUCT:
	// 	_, e := p.ReadStructBegin()
	// 	if e != nil {
	// 		return wrapError(meta.ErrRead, "", e)
	// 	}
	// 	*out = json.EncodeObjectBegin(*out)

	// 	r := thrift.NewRequiresBitmap()
	// 	desc.Struct().Requires().CopyTo(r)
	// 	comma := false

	// 	for {
	// 		_, typeId, id, e := p.ReadFieldBegin()
	// 		if e != nil {
	// 			return wrapError(meta.ErrRead, "", e)
	// 		}
	// 		if typeId == 0 {
	// 			break
	// 		}

	// 		field := desc.Struct().FieldById(thrift.FieldID(id))
	// 		if field == nil {
	// 			if self.opts.DisallowUnknownField {
	// 				return wrapError(meta.ErrUnknownField, fmt.Sprintf("unknown field %d", id), nil)
	// 			}
	// 			if e := p.Skip(typeId, self.opts.UseNativeSkip); e != nil {
	// 				return wrapError(meta.ErrRead, "", e)
	// 			}
	// 			continue
	// 		}

	// 		r.Set(field.ID(), thrift.OptionalRequireness)
	// 		restart := p.Read

	// 		if resp != nil && self.opts.EnableHttpMapping && field.HTTPMappings() != nil {
	// 			ok, err := self.writeHttpValue(ctx, resp, p, field)
	// 			if err != nil {
	// 				return unwrapError(fmt.Sprintf("mapping field %s of STRUCT %s failed", field.Name(), desc.Name()), err)
	// 			}
	// 			// NOTICE: if option HttpMappingAsExtra is false and http-mapping failed,
	// 			// continue to write to json body
	// 			if !self.opts.WriteHttpValueFallback || ok {
	// 				continue
	// 			}
	// 		}

	// 		// HttpMappingAsExtra is true, return to begining and write json body
	// 		p.Read = restart
	// 		if comma {
	// 			*out = json.EncodeObjectComma(*out)
	// 			if err != nil {
	// 				return wrapError(meta.ErrWrite, "", err)
	// 			}
	// 		} else {
	// 			comma = true
	// 		}
	// 		// NOTICE: always use field.Alias() here, because alias equals to name by default
	// 		*out = json.EncodeString(*out, field.Alias())
	// 		*out = json.EncodeObjectColon(*out)

	// 		if self.opts.EnableValueMapping && field.ValueMapping() != nil {
	// 			if err = field.ValueMapping().Read(ctx, p, field, out); err != nil {
	// 				return unwrapError(fmt.Sprintf("mapping field %s of STRUCT %s failed", field.Name(), desc.Type()), err)
	// 			}
	// 		} else {
	// 			err = self.doRecurse(ctx, field.Type(), out, nil, p)
	// 			if err != nil {
	// 				return unwrapError(fmt.Sprintf("converting field %s of STRUCT %s failed", field.Name(), desc.Type()), err)
	// 			}
	// 		}
	// 	}

	// 	if err = self.handleUnsets(r, desc.Struct(), out, comma, ctx, resp); err != nil {
	// 		return err
	// 	}

	// 	thrift.FreeRequiresBitmap(r)
	// 	*out = json.EncodeObjectEnd(*out)
	// case thrift.MAP:
	// 	keyType, valueType, size, e := p.ReadMapBegin()
	// 	if e != nil {
	// 		return wrapError(meta.ErrRead, "", e)
	// 	}
	// 	if keyType != desc.Key().Type() {
	// 		return wrapError(meta.ErrDismatchType, fmt.Sprintf("expect type %s but got type %s", desc.Key().Type(), keyType), nil)
	// 	} else if valueType != desc.Elem().Type() {
	// 		return wrapError(meta.ErrDismatchType, fmt.Sprintf("expect type %s but got type %s", desc.Elem().Type(), valueType), nil)
	// 	}
	// 	*out = json.EncodeObjectBegin(*out)
	// 	for i := 0; i < size; i++ {
	// 		if i != 0 {
	// 			*out = json.EncodeObjectComma(*out)
	// 		}
	// 		err = self.buildinTypeToKey(p, desc.Key(), out)
	// 		if err != nil {
	// 			return wrapError(meta.ErrConvert, "", err)
	// 		}
	// 		*out = json.EncodeObjectColon(*out)
	// 		err = self.doRecurse(ctx, desc.Elem(), out, nil, p)
	// 		if err != nil {
	// 			return unwrapError(fmt.Sprintf("converting %dth element of MAP failed", i), err)
	// 		}
	// 	}
	// 	e = p.ReadMapEnd()
	// 	if e != nil {
	// 		return wrapError(meta.ErrRead, "", e)
	// 	}
	// 	*out = json.EncodeObjectEnd(*out)
	// case thrift.SET, thrift.LIST:
	// 	elemType, size, e := p.ReadSetBegin()
	// 	if e != nil {
	// 		return wrapError(meta.ErrRead, "", e)
	// 	}
	// 	if elemType != desc.Elem().Type() {
	// 		return wrapError(meta.ErrDismatchType, fmt.Sprintf("expect type %s but got type %s", desc.Elem().Type(), elemType), nil)
	// 	}
	// 	*out = json.EncodeArrayBegin(*out)
	// 	for i := 0; i < size; i++ {
	// 		if i != 0 {
	// 			*out = json.EncodeArrayComma(*out)
	// 		}
	// 		err = self.doRecurse(ctx, desc.Elem(), out, nil, p)
	// 		if err != nil {
	// 			return unwrapError(fmt.Sprintf("converting %dth element of SET failed", i), err)
	// 		}
	// 	}
	// 	e = p.ReadSetEnd()
	// 	if e != nil {
	// 		return wrapError(meta.ErrRead, "", e)
	// 	}
	// 	*out = json.EncodeArrayEnd(*out)

	// default:
	// 	return wrapError(meta.ErrUnsupportedType, fmt.Sprintf("unknown descriptor type %s", tt), nil)
	// }
}

// parse Singular MessageType
func (self *ProtoConv) unmarshalSingular(ctx context.Context, resp http.ResponseSetter, p *base.BinaryProtocol, typeId proto.WireType, out *[]byte, fd *proto.FieldDescriptor) (err error) {
	switch (*fd).Kind() {
	case protoreflect.BoolKind:
		v, e := p.ReadBool()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Boolkind error", e)
		}
		*out = json.EncodeBool(*out, v)
	case protoreflect.EnumKind:
		v, e := p.ReadEnum()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Enumkind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case protoreflect.Int32Kind:
		v, e := p.ReadI32()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Int32kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case protoreflect.Sint32Kind:
		v, e := p.ReadSint32()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Sint32kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case protoreflect.Uint32Kind:
		v, e := p.ReadUint32()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Uint32kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case protoreflect.Fixed32Kind:
		v, e := p.ReadFixed32()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Fixed32kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case protoreflect.Sfixed32Kind:
		v, e := p.ReadSfixed32()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Sfixed32kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case protoreflect.Int64Kind:
		v, e := p.ReadI64()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Int64kind error", e)
		}
		if self.opts.String2Int64 {
			*out = append(*out, '"')
			*out = json.EncodeInt64(*out, int64(v))
			*out = append(*out, '"')
		} else {
			*out = json.EncodeInt64(*out, int64(v))
		}
	case protoreflect.Sint64Kind:
		v, e := p.ReadI64()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Sint64kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case protoreflect.Uint64Kind:
		v, e := p.ReadUint64()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Uint64kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case protoreflect.Sfixed64Kind:
		v, e := p.ReadSfixed64()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Sfixed64kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case protoreflect.FloatKind:
		v, e := p.ReadFloat()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Floatkind error", e)
		}
		*out = json.EncodeFloat64(*out, float64(v))
	case protoreflect.DoubleKind:
		v, e := p.ReadDouble()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Doublekind error", e)
		}
		*out = json.EncodeFloat64(*out, float64(v))
	case protoreflect.StringKind:
		v, e := p.ReadString()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Stringkind error", e)
		}
		*out = json.EncodeString(*out, v)
	case protoreflect.BytesKind:
		v, e := p.ReadBytes()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Byteskind error", e)
		}
		*out = json.EncodeBaniry(*out, v)
	case protoreflect.MessageKind:
		// get the message data length
		l, e := p.ReadLength()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Byteskind error", e)
		}
		fields := (*fd).Message().Fields()
		comma := false
		existExceptionField := false
		start := p.Read

		*out = json.EncodeObjectBegin(*out)

		for p.Read < start+l {
			fieldId, typeId, _, e := p.ConsumeTag()
			if e != nil {
				return wrapError(meta.ErrRead, "", e)
			}

			fd := fields.ByNumber(protowire.Number(fieldId))
			if fd == nil {
				return wrapError(meta.ErrRead, "invalid field", nil)
			}

			if comma {
				*out = json.EncodeObjectComma(*out)
			} else {
				comma = true
			}

			// serizalize jsonname
			*out = json.EncodeString(*out, fd.JSONName())
			*out = json.EncodeObjectColon(*out)

			if self.opts.EnableValueMapping {

			} else {
				err := self.doRecurse(ctx, &fd, out, resp, p, typeId)
				if err != nil {
					return unwrapError(fmt.Sprintf("converting field %s of MESSAGE %s failed", fd.Name(), fd.Kind()), err)
				}
			}

			if existExceptionField {
				break
			}
		}
		*out = json.EncodeObjectEnd(*out)
	default:
		return wrapError(meta.ErrUnsupportedType, fmt.Sprintf("unknown descriptor type %s", (*fd).Kind()), nil)
	}
	return
}

// parse ListType
func (self *ProtoConv) unmarshalList(ctx context.Context, resp http.ResponseSetter, p *base.BinaryProtocol, typeId proto.WireType, out *[]byte, fd *proto.FieldDescriptor) (err error) {
	data, err := p.ReadAnyWithDesc(fd, true, false, false)
	if err != nil {
		*out = json.EncodeArrayBegin(*out)
		// write json ...

		*out = json.EncodeArrayEnd(*out)
	}
	return nil
}
