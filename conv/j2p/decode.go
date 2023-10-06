package j2p

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"unsafe"

	"github.com/bytedance/sonic/ast"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type visitorUserNode interface {
	// add descriptor
	UserNode(desc *proto.FieldDescriptor) interface{}
}

type (
	visitorUserNull    struct{}
	visitorUserBool    struct{ Value bool }
	visitorUserInt64   struct{ Value int64 }
	visitorUserFloat64 struct{ Value float64 }
	visitorUserString  struct{ Value string }
	visitorUserObject  struct{ Value map[string]visitorUserNode }
	visitorUserArray   struct{ Value []visitorUserNode }
)

// convert visitorUserNode dataFormat
func (self *visitorUserNull) UserNode(desc *proto.FieldDescriptor) interface{} {
	return nil
}

func (self *visitorUserBool) UserNode(desc *proto.FieldDescriptor) interface{} {
	return nil
}

// pay attention to digital precision
func (self *visitorUserInt64) UserNode(desc *proto.FieldDescriptor) interface{} {
	return nil
}

func (self *visitorUserFloat64) UserNode(desc *proto.FieldDescriptor) interface{} {
	return nil
}

func (self *visitorUserString) UserNode(desc *proto.FieldDescriptor) interface{} {
	return nil
}

func (self *visitorUserObject) UserNode(desc *proto.FieldDescriptor) interface{} {
	return nil
}

func (self *visitorUserArray) UserNode(desc *proto.FieldDescriptor) interface{} {
	return nil
}

/** p use to encode pbEncode
 *  desc represent fieldDescriptor
 *  pos is used when encode message\mapValue\unpackedList
 */
type visitorUserNodeVisitorDecoder struct {
	stk visitorUserNodeStack
	sp  uint8
	p   *binary.BinaryProtocol
}

// keep hierarchy of Array and Object
type visitorUserNodeStack = [256]struct {
	val   visitorUserNode
	obj   map[string]visitorUserNode
	arr   []visitorUserNode
	key   string
	state visitorUserNodeState
}

// record descriptor、preWrite lenPos
type visitorUserNodeState struct {
	desc   *proto.Descriptor
	lenPos int
}

// init visitorUserNodeVisitorDecoder
func (self *visitorUserNodeVisitorDecoder) Reset(p *binary.BinaryProtocol) {
	self.stk = visitorUserNodeStack{}
	self.sp = 0
	self.p = p
}

func (self *visitorUserNodeVisitorDecoder) Decode(str string, desc *proto.Descriptor) ([]byte, error) {
	// init initial visitorUserNodeState
	self.stk[self.sp].state = visitorUserNodeState{desc: desc, lenPos: -1}
	if err := ast.Preorder(str, self, nil); err != nil {
		return nil, err
	}
	return self.result()
}

func (self *visitorUserNodeVisitorDecoder) result() ([]byte, error) {
	if self.sp != 1 {
		return nil, fmt.Errorf("incorrect sp: %d", self.sp)
	}
	return self.p.RawBuf(), nil
}

func (self *visitorUserNodeVisitorDecoder) incrSP() error {
	self.sp++
	if self.sp == 0 {
		return fmt.Errorf("reached max depth: %d", len(self.stk))
	}
	return nil
}

func (self *visitorUserNodeVisitorDecoder) OnNull() error {
	self.stk[self.sp].val = &visitorUserNull{}
	if err := self.incrSP(); err != nil {
		return err
	}
	return self.onValueEnd()
}

func (self *visitorUserNodeVisitorDecoder) OnBool(v bool) error {
	var err error
	self.stk[self.sp].val = &visitorUserBool{Value: v}
	curDesc := self.stk[self.SearchPrevStateNodeIndex()].state.desc
	convertDesc := (*curDesc).(proto.FieldDescriptor)
	if err = self.p.WriteAnyWithDesc(&convertDesc, v, false, false, true); err != nil {
		return err
	}
	if err = self.incrSP(); err != nil {
		return err
	}
	return self.onValueEnd()
}

