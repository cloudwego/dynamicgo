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

const defaultBytesSize = 64

type Value struct {
	Node
	Desc *proto.TypeDescriptor
	IsRoot  bool
}

var (
	pnsPool = sync.Pool{
		New: func() interface{} {
			return &pnSlice{
				a: make([]PathNode, 0, DefaultNodeSliceCap),
			}
		},
	}
	bytesPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, defaultBytesSize)
		},
	}
)

func NewBytesFromPool() []byte {
	return bytesPool.Get().([]byte)
}

func FreeBytesToPool(b []byte) {
	b = b[:0]
	bytesPool.Put(b)
}

func NewRootValue(desc *proto.TypeDescriptor, src []byte) Value {
	return Value{
		Node: NewNode(proto.MESSAGE, src),
		Desc: desc,
		IsRoot: true,
	}
}

// only for basic Node
func NewValue(desc *proto.TypeDescriptor, src []byte) Value {
	return Value{
		Node: NewNode(desc.Type(), src),
		Desc: desc,
	}
}

// only for LIST/MAP parent Node
func NewComplexValue(desc *proto.TypeDescriptor, src []byte) Value {
	t := desc.Type()
	et := proto.UNKNOWN
	kt := proto.UNKNOWN
	if t == proto.LIST {
		et = desc.Elem().Type()
	} else if t == proto.MAP {
		t = proto.MAP
		et = desc.Elem().Type()
		kt = desc.Key().Type()
	} else {
		panic("only used for list/map node")
	}

	return Value{
		Node: NewComplexNode(t, et, kt, src),
		Desc: desc,
	}
}


// NewValueFromNode copy both Node and TypeDescriptor from another Value.
func (self Value) Fork() Value {
	ret := self
	ret.Node = self.Node.Fork()
	return ret
}


func (self Value) sliceWithDesc(s int, e int, desc *proto.TypeDescriptor) Value {
	node := self.sliceNodeWithDesc(s, e, desc)
	return Value{Node: node, Desc: desc}
}

// searchFieldId in MESSAGE Node
// if id is found, return the field tag position, otherwise return the end of p.Buf
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

