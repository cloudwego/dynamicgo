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
	return self.Value
}

// pay attention to digital precision
func (self *visitorUserInt64) UserNode(desc *proto.FieldDescriptor) interface{} {
	var value interface{}
	switch (*desc).Kind() {
	case protoreflect.Kind(proto.Int32Kind), protoreflect.Kind(proto.Sint32Kind), protoreflect.Kind(proto.Sfixed32Kind), protoreflect.Kind(proto.Fixed32Kind):
		value = *(*int32)(unsafe.Pointer(&self.Value))
	case protoreflect.Kind(proto.Uint32Kind):
		value = *(*uint32)(unsafe.Pointer(&self.Value))
	case protoreflect.Kind(proto.Uint64Kind):
		value = *(*uint64)(unsafe.Pointer(&self.Value))
	default:
		value = self.Value
	}
	return value
}

func (self *visitorUserFloat64) UserNode(desc *proto.FieldDescriptor) interface{} {
	var value interface{}
	if (*desc).Kind() == protoreflect.Kind(proto.FloatKind) {
		value = *(*float32)(unsafe.Pointer(&self.Value))
	} else {
		value = self.Value
	}
	return value
}

func (self *visitorUserString) UserNode(desc *proto.FieldDescriptor) interface{} {
	var value interface{}
	switch (*desc).Kind() {
	case protoreflect.Kind(proto.BytesKind):
		// encoding/json encode byte[] use base64 format
		base64DecodeData, err := base64.StdEncoding.DecodeString(self.Value)
		if err != nil {
			panic(err)
		}
		value = []byte(string(base64DecodeData))
	default:
		value = self.Value
	}
	return value
}

func (self *visitorUserObject) UserNode(desc *proto.FieldDescriptor) interface{} {
	return nil
}

func (self *visitorUserArray) UserNode(desc *proto.FieldDescriptor) interface{} {
	return nil
}

type visitorUserNodeVisitorDecoder struct {
	stk visitorUserNodeStack
	sp  uint8
}

// keep hierarchy of Array and Object
type visitorUserNodeStack = [256]struct {
	val visitorUserNode
	obj map[string]visitorUserNode
	arr []visitorUserNode
	key string
}

func (self *visitorUserNodeVisitorDecoder) Reset() {
	self.stk = visitorUserNodeStack{}
	self.sp = 0
}

func (self *visitorUserNodeVisitorDecoder) Decode(str string) (visitorUserNode, error) {
	if err := ast.Preorder(str, self, nil); err != nil {
		return nil, err
	}
	return self.result()
}

func (self *visitorUserNodeVisitorDecoder) result() (visitorUserNode, error) {
	if self.sp != 1 {
		return nil, fmt.Errorf("incorrect sp: %d", self.sp)
	}
	return self.stk[0].val, nil
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
	self.stk[self.sp].val = &visitorUserBool{Value: v}
	if err := self.incrSP(); err != nil {
		return err
	}
	return self.onValueEnd()
}

func (self *visitorUserNodeVisitorDecoder) OnString(v string) error {
	self.stk[self.sp].val = &visitorUserString{Value: v}
	if err := self.incrSP(); err != nil {
		return err
	}
	return self.onValueEnd()
}

func (self *visitorUserNodeVisitorDecoder) OnInt64(v int64, n json.Number) error {
	self.stk[self.sp].val = &visitorUserInt64{Value: v}
	if err := self.incrSP(); err != nil {
		return err
	}
	return self.onValueEnd()
}

func (self *visitorUserNodeVisitorDecoder) OnFloat64(v float64, n json.Number) error {
	self.stk[self.sp].val = &visitorUserFloat64{Value: v}
	if err := self.incrSP(); err != nil {
		return err
	}
	return self.onValueEnd()
}

func (self *visitorUserNodeVisitorDecoder) OnObjectBegin(capacity int) error {
	self.stk[self.sp].obj = make(map[string]visitorUserNode, capacity)
	return self.incrSP()
}

func (self *visitorUserNodeVisitorDecoder) OnObjectKey(key string) error {
	self.stk[self.sp].key = key
	return self.incrSP()
}

func (self *visitorUserNodeVisitorDecoder) OnObjectEnd() error {
	self.stk[self.sp-1].val = &visitorUserObject{Value: self.stk[self.sp-1].obj}
	self.stk[self.sp-1].obj = nil
	return self.onValueEnd()
}

