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

const (
	_ = -iota
	ErrCodeTruncated
	ErrCodeFieldNumber
	ErrCodeOverflow
	ErrCodeReserved
	ErrCodeEndGroup
	ErrCodeRecursionDepth
)

// proto message kind
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


// Node type byte=uint8 reflect protokind
type Type uint8

const (
	NIL     Type = 0 // nil
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
	case NIL, BOOL, ENUM, BYTE, INT32, SINT32, UINT32, SFIX32, FIX32, INT64, SINT64, UINT64, SFIX64, FIX64, FLOAT,
		DOUBLE, STRING, MESSAGE, LIST, MAP:
		return true
	default:
		return false
	}
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
// IsInt tells if the type is one of I08, I16, I32, I64
func (p Type) IsInt() bool {
	return p == INT32 || p == INT64 || p == SFIX32 || p == SFIX32 || p == SINT64 || p == SINT32
}

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


type FieldName = protoreflect.Name
