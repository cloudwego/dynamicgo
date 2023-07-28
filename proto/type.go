package proto

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

// protobuf encoding wire type
type wireType int8

const (
	VarintType     wireType = 0
	Fixed32Type    wireType = 5
	Fixed64Type    wireType = 1
	BytesType      wireType = 2
	StartGroupType wireType = 3	 // deprecated
	EndGroupType   wireType = 4  // deprecated
)

// proto message kind
type ProtoKind = protoreflect.Kind

const (
	BoolKind     protoKind = 8
	EnumKind     protoKind = 14
	Int32Kind    protoKind = 5
	Sint32Kind   protoKind = 17
	Uint32Kind   protoKind = 13
	Int64Kind    protoKind = 3
	Sint64Kind   protoKind = 18
	Uint64Kind   protoKind = 4
	Sfixed32Kind protoKind = 15
	Fixed32Kind  protoKind = 7
	FloatKind    protoKind = 2
	Sfixed64Kind protoKind = 16
	Fixed64Kind  protoKind = 6
	DoubleKind   protoKind = 1
	StringKind   protoKind = 9
	BytesKind    protoKind = 12
	MessageKind  protoKind = 11
	GroupKind    protoKind = 10  // deprecated
)

// filed type byte=uint8
type Type byte

// built-in Types
//
//	╔════════════╤═════════════════════════════════════╗
//	║ Go type    │ Protobuf kind                       ║
//	╠════════════╪═════════════════════════════════════╣
//	║ bool       │ BoolKind                            ║
//	║ int32      │ Int32Kind, Sint32Kind, Sfixed32Kind ║
//	║ int64      │ Int64Kind, Sint64Kind, Sfixed64Kind ║
//	║ uint32     │ Uint32Kind, Fixed32Kind             ║
//	║ uint64     │ Uint64Kind, Fixed64Kind             ║
//	║ float32    │ FloatKind                           ║
//	║ float64    │ DoubleKind                          ║
//	║ string     │ StringKind                          ║
//	║ []byte     │ BytesKind                           ║
//	║ EnumNumber │ EnumKind                            ║
//	║ Message    │ MessageKind                         ║
//	║ LIST       │ MessageKind                         ║
//	║ MAP        │ MessageKind                         ║
//	╚════════════╧═════════════════════════════════════╝
//

const (
	NIL     Type = 0 // nil
	BOOL    Type = 1
	INT32   Type = 2
	INT64   Type = 3
	UINT32  Type = 3
	UINT64  Type = 4
	FLOAT32 Type = 6
	FLOAT64 Type = 8
	STRING  Type = 10
	BYTES   Type = 11 // []byte
	ENUM    Type = 12
	MAP     Type = 13
	LIST    Type = 14
	MESSAGE Type = 15
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

type Number int32

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
