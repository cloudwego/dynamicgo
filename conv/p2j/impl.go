package p2j

import (
	"context"
	"fmt"

	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/json"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
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

func (self *BinaryConv) do(ctx context.Context, src []byte, desc *proto.TypeDescriptor, out *[]byte, resp http.ResponseSetter) (err error) {
	//NOTICE: output buffer must be larger than src buffer
	rt.GuardSlice(out, len(src)*_GUARD_SLICE_FACTOR)

	var p = binary.BinaryProtocol{
		Buf: src,
	}

	// when desc is Singular/Map/List
	if desc.Type() != proto.MESSAGE {
		wtyp := proto.Kind2Wire[protoreflect.Kind(desc.Type())]
		return self.doRecurse(ctx, desc, out, resp, &p, wtyp)
	}

	// when desc is Message
	messageDesc := desc.Message()
	comma := false

	*out = json.EncodeObjectBegin(*out)

	for p.Read < len(src) {
		// Parse Tag to preprocess Descriptor does not have the field
		fieldId, typeId, _, e := p.ConsumeTag()
		if e != nil {
			return wrapError(meta.ErrRead, "", e)
		}

		fd := messageDesc.ByNumber(fieldId)
		if fd == nil {
			if self.opts.DisallowUnknownField {
				return wrapError(meta.ErrUnknownField, fmt.Sprintf("unknown field %d", fieldId), nil)
			}
			if e := p.Skip(typeId, self.opts.UseNativeSkip); e != nil {
				return wrapError(meta.ErrRead, "", e)
			}
			continue
		}

		if comma {
			*out = json.EncodeObjectComma(*out)
		} else {
			comma = true
		}

		// NOTICE: always use jsonName here, because jsonName always equals to name by default
		*out = json.EncodeString(*out, fd.JSONName())
		*out = json.EncodeObjectColon(*out)
		// Parse ProtoData and encode into json format
		err := self.doRecurse(ctx, fd.Type(), out, resp, &p, typeId)
		if err != nil {
			return unwrapError(fmt.Sprintf("converting field %s of MESSAGE %s failed", fd.Name(), fd.Kind()), err)
		}
	}

	*out = json.EncodeObjectEnd(*out)
	return err
}

// Parse ProtoData into JSONData by DescriptorType
func (self *BinaryConv) doRecurse(ctx context.Context, fd *proto.TypeDescriptor, out *[]byte, resp http.ResponseSetter, p *binary.BinaryProtocol, typeId proto.WireType) error {
	switch {
	case (*fd).IsList():
		return self.unmarshalList(ctx, resp, p, typeId, out, fd)
	case (*fd).IsMap():
		return self.unmarshalMap(ctx, resp, p, typeId, out, fd)
	default:
		return self.unmarshalSingular(ctx, resp, p, out, fd)
	}
}

// parse Singular/MessageType
// field tag is processed outside before doRecurse
// Singular format:	[(L)V]
// Message format: [Length][[Tag][(L)V] [Tag][(L)V]....]
func (self *BinaryConv) unmarshalSingular(ctx context.Context, resp http.ResponseSetter, p *binary.BinaryProtocol, out *[]byte, fd *proto.TypeDescriptor) (err error) {
	switch fd.Type() {
	case proto.BOOL:
		v, e := p.ReadBool()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Boolkind error", e)
		}
		*out = json.EncodeBool(*out, v)
	case proto.ENUM:
		v, e := p.ReadEnum()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Enumkind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case proto.INT32:
		v, e := p.ReadInt32()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Int32kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case proto.SINT32:
		v, e := p.ReadSint32()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Sint32kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case proto.UINT32:
		v, e := p.ReadUint32()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Uint32kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case proto.FIX32:
		v, e := p.ReadFixed32()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Fixed32kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case proto.SFIX32:
		v, e := p.ReadSfixed32()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Sfixed32kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case proto.INT64:
		v, e := p.ReadInt64()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Int64kind error", e)
		}
		if self.opts.Int642String {
			*out = append(*out, '"')
			*out = json.EncodeInt64(*out, int64(v))
			*out = append(*out, '"')
		} else {
			*out = json.EncodeInt64(*out, int64(v))
		}
	case proto.SINT64:
		v, e := p.ReadSint64()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Sint64kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case proto.UINT64:
		v, e := p.ReadUint64()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Uint64kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case proto.FIX64:
		v, e := p.ReadFixed64()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Fixed64kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case proto.SFIX64:
		v, e := p.ReadSfixed64()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Sfixed64kind error", e)
		}
		*out = json.EncodeInt64(*out, int64(v))
	case proto.FLOAT:
		v, e := p.ReadFloat()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Floatkind error", e)
		}
		*out = json.EncodeFloat64(*out, float64(v))
	case proto.DOUBLE:
		v, e := p.ReadDouble()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Doublekind error", e)
		}
		*out = json.EncodeFloat64(*out, float64(v))
	case proto.STRING:
		v, e := p.ReadString(false)
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Stringkind error", e)
		}
		*out = json.EncodeString(*out, v)
	case proto.BYTE:
		v, e := p.ReadBytes()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Byteskind error", e)
		}
		*out = json.EncodeBaniry(*out, v)
	case proto.MESSAGE:
		l, e := p.ReadLength()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Byteskind error", e)
		}
		message := (*fd).Message()
		comma := false
		start := p.Read

		*out = json.EncodeObjectBegin(*out)

		for p.Read < start+l {
			fieldId, typeId, _, e := p.ConsumeTag()
			if e != nil {
				return wrapError(meta.ErrRead, "", e)
			}

			fd := message.ByNumber(fieldId)
			if fd == nil {
				if self.opts.DisallowUnknownField {
					return wrapError(meta.ErrUnknownField, fmt.Sprintf("unknown field %d", fieldId), nil)
				}
				if e := p.Skip(typeId, self.opts.UseNativeSkip); e != nil {
					return wrapError(meta.ErrRead, "", e)
				}
				continue
			}

			if comma {
				*out = json.EncodeObjectComma(*out)
			} else {
				comma = true
			}

			*out = json.EncodeString(*out, fd.JSONName())
			*out = json.EncodeObjectColon(*out)

			// parse MessageFieldValue recursive
			err := self.doRecurse(ctx, fd.Type(), out, resp, p, typeId)
			if err != nil {
				return unwrapError(fmt.Sprintf("converting field %s of MESSAGE %s failed", fd.Name(), fd.Kind()), err)
			}
		}
		*out = json.EncodeObjectEnd(*out)
	default:
		return wrapError(meta.ErrUnsupportedType, fmt.Sprintf("unknown descriptor type %s", fd.Type()), nil)
	}
	return
}

