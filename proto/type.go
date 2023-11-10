package proto

import (
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// protobuf encoding wire type
type WireType int8

const (
	VarintType     WireType = 0
	Fixed32Type    WireType = 5
	Fixed64Type    WireType = 1
	BytesType      WireType = 2
	StartGroupType WireType = 3 // deprecated
	EndGroupType   WireType = 4 // deprecated
)

func (p WireType) String() string {
	switch p {
	case VarintType:
		return "VarintType"
	case Fixed32Type:
		return "Fixed32Type"
	case Fixed64Type:
		return "Fixed64Type"
	case BytesType:
		return "BytesType"
	case StartGroupType:
		return "StartGroupType"
	case EndGroupType:
		return "EndGroupType"
	default:
		return "UnknownWireType"
	}
}

// define ProtoKind = protoreflect.Kind (int8)
type ProtoKind = protoreflect.Kind

const (
	DoubleKind ProtoKind = iota + 1
	FloatKind
	Int64Kind
	Uint64Kind
	Int32Kind
	Fixed64Kind
	Fixed32Kind
	BoolKind
	StringKind
	GroupKind
	MessageKind
	BytesKind
	Uint32Kind
	EnumKind
	Sfixed32Kind
	Sfixed64Kind
	Sint32Kind
	Sint64Kind
)

// map from proto.ProtoKind to proto.WireType
var Kind2Wire = map[ProtoKind]WireType{
	BoolKind:     VarintType,
	EnumKind:     VarintType,
	Int32Kind:    VarintType,
	Sint32Kind:   VarintType,
	Uint32Kind:   VarintType,
	Int64Kind:    VarintType,
	Sint64Kind:   VarintType,
	Uint64Kind:   VarintType,
	Sfixed32Kind: Fixed32Type,
	Fixed32Kind:  Fixed32Type,
	FloatKind:    Fixed32Type,
	Sfixed64Kind: Fixed64Type,
	Fixed64Kind:  Fixed64Type,
	DoubleKind:   Fixed64Type,
	StringKind:   BytesType,
	BytesKind:    BytesType,
	MessageKind:  BytesType,
	GroupKind:    StartGroupType,
}

// Node type (uint8) mapping ProtoKind the same value, except for UNKNOWN, LIST, MAP, ERROR
type Type uint8

const (
	UNKNOWN Type = 0 // unknown field type
	DOUBLE  Type = 1
	FLOAT   Type = 2
	INT64   Type = 3
	UINT64  Type = 4
	INT32   Type = 5
	FIX64   Type = 6
	FIX32   Type = 7
	BOOL    Type = 8
	STRING  Type = 9
	GROUP   Type = 10 // deprecated
	MESSAGE Type = 11
	BYTE    Type = 12
	UINT32  Type = 13
	ENUM    Type = 14
	SFIX32  Type = 15
	SFIX64  Type = 16
	SINT32  Type = 17
	SINT64  Type = 18
	LIST    Type = 19
	MAP     Type = 20
	ERROR   Type = 255
)

func (p Type) Valid() bool {
	switch p {
	case UNKNOWN, BOOL, ENUM, BYTE, INT32, SINT32, UINT32, SFIX32, FIX32, INT64, SINT64, UINT64, SFIX64, FIX64, FLOAT,
		DOUBLE, STRING, MESSAGE, LIST, MAP:
		return true
	default:
		return false
	}
}

func (p Type) TypeToKind() ProtoKind {
	switch p {
	case UNKNOWN, ERROR:
		return 0
	case MAP:
		return MessageKind
	case LIST:
		panic("LIST type has no kind, only list element type has kind")
	}
	return ProtoKind(p)
}

// FromProtoKindTType converts ProtoKind to Type
func FromProtoKindToType(kind ProtoKind, isList bool, isMap bool) Type {
	t := Type(kind)
	if isList {
		t = LIST
	} else if isMap {
		t = MAP
	}
	return t
}

// check if the type need Varint encoding, also used in check list isPacked 
func (p Type) NeedVarint() bool {
	return p == BOOL || p == ENUM || p == INT32 || p == SINT32 || p == UINT32 || p == INT64 || p == SINT64 || p == UINT64
}

// IsInt containing isUint
func (p Type) IsInt() bool {
	return p == INT32 || p == INT64 || p == SFIX32 || p == SFIX64 || p == SINT64 || p == SINT32 || p == UINT32 || p == UINT64 || p == FIX32 || p == FIX64
}

func (p Type) IsUint() bool {
	return p == UINT32 || p == UINT64 || p == FIX32 || p == FIX64
}

// IsComplex tells if the type is one of STRUCT, MAP, SET, LIST
func (p Type) IsComplex() bool {
	return p == MESSAGE || p == MAP || p == LIST
}

// String for format and print
func (p Type) String() string {
	switch p {
	case UNKNOWN:
		return "UNKNOWN"
	case BOOL:
		return "BOOL"
	case ENUM:
		return "ENUM"
	case BYTE:
		return "BYTE"
	case INT32:
		return "INT32"
	case SINT32:
		return "SINT32"
	case UINT32:
		return "UINT32"
	case SFIX32:
		return "SFIX32"
	case FIX32:
		return "FIX32"
	case INT64:
		return "INT64"
	case SINT64:
		return "SINT64"
	case UINT64:
		return "UINT64"
	case SFIX64:
		return "SFIX64"
	case FIX64:
		return "FIX64"
	case FLOAT:
		return "FLOAT"
	case DOUBLE:
		return "DOUBLE"
	case STRING:
		return "STRING"
	case MESSAGE:
		return "MESSAGE"
	case LIST:
		return "LIST"
	case MAP:
		return "MAP"
	default:
		return "ERROR"
	}
}

// define Number = protowire.Number (int32)
type Number = protowire.Number

type FieldNumber = Number
type EnumNumber = Number

// reserved field number min-max ranges in a proto message
const (
	MinValidNumber        Number = 1
	FirstReservedNumber   Number = 19000
	LastReservedNumber    Number = 19999
	MaxValidNumber        Number = 1<<29 - 1
	DefaultRecursionLimit        = 10000
)

// define FieldName = protoreflect.Name (string) used in Descriptor.Name()
type FieldName = protoreflect.Name
