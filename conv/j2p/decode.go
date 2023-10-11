package j2p

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"unsafe"

	"github.com/bytedance/sonic/ast"
	"github.com/chenzhuoyu/base64x"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
)

// memory resize factor
const (
	// new = old + old >> growSliceFactor
	growStkFactory = 1

	defaultStkDepth = 256
	nilStkType uint8 = 0
	objStkType uint8 = 1
	arrStkType uint8 = 2
	mapStkType uint8 = 3
)

var (
	vuPool = sync.Pool{
		New: func() interface{} {
			return &VisitorUserNode{
				sp:  0,
				p:   binary.NewBinaryProtocolBuffer(),
				stk: make([]VisitorUserNodeStack, defaultStkDepth),
			}
		},
	}
)

// NewVisitorUserNode get a new VisitorUserNode from sync.Pool
func NewVisitorUserNode(buf []byte) *VisitorUserNode {
	vu := vuPool.Get().(*VisitorUserNode)
	vu.p.Buf = buf
	return vu
}

// NewVisitorUserNode gets a new VisitorUserNode from sync.Pool
// and reuse the buffer in pool
func NewVisitorUserNodeBuffer() *VisitorUserNode {
	vu := vuPool.Get().(*VisitorUserNode)
	return vu
}

// FreeVisitorUserNode resets the buffer and puts the VisitorUserNode back to sync.Pool
func FreeVisitorUserNodePool(vu *VisitorUserNode) {
	vu.Reset()
	vuPool.Put(vu)
}

// Recycle put the VisitorUserNode back to sync.Pool
func (self *VisitorUserNode) Recycle() {
	self.Reset()
	vuPool.Put(self)
}

// Reset resets the buffer and read position
func (self *VisitorUserNode) Reset() {
	self.sp = 0
	self.p.Reset()
	for i := 0; i < len(self.stk); i++ {
		self.stk[i].Reset()
	}
}

/** p use to encode pbEncode
 *  desc represent fieldDescriptor
 *  pos is used when encode message\mapValue\unpackedList
 */
type VisitorUserNode struct {
	stk             []VisitorUserNodeStack
	sp              uint8
	p               *binary.BinaryProtocol
	globalFieldDesc *proto.FieldDescriptor
}

// keep hierarchy of Array and Object, arr represent current is List, obj represent current is Map/Object
// objStkType —— Message、arrStkType —— List、mapStkType —— Map
type VisitorUserNodeStack struct {
	typ       uint8
	state     visitorUserNodeState
}

func (stk *VisitorUserNodeStack) Reset() {
	stk.typ = nilStkType
	stk.state.lenPos = -1
	stk.state.msgDesc = nil
	stk.state.fieldDesc = nil
}

func (self *VisitorUserNode) Push(isMap bool, isObj bool, isList bool, desc *proto.FieldDescriptor, pos int) error {
	err := self.incrSP()
	self.stk[self.sp].state = visitorUserNodeState{
		msgDesc:   nil,
		fieldDesc: desc,
		lenPos:    pos,
	}
	if isMap {
		self.stk[self.sp].typ = mapStkType
	} else if isObj {
		self.stk[self.sp].typ = objStkType
	} else {
		self.stk[self.sp].typ = arrStkType
	}
	return err
}

func (self *VisitorUserNode) Pop() {
	self.stk[self.sp].Reset()
	self.sp--
}

// record descriptor、preWrite lenPos
type visitorUserNodeState struct {
	msgDesc   *proto.MessageDescriptor
	fieldDesc *proto.FieldDescriptor
	lenPos    int
}

func (self *VisitorUserNode) Decode(bytes []byte, desc *proto.Descriptor) ([]byte, error) {
	// init initial visitorUserNodeState
	switch (*desc).(type) {
	case proto.MessageDescriptor:
		convDesc := (*desc).(proto.MessageDescriptor)
		self.stk[self.sp].state = visitorUserNodeState{msgDesc: &convDesc, fieldDesc: nil, lenPos: -1}
	case proto.FieldDescriptor:
		convDesc := (*desc).(proto.FieldDescriptor)
		self.stk[self.sp].state = visitorUserNodeState{msgDesc: nil, fieldDesc: &convDesc, lenPos: -1}
	}
	str := rt.Mem2Str(bytes)
	if err := ast.Preorder(str, self, nil); err != nil {
		return nil, err
	}
	return self.result()
}

func (self *VisitorUserNode) result() ([]byte, error) {
	if self.sp != 0 {
		return nil, fmt.Errorf("incorrect sp: %d", self.sp)
	}
	return self.p.RawBuf(), nil
}

func (self *VisitorUserNode) incrSP() error {
	self.sp++
	if self.sp == 0 {
		return fmt.Errorf("reached max depth: %d", len(self.stk))
	}
	return nil
}

func (self *VisitorUserNode) OnNull() error {
	// self.stk[self.sp].val = &visitorUserNull{}
	if err := self.incrSP(); err != nil {
		return err
	}
	return self.onValueEnd()
}

func (self *VisitorUserNode) OnBool(v bool) error {
	var err error
	if self.globalFieldDesc == nil {
		return newError(meta.ErrConvert, "self.globalFieldDescriptor is nil, type Onbool", nil)
	}
	// TODO
	if err = self.p.AppendTagByDesc(self.globalFieldDesc); err != nil {
		return err
	}
	if err = self.p.WriteBool(v); err != nil {
		return err
	}
	return self.onValueEnd()
}

func (self *VisitorUserNode) OnString(v string) error {
	var err error
	top := self.stk[self.sp].state.fieldDesc
	fieldDesc := self.globalFieldDesc
	if fieldDesc == nil && top != nil && (*top).IsList() {
		fieldDesc = top
	}

	if err = self.p.AppendTagByDesc(fieldDesc); err != nil {
		return err
	}
	// convert string、bytesType
	switch (*fieldDesc).Kind() {
	case proto.BytesKind:
		bytesData, err := base64x.StdEncoding.DecodeString(v)
		if err = self.p.WriteBytes(bytesData); err != nil {
			return err
		}
	case proto.StringKind:
		if err = self.p.WriteString(v); err != nil {
			return err
		}
	default:
		return newError(meta.ErrDismatchType, "param isn't stringType", nil)
	}
	if self.globalFieldDesc != nil {
		return self.onValueEnd()
	}
	return err
}

func (self *VisitorUserNode) OnInt64(v int64, n json.Number) error {
	var err error
	top := self.stk[self.sp]
	fieldDesc := self.globalFieldDesc
	// case List<int>
	if self.globalFieldDesc == nil && top.typ == arrStkType {
		fieldDesc = top.state.fieldDesc
	}

	// packed list no need to write tag
	if !(*fieldDesc).IsList() {
		if err = self.p.AppendTagByDesc(fieldDesc); err != nil {
			return err
		}
	}

	switch (*fieldDesc).Kind() {
	case proto.Int32Kind, proto.Sint32Kind, proto.Sfixed32Kind, proto.Fixed32Kind:
		convertData := *(*int32)(unsafe.Pointer(&v))

		if err = self.p.WriteI32(convertData); err != nil {
			return err
		}
	case proto.Uint32Kind:
		convertData := *(*uint32)(unsafe.Pointer(&v))
		if err = self.p.WriteUint32(convertData); err != nil {
			return err
		}
	case proto.Uint64Kind:
		convertData := *(*uint64)(unsafe.Pointer(&v))
		if err = self.p.WriteUint64(convertData); err != nil {
			return err
		}
	case proto.Int64Kind:
		if err = self.p.WriteI64(v); err != nil {
			return err
		}
	default:
		return newError(meta.ErrDismatchType, "param isn't intType", nil)
	}
	// globalFieldDesc must belong to MessageDescriptor
	if self.globalFieldDesc != nil {
		err = self.onValueEnd()
	}
	return err
}

// todo when list
func (self *VisitorUserNode) OnFloat64(v float64, n json.Number) error {
	var err error
	top := self.stk[self.sp]
	fieldDesc := self.globalFieldDesc
	// case List<float>
	if self.globalFieldDesc == nil && top.typ == arrStkType {
		fieldDesc = top.state.fieldDesc
	}
	switch (*fieldDesc).Kind() {
	case proto.FloatKind:
		convertData := *(*float32)(unsafe.Pointer(&v))
		if err = self.p.AppendTagByDesc(fieldDesc); err != nil {
			return err
		}
		if err = self.p.WriteFloat(convertData); err != nil {
			return err
		}
	case proto.DoubleKind:
		convertData := *(*float64)(unsafe.Pointer(&v))
		if err = self.p.AppendTagByDesc(fieldDesc); err != nil {
			return err
		}
		if err = self.p.WriteDouble(convertData); err != nil {
			return err
		}
	default:
		return newError(meta.ErrDismatchType, "param isn't floatType", nil)
	}
	if self.globalFieldDesc != nil {
		err = self.onValueEnd()
	}
	return err
}

func (self *VisitorUserNode) OnObjectBegin(capacity int) error {
	var err error
	fieldDesc := self.globalFieldDesc
	top := self.stk[self.sp]
	curNodeLenPos := -1

	// case List<Message>
	if self.globalFieldDesc == nil && top.typ == arrStkType {
		fieldDesc = top.state.fieldDesc
	}
	
	if fieldDesc != nil {
		if (*fieldDesc).IsMap() {
			// case Map, push MapDesc
			if err = self.Push(true, false, false, fieldDesc, curNodeLenPos); err != nil {
				return err
			}
		} else {
			// case Message, encode Tag、Len
			if err = self.p.AppendTag(proto.Number((*fieldDesc).Number()), proto.BytesType); err != nil {
				return meta.NewError(meta.ErrWrite, "append prefix tag failed", nil)
			}
			self.p.Buf, curNodeLenPos = binary.AppendSpeculativeLength(self.p.Buf)
			if err = self.Push(false, true, false, fieldDesc, curNodeLenPos); err != nil {
				return err
			}
		}
	}
	return err
}

func (self *VisitorUserNode) DecodeMapKey(key string, mapKeyDesc *proto.FieldDescriptor) error {
	switch (*mapKeyDesc).Kind() {
	case proto.Int32Kind, proto.Sint32Kind, proto.Sfixed32Kind, proto.Fixed32Kind:
		t, _ := strconv.ParseInt(key, 10, 32)
		if err := self.p.WriteI32(int32(t)); err != nil {
			return err
		}
	case proto.Uint32Kind:
		t, _ := strconv.ParseInt(key, 10, 32)
		if err := self.p.WriteUint32(uint32(t)); err != nil {
			return err
		}
	case proto.Uint64Kind:
		t, _ := strconv.ParseInt(key, 10, 64)
		if err := self.p.WriteUint64(uint64(t)); err != nil {
			return err
		}
	case proto.Int64Kind:
		t, _ := strconv.ParseInt(key, 10, 64)
		if err := self.p.WriteI64(int64(t)); err != nil {
			return err
		}
	case proto.BoolKind:
		t, _ := strconv.ParseBool(key)
		if err := self.p.WriteBool(t); err != nil {
			return err
		}
	case proto.StringKind:
		if err := self.p.WriteString(key); err != nil {
			return err
		}
	default:
		return newError(meta.ErrDismatchType, "invalid mapKeyDescriptor Type", nil)
	}

	return nil
}

func (self *VisitorUserNode) OnObjectKey(key string) error {
	var err error
	var top *VisitorUserNodeStack
	var curDesc proto.FieldDescriptor
	curNodeLenPos := -1

	// get stack top, and recognize currentDescriptor type
	top = &self.stk[self.sp]
	if top.state.msgDesc != nil {
		// first hierarchy
		rootDesc := top.state.msgDesc
		curDesc = (*rootDesc).Fields().ByJSONName(key)
	} else {
		fieldDesc := top.state.fieldDesc
		// case MessageField
		if top.typ == objStkType {
			curDesc = (*fieldDesc).Message().Fields().ByJSONName(key)
		} else if top.typ == mapStkType {
			// case MapKey, write PairTag、PairLen、MapKeyTag、MapKeyLen、MapKeyData, push MapDesc into stack

			// encode PairTag、PairLen
			self.p.AppendTag((*top.state.fieldDesc).Number(), proto.BytesType)
			self.p.Buf, curNodeLenPos = binary.AppendSpeculativeLength(self.p.Buf)

			// encode MapKeyTag、MapKeyLen、MapKeyData
			mapKeyDesc := (*fieldDesc).MapKey()
			if err := self.p.AppendTagByDesc(&mapKeyDesc); err != nil {
				return newError(meta.ErrUnsupportedType, "unsatfisied mapKeyDescriptor Type", err)
			}

			if err := self.DecodeMapKey(key, &mapKeyDesc); err != nil {
				return newError(meta.ErrUnsupportedType, "unsatfisied mapKeyDescriptor Type", err)
			}
			
			// push MapDesc into stack
			if err = self.Push(true, false, false, top.state.fieldDesc, curNodeLenPos); err != nil {
				return err
			}
			curDesc = (*fieldDesc).MapValue()
		} else if top.typ == arrStkType {
			// case List
			curDesc = *top.state.fieldDesc
		}
	}
	self.globalFieldDesc = &curDesc
	return err
}

func (self *VisitorUserNode) OnObjectEnd() error {
	top := &self.stk[self.sp]
	// root layer no need to write tag and len
	if top.state.lenPos != -1 {
		self.p.Buf = binary.FinishSpeculativeLength(self.p.Buf, top.state.lenPos)
	}
	return self.onValueEnd()
}

func (self *VisitorUserNode) OnArrayBegin(capacity int) error {
	var err error
	curNodeLenPos := -1
	if self.globalFieldDesc != nil {
		// PackedList: encode Tag、Len
		if (*self.globalFieldDesc).IsPacked() {
			if err = self.p.AppendTag(proto.Number((*self.globalFieldDesc).Number()), proto.BytesType); err != nil {
				return meta.NewError(meta.ErrWrite, "append prefix tag failed", nil)
			}
			self.p.Buf, curNodeLenPos = binary.AppendSpeculativeLength(self.p.Buf)
		}
		if err = self.Push(false, false, true, self.globalFieldDesc, curNodeLenPos); err != nil {
			return err
		}
	}
	return err
}

func (self *VisitorUserNode) OnArrayEnd() error {
	var top *VisitorUserNodeStack
	top = &self.stk[self.sp]
	// case PackedList
	if (top.state.lenPos != -1) && (*top.state.fieldDesc).IsPacked() {
		self.p.Buf = binary.FinishSpeculativeLength(self.p.Buf, top.state.lenPos)
	}
	return self.onValueEnd()
}

func (self *VisitorUserNode) onValueEnd() error {
	if self.sp == 0 {
		return nil
	}
	top := self.stk[self.sp]

	// basic Type
	if self.globalFieldDesc != nil {
		self.globalFieldDesc = nil
		// basic Type belong to MapValue
		// rewrite pairlen in one kv-item of map
		if top.typ == mapStkType {
			self.p.Buf = binary.FinishSpeculativeLength(self.p.Buf, top.state.lenPos)
			self.Pop()
		}
		return nil
	}

	// complexType Type
	if top.typ == objStkType {
		self.Pop()
		// Message belong to Map<int, Message>、Message{Message}
		ntop := self.stk[self.sp]
		if ntop.typ == mapStkType {
			self.p.Buf = binary.FinishSpeculativeLength(self.p.Buf, ntop.state.lenPos)
			self.Pop()
		}
	} else if top.typ == arrStkType {
		self.Pop()
	} else if top.typ == mapStkType {
		self.Pop()
	} else {
		return newError(meta.ErrWrite, "disMatched ValueEnd", nil)
	}
	return nil
}
