// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.23.1
// source: baseline.proto

package baseline

import (
	context "context"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Simple struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ByteField   []byte   `protobuf:"bytes,1,opt,name=ByteField,proto3" json:"ByteField,omitempty"`
	I64Field    int64    `protobuf:"varint,2,opt,name=I64Field,proto3" json:"I64Field,omitempty"`
	DoubleField float64  `protobuf:"fixed64,3,opt,name=DoubleField,proto3" json:"DoubleField,omitempty"`
	I32Field    int32    `protobuf:"varint,4,opt,name=I32Field,proto3" json:"I32Field,omitempty"`
	StringField string   `protobuf:"bytes,5,opt,name=StringField,proto3" json:"StringField,omitempty"`
	BinaryField []byte   `protobuf:"bytes,6,opt,name=BinaryField,proto3" json:"BinaryField,omitempty"`
	ListsString []string `protobuf:"bytes,7,rep,name=ListsString,proto3" json:"ListsString,omitempty"`
}

func (x *Simple) Reset() {
	*x = Simple{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baseline_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Simple) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Simple) ProtoMessage() {}

func (x *Simple) ProtoReflect() protoreflect.Message {
	mi := &file_baseline_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Simple.ProtoReflect.Descriptor instead.
func (*Simple) Descriptor() ([]byte, []int) {
	return file_baseline_proto_rawDescGZIP(), []int{0}
}

func (x *Simple) GetByteField() []byte {
	if x != nil {
		return x.ByteField
	}
	return nil
}

func (x *Simple) GetI64Field() int64 {
	if x != nil {
		return x.I64Field
	}
	return 0
}

func (x *Simple) GetDoubleField() float64 {
	if x != nil {
		return x.DoubleField
	}
	return 0
}

func (x *Simple) GetI32Field() int32 {
	if x != nil {
		return x.I32Field
	}
	return 0
}

func (x *Simple) GetStringField() string {
	if x != nil {
		return x.StringField
	}
	return ""
}

func (x *Simple) GetBinaryField() []byte {
	if x != nil {
		return x.BinaryField
	}
	return nil
}

func (x *Simple) GetListsString() []string {
	if x != nil {
		return x.ListsString
	}
	return nil
}

type PartialSimple struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ByteField   []byte  `protobuf:"bytes,1,opt,name=ByteField,proto3" json:"ByteField,omitempty"`
	DoubleField float64 `protobuf:"fixed64,2,opt,name=DoubleField,proto3" json:"DoubleField,omitempty"`
	BinaryField []byte  `protobuf:"bytes,3,opt,name=BinaryField,proto3" json:"BinaryField,omitempty"`
}

func (x *PartialSimple) Reset() {
	*x = PartialSimple{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baseline_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PartialSimple) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PartialSimple) ProtoMessage() {}

