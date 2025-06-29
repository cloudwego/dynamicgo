package generic

import (
	"fmt"
	"sort"
	"sync"
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/types"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
	"github.com/cloudwego/dynamicgo/proto/protowire"
)

var (
	_SkipMaxDepth = types.TB_SKIP_STACK_SIZE - 1
)

// PathType is the type of path
type PathType uint8

const (
	// PathFieldId represents a field id of MESSAGE type
	PathFieldId PathType = 1 + iota

	// PathFieldName represents a field name of MESSAGE type
	// NOTICE: it is only supported by Value
	PathFieldName

	// PathIndex represents a index of LIST type
	PathIndex

	// Path represents a string key of MAP type
	PathStrKey

	// Path represents a int key of MAP type
	PathIntKey

	// Path represents a raw-bytes key of MAP type
	// It is usually used for neither-string-nor-integer type key
	PathBinKey // not supported in protobuf MapKeyDescriptor
)

// Path represents the relative position of a sub node in a complex parent node
type Path struct {
	t PathType // path type
	v unsafe.Pointer
	l int
}

// Str returns the string value of a PathFieldName\PathStrKey path
func (self Path) Str() string {
	switch self.t {
	case PathFieldName, PathStrKey:
		return self.str()
	default:
		return ""
	}
}

func (self Path) str() string {
	return *(*string)(unsafe.Pointer(&self.v))
}

// Int returns the int value of a PathIndex\PathIntKey path
func (self Path) Int() int {
	switch self.t {
	case PathIndex, PathIntKey:
		return self.int()
	default:
		return -1
	}
}

func (self Path) int() int {
	return self.l
}

// Id returns the field id of a PathFieldId path
func (self Path) Id() proto.FieldNumber {
	switch self.t {
	case PathFieldId:
		return self.id()
	default:
		return proto.FieldNumber(0)
	}
}

func (self Path) id() proto.FieldNumber {
	return proto.FieldNumber(self.l)
}

// Type returns the type of a Path
func (self Path) Type() PathType {
	return self.t
}

// Value returns the equivalent go interface of a Path
func (self Path) Value() interface{} {
	switch self.t {
	case PathFieldId:
		return self.id()
	case PathFieldName, PathStrKey:
		return self.str()
	case PathIndex, PathIntKey:
		return self.int()
	default:
		return nil
	}
}

// ToRaw converts underlying value to protobuf-encoded bytes
func (self Path) ToRaw(t proto.Type) []byte {
	switch self.t {
	case PathFieldId:
		// tag
		ret := make([]byte, 0, DefaultTagSliceCap)
		if t != proto.LIST && t != proto.MAP {
			kind := t.TypeToKind()
			tag := uint64(self.l)<<3 | uint64(proto.Kind2Wire[kind])
			ret = protowire.BinaryEncoder{}.EncodeUint64(ret, tag)
		}
		return ret
	case PathStrKey:
		// tag + string key
		ret := make([]byte, 0, DefaultTagSliceCap)
		tag := uint64(1)<<3 | uint64(proto.STRING)
		ret = protowire.BinaryEncoder{}.EncodeUint64(ret, tag)
		ret = protowire.BinaryEncoder{}.EncodeString(ret, self.str())
		return ret
	case PathIntKey:
		// tag + int key
		kind := t.TypeToKind()
		ret := make([]byte, 0, DefaultTagSliceCap)
		tag := uint64(1)<<3 | uint64(proto.Kind2Wire[kind])
		ret = protowire.BinaryEncoder{}.EncodeUint64(ret, tag)
		switch t {
		case proto.INT32:
			ret = protowire.BinaryEncoder{}.EncodeInt32(ret, int32(self.l))
		case proto.SINT32:
			ret = protowire.BinaryEncoder{}.EncodeSint32(ret, int32(self.l))
		case proto.SFIX32:
			ret = protowire.BinaryEncoder{}.EncodeSfixed32(ret, int32(self.l))
		case proto.INT64:
			ret = protowire.BinaryEncoder{}.EncodeInt64(ret, int64(self.l))
		case proto.SINT64:
			ret = protowire.BinaryEncoder{}.EncodeSint64(ret, int64(self.l))
		case proto.SFIX64:
			ret = protowire.BinaryEncoder{}.EncodeSfixed64(ret, int64(self.l))
		}
		return ret
	case PathBinKey:
		return nil
	default:
		return nil
	}
}

