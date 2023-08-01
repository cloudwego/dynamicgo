package generic

import (
	"github.com/cloudwego/dynamicgo/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Value struct {
	Node
	Desc *proto.FieldDescriptor
}

func NewValue(desc *proto.FieldDescriptor, src []byte) Value {
	kind := (*desc).Kind()
	typ := proto.Type(kind)
	if kind == protoreflect.MessageKind{
		if (*desc).IsList(){
			typ = proto.LIST
		}else if (*desc).IsMap(){
			typ = proto.MAP
		}
	}
	return Value{
		Node: NewNode(typ, src),
		Desc: desc,
	}
}