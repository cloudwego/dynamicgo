package generic

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
	"github.com/cloudwego/dynamicgo/proto/protowire"
)

type Value struct {
	Node
	rootDesc *proto.MessageDescriptor
	Desc     *proto.FieldDescriptor
}

var pnsPool = sync.Pool{
	New: func() interface{} {
		return &pnSlice{
			a: make([]PathNode, 0, DefaultNodeSliceCap),
		}
	},
}

func NewRootValue(desc *proto.MessageDescriptor, src []byte) Value {
	return Value{
		Node:     NewNode(proto.MESSAGE, src),
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
func (self Value) sliceWithDesc(s int, e int, desc *proto.FieldDescriptor) Value {
	t := proto.FromProtoKindToType((*desc).Kind(), (*desc).IsList(), (*desc).IsMap())
	ret := Value{
		Node: Node{
			t: t,
			l: (e - s),
			v: rt.AddPtr(self.v, uintptr(s)),
		},
		Desc: desc,
	}
	if t == proto.LIST {
		ret.et = proto.FromProtoKindToType((*desc).Kind(), false, false) // hard code, may have error?
	} else if t == proto.MAP {
		mapkey := (*desc).MapKey()
		mapvalue := (*desc).MapValue()
		ret.kt = proto.FromProtoKindToType(mapkey.Kind(), mapkey.IsList(), mapkey.IsMap())
		ret.et = proto.FromProtoKindToType(mapvalue.Kind(), mapvalue.IsList(), mapvalue.IsMap())
	}
	return ret
}

func searchFieldId(p *binary.BinaryProtocol, id proto.FieldNumber, messageLen int) (int, error) {
	start := p.Read
	for p.Read < start+messageLen {
		fieldNumber, wireType, tagLen, err := p.ConsumeTagWithoutMove()
		if err != nil {
			return 0, err
		}

		if fieldNumber == id {
			return p.Read, nil
		}
		p.Read += tagLen

		if err := p.Skip(wireType, false); err != nil {
			return 0, errNode(meta.ErrRead, "searchFieldId: skip field error.", err)
		}
	}
	return p.Read, errNotFound
}

func searchIndex(p *binary.BinaryProtocol, idx int, elementWireType proto.WireType, isPacked bool, fieldNumber proto.Number) (int, error) {
	// packed list
	cnt := 0
	result := p.Read
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
				return 0, errNode(meta.ErrRead, "searchIndex: skip packed list element error.", err)
			}
			cnt++
		}
		result = p.Read
	} else {
		// normal Type : [tag][(length)][value][tag][(length)][value][tag][(length)][value]....
		for p.Read < len(p.Buf) && cnt < idx {
			// don't move p.Read and judge whether readList completely
			if err := p.Skip(elementWireType, false); err != nil {
				return 0, errNode(meta.ErrRead, "searchIndex: skip unpacked list element error.", err)
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
		return p.Read, errNotFound
	}

	return result, nil
}

func searchIntKey(p *binary.BinaryProtocol, key int, keyType proto.Type, mapFieldNumber proto.FieldNumber) (int, error) {
	exist := false
	start := p.Read
	// normal Type : [tag][(length)][value][tag][(length)][value][tag][(length)][value]....
	for p.Read < len(p.Buf) {
		if _, err := p.ReadLength(); err != nil {
			return 0, wrapError(meta.ErrRead, "searchIntKey: read pair length failed", nil)
		}

		if _, _, _, keyTagErr := p.ConsumeTag(); keyTagErr != nil {
			return 0, wrapError(meta.ErrRead, "searchIntKey: read key tag failed", nil)
		}

		k, err := p.ReadInt(keyType)
		if err != nil {
			return 0, wrapError(meta.ErrRead, "searchIntKey: can not read map key", nil)
		}

		if k == key {
			exist = true
			start = p.Read // then p.Read will point to value tag
			break
		}

		_, valueWireType, _, valueTagErr := p.ConsumeTag()
		if valueTagErr != nil {
			return 0, wrapError(meta.ErrRead, "searchIntKey: read value tag failed", nil)
		}

		// if key not match, skip value
		if err := p.Skip(valueWireType, false); err != nil {
			return 0, errNode(meta.ErrRead, "searchIntKey: searchIntKey: can not read value.", err)
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
		return p.Read, errNotFound
	}
	return start, nil
}

func searchStrKey(p *binary.BinaryProtocol, key string, keyType proto.Type, mapFieldNumber proto.FieldNumber) (int, error) {
	exist := false
	start := p.Read
	// normal Type : [tag][(length)][value][tag][(length)][value][tag][(length)][value]....
	for p.Read < len(p.Buf) {
		if _, err := p.ReadLength(); err != nil {
			return 0, wrapError(meta.ErrRead, "searchStrKey: read pair length failed", nil)
		}

		if _, _, _, keyTagErr := p.ConsumeTag(); keyTagErr != nil {
			return 0, wrapError(meta.ErrRead, "searchStrKey: read key tag failed", nil)
		}

		k, err := p.ReadString(false)
		if err != nil {
			return 0, wrapError(meta.ErrRead, "searchStrKey: can not read map key", nil)
		}

		if k == key {
			exist = true
			start = p.Read // then p.Read will point to value tag
			break
		}

		_, valueWireType, _, valueTagErr := p.ConsumeTag()
		if valueTagErr != nil {
			return 0, wrapError(meta.ErrRead, "searchStrKey: read value tag failed", nil)
		}

		// if key not match, skip value
		if err := p.Skip(valueWireType, false); err != nil {
			return 0, errNode(meta.ErrRead, "searchStrKey: searchStrKey: can not read value.", err)
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
		return p.Read, errNotFound
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
	size := 0
	isRoot := false
	if len(pathes) == 0 {
		return self, address
	}

	if self.Error() != "" {
		return self, address
	}

	p := binary.BinaryProtocol{
		Buf: self.raw(),
	}

	if self.rootDesc != nil {
		isRoot = true
	} else {
		desc = self.Desc
		if self.t == proto.LIST || self.t == proto.MAP {
			p.ConsumeTag()
		}
	}

	for i, path := range pathes {
		switch path.t {
		case PathFieldId:
			id := path.id()
			var fd proto.FieldDescriptor
			messageLen := 0
			if isRoot {
				fd = (*self.rootDesc).Fields().ByNumber(id)
				messageLen = len(p.Buf)
				isRoot = false
			} else {
				fd = (*desc).Message().Fields().ByNumber(id)
				Len, err := p.ReadLength()
				if err != nil {
					return errValue(meta.ErrRead, "GetByPath: read field length failed.", err), address
				}
				messageLen += Len
			}
			if fd != nil {
				desc = &fd
				tt = proto.FromProtoKindToType(fd.Kind(), fd.IsList(), fd.IsMap())
			}
			start, err = searchFieldId(&p, id, messageLen)
			if err == errNotFound {
				tt = proto.MESSAGE
			}
		case PathFieldName:
			name := proto.FieldName(path.str())
			var fd proto.FieldDescriptor
			messageLen := 0
			if isRoot {
				fd = (*self.rootDesc).Fields().ByName(name)
				messageLen = len(p.Buf)
				isRoot = false
			} else {
				fd = (*desc).Message().Fields().ByName(name)
				Len, err := p.ReadLength()
				if err != nil {
					return errValue(meta.ErrRead, "GetByPath: read field length failed.", err), address
				}
				messageLen += Len
			}
			if fd == nil {
				return errValue(meta.ErrUnknownField, fmt.Sprintf("field name '%s' is not defined in IDL", name), nil), address
			}
			tt = proto.FromProtoKindToType(fd.Kind(), fd.IsList(), fd.IsMap())
			desc = &fd
			start, err = searchFieldId(&p, fd.Number(), messageLen)
			if err == errNotFound {
				tt = proto.MESSAGE
			}
		case PathIndex:
			elemKind := (*desc).Kind()
			elementWireType := proto.Kind2Wire[elemKind]
			isPacked := (*desc).IsPacked()
			start, err = searchIndex(&p, path.int(), elementWireType, isPacked, (*desc).Number())
			tt = proto.FromProtoKindToType(elemKind, false, false)
			if err == errNotFound {
				tt = proto.LIST
			}
		case PathStrKey:
			mapFieldNumber := (*desc).Number()
			start, err = searchStrKey(&p, path.str(), proto.STRING, mapFieldNumber)
			tt = proto.FromProtoKindToType((*desc).MapValue().Kind(), false, false)
			valueDesc := (*desc).MapValue()
			desc = &valueDesc
			if err == errNotFound {
				tt = proto.MAP
			}
		case PathIntKey:
			keyType := proto.FromProtoKindToType((*desc).MapKey().Kind(), false, false)
			mapFieldNumber := (*desc).Number()
			start, err = searchIntKey(&p, path.int(), keyType, mapFieldNumber)
			tt = proto.FromProtoKindToType((*desc).MapValue().Kind(), false, false)
			valueDesc := (*desc).MapValue()
			desc = &valueDesc
			if err == errNotFound {
				tt = proto.MAP
			}
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
			return errValue(en.ErrCode().Behavior(), "invalid value node.", err), address
		}

		if i != len(pathes)-1 {
			if _, _, _, err := p.ConsumeTag(); err != nil {
				return errValue(meta.ErrRead, "invalid field tag failed.", err), address
			}
		}

	}

	switch tt {
	case proto.MAP:
		kt = proto.FromProtoKindToType((*desc).MapKey().Kind(), false, false)
		et = proto.FromProtoKindToType((*desc).MapValue().Kind(), false, false)
		if s, err := p.SkipAllElements((*desc).Number(),(*desc).IsPacked()); err != nil {
			en := err.(Node)
			return errValue(en.ErrCode().Behavior(), "invalid list node.", err), address
		} else {
			size = s
		}
	case proto.LIST:
		et = proto.FromProtoKindToType((*desc).Kind(), false, false)
		if s, err := p.SkipAllElements((*desc).Number(),(*desc).IsPacked()); err != nil {
			en := err.(Node)
			return errValue(en.ErrCode().Behavior(), "invalid list node.", err), address
		} else {
			size = s
		}
	default:
		// node condition: simple field element, message, packed list element, map value element
		skipType := proto.Kind2Wire[(*desc).Kind()]
		
		// only packed list element no tag to skip
		if (*desc).IsPacked() == false {
			if _, _, _, err := p.ConsumeTag(); err != nil {
				return errValue(meta.ErrRead, "invalid field tag.", err), address
			}
			start = p.Read
		}

		if err := p.Skip(skipType, false); err != nil {
			return errValue(meta.ErrRead, "skip field error.", err), address
		}
	}
	
	return wrapValue(self.Node.sliceComplex(start, p.Read, tt, kt, et, size), desc), address
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
		return false, wrapError(meta.ErrInvalidParam, "given node is invalid.", sub)
	}

	// search source node by path
	v, address := self.GetByPath(path...)
	if v.IsError() {
		if !v.isErrNotFoundLast() {
			return false, v
		}

		targetPath := path[l-1]
		desc, err := GetDescByPath(self.rootDesc, path[:l-1]...)
		var fd *proto.FieldDescriptor
		if err != nil {
			return false, err
		}
		if f, ok := (*desc).(proto.FieldDescriptor); ok {
			fd = &f
		}

		if targetPath.t == PathFieldName {
			if d, ok := (*desc).(proto.MessageDescriptor); ok {
				f := d.Fields().ByName(proto.FieldName(targetPath.str()))
				targetPath = NewPathFieldId(f.Number())
				fd = &f
			} else if d, ok := (*desc).(proto.FieldDescriptor); ok {
				f := d.Message().Fields().ByName(proto.FieldName(targetPath.str()))
				targetPath = NewPathFieldId(f.Number())
				fd = &f
			}
		}

		if err := v.setNotFound(targetPath, &sub.Node, fd); err != nil {
			return false, err
		}
	} else {
		exist = true
	}

	
	originLen := len(self.raw())
	err = self.replace(v.Node, sub.Node)
	isPacked := path[l-1].t == PathIndex && sub.Node.t != proto.MESSAGE && sub.Node.t != proto.STRING
	self.UpdateByteLen(originLen, address, isPacked, path...)
	return
}


func (self *Value) UpdateByteLen(originLen int, address []int, isPacked bool, path ...Path) {
	afterLen := self.l
	diffLen := afterLen - originLen
	previousType := proto.UNKNOWN

	for i := len(address) - 1; i >= 0; i-- {
		pathType := path[i].t
		addressPtr := address[i]
		if previousType == proto.MESSAGE || (previousType == proto.LIST && isPacked) {
			// tag
			buf := rt.BytesFrom(rt.AddPtr(self.v, uintptr(addressPtr)), self.l-addressPtr, self.l-addressPtr)
			_, tagOffset := protowire.ConsumeVarint(buf)
			// length
			length, lenOffset := protowire.ConsumeVarint(buf[tagOffset:])
			newLength := int(length) + diffLen
			newBytes := protowire.AppendVarint(nil, uint64(newLength))
			// length == 0 means had been deleted all the data in the field
			if newLength == 0 {
				newBytes = newBytes[:0]
			}

			subLen := len(newBytes) - lenOffset

			if subLen == 0 {
				// no need to change length
				copy(buf[tagOffset:tagOffset+lenOffset], newBytes)
				continue
			}

			// split length
			srcHead := rt.AddPtr(self.v, uintptr(addressPtr+tagOffset))
			if newLength == 0 {
				// delete tag
				srcHead = rt.AddPtr(self.v, uintptr(addressPtr))
				subLen -= tagOffset
			}

			srcTail := rt.AddPtr(self.v, uintptr(addressPtr+tagOffset+lenOffset))
			l0 := int(uintptr(srcHead) - uintptr(self.v))
			l1 := len(newBytes)
			l2 := int(uintptr(self.v) + uintptr(self.l) - uintptr(srcTail))

			// copy three slices into new buffer
			newBuf := make([]byte, l0+l1+l2)
			copy(newBuf[:l0], rt.BytesFrom(self.v, l0, l0))
			copy(newBuf[l0:l0+l1], newBytes)
			copy(newBuf[l0+l1:l0+l1+l2], rt.BytesFrom(srcTail, l2, l2))
			self.v = rt.GetBytePtr(newBuf)
			self.l = int(len(newBuf))
			if isPacked {
				isPacked = false
			}
			diffLen += subLen
		}

		if pathType == PathStrKey || pathType == PathIntKey {
			previousType = proto.MAP
		} else if pathType == PathIndex {
			previousType = proto.LIST
		} else {
			previousType = proto.MESSAGE
		}
	}
}

 
func (self *Value) findDeleteChild(path Path) (Node, int) {
	p := binary.BinaryProtocol{}
	p.Buf = self.raw()
	tt := self.t // in fact, no need to judge which type
	valueLen := self.l
	exist := false
	var start, end int

	switch self.t {
	case proto.MESSAGE:
		if self.rootDesc == nil {
			if l, err := p.ReadLength(); err != nil {
				return errNode(meta.ErrRead, "", err), -1
			} else {
				valueLen = l
			}
		}
		start = valueLen
		end = 0
		// previous has change PathFieldName to PathFieldId
		if path.Type() != PathFieldId {
			return errNode(meta.ErrDismatchType, "path type is not PathFieldId", nil), -1
		}
		messageStart := p.Read
		id := path.id()
		for p.Read < messageStart+valueLen {
			fieldStart := p.Read
			fieldNumber, wireType, _, tagErr := p.ConsumeTag()
			if tagErr != nil {
				return errNode(meta.ErrRead, "", tagErr), -1
			}
			if err := p.Skip(wireType, false); err != nil {
				return errNode(meta.ErrRead, "", err), -1
			}
			fieldEnd := p.Read
			if id == fieldNumber {
				exist = true
				if fieldStart < start {
					start = fieldStart
				}
				if fieldEnd > end {
					end = fieldEnd
				}
			}
		}
		if !exist {
			return errNotFound, -1
		}
	case proto.LIST:
		listIndex := 0
		if path.Type() != PathIndex {
			return errNode(meta.ErrDismatchType, "path type is not PathIndex", nil), -1
		}
		idx := path.int()
		fieldNumber := (*self.Desc).Number()
		elemType := proto.Kind2Wire[(*self.Desc).Kind()]
		size, err := self.Len()
		if err != nil {
			return errNode(meta.ErrRead, "", err), -1
		}
		
		if idx >= size{
			return errNotFound, -1
		}
		// packed : [tag][l][(l)v][(l)v][(l)v][(l)v].....
		if (*self.Desc).IsPacked() {
			if _, _, _, tagErr := p.ConsumeTag(); tagErr != nil {
				return errNode(meta.ErrRead, "", tagErr), -1
			}

			if _, lengthErr := p.ReadLength(); lengthErr != nil {
				return errNode(meta.ErrRead, "", lengthErr), -1
			}

			for p.Read < valueLen && listIndex <= idx {
				start = p.Read
				if err := p.Skip(elemType, false); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}
				end = p.Read
				listIndex++
			}
		} else {
			// unpacked : [tag][l][v][tag][l][v][tag][l][v][tag][l][v]....
			for p.Read < valueLen && listIndex <= idx {
				start = p.Read
				itemNumber, _, tagLen, _ := p.ConsumeTagWithoutMove()
				if itemNumber != fieldNumber {
					break
				}
				p.Read += tagLen
				if err := p.Skip(elemType, false); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}
				end = p.Read
				listIndex++
			}
		}
	case proto.MAP:
		pairNumber := (*self.Desc).Number()
		if self.kt == proto.STRING {
			key := path.str()
			for p.Read < valueLen {
				start = p.Read
				itemNumber, _, pairLen, _ := p.ConsumeTagWithoutMove()
				if itemNumber != pairNumber {
					break
				}
				p.Read += pairLen
				if _, err := p.ReadLength(); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}
				// key
				if _, _, _, err := p.ConsumeTag(); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}

				k, err := p.ReadString(false)
				if err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}

				// value
				_, valueWire, _, _ := p.ConsumeTag()
				if err := p.Skip(valueWire, false); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}
				end = p.Read
				if k == key {
					exist = true
					break
				}
			}
		} else if self.kt.IsInt() {
			key := path.Int()
			pairNumber, _, _, _ := p.ConsumeTagWithoutMove()
			for p.Read < valueLen {
				start = p.Read
				itemNumber, _, pairLen, _ := p.ConsumeTagWithoutMove()
				if itemNumber != pairNumber {
					break
				}
				p.Read += pairLen
				if _, err := p.ReadLength(); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}

				// key
				if _, _, _, err := p.ConsumeTag(); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}

				k, err := p.ReadInt(self.kt)
				if err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}

				//value
				_, valueWire, _, _ := p.ConsumeTag()
				if err := p.Skip(valueWire, false); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}
				end = p.Read
				if k == key {
					exist = true
					break
				}
			}
		}
		if !exist {
			return errNotFound, -1
		}
	default:
		return errNotFound, -1
	}
	return Node{
		t: tt,
		v: rt.AddPtr(self.v, uintptr(start)),
		l: end - start,
	}, start
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

	// search target node by path
	var parentValue, address = self.GetByPath(path[:l-1]...)
	if parentValue.IsError() {
		if parentValue.IsErrNotFound() {
			print(address)
			return nil
		}
		return parentValue
	}

	isPacked := false
	p := path[l-1]
	var desc *proto.Descriptor
	var err error
	desc, err = GetDescByPath(self.rootDesc, path...)
	if err != nil {
		return err
	}

	if p.t == PathFieldName {
		if d, ok := (*desc).(proto.FieldDescriptor); ok {
			p = NewPathFieldId(d.Number())
		}
	} else if p.t == PathIndex {
		if d, ok := (*desc).(proto.FieldDescriptor); ok {
			isPacked = d.IsPacked()
		}
	}

	ret, position := parentValue.findDeleteChild(p)
	if ret.IsError() {
		return ret
	}
	
	originLen := len(self.raw())
	if err := self.replace(ret, Node{t: ret.t}); err != nil {
		return errValue(meta.ErrWrite, "replace node by empty node failed", err)
	}
	address = append(address, position) // must add one address align with path length
	self.UpdateByteLen(originLen, address, isPacked, path...)
	return nil
}