// NewPathFieldId creates a PathFieldId path
func NewPathFieldId(id proto.FieldNumber) Path {
	return Path{
		t: PathFieldId,
		l: int(id),
	}
}

// NewPathFieldName creates a PathFieldName path
func NewPathFieldName(name string) Path {
	return Path{
		t: PathFieldName,
		v: *(*unsafe.Pointer)(unsafe.Pointer(&name)),
		l: len(name),
	}
}

// NewPathIndex creates a PathIndex path
func NewPathIndex(index int) Path {
	return Path{
		t: PathIndex,
		l: index,
	}
}

// NewPathStrKey creates a PathStrKey path
func NewPathStrKey(key string) Path {
	return Path{
		t: PathStrKey,
		v: *(*unsafe.Pointer)(unsafe.Pointer(&key)),
		l: len(key),
	}
}

// NewPathIntKey creates a PathIntKey path
func NewPathIntKey(key int) Path {
	return Path{
		t: PathIntKey,
		l: key,
	}
}

// PathNode is a three node of DOM tree
type PathNode struct {
	Path
	Node
	Next []PathNode
}

var pathNodePool = sync.Pool{
	New: func() interface{} {
		return &PathNode{}
	},
}

// NewPathNode get a new PathNode from memory pool
func NewPathNode() *PathNode {
	return pathNodePool.Get().(*PathNode)
}

// CopyTo deeply copy self and its children to a PathNode
func (self PathNode) CopyTo(to *PathNode) {
	to.Path = self.Path
	to.Node = self.Node
	if cap(to.Next) < len(self.Next) {
		to.Next = make([]PathNode, len(self.Next))
	}
	to.Next = to.Next[:len(self.Next)]
	for i, c := range self.Next {
		c.CopyTo(&to.Next[i])
	}
}

// ResetValue resets self's node and its children's node
func (self *PathNode) ResetValue() {
	self.Node = Node{}
	for i := range self.Next {
		self.Next[i].ResetValue()
	}
}

// ResetAll resets self and its children, including path and node both
func (self *PathNode) ResetAll() {
	for i := range self.Next {
		self.Next[i].ResetAll()
	}
	self.Node = Node{}
	self.Path = Path{}
	self.Next = self.Next[:0]
}

// FreePathNode put a PathNode back to memory pool
func FreePathNode(p *PathNode) {
	p.Path = Path{}
	p.Node = Node{}
	p.Next = p.Next[:0]
	pathNodePool.Put(p)
}

// extend cap of a PathNode slice
func guardPathNodeSlice(con *[]PathNode, l int) {
	c := cap(*con) // Get the current capacity of the slice
	if l >= c {
		tmp := make([]PathNode, len(*con), l+DefaultNodeSliceCap) // Create a new slice 'tmp'
		copy(tmp, *con)                                           // Copy elements from the original slice to the new slice 'tmp'
		*con = tmp                                                // Update the reference of the original slice to point to the new slice 'tmp'
	}
}