func (self *visitorUserNodeVisitorDecoder) OnString(v string) error {
	var err error
	curDesc := self.stk[self.SearchPrevStateNodeIndex()].state.desc
	convertDesc := (*curDesc).(proto.FieldDescriptor)
	convertData := self.EncodePbData(convertDesc, v)
	self.stk[self.sp].val = &visitorUserString{Value: v}
	if err = self.p.WriteAnyWithDesc(&convertDesc, convertData, false, false, true); err != nil {
		return err
	}
	if err = self.incrSP(); err != nil {
		return err
	}
	return self.onValueEnd()
}

func (self *visitorUserNodeVisitorDecoder) OnInt64(v int64, n json.Number) error {
	var err error
	curDesc := self.stk[self.SearchPrevStateNodeIndex()].state.desc
	convertDesc := (*curDesc).(proto.FieldDescriptor)
	convertData := self.EncodePbData(convertDesc, v)
	self.stk[self.sp].val = &visitorUserInt64{Value: v}
	if convertDesc.IsList() {
		if err = self.p.WriteBaseTypeWithDesc(&convertDesc, convertData, false, false, true); err != nil {
			return err
		}
	} else {
		if err = self.p.WriteAnyWithDesc(&convertDesc, convertData, false, false, true); err != nil {
			return nil
		}
	}
	if err = self.incrSP(); err != nil {
		return err
	}
	return self.onValueEnd()
}

func (self *visitorUserNodeVisitorDecoder) OnFloat64(v float64, n json.Number) error {
	var err error
	curDesc := self.stk[self.SearchPrevStateNodeIndex()].state.desc
	convertDesc := (*curDesc).(proto.FieldDescriptor)
	convertData := self.EncodePbData(convertDesc, v)
	self.stk[self.sp].val = &visitorUserFloat64{Value: v}
	if err = self.p.WriteAnyWithDesc(&convertDesc, convertData, false, false, true); err != nil {
		return nil
	}
	if err = self.incrSP(); err != nil {
		return err
	}
	return self.onValueEnd()
}

// write data into pbEncode, pay attention to dataType change
func (self *visitorUserNodeVisitorDecoder) EncodePbData(fieldDesc proto.FieldDescriptor, val interface{}) interface{} {
	var value interface{}
	switch fieldDesc.Kind() {
	case proto.Int32Kind, proto.Sint32Kind, proto.Sfixed32Kind, proto.Fixed32Kind:
		num := val.(int64)
		value = *(*int32)(unsafe.Pointer(&num))
	case proto.Uint32Kind:
		num := val.(int64)
		value = *(*uint32)(unsafe.Pointer(&num))
	case proto.Uint64Kind:
		num := val.(int64)
		value = *(*uint64)(unsafe.Pointer(&num))
	case proto.Int64Kind:
		num := val.(int64)
		value = *(*int64)(unsafe.Pointer(&num))
	case proto.FloatKind:
		num := val.(float64)
		value = *(*float32)(unsafe.Pointer(&num))
	case proto.DoubleKind:
		num := val.(float64)
		value = *(*float64)(unsafe.Pointer(&num))
	case proto.StringKind:
		value = val.(string)
	case proto.BytesKind:
		// encoding/json encode byte[] use base64 format
		str := val.(string)
		base64DecodeData, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			panic(err)
		}
		value = []byte(string(base64DecodeData))
	}
	return value
}

