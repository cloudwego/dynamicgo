package j2p

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/bytedance/sonic/ast"
	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
)

// memory resize factor
const (
	defaultStkDepth       = 256
	nilStkType      uint8 = 0
	objStkType      uint8 = 1
	arrStkType      uint8 = 2
	mapStkType      uint8 = 3
)

const (
	mapkeyFieldNumber   = 1
	mapValueFieldNumber = 2
)

var (
	vuPool = sync.Pool{
		New: func() interface{} {
			return &visitorUserNode{
				sp:  0,
				stk: make([]visitorUserNodeStack, defaultStkDepth),
			}
		},
	}
)

func decodeBinary(val string) ([]byte, error) {
	return decodeBase64(val)
}

// NewvisitorUserNode gets a new visitorUserNode from sync.Pool
// and reuse the buffer in pool
func newVisitorUserNodeBuffer() *visitorUserNode {
	vu := vuPool.Get().(*visitorUserNode)
	for i := 0; i < len(vu.stk); i++ {
		vu.stk[i].state.lenPos = -1
	}
	return vu
}

// FreevisitorUserNode resets the buffer and puts the visitorUserNode back to sync.Pool
func freeVisitorUserNodePool(vu *visitorUserNode) {
	vu.reset()
	vuPool.Put(vu)
}

// recycle put the visitorUserNode back to sync.Pool
func (self *visitorUserNode) recycle() {
	self.reset()
	vuPool.Put(self)
}

// reset resets the buffer and read position
func (self *visitorUserNode) reset() {
	self.sp = 0
	self.p = nil
	for i := 0; i < len(self.stk); i++ {
		self.stk[i].reset()
	}
	self.globalFieldDesc = nil
	self.opts = nil
	self.inskip = false
}

// visitorUserNode is used to store some conditional variables about Protobuf when parsing json
// Stk: Record MessageDescriptor and PrefixLen when parsing MessageType; Record FieldDescriptor and PairPrefixLen and MapKeyDescriptor when parsing MapType; Record FieldDescriptor and PrefixListLen when parsing List
// Sp: using to Record the level of current stack(stk)
// P: Output of Protobuf BinaryData
// GlobalFieldDesc：After parsing the FieldKey, save the FieldDescriptor
type visitorUserNode struct {
	sp              uint8
	inskip          bool
	stk             []visitorUserNodeStack
	p               *binary.BinaryProtocol
	globalFieldDesc *proto.FieldDescriptor
	opts            *conv.Options
}

// keep hierarchy when parsing, objStkType represents Message and arrStkType represents List and mapStkType represents Map
type visitorUserNodeStack struct {
	typ   uint8
	state visitorUserNodeState
}

func (stk *visitorUserNodeStack) reset() {
	stk.typ = nilStkType
	stk.state.lenPos = -1
	stk.state.msgDesc = nil
	stk.state.fieldDesc = nil
}

// push FieldDescriptor、position of prefixLen into stack
func (self *visitorUserNode) push(isMap bool, isObj bool, isList bool, desc *proto.FieldDescriptor, pos int) error {
	err := self.incrSP()
	t := nilStkType
	if isMap {
		t = mapStkType
	} else if isObj {
		t = objStkType
	} else if isList {
		t = arrStkType
	}

	self.stk[self.sp] = visitorUserNodeStack{typ: t, state: visitorUserNodeState{
		msgDesc:   nil,
		fieldDesc: desc,
		lenPos:    pos,
	}}
	return err
}

// After parsing MessageType/MapType/ListType, pop stack
func (self *visitorUserNode) pop() {
	self.stk[self.sp].reset()
	self.sp--
}

// combine of Descriptor and position of prefixLength
type visitorUserNodeState struct {
	msgDesc   *proto.MessageDescriptor
	fieldDesc *proto.FieldDescriptor
	lenPos    int
}

