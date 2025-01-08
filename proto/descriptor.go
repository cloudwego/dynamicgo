package proto

import "github.com/cloudwego/dynamicgo/internal/util"

type TypeDescriptor struct {
	baseId FieldNumber // for LIST/MAP to write field tag by baseId
	typ    Type
	name   string
	key    *TypeDescriptor
	elem   *TypeDescriptor
	msg    *MessageDescriptor // for message, list+message element and map key-value entry
}

func (t *TypeDescriptor) Type() Type {
	return t.typ
}

func (t *TypeDescriptor) Key() *TypeDescriptor {
	return t.key
}

func (t *TypeDescriptor) Elem() *TypeDescriptor {
	return t.elem
}

func (t *TypeDescriptor) Message() *MessageDescriptor {
	return t.msg
}

func (t *TypeDescriptor) BaseId() FieldNumber {
	return t.baseId
}

func (t *TypeDescriptor) IsPacked() bool {
	if t.typ != LIST {
		return false // if not list, return false forever
	}
	return t.elem.typ.IsPacked()
}

func (f *TypeDescriptor) IsMap() bool {
	return f.typ == MAP
}

func (f *TypeDescriptor) IsList() bool {
	return f.typ == LIST
}

func (f *TypeDescriptor) WireType() WireType {
	kind := f.typ.TypeToKind()
	return Kind2Wire[kind]
}

func (f *TypeDescriptor) Name() string {
	return f.name
}

type FieldDescriptor struct {
	kind     ProtoKind // the same value with protobuf descriptor
	id       FieldNumber
	name     string
	jsonName string
	typ      *TypeDescriptor
}

func (f *FieldDescriptor) Number() FieldNumber {
	return f.id
}

func (f *FieldDescriptor) Kind() ProtoKind {
	return f.kind
}

func (f *FieldDescriptor) Name() string {
	return f.name
}

func (f *FieldDescriptor) JSONName() string {
	return f.jsonName
}

func (f *FieldDescriptor) Type() *TypeDescriptor {
	return f.typ
}

// when List+Message it can get message element descriptor
// when Map it can get map key-value entry massage descriptor
// when Message it can get sub message descriptor
func (f *FieldDescriptor) Message() *MessageDescriptor {
	return f.typ.Message()
}

func (f *FieldDescriptor) MapKey() *TypeDescriptor {
	if f.typ.Type() != MAP {
		panic("not map")
	}
	return f.typ.Key()
}

func (f *FieldDescriptor) MapValue() *TypeDescriptor {
	if f.typ.Type() != MAP {
		panic("not map")
	}
	return f.typ.Elem()
}

func (f *FieldDescriptor) IsMap() bool {
	return f.typ.IsMap()
}

func (f *FieldDescriptor) IsList() bool {
	return f.typ.IsList()
}

type MessageDescriptor struct {
	baseId FieldNumber
	name   string
	ids    util.FieldIDMap
	names  util.FieldNameMap // store name and jsonName for FieldDescriptor
}

func (m *MessageDescriptor) Name() string {
	return m.name
}

func (m *MessageDescriptor) ByJSONName(name string) *FieldDescriptor {
	return (*FieldDescriptor)(m.names.Get(name))
}

func (m *MessageDescriptor) ByName(name string) *FieldDescriptor {
	return (*FieldDescriptor)(m.names.Get(name))
}

func (m *MessageDescriptor) ByNumber(id FieldNumber) *FieldDescriptor {
	return (*FieldDescriptor)(m.ids.Get(int32(id)))
}

func (m *MessageDescriptor) FieldsCount() int {
	return m.ids.Size() - 1
}

type MethodDescriptor struct {
	name              string
	input             *TypeDescriptor
	output            *TypeDescriptor
	isClientStreaming bool
	isServerStreaming bool
}

func (m *MethodDescriptor) Name() string {
	return m.name
}

func (m *MethodDescriptor) Input() *TypeDescriptor {
	return m.input
}

func (m *MethodDescriptor) Output() *TypeDescriptor {
	return m.output
}

func (m *MethodDescriptor) IsClientStreaming() bool {
	return m.isClientStreaming
}

func (m *MethodDescriptor) IsServerStreaming() bool {
	return m.isServerStreaming
}

type ServiceDescriptor struct {
	serviceName        string
	methods            map[string]*MethodDescriptor
	isCombinedServices bool
	packageName        string
}

func (s *ServiceDescriptor) Name() string {
	return s.serviceName
}

func (s *ServiceDescriptor) Methods() map[string]*MethodDescriptor {
	return s.methods
}

func (s *ServiceDescriptor) LookupMethodByName(name string) *MethodDescriptor {
	return s.methods[name]
}

func (s *ServiceDescriptor) IsCombinedServices() bool {
	return s.isCombinedServices
}

func (s *ServiceDescriptor) PackageName() string {
	return s.packageName
}