// scanChildren scans all children of self and store them in self.Next
// messageLen is only used when self.Node.t == proto.MESSAGE
func (self *PathNode) scanChildren(p *binary.BinaryProtocol, recurse bool, opts *Options, desc *proto.TypeDescriptor, messageLen int) (err error) {
	next := self.Next[:0] // []PathNode len=0
	l := len(next)
	c := cap(next)

	var v *PathNode
	switch self.Node.t {
	case proto.MESSAGE:
		messageDesc := desc.Message()
		start := p.Read
		// range all fields formats: [FieldTag(L)V][FieldTag(L)V][FieldTag(L)V]...
		for p.Read < start+messageLen {
			fieldNumber, wireType, tagLen, tagErr := p.ConsumeTag()
			if tagErr != nil {
				return wrapError(meta.ErrRead, "PathNode.scanChildren: invalid field tag.", tagErr)
			}

			field := messageDesc.ByNumber(fieldNumber)
			if field != nil {
				v, err = self.handleChild(&next, &l, &c, p, recurse, field.Type(), tagLen, opts)
			} else {
				// store unknown field without recurse subnodes, containing the whole [TLV] of unknown field
				v, err = self.handleUnknownChild(&next, &l, &c, p, recurse, opts, fieldNumber, wireType, tagLen)
			}

			if err != nil {
				return err
			}
			v.Path = NewPathFieldId(fieldNumber)
		}
	case proto.LIST:
		// range all elements
		// FieldDesc := (*desc).(proto.FieldDescriptor)
		fieldNumber := desc.BaseId()
		start := p.Read
		// must set list element type first, so that handleChild will can handle the element correctly
		self.et = desc.Elem().Type()
		listIndex := 0
		// packed formats: [ListFieldTag][ListByteLen][VVVVV]...
		if desc.IsPacked() {
			if _, _, _, tagErr := p.ConsumeTag(); tagErr != nil {
				return wrapError(meta.ErrRead, "PathNode.scanChildren: invalid list tag", tagErr)
			}

			listLen, lengthErr := p.ReadLength()
			if lengthErr != nil {
				return wrapError(meta.ErrRead, "PathNode.scanChildren: invalid list len", lengthErr)
			}

			start = p.Read
			for p.Read < start+listLen {
				// listLen is not used when node is LIST
				v, err = self.handleChild(&next, &l, &c, p, recurse, desc.Elem(), listLen, opts)
				if err != nil {
					return err
				}
				v.Path = NewPathIndex(listIndex)
				listIndex++
			}
		} else {
			// unpacked formats: [FieldTag(L)V][FieldTag(L)V]...
			for p.Read < len(p.Buf) {
				itemNumber, _, tagLen, tagErr := p.ConsumeTagWithoutMove()
				if tagErr != nil {
					return wrapError(meta.ErrRead, "PathNode.scanChildren: invalid element tag", tagErr)
				}

				if itemNumber != fieldNumber {
					break
				}
				p.Read += tagLen
				// tagLen is not used when node is MAP
				v, err = self.handleChild(&next, &l, &c, p, recurse, desc.Elem(), tagLen, opts)
				if err != nil {
					return err
				}
				v.Path = NewPathIndex(listIndex)
				listIndex++
			}
		}
		self.size = listIndex
	case proto.MAP:
		// range all map entries: [PairTag][PairLen][MapKeyT(L)V][MapValueT(L)V], [PairTag][PairLen][MapKeyT(L)V][MapValueT(L)V]...
		keyDesc := desc.Key()
		valueDesc := desc.Elem()
		// set map key type and value type
		self.kt = keyDesc.Type() // map key type only support int/string
		self.et = valueDesc.Type()
		mapNumber := desc.BaseId()
		for p.Read < len(p.Buf) {
			pairNumber, _, pairTagLen, pairTagErr := p.ConsumeTagWithoutMove()
			if pairTagErr != nil {
				return wrapError(meta.ErrRead, "PathNode.scanChildren: Consume pair tag failed", nil)
			}
			if pairNumber != mapNumber {
				break
			}
			p.Read += pairTagLen

			pairLen, pairLenErr := p.ReadLength()
			if pairLen <= 0 || pairLenErr != nil {
				return wrapError(meta.ErrRead, "PathNode.scanChildren:invalid pair len", pairLenErr)
			}

			if _, _, _, keyTagErr := p.ConsumeTag(); keyTagErr != nil {
				return wrapError(meta.ErrRead, "PathNode.scanChildren: Consume map key tag failed", nil)
			}

			var keyString string
			var keyInt int
			var keyErr error

			if self.kt == proto.STRING {
				keyString, keyErr = p.ReadString(false)
			} else if self.kt.IsInt() {
				keyInt, keyErr = p.ReadInt(self.kt)
			} else {
				return wrapError(meta.ErrUnsupportedType, "PathNode.scanChildren: Unsupported map key type", nil)
			}

			if keyErr != nil {
				return wrapError(meta.ErrRead, "PathNode.scanChildren: can not read map key.", keyErr)
			}

			_, _, valueLen, valueTagErr := p.ConsumeTag() // consume value tag

			if valueTagErr != nil {
				return wrapError(meta.ErrRead, "PathNode.scanChildren: Consume map value tag failed", nil)
			}

			v, err = self.handleChild(&next, &l, &c, p, recurse, valueDesc, valueLen, opts)
			if err != nil {
				return err
			}

			if self.kt == proto.STRING {
				v.Path = NewPathStrKey(keyString)
			} else if self.kt.IsInt() {
				v.Path = NewPathIntKey(keyInt)
			} else {
				return wrapError(meta.ErrUnsupportedType, "PathNode.scanChildren: Unsupported map key type", nil)
			}
			self.size++ // map enrty size ++
		}
	default:
		return wrapError(meta.ErrUnsupportedType, "PathNode.scanChildren: Unsupported children type", nil)
	}

	self.Next = next
	return nil
}

