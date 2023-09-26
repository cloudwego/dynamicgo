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





func searchFieldId(p *binary.BinaryProtocol, id proto.FieldNumber, messageLen int) (int, error) {
	for p.Read < messageLen {
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

func searchIndex(p *binary.BinaryProtocol, idx int, elementWireType proto.WireType, isPacked bool, fieldNumber proto.Number) (int, error) {
	// packed list
	cnt := 0
	result := 0
	if isPacked {
		// read length
		length, err := p.ReadLength()
		if err != nil {
			return 0, err
		}
		// read list
		start := p.Read
		for p.Read < start+length && cnt < idx {
			if err := p.Skip(elementWireType, false); err != nil {
				return 0, errNode(meta.ErrRead, "", err)
			}
			cnt++
		}
		result = p.Read
	} else {
		// normal Type : [tag][(length)][value][tag][(length)][value][tag][(length)][value]....
		for p.Read < len(p.Buf) && cnt < idx {
			// don't move p.Read and judge whether readList completely
			if err := p.Skip(elementWireType, false); err != nil {
				return 0, errNode(meta.ErrRead, "", err)
			}
			cnt++
			if p.Read < len(p.Buf) {
				// don't move p.Read and judge whether readList completely
				elementFieldNumber, _, n, err := p.ConsumeTagWithoutMove()
				if err != nil {
					return 0, err
				}
				if elementFieldNumber != fieldNumber {
					break
				}
				if cnt < idx {
					p.Read += n
				}
				result = p.Read + n
			}
		}

	}

	if cnt < idx {
		return 0, errNotFound
	} 

	return result, nil
}

func searchIntKey(p *binary.BinaryProtocol, key int, keyType proto.Type, mapFieldNumber proto.FieldNumber) (int, error) {
	exist := false
	start := p.Read
	// normal Type : [tag][(length)][value][tag][(length)][value][tag][(length)][value]....
	for p.Read < len(p.Buf) {
		if _, err := p.ReadLength() ; err != nil {
			return 0, meta.NewError(meta.ErrRead, "ReadLength failed", nil)
		}

		if _, _, _, keyTagErr := p.ConsumeTag(); keyTagErr != nil {
			return 0, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
		}

		k, err := p.ReadInt(keyType)
		if err != nil {
			return 0, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
		}
		

		if k == key {
			exist = true
			_,_, tagLen, err := p.ConsumeTagWithoutMove() // skip value taglen
			if err != nil {
				return 0, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
			}
			start = p.Read + tagLen // then p.Read will point to real value part
			break
		}

		_, valueWireType, _, valueTagErr := p.ConsumeTag()
		if valueTagErr != nil {
			return 0, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
		}

		// if key not match, skip value
		if err := p.Skip(valueWireType, false); err != nil {
			return 0, errNode(meta.ErrRead, "", err)
		}

		if p.Read >= len(p.Buf) {
			break
		}

		// don't move p.Read and judge whether readList completely
		elementFieldNumber, _, n, err := p.ConsumeTagWithoutMove()
		if err != nil {
			return 0, err
		}
		if elementFieldNumber != mapFieldNumber {
			break
		}
		p.Read += n
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
		if _, err := p.ReadLength() ; err != nil {
			return 0, meta.NewError(meta.ErrRead, "ReadLength failed", nil)
		}

		if _, _, _, keyTagErr := p.ConsumeTag(); keyTagErr != nil {
			return 0, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
		}
		
		k, err := p.ReadString(false)
		if err != nil {
			return 0, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
		}

		if k == key {
			exist = true
			_,_, tagLen, err := p.ConsumeTagWithoutMove() // skip value taglen
			if err != nil {
				return 0, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
			}
			start = p.Read + tagLen // then p.Read will point to real value part
			break
		}


		_, valueWireType, _, valueTagErr := p.ConsumeTag()
		if valueTagErr != nil {
			return 0, meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
		}

		// if key not match, skip value
		if err := p.Skip(valueWireType, false); err != nil {
			return 0, errNode(meta.ErrRead, "", err)
		}

		if p.Read >= len(p.Buf) {
			break
		}

		// don't move p.Read and judge whether readList completely
		elementFieldNumber, _, n, err := p.ConsumeTagWithoutMove()
		if err != nil {
			return 0, err
		}
		if elementFieldNumber != mapFieldNumber {
			break
		}
		p.Read += n
	}
	if !exist {
		return 0, errNotFound
	}
	return start, nil
}

func (self Value) GetByPath(pathes ...Path) (Value, []int) {
	address := make([]int, len(pathes))
	start := 0
	var desc *proto.FieldDescriptor
	var err error
	tt := self.t
	kt := self.kt
	et := self.et

	if self.Error() != "" {
		return self, address
	}

	p := binary.BinaryProtocol{
		Buf: self.raw(),
	}

	for i, path := range pathes {
		switch path.t {
		case PathFieldId:
			id := path.id()
			var fd proto.FieldDescriptor
			messageLen := 0
			if i == 0 {
				fd = (*self.rootDesc).Fields().ByNumber(id)
				messageLen = len(p.Buf)
			} else {
				fd = (*desc).Message().Fields().ByNumber(id)
				Len, err := p.ReadLength()
				if err != nil {
					return errValue(meta.ErrRead, "", err), address
				}
				messageLen += Len
			}
			if fd != nil {
				desc = &fd
				tt = proto.FromProtoKindToType(fd.Kind(),fd.IsList(),fd.IsMap())
			}
			start, err = searchFieldId(&p, id, messageLen)
		case PathFieldName:
			name := proto.FieldName(path.str())
			var fd proto.FieldDescriptor
			messageLen := 0
			if i == 0 {
				fd = (*self.rootDesc).Fields().ByName(name)
				messageLen = len(p.Buf)
			} else {
				fd = (*desc).Message().Fields().ByName(name)
				Len, err := p.ReadLength()
				if err != nil {
					return errValue(meta.ErrRead, "", err), address
				}
				messageLen += Len
			}
			if fd == nil {
				return errValue(meta.ErrUnknownField, fmt.Sprintf("field name '%s' is not defined in IDL", name), nil), address
			}
			tt = proto.FromProtoKindToType(fd.Kind(),fd.IsList(),fd.IsMap())
			desc = &fd
			start, err = searchFieldId(&p, fd.Number(), messageLen)
		case PathIndex:
			elemKind := (*desc).Kind()
			elementWireType := proto.Kind2Wire[elemKind]
			isPacked := (*desc).IsPacked()
			start, err = searchIndex(&p, path.int(),elementWireType, isPacked, (*desc).Number())
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
			return errValue(meta.ErrUnsupportedType, fmt.Sprintf("invalid %dth path: %#v", i, p), nil), address
		}

		address[i] = start

		if err != nil {
			// the last one not foud, return start pointer for subsequently inserting operation on `SetByPath()`
			if i == len(pathes)-1 && err == errNotFound {
				return Value{errNotFoundLast(unsafe.Pointer(uintptr(self.v)+uintptr(start)), tt), nil, nil}, address
			}
			en := err.(Node)
			return errValue(en.ErrCode().Behavior(), "", err), address
		}
		
		if i != len(pathes)-1 {
			if _,_,_,err := p.ConsumeTag(); err != nil {
				return errValue(meta.ErrRead, "", err), address
			}
			start = p.Read
		}
		
	}
	

	if tt == proto.MAP {
		kt = proto.FromProtoKindToType((*desc).MapKey().Kind(),false,false)
		et = proto.FromProtoKindToType((*desc).MapValue().Kind(),false,false)
		if _, err = p.ReadMap(desc, false, false, false); err != nil {
			en := err.(Node)
			return errValue(en.ErrCode().Behavior(), "", err), address
		}
	} else if tt == proto.LIST {
		et = proto.FromProtoKindToType((*desc).Kind(),false,false)
		if _, err = p.ReadList(desc, false, false, false); err != nil {
			en := err.(Node)
			return errValue(en.ErrCode().Behavior(), "", err), address
		}
	} else {
		skipType := proto.Kind2Wire[(*desc).Kind()]
		if (*desc).IsPacked() == false {
			if _, _, _, err := p.ConsumeTag(); err != nil {
				return errValue(meta.ErrRead, "", err), address
			}
			start = p.Read
		}
		
		if err := p.Skip(skipType, false); err != nil {
			return errValue(meta.ErrRead, "", err), address
		}
	}
	return wrapValue(self.Node.slice(start, p.Read, tt, kt, et), desc), address
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
	v, address := self.GetByPath(path...)
	if v.IsError() {
		if !v.isErrNotFoundLast() {
			return false, v
		}

		parentPath := path[l-1]
		desc, err := GetDescByPath(self.rootDesc, path[:l-1]...)
		if err != nil {
			return false, err
		}
		// may have error
		if parentPath.t == PathFieldName {
			f := (*desc).Message().Fields().ByName(proto.FieldName(parentPath.str()))
			parentPath = NewPathFieldId(f.Number())
		}

		if err := v.setNotFound(parentPath, &sub.Node, desc); err != nil {
			return false, err
		}
	} else {
		exist = true
	}
	originLen := len(self.raw())
	err = self.replace(v.Node, sub.Node)
	self.UpdateByteLen(originLen,address, path...)
	return
}

func (self *Value) UpdateByteLen(originLen int, address []int, path ...Path) {
	afterLen := len(self.raw())
	diffLen := afterLen - originLen
	fmt.Println("diffLen", diffLen)
	for i := len(address) - 1; i > 0; i-- {
		fmt.Println("address",i, address[i])
		fmt.Println("path",i, path[i])
	}
}


// UnsetByPath searches longitudinally and unsets a sub value at the given path from the value.
func (self *Value) UnsetByPath(path ...Path) error {
	l := len(path)
	if l == 0 {
		*self = Value{}
		return nil
	}
	if err := self.Check(); err != nil {
		return err
	}
	// search parent node by path
	var v,_ = self.GetByPath(path[:l-1]...)
	if v.IsError() {
		if v.IsErrNotFound() {
			return nil
		}
		return v
	}
	p := path[l-1]
	desc, err := GetDescByPath(self.rootDesc, path[:l-1]...)
	if p.t == PathFieldName {
		if err != nil {
			return err
		}
		f := (*desc).Message().Fields().ByName(proto.FieldName(p.str()))
		p = NewPathFieldId(f.Number())
	}
	ret := v.deleteChild(p)
	if ret.IsError() {
		return ret
	}
	return self.replace(ret, Node{t: ret.t})
}

// MarshalTo marshals self value into a sub value descripted by the to descriptor, alse called as "Cutting".
// Usually, the to descriptor is a subset of self descriptor.
func (self Value) MarshalTo(to *proto.MessageDescriptor, opts *Options) ([]byte, error) {
	var w = binary.NewBinaryProtocolBuffer()
	var r = binary.BinaryProtocol{}
	r.Buf = self.raw()
	var from = self.rootDesc
	messageLen := len(r.Buf)
	if err := marshalTo(&r, w, from, to, opts,messageLen); err != nil {
		return nil, err
	}
	ret := make([]byte, len(w.Buf))
	copy(ret, w.Buf)
	binary.FreeBinaryProtocol(w)
	return ret, nil
}


func marshalTo(read *binary.BinaryProtocol, write *binary.BinaryProtocol, from *proto.MessageDescriptor, to *proto.MessageDescriptor, opts *Options, massageLen int) error {
	tail := read.Read + massageLen
	for read.Read < tail {
		fieldNumber, wireType, _, _ := read.ConsumeTag()
		fromField := (*from).Fields().ByNumber(fieldNumber)

		if fromField == nil {
			if opts.DisallowUnknow {
				return wrapError(meta.ErrUnknownField, fmt.Sprintf("unknown field %d", fieldNumber), nil)
			} else {
				// if not set, skip to the next field
				if err := read.Skip(wireType, opts.UseNativeSkip); err != nil {
					return wrapError(meta.ErrRead, "", err)
				}
				continue
			}
		}

		toField := (*to).Fields().ByNumber(fieldNumber)

		if toField == nil {
			// if not set, skip to the next field
			if err := read.Skip(wireType, opts.UseNativeSkip); err != nil {
				return wrapError(meta.ErrRead, "", err)
			}
			continue
		}

		fromType := fromField.Kind()
		toType := toField.Kind()
		if fromType != toType {
			return meta.NewError(meta.ErrDismatchType, "to descriptor dismatches from descriptor", nil)
		}

		if fromType == proto.MessageKind {
			fromDesc := fromField.Message()
			toDesc := toField.Message()
			write.AppendTag(fieldNumber, wireType)
			var pos int
			subMessageLen, err := read.ReadLength()
			if err != nil {
				return wrapError(meta.ErrRead, "", err)
			}
			write.Buf, pos = binary.AppendSpeculativeLength(write.Buf)
			marshalTo(read, write, &fromDesc, &toDesc, opts,subMessageLen)
			write.Buf = binary.FinishSpeculativeLength(write.Buf, pos)
		} else{
			start := read.Read
			if err := read.Skip(wireType, opts.UseNativeSkip); err != nil {
				return wrapError(meta.ErrRead, "", err)
			}
			end := read.Read
			value := read.Buf[start:end]

			write.AppendTag(fieldNumber, wireType)
			write.Buf = append(write.Buf, value...)
		}
	}
	return nil
}


func (self Value) GetByStr(key string) (v Value) {
	if err := self.should("Get", proto.MAP); err != "" {
		return errValue(meta.ErrUnsupportedType, err, nil)
	}

	if self.kt != proto.STRING {
		return errValue(meta.ErrUnsupportedType, "key type is not string", nil)
	}

	it := self.iterPairs()
	if it.Err != nil {
		return errValue(meta.ErrRead, "", it.Err)
	}

	for it.HasNext() {
		_, s, ss, e := it.NextStr(UseNativeSkipForGet)
		if it.Err != nil {
			v = errValue(meta.ErrRead, "", it.Err)
			goto ret
		}
		
		if s == key {
			vd := (*self.Desc).MapValue()
			v = self.slice(ss, e, &vd)
			goto ret
		}
	}
	v = errValue(meta.ErrNotFound, "", nil)
ret:
	return
}

func (self Value) GetByInt(key int) (v Value) {
	if err := self.should("Get", proto.MAP); err != "" {
		return errValue(meta.ErrUnsupportedType, err, nil)
	}

	if !self.kt.IsInt() {
		return errValue(meta.ErrUnsupportedType, "key type is not int", nil)
	}

	it := self.iterPairs()
	if it.Err != nil {
		return errValue(meta.ErrRead, "", it.Err)
	}

	for it.HasNext() {
		_, i, ss, e := it.NextInt(UseNativeSkipForGet)
		if it.Err != nil {
			v = errValue(meta.ErrRead, "", it.Err)
			goto ret
		}
		if i == key {
			vd := (*self.Desc).MapValue()
			v = self.slice(ss, e, &vd)
			goto ret
		}
	}
	v = errValue(meta.ErrNotFound, fmt.Sprintf("key '%d' is not found in this value", key), nil)
ret:
	return
}

// Index returns a sub node at the given index from a LIST value.
func (self Value) Index(i int) (v Value) {
	if err := self.should("Index", proto.LIST); err != "" {
		return errValue(meta.ErrUnsupportedType, err, nil)
	}

	var s, e int
	dataLen := 0 // when not packed, start need to contation length, like STRING and MESSAGE type
	it := self.iterElems()
	if it.Err != nil {
		return errValue(meta.ErrRead, "", it.Err)
	}

	if (*self.Desc).IsPacked() {
		if _, err := it.p.ReadLength(); err != nil {
			return errValue(meta.ErrRead, "", err)
		}

		for j := 0; it.HasNext() && j < i; j++ {
			it.Next(UseNativeSkipForGet)
		}
	} else {
		for j := 0; it.HasNext() && j < i; j++ {
			it.Next(UseNativeSkipForGet)
			if it.Err != nil {
				return errValue(meta.ErrRead, "", it.Err)
			}

			if it.HasNext() {
				if _,_,_,err := it.p.ConsumeTag(); err!=nil {
					return errValue(meta.ErrRead, "", err)
				}
			}
		}
	}

	if it.Err != nil {
		return errValue(meta.ErrRead, "", it.Err)
	}

	if i > it.k {
		return errValue(meta.ErrInvalidParam, fmt.Sprintf("index '%d' is out of range", i), nil)
	}

	s, e = it.Next(UseNativeSkipForGet)
	
	v = wrapValue(self.Node.slice(s-dataLen, e, self.et, 0, 0), self.Desc)

	return
}

// FieldByName returns a sub node at the given field name from a STRUCT value.
func (self Value) FieldByName(name string) (v Value) {
	if err := self.should("FieldByName", proto.MESSAGE); err != "" {
		return errValue(meta.ErrUnsupportedType, err, nil)
	}

	var f proto.FieldDescriptor
	if self.rootDesc != nil {
		f = (*self.rootDesc).Fields().ByName(proto.FieldName(name))
	} else {
		f = (*self.Desc).Message().Fields().ByName(proto.FieldName(name))
	}

	if f == nil {
		return errValue(meta.ErrUnknownField, fmt.Sprintf("field '%s' is not defined in IDL", name), nil)
	}

	// not found, try to scan the whole bytes
	it := self.iterFields()
	if it.Err != nil {
		return errValue(meta.ErrRead, "", it.Err)
	}
	for it.HasNext() {
		i, wt, s, e, tagPos := it.Next(UseNativeSkipForGet)
		if i == f.Number() {
			if f.IsMap() {
				it.p.Read = tagPos
				if _, err := it.p.ReadMap(&f, false, false, false); err != nil {
					return errValue(meta.ErrRead, "", err)
				}
				v = self.slice(s, it.p.Read, &f)
				goto ret
			}

			t := proto.Kind2Wire[f.Kind()]
			if wt != t {
				v = errValue(meta.ErrDismatchType, fmt.Sprintf("field '%s' expects type %s, buf got type %s", string(f.Name()), t, wt), nil)
				goto ret
			}
			v = self.slice(s, e, &f)
			goto ret
		} else if it.Err != nil {
			v = errValue(meta.ErrRead, "", it.Err)
			goto ret
		}
	}

	v = errValue(meta.ErrNotFound, fmt.Sprintf("field '%s' is not found in this value", name), errNotFound)
ret:
	// it.Recycle()
	return
}

// Field returns a sub node at the given field id from a STRUCT value.
func (self Value) Field(id proto.FieldNumber) (v Value) {
	if err := self.should("Field", proto.MESSAGE); err != "" {
		return errValue(meta.ErrUnsupportedType, err, nil)
	}

	it := self.iterFields()

	var f proto.FieldDescriptor
	if self.rootDesc != nil {
		f = (*self.rootDesc).Fields().ByNumber(id)
	} else {
		f = (*self.Desc).Message().Fields().ByNumber(id)
		if _, err := it.p.ReadLength(); err != nil {
			return errValue(meta.ErrRead, "", err)
		}
	}

	if f == nil {
		return errValue(meta.ErrUnknownField, fmt.Sprintf("field '%d' is not defined in IDL", id), nil)
	}
	// not found, try to scan the whole bytes
	
	if it.Err != nil {
		return errValue(meta.ErrRead, "", it.Err)
	}
	

	for it.HasNext() {
		i, wt, s, e, tagPos := it.Next(UseNativeSkipForGet)
		if i == f.Number() {
			if f.IsMap() || f.IsList() {
				it.p.Read = tagPos
				if f.IsMap() {
					if _, err := it.p.ReadMap(&f, false, false, false); err != nil {
						return errValue(meta.ErrRead, "", err)
					}
				} else {
					if _,err := it.p.ReadList(&f, false, false, false); err != nil {
						return errValue(meta.ErrRead, "", err)
					}
				}
				s = tagPos
				e = it.p.Read

				v = self.slice(s, e, &f)
				goto ret
			}

			t := proto.Kind2Wire[f.Kind()]
			if wt != t {
				v = errValue(meta.ErrDismatchType, fmt.Sprintf("field '%s' expects type %s, buf got type %s", f.Name(), t, wt), nil)
				goto ret
			}
			v = self.slice(s, e, &f)
			goto ret
		} else if it.Err != nil {
			v = errValue(meta.ErrRead, "", it.Err)
			goto ret
		}
	}

	v = errValue(meta.ErrNotFound, fmt.Sprintf("field '%d' is not found in this value", id), errNotFound)
ret:
	return
}