package generic

import (
	"sync"
	"unsafe"

	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/proto"
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
	t PathType
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
