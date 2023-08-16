package generic

import (
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/proto"
)

type Value struct {
	Node
	Desc *proto.FieldDescriptor
}

func NewValue(desc *proto.FieldDescriptor, src []byte) Value {
	typ := proto.FromProtoKindToType((*desc).Kind(), (*desc).IsList(), (*desc).IsMap())
	return Value{
		Node: NewNode(typ, src),
		Desc: desc,
	}
}

// NewValueFromNode copy both Node and TypeDescriptor from another Value.
func (self Value) Fork() Value {
	ret := self
	ret.Node = self.Node.Fork()
	return ret
}

func (self Value) slice(s int, e int, desc *proto.FieldDescriptor) Value {
	t := proto.FromProtoKindToType((*desc).Kind(),(*desc).IsList(),(*desc).IsMap())
	ret := Value{
		Node: Node{
			t: t,
			l: (e - s),
			v: rt.AddPtr(self.v, uintptr(s)),
		},
		Desc: desc,
	}
	if t == proto.LIST {
		ret.et = proto.FromProtoKindToType((*desc).Kind(),false,false) // hard code, may have error?
	} else if t == proto.MAP {
		mapkey := (*desc).MapKey()
		mapvalue := (*desc).MapValue()
		ret.kt = proto.FromProtoKindToType(mapkey.Kind(),mapkey.IsList(),mapkey.IsMap())
		ret.et = proto.FromProtoKindToType(mapvalue.Kind(),mapvalue.IsList(),mapvalue.IsMap())
	}
	return ret
}