func (self *visitorUserNodeVisitorDecoder) OnObjectBegin(capacity int) error {
	self.stk[self.sp].obj = make(map[string]visitorUserNode, capacity)
	lastDescIdx := self.SearchPrevStateNodeIndex()
	if lastDescIdx > 0 {
		var elemLenPos int
		objDesc := self.stk[lastDescIdx].state.desc
		convertDesc := (*objDesc).(proto.FieldDescriptor)
		// case List<Message>/Map<xxx,Message>, insert MessageTag、MessageLen, store innerMessageDesc into stack
		if convertDesc != nil && convertDesc.Message() != nil {
			parentMsg := convertDesc.ContainingMessage()
			if convertDesc.IsList() || parentMsg.IsMapEntry() {
				innerElemDesc := (*objDesc).(proto.FieldDescriptor)
				if err := self.p.AppendTag(proto.Number(innerElemDesc.Number()), proto.BytesType); err != nil {
					return meta.NewError(meta.ErrWrite, "append prefix tag failed", nil)
				}
				self.p.Buf, elemLenPos = binary.AppendSpeculativeLength(self.p.Buf)
				self.stk[self.sp].state = visitorUserNodeState{
					desc:   objDesc,
					lenPos: elemLenPos,
				}
			}
		}
	}
	return self.incrSP()
}

func (self *visitorUserNodeVisitorDecoder) OnObjectKey(key string) error {
	self.stk[self.sp].key = key
	var rootDesc proto.MessageDescriptor
	var curDesc proto.FieldDescriptor
	var preNodeState visitorUserNodeState
	// search preNodeState
	i := self.SearchPrevStateNodeIndex()
	preNodeState = self.stk[i].state

	// recognize descriptor type
	if i == 0 {
		rootDesc = ((*preNodeState.desc).(proto.MessageDescriptor))
	} else {
		curDesc = ((*preNodeState.desc).(proto.FieldDescriptor))
	}
	// initial hierarchy
	if rootDesc != nil {
		curDesc = rootDesc.Fields().ByJSONName(key)
		// complex structure, get inner fieldDesc by jsonname
	} else if curDesc != nil && curDesc.Message() != nil {
		if curDesc.IsList() && !curDesc.IsPacked() {
			// case Unpacked List(inner is Message)
			if fieldDesc := curDesc.Message().Fields().ByJSONName(key); fieldDesc != nil {
				curDesc = fieldDesc
			}
		} else if curDesc.IsMap() {
			// case Map
		} else if !curDesc.IsPacked() {
			// case Message
			curDesc = curDesc.Message().Fields().ByJSONName(key)
		}
	}

	if curDesc != nil {
		curNodeLenPos := -1
		convertDescriptor := curDesc.(proto.Descriptor)
		// case PackedList/Message, need to write prefix Tag、Len, and store to stack
		if (curDesc.IsList() && curDesc.IsPacked()) || (curDesc.Kind() == proto.MessageKind && !curDesc.IsMap() && !curDesc.IsList()) {
			if err := self.p.AppendTag(proto.Number(curDesc.Number()), proto.BytesType); err != nil {
				return meta.NewError(meta.ErrWrite, "append prefix tag failed", nil)
			}
			self.p.Buf, curNodeLenPos = binary.AppendSpeculativeLength(self.p.Buf)
		} else {
			// mapkey but not mapName, write pairTag、pairLen、mapKeyTag、mapKeyLen、mapData, may have bug here
			if curDesc.IsMap() && curDesc.JSONName() != key {
				// pay attention to Key type conversion
				var keyData interface{}
				mapKeyDesc := curDesc.MapKey()
				switch mapKeyDesc.Kind() {
				case protoreflect.Kind(proto.Int32Kind), protoreflect.Kind(proto.Sint32Kind), protoreflect.Kind(proto.Sfixed32Kind), protoreflect.Kind(proto.Fixed32Kind):
					t, _ := strconv.ParseInt(key, 10, 32)
					keyData = int32(t)
				case protoreflect.Kind(proto.Uint32Kind):
					t, _ := strconv.ParseInt(key, 10, 32)
					keyData = uint32(t)
				case protoreflect.Kind(proto.Uint64Kind):
					t, _ := strconv.ParseInt(key, 10, 64)
					keyData = uint64(t)
				case protoreflect.Kind(proto.Int64Kind):
					t, _ := strconv.ParseInt(key, 10, 64)
					keyData = t
				case protoreflect.Kind(proto.BoolKind):
					t, _ := strconv.ParseBool(key)
					keyData = t
				default:
					keyData = key
				}
				self.p.AppendTag(curDesc.Number(), proto.BytesType)
				self.p.Buf, curNodeLenPos = binary.AppendSpeculativeLength(self.p.Buf)
				if err := self.p.WriteAnyWithDesc(&mapKeyDesc, keyData, false, false, true); err != nil {
					return err
				}
				// store MapValueDesc into stack
				convertDescriptor = curDesc.MapValue().(proto.Descriptor)
			}
		}
		// store fieldDesc into stack
		self.stk[self.sp].state = visitorUserNodeState{
			desc:   &convertDescriptor,
			lenPos: curNodeLenPos,
		}
	}
	return self.incrSP()
}

