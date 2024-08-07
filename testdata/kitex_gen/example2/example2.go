// Code generated by thriftgo (0.3.15). DO NOT EDIT.

package example2

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/base"
)

type FOO int64

const (
	FOO_A FOO = 1
)

func (p FOO) String() string {
	switch p {
	case FOO_A:
		return "A"
	}
	return "<UNSET>"
}

func FOOFromString(s string) (FOO, error) {
	switch s {
	case "A":
		return FOO_A, nil
	}
	return FOO(0), fmt.Errorf("not a valid FOO string")
}

func FOOPtr(v FOO) *FOO { return &v }
func (p *FOO) Scan(value interface{}) (err error) {
	var result sql.NullInt64
	err = result.Scan(value)
	*p = FOO(result.Int64)
	return
}

func (p *FOO) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return int64(*p), nil
}

type InnerBase struct {
	Bool                  bool                      `thrift:"Bool,1" json:"Bool"`
	Byte                  int8                      `thrift:"Byte,2" json:"Byte"`
	Int16                 int16                     `thrift:"Int16,3" json:"Int16"`
	Int32                 int32                     `thrift:"Int32,4" json:"Int32"`
	Int64                 int64                     `thrift:"Int64,5" json:"Int64"`
	Double                float64                   `thrift:"Double,6" json:"Double"`
	String_               string                    `thrift:"String,7" json:"String"`
	ListInt32             []int32                   `thrift:"ListInt32,8" json:"ListInt32"`
	MapStringString       map[string]string         `thrift:"MapStringString,9" json:"MapStringString"`
	SetInt32              []int32                   `thrift:"SetInt32,10" json:"SetInt32"`
	Foo                   FOO                       `thrift:"Foo,11" json:"Foo"`
	MapInt32String        map[int32]string          `thrift:"MapInt32String,12" json:"MapInt32String"`
	Binary                []byte                    `thrift:"Binary,13" json:"Binary"`
	MapInt8String         map[int8]string           `thrift:"MapInt8String,14" json:"MapInt8String"`
	MapInt16String        map[int16]string          `thrift:"MapInt16String,15" json:"MapInt16String"`
	MapInt64String        map[int64]string          `thrift:"MapInt64String,16" json:"MapInt64String"`
	MapDoubleString       map[float64]string        `thrift:"MapDoubleString,17" json:"MapDoubleString"`
	ListInnerBase         []*InnerBase              `thrift:"ListInnerBase,18" json:"ListInnerBase"`
	MapInnerBaseInnerBase map[*InnerBase]*InnerBase `thrift:"MapInnerBaseInnerBase,19" json:"MapInnerBaseInnerBase"`
	Base                  *base.Base                `thrift:"Base,255" json:"Base"`
}

func NewInnerBase() *InnerBase {
	return &InnerBase{}
}

func (p *InnerBase) InitDefault() {
}

func (p *InnerBase) GetBool() (v bool) {
	return p.Bool
}

func (p *InnerBase) GetByte() (v int8) {
	return p.Byte
}

func (p *InnerBase) GetInt16() (v int16) {
	return p.Int16
}

func (p *InnerBase) GetInt32() (v int32) {
	return p.Int32
}

func (p *InnerBase) GetInt64() (v int64) {
	return p.Int64
}

func (p *InnerBase) GetDouble() (v float64) {
	return p.Double
}

func (p *InnerBase) GetString() (v string) {
	return p.String_
}

func (p *InnerBase) GetListInt32() (v []int32) {
	return p.ListInt32
}

func (p *InnerBase) GetMapStringString() (v map[string]string) {
	return p.MapStringString
}

func (p *InnerBase) GetSetInt32() (v []int32) {
	return p.SetInt32
}

func (p *InnerBase) GetFoo() (v FOO) {
	return p.Foo
}

func (p *InnerBase) GetMapInt32String() (v map[int32]string) {
	return p.MapInt32String
}

func (p *InnerBase) GetBinary() (v []byte) {
	return p.Binary
}