func (x *PartialSimple) ProtoReflect() protoreflect.Message {
	mi := &file_baseline_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PartialSimple.ProtoReflect.Descriptor instead.
func (*PartialSimple) Descriptor() ([]byte, []int) {
	return file_baseline_proto_rawDescGZIP(), []int{1}
}

func (x *PartialSimple) GetByteField() []byte {
	if x != nil {
		return x.ByteField
	}
	return nil
}

func (x *PartialSimple) GetDoubleField() float64 {
	if x != nil {
		return x.DoubleField
	}
	return 0
}

func (x *PartialSimple) GetBinaryField() []byte {
	if x != nil {
		return x.BinaryField
	}
	return nil
}

type Nesting struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	String_         string             `protobuf:"bytes,1,opt,name=String,proto3" json:"String,omitempty"`
	ListSimple      []*Simple          `protobuf:"bytes,2,rep,name=ListSimple,proto3" json:"ListSimple,omitempty"`
	Double          float64            `protobuf:"fixed64,3,opt,name=Double,proto3" json:"Double,omitempty"`
	I32             int32              `protobuf:"varint,4,opt,name=I32,proto3" json:"I32,omitempty"`
	ListI32         []int32            `protobuf:"varint,5,rep,packed,name=ListI32,proto3" json:"ListI32,omitempty"`
	I64             int64              `protobuf:"varint,6,opt,name=I64,proto3" json:"I64,omitempty"`
	MapStringString map[string]string  `protobuf:"bytes,7,rep,name=MapStringString,proto3" json:"MapStringString,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	SimpleStruct    *Simple            `protobuf:"bytes,8,opt,name=SimpleStruct,proto3" json:"SimpleStruct,omitempty"`
	MapI32I64       map[int32]int64    `protobuf:"bytes,9,rep,name=MapI32I64,proto3" json:"MapI32I64,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	ListString      []string           `protobuf:"bytes,10,rep,name=ListString,proto3" json:"ListString,omitempty"`
	Binary          []byte             `protobuf:"bytes,11,opt,name=Binary,proto3" json:"Binary,omitempty"`
	MapI64String    map[int64]string   `protobuf:"bytes,12,rep,name=MapI64String,proto3" json:"MapI64String,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	ListI64         []int64            `protobuf:"varint,13,rep,packed,name=ListI64,proto3" json:"ListI64,omitempty"`
	Byte            []byte             `protobuf:"bytes,14,opt,name=Byte,proto3" json:"Byte,omitempty"`
	MapStringSimple map[string]*Simple `protobuf:"bytes,15,rep,name=MapStringSimple,proto3" json:"MapStringSimple,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Nesting) Reset() {
	*x = Nesting{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baseline_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Nesting) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Nesting) ProtoMessage() {}