// search the last StateNode which desc isn't empty
func (self *visitorUserNodeVisitorDecoder) SearchPrevStateNodeIndex() int {
	var i int
	for i = int(self.sp); i >= 0; i-- {
		if i == 0 || self.stk[i].state.desc != nil {
			break
		}
	}
	return i
}

func (self *visitorUserNodeVisitorDecoder) OnObjectEnd() error {
	self.stk[self.sp-1].val = &visitorUserObject{Value: self.stk[self.sp-1].obj}
	self.stk[self.sp-1].obj = nil

	// fill prefix_length when MessageEnd, may have problem with rootDesc
	parentDescIdx := self.SearchPrevStateNodeIndex()
	if parentDescIdx > 0 {
		parentNodeState := self.stk[parentDescIdx].state
		if parentNodeState.desc != nil && parentNodeState.lenPos != -1 {
			self.p.Buf = binary.FinishSpeculativeLength(self.p.Buf, parentNodeState.lenPos)
			self.stk[parentDescIdx].state.desc = nil
			self.stk[parentDescIdx].state.lenPos = -1
		}
	}
	return self.onValueEnd()
}

func (self *visitorUserNodeVisitorDecoder) OnArrayBegin(capacity int) error {
	self.stk[self.sp].arr = make([]visitorUserNode, 0, capacity)
	return self.incrSP()
}

func (self *visitorUserNodeVisitorDecoder) OnArrayEnd() error {
	self.stk[self.sp-1].val = &visitorUserArray{Value: self.stk[self.sp-1].arr}
	self.stk[self.sp-1].arr = nil

	// case arrayEnd, fill arrayPrefixLen
	var parentNodeState visitorUserNodeState
	if self.sp >= 2 {
		parentNodeState = self.stk[self.sp-2].state
		if parentNodeState.lenPos != -1 {
			// case Packedlist, fill prefixLength
			self.p.Buf = binary.FinishSpeculativeLength(self.p.Buf, parentNodeState.lenPos)
			self.stk[self.sp-2].state.desc = nil
			self.stk[self.sp-2].state.lenPos = -1
		}
	}
	return self.onValueEnd()
}

func (self *visitorUserNodeVisitorDecoder) onValueEnd() error {
	if self.sp == 1 {
		return nil
	}
	// [..., Array, Value, sp]
	if self.stk[self.sp-2].arr != nil {
		self.stk[self.sp-2].arr = append(self.stk[self.sp-2].arr, self.stk[self.sp-1].val)
		self.sp--
		return nil
	}
	// [..., Object, ObjectKey, Value, sp]
	if self.stk[self.sp-3].obj != nil {
		self.stk[self.sp-3].obj[self.stk[self.sp-2].key] = self.stk[self.sp-1].val
		self.sp -= 2

		// case MapValueEnd, fill pairLength, need to exclude Message[Message]
		if self.sp >= 2 {
			parentNodeState := self.stk[self.sp-2].state
			if parentNodeState.desc != nil {
				if (*parentNodeState.desc).(proto.FieldDescriptor).IsMap() {
					// case mapValueEnd, fill pairLength
					self.p.Buf = binary.FinishSpeculativeLength(self.p.Buf, self.stk[self.sp].state.lenPos)
				}
			}
		}

		self.stk[self.sp].state.desc = nil
		self.stk[self.sp].state.lenPos = -1
	}
	return nil
}