func (p *InnerBase) GetMapInt8String() (v map[int8]string) {
	return p.MapInt8String
}

func (p *InnerBase) GetMapInt16String() (v map[int16]string) {
	return p.MapInt16String
}

func (p *InnerBase) GetMapInt64String() (v map[int64]string) {
	return p.MapInt64String
}

func (p *InnerBase) GetMapDoubleString() (v map[float64]string) {
	return p.MapDoubleString
}

func (p *InnerBase) GetListInnerBase() (v []*InnerBase) {
	return p.ListInnerBase
}

func (p *InnerBase) GetMapInnerBaseInnerBase() (v map[*InnerBase]*InnerBase) {
	return p.MapInnerBaseInnerBase
}

var InnerBase_Base_DEFAULT *base.Base

func (p *InnerBase) GetBase() (v *base.Base) {
	if !p.IsSetBase() {
		return InnerBase_Base_DEFAULT
	}
	return p.Base
}

func (p *InnerBase) IsSetBase() bool {
	return p.Base != nil
}

func (p *InnerBase) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("InnerBase(%+v)", *p)
}

var fieldIDToName_InnerBase = map[int16]string{
	1:   "Bool",
	2:   "Byte",
	3:   "Int16",
	4:   "Int32",
	5:   "Int64",
	6:   "Double",
	7:   "String",
	8:   "ListInt32",
	9:   "MapStringString",
	10:  "SetInt32",
	11:  "Foo",
	12:  "MapInt32String",
	13:  "Binary",
	14:  "MapInt8String",
	15:  "MapInt16String",
	16:  "MapInt64String",
	17:  "MapDoubleString",
	18:  "ListInnerBase",
	19:  "MapInnerBaseInnerBase",
	255: "Base",
}

type InnerBasePartial struct {
	Bool                  bool                                    `thrift:"Bool,1" json:"Bool"`
	ListInt32             []int32                                 `thrift:"ListInt32,8" json:"ListInt32"`
	MapStringString       map[string]string                       `thrift:"MapStringString,9" json:"MapStringString"`
	MapDoubleString       map[float64]string                      `thrift:"MapDoubleString,17" json:"MapDoubleString"`
	ListInnerBase         []*InnerBasePartial                     `thrift:"ListInnerBase,18" json:"ListInnerBase"`
	MapInnerBaseInnerBase map[*InnerBasePartial]*InnerBasePartial `thrift:"MapInnerBaseInnerBase,19" json:"MapInnerBaseInnerBase"`
	MapStringString2      map[string]string                       `thrift:"MapStringString2,127" json:"MapStringString2"`
}

func NewInnerBasePartial() *InnerBasePartial {
	return &InnerBasePartial{}
}

func (p *InnerBasePartial) InitDefault() {
}

func (p *InnerBasePartial) GetBool() (v bool) {
	return p.Bool
}

func (p *InnerBasePartial) GetListInt32() (v []int32) {
	return p.ListInt32
}

func (p *InnerBasePartial) GetMapStringString() (v map[string]string) {
	return p.MapStringString
}

func (p *InnerBasePartial) GetMapDoubleString() (v map[float64]string) {
	return p.MapDoubleString
}

func (p *InnerBasePartial) GetListInnerBase() (v []*InnerBasePartial) {
	return p.ListInnerBase
}

func (p *InnerBasePartial) GetMapInnerBaseInnerBase() (v map[*InnerBasePartial]*InnerBasePartial) {
	return p.MapInnerBaseInnerBase
}

func (p *InnerBasePartial) GetMapStringString2() (v map[string]string) {
	return p.MapStringString2
}

func (p *InnerBasePartial) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("InnerBasePartial(%+v)", *p)
}

var fieldIDToName_InnerBasePartial = map[int16]string{
	1:   "Bool",
	8:   "ListInt32",
	9:   "MapStringString",
	17:  "MapDoubleString",
	18:  "ListInnerBase",
	19:  "MapInnerBaseInnerBase",
	127: "MapStringString2",
}