func (self *PathNode) handleChild(in *[]PathNode, lp *int, cp *int, p *binary.BinaryProtocol, recurse bool, desc *proto.TypeDescriptor, tagL int, opts *Options) (*PathNode, error) {
	var con = *in
	var l = *lp
	guardPathNodeSlice(&con, l) // extend cap of con
	if l >= len(con) {
		con = con[:l+1]
	}
	v := &con[l]
	l += 1

	start := p.Read
	buf := p.Buf

	var skipType proto.WireType
	tt := desc.Type()
	IsPacked := desc.IsPacked()

	if tt != proto.LIST {
		skipType = desc.WireType()
	}

	// if parent node is LIST, the children type of parent node is the element type of LIST
	if self.Node.t == proto.LIST {
		tt = self.et
	}

	if tt == proto.LIST || tt == proto.MAP {
		// when list/map the parent node will contain the tag, so the start need to move back tagL
		start = start - tagL
		if start < 0 {
			return nil, wrapError(meta.ErrRead, "invalid start", nil)
		}
		skipType = proto.BytesType
	}

	// notice: when parent node is packed LIST, the size is not calculated in order to fast read all elements
	if e := p.Skip(skipType, opts.UseNativeSkip); e != nil {
		return nil, wrapError(meta.ErrRead, "skip field failed", e)
	}

	// unpacked LIST or MAP
	if (tt == proto.LIST && IsPacked == false) || tt == proto.MAP {
		fieldNumber := desc.BaseId()
		// skip remain elements with the same field number
		for p.Read < len(p.Buf) {
			number, wt, tagLen, err := p.ConsumeTagWithoutMove()
			if err != nil {
				return nil, err
			}
			if number != fieldNumber {
				break
			}
			p.Read += tagLen
			if e := p.Skip(wt, opts.UseNativeSkip); e != nil {
				return nil, wrapError(meta.ErrRead, "skip field failed", e)
			}
		}
	}
	v.Node = self.slice(start, p.Read, tt)

	if tt.IsComplex() {
		if recurse {
			p.Buf = p.Buf[start:]
			p.Read = 0
			parentDesc := desc
			messageLen := 0
			if tt == proto.MESSAGE {
				// parentDesc = desc.Message()
				var err error
				messageLen, err = p.ReadLength() // the sub message has message byteLen need to read before next recurse for scanChildren
				if messageLen <= 0 || err != nil {
					return nil, wrapError(meta.ErrRead, "read message length failed", err)
				}
			}

			if err := v.scanChildren(p, recurse, opts, parentDesc, messageLen); err != nil {
				return nil, err
			}
			p.Buf = buf
			p.Read = start + p.Read
		} else {
			// set complex Node type when lazy load
			if tt == proto.LIST {
				v.et = desc.Elem().Type()
			} else if tt == proto.MAP {
				v.kt = desc.Key().Type()
				v.et = desc.Elem().Type()
			}
		}

	}

	*in = con
	*lp = l
	*cp = cap(con)
	return v, nil
}