func (self *visitorUserNode) decode(bytes []byte, desc *proto.TypeDescriptor) ([]byte, error) {
	// init initial visitorUserNodeState, may be MessageDescriptor or FieldDescriptor
	switch desc.Type() {
	case proto.MESSAGE:
		convDesc := desc.Message()
		self.stk[self.sp].state = visitorUserNodeState{msgDesc: convDesc, fieldDesc: nil, lenPos: -1}
		self.stk[self.sp].typ = objStkType
	}
	str := rt.Mem2Str(bytes)
	if err := ast.Preorder(str, self, nil); err != nil {
		return nil, err
	}
	return self.result()
}

func (self *visitorUserNode) result() ([]byte, error) {
	if self.sp != 0 {
		return nil, fmt.Errorf("incorrect sp: %d", self.sp)
	}
	return self.p.RawBuf(), nil
}

func (self *visitorUserNode) incrSP() error {
	self.sp++
	if self.sp == 0 {
		return fmt.Errorf("reached max depth: %d", len(self.stk))
	}
	return nil
}

func (self *visitorUserNode) OnNull() error {
	if self.inskip {
		self.inskip = false
		return nil
	}
	// self.stk[self.sp].val = &visitorUserNull{}
	if err := self.incrSP(); err != nil {
		return err
	}
	return self.onValueEnd()
}

func (self *visitorUserNode) OnBool(v bool) error {
	if self.inskip {
		self.inskip = false
		return nil
	}

	var err error
	top := self.stk[self.sp]
	fieldDesc := self.globalFieldDesc
	// case PackedList(List bool), get fieldDescriptor from Stack
	if self.globalFieldDesc == nil && top.typ == arrStkType {
		fieldDesc = top.state.fieldDesc
	}

	// packed list no need to write tag
	if !fieldDesc.Type().IsList() {
		if err = self.p.AppendTagByKind(fieldDesc.Number(), fieldDesc.Kind()); err != nil {
			return err
		}
	}

	if err = self.p.WriteBool(v); err != nil {
		return err
	}
	// globalFieldDesc must belong to MessageDescriptor
	if self.globalFieldDesc != nil {
		err = self.onValueEnd()
	}
	return err
}