type BasePartial struct {
	TrafficEnv *base.TrafficEnv `thrift:"TrafficEnv,5,optional" json:"TrafficEnv,omitempty"`
}

func NewBasePartial() *BasePartial {
	return &BasePartial{}
}

func (p *BasePartial) InitDefault() {
}

var BasePartial_TrafficEnv_DEFAULT *base.TrafficEnv

func (p *BasePartial) GetTrafficEnv() (v *base.TrafficEnv) {
	if !p.IsSetTrafficEnv() {
		return BasePartial_TrafficEnv_DEFAULT
	}
	return p.TrafficEnv
}

func (p *BasePartial) IsSetTrafficEnv() bool {
	return p.TrafficEnv != nil
}

func (p *BasePartial) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("BasePartial(%+v)", *p)
}

var fieldIDToName_BasePartial = map[int16]string{
	5: "TrafficEnv",
}

type ExampleReq struct {
	Msg       *string    `thrift:"Msg,1,optional" json:"Msg,omitempty"`
	A         *int32     `thrift:"A,2,optional" json:"A,omitempty"`
	InnerBase *InnerBase `thrift:"InnerBase,3" json:"InnerBase"`
	Base      *base.Base `thrift:"Base,255,required" json:"Base"`
	Subfix    float64    `thrift:"Subfix,32767" json:"Subfix"`
}

func NewExampleReq() *ExampleReq {
	return &ExampleReq{}
}

func (p *ExampleReq) InitDefault() {
}

var ExampleReq_Msg_DEFAULT string

func (p *ExampleReq) GetMsg() (v string) {
	if !p.IsSetMsg() {
		return ExampleReq_Msg_DEFAULT
	}
	return *p.Msg
}

var ExampleReq_A_DEFAULT int32

func (p *ExampleReq) GetA() (v int32) {
	if !p.IsSetA() {
		return ExampleReq_A_DEFAULT
	}
	return *p.A
}

var ExampleReq_InnerBase_DEFAULT *InnerBase

func (p *ExampleReq) GetInnerBase() (v *InnerBase) {
	if !p.IsSetInnerBase() {
		return ExampleReq_InnerBase_DEFAULT
	}
	return p.InnerBase
}

var ExampleReq_Base_DEFAULT *base.Base

func (p *ExampleReq) GetBase() (v *base.Base) {
	if !p.IsSetBase() {
		return ExampleReq_Base_DEFAULT
	}
	return p.Base
}

func (p *ExampleReq) GetSubfix() (v float64) {
	return p.Subfix
}

func (p *ExampleReq) IsSetMsg() bool {
	return p.Msg != nil
}

func (p *ExampleReq) IsSetA() bool {
	return p.A != nil
}

func (p *ExampleReq) IsSetInnerBase() bool {
	return p.InnerBase != nil
}

func (p *ExampleReq) IsSetBase() bool {
	return p.Base != nil
}

func (p *ExampleReq) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleReq(%+v)", *p)
}

var fieldIDToName_ExampleReq = map[int16]string{
	1:     "Msg",
	2:     "A",
	3:     "InnerBase",
	255:   "Base",
	32767: "Subfix",
}

type ExampleSuper struct {
	Msg       *string    `thrift:"Msg,1,optional" json:"Msg,omitempty"`
	A         *int32     `thrift:"A,2,optional" json:"A,omitempty"`
	InnerBase *InnerBase `thrift:"InnerBase,3" json:"InnerBase"`
	Ex1       string     `thrift:"Ex1,4" json:"Ex1"`
	Ex2       *string    `thrift:"Ex2,5,optional" json:"Ex2,omitempty"`
	Ex3       *string    `thrift:"Ex3,6,optional" json:"Ex3,omitempty"`
	Ex4       string     `thrift:"Ex4,7,required" json:"Ex4"`
	SelfRef   *SelfRef   `thrift:"SelfRef,9" json:"SelfRef"`
	Base      *base.Base `thrift:"Base,255,required" json:"Base"`
	Subfix    float64    `thrift:"Subfix,32767" json:"Subfix"`
}

