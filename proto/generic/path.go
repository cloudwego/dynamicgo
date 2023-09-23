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

// PathType is the type of path
type PathType uint8

const (
	// PathFieldId represents a field id of STRUCT type
	PathFieldId PathType = 1 + iota

	// PathFieldName represents a field name of STRUCT type
	// NOTICE: it is only supported by Value
	PathFieldName

	// PathIndex represents a index of LIST\SET type
	PathIndex

	// Path represents a string key of MAP type
	PathStrKey

	// Path represents a int key of MAP type
	PathIntKey

	// Path represents a raw-bytes key of MAP type
	// It is usually used for neither-string-nor-integer type key
	PathBinKey
)

// Path represents the relative position of a sub node in a complex parent node
type Path struct {
	t PathType // path type
	v unsafe.Pointer // value ptr
	l int // field number
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

// Bin returns the raw bytes value of a PathBinKey path
func (self Path) Bin() []byte {
	switch self.t {
	case PathBinKey:
		return self.bin()
	default:
		return nil
	}
}

func (self Path) bin() []byte {
	return rt.BytesFrom(self.v, self.l, self.l)
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
	case PathBinKey:
		return self.bin()
	default:
		return nil
	}
}

// ToRaw converts underlying value to thrift-encoded bytes
func (self Path) ToRaw(t proto.Type) []byte {
	kind := t.TypeToKind()
	switch self.t {
	case PathFieldId:
		// tag
		ret := make([]byte, 0, 8)
		tag := uint64(self.l) << 3 | uint64(proto.Kind2Wire[kind])
		ret = protowire.BinaryEncoder{}.EncodeUint64(ret, tag)
		return ret
	case PathStrKey:
		// tag + string key
		ret := make([]byte, 0, 8)
		tag := uint64(0) << 3 | uint64(proto.STRING)
		ret = protowire.BinaryEncoder{}.EncodeUint64(ret, tag)
		ret = protowire.BinaryEncoder{}.EncodeString(ret, self.str())
		return ret
	case PathIntKey:
		// tag + int key
		ret := make([]byte, 0, 8)
		tag := uint64(0) << 3 | uint64(proto.Kind2Wire[kind])
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

// NewPathBinKey creates a PathBinKey path
func NewPathBinKey(key []byte) Path {
	return Path{
		t: PathBinKey,
		v: *(*unsafe.Pointer)(unsafe.Pointer(&key)),
		l: len(key),
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

// FreePathNode put a PathNode back to memory pool
func FreePathNode(p *PathNode) {
	p.Path = Path{}
	p.Node = Node{}
	p.Next = p.Next[:0]
	pathNodePool.Put(p)
}

// extend cap of a PathNode slice
func guardPathNodeSlice(con *[]PathNode, l int) {
    c := cap(*con)  // Get the current capacity of the slice
    if l >= c {
        tmp := make([]PathNode, len(*con), l + DefaultNodeSliceCap)  // Create a new slice 'tmp'
        copy(tmp, *con)  // Copy elements from the original slice to the new slice 'tmp'
        *con = tmp  // Update the reference of the original slice to point to the new slice 'tmp'
    }
}

// DescriptorToPathNode converts a proto kind descriptor to a DOM, assgining path to root
// NOTICE: it only recursively converts MESSAGE type
func DescriptorToPathNode(desc *proto.FieldDescriptor, root *PathNode, opts *Options) error {
	if desc == nil || root == nil {
		panic("nil pointer")
	}
	
	if (*desc).Kind() == proto.MessageKind && (*desc).IsMap() == false {
		fs := (*desc).Message().Fields()
		ns := root.Next
		if cap(ns) == 0 {
			ns = make([]PathNode, 0, fs.Len())
		} else {
			ns = ns[:0]
		}
		for i := 0; i <= fs.Len(); i++ {
			fd := fs.Get(i)
			var p PathNode
			// if opts.FieldByName {
			// 	p.Path = NewPathFieldName(f.Name())
			// } else {
			p.Path = NewPathFieldId(fd.Number())
			// }
			if err := DescriptorToPathNode(&fd, &p, opts); err != nil {
				return err
			}
			ns = append(ns, p)
		}
		root.Next = ns

	}
	return nil
}




func (self *PathNode) scanChildren(p *binary.BinaryProtocol, recurse bool, opts *Options, desc *proto.Descriptor,messageLen int) (err error) {
	next := self.Next[:0] // []PathNode len=0
	l := len(next)
	c := cap(next)
	maxId := StoreChildrenByIdShreshold
	var v *PathNode
	switch self.Node.t {
	case proto.MESSAGE:
		messageDesc := (*desc).(proto.MessageDescriptor)
		fields := messageDesc.Fields()
		start := p.Read
		for p.Read < start + messageLen{
			fieldNumber, wireType, _, tagErr := p.ConsumeTag()
			if tagErr != nil {
				return errNode(meta.ErrRead, "", tagErr)
			}
			
			// OPT: store children by id here, thus we can use id as index to access children.
			if opts.StoreChildrenById {
				// if id is larger than the threshold, we store children after the threshold.
				if int(fieldNumber) < StoreChildrenByIdShreshold {
					l = int(fieldNumber)
				} else {
					l = maxId
					maxId += 1
				}
			}
			if fieldNumber == 255{
				fmt.Println("fieldNumber == 255")
			}
			
			field := fields.ByNumber(fieldNumber)
			if field != nil {
				v, err = self.handleChild(&next, &l, &c, p, recurse, opts, &field)
			} else {
				// store unknown field without recurse subnodes
				v, err = self.handleUnknownChild(&next, &l, &c, p, recurse, opts, wireType)
			}
			
			if err != nil {
				return err
			}
			v.Path = NewPathFieldId(fieldNumber)
		}
	case proto.LIST:
		FieldDesc := (*desc).(proto.FieldDescriptor)
		fieldNumber := FieldDesc.Number()
		start := p.Read
		self.et = proto.FromProtoKindToType(FieldDesc.Kind(),false,false)
		// packed
		if FieldDesc.IsPacked() {
			listLen, lengthErr := p.ReadLength()
			if lengthErr != nil {
				return errNode(meta.ErrRead, "", lengthErr)
			}
			start = p.Read
			listIndex := 0
			for p.Read < start + listLen {
				v, err = self.handleListChild(&next, &l, &c, p, recurse, opts, &FieldDesc)
				if err != nil {
					return err
				}
				v.Path = NewPathIndex(listIndex)
				listIndex++
			}
		} else {
			listIndex := 0
			for p.Read < len(p.Buf) {
				v, err = self.handleListChild(&next, &l, &c, p, recurse, opts, &FieldDesc)
				if err != nil {
					return err
				}
				v.Path = NewPathIndex(listIndex)
				listIndex++

				if p.Read >= len(p.Buf) {
					break
				}

				itemNumber, _, tagLen, tagErr := p.ConsumeTagWithoutMove()
				if tagErr != nil {
					return errNode(meta.ErrRead, "", tagErr)
				}
				
				if itemNumber != fieldNumber {
					break
				}
				p.Read += tagLen
			}
		}
	case proto.MAP:
		mapDesc := (*desc).(proto.FieldDescriptor)
		keyDesc := mapDesc.MapKey()
		valueDesc := mapDesc.MapValue()
		self.kt = proto.FromProtoKindToType(keyDesc.Kind(),keyDesc.IsList(),keyDesc.IsMap())
		self.et = proto.FromProtoKindToType(valueDesc.Kind(),valueDesc.IsList(),valueDesc.IsMap())
		for p.Read < len(p.Buf) {
			pairLen, pairLenErr := p.ReadLength()
			if pairLen <= 0 || pairLenErr != nil {
				return errNode(meta.ErrRead, "", pairLenErr)
			}
		
			_, _, _, keyTagErr := p.ConsumeTag()
			if keyTagErr != nil {
				return meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
			}
		
			var key interface{}
			var keyErr error
		
			if self.kt == proto.STRING {
				key, keyErr = p.ReadString(false)
			} else if self.kt.IsInt() {
				key, keyErr = p.ReadInt(self.kt)
			} else {
				return meta.NewError(meta.ErrRead, "Unsupported key type", nil)
			}
		
			if keyErr != nil {
				return errNode(meta.ErrRead, "", keyErr)
			}
		
			_, _, _, valueTagErr := p.ConsumeTag()
			if valueTagErr != nil {
				return meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
			}
		
			v, err = self.handleChild(&next, &l, &c, p, recurse, opts, &valueDesc)
			if err != nil {
				return err
			}
		
			if self.kt == proto.STRING {
				v.Path = NewPathStrKey(key.(string))
			} else if self.kt.IsInt() {
				v.Path = NewPathIntKey(key.(int))
			}

			if p.Read >= len(p.Buf) {
				break
			}
		
			pairNumber, _, tagLen, pairTagErr := p.ConsumeTagWithoutMove()
			if pairTagErr != nil {
				return meta.NewError(meta.ErrRead, "ConsumeTag failed", nil)
			}
		
			if pairNumber != mapDesc.Number() {
				break
			}
		
			p.Read += tagLen
		}
	default:
		return errNode(meta.ErrUnsupportedType, "", nil)
	}

	self.Next = next
	return nil
}

func (self *PathNode) handleChild(in *[]PathNode, lp *int, cp *int, p *binary.BinaryProtocol, recurse bool, opts *Options, desc *proto.FieldDescriptor) (*PathNode, error) {
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
	kind := (*desc).Kind()
	et := proto.FromProtoKindToType(kind,(*desc).IsList(),(*desc).IsMap())

	if recurse && (et.IsComplex() && opts.NotScanParentNode) {
		v.Node = Node{
			t: et,
			l: 0,
			v: unsafe.Pointer(uintptr(self.Node.v) + uintptr(start)),
		}
	} else {
		if e := p.Skip(proto.Kind2Wire[kind], opts.UseNativeSkip); e != nil {
			return nil, errNode(meta.ErrRead, "", e)
		}
		v.Node = self.slice(start, p.Read, et, 0, 0)
	}

	if recurse && et.IsComplex() {
		p.Buf = p.Buf[start:]
		p.Read = 0
		parentDesc := (*desc).(proto.Descriptor)
		messageLen := 0
		if et == proto.MESSAGE {
			parentDesc = (*desc).Message().(proto.Descriptor)
			var err error
			messageLen, err = p.ReadLength()
			if messageLen<= 0 || err != nil {
				panic("read message length failed")
			}
		}

		if err := v.scanChildren(p, recurse, opts, &parentDesc,messageLen); err != nil {
			return nil, err
		}
		p.Buf = buf
		p.Read = start + p.Read
	}

	*in = con
	*lp = l
	*cp = cap(con)
	return v, nil
}

func (self *PathNode) handleUnknownChild(in *[]PathNode, lp *int, cp *int, p *binary.BinaryProtocol, recurse bool, opts *Options, wireType proto.WireType) (*PathNode, error) {
	var con = *in
	var l = *lp
	guardPathNodeSlice(&con, l) // extend cap of con
	if l >= len(con) {
		con = con[:l+1]
	}
	v := &con[l]
	l += 1

	start := p.Read
	
	if e := p.Skip(wireType, opts.UseNativeSkip); e != nil {
		return nil, errNode(meta.ErrRead, "", e)
	}
	v.Node = self.slice(start, p.Read, proto.UNKNOWN, 0, 0)

	*in = con
	*lp = l
	*cp = cap(con)
	return v, nil
}


func (self *PathNode) handleListChild(in *[]PathNode, lp *int, cp *int, p *binary.BinaryProtocol, recurse bool, opts *Options, desc *proto.FieldDescriptor) (*PathNode, error) {
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
	kind := (*desc).Kind()
	et := proto.FromProtoKindToType(kind, false, false)

	if recurse && (et.IsComplex() && opts.NotScanParentNode) {
		v.Node = Node{
			t: et,
			l: 0,
			v: unsafe.Pointer(uintptr(self.Node.v) + uintptr(start)),
		}
	} else {
		if e := p.Skip(proto.Kind2Wire[kind], opts.UseNativeSkip); e != nil {
			return nil, errNode(meta.ErrRead, "", e)
		}
		v.Node = self.slice(start, p.Read, et, 0, 0)
	}

	if recurse && et.IsComplex() {
		p.Buf = p.Buf[start:]
		p.Read = 0
		messageLen := 0
		parentDesc := (*desc).(proto.Descriptor)
		if et == proto.MESSAGE {
			parentDesc = (*desc).Message().(proto.Descriptor)
			var err error
			messageLen, err = p.ReadLength()
			if messageLen<= 0 || err != nil {
				panic("read message length failed")
			}
		}

		if err := v.scanChildren(p, recurse, opts, &parentDesc, messageLen); err != nil {
			return nil, err
		}
		p.Buf = buf
		p.Read = start + p.Read
	}

	*in = con
	*lp = l
	*cp = cap(con)
	return v, nil
}

// Load loads self's all children ( and children's children if recurse is true) into self.Next,
// no matter whether self.Next is empty or set before (will be reset).
// NOTICE: if opts.NotScanParentNode is true, the parent nodes (PathNode.Node) of complex (map/list/struct) type won't be assgined data
func (self *PathNode) Load(recurse bool, opts *Options, desc *proto.Descriptor) error {
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
	return self.scanChildren(&p, recurse, opts, desc, len(p.Buf))
}


func GetDescByPath(rootDesc *proto.MessageDescriptor, pathes ...Path) (ret *proto.FieldDescriptor, err error) {
	var desc *proto.FieldDescriptor
	ret = desc
	for i, p := range pathes {
		if i == 0 {
			switch p.Type() {
			case PathFieldId:
				f := (*rootDesc).Fields().ByNumber(p.id())
				if f == nil {
					return nil, errNode(meta.ErrUnknownField, fmt.Sprintf("unknown field %d", p.id()), nil)
				}
				desc = &f
			case PathFieldName:
				f := (*rootDesc).Fields().ByName(proto.FieldName(p.str()))
				if f == nil {
					return nil, errNode(meta.ErrUnknownField, fmt.Sprintf("unknown field %s", p.str()), nil)
				}
				desc = &f
			default:
				return nil, errNode(meta.ErrInvalidParam, "", nil)
			}
		} else {
			kind := (*desc).Kind()
			et := proto.FromProtoKindToType(kind,(*desc).IsList(),(*desc).IsMap())
			switch et {
			case proto.MESSAGE:
				switch p.Type() {
				case PathFieldId:
					f := (*desc).Message().Fields().ByNumber(p.id())
					if f == nil {
						return nil, errNode(meta.ErrUnknownField, fmt.Sprintf("unknown field %d", p.id()), nil)
					}
					desc = &f
				case PathFieldName:
					f := (*desc).Message().Fields().ByName(proto.FieldName(p.str()))
					if f == nil {
						return nil, errNode(meta.ErrUnknownField, fmt.Sprintf("unknown field %s", p.str()), nil)
					}
					desc = &f
				default:
					return nil, errNode(meta.ErrInvalidParam, "", nil)
				}
			case proto.MAP:
				valueDesc := (*desc).MapValue()
				desc = &valueDesc
			// if LIST keep the same desc
			default:
				return nil, errNode(meta.ErrInvalidParam, "", nil)
			}
		}

		ret = desc
		if ret == nil {
			return nil, errNode(meta.ErrNotFound, "", err)
		}
	}
	return
}