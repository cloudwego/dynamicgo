package proto

// type Descriptor = protoreflect.Descriptor

// type FileDescriptor = protoreflect.FileDescriptor

// type ServiceDescriptor = protoreflect.ServiceDescriptor

// type MethodDescriptor = protoreflect.MethodDescriptor

// type MessageDescriptor = protoreflect.MessageDescriptor

// type FieldDescriptor = protoreflect.FieldDescriptor

type TypeDescriptor struct {
	baseId FieldNumber // for LIST/MAP to write field tag
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

// if List+Message it can get message element descriptor
// if Map it can get map key-value entry massage descriptor
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
	name     FieldName
	jsonName string
	typ      *TypeDescriptor
}

func (f *FieldDescriptor) Number() FieldNumber {
	return f.id
}

func (f *FieldDescriptor) Kind() ProtoKind {
	return f.kind
}

func (f *FieldDescriptor) Name() FieldName {
	return f.name
}

func (f *FieldDescriptor) JSONName() string {
	return f.jsonName
}

func (f *FieldDescriptor) Type() *TypeDescriptor {
	return f.typ
}

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
	baseId    FieldNumber
	name      FieldName
	ids       map[FieldNumber]*FieldDescriptor
	names     map[FieldName]*FieldDescriptor
	jsonNames map[string]*FieldDescriptor
}

func (m *MessageDescriptor) Name() FieldName {
	return m.name
}

func (m *MessageDescriptor) ByJSONName(name string) *FieldDescriptor {
	return m.jsonNames[name]
}

func (m *MessageDescriptor) ByName(name FieldName) *FieldDescriptor {
	return m.names[name]
}

func (m *MessageDescriptor) ByNumber(id FieldNumber) *FieldDescriptor {
	return m.ids[id]
}

func (m *MessageDescriptor) FieldsCount() int {
	return len(m.ids)
}

type MethodDescriptor struct {
	name   string
	input  *TypeDescriptor
	output *TypeDescriptor
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

type ServiceDescriptor struct {
	serviceName string
	methods     map[string]*MethodDescriptor
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