func NewExampleSuper() *ExampleSuper {
	return &ExampleSuper{}
}

func (p *ExampleSuper) InitDefault() {
}

var ExampleSuper_Msg_DEFAULT string

func (p *ExampleSuper) GetMsg() (v string) {
	if !p.IsSetMsg() {
		return ExampleSuper_Msg_DEFAULT
	}
	return *p.Msg
}

var ExampleSuper_A_DEFAULT int32

func (p *ExampleSuper) GetA() (v int32) {
	if !p.IsSetA() {
		return ExampleSuper_A_DEFAULT
	}
	return *p.A
}

var ExampleSuper_InnerBase_DEFAULT *InnerBase

func (p *ExampleSuper) GetInnerBase() (v *InnerBase) {
	if !p.IsSetInnerBase() {
		return ExampleSuper_InnerBase_DEFAULT
	}
	return p.InnerBase
}

func (p *ExampleSuper) GetEx1() (v string) {
	return p.Ex1
}

var ExampleSuper_Ex2_DEFAULT string

func (p *ExampleSuper) GetEx2() (v string) {
	if !p.IsSetEx2() {
		return ExampleSuper_Ex2_DEFAULT
	}
	return *p.Ex2
}

var ExampleSuper_Ex3_DEFAULT string

func (p *ExampleSuper) GetEx3() (v string) {
	if !p.IsSetEx3() {
		return ExampleSuper_Ex3_DEFAULT
	}
	return *p.Ex3
}

func (p *ExampleSuper) GetEx4() (v string) {
	return p.Ex4
}

var ExampleSuper_SelfRef_DEFAULT *SelfRef

func (p *ExampleSuper) GetSelfRef() (v *SelfRef) {
	if !p.IsSetSelfRef() {
		return ExampleSuper_SelfRef_DEFAULT
	}
	return p.SelfRef
}

var ExampleSuper_Base_DEFAULT *base.Base

func (p *ExampleSuper) GetBase() (v *base.Base) {
	if !p.IsSetBase() {
		return ExampleSuper_Base_DEFAULT
	}
	return p.Base
}

func (p *ExampleSuper) GetSubfix() (v float64) {
	return p.Subfix
}

func (p *ExampleSuper) IsSetMsg() bool {
	return p.Msg != nil
}

func (p *ExampleSuper) IsSetA() bool {
	return p.A != nil
}

func (p *ExampleSuper) IsSetInnerBase() bool {
	return p.InnerBase != nil
}

func (p *ExampleSuper) IsSetEx2() bool {
	return p.Ex2 != nil
}

func (p *ExampleSuper) IsSetEx3() bool {
	return p.Ex3 != nil
}

func (p *ExampleSuper) IsSetSelfRef() bool {
	return p.SelfRef != nil
}

func (p *ExampleSuper) IsSetBase() bool {
	return p.Base != nil
}

func (p *ExampleSuper) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleSuper(%+v)", *p)
}

var fieldIDToName_ExampleSuper = map[int16]string{
	1:     "Msg",
	2:     "A",
	3:     "InnerBase",
	4:     "Ex1",
	5:     "Ex2",
	6:     "Ex3",
	7:     "Ex4",
	9:     "SelfRef",
	255:   "Base",
	32767: "Subfix",
}

type SelfRef struct {
	Self *SelfRef `thrift:"self,1,optional" json:"self,omitempty"`
}

func NewSelfRef() *SelfRef {
	return &SelfRef{}
}

func (p *SelfRef) InitDefault() {
}

var SelfRef_Self_DEFAULT *SelfRef

func (p *SelfRef) GetSelf() (v *SelfRef) {
	if !p.IsSetSelf() {
		return SelfRef_Self_DEFAULT
	}
	return p.Self
}

func (p *SelfRef) IsSetSelf() bool {
	return p.Self != nil
}

func (p *SelfRef) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("SelfRef(%+v)", *p)
}