func (self *visitorUserNodeVisitorDecoder) OnArrayBegin(capacity int) error {
	self.stk[self.sp].arr = make([]visitorUserNode, 0, capacity)
	return self.incrSP()
}

func (self *visitorUserNodeVisitorDecoder) OnArrayEnd() error {
	self.stk[self.sp-1].val = &visitorUserArray{Value: self.stk[self.sp-1].arr}
	self.stk[self.sp-1].arr = nil
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
	self.stk[self.sp-3].obj[self.stk[self.sp-2].key] = self.stk[self.sp-1].val
	self.sp -= 2
	return nil
}

func parseUserNodeRecursive(node visitorUserNode, desc *proto.Descriptor, p *binary.BinaryProtocol, self *BinaryConv, hierarchy int) error {
	var fieldDesc proto.FieldDescriptor
	var messageDesc proto.MessageDescriptor

	if hierarchy == 0 {
		messageDesc = (*desc).(proto.MessageDescriptor)
	} else {
		fieldDesc = (*desc).(proto.FieldDescriptor)
	}

	switch node.(type) {
	case *visitorUserNull:

	case *visitorUserInt64:
		// need to concern String2Int64
		nhs, ok := node.(*visitorUserInt64)
		if !ok {
			return meta.NewError(meta.ErrDismatchType, "json2pb:: unsatified visitorUserNodeType: visitorUserInt64", nil)
		}
		if err := p.WriteAnyWithDesc(&fieldDesc, nhs.UserNode(&fieldDesc), false, self.opts.DisallowUnknownField, true); err != nil {
			return meta.NewError(meta.ErrWrite, fmt.Sprintf("json2pb:: write int failed : %v", nhs.Value), err)
		}
	case *visitorUserFloat64:
		nhs, ok := node.(*visitorUserFloat64)
		if !ok {
			return meta.NewError(meta.ErrDismatchType, "json2pb:: unsatified visitorUserNodeType: visitorUserFloat64", nil)
		}
		if err := p.WriteAnyWithDesc(&fieldDesc, nhs.UserNode(&fieldDesc), false, self.opts.DisallowUnknownField, true); err != nil {
			return meta.NewError(meta.ErrWrite, fmt.Sprintf("json2pb:: write float failed : %v", nhs.Value), err)
		}
	case *visitorUserBool:
		nhs, ok := node.(*visitorUserBool)
		if !ok {
			return meta.NewError(meta.ErrDismatchType, "json2pb:: unsatified visitorUserNodeType: visitorUserBool", nil)
		}
		if err := p.WriteAnyWithDesc(&fieldDesc, nhs.UserNode(&fieldDesc), false, self.opts.DisallowUnknownField, true); err != nil {
			return meta.NewError(meta.ErrWrite, fmt.Sprintf("json2pb:: write bool failed : %v", nhs.Value), err)
		}
	case *visitorUserString:
		nhs, ok := node.(*visitorUserString)
		if !ok {
			return meta.NewError(meta.ErrConvert, "json2pb:: unsatified visitorUserNodeType: visitorUserString", nil)
		}
		if err := p.WriteAnyWithDesc(&fieldDesc, nhs.UserNode(&fieldDesc), false, self.opts.DisallowUnknownField, true); err != nil {
			return meta.NewError(meta.ErrWrite, fmt.Sprintf("json2pb:: write string failed : %v", nhs.Value), err)
		}
	case *visitorUserObject:
		nhs, ok := node.(*visitorUserObject)
		if !ok {
			return meta.NewError(meta.ErrDismatchType, "json2pb:: unsatified visitorUserNodeType: visitorUserObject", nil)
		}
		vs := nhs.Value
		// case Map
		if fieldDesc != nil && fieldDesc.IsMap() && len(vs) > 0 {
			MapKeyDesc := fieldDesc.MapKey()
			MapValueDesc := fieldDesc.MapValue()
			var pos int
			var value interface{}
			for k, v := range vs {
				// pay attention to Key type conversion
				switch MapKeyDesc.Kind() {
				case protoreflect.Kind(proto.Int32Kind), protoreflect.Kind(proto.Sint32Kind), protoreflect.Kind(proto.Sfixed32Kind), protoreflect.Kind(proto.Fixed32Kind):
					t, _ := strconv.ParseInt(k, 10, 32)
					value = int32(t)
				case protoreflect.Kind(proto.Uint32Kind):
					t, _ := strconv.ParseInt(k, 10, 32)
					value = uint32(t)
				case protoreflect.Kind(proto.Uint64Kind):
					t, _ := strconv.ParseInt(k, 10, 64)
					value = uint64(t)
				case protoreflect.Kind(proto.Int64Kind):
					t, _ := strconv.ParseInt(k, 10, 64)
					value = t
				case protoreflect.Kind(proto.BoolKind):
					t, _ := strconv.ParseBool(k)
					value = t
				default:
					value = k
				}
				p.AppendTag(fieldDesc.Number(), proto.BytesType)
				p.Buf, pos = binary.AppendSpeculativeLength(p.Buf)
				if err := p.WriteAnyWithDesc(&MapKeyDesc, value, false, self.opts.DisallowUnknownField, true); err != nil {
					return meta.NewError(meta.ErrWrite, fmt.Sprintf("json2pb:: write MapKey failed : %v", k), err)
				}
				convertDesc := MapValueDesc.(proto.Descriptor)
				if err := parseUserNodeRecursive(v, &convertDesc, p, self, hierarchy+1); err != nil {
					return meta.NewError(meta.ErrWrite, fmt.Sprintf("json2pb:: write MapValue failed : %v", v), err)
				}
				p.Buf = binary.FinishSpeculativeLength(p.Buf, pos)
			}
		} else {
			// case Message, contain inner Message and initial Message
			var pos int
			if hierarchy > 0 {
				messageDesc = fieldDesc.Message()
				p.AppendTag(fieldDesc.Number(), proto.BytesType)
				p.Buf, pos = binary.AppendSpeculativeLength(p.Buf)
			}
			fieldDescs := messageDesc.Fields()
			for k, v := range vs {
				childDesc := fieldDescs.ByJSONName(k)
				if self.opts.DisallowUnknownField && childDesc == nil {
					return meta.NewError(meta.ErrUnknownField, fmt.Sprintf("json2pb:: json data exists unknownField : %v", k), nil)
				} else if childDesc == nil {
					// jsonField not exists in protoFile, consider to skip
					continue
				}
				convertDesc := childDesc.(proto.Descriptor)
				if err := parseUserNodeRecursive(v, &convertDesc, p, self, hierarchy+1); err != nil {
					return meta.NewError(meta.ErrWrite, fmt.Sprintf("json2pb:: write Message failed : %v", k), err)
				}
			}
			if hierarchy > 0 {
				p.Buf = binary.FinishSpeculativeLength(p.Buf, pos)
			}
		}
	case *visitorUserArray:
		nhs, ok := node.(*visitorUserArray)
		if !ok {
			return meta.NewError(meta.ErrDismatchType, "json2pb:: unsatified visitorUserNodeType: visitorUserArray", nil)
		}
		vs := nhs.Value
		// packed List bytes format: [tag][length][(L)v][(L)v][(L)v][(L)v]....
		if fieldDesc.IsPacked() && len(vs) > 0 {
			p.AppendTag(fieldDesc.Number(), proto.BytesType)
			var pos int
			p.Buf, pos = binary.AppendSpeculativeLength(p.Buf)
			for _, v := range vs {
				if err := p.WriteBaseTypeWithDesc(&fieldDesc, v.UserNode(&fieldDesc), false, self.opts.DisallowUnknownField, true); err != nil {
					return meta.NewError(meta.ErrWrite, fmt.Sprintf("json2pb:: write PackedList failed : %v", fieldDesc.JSONName()), err)
				}
			}
			p.Buf = binary.FinishSpeculativeLength(p.Buf, pos)
		} else {
			// unpacked List bytes format: [tag][(L)V][tag][(L)V][tag][(L)V][tag][(L)V]...
			convertDesc := fieldDesc.(proto.Descriptor)
			for _, v := range vs {
				if err := parseUserNodeRecursive(v, &convertDesc, p, self, hierarchy+1); err != nil {
					return meta.NewError(meta.ErrWrite, fmt.Sprintf("json2pb:: write unPackedList failed : %v", fieldDesc.JSONName()), err)
				}
			}
		}
	default:
		fmt.Println("unknown field")
	}
	return nil
}