// searchIndex in LIST Node
// packed: if idx is found, return the element[V] value start position, otherwise return the end of p.Buf
// unpacked: if idx is found, return the element[TLV] tag position, otherwise return the end of p.Buf
func searchIndex(p *binary.BinaryProtocol, idx int, elementWireType proto.WireType, isPacked bool, fieldNumber proto.FieldNumber) (int, error) {
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

// searchIntKey in MAP Node
// if key is found, return the value tag position, otherwise return the end of p.Buf
func searchIntKey(p *binary.BinaryProtocol, key int, keyType proto.Type, mapFieldNumber proto.FieldNumber) (int, error) {
	exist := false
	start := p.Read
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
			start = p.Read // p.Read will point to value tag
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

// searchStrKey in MAP Node
// if key is found, return the value tag position, otherwise return the end of p.Buf
func searchStrKey(p *binary.BinaryProtocol, key string, keyType proto.Type, mapFieldNumber proto.FieldNumber) (int, error) {
	exist := false
	start := p.Read

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
			start = p.Read // p.Read will point to value tag
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

func (self Value) GetByPath(pathes ...Path) Value {
	value, _ := self.getByPath(pathes...)
	return value
}

func (self Value) GetByPathWithAddress(pathes ...Path) (Value, []int) {
	return self.getByPath(pathes...)
}

// inner use
func (self Value) getByPath(pathes ...Path) (Value, []int) {
	address := make([]int, len(pathes))
	start := 0
	var err error
	tt := self.t
	kt := self.kt
	et := self.et
	size := 0
	desc := self.Desc
	isRoot := self.IsRoot
	if len(pathes) == 0 {
		return self, address
	}

	if self.Error() != "" {
		return self, address
	}

	p := binary.BinaryProtocol{
		Buf: self.raw(),
	}

	if !isRoot {
		if self.t == proto.LIST || self.t == proto.MAP {
			p.ConsumeTag()
		}
	}

	for i, path := range pathes {
		switch path.t {
		case PathFieldId:
			id := path.id()
			messageLen := 0
			if isRoot {
				messageLen = len(p.Buf)
				isRoot = false
			} else {
				Len, err := p.ReadLength()
				if err != nil {
					return errValue(meta.ErrRead, "GetByPath: read field length failed.", err), address
				}
				messageLen += Len
			}

			fd := desc.Message().ByNumber(id)
			if fd != nil {
				desc = fd.Type()
				tt = desc.Type()
			}

			start, err = searchFieldId(&p, id, messageLen)
			if err == errNotFound {
				tt = proto.MESSAGE
			}
		case PathFieldName:
			name := path.str()
			messageLen := 0
			if isRoot {
				messageLen = len(p.Buf)
				isRoot = false
			} else {
				Len, err := p.ReadLength()
				if err != nil {
					return errValue(meta.ErrRead, "GetByPath: read field length failed.", err), address
				}
				messageLen += Len
			}

			fd := desc.Message().ByName(name)
			if fd == nil {
				return errValue(meta.ErrUnknownField, fmt.Sprintf("field name '%s' is not defined in IDL", name), nil), address
			}

			desc = fd.Type()
			tt = desc.Type()
			start, err = searchFieldId(&p, fd.Number(), messageLen)
			if err == errNotFound {
				tt = proto.MESSAGE
			}
		case PathIndex:
			elementWireType := desc.Elem().WireType()
			isPacked := desc.IsPacked()
			start, err = searchIndex(&p, path.int(), elementWireType, isPacked, desc.BaseId())
			tt = desc.Elem().Type()
			if err == errNotFound {
				tt = proto.LIST
			}
		case PathStrKey:
			mapFieldNumber := desc.BaseId()
			start, err = searchStrKey(&p, path.str(), proto.STRING, mapFieldNumber)
			tt = desc.Elem().Type()
			valueDesc := desc.Elem()
			desc = valueDesc
			if err == errNotFound {
				tt = proto.MAP
			}
		case PathIntKey:
			keyType := desc.Key().Type()
			mapFieldNumber := desc.BaseId()
			start, err = searchIntKey(&p, path.int(), keyType, mapFieldNumber)
			tt = desc.Elem().Type()
			valueDesc := desc.Elem()
			desc = valueDesc
			if err == errNotFound {
				tt = proto.MAP
			}
		default:
			return errValue(meta.ErrUnsupportedType, fmt.Sprintf("invalid %dth path: %#v", i, p), nil), address
		}
		// after search function, p.Read will always point to the tag position except when packed list element
		address[i] = start

		if err != nil {
			// the last one not foud, return start pointer for subsequently inserting operation on `SetByPath()`
			if i == len(pathes)-1 && err == errNotFound {
				return Value{errNotFoundLast(unsafe.Pointer(uintptr(self.v)+uintptr(start)), tt), nil, false}, address
			}
			en := err.(Node)
			return errValue(en.ErrCode().Behavior(), "invalid value node.", err), address
		}
		// if not the last one, it must be a complex node, so need to skip tag
		if i != len(pathes)-1 {
			if _, _, _, err := p.ConsumeTag(); err != nil {
				return errValue(meta.ErrRead, "invalid field tag failed.", err), address
			}
		}

	}

	// check the result node type to get the right slice bytes
	switch tt {
	case proto.MAP:
		kt = desc.Key().Type()
		et = desc.Elem().Type()
		if s, err := p.SkipAllElements(desc.BaseId(), desc.IsPacked()); err != nil {
			en := err.(Node)
			return errValue(en.ErrCode().Behavior(), "invalid map node.", err), address
		} else {
			size = s
		}
	case proto.LIST:
		et = desc.Elem().Type()
		if s, err := p.SkipAllElements(desc.BaseId(), desc.IsPacked()); err != nil {
			en := err.(Node)
			return errValue(en.ErrCode().Behavior(), "invalid list node.", err), address
		} else {
			size = s
		}
	default:
		// node condition: field element, list element, map value element
		var skipType proto.WireType
		
		// only packed list element no tag to skip
		if desc.IsPacked() == false {
			if _, _, _, err := p.ConsumeTag(); err != nil {
				return errValue(meta.ErrRead, "invalid field tag.", err), address
			}
			start = p.Read
		}
		
		if desc.Type() != proto.LIST{
			skipType = desc.WireType()
		} else {
			skipType = desc.Elem().WireType()
			desc = desc.Elem()
		}

		if err := p.Skip(skipType, false); err != nil {
			return errValue(meta.ErrRead, "skip field error.", err), address
		}
	}

	return wrapValue(self.Node.sliceComplex(start, p.Read, tt, kt, et, size), desc), address
}

// SetByPath searches longitudinally and sets a sub value at the given path from the value.
// exist tells whether the node is already exists.
func (self *Value) SetByPath(sub Node, path ...Path) (exist bool, err error) {
	l := len(path)
	if l == 0 {
		if self.t != sub.t {
			return false, errValue(meta.ErrDismatchType, "self node type is not equal to sub node type.", nil)
		}
		self.Node = sub // it means replace the root value ?
		return true, nil
	}

	if err := self.Check(); err != nil {
		return false, err
	}

	if self.Error() != "" {
		return false, wrapError(meta.ErrInvalidParam, "given node is invalid.", sub)
	}

	// search source node by path
	v, address := self.getByPath(path...)
	if v.IsError() {
		if !v.isErrNotFoundLast() {
			return false, v
		}

		// find target node descriptor
		targetPath := path[l-1]
		desc, err := getDescByPath(self.Desc, path[:l-1]...)
		// var fd *proto.FieldDescriptor
		if err != nil {
			return false, err
		}

		// exchange PathFieldName to create PathFieldId
		if targetPath.t == PathFieldName {
			f := desc.Message().ByName(targetPath.str())
			targetPath = NewPathFieldId(f.Number())
			desc = f.Type()
		}
		
		// set sub node bytes by path and descriptor to check whether the node need to append tag
		if err := v.setNotFound(targetPath, &sub, desc); err != nil {
			return false, err
		}
	} else {
		exist = true
	}

	originLen := len(self.raw()) // root buf length
	err = self.replace(v.Node, sub) // replace ErrorNode bytes by sub Node bytes
	isPacked := path[l-1].t == PathIndex && sub.t.IsPacked()
	self.updateByteLen(originLen, address, isPacked, path...)
	return
}

// update parent node bytes length
func (self *Value) updateByteLen(originLen int, address []int, isPacked bool, path ...Path) {
	afterLen := self.l
	diffLen := afterLen - originLen
	previousType := proto.UNKNOWN

	for i := len(address) - 1; i >= 0; i-- {
		// notice: when i == len(address) - 1, it do not change bytes length because it has been changed in replace function, just change previousType
		pathType := path[i].t
		addressPtr := address[i]
		if previousType == proto.MESSAGE || (previousType == proto.LIST && isPacked) {
			newBytes := NewBytesFromPool()
			// tag
			buf := rt.BytesFrom(rt.AddPtr(self.v, uintptr(addressPtr)), self.l-addressPtr, self.l-addressPtr)
			_, tagOffset := protowire.ConsumeVarint(buf)
			// length
			length, lenOffset := protowire.ConsumeVarint(buf[tagOffset:])
			newLength := int(length) + diffLen
			newBytes = protowire.AppendVarint(newBytes, uint64(newLength))
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
			FreeBytesToPool(newBytes)
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
	var parentValue, address = self.getByPath(path[:l-1]...)
	if parentValue.IsError() {
		if parentValue.IsErrNotFound() {
			print(address)
			return nil
		}
		return parentValue
	}

	p := path[l-1]
	var desc *proto.TypeDescriptor
	var err error
	
	desc, err = getDescByPath(self.Desc, path[:l-1]...)
	if err != nil {
		return err
	}
	isPacked := desc.IsPacked()
	
	if p.t == PathFieldName {
		f := desc.Message().ByName(p.str())
		p = NewPathFieldId(f.Number())
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
	self.updateByteLen(originLen, address, isPacked, path...)
	return nil
}

func (self *Value) findDeleteChild(path Path) (Node, int) {
	tt := self.t // in fact, no need to judge which type
	valueLen := self.l
	exist := false
	var start, end int

	switch self.t {
	case proto.MESSAGE:
		it := self.iterFields()
		// if not root node
		if !self.IsRoot {
			if l, err := it.p.ReadLength(); err != nil {
				return errNode(meta.ErrRead, "", err), -1
			} else {
				valueLen = l
			}
		}
		start = valueLen // start initial at the end of the message
		end = 0          // end initial at the begin of the message

		// previous has change PathFieldName to PathFieldId
		if path.Type() != PathFieldId {
			return errNode(meta.ErrDismatchType, "path type is not PathFieldId", nil), -1
		}

		messageStart := it.p.Read
		id := path.id()
		for it.p.Read < messageStart+valueLen {
			fieldStart := it.p.Read
			fieldNumber, wireType, _, tagErr := it.p.ConsumeTag()
			if tagErr != nil {
				return errNode(meta.ErrRead, "", tagErr), -1
			}
			if err := it.p.Skip(wireType, false); err != nil {
				return errNode(meta.ErrRead, "", err), -1
			}
			fieldEnd := it.p.Read
			// get all items of the same fieldNumber, find the min start and max end, because the fieldNumber may be repeated
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
		it := self.iterElems()
		listIndex := 0
		if path.Type() != PathIndex {
			return errNode(meta.ErrDismatchType, "path type is not PathIndex", nil), -1
		}
		idx := path.int()

		elemType := self.Desc.Elem().WireType()
		size, err := self.Len()
		if err != nil {
			return errNode(meta.ErrRead, "", err), -1
		}

		// size = 0 maybe list in lazy load, need to check idx
		if size > 0 && idx >= size {
			return errNotFound, -1
		}
		// packed : [ListTag][ListLen][(l)v][(l)v][(l)v][(l)v].....
		if self.Desc.IsPacked() {
			if _, _, _, err := it.p.ConsumeTag(); err != nil {
				return errNode(meta.ErrRead, "", err), -1
			}

			if _, lengthErr := it.p.ReadLength(); lengthErr != nil {
				return errNode(meta.ErrRead, "", lengthErr), -1
			}

			for it.p.Read < valueLen && listIndex <= idx {
				start = it.p.Read
				if err := it.p.Skip(elemType, false); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}
				end = it.p.Read
				listIndex++
			}
		} else {
			// unpacked : [tag][l][v][tag][l][v][tag][l][v][tag][l][v]....
			for it.p.Read < valueLen && listIndex <= idx {
				start = it.p.Read
				if _, _, _, err := it.p.ConsumeTag(); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}

				if err := it.p.Skip(elemType, false); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}
				end = it.p.Read
				listIndex++
			}
		}
	case proto.MAP:
		it := self.iterPairs()

		if self.kt == proto.STRING {
			key := path.str()
			for it.p.Read < valueLen {
				start = it.p.Read
				// pair
				if _, _, _, err := it.p.ConsumeTag(); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}
				if _, err := it.p.ReadLength(); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}

				// key
				if _, _, _, err := it.p.ConsumeTag(); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}
				k, err := it.p.ReadString(false)
				if err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}

				// value
				_, valueWire, _, _ := it.p.ConsumeTag()
				if err := it.p.Skip(valueWire, false); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}
				end = it.p.Read
				if k == key {
					exist = true
					break
				}
			}
		} else if self.kt.IsInt() {
			key := path.Int()
			for it.p.Read < valueLen {
				start = it.p.Read

				if _, _, _, err := it.p.ConsumeTag(); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}

				if _, err := it.p.ReadLength(); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}

				// key
				if _, _, _, err := it.p.ConsumeTag(); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}

				k, err := it.p.ReadInt(self.kt)
				if err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}

				//value
				_, valueWire, _, _ := it.p.ConsumeTag()
				if err := it.p.Skip(valueWire, false); err != nil {
					return errNode(meta.ErrRead, "", err), -1
				}
				end = it.p.Read
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