var fieldIDToName_SelfRef = map[int16]string{
	1: "self",
}

type ExampleReqPartial struct {
	Msg       *string           `thrift:"Msg,1,optional" json:"Msg,omitempty"`
	InnerBase *InnerBasePartial `thrift:"InnerBase,3" json:"InnerBase"`
	Base      *BasePartial      `thrift:"Base,255" json:"Base"`
}

func NewExampleReqPartial() *ExampleReqPartial {
	return &ExampleReqPartial{}
}

func (p *ExampleReqPartial) InitDefault() {
}

var ExampleReqPartial_Msg_DEFAULT string

func (p *ExampleReqPartial) GetMsg() (v string) {
	if !p.IsSetMsg() {
		return ExampleReqPartial_Msg_DEFAULT
	}
	return *p.Msg
}

var ExampleReqPartial_InnerBase_DEFAULT *InnerBasePartial

func (p *ExampleReqPartial) GetInnerBase() (v *InnerBasePartial) {
	if !p.IsSetInnerBase() {
		return ExampleReqPartial_InnerBase_DEFAULT
	}
	return p.InnerBase
}

var ExampleReqPartial_Base_DEFAULT *BasePartial

func (p *ExampleReqPartial) GetBase() (v *BasePartial) {
	if !p.IsSetBase() {
		return ExampleReqPartial_Base_DEFAULT
	}
	return p.Base
}

func (p *ExampleReqPartial) IsSetMsg() bool {
	return p.Msg != nil
}

func (p *ExampleReqPartial) IsSetInnerBase() bool {
	return p.InnerBase != nil
}

func (p *ExampleReqPartial) IsSetBase() bool {
	return p.Base != nil
}

func (p *ExampleReqPartial) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleReqPartial(%+v)", *p)
}

var fieldIDToName_ExampleReqPartial = map[int16]string{
	1:   "Msg",
	3:   "InnerBase",
	255: "Base",
}

type ExampleResp struct {
	Msg           *string        `thrift:"Msg,1,optional" json:"Msg,omitempty"`
	RequiredField string         `thrift:"required_field,2,required" json:"required_field"`
	BaseResp      *base.BaseResp `thrift:"BaseResp,255" json:"BaseResp"`
}

func NewExampleResp() *ExampleResp {
	return &ExampleResp{}
}

func (p *ExampleResp) InitDefault() {
}

var ExampleResp_Msg_DEFAULT string

func (p *ExampleResp) GetMsg() (v string) {
	if !p.IsSetMsg() {
		return ExampleResp_Msg_DEFAULT
	}
	return *p.Msg
}

func (p *ExampleResp) GetRequiredField() (v string) {
	return p.RequiredField
}

var ExampleResp_BaseResp_DEFAULT *base.BaseResp

func (p *ExampleResp) GetBaseResp() (v *base.BaseResp) {
	if !p.IsSetBaseResp() {
		return ExampleResp_BaseResp_DEFAULT
	}
	return p.BaseResp
}

func (p *ExampleResp) IsSetMsg() bool {
	return p.Msg != nil
}

func (p *ExampleResp) IsSetBaseResp() bool {
	return p.BaseResp != nil
}

func (p *ExampleResp) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleResp(%+v)", *p)
}

var fieldIDToName_ExampleResp = map[int16]string{
	1:   "Msg",
	2:   "required_field",
	255: "BaseResp",
}

type A struct {
	Self *A `thrift:"self,1" json:"self"`
}

func NewA() *A {
	return &A{}
}

func (p *A) InitDefault() {
}

var A_Self_DEFAULT *A

func (p *A) GetSelf() (v *A) {
	if !p.IsSetSelf() {
		return A_Self_DEFAULT
	}
	return p.Self
}

func (p *A) IsSetSelf() bool {
	return p.Self != nil
}

func (p *A) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("A(%+v)", *p)
}

var fieldIDToName_A = map[int16]string{
	1: "self",
}

type Exception struct {
	Code int32  `thrift:"code,1" json:"code"`
	Msg  string `thrift:"msg,255" json:"msg"`
}