func (x *Nesting) ProtoReflect() protoreflect.Message {
	mi := &file_baseline_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Nesting.ProtoReflect.Descriptor instead.
func (*Nesting) Descriptor() ([]byte, []int) {
	return file_baseline_proto_rawDescGZIP(), []int{2}
}

func (x *Nesting) GetString_() string {
	if x != nil {
		return x.String_
	}
	return ""
}

func (x *Nesting) GetListSimple() []*Simple {
	if x != nil {
		return x.ListSimple
	}
	return nil
}

func (x *Nesting) GetDouble() float64 {
	if x != nil {
		return x.Double
	}
	return 0
}

func (x *Nesting) GetI32() int32 {
	if x != nil {
		return x.I32
	}
	return 0
}

func (x *Nesting) GetListI32() []int32 {
	if x != nil {
		return x.ListI32
	}
	return nil
}

func (x *Nesting) GetI64() int64 {
	if x != nil {
		return x.I64
	}
	return 0
}

func (x *Nesting) GetMapStringString() map[string]string {
	if x != nil {
		return x.MapStringString
	}
	return nil
}

func (x *Nesting) GetSimpleStruct() *Simple {
	if x != nil {
		return x.SimpleStruct
	}
	return nil
}

func (x *Nesting) GetMapI32I64() map[int32]int64 {
	if x != nil {
		return x.MapI32I64
	}
	return nil
}

func (x *Nesting) GetListString() []string {
	if x != nil {
		return x.ListString
	}
	return nil
}

func (x *Nesting) GetBinary() []byte {
	if x != nil {
		return x.Binary
	}
	return nil
}

func (x *Nesting) GetMapI64String() map[int64]string {
	if x != nil {
		return x.MapI64String
	}
	return nil
}

func (x *Nesting) GetListI64() []int64 {
	if x != nil {
		return x.ListI64
	}
	return nil
}

func (x *Nesting) GetByte() []byte {
	if x != nil {
		return x.Byte
	}
	return nil
}

func (x *Nesting) GetMapStringSimple() map[string]*Simple {
	if x != nil {
		return x.MapStringSimple
	}
	return nil
}

type PartialNesting struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ListSimple      []*PartialSimple          `protobuf:"bytes,2,rep,name=ListSimple,proto3" json:"ListSimple,omitempty"`
	SimpleStruct    *PartialSimple            `protobuf:"bytes,8,opt,name=SimpleStruct,proto3" json:"SimpleStruct,omitempty"`
	MapStringSimple map[string]*PartialSimple `protobuf:"bytes,15,rep,name=MapStringSimple,proto3" json:"MapStringSimple,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *PartialNesting) Reset() {
	*x = PartialNesting{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baseline_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PartialNesting) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PartialNesting) ProtoMessage() {}

func (x *PartialNesting) ProtoReflect() protoreflect.Message {
	mi := &file_baseline_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PartialNesting.ProtoReflect.Descriptor instead.
func (*PartialNesting) Descriptor() ([]byte, []int) {
	return file_baseline_proto_rawDescGZIP(), []int{3}
}

func (x *PartialNesting) GetListSimple() []*PartialSimple {
	if x != nil {
		return x.ListSimple
	}
	return nil
}

func (x *PartialNesting) GetSimpleStruct() *PartialSimple {
	if x != nil {
		return x.SimpleStruct
	}
	return nil
}

func (x *PartialNesting) GetMapStringSimple() map[string]*PartialSimple {
	if x != nil {
		return x.MapStringSimple
	}
	return nil
}

type Nesting2 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MapSimpleNesting map[int64]*Nesting `protobuf:"bytes,1,rep,name=MapSimpleNesting,proto3" json:"MapSimpleNesting,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	SimpleStruct     *Simple            `protobuf:"bytes,2,opt,name=SimpleStruct,proto3" json:"SimpleStruct,omitempty"`
	Byte             []byte             `protobuf:"bytes,3,opt,name=Byte,proto3" json:"Byte,omitempty"`
	Double           float64            `protobuf:"fixed64,4,opt,name=Double,proto3" json:"Double,omitempty"`
	ListNesting      []*Nesting         `protobuf:"bytes,5,rep,name=ListNesting,proto3" json:"ListNesting,omitempty"`
	I64              int64              `protobuf:"varint,6,opt,name=I64,proto3" json:"I64,omitempty"`
	NestingStruct    *Nesting           `protobuf:"bytes,7,opt,name=NestingStruct,proto3" json:"NestingStruct,omitempty"`
	Binary           []byte             `protobuf:"bytes,8,opt,name=Binary,proto3" json:"Binary,omitempty"`
	String_          string             `protobuf:"bytes,9,opt,name=String,proto3" json:"String,omitempty"`
	SetNesting       []*Nesting         `protobuf:"bytes,10,rep,name=SetNesting,proto3" json:"SetNesting,omitempty"`
	I32              int32              `protobuf:"varint,11,opt,name=I32,proto3" json:"I32,omitempty"`
}

func (x *Nesting2) Reset() {
	*x = Nesting2{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baseline_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Nesting2) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Nesting2) ProtoMessage() {}

