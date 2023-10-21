package proto

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Descriptor = protoreflect.Descriptor

type FileDescriptor = protoreflect.FieldDescriptor

type ServiceDescriptor = protoreflect.ServiceDescriptor

type MethodDescriptor = protoreflect.MethodDescriptor

type MessageDescriptor = protoreflect.MessageDescriptor

type FieldDescriptor = protoreflect.FieldDescriptor

// special fielddescriptor for enum, oneof, extension, extensionrange
type EnumDescriptor = protoreflect.EnumDescriptor

type EnumValueDescriptor = protoreflect.EnumValueDescriptor

type OneofDescriptor = protoreflect.OneofDescriptor

// ExtensionDescriptor is the same as FieldDescriptor
type ExtensionDescriptor = protoreflect.ExtensionDescriptor

type ExtensionTypeDescriptor = protoreflect.ExtensionTypeDescriptor

type FieldDescriptors = protoreflect.FieldDescriptors
type TypeDescriptor struct {
	typ Type
	name string
	isPacked bool
	key *TypeDescriptor
	elem *TypeDescriptor
	struc *MessageDescriptor
}

func (t TypeDescriptor) Name() string {
	return t.name
}

func (t TypeDescriptor) Type() Type {
	return t.typ
}

// Key returns the key type descriptor of a MAP type
func (d TypeDescriptor) Key() *TypeDescriptor {
	return d.key
}

// Elem returns the element type descriptor of a LIST, SET or MAP type
func (d TypeDescriptor) Elem() *TypeDescriptor {
	return d.elem
}

// Struct returns the struct type descriptor of a STRUCT type
func (d TypeDescriptor) Struct() *MessageDescriptor {
	return d.struc
}


func Field2TypeDescriptor(desc *FieldDescriptor) *TypeDescriptor {
	fd := *desc
	typ := FromProtoKindToType(fd.Kind(), fd.IsList(), fd.IsMap())

	t := &TypeDescriptor{
		typ: typ,
		name: typ.String(),
	}

	if typ == LIST {
		et := FromProtoKindToType(fd.Kind(), false, fd.IsMap())
		t.elem = &TypeDescriptor{
			typ: et,
			name: et.String(),
		}

		if et == MESSAGE {
			msgDesc := fd.Message()
			t.struc = &msgDesc
		} else if et == MAP {
			key := fd.MapKey()
			value := fd.MapValue()
			t.key = Field2TypeDescriptor(&key)
			t.elem = Field2TypeDescriptor(&value)
		}
	} else if typ == MAP {
		key := fd.MapKey()
		value := fd.MapValue()
		t.key = Field2TypeDescriptor(&key)
		t.elem = Field2TypeDescriptor(&value)
	} else if typ == MESSAGE {
		msgDesc := fd.Message()
		t.struc = &msgDesc
	}

	return t
}

func Message2TypeDescriptor(desc *MessageDescriptor) *TypeDescriptor {
	return &TypeDescriptor{
		typ: MESSAGE,
		name: "MESSAGE",
		struc: desc,
	}
}