func NewException() *Exception {
	return &Exception{}
}

func (p *Exception) InitDefault() {
}

func (p *Exception) GetCode() (v int32) {
	return p.Code
}

func (p *Exception) GetMsg() (v string) {
	return p.Msg
}

func (p *Exception) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Exception(%+v)", *p)
}
func (p *Exception) Error() string {
	return p.String()
}

var fieldIDToName_Exception = map[int16]string{
	1:   "code",
	255: "msg",
}

type ExampleService interface {
	ExampleMethod(ctx context.Context, req *ExampleReq) (r *ExampleResp, err error)

	ExamplePartialMethod(ctx context.Context, req *ExampleReqPartial) (r *A, err error)

	ExampleSuperMethod(ctx context.Context, req *ExampleSuper) (r *A, err error)

	Foo(ctx context.Context, req *A) (r *A, err error)

	Ping(ctx context.Context, msg string) (r string, err error)

	Oneway(ctx context.Context, msg string) (err error)

	Void(ctx context.Context, msg string) (err error)
}

type ExampleServiceExampleMethodArgs struct {
	Req *ExampleReq `thrift:"req,1" json:"req"`
}

func NewExampleServiceExampleMethodArgs() *ExampleServiceExampleMethodArgs {
	return &ExampleServiceExampleMethodArgs{}
}

func (p *ExampleServiceExampleMethodArgs) InitDefault() {
}

var ExampleServiceExampleMethodArgs_Req_DEFAULT *ExampleReq

func (p *ExampleServiceExampleMethodArgs) GetReq() (v *ExampleReq) {
	if !p.IsSetReq() {
		return ExampleServiceExampleMethodArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *ExampleServiceExampleMethodArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *ExampleServiceExampleMethodArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceExampleMethodArgs(%+v)", *p)
}

var fieldIDToName_ExampleServiceExampleMethodArgs = map[int16]string{
	1: "req",
}

type ExampleServiceExampleMethodResult struct {
	Success *ExampleResp `thrift:"success,0,optional" json:"success,omitempty"`
	Err     *Exception   `thrift:"err,1,optional" json:"err,omitempty"`
}

func NewExampleServiceExampleMethodResult() *ExampleServiceExampleMethodResult {
	return &ExampleServiceExampleMethodResult{}
}

func (p *ExampleServiceExampleMethodResult) InitDefault() {
}

var ExampleServiceExampleMethodResult_Success_DEFAULT *ExampleResp

func (p *ExampleServiceExampleMethodResult) GetSuccess() (v *ExampleResp) {
	if !p.IsSetSuccess() {
		return ExampleServiceExampleMethodResult_Success_DEFAULT
	}
	return p.Success
}

var ExampleServiceExampleMethodResult_Err_DEFAULT *Exception

func (p *ExampleServiceExampleMethodResult) GetErr() (v *Exception) {
	if !p.IsSetErr() {
		return ExampleServiceExampleMethodResult_Err_DEFAULT
	}
	return p.Err
}

func (p *ExampleServiceExampleMethodResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *ExampleServiceExampleMethodResult) IsSetErr() bool {
	return p.Err != nil
}

func (p *ExampleServiceExampleMethodResult) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceExampleMethodResult(%+v)", *p)
}

var fieldIDToName_ExampleServiceExampleMethodResult = map[int16]string{
	0: "success",
	1: "err",
}

type ExampleServiceExamplePartialMethodArgs struct {
	Req *ExampleReqPartial `thrift:"req,1" json:"req"`
}

func NewExampleServiceExamplePartialMethodArgs() *ExampleServiceExamplePartialMethodArgs {
	return &ExampleServiceExamplePartialMethodArgs{}
}

func (p *ExampleServiceExamplePartialMethodArgs) InitDefault() {
}

var ExampleServiceExamplePartialMethodArgs_Req_DEFAULT *ExampleReqPartial