// parse ListType
// Packed List format: [Tag][Length][Value Value Value Value Value]....
// Unpacked List format: [Tag][Length][Value] [Tag][Length][Value]....
func (self *BinaryConv) unmarshalList(ctx context.Context, resp http.ResponseSetter, p *binary.BinaryProtocol, typeId proto.WireType, out *[]byte, fd *proto.TypeDescriptor) (err error) {
	*out = json.EncodeArrayBegin(*out)

	fileldNumber := fd.BaseId()
	// packedList(format)：[Tag] [Length] [Value Value Value Value Value]
	if typeId == proto.BytesType && (*fd).IsPacked() {
		len, err := p.ReadLength()
		if err != nil {
			return wrapError(meta.ErrRead, "unmarshal List Length error", err)
		}
		start := p.Read
		// parse Value repeated
		for p.Read < start+len {
			self.unmarshalSingular(ctx, resp, p, out, fd.Elem())
			if p.Read != start && p.Read != start+len {
				*out = json.EncodeArrayComma(*out)
			}
		}
	} else {
		// unpackedList(format)：[Tag][Length][Value] [Tag][Length][Value]....
		self.unmarshalSingular(ctx, resp, p, out, fd.Elem())
		for p.Read < len(p.Buf) {
			elementFieldNumber, _, tagLen, err := p.ConsumeTagWithoutMove()

			if err != nil {
				return wrapError(meta.ErrRead, "consume list child Tag error", err)
			}
			// List parse end, pay attention to remove the last ','
			if elementFieldNumber != fileldNumber {
				break
			}
			*out = json.EncodeArrayComma(*out)
			p.Read += tagLen
			self.unmarshalSingular(ctx, resp, p, out, fd.Elem())
		}
	}

	*out = json.EncodeArrayEnd(*out)
	return nil
}