// handleUnknownChild handles unknown field and store the [TLV] of unknown field to the proto.UNKNOWN Node
func (self *PathNode) handleUnknownChild(in *[]PathNode, lp *int, cp *int, p *binary.BinaryProtocol, recurse bool, opts *Options, fieldNumber proto.FieldNumber, wireType proto.WireType, tagL int) (*PathNode, error) {
	var con = *in
	var l = *lp
	guardPathNodeSlice(&con, l) // extend cap of con
	if l >= len(con) {
		con = con[:l+1]
	}
	v := &con[l]
	l += 1

	start := p.Read - tagL

	if e := p.Skip(wireType, opts.UseNativeSkip); e != nil {
		return nil, wrapError(meta.ErrRead, "skip unknown field failed", e)
	}

	// maybe there are more unknown fields with the same field number such as unknown packed list or map
	for p.Read < len(p.Buf) {
		number, wt, tagLen, err := p.ConsumeTagWithoutMove()
		if err != nil {
			return nil, err
		}
		if number != fieldNumber {
			break
		}
		p.Read += tagLen
		if e := p.Skip(wt, opts.UseNativeSkip); e != nil {
			return nil, wrapError(meta.ErrRead, "skip unknown field failed", e)
		}
	}
	v.Node = self.slice(start, p.Read, proto.UNKNOWN)

	*in = con
	*lp = l
	*cp = cap(con)
	return v, nil
}

// Load loads self's all children ( and children's children if recurse is true) into self.Next,
// no matter whether self.Next is empty or set before (will be reset).
// NOTICE: if opts.NotScanParentNode is true, the parent nodes (PathNode.Node) of complex (map/list/struct) type won't be assgined data
func (self *PathNode) Load(recurse bool, opts *Options, desc *proto.TypeDescriptor) error {
	if self == nil {
		panic("nil PathNode")
	}
	if self.Error() != "" {
		return self
	}
	self.Next = self.Next[:0]
	p := binary.BinaryProtocol{
		Buf: self.Node.raw(),
	}
	// fd, ok := (*desc).(proto.Descriptor)
	// if !ok {
	// 	return wrapError(meta.ErrInvalidParam, "invalid descriptor", nil)
	// }
	return self.scanChildren(&p, recurse, opts, desc, len(p.Buf))
}

func getDescByPath(root *proto.TypeDescriptor, pathes ...Path) (*proto.TypeDescriptor, error) {
	desc := root
	for i, p := range pathes {
		if i == 0 {
			switch p.Type() {
			case PathFieldId:
				f := desc.Message().ByNumber(p.id())
				if f == nil {
					return nil, wrapError(meta.ErrUnknownField, fmt.Sprintf("unknown field %d", p.id()), nil)
				}
				desc = f.Type()
			case PathFieldName:
				f := desc.Message().ByName(p.str())
				if f == nil {
					return nil, wrapError(meta.ErrUnknownField, fmt.Sprintf("unknown field %s", p.str()), nil)
				}
				desc = f.Type()
			default:
				return nil, wrapError(meta.ErrUnsupportedType, "unsupported path type", nil)
			}
		} else {
			t := desc.Type()
			switch t {
			case proto.MESSAGE:
				switch p.Type() {
				case PathFieldId:
					f := desc.Message().ByNumber(p.id())
					if f == nil {
						return nil, wrapError(meta.ErrUnknownField, fmt.Sprintf("unknown field %d", p.id()), nil)
					}
					desc = f.Type()
				case PathFieldName:
					f := desc.Message().ByName(p.str())
					if f == nil {
						return nil, wrapError(meta.ErrUnknownField, fmt.Sprintf("unknown field %s", p.str()), nil)
					}
					desc = f.Type()
				default:
					return nil, wrapError(meta.ErrUnsupportedType, "unsupported path type", nil)
				}
			case proto.MAP:
				valueDesc := desc.Elem()
				desc = valueDesc
			case proto.LIST:
				desc = desc.Elem()
			default:
				return nil, wrapError(meta.ErrInvalidParam, "unsupported path type", nil)
			}
		}

		if desc == nil {
			return nil, wrapError(meta.ErrNotFound, "descriptor is not found.", nil)
		}
	}

	return desc, nil
}