func (x *Nesting2) ProtoReflect() protoreflect.Message {
	mi := &file_baseline_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Nesting2.ProtoReflect.Descriptor instead.
func (*Nesting2) Descriptor() ([]byte, []int) {
	return file_baseline_proto_rawDescGZIP(), []int{4}
}

func (x *Nesting2) GetMapSimpleNesting() map[int64]*Nesting {
	if x != nil {
		return x.MapSimpleNesting
	}
	return nil
}

func (x *Nesting2) GetSimpleStruct() *Simple {
	if x != nil {
		return x.SimpleStruct
	}
	return nil
}

func (x *Nesting2) GetByte() []byte {
	if x != nil {
		return x.Byte
	}
	return nil
}

func (x *Nesting2) GetDouble() float64 {
	if x != nil {
		return x.Double
	}
	return 0
}

func (x *Nesting2) GetListNesting() []*Nesting {
	if x != nil {
		return x.ListNesting
	}
	return nil
}

func (x *Nesting2) GetI64() int64 {
	if x != nil {
		return x.I64
	}
	return 0
}

func (x *Nesting2) GetNestingStruct() *Nesting {
	if x != nil {
		return x.NestingStruct
	}
	return nil
}

func (x *Nesting2) GetBinary() []byte {
	if x != nil {
		return x.Binary
	}
	return nil
}

func (x *Nesting2) GetString_() string {
	if x != nil {
		return x.String_
	}
	return ""
}

func (x *Nesting2) GetSetNesting() []*Nesting {
	if x != nil {
		return x.SetNesting
	}
	return nil
}

func (x *Nesting2) GetI32() int32 {
	if x != nil {
		return x.I32
	}
	return 0
}

var File_baseline_proto protoreflect.FileDescriptor

var file_baseline_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x62, 0x61, 0x73, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x03, 0x70, 0x62, 0x33, 0x22, 0xe6, 0x01, 0x0a, 0x06, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65,
	0x12, 0x1c, 0x0a, 0x09, 0x42, 0x79, 0x74, 0x65, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x09, 0x42, 0x79, 0x74, 0x65, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x1a,
	0x0a, 0x08, 0x49, 0x36, 0x34, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x08, 0x49, 0x36, 0x34, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x20, 0x0a, 0x0b, 0x44, 0x6f,
	0x75, 0x62, 0x6c, 0x65, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x01, 0x52,
	0x0b, 0x44, 0x6f, 0x75, 0x62, 0x6c, 0x65, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x1a, 0x0a, 0x08,
	0x49, 0x33, 0x32, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08,
	0x49, 0x33, 0x32, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x20, 0x0a, 0x0b, 0x53, 0x74, 0x72, 0x69,
	0x6e, 0x67, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x53,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x20, 0x0a, 0x0b, 0x42, 0x69,
	0x6e, 0x61, 0x72, 0x79, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x0b, 0x42, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x20, 0x0a, 0x0b,
	0x4c, 0x69, 0x73, 0x74, 0x73, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x07, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x0b, 0x4c, 0x69, 0x73, 0x74, 0x73, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x22, 0x71,
	0x0a, 0x0d, 0x50, 0x61, 0x72, 0x74, 0x69, 0x61, 0x6c, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x12,
	0x1c, 0x0a, 0x09, 0x42, 0x79, 0x74, 0x65, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x09, 0x42, 0x79, 0x74, 0x65, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x20, 0x0a,
	0x0b, 0x44, 0x6f, 0x75, 0x62, 0x6c, 0x65, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x01, 0x52, 0x0b, 0x44, 0x6f, 0x75, 0x62, 0x6c, 0x65, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12,
	0x20, 0x0a, 0x0b, 0x42, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x42, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x46, 0x69, 0x65, 0x6c,
	0x64, 0x22, 0xe8, 0x06, 0x0a, 0x07, 0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x12, 0x16, 0x0a,
	0x06, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x53,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x2b, 0x0a, 0x0a, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x69, 0x6d,
	0x70, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x70, 0x62, 0x33, 0x2e,
	0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x0a, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x69, 0x6d, 0x70,
	0x6c, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x44, 0x6f, 0x75, 0x62, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x01, 0x52, 0x06, 0x44, 0x6f, 0x75, 0x62, 0x6c, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x49, 0x33,
	0x32, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x49, 0x33, 0x32, 0x12, 0x18, 0x0a, 0x07,
	0x4c, 0x69, 0x73, 0x74, 0x49, 0x33, 0x32, 0x18, 0x05, 0x20, 0x03, 0x28, 0x05, 0x52, 0x07, 0x4c,
	0x69, 0x73, 0x74, 0x49, 0x33, 0x32, 0x12, 0x10, 0x0a, 0x03, 0x49, 0x36, 0x34, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x03, 0x49, 0x36, 0x34, 0x12, 0x4b, 0x0a, 0x0f, 0x4d, 0x61, 0x70, 0x53,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x07, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x21, 0x2e, 0x70, 0x62, 0x33, 0x2e, 0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e,
	0x4d, 0x61, 0x70, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x0f, 0x4d, 0x61, 0x70, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x53,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x2f, 0x0a, 0x0c, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x53,
	0x74, 0x72, 0x75, 0x63, 0x74, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x70, 0x62,
	0x33, 0x2e, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x0c, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65,
	0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x12, 0x39, 0x0a, 0x09, 0x4d, 0x61, 0x70, 0x49, 0x33, 0x32,
	0x49, 0x36, 0x34, 0x18, 0x09, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x70, 0x62, 0x33, 0x2e,
	0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x4d, 0x61, 0x70, 0x49, 0x33, 0x32, 0x49, 0x36,
	0x34, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x09, 0x4d, 0x61, 0x70, 0x49, 0x33, 0x32, 0x49, 0x36,
	0x34, 0x12, 0x1e, 0x0a, 0x0a, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18,
	0x0a, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0a, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x74, 0x72, 0x69, 0x6e,
	0x67, 0x12, 0x16, 0x0a, 0x06, 0x42, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x18, 0x0b, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x06, 0x42, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x12, 0x42, 0x0a, 0x0c, 0x4d, 0x61, 0x70,
	0x49, 0x36, 0x34, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x0c, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1e, 0x2e, 0x70, 0x62, 0x33, 0x2e, 0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x4d, 0x61,
	0x70, 0x49, 0x36, 0x34, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x0c, 0x4d, 0x61, 0x70, 0x49, 0x36, 0x34, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x18, 0x0a,
	0x07, 0x4c, 0x69, 0x73, 0x74, 0x49, 0x36, 0x34, 0x18, 0x0d, 0x20, 0x03, 0x28, 0x03, 0x52, 0x07,
	0x4c, 0x69, 0x73, 0x74, 0x49, 0x36, 0x34, 0x12, 0x12, 0x0a, 0x04, 0x42, 0x79, 0x74, 0x65, 0x18,
	0x0e, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x42, 0x79, 0x74, 0x65, 0x12, 0x4b, 0x0a, 0x0f, 0x4d,
	0x61, 0x70, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x18, 0x0f,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x70, 0x62, 0x33, 0x2e, 0x4e, 0x65, 0x73, 0x74, 0x69,
	0x6e, 0x67, 0x2e, 0x4d, 0x61, 0x70, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x53, 0x69, 0x6d, 0x70,
	0x6c, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0f, 0x4d, 0x61, 0x70, 0x53, 0x74, 0x72, 0x69,
	0x6e, 0x67, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x1a, 0x42, 0x0a, 0x14, 0x4d, 0x61, 0x70, 0x53,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x3c, 0x0a, 0x0e,
	0x4d, 0x61, 0x70, 0x49, 0x33, 0x32, 0x49, 0x36, 0x34, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x3f, 0x0a, 0x11, 0x4d, 0x61,
	0x70, 0x49, 0x36, 0x34, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x4f, 0x0a, 0x14, 0x4d,
	0x61, 0x70, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x21, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x70, 0x62, 0x33, 0x2e, 0x53, 0x69, 0x6d, 0x70, 0x6c,
	0x65, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xa8, 0x02, 0x0a,
	0x0e, 0x50, 0x61, 0x72, 0x74, 0x69, 0x61, 0x6c, 0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x12,
	0x32, 0x0a, 0x0a, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x70, 0x62, 0x33, 0x2e, 0x50, 0x61, 0x72, 0x74, 0x69, 0x61,
	0x6c, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x0a, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x69, 0x6d,
	0x70, 0x6c, 0x65, 0x12, 0x36, 0x0a, 0x0c, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x53, 0x74, 0x72,
	0x75, 0x63, 0x74, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x70, 0x62, 0x33, 0x2e,
	0x50, 0x61, 0x72, 0x74, 0x69, 0x61, 0x6c, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x0c, 0x53,
	0x69, 0x6d, 0x70, 0x6c, 0x65, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x12, 0x52, 0x0a, 0x0f, 0x4d,
	0x61, 0x70, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x18, 0x0f,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x70, 0x62, 0x33, 0x2e, 0x50, 0x61, 0x72, 0x74, 0x69,
	0x61, 0x6c, 0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x4d, 0x61, 0x70, 0x53, 0x74, 0x72,
	0x69, 0x6e, 0x67, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0f,
	0x4d, 0x61, 0x70, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x1a,
	0x56, 0x0a, 0x14, 0x4d, 0x61, 0x70, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x53, 0x69, 0x6d, 0x70,
	0x6c, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x28, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x70, 0x62, 0x33, 0x2e, 0x50,
	0x61, 0x72, 0x74, 0x69, 0x61, 0x6c, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xf1, 0x03, 0x0a, 0x08, 0x4e, 0x65, 0x73, 0x74,
	0x69, 0x6e, 0x67, 0x32, 0x12, 0x4f, 0x0a, 0x10, 0x4d, 0x61, 0x70, 0x53, 0x69, 0x6d, 0x70, 0x6c,
	0x65, 0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x23,
	0x2e, 0x70, 0x62, 0x33, 0x2e, 0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x32, 0x2e, 0x4d, 0x61,
	0x70, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x10, 0x4d, 0x61, 0x70, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x4e, 0x65,
	0x73, 0x74, 0x69, 0x6e, 0x67, 0x12, 0x2f, 0x0a, 0x0c, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x53,
	0x74, 0x72, 0x75, 0x63, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x70, 0x62,
	0x33, 0x2e, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x52, 0x0c, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65,
	0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x42, 0x79, 0x74, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x42, 0x79, 0x74, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x44, 0x6f,
	0x75, 0x62, 0x6c, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x01, 0x52, 0x06, 0x44, 0x6f, 0x75, 0x62,
	0x6c, 0x65, 0x12, 0x2e, 0x0a, 0x0b, 0x4c, 0x69, 0x73, 0x74, 0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e,
	0x67, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x70, 0x62, 0x33, 0x2e, 0x4e, 0x65,
	0x73, 0x74, 0x69, 0x6e, 0x67, 0x52, 0x0b, 0x4c, 0x69, 0x73, 0x74, 0x4e, 0x65, 0x73, 0x74, 0x69,
	0x6e, 0x67, 0x12, 0x10, 0x0a, 0x03, 0x49, 0x36, 0x34, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x03, 0x49, 0x36, 0x34, 0x12, 0x32, 0x0a, 0x0d, 0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x53,
	0x74, 0x72, 0x75, 0x63, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x70, 0x62,
	0x33, 0x2e, 0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x52, 0x0d, 0x4e, 0x65, 0x73, 0x74, 0x69,
	0x6e, 0x67, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x42, 0x69, 0x6e, 0x61,
	0x72, 0x79, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x42, 0x69, 0x6e, 0x61, 0x72, 0x79,
	0x12, 0x16, 0x0a, 0x06, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x2c, 0x0a, 0x0a, 0x53, 0x65, 0x74, 0x4e,
	0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x18, 0x0a, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x70,
	0x62, 0x33, 0x2e, 0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x52, 0x0a, 0x53, 0x65, 0x74, 0x4e,
	0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x12, 0x10, 0x0a, 0x03, 0x49, 0x33, 0x32, 0x18, 0x0b, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x03, 0x49, 0x33, 0x32, 0x1a, 0x51, 0x0a, 0x15, 0x4d, 0x61, 0x70, 0x53,
	0x69, 0x6d, 0x70, 0x6c, 0x65, 0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x22, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x70, 0x62, 0x33, 0x2e, 0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x32, 0x99, 0x02, 0x0a, 0x0f,
	0x42, 0x61, 0x73, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x28, 0x0a, 0x0c, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12,
	0x0b, 0x2e, 0x70, 0x62, 0x33, 0x2e, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x1a, 0x0b, 0x2e, 0x70,
	0x62, 0x33, 0x2e, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x12, 0x3d, 0x0a, 0x13, 0x50, 0x61, 0x72,
	0x74, 0x69, 0x61, 0x6c, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64,
	0x12, 0x12, 0x2e, 0x70, 0x62, 0x33, 0x2e, 0x50, 0x61, 0x72, 0x74, 0x69, 0x61, 0x6c, 0x53, 0x69,
	0x6d, 0x70, 0x6c, 0x65, 0x1a, 0x12, 0x2e, 0x70, 0x62, 0x33, 0x2e, 0x50, 0x61, 0x72, 0x74, 0x69,
	0x61, 0x6c, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x12, 0x2b, 0x0a, 0x0d, 0x4e, 0x65, 0x73, 0x74,
	0x69, 0x6e, 0x67, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x0c, 0x2e, 0x70, 0x62, 0x33, 0x2e,
	0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x1a, 0x0c, 0x2e, 0x70, 0x62, 0x33, 0x2e, 0x4e, 0x65,
	0x73, 0x74, 0x69, 0x6e, 0x67, 0x12, 0x40, 0x0a, 0x14, 0x50, 0x61, 0x72, 0x74, 0x69, 0x61, 0x6c,
	0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x13, 0x2e,
	0x70, 0x62, 0x33, 0x2e, 0x50, 0x61, 0x72, 0x74, 0x69, 0x61, 0x6c, 0x4e, 0x65, 0x73, 0x74, 0x69,
	0x6e, 0x67, 0x1a, 0x13, 0x2e, 0x70, 0x62, 0x33, 0x2e, 0x50, 0x61, 0x72, 0x74, 0x69, 0x61, 0x6c,
	0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x12, 0x2e, 0x0a, 0x0e, 0x4e, 0x65, 0x73, 0x74, 0x69,
	0x6e, 0x67, 0x32, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x0d, 0x2e, 0x70, 0x62, 0x33, 0x2e,
	0x4e, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x32, 0x1a, 0x0d, 0x2e, 0x70, 0x62, 0x33, 0x2e, 0x4e,
	0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x32, 0x42, 0x3f, 0x5a, 0x3d, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x77, 0x65, 0x67, 0x6f, 0x2f,
	0x64, 0x79, 0x6e, 0x61, 0x6d, 0x69, 0x63, 0x67, 0x6f, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x64, 0x61,
	0x74, 0x61, 0x2f, 0x6b, 0x69, 0x74, 0x65, 0x78, 0x5f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x62, 0x2f,
	0x62, 0x61, 0x73, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_baseline_proto_rawDescOnce sync.Once
	file_baseline_proto_rawDescData = file_baseline_proto_rawDesc
)

