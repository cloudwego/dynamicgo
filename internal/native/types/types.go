/*
 * Copyright 2023 CloudWeGo Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

import (
	"fmt"
	"sync"
	"unsafe"
)

const (
	MaxInt64StringLen   = 21
	MaxFloat64StringLen = 32
)

type ValueType int64

func (v ValueType) String() string {
	return _ValueTypes[v]
}

var _ValueTypes = []string{
	0:         "none",
	1:         "error",
	V_NULL:    "null",
	V_TRUE:    "bool",
	V_FALSE:   "bool",
	V_ARRAY:   "array",
	V_OBJECT:  "object",
	V_STRING:  "string",
	V_DOUBLE:  "double",
	V_INTEGER: "integer",
	33:        "number",
	34:        "any",
}

type ParsingError uint

type SearchingError uint

const (
	V_EOF     ValueType = 1
	V_NULL    ValueType = 2
	V_TRUE    ValueType = 3
	V_FALSE   ValueType = 4
	V_ARRAY   ValueType = 5
	V_OBJECT  ValueType = 6
	V_STRING  ValueType = 7
	V_DOUBLE  ValueType = 8
	V_INTEGER ValueType = 9
	_         ValueType = 10 // V_KEY_SEP
	_         ValueType = 11 // V_ELEM_SEP
	_         ValueType = 12 // V_ARRAY_END
	_         ValueType = 13 // V_OBJECT_END
	V_MAX
)

const (
	B_DOUBLE_UNQUOTE  = 0
	B_UNICODE_REPLACE = 1
)

const (
	// for unquote
	F_DOUBLE_UNQUOTE  = 1 << B_DOUBLE_UNQUOTE
	F_UNICODE_REPLACE = 1 << B_UNICODE_REPLACE
	// for j2t_fsm_exec
	F_ALLOW_UNKNOWN  = 1
	F_WRITE_DEFAULT  = 1 << 1
	F_VALUE_MAPPING  = 1 << 2
	F_HTTP_MAPPING   = 1 << 3
	F_STRING_INT     = 1 << 4
	F_WRITE_REQUIRE  = 1 << 5
	F_NO_BASE64      = 1 << 6
	F_WRITE_OPTIONAL = 1 << 7
	F_TRACE_BACK     = 1 << 8
)

const (
	MAX_RECURSE = 4096
)

const (
	SPACE_MASK = (1 << ' ') | (1 << '\t') | (1 << '\r') | (1 << '\n')
)

const (
	ERR_WRAP_SHIFT_CODE = 8
	ERR_WRAP_SHIFT_POS  = 32
)

const (
	ERR_EOF                   ParsingError = 1
	ERR_INVALID_CHAR          ParsingError = 2
	ERR_INVALID_ESCAPE        ParsingError = 3
	ERR_INVALID_UNICODE       ParsingError = 4
	ERR_INTEGER_OVERFLOW      ParsingError = 5
	ERR_INVALID_NUMBER_FMT    ParsingError = 6
	ERR_RECURSE_EXCEED_MAX    ParsingError = 7
	ERR_FLOAT_INFINITY        ParsingError = 8
	ERR_DISMATCH_TYPE         ParsingError = 9
	ERR_NULL_REQUIRED         ParsingError = 10
	ERR_UNSUPPORT_THRIFT_TYPE ParsingError = 11
	ERR_UNKNOWN_FIELD         ParsingError = 12
	ERR_DISMATCH_TYPE2        ParsingError = 13
	ERR_DECODE_BASE64         ParsingError = 14
	ERR_OOM_BM                ParsingError = 16
	ERR_OOM_BUF               ParsingError = 17
	ERR_OOM_KEY               ParsingError = 18
	ERR_OOM_FIELD             ParsingError = 22
	ERR_OOM_FVAL              ParsingError = 23
	ERR_HTTP_MAPPING          ParsingError = 19
	ERR_HTTP_MAPPING_END      ParsingError = 21
	ERR_UNSUPPORT_VM_TYPE     ParsingError = 20
	ERR_VALUE_MAPPING_END     ParsingError = 24
)

var _ParsingErrors = []string{
	0:                         "ok",
	ERR_EOF:                   "eof",
	ERR_INVALID_CHAR:          "invalid char",
	ERR_INVALID_ESCAPE:        "invalid escape char",
	ERR_INVALID_UNICODE:       "invalid unicode escape",
	ERR_INTEGER_OVERFLOW:      "integer overflow",
	ERR_INVALID_NUMBER_FMT:    "invalid number format",
	ERR_RECURSE_EXCEED_MAX:    "recursion exceeded max depth",
	ERR_FLOAT_INFINITY:        "float number is infinity",
	ERR_DISMATCH_TYPE:         "dismatched type",
	ERR_NULL_REQUIRED:         "required field is not set",
	ERR_UNSUPPORT_THRIFT_TYPE: "unsupported type",
	ERR_UNKNOWN_FIELD:         "unknown field",
	ERR_DISMATCH_TYPE2:        "dismatched types",
	ERR_DECODE_BASE64:         "decode base64 error",
	ERR_OOM_BM:                "requireness cache is not enough",
	ERR_OOM_BUF:               "output buffer is not enough",
	ERR_OOM_KEY:               "key cache is not enough",
	ERR_HTTP_MAPPING:          "http mapping error",
	ERR_UNSUPPORT_VM_TYPE:     "unsupported value mapping type",
}

func (self ParsingError) Error() string {
	return "json: error when parsing input: " + self.Message()
}

func (self ParsingError) Message() string {
	if int(self) < len(_ParsingErrors) {
		return _ParsingErrors[self]
	} else {
		return fmt.Sprintf("unknown error %d", self)
	}
}

type JsonState struct {
	Vt   ValueType
	Dv   float64
	Iv   int64
	Ep   int64
	Dbuf *byte
	Dcap int
}

func (s JsonState) String() string {
	return fmt.Sprintf(`{Vt: %d, Dv: %f, Iv: %d, Ep: %d}`, s.Vt, s.Dv, s.Iv, s.Ep)
}

type StateMachine struct {
	Sp int
	Vt [MAX_RECURSE]int64
}

var stackPool = sync.Pool{
	New: func() interface{} {
		return &StateMachine{}
	},
}

func NewStateMachine() *StateMachine {
	return stackPool.Get().(*StateMachine)
}

func FreeStateMachine(fsm *StateMachine) {
	stackPool.Put(fsm)
}

const _SIZE_J2TEXTRA = 3

type J2TExtra [_SIZE_J2TEXTRA]uint64

type J2TState struct {
	State    int
	JsonPos  int
	TypeDesc uintptr
	Extra    J2TExtra
}

//go:nocheckptr
func (s J2TState) String() string {
	name := ""
	typ := 0
	if s.TypeDesc != 0 {
		desc := (*tTypeDesc)(unsafe.Pointer(s.TypeDesc))
		name = desc.name
		typ = int(desc.ttype)
	}
	return fmt.Sprintf("{State: %x, JsonPos: %d, TypeDesc: %s(%d), Extra:%v}", s.State, s.JsonPos, name, typ, s.Extra)
}

type tTypeDesc struct
{
    ttype uint8;
    name string;
    key *tTypeDesc;
    elem *tTypeDesc;
    st unsafe.Pointer;
};

func (self *J2TState) TdPointer() uintptr {
	return uintptr(self.TypeDesc)
}

type J2T_STATE uint8

const (
	J2T_VAL J2T_STATE = iota
	J2T_ARR
	J2T_OBJ
	J2T_KEY
	J2T_ELEM
	J2T_ARR_0
	J2T_OBJ_0

	J2T_VM J2T_STATE = 16
)

func (v J2T_STATE) String() string {
	return _J2T_STATEs[v]
}

var _J2T_STATEs = []string{
	J2T_VAL:   "J2T_VAL",
	J2T_ARR:   "J2T_ARR",
	J2T_OBJ:   "J2T_OBJ",
	J2T_KEY:   "J2T_KEY",
	J2T_ELEM:  "J2T_ELEM",
	J2T_ARR_0: "J2T_ARR_0",
	J2T_OBJ_0: "J2T_OBJ_0",
	J2T_VM:    "J2T_VM",
}

const (
	J2T_KEY_CACHE_SIZE   = 1024
	J2T_FIELD_CACHE_SIZE = 4096
	J2T_REQS_CACHE_SIZE  = 4096
	J2T_DBUF_SIZE        = 800
)

type J2TStateMachine struct {
	SP              int
	JT              JsonState
	ReqsCache       []byte
	KeyCache        []byte
	VT              [MAX_RECURSE]J2TState
	SM              StateMachine
	FieldCache      []int32
	FieldValueCache FieldValue
}

func (fsm *J2TStateMachine) String() string {
	var vt1 *J2TState
	if fsm.SP > 0 {
		vt1 = &fsm.VT[fsm.SP-1]
	}
	var vt2 *J2TState
	if fsm.SP > 1 {
		vt2 = &fsm.VT[fsm.SP-2]
	}
	var svt int64
	if fsm.SM.Sp > 0 {
		svt = fsm.SM.Vt[fsm.SM.Sp-1]
	}
	return fmt.Sprintf(`SP: %d
JsonState: %v
VT[SP-1]: %v
VT[SP-2]: %v
SM.SP: %d
SM.VT[SP-1]: %x
ReqsCache: %v
KeyCache: %s
FieldCache: %v
FieldValue: %#v`, fsm.SP, fsm.JT, vt1, vt2, fsm.SM.Sp, svt, fsm.ReqsCache, fsm.KeyCache, fsm.FieldCache, fsm.FieldValueCache)
}

type FieldValue struct {
	FieldID  int32
	ValBegin uint32
	ValEnd   uint32
}

var j2tStackPool = sync.Pool{
	New: func() interface{} {
		ret := &J2TStateMachine{}
		ret.ReqsCache = make([]byte, 0, J2T_REQS_CACHE_SIZE)
		ret.KeyCache = make([]byte, 0, J2T_KEY_CACHE_SIZE)
		ret.FieldCache = make([]int32, 0, J2T_FIELD_CACHE_SIZE)
		// ret.FieldValueCache = make([]FieldValue, 0, J2T_FIELD_CACHE_SIZE)
		tmp := make([]byte, 0, J2T_DBUF_SIZE)
		ret.JT.Dbuf = *(**byte)(unsafe.Pointer(&tmp))
		ret.JT.Dcap = J2T_DBUF_SIZE
		return ret
	},
}

func NewJ2TStateMachine() *J2TStateMachine {
	return j2tStackPool.Get().(*J2TStateMachine)
}

func FreeJ2TStateMachine(ret *J2TStateMachine) {
	ret.SP = 0
	ret.ReqsCache = ret.ReqsCache[:0]
	ret.KeyCache = ret.KeyCache[:0]
	ret.FieldCache = ret.FieldCache[:0]
	// ret.FieldValueCache = ret.FieldValueCache[:0]
	j2tStackPool.Put(ret)
}

func (ret *J2TStateMachine) Init(start int, desc unsafe.Pointer) {
	ret.SP = 1
	ret.VT[0] = J2TState{
		State:    0,
		JsonPos:  start,
		TypeDesc: uintptr(desc),
	}
}

func (self *J2TStateMachine) At(i int) *J2TState {
	if i < 0 || i >= self.SP {
		return nil
	}
	return &self.VT[i]
}

func (self *J2TStateMachine) Now() *J2TState {
	if self.SP == 0 || self.SP >= MAX_RECURSE {
		return nil
	}
	return &self.VT[self.SP-1]
}

func (self *J2TStateMachine) Drop() {
	if self.SP == 0 {
		return
	}
	self.SP--
}

func (self *J2TStateMachine) Next() *J2TState {
	if self.SP > MAX_RECURSE {
		return nil
	}
	return &self.VT[self.SP]
}

const (
	resizeFactor  = 2
	int64ByteSize = unsafe.Sizeof(int64(0))
)

func (ret *J2TStateMachine) GrowReqCache(n int) {
	c := cap(ret.ReqsCache) + n*resizeFactor
	tmp := make([]byte, len(ret.ReqsCache), c)
	copy(tmp, ret.ReqsCache)
	ret.ReqsCache = tmp
}

func (ret *J2TStateMachine) GrowKeyCache(n int) {
	c := cap(ret.KeyCache) + n*resizeFactor
	tmp := make([]byte, len(ret.KeyCache), c)
	copy(tmp, ret.KeyCache)
	ret.KeyCache = tmp
}

func (ret *J2TStateMachine) GrowFieldCache(n int) {
	c := cap(ret.FieldCache) + n
	tmp := make([]int32, len(ret.FieldCache), c)
	copy(tmp, ret.FieldCache)
	ret.FieldCache = tmp
}

// func (ret *J2TStateMachine) GrowFieldValueCache(n int) {
// 	c := cap(ret.FieldValueCache) + n
// 	tmp := make([]FieldValue, len(ret.FieldValueCache), c)
// 	copy(tmp, ret.FieldValueCache)
// 	ret.FieldValueCache = tmp
// }

func (ret *J2TStateMachine) SetPos(pos int) {
	vt := ret.Now()
	vt.JsonPos = pos
}

type TState struct {
	t uint8
	k uint8
	v uint8
	n int32
}

const (
	TB_SKIP_STACK_SIZE = 1024
)

type TStateMachine [TB_SKIP_STACK_SIZE]TState

var tsmPool = sync.Pool{
	New: func() interface{} {
		return &TStateMachine{}
	},
}

func NewTStateMachine() *TStateMachine {
	ret := tsmPool.Get().(*TStateMachine)
	return ret
}

func FreeTStateMachine(ret *TStateMachine) {
	tsmPool.Put(ret)
}

const (
	VM_NONE   uint16 = 0
	VM_JSCONV uint16 = 101
)