func (self PathNode) Marshal(opt *Options) (out []byte, err error) {
	p := binary.NewBinaryProtocolBuffer()
	rootLayer := true // TODO: make rootLayer to be a parameter will be better
	err = self.marshal(p, rootLayer, opt)
	if err == nil {
		out = make([]byte, len(p.Buf))
		copy(out, p.Buf)
	}
	binary.FreeBinaryProtocol(p)
	return
}

func (self PathNode) marshal(p *binary.BinaryProtocol, rootLayer bool, opts *Options) error {
	if self.IsError() {
		return self.Node
	}

	// if node is basic type, Next[] is empty, directly append to buf
	if len(self.Next) == 0 {
		p.Buf = append(p.Buf, self.Node.raw()...)
		return nil
	}

	var err error

	switch self.Node.t {
	case proto.MESSAGE:
		pos := -1
		// only root layer no need append message tag and write prefix length
		if !rootLayer {
			p.Buf, pos = binary.AppendSpeculativeLength(p.Buf)
			if pos < 0 {
				return wrapError(meta.ErrWrite, "PathNode.marshal: append speculative length failed", nil)
			}
		}

		for _, v := range self.Next {
			// when node type is not LIST/MAP write tag
			if v.Node.t != proto.LIST && v.Node.t != proto.MAP && v.Node.t != proto.UNKNOWN {
				err = p.AppendTag(v.Path.Id(), proto.Kind2Wire[v.Node.t.TypeToKind()])
			}
			if err != nil {
				return wrapError(meta.ErrWrite, "PathNode.marshal: append tag failed", err)
			}

			if err = v.marshal(p, false, opts); err != nil {
				return unwrapError(fmt.Sprintf("field %d  marshal failed", v.Path.id()), err)
			}
		}
		if !rootLayer {
			p.Buf = binary.FinishSpeculativeLength(p.Buf, pos)
		}
	case proto.LIST:
		et := self.et
		filedNumber := self.Path.Id()
		IsPacked := et.IsPacked()
		pos := -1
		// packed need first list tag and write prefix length
		if IsPacked {
			if err = p.AppendTag(filedNumber, proto.BytesType); err != nil {
				return wrapError(meta.ErrWrite, "PathNode.marshal: append tag failed", err)
			}
			p.Buf, pos = binary.AppendSpeculativeLength(p.Buf)
			if pos < 0 {
				return wrapError(meta.ErrWrite, "PathNode.marshal: append speculative length failed", nil)
			}
		}

		for _, v := range self.Next {
			// unpacked need append tag first
			if !IsPacked {
				if err = p.AppendTag(filedNumber, proto.Kind2Wire[v.Node.t.TypeToKind()]); err != nil {
					return wrapError(meta.ErrWrite, "PathNode.marshal: append tag failed", err)
				}
			}

			if err = v.marshal(p, false, opts); err != nil {
				return unwrapError(fmt.Sprintf("field %d  marshal failed", v.Path.id()), err)
			}
		}

		// packed mode need update prefix list length
		if IsPacked {
			p.Buf = binary.FinishSpeculativeLength(p.Buf, pos)
		}

	case proto.MAP:
		kt := self.kt
		et := self.et
		filedNumber := self.Path.Id()
		pos := -1

		for _, v := range self.Next {
			// append pair tag and write prefix length
			if err = p.AppendTag(filedNumber, proto.BytesType); err != nil {
				return wrapError(meta.ErrWrite, "PathNode.marshal: append tag failed", err)
			}
			p.Buf, pos = binary.AppendSpeculativeLength(p.Buf)
			if pos < 0 {
				return wrapError(meta.ErrWrite, "PathNode.marshal: append speculative length failed", nil)
			}

			// write key tag + value
			if kt == proto.STRING {
				// Mapkey field number is 1
				if err = p.AppendTag(1, proto.BytesType); err != nil {
					return wrapError(meta.ErrWrite, "PathNode.marshal: append tag failed", err)
				}

				if err = p.WriteString(v.Path.Str()); err != nil {
					return wrapError(meta.ErrWrite, "PathNode.marshal: append string failed", err)
				}
			} else if kt.IsInt() {
				wt := proto.Kind2Wire[kt.TypeToKind()]
				// Mapkey field number is 1
				if err = p.AppendTag(1, wt); err != nil {
					return wrapError(meta.ErrWrite, "PathNode.marshal: append tag failed", err)
				}

				if wt == proto.VarintType {
					err = p.WriteInt64(int64(v.Path.int()))
				} else if wt == proto.Fixed32Type {
					err = p.WriteSfixed32(int32(v.Path.int()))
				} else if wt == proto.Fixed64Type {
					err = p.WriteSfixed64(int64(v.Path.int()))
				}
				if err != nil {
					return wrapError(meta.ErrWrite, "PathNode.marshal: append int failed", err)
				}
			} else {
				return wrapError(meta.ErrUnsupportedType, "PathNode.marshal: unsupported map key type", nil)
			}

			// if value is basic type, need append tag first
			if v.Node.t != proto.LIST && v.Node.t != proto.MAP {
				// MapValue field number is 2
				if err = p.AppendTag(2, proto.Kind2Wire[et.TypeToKind()]); err != nil {
					return wrapError(meta.ErrWrite, "PathNode.marshal: append tag failed", err)
				}
			}

			if err = v.marshal(p, false, opts); err != nil {
				return unwrapError(fmt.Sprintf("field %d  marshal failed", v.Path.id()), err)
			}
			p.Buf = binary.FinishSpeculativeLength(p.Buf, pos)
		}
	case proto.BOOL, proto.INT32, proto.SINT32, proto.UINT32, proto.FIX32, proto.SFIX32, proto.INT64, proto.SINT64, proto.UINT64, proto.FIX64, proto.SFIX64, proto.FLOAT, proto.DOUBLE, proto.STRING, proto.BYTE:
		p.Buf = append(p.Buf, self.Node.raw()...)
	case proto.UNKNOWN:
		// unknown bytes can also be marshaled, but we don't know its real type, so we can't read it, just skip it.
		p.Buf = append(p.Buf, self.Node.raw()...)
	default:
		return wrapError(meta.ErrUnsupportedType, "PathNode.marshal: unsupported type", nil)
	}

	return err
}

// pathNode Slice Pool
type pnSlice struct {
	a []PathNode
	b []PathNode
}

func (self pnSlice) Len() int {
	return len(self.a)
}

func (self *pnSlice) Swap(i, j int) {
	self.a[i], self.a[j] = self.a[j], self.a[i]
	self.b[i], self.b[j] = self.b[j], self.b[i]
}

func (self pnSlice) Less(i, j int) bool {
	return int(uintptr(self.a[i].Node.v)) < int(uintptr(self.a[j].Node.v))
}

func (self *pnSlice) Sort() {
	sort.Sort(self)
}