func (p *ExampleServiceExamplePartialMethodArgs) GetReq() (v *ExampleReqPartial) {
	if !p.IsSetReq() {
		return ExampleServiceExamplePartialMethodArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *ExampleServiceExamplePartialMethodArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *ExampleServiceExamplePartialMethodArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceExamplePartialMethodArgs(%+v)", *p)
}

var fieldIDToName_ExampleServiceExamplePartialMethodArgs = map[int16]string{
	1: "req",
}

type ExampleServiceExamplePartialMethodResult struct {
	Success *A `thrift:"success,0,optional" json:"success,omitempty"`
}

func NewExampleServiceExamplePartialMethodResult() *ExampleServiceExamplePartialMethodResult {
	return &ExampleServiceExamplePartialMethodResult{}
}

func (p *ExampleServiceExamplePartialMethodResult) InitDefault() {
}

var ExampleServiceExamplePartialMethodResult_Success_DEFAULT *A

func (p *ExampleServiceExamplePartialMethodResult) GetSuccess() (v *A) {
	if !p.IsSetSuccess() {
		return ExampleServiceExamplePartialMethodResult_Success_DEFAULT
	}
	return p.Success
}

func (p *ExampleServiceExamplePartialMethodResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *ExampleServiceExamplePartialMethodResult) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceExamplePartialMethodResult(%+v)", *p)
}

var fieldIDToName_ExampleServiceExamplePartialMethodResult = map[int16]string{
	0: "success",
}

type ExampleServiceExampleSuperMethodArgs struct {
	Req *ExampleSuper `thrift:"req,1" json:"req"`
}

func NewExampleServiceExampleSuperMethodArgs() *ExampleServiceExampleSuperMethodArgs {
	return &ExampleServiceExampleSuperMethodArgs{}
}

func (p *ExampleServiceExampleSuperMethodArgs) InitDefault() {
}

var ExampleServiceExampleSuperMethodArgs_Req_DEFAULT *ExampleSuper

func (p *ExampleServiceExampleSuperMethodArgs) GetReq() (v *ExampleSuper) {
	if !p.IsSetReq() {
		return ExampleServiceExampleSuperMethodArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *ExampleServiceExampleSuperMethodArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *ExampleServiceExampleSuperMethodArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceExampleSuperMethodArgs(%+v)", *p)
}

var fieldIDToName_ExampleServiceExampleSuperMethodArgs = map[int16]string{
	1: "req",
}

type ExampleServiceExampleSuperMethodResult struct {
	Success *A `thrift:"success,0,optional" json:"success,omitempty"`
}

func NewExampleServiceExampleSuperMethodResult() *ExampleServiceExampleSuperMethodResult {
	return &ExampleServiceExampleSuperMethodResult{}
}

func (p *ExampleServiceExampleSuperMethodResult) InitDefault() {
}

var ExampleServiceExampleSuperMethodResult_Success_DEFAULT *A

func (p *ExampleServiceExampleSuperMethodResult) GetSuccess() (v *A) {
	if !p.IsSetSuccess() {
		return ExampleServiceExampleSuperMethodResult_Success_DEFAULT
	}
	return p.Success
}

func (p *ExampleServiceExampleSuperMethodResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *ExampleServiceExampleSuperMethodResult) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceExampleSuperMethodResult(%+v)", *p)
}

var fieldIDToName_ExampleServiceExampleSuperMethodResult = map[int16]string{
	0: "success",
}

type ExampleServiceFooArgs struct {
	Req *A `thrift:"req,1" json:"req"`
}

func NewExampleServiceFooArgs() *ExampleServiceFooArgs {
	return &ExampleServiceFooArgs{}
}

func (p *ExampleServiceFooArgs) InitDefault() {
}

var ExampleServiceFooArgs_Req_DEFAULT *A

