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

func (self *ProtoConv) do(ctx context.Context, src []byte, desc *proto.Descriptor, out *[]byte, resp http.ResponseSetter) (err error) {
	//NOTICE: output buffer must be larger than src buffer
	rt.GuardSlice(out, len(src)*_GUARD_SLICE_FACTOR)

	var p = binary.BinaryProtocol{
		Buf: src,
	}

	// when desc is Singular/Map/List
	fieldDesc, ok := (*desc).(proto.FieldDescriptor)
	if ok {
		wtyp := proto.Kind2Wire[protoreflect.Kind(fieldDesc.Kind())]
		return self.doRecurse(ctx, &fieldDesc, out, resp, &p, wtyp)
	}

	// when desc is Message
	messageDesc, ok := (*desc).(protoreflect.MessageDescriptor)
	if !ok {
		return wrapError(meta.ErrConvert, "invalid descriptor", nil)
	}
	fields := messageDesc.Fields()
	comma := false

	*out = json.EncodeObjectBegin(*out)

	for p.Read < len(src) {
		// Parse Tag to preprocess Descriptor does not have the field
		fieldId, typeId, _, e := p.ConsumeTag()
		if e != nil {
			return wrapError(meta.ErrRead, "", e)
		}

		fd := fields.ByNumber(protowire.Number(fieldId))
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
		err := self.doRecurse(ctx, &fd, out, resp, &p, typeId)
		if err != nil {
			return unwrapError(fmt.Sprintf("converting field %s of MESSAGE %s failed", fd.Name(), fd.Kind()), err)
		}
	}

	*out = json.EncodeObjectEnd(*out)
	return err
}

// Parse ProtoData into JSONData by DescriptorType
func (self *ProtoConv) doRecurse(ctx context.Context, fd *proto.FieldDescriptor, out *[]byte, resp http.ResponseSetter, p *binary.BinaryProtocol, typeId proto.WireType) error {
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
// Singular format: [Tag][Length][Value]
// Message format: [Tag][Length][[Tag][Length][Value] [Tag][Length][Value]....]
func (self *ProtoConv) unmarshalSingular(ctx context.Context, resp http.ResponseSetter, p *binary.BinaryProtocol, out *[]byte, fd *proto.FieldDescriptor) (err error) {
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
		v, e := p.ReadInt32()
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
		v, e := p.ReadInt64()
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
		v, e := p.ReadInt64()
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
		v, e := p.ReadString(false)
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
		l, e := p.ReadLength()
		if e != nil {
			return wrapError(meta.ErrRead, "unmarshal Byteskind error", e)
		}
		fields := (*fd).Message().Fields()
		comma := false
		start := p.Read

		*out = json.EncodeObjectBegin(*out)

		for p.Read < start+l {
			fieldId, typeId, _, e := p.ConsumeTag()
			if e != nil {
				return wrapError(meta.ErrRead, "", e)
			}

			fd := fields.ByNumber(protowire.Number(fieldId))
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
			err := self.doRecurse(ctx, &fd, out, resp, p, typeId)
			if err != nil {
				return unwrapError(fmt.Sprintf("converting field %s of MESSAGE %s failed", fd.Name(), fd.Kind()), err)
			}
		}
		*out = json.EncodeObjectEnd(*out)
	default:
		return wrapError(meta.ErrUnsupportedType, fmt.Sprintf("unknown descriptor type %s", (*fd).Kind()), nil)
	}
	return
}

// parse ListType
// Packed List format: [Tag][Length][Value Value Value Value Value]....
// Unpacked List format: [Tag][Length][Value] [Tag][Length][Value]....
func (self *ProtoConv) unmarshalList(ctx context.Context, resp http.ResponseSetter, p *binary.BinaryProtocol, typeId proto.WireType, out *[]byte, fd *proto.FieldDescriptor) (err error) {
	*out = json.EncodeArrayBegin(*out)

	fileldNumber := (*fd).Number()
	// packedList(format)：[Tag] [Length] [Value Value Value Value Value]
	if typeId == proto.BytesType && (*fd).IsPacked() {
		len, err := p.ReadLength()
		if err != nil {
			return wrapError(meta.ErrRead, "unmarshal List Length error", err)
		}
		start := p.Read
		// parse Value repeated
		for p.Read < start+len {
			self.unmarshalSingular(ctx, resp, p, out, fd)
			if p.Read != start && p.Read != start+len {
				*out = json.EncodeArrayComma(*out)
			}
		}
	} else {
		// unpackedList(format)：[Tag][Length][Value] [Tag][Length][Value]....
		self.unmarshalSingular(ctx, resp, p, out, fd)
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
			self.unmarshalSingular(ctx, resp, p, out, fd)
		}
	}

	*out = json.EncodeArrayEnd(*out)
	return nil
}

// parse MapType
// Map bytes format: [Pairtag][Pairlength][keyTag(L)V][valueTag(L)V] [Pairtag][Pairlength][T(L)V][T(L)V]...
// Pairtag = MapFieldnumber << 3 | wiretype:BytesType
func (self *ProtoConv) unmarshalMap(ctx context.Context, resp http.ResponseSetter, p *binary.BinaryProtocol, typeId proto.WireType, out *[]byte, fd *proto.FieldDescriptor) (err error) {
	fileldNumber := (*fd).Number()
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
	mapKeyDesc := (*fd).MapKey()
	isIntKey := (mapKeyDesc.Kind() == proto.Int32Kind) || (mapKeyDesc.Kind() == proto.Int64Kind) || (mapKeyDesc.Kind() == proto.Uint32Kind) || (mapKeyDesc.Kind() == proto.Uint64Kind)
	if isIntKey {
		*out = append(*out, '"')
	}
	if self.unmarshalSingular(ctx, resp, p, out, &mapKeyDesc) != nil {
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
	mapValueDesc := (*fd).MapValue()
	if self.unmarshalSingular(ctx, resp, p, out, &mapValueDesc) != nil {
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
		if self.unmarshalSingular(ctx, resp, p, out, &mapKeyDesc) != nil {
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
		if self.unmarshalSingular(ctx, resp, p, out, &mapValueDesc) != nil {
			return wrapError(meta.ErrRead, "parse MapValue Value error", err)
		}
	}

	*out = json.EncodeObjectEnd(*out)
	return nil
}