// MarshalTo marshals self value into a sub value descripted by the to descriptor, alse called as "Cutting".
// Usually, the to descriptor is a subset of self descriptor.
func (self Value) MarshalTo(to *proto.MessageDescriptor, opts *Options) ([]byte, error) {
	var w = binary.NewBinaryProtocolBuffer()
	var r = binary.BinaryProtocol{}
	r.Buf = self.raw()
	var from = self.rootDesc
	messageLen := len(r.Buf)
	if err := marshalTo(&r, w, from, to, opts, messageLen); err != nil {
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
			return wrapError(meta.ErrDismatchType, "to descriptor dismatches from descriptor", nil)
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
			marshalTo(read, write, &fromDesc, &toDesc, opts, subMessageLen)
			write.Buf = binary.FinishSpeculativeLength(write.Buf, pos)
		} else {
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
	n := self.Node.GetByStr(key)
	vd := (*self.Desc).MapValue()
	if n.IsError() {
		return wrapValue(n, nil)
	}
	return wrapValue(n, &vd)
}

func (self Value) GetByInt(key int) (v Value) {
	n := self.Node.GetByInt(key)
	vd := (*self.Desc).MapValue()
	if n.IsError() {
		return wrapValue(n, nil)
	}
	return wrapValue(n, &vd)
}

// Index returns a sub node at the given index from a LIST value.
func (self Value) Index(i int) (v Value) {
	n := self.Node.Index(i)
	if n.IsError() {
		return wrapValue(n, nil)
	}
	v = wrapValue(n, self.Desc)
	return
}

// FieldByName returns a sub node at the given field name from a STRUCT value.
func (self Value) FieldByName(name string) (v Value) {
	if err := self.should("FieldByName", proto.MESSAGE); err != "" {
		return errValue(meta.ErrUnsupportedType, err, nil)
	}

	it := self.iterFields()

	var f proto.FieldDescriptor
	if self.rootDesc != nil {
		f = (*self.rootDesc).Fields().ByName(proto.FieldName(name))
	} else {
		f = (*self.Desc).Message().Fields().ByName(proto.FieldName(name))
		if _, err := it.p.ReadLength(); err != nil {
			return errValue(meta.ErrRead, "", err)
		}
	}

	if f == nil {
		return errValue(meta.ErrUnknownField, fmt.Sprintf("field '%s' is not defined in IDL", name), nil)
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
				if _, err := it.p.SkipAllElements(i,f.IsPacked()); err != nil {
					return errValue(meta.ErrRead, "SkipAllElements in LIST/MAP failed", err)
				}
				s = tagPos
				e = it.p.Read

				v = self.sliceWithDesc(s, e, &f)
				goto ret
			}

			t := proto.Kind2Wire[f.Kind()]
			if wt != t {
				v = errValue(meta.ErrDismatchType, fmt.Sprintf("field '%s' expects type %s, buf got type %s", string(f.Name()), t, wt), nil)
				goto ret
			}
			v = self.sliceWithDesc(s, e, &f)
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
				if _, err := it.p.SkipAllElements(i,f.IsPacked()); err != nil {
					return errValue(meta.ErrRead, "SkipAllElements in LIST/MAP failed", err)
				}
				s = tagPos
				e = it.p.Read

				v = self.sliceWithDesc(s, e, &f)
				goto ret
			}

			t := proto.Kind2Wire[f.Kind()]
			if wt != t {
				v = errValue(meta.ErrDismatchType, fmt.Sprintf("field '%s' expects type %s, buf got type %s", f.Name(), t, wt), nil)
				goto ret
			}
			v = self.sliceWithDesc(s, e, &f)
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


// GetMany searches transversely and returns all the sub nodes along with the given pathes.
func (self Value) GetMany(pathes []PathNode, opts *Options) error {
	if len(pathes) == 0 {
		return nil
	}
	return self.getMany(pathes, opts.ClearDirtyValues, opts)
}

func (self Value) getMany(pathes []PathNode, clearDirty bool, opts *Options) error {
	if clearDirty {
		for i := range pathes {
			pathes[i].Node = Node{}
		}
	}
	p := pathes[0]
	switch p.Path.t {
	case PathFieldId:
		return self.Fields(pathes, opts)
	case PathIndex:
		return self.Indexes(pathes, opts)
	case PathStrKey, PathIntKey, PathBinKey:
		return self.Gets(pathes, opts)
	default:
		return errValue(meta.ErrUnsupportedType, fmt.Sprintf("invalid path: %#v", p), nil)
	}
}

func (self Value) Fields(ids []PathNode, opts *Options) error {
	if err := self.should("Fields", proto.MESSAGE); err != "" {
		return errValue(meta.ErrUnsupportedType, err, nil)
	}

	if len(ids) == 0 {
		return nil
	}

	if opts.ClearDirtyValues {
		for i := range ids {
			ids[i].Node = Node{}
		}
	}

	it := self.iterFields()
	if it.Err != nil {
		return errValue(meta.ErrRead, "", it.Err)
	}

	var Fields proto.FieldDescriptors
	if self.rootDesc != nil {
		Fields = (*self.rootDesc).Fields()
	} else {
		Fields = (*self.Desc).Message().Fields()
		if _, err := it.p.ReadLength(); err != nil {
			return errValue(meta.ErrRead, "", err)
		}
	}

	need := len(ids)
	for count := 0; it.HasNext() && count < need; {
		var p *PathNode
		i, _, s, e, tagPos := it.Next(UseNativeSkipForGet)
		if it.Err != nil {
			return errValue(meta.ErrRead, "", it.Err)
		}
		f := Fields.ByNumber(i)
		if f.IsMap() || f.IsList() {
			it.p.Read = tagPos
			if _, err := it.p.SkipAllElements(i,f.IsPacked()); err != nil {
				return errValue(meta.ErrRead, "SkipAllElements in LIST/MAP failed", err)
			}
			s = tagPos
			e = it.p.Read
		}


		v := self.sliceWithDesc(s, e, &f)

		//TODO: use bitmap to avoid repeatedly scan
		for j, id := range ids {
			if id.Path.t == PathFieldId && id.Path.id() == i {
				p = &ids[j]
				count += 1
				break
			}
		}

		if p != nil {
			p.Node = v.Node
		}
	}

	return nil
}

func (self Value) Indexes(ins []PathNode, opts *Options) error {
	if err := self.should("Indexes", proto.LIST); err != "" {
		return errValue(meta.ErrUnsupportedType, err, nil)
	}

	if len(ins) == 0 {
		return nil
	}

	if opts.ClearDirtyValues {
		for i := range ins {
			ins[i].Node = Node{}
		}
	}

	it := self.iterElems()
	if it.Err != nil {
		return errValue(meta.ErrRead, "", it.Err)
	}

	need := len(ins)
	IsPacked := (*self.Desc).IsPacked()

	if IsPacked {
		if _, err := it.p.ReadLength(); err != nil {
			return errValue(meta.ErrRead, "", err)
		}
	}
	
	for count, i := 0, 0; it.HasNext() && count < need; i++ {
		s, e := it.Next(UseNativeSkipForGet)
		if it.Err != nil {
			return errValue(meta.ErrRead, "", it.Err)
		}

		var p *PathNode
		for j, id := range ins {
			if id.Path.t != PathIndex {
				continue
			}
			k := id.Path.int()
			if k >= it.size {
				continue
			}
			if k == i {
				p = &ins[j]
				count += 1
				break
			}
		}
		// unpacked mode, skip tag
		if IsPacked == false && it.HasNext() {
			if _, _, _, err := it.p.ConsumeTag(); err != nil {
				return errValue(meta.ErrRead, "", err)
			}
		}
		if p != nil {
			p.Node = self.Node.slice(s, e, self.et)
		}
	}
	return nil
}

func (self Value) Gets(keys []PathNode, opts *Options) error {
	if err := self.should("Gets", proto.MAP); err != "" {
		return errValue(meta.ErrUnsupportedType, err, nil)
	}

	if len(keys) == 0 {
		return nil
	}

	if opts.ClearDirtyValues {
		for i := range keys {
			keys[i].Node = Node{}
		}
	}

	et := self.et
	it := self.iterPairs()
	if it.Err != nil {
		return errValue(meta.ErrRead, "", it.Err)
	}

	need := len(keys)
	for count := 0; it.HasNext() && count < need; {
		for j, id := range keys {
			if id.Path.Type() == PathStrKey {
				exp := id.Path.str()
				_, key, s, e := it.NextStr(UseNativeSkipForGet)
				if it.Err != nil {
					return errValue(meta.ErrRead, "", it.Err)
				}
				if key == exp {
					keys[j].Node = self.Node.slice(s, e, et)
					count += 1
					break
				}
			} else if id.Path.Type() == PathIntKey {
				exp := id.Path.int()
				_, key, s, e := it.NextInt(UseNativeSkipForGet)
				if it.Err != nil {
					return errValue(meta.ErrRead, "", it.Err)
				}
				if key == exp {
					keys[j].Node = self.Node.slice(s, e, et)
					count += 1
					break
				}
			}
		}
	}
	return nil
}


func (self *Value) SetMany(pathes []PathNode, opts *Options, root *Value, address []int, path ...Path) (err error) {
	if len(pathes) == 0 {
		return nil
	}
	
	if err := self.Check(); err != nil {
		return err
	}

	// copy original pathes
	ps := pnsPool.Get().(*pnSlice)
	if cap(ps.a) < len(pathes) {
		ps.a = make([]PathNode, len(pathes))
	} else {
		ps.a = ps.a[:len(pathes)]
	}
	copy(ps.a, pathes)
	ps.b = pathes
	isPacked := false
	originLen := len(self.raw()) // current buf length
	rootLen := len(root.raw()) // root buf length
	
	// get original values
	if err = self.getMany(ps.a, true, opts); err != nil {
		goto ret
	}

	// handle not found values
	for i, a := range ps.a {
		if a.IsUnKnown() {
			var sp unsafe.Pointer
			if self.t.IsComplex() {
				sp = rt.AddPtr(self.v, uintptr(self.l))
			}
			ps.a[i].Node = errNotFoundLast(sp, self.t)
			ps.a[i].Node.setNotFound(a.Path, &ps.b[i].Node, self.Desc)
			if self.t == proto.LIST || self.t == proto.MAP{
				self.size += 1
			}
		}
	}

	if self.t == proto.LIST {
		isPacked = (*self.Desc).IsPacked()
	}

	// update current Node length
	if self.rootDesc == nil {
		err = self.replaceMany(ps)
		if self.t == proto.LIST && isPacked {
			currentAdd := []int{0, -1}
			currentPath := []Path{NewPathIndex(-1), NewPathIndex(-1)}
			self.UpdateByteLen(originLen,currentAdd,isPacked,currentPath...)
		} else if self.t == proto.MESSAGE {
			buf := self.raw()
			_, lenOffset := protowire.ConsumeVarint(buf)
			byteLen := len(buf) - lenOffset
			newLen := protowire.AppendVarint(nil, uint64(byteLen))
			// newLen + buf[lenOffset:]
			l0 := len(newLen)
			l1 := byteLen
			newBuf := make([]byte, l0+l1)
			copy(newBuf[:l0], newLen)
			copy(newBuf[l0:], buf[lenOffset:])
			self.v = rt.GetBytePtr(newBuf)
			self.l = int(len(newBuf))
		}
	}
	

	// update root length
	err = root.replaceMany(ps)
	root.UpdateByteLen(rootLen, address, isPacked, path...)
ret:
	ps.b = nil
	pnsPool.Put(ps)
	return
}