// parse MapType
// Map bytes format: [Pairtag][Pairlength][keyTag(L)V][valueTag(L)V] [Pairtag][Pairlength][T(L)V][T(L)V]...
// Pairtag = MapFieldnumber << 3 | wiretype:BytesType
func (self *BinaryConv) unmarshalMap(ctx context.Context, resp http.ResponseSetter, p *binary.BinaryProtocol, typeId proto.WireType, out *[]byte, fd *proto.TypeDescriptor) (err error) {
	fileldNumber := (*fd).BaseId()
	_, lengthErr := p.ReadLength()
	if lengthErr != nil {
		return wrapError(meta.ErrRead, "parse Tag length error", err)
	}

	*out = json.EncodeObjectBegin(*out)

	// parse first k-v pair, [KeyTag][KeyLength][KeyValue][ValueTag][ValueLength][ValueValue]
	_, _, _, keyErr := p.ConsumeTag()
	if keyErr != nil {
		return wrapError(meta.ErrRead, "parse MapKey Tag error", err)
	}
	mapKeyDesc := fd.Key()
	isIntKey := (mapKeyDesc.Type() == proto.INT32) || (mapKeyDesc.Type() == proto.INT64) || (mapKeyDesc.Type() == proto.UINT32) || (mapKeyDesc.Type() == proto.UINT64)
	if isIntKey {
		*out = append(*out, '"')
	}
	if self.unmarshalSingular(ctx, resp, p, out, mapKeyDesc) != nil {
		return wrapError(meta.ErrRead, "parse MapKey Value error", err)
	}
	if isIntKey {
		*out = append(*out, '"')
	}
	*out = json.EncodeObjectColon(*out)
	_, _, _, valueErr := p.ConsumeTag()
	if valueErr != nil {
		return wrapError(meta.ErrRead, "parse MapValue Tag error", err)
	}
	mapValueDesc := fd.Elem()
	if self.unmarshalSingular(ctx, resp, p, out, mapValueDesc) != nil {
		return wrapError(meta.ErrRead, "parse MapValue Value error", err)
	}

	// parse the remaining k-v pairs
	for p.Read < len(p.Buf) {
		pairNumber, _, tagLen, err := p.ConsumeTagWithoutMove()
		if err != nil {
			return wrapError(meta.ErrRead, "consume list child Tag error", err)
		}
		// parse second Tag
		if pairNumber != fileldNumber {
			break
		}
		p.Read += tagLen
		*out = json.EncodeObjectComma(*out)
		// parse second length
		_, lengthErr := p.ReadLength()
		if lengthErr != nil {
			return wrapError(meta.ErrRead, "parse Tag length error", err)
		}
		// parse second [KeyTag][KeyLength][KeyValue][ValueTag][ValueLength][ValueValue]
		_, _, _, keyErr = p.ConsumeTag()
		if keyErr != nil {
			return wrapError(meta.ErrRead, "parse MapKey Tag error", err)
		}
		if isIntKey {
			*out = append(*out, '"')
		}
		if self.unmarshalSingular(ctx, resp, p, out, mapKeyDesc) != nil {
			return wrapError(meta.ErrRead, "parse MapKey Value error", err)
		}
		if isIntKey {
			*out = append(*out, '"')
		}
		*out = json.EncodeObjectColon(*out)
		_, _, _, valueErr = p.ConsumeTag()
		if valueErr != nil {
			return wrapError(meta.ErrRead, "parse MapValue Tag error", err)
		}
		if self.unmarshalSingular(ctx, resp, p, out, mapValueDesc) != nil {
			return wrapError(meta.ErrRead, "parse MapValue Value error", err)
		}
	}

	*out = json.EncodeObjectEnd(*out)
	return nil
}