func file_baseline_proto_rawDescGZIP() []byte {
	file_baseline_proto_rawDescOnce.Do(func() {
		file_baseline_proto_rawDescData = protoimpl.X.CompressGZIP(file_baseline_proto_rawDescData)
	})
	return file_baseline_proto_rawDescData
}

var file_baseline_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_baseline_proto_goTypes = []interface{}{
	(*Simple)(nil),         // 0: pb3.Simple
	(*PartialSimple)(nil),  // 1: pb3.PartialSimple
	(*Nesting)(nil),        // 2: pb3.Nesting
	(*PartialNesting)(nil), // 3: pb3.PartialNesting
	(*Nesting2)(nil),       // 4: pb3.Nesting2
	nil,                    // 5: pb3.Nesting.MapStringStringEntry
	nil,                    // 6: pb3.Nesting.MapI32I64Entry
	nil,                    // 7: pb3.Nesting.MapI64StringEntry
	nil,                    // 8: pb3.Nesting.MapStringSimpleEntry
	nil,                    // 9: pb3.PartialNesting.MapStringSimpleEntry
	nil,                    // 10: pb3.Nesting2.MapSimpleNestingEntry
}
var file_baseline_proto_depIdxs = []int32{
	0,  // 0: pb3.Nesting.ListSimple:type_name -> pb3.Simple
	5,  // 1: pb3.Nesting.MapStringString:type_name -> pb3.Nesting.MapStringStringEntry
	0,  // 2: pb3.Nesting.SimpleStruct:type_name -> pb3.Simple
	6,  // 3: pb3.Nesting.MapI32I64:type_name -> pb3.Nesting.MapI32I64Entry
	7,  // 4: pb3.Nesting.MapI64String:type_name -> pb3.Nesting.MapI64StringEntry
	8,  // 5: pb3.Nesting.MapStringSimple:type_name -> pb3.Nesting.MapStringSimpleEntry
	1,  // 6: pb3.PartialNesting.ListSimple:type_name -> pb3.PartialSimple
	1,  // 7: pb3.PartialNesting.SimpleStruct:type_name -> pb3.PartialSimple
	9,  // 8: pb3.PartialNesting.MapStringSimple:type_name -> pb3.PartialNesting.MapStringSimpleEntry
	10, // 9: pb3.Nesting2.MapSimpleNesting:type_name -> pb3.Nesting2.MapSimpleNestingEntry
	0,  // 10: pb3.Nesting2.SimpleStruct:type_name -> pb3.Simple
	2,  // 11: pb3.Nesting2.ListNesting:type_name -> pb3.Nesting
	2,  // 12: pb3.Nesting2.NestingStruct:type_name -> pb3.Nesting
	2,  // 13: pb3.Nesting2.SetNesting:type_name -> pb3.Nesting
	0,  // 14: pb3.Nesting.MapStringSimpleEntry.value:type_name -> pb3.Simple
	1,  // 15: pb3.PartialNesting.MapStringSimpleEntry.value:type_name -> pb3.PartialSimple
	2,  // 16: pb3.Nesting2.MapSimpleNestingEntry.value:type_name -> pb3.Nesting
	0,  // 17: pb3.BaselineService.SimpleMethod:input_type -> pb3.Simple
	1,  // 18: pb3.BaselineService.PartialSimpleMethod:input_type -> pb3.PartialSimple
	2,  // 19: pb3.BaselineService.NestingMethod:input_type -> pb3.Nesting
	3,  // 20: pb3.BaselineService.PartialNestingMethod:input_type -> pb3.PartialNesting
	4,  // 21: pb3.BaselineService.Nesting2Method:input_type -> pb3.Nesting2
	0,  // 22: pb3.BaselineService.SimpleMethod:output_type -> pb3.Simple
	1,  // 23: pb3.BaselineService.PartialSimpleMethod:output_type -> pb3.PartialSimple
	2,  // 24: pb3.BaselineService.NestingMethod:output_type -> pb3.Nesting
	3,  // 25: pb3.BaselineService.PartialNestingMethod:output_type -> pb3.PartialNesting
	4,  // 26: pb3.BaselineService.Nesting2Method:output_type -> pb3.Nesting2
	22, // [22:27] is the sub-list for method output_type
	17, // [17:22] is the sub-list for method input_type
	17, // [17:17] is the sub-list for extension type_name
	17, // [17:17] is the sub-list for extension extendee
	0,  // [0:17] is the sub-list for field type_name
}

func init() { file_baseline_proto_init() }
func file_baseline_proto_init() {
	if File_baseline_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_baseline_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Simple); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_baseline_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PartialSimple); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_baseline_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Nesting); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_baseline_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PartialNesting); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_baseline_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Nesting2); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_baseline_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_baseline_proto_goTypes,
		DependencyIndexes: file_baseline_proto_depIdxs,
		MessageInfos:      file_baseline_proto_msgTypes,
	}.Build()
	File_baseline_proto = out.File
	file_baseline_proto_rawDesc = nil
	file_baseline_proto_goTypes = nil
	file_baseline_proto_depIdxs = nil
}

var _ context.Context

// Code generated by Kitex v0.7.2. DO NOT EDIT.

type BaselineService interface {
	SimpleMethod(ctx context.Context, req *Simple) (res *Simple, err error)
	PartialSimpleMethod(ctx context.Context, req *PartialSimple) (res *PartialSimple, err error)
	NestingMethod(ctx context.Context, req *Nesting) (res *Nesting, err error)
	PartialNestingMethod(ctx context.Context, req *PartialNesting) (res *PartialNesting, err error)
	Nesting2Method(ctx context.Context, req *Nesting2) (res *Nesting2, err error)
}