// MarshalTo marshals self value into a sub value descripted by the to descriptor, alse called as "Cutting".
// Usually, the to descriptor is a subset of self descriptor.
func (self Value) MarshalTo(to *proto.TypeDescriptor, opts *Options) ([]byte, error) {
	var w = binary.NewBinaryProtocolBuffer()
	var r = binary.BinaryProtocol{}
	r.Buf = self.raw()
	var from = self.Desc
	messageLen := len(r.Buf)
	if err := marshalTo(&r, w, from, to, opts, messageLen); err != nil {
		return nil, err
	}
	ret := make([]byte, len(w.Buf))
	copy(ret, w.Buf)
	binary.FreeBinaryProtocol(w)
	return ret, nil
}

func marshalTo(read *binary.BinaryProtocol, write *binary.BinaryProtocol, from *proto.TypeDescriptor, to *proto.TypeDescriptor, opts *Options, massageLen int) error {
	tail := read.Read + massageLen
	for read.Read < tail {
		fieldNumber, wireType, _, _ := read.ConsumeTag()
		fromField := from.Message().ByNumber(fieldNumber)

		if fromField == nil {
			if opts.DisallowUnknown {
				return wrapError(meta.ErrUnknownField, fmt.Sprintf("unknown field %d", fieldNumber), nil)
			} else {
				// if not set, skip to the next field
				if err := read.Skip(wireType, opts.UseNativeSkip); err != nil {
					return wrapError(meta.ErrRead, "", err)
				}
				continue
			}
		}

		toField := to.Message().ByNumber(fieldNumber)

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

		// when FieldDescriptor.Kind() == MessageKind, it contained 3 conditions: message list, message map value, message;
		if fromType == proto.MessageKind {
			fromDesc := fromField.Type()
			toDesc := toField.Type()
			// message list
			if fromDesc.Type() == proto.LIST{
				fromDesc = fromDesc.Elem()
				toDesc = toDesc.Elem()
			}

			write.AppendTag(fieldNumber, wireType)
			var pos int
			subMessageLen, err := read.ReadLength()
			if err != nil {
				return wrapError(meta.ErrRead, "", err)
			}
			write.Buf, pos = binary.AppendSpeculativeLength(write.Buf)
			marshalTo(read, write, fromDesc, toDesc, opts, subMessageLen)
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

// GetByInt returns a sub node at the given key from a MAP value.
func (self Value) GetByStr(key string) (v Value) {
	n := self.Node.GetByStr(key)
	vd := self.Desc.Elem()
	if n.IsError() {
		return wrapValue(n, nil)
	}
	return wrapValue(n, vd)
}

// GetByInt returns a sub node at the given key from a MAP value.
func (self Value) GetByInt(key int) (v Value) {
	n := self.Node.GetByInt(key)
	vd := self.Desc.Elem()
	if n.IsError() {
		return wrapValue(n, nil)
	}
	return wrapValue(n, vd)
}

// Index returns a sub node at the given index from a LIST value.
func (self Value) Index(i int) (v Value) {
	n := self.Node.Index(i)
	if n.IsError() {
		return wrapValue(n, nil)
	}
	v = wrapValue(n, self.Desc.Elem())
	return
}

// FieldByName returns a sub node at the given field name from a MESSAGE value.
func (self Value) FieldByName(name string) (v Value) {
	if err := self.should("FieldByName", proto.MESSAGE); err != "" {
		return errValue(meta.ErrUnsupportedType, err, nil)
	}

	it := self.iterFields()
	if !self.IsRoot {
		if _, err := it.p.ReadLength(); err != nil {
			return errValue(meta.ErrRead, "", err)
		}
	}

	f := self.Desc.Message().ByName(name)
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
			typDesc := f.Type()
			if typDesc.IsMap() || typDesc.IsList() {
				it.p.Read = tagPos
				if _, err := it.p.SkipAllElements(i, typDesc.IsPacked()); err != nil {
					return errValue(meta.ErrRead, "SkipAllElements in LIST/MAP failed", err)
				}
				s = tagPos
				e = it.p.Read

				v = self.sliceWithDesc(s, e, typDesc)
				goto ret
			}

			t := proto.Kind2Wire[f.Kind()]
			if wt != t {
				v = errValue(meta.ErrDismatchType, fmt.Sprintf("field '%s' expects type %s, buf got type %s", string(f.Name()), t, wt), nil)
				goto ret
			}
			v = self.sliceWithDesc(s, e, typDesc)
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

// Field returns a sub node at the given field id from a MESSAGE value.
func (self Value) Field(id proto.FieldNumber) (v Value) {
	rootLayer := self.IsRoot
	msgDesc := self.Desc.Message()
	 
	n, d := self.Node.Field(id, rootLayer, msgDesc)

	if n.IsError() {
		return wrapValue(n, nil)
	}
	return wrapValue(n, d.Type())
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
	rootLayer := self.IsRoot
	msgDesc := self.Desc.Message()

	err := self.Node.Fields(ids, rootLayer, msgDesc, opts)
	return err
}

// SetMany: set a list of sub nodes at the given pathes from the value.
// root *Value: the root Node
// self *Value: the current Node (maybe root Node)
// address []int: the address from target Nodes to the root Node
// path ...Path: the path from root Node to target Nodes
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
	originLen := len(self.raw()) // current buf length
	rootLen := len(root.raw())   // root buf length
	isPacked := self.Desc.IsPacked()

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
			if self.t == proto.LIST || self.t == proto.MAP {
				self.size += 1
			}
		}
	}

	// if current is not root node, update current Node length
	if !self.IsRoot {
		err = self.replaceMany(ps)
		if self.t == proto.LIST && isPacked {
			currentAdd := []int{0, -1}
			currentPath := []Path{NewPathIndex(-1), NewPathIndex(-1)}
			self.updateByteLen(originLen, currentAdd, isPacked, currentPath...)
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
	root.updateByteLen(rootLen, address, isPacked, path...)
ret:
	ps.b = nil
	pnsPool.Put(ps)
	return
}
