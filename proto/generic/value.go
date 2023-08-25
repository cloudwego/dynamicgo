package generic

import (
	"fmt"
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
)

type Value struct {
	Node
	rootDesc *proto.MessageDescriptor
	Desc *proto.FieldDescriptor
}

func NewRootValue(desc *proto.MessageDescriptor, src []byte) Value {
	return Value{
		Node: NewNode(proto.MESSAGE, src),
		rootDesc: desc,
	}
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

// TODO: need change
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

func searchFieldId(p *binary.BinaryProtocol, id proto.FieldNumber) (int, error) {
	for p.Read < len(p.Buf) {
		fieldNumber, wireType, tagLen, err := p.ConsumeTagWithoutMove()
		if err != nil {
			return 0, err
		}

		if fieldNumber == id {
			return p.Read, nil
		}
		p.Read += tagLen
		if err := p.Skip(wireType, false); err != nil {
			return 0, errNode(meta.ErrRead, "", err)
		}
	}
	return 0, errNotFound
}

func searchIndex(p *binary.BinaryProtocol, idx int, elementWireType proto.WireType) (int, error) {
	fieldNumber, listWireType, _, listTagErr := p.ConsumeTag()
	if listTagErr != nil {
		return 0, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
	}
	
	// packed list
	if listWireType == proto.BytesType {
		// read length
		length, err := p.ReadLength()
		if err != nil {
			return 0, err
		}
		// read list
		start := p.Read
		cnt := 0
		for p.Read < start+length && cnt < idx {
			if err := p.Skip(elementWireType, false); err != nil {
				return 0, errNode(meta.ErrRead, "", err)
			}
			cnt++
		}

		if cnt < idx {
			return 0, errNotFound
		}
	} else {
		cnt := 0
		// normal Type : [tag][(length)][value][tag][(length)][value][tag][(length)][value]....
		for p.Read < len(p.Buf) && cnt < idx {
			// don't move p.Read and judge whether readList completely
			elementFieldNumber, _, n, err := p.ConsumeTagWithoutMove()
			if err != nil {
				return 0, err
			}
			if elementFieldNumber != fieldNumber {
				break
			}
			p.Read += n

			if err := p.Skip(elementWireType, false); err != nil {
				return 0, errNode(meta.ErrRead, "", err)
			}
			cnt++
		}

		if cnt < idx {
			return 0, errNotFound
		} else {
			if cnt == 0 {
				p.ConsumeTag() // skip first taglen
			}
		}
	}
	return p.Read, nil
}

func searchIntKey(p *binary.BinaryProtocol, key int, keyType proto.Type, mapFieldNumber proto.FieldNumber) (int, error) {
	exist := false
	start := p.Read
	// normal Type : [tag][(length)][value][tag][(length)][value][tag][(length)][value]....
	for p.Read < len(p.Buf) {
		// don't move p.Read and judge whether readList completely
		elementFieldNumber, _, n, err := p.ConsumeTagWithoutMove()
		if err != nil {
			return 0, err
		}
		if elementFieldNumber != mapFieldNumber {
			break
		}
		p.Read += n

		p.ReadLength()

		_, _, _, keyTagErr := p.ConsumeTag()
		if keyTagErr != nil {
			return 0, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
		}
		
		k, err := p.ReadInt(keyType)
		if err != nil {
			return 0, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
		}
		// must read valueTag first
		_, valueWireType, _, valueTagErr := p.ConsumeTag()
		if valueTagErr != nil {
			return 0, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
		}

		if k == key {
			exist = true
			start = p.Read // then p.Read will point to real value part
			break
		}
		// if key not match, skip value
		if err := p.Skip(valueWireType, false); err != nil {
			return 0, errNode(meta.ErrRead, "", err)
		}

	}
	if !exist {
		return 0, errNotFound
	}
	return start, nil
}

func searchStrKey(p *binary.BinaryProtocol, key string, keyType proto.Type, mapFieldNumber proto.FieldNumber) (int, error) {
	exist := false
	start := p.Read
	// normal Type : [tag][(length)][value][tag][(length)][value][tag][(length)][value]....
	for p.Read < len(p.Buf) {
		// don't move p.Read and judge whether readList completely
		elementFieldNumber, _, n, err := p.ConsumeTagWithoutMove()
		if err != nil {
			return 0, err
		}
		if elementFieldNumber != mapFieldNumber {
			break
		}
		p.Read += n

		p.ReadLength()

		_, _, _, keyTagErr := p.ConsumeTag()
		if keyTagErr != nil {
			return 0, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
		}
		
		k, err := p.ReadString(false)
		if err != nil {
			return 0, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
		}
		// must read valueTag first
		_, valueWireType, _, valueTagErr := p.ConsumeTag()
		if valueTagErr != nil {
			return 0, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
		}

		if k == key {
			exist = true
			start = p.Read // then p.Read will point to real value part
			break
		}
		// if key not match, skip value
		if err := p.Skip(valueWireType, false); err != nil {
			return 0, errNode(meta.ErrRead, "", err)
		}

	}
	if !exist {
		return 0, errNotFound
	}
	return start, nil
}

func (self Value) GetByPath(pathes ...Path) Value {
	if self.Error() != "" {
		return self
	}

	p := binary.BinaryProtocol{
		Buf: self.raw(),
	}
	start := 0
	var desc *proto.FieldDescriptor
	var err error
	tt := self.t

	for i, path := range pathes {
		switch path.t {
		case PathFieldId:
			id := path.id()
			var fd proto.FieldDescriptor
			start, err = searchFieldId(&p, id)
			if i == 0 {
				fd = (*self.rootDesc).Fields().ByNumber(id)
			} else {
				fd = (*desc).Message().Fields().ByNumber(id)
			}
			if fd != nil {
				desc = &fd
				tt = proto.FromProtoKindToType(fd.Kind(),fd.IsList(),fd.IsMap())
			}
			desc = nil
			
		case PathFieldName:
			name := proto.FieldName(path.str())
			var fd proto.FieldDescriptor
			if i == 0 {
				fd = (*self.rootDesc).Fields().ByName(name)
			} else {
				fd = (*desc).Message().Fields().ByName(name)
			}
			if fd == nil {
				return errValue(meta.ErrUnknownField, fmt.Sprintf("field name '%s' is not defined in IDL", name), nil)
			}
			tt = proto.FromProtoKindToType(fd.Kind(),fd.IsList(),fd.IsMap())
			desc = &fd
			start, err = searchFieldId(&p, fd.Number())
		case PathIndex:
			elemKind := (*desc).Kind()
			elementWireType := proto.Kind2Wire[elemKind]
			start, err = searchIndex(&p, path.int(),elementWireType)
			tt = proto.FromProtoKindToType(elemKind,false,false)
		case PathStrKey:
			mapFieldNumber := (*desc).Number()
			start, err = searchStrKey(&p, path.str(), proto.STRING, mapFieldNumber)
			tt = proto.FromProtoKindToType((*desc).MapValue().Kind(),false,false)
		case PathIntKey:
			keyType := proto.FromProtoKindToType((*desc).MapKey().Kind(),false,false) 
			mapFieldNumber := (*desc).Number()
			start, err = searchIntKey(&p, path.int(), keyType, mapFieldNumber)
			tt = proto.FromProtoKindToType((*desc).MapValue().Kind(),false,false)
		default:
			return errValue(meta.ErrUnsupportedType, fmt.Sprintf("invalid %dth path: %#v", i, p), nil)
		}
		if err != nil {
			// the last one not foud, return start pointer for subsequently inserting operation on `SetByPath()`
			if i == len(pathes)-1 && err == errNotFound {
				return Value{errNotFoundLast(unsafe.Pointer(uintptr(self.v)+uintptr(start)), tt), nil, nil}
			}
			en := err.(Node)
			return errValue(en.ErrCode().Behavior(), "", err)
		}
	}
	
	if err := p.Skip(proto.Kind2Wire[(*desc).Kind()], false); err != nil {
		return errValue(meta.ErrRead, "", err)
	}
	return self.slice(start, p.Read, desc)
}


// SetByPath searches longitudinally and sets a sub value at the given path from the value.
// exist tells whether the node is already exists.
func (self *Value) SetByPath(sub Value, path ...Path) (exist bool, err error) {
	l := len(path)
	if l == 0 {
		*self = sub // it means replace the root value ?
		return true, nil
	}

	if err := self.Check(); err != nil {
		return false, err
	}

	if self.Error() != "" {
		return false, meta.NewError(meta.ErrInvalidParam, "given node is invalid", sub)
	}


	// search source node by path
	v := self.GetByPath(path...)
	if v.IsError() {
		if !v.isErrNotFoundLast() {
			return false, v
		}

		parentPath := path[l-1]
		desc, err := GetDescByPath(self.Desc, path[:l-1]...)
		// if p.t == PathFieldName {
		if err := v.setNotFound(parentPath, &sub.Node); err != nil {
			return false, err
		}
	} else {
		exist = true
	}
	err = self.replace(v.Node, sub.Node)
	return
}
