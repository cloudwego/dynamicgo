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