// Parse stringType/bytesType
func (self *visitorUserNode) OnString(v string) error {
	if self.inskip {
		self.inskip = false
		return nil
	}
	var err error
	top := self.stk[self.sp].state.fieldDesc
	fieldDesc := self.globalFieldDesc
	if fieldDesc == nil && top != nil && top.Type().IsList() {
		fieldDesc = top
	}

	if err = self.p.AppendTagByKind(fieldDesc.Number(), fieldDesc.Kind()); err != nil {
		return err
	}
	// convert string、bytesType
	switch fieldDesc.Kind() {
	case proto.BytesKind:
		// JSON format string data needs to be decoded via Base64x
		bytesData, err := decodeBinary(v)
		if err != nil {
			return err
		}
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

func (self *visitorUserNode) OnInt64(v int64, n json.Number) error {
	if self.inskip {
		self.inskip = false
		return nil
	}
	var err error
	top := self.stk[self.sp]
	fieldDesc := self.globalFieldDesc
	// case PackedList(List<int32/int64/...), get fieldDescriptor from Stack
	if self.globalFieldDesc == nil && top.typ == arrStkType {
		fieldDesc = top.state.fieldDesc
	}

	// packed list no need to write tag
	if !fieldDesc.Type().IsList() {
		if err = self.p.AppendTagByKind(fieldDesc.Number(), fieldDesc.Kind()); err != nil {
			return err
		}
	}

	switch fieldDesc.Kind() {
	case proto.Int32Kind:
		convertData := int32(v)
		if err = self.p.WriteInt32(convertData); err != nil {
			return err
		}
	case proto.Sint32Kind:
		convertData := int32(v)
		if err = self.p.WriteSint32(convertData); err != nil {
			return err
		}
	case proto.Sfixed32Kind:
		convertData := int32(v)
		if err = self.p.WriteSfixed32(convertData); err != nil {
			return err
		}
	case proto.Fixed32Kind:
		convertData := uint32(v)
		if err = self.p.WriteFixed32(convertData); err != nil {
			return err
		}
	case proto.Uint32Kind:
		convertData := uint32(v)
		if err = self.p.WriteUint32(convertData); err != nil {
			return err
		}
	case proto.Uint64Kind:
		convertData := uint64(v)
		if err = self.p.WriteUint64(convertData); err != nil {
			return err
		}
	case proto.Int64Kind:
		if err = self.p.WriteInt64(v); err != nil {
			return err
		}
	case proto.Sint64Kind:
		if err = self.p.WriteSint64(v); err != nil {
			return err
		}
	case proto.Sfixed64Kind:
		if err = self.p.WriteSfixed64(v); err != nil {
			return err
		}
	case proto.Fixed64Kind:
		convertData := uint64(v)
		if err = self.p.WriteFixed64(convertData); err != nil {
			return err
		}
	// cast int2float, int2double
	case proto.FloatKind:
		convertData := float32(v)
		if err = self.p.WriteFloat(convertData); err != nil {
			return err
		}
	case proto.DoubleKind:
		convertData := float64(v)
		if err = self.p.WriteDouble(convertData); err != nil {
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

func (self *visitorUserNode) OnFloat64(v float64, n json.Number) error {
	if self.inskip {
		self.inskip = false
		return nil
	}
	var err error
	top := self.stk[self.sp]
	fieldDesc := self.globalFieldDesc

	if self.globalFieldDesc == nil && top.typ == arrStkType {
		fieldDesc = top.state.fieldDesc
	}

	// packed list no need to write tag
	if !fieldDesc.Type().IsList() {
		if err = self.p.AppendTagByKind(fieldDesc.Number(), fieldDesc.Kind()); err != nil {
			return err
		}
	}

	switch fieldDesc.Kind() {
	case proto.FloatKind:
		convertData := float32(v)
		if err = self.p.WriteFloat(convertData); err != nil {
			return err
		}
	case proto.DoubleKind:
		convertData := v
		if err = self.p.WriteDouble(convertData); err != nil {
			return err
		}
	// cast double2int32, double2int64
	case proto.Int32Kind:
		convertData := int32(v)
		if err = self.p.WriteInt32(convertData); err != nil {
			return err
		}
	case proto.Int64Kind:
		convertData := int64(v)
		if err = self.p.WriteInt64(convertData); err != nil {
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

// Start Parsing JSONObject, which may correspond to Protobuf Map or Protobuf Message
//  1. Get fieldDescriptor from globalFieldDesc first, corresponding to Message/Map embedded in a Message
//  2. Then get fieldDesc from the top of stack, corresponding to Message embedded in List
//  3. When Field is Message type, encode Tag and PrefixLen, and push FieldDescriptor and PrefixLen to the stack.
//     When Field is Map type, only perform pressing stack operation.
func (self *visitorUserNode) OnObjectBegin(capacity int) error {
	if self.inskip {
		return ast.VisitOPSkip
	}
	var err error
	fieldDesc := self.globalFieldDesc
	top := self.stk[self.sp]
	curNodeLenPos := -1

	// case List<Message>, get fieldDescriptor
	if self.globalFieldDesc == nil && top.typ == arrStkType {
		fieldDesc = top.state.fieldDesc
	}

	if fieldDesc != nil {
		if fieldDesc.Type().IsMap() {
			// case Map, push MapDesc
			if err = self.push(true, false, false, fieldDesc, curNodeLenPos); err != nil {
				return err
			}
		} else {
			// case Message, encode Tag、PrefixLen, push MessageDesc、PrefixLen
			if err = self.p.AppendTag(fieldDesc.Number(), proto.BytesType); err != nil {
				return meta.NewError(meta.ErrWrite, "append prefix tag failed", nil)
			}
			self.p.Buf, curNodeLenPos = binary.AppendSpeculativeLength(self.p.Buf)
			if err = self.push(false, true, false, fieldDesc, curNodeLenPos); err != nil {
				return err
			}
		}
	}
	return err
}

// MapKey maybe int32/sint32/uint32/uint64 etc....
func (self *visitorUserNode) encodeMapKey(key string, t proto.Type) error {
	switch t {
	case proto.INT32:
		t, _ := strconv.ParseInt(key, 10, 32)
		if err := self.p.WriteInt32(int32(t)); err != nil {
			return err
		}
	case proto.UINT32:
		t, _ := strconv.ParseInt(key, 10, 32)
		if err := self.p.WriteUint32(uint32(t)); err != nil {
			return err
		}
	case proto.UINT64:
		t, _ := strconv.ParseInt(key, 10, 64)
		if err := self.p.WriteUint64(uint64(t)); err != nil {
			return err
		}
	case proto.INT64:
		t, _ := strconv.ParseInt(key, 10, 64)
		if err := self.p.WriteInt64(int64(t)); err != nil {
			return err
		}
	case proto.BOOL:
		t, _ := strconv.ParseBool(key)
		if err := self.p.WriteBool(t); err != nil {
			return err
		}
	case proto.STRING:
		if err := self.p.WriteString(key); err != nil {
			return err
		}
	default:
		return newError(meta.ErrDismatchType, "invalid mapKeyDescriptor Type", nil)
	}

	return nil
}

// Start Parsing JSONField, which may correspond to Protobuf MessageField or Protobuf MapKey
//  1. The initial layer obtains the MessageFieldDescriptor, saves it to globalFieldDesc and exits;
//  2. The inner layer performs different operations depending on the top-of-stack descriptor type
//     2.1 The top of the stack is MessageType, which exits after saving globalFieldDesc;
//     2.2 The top of the stack is MapType, write PairTag, PairLen, MapKey, MapKeyLen, and MapKeyData; save MapDesc、PairPrefixLen into stack; save MapValueDescriptor into globalFieldDesc;
//     2.3 The top of the stack is ListType, which exits after saving globalFieldDesc;
func (self *visitorUserNode) OnObjectKey(key string) error {
	var err error
	var top *visitorUserNodeStack
	var curDesc proto.FieldDescriptor
	curNodeLenPos := -1

	// get stack top, and recognize currentDescriptor type
	top = &self.stk[self.sp]
	if top.state.msgDesc != nil {
		// first hierarchy
		rootDesc := top.state.msgDesc
		fd := rootDesc.ByJSONName(key)
		if fd == nil {
			if self.opts.DisallowUnknownField {
				return newError(meta.ErrUnknownField, fmt.Sprintf("json key '%s' is unknown", key), err)
			} else {
				self.inskip = true
				return nil
			}
		}
		curDesc = *fd
	} else {
		fieldDesc := top.state.fieldDesc
		// case MessageField
		if top.typ == objStkType {
			fd := fieldDesc.Message().ByJSONName(key)
			if fd == nil {
				if self.opts.DisallowUnknownField {
					return newError(meta.ErrUnknownField, fmt.Sprintf("json key '%s' is unknown", key), err)
				} else {
					self.inskip = true
					return nil
				}
			}
			curDesc = *fd
		} else if top.typ == mapStkType {
			// case MapKey, write PairTag、PairLen、MapKeyTag、MapKeyLen、MapKeyData, push MapDesc into stack

			// encode PairTag、PairLen
			fd := top.state.fieldDesc
			if err := self.p.AppendTag(fd.Number(), proto.BytesType); err != nil {
				return newError(meta.ErrWrite, fmt.Sprintf("field '%s' encode pair tag faield", fd.Name()), err)
			}
			self.p.Buf, curNodeLenPos = binary.AppendSpeculativeLength(self.p.Buf)

			// encode MapKeyTag、MapKeyLen、MapKeyData
			mapKeyDesc := fieldDesc.MapKey()
			if err := self.p.AppendTag(mapkeyFieldNumber, mapKeyDesc.WireType()); err != nil {
				return newError(meta.ErrUnsupportedType, fmt.Sprintf("field '%s' encode map key tag faield", fd.Name()), err)
			}

			if err := self.encodeMapKey(key, mapKeyDesc.Type()); err != nil {
				return newError(meta.ErrUnsupportedType, fmt.Sprintf("field '%s' encode map key faield", fd.Name()), err)
			}

			// push MapDesc into stack
			if err = self.push(true, false, false, top.state.fieldDesc, curNodeLenPos); err != nil {
				return err
			}
			// save MapValueDesc into globalFieldDesc
			curDesc = *fieldDesc.Message().ByNumber(mapValueFieldNumber)
		} else if top.typ == arrStkType {
			// case List<Message>
			curDesc = *top.state.fieldDesc
		}
	}
	self.globalFieldDesc = &curDesc
	return err
}

// After parsing JSONObject, write back prefixLen of Message
func (self *visitorUserNode) OnObjectEnd() error {
	if self.inskip {
		self.inskip = false
		return nil
	}
	top := &self.stk[self.sp]
	// root layer no need to write tag and len
	if top.state.lenPos != -1 {
		self.p.Buf = binary.FinishSpeculativeLength(self.p.Buf, top.state.lenPos)
	}
	return self.onValueEnd()
}

// Start Parsing JSONArray, which may correspond to ProtoBuf PackedList、UnPackedList
// 1. If PackedList, Encode ListTag、PrefixLen
// 2. push ListDescriptor、PrefixLen(UnPackedList is -1) into stack
func (self *visitorUserNode) OnArrayBegin(capacity int) error {
	if self.inskip {
		return ast.VisitOPSkip
	}
	var err error
	curNodeLenPos := -1
	if self.globalFieldDesc != nil {
		// PackedList: encode Tag、Len
		if self.globalFieldDesc.Type().IsPacked() {
			if err = self.p.AppendTag(self.globalFieldDesc.Number(), proto.BytesType); err != nil {
				return meta.NewError(meta.ErrWrite, "append prefix tag failed", nil)
			}
			self.p.Buf, curNodeLenPos = binary.AppendSpeculativeLength(self.p.Buf)
		}
		if err = self.push(false, false, true, self.globalFieldDesc, curNodeLenPos); err != nil {
			return err
		}
	}
	return err
}

// After Parsing JSONArray, writing back PrefixLen If PackedList
func (self *visitorUserNode) OnArrayEnd() error {
	if self.inskip {
		self.inskip = false
		return nil
	}
	var top *visitorUserNodeStack
	top = &self.stk[self.sp]
	// case PackedList
	if (top.state.lenPos != -1) && top.state.fieldDesc.Type().IsPacked() {
		self.p.Buf = binary.FinishSpeculativeLength(self.p.Buf, top.state.lenPos)
	}
	return self.onValueEnd()
}

// After parsing one JSON field, maybe basicType(string/int/float/bool/bytes) or complexType(message/map/list)
// 1. When type is basicType, note that the format of Map<basicType, basicType>, needs to write back PariPrefixLen of the current pair;
// 2. When type is MessageType, current Message may belong to Map<any, Message> or Message{Message}, note that Map<int, Message>, needs to write back PairPrefixLen of current pair;
// 3. When type is MapType/ListType, only execute pop operation;
func (self *visitorUserNode) onValueEnd() error {
	if self.sp == 0 && self.globalFieldDesc == nil {
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
			self.pop()
		}
		return nil
	}

	// complexType Type
	if top.typ == objStkType {
		self.pop()
		// Message belong to Map<int, Message>、Message{Message}
		ntop := self.stk[self.sp]
		if ntop.typ == mapStkType {
			self.p.Buf = binary.FinishSpeculativeLength(self.p.Buf, ntop.state.lenPos)
			self.pop()
		}
	} else if top.typ == arrStkType {
		self.pop()
	} else if top.typ == mapStkType {
		self.pop()
	} else {
		return newError(meta.ErrWrite, "disMatched ValueEnd", nil)
	}
	return nil
}