func (p *ExampleServiceFooArgs) GetReq() (v *A) {
	if !p.IsSetReq() {
		return ExampleServiceFooArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *ExampleServiceFooArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *ExampleServiceFooArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceFooArgs(%+v)", *p)
}

var fieldIDToName_ExampleServiceFooArgs = map[int16]string{
	1: "req",
}

type ExampleServiceFooResult struct {
	Success *A `thrift:"success,0,optional" json:"success,omitempty"`
}

func NewExampleServiceFooResult() *ExampleServiceFooResult {
	return &ExampleServiceFooResult{}
}

func (p *ExampleServiceFooResult) InitDefault() {
}

var ExampleServiceFooResult_Success_DEFAULT *A

func (p *ExampleServiceFooResult) GetSuccess() (v *A) {
	if !p.IsSetSuccess() {
		return ExampleServiceFooResult_Success_DEFAULT
	}
	return p.Success
}

func (p *ExampleServiceFooResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *ExampleServiceFooResult) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceFooResult(%+v)", *p)
}

var fieldIDToName_ExampleServiceFooResult = map[int16]string{
	0: "success",
}

type ExampleServicePingArgs struct {
	Msg string `thrift:"msg,1" json:"msg"`
}

func NewExampleServicePingArgs() *ExampleServicePingArgs {
	return &ExampleServicePingArgs{}
}

func (p *ExampleServicePingArgs) InitDefault() {
}

func (p *ExampleServicePingArgs) GetMsg() (v string) {
	return p.Msg
}

func (p *ExampleServicePingArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServicePingArgs(%+v)", *p)
}

var fieldIDToName_ExampleServicePingArgs = map[int16]string{
	1: "msg",
}

type ExampleServicePingResult struct {
	Success *string `thrift:"success,0,optional" json:"success,omitempty"`
}

func NewExampleServicePingResult() *ExampleServicePingResult {
	return &ExampleServicePingResult{}
}

func (p *ExampleServicePingResult) InitDefault() {
}

var ExampleServicePingResult_Success_DEFAULT string

func (p *ExampleServicePingResult) GetSuccess() (v string) {
	if !p.IsSetSuccess() {
		return ExampleServicePingResult_Success_DEFAULT
	}
	return *p.Success
}

func (p *ExampleServicePingResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *ExampleServicePingResult) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServicePingResult(%+v)", *p)
}

var fieldIDToName_ExampleServicePingResult = map[int16]string{
	0: "success",
}

type ExampleServiceOnewayArgs struct {
	Msg string `thrift:"msg,1" json:"msg"`
}

func NewExampleServiceOnewayArgs() *ExampleServiceOnewayArgs {
	return &ExampleServiceOnewayArgs{}
}

func (p *ExampleServiceOnewayArgs) InitDefault() {
}

func (p *ExampleServiceOnewayArgs) GetMsg() (v string) {
	return p.Msg
}

func (p *ExampleServiceOnewayArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceOnewayArgs(%+v)", *p)
}

var fieldIDToName_ExampleServiceOnewayArgs = map[int16]string{
	1: "msg",
}

type ExampleServiceVoidArgs struct {
	Msg string `thrift:"msg,1" json:"msg"`
}

func NewExampleServiceVoidArgs() *ExampleServiceVoidArgs {
	return &ExampleServiceVoidArgs{}
}

func (p *ExampleServiceVoidArgs) InitDefault() {
}

func (p *ExampleServiceVoidArgs) GetMsg() (v string) {
	return p.Msg
}

func (p *ExampleServiceVoidArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceVoidArgs(%+v)", *p)
}

var fieldIDToName_ExampleServiceVoidArgs = map[int16]string{
	1: "msg",
}

type ExampleServiceVoidResult struct {
}

func NewExampleServiceVoidResult() *ExampleServiceVoidResult {
	return &ExampleServiceVoidResult{}
}

func (p *ExampleServiceVoidResult) InitDefault() {
}

func (p *ExampleServiceVoidResult) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceVoidResult(%+v)", *p)
}

var fieldIDToName_ExampleServiceVoidResult = map[int16]string{}

// exceptions of methods in ExampleService.
var (
	_ error = (*Exception)(nil)
)
