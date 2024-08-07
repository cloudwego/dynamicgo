// Code generated by thriftgo (0.3.15). DO NOT EDIT.

package example

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/base"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/deep"
)

type FOO int64

const (
	FOO_B FOO = 0
	FOO_A FOO = 1
)

func (p FOO) String() string {
	switch p {
	case FOO_B:
		return "B"
	case FOO_A:
		return "A"
	}
	return "<UNSET>"
}

func FOOFromString(s string) (FOO, error) {
	switch s {
	case "B":
		return FOO_B, nil
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

type ExampleReq struct {
	Msg       *string    `thrift:"Msg,1,optional" json:"Msg,omitempty"`
	InnerBase int32      `thrift:"InnerBase,3" json:"InnerBase"`
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

func (p *ExampleReq) GetInnerBase() (v int32) {
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
	3:     "InnerBase",
	255:   "Base",
	32767: "Subfix",
}

type ExampleToSnakeCase struct {
	Msg       *string    `thrift:"Msg,1,optional" json:"Msg,omitempty"`
	ReqList   []string   `thrift:"req_list,2" json:"req_list"`
	InnerBase int32      `thrift:"InnerBase,3" json:"InnerBase"`
	Base      *base.Base `thrift:"Base,255,required" json:"Base"`
}

func NewExampleToSnakeCase() *ExampleToSnakeCase {
	return &ExampleToSnakeCase{}
}

func (p *ExampleToSnakeCase) InitDefault() {
}

var ExampleToSnakeCase_Msg_DEFAULT string

func (p *ExampleToSnakeCase) GetMsg() (v string) {
	if !p.IsSetMsg() {
		return ExampleToSnakeCase_Msg_DEFAULT
	}
	return *p.Msg
}

func (p *ExampleToSnakeCase) GetReqList() (v []string) {
	return p.ReqList
}

func (p *ExampleToSnakeCase) GetInnerBase() (v int32) {
	return p.InnerBase
}

var ExampleToSnakeCase_Base_DEFAULT *base.Base

func (p *ExampleToSnakeCase) GetBase() (v *base.Base) {
	if !p.IsSetBase() {
		return ExampleToSnakeCase_Base_DEFAULT
	}
	return p.Base
}

func (p *ExampleToSnakeCase) IsSetMsg() bool {
	return p.Msg != nil
}

func (p *ExampleToSnakeCase) IsSetBase() bool {
	return p.Base != nil
}

func (p *ExampleToSnakeCase) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleToSnakeCase(%+v)", *p)
}

var fieldIDToName_ExampleToSnakeCase = map[int16]string{
	1:   "Msg",
	2:   "req_list",
	3:   "InnerBase",
	255: "Base",
}

type ExampleResp struct {
	Msg           *string        `thrift:"Msg,1,optional" json:"Msg,omitempty"`
	RequiredField string         `thrift:"required_field,2,required" json:"required_field"`
	BaseResp      *base.BaseResp `thrift:"BaseResp,32767" json:"BaseResp"`
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
	1:     "Msg",
	2:     "required_field",
	32767: "BaseResp",
}

type A struct {
	Self string `thrift:"self,1" json:"self"`
	Foo  FOO    `thrift:"foo,2" json:"foo"`
}

func NewA() *A {
	return &A{}
}

func (p *A) InitDefault() {
}

func (p *A) GetSelf() (v string) {
	return p.Self
}

func (p *A) GetFoo() (v FOO) {
	return p.Foo
}

func (p *A) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("A(%+v)", *p)
}

var fieldIDToName_A = map[int16]string{
	1: "self",
	2: "foo",
}

type ExampleDefaultValue struct {
	String_     string            `thrift:"String,1" json:"String"`
	Int         int32             `thrift:"Int,2" json:"Int"`
	Double      float64           `thrift:"Double,3" json:"Double"`
	Bool        bool              `thrift:"Bool,4" json:"Bool"`
	List        []string          `thrift:"List,5" json:"List"`
	Map         map[string]string `thrift:"Map,6" json:"Map"`
	Set         []string          `thrift:"Set,7" json:"Set"`
	ConstString string            `thrift:"ConstString,8" json:"ConstString"`
	Enum        FOO               `thrift:"Enum,9" json:"Enum"`
}

func NewExampleDefaultValue() *ExampleDefaultValue {
	return &ExampleDefaultValue{

		String_: "default",
		Int:     1,
		Double:  1.1,
		Bool:    true,
		List: []string{
			"a",
			"b",
		},
		Map: map[string]string{
			"a": "b",
		},
		Set: []string{
			"a",
			"b",
		},
		ConstString: deep.ConstString,
		Enum:        FOO_A,
	}
}

func (p *ExampleDefaultValue) InitDefault() {
	p.String_ = "default"
	p.Int = 1
	p.Double = 1.1
	p.Bool = true
	p.List = []string{
		"a",
		"b",
	}
	p.Map = map[string]string{
		"a": "b",
	}
	p.Set = []string{
		"a",
		"b",
	}
	p.ConstString = deep.ConstString
	p.Enum = FOO_A
}

func (p *ExampleDefaultValue) GetString() (v string) {
	return p.String_
}

func (p *ExampleDefaultValue) GetInt() (v int32) {
	return p.Int
}

func (p *ExampleDefaultValue) GetDouble() (v float64) {
	return p.Double
}

func (p *ExampleDefaultValue) GetBool() (v bool) {
	return p.Bool
}

func (p *ExampleDefaultValue) GetList() (v []string) {
	return p.List
}

func (p *ExampleDefaultValue) GetMap() (v map[string]string) {
	return p.Map
}

func (p *ExampleDefaultValue) GetSet() (v []string) {
	return p.Set
}

func (p *ExampleDefaultValue) GetConstString() (v string) {
	return p.ConstString
}

func (p *ExampleDefaultValue) GetEnum() (v FOO) {
	return p.Enum
}

func (p *ExampleDefaultValue) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleDefaultValue(%+v)", *p)
}

var fieldIDToName_ExampleDefaultValue = map[int16]string{
	1: "String",
	2: "Int",
	3: "Double",
	4: "Bool",
	5: "List",
	6: "Map",
	7: "Set",
	8: "ConstString",
	9: "Enum",
}

type DeepRef struct {
	DeepRef *deep.TestStruct `thrift:"DeepRef,1" json:"DeepRef"`
}

func NewDeepRef() *DeepRef {
	return &DeepRef{}
}

func (p *DeepRef) InitDefault() {
}

var DeepRef_DeepRef_DEFAULT *deep.TestStruct

func (p *DeepRef) GetDeepRef() (v *deep.TestStruct) {
	if !p.IsSetDeepRef() {
		return DeepRef_DeepRef_DEFAULT
	}
	return p.DeepRef
}

func (p *DeepRef) IsSetDeepRef() bool {
	return p.DeepRef != nil
}

func (p *DeepRef) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("DeepRef(%+v)", *p)
}

var fieldIDToName_DeepRef = map[int16]string{
	1: "DeepRef",
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
	ExampleDeepRef(ctx context.Context, req *DeepRef) (r *DeepRef, err error)

	ExampleMethod(ctx context.Context, req *ExampleReq) (r *ExampleResp, err error)

	Foo(ctx context.Context, req *A) (r *A, err error)

	Ping(ctx context.Context, msg string) (r string, err error)

	Oneway(ctx context.Context, msg string) (err error)

	Void(ctx context.Context, msg string) (err error)

	ExampleToSnakeCase(ctx context.Context, req *ExampleToSnakeCase) (err error)

	ExampleDefaultValue(ctx context.Context, req *ExampleDefaultValue) (r *ExampleDefaultValue, err error)
}

type ExampleServiceExampleDeepRefArgs struct {
	Req *DeepRef `thrift:"req,1" json:"req"`
}

func NewExampleServiceExampleDeepRefArgs() *ExampleServiceExampleDeepRefArgs {
	return &ExampleServiceExampleDeepRefArgs{}
}

func (p *ExampleServiceExampleDeepRefArgs) InitDefault() {
}

var ExampleServiceExampleDeepRefArgs_Req_DEFAULT *DeepRef

func (p *ExampleServiceExampleDeepRefArgs) GetReq() (v *DeepRef) {
	if !p.IsSetReq() {
		return ExampleServiceExampleDeepRefArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *ExampleServiceExampleDeepRefArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *ExampleServiceExampleDeepRefArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceExampleDeepRefArgs(%+v)", *p)
}

var fieldIDToName_ExampleServiceExampleDeepRefArgs = map[int16]string{
	1: "req",
}

type ExampleServiceExampleDeepRefResult struct {
	Success *DeepRef `thrift:"success,0,optional" json:"success,omitempty"`
}

func NewExampleServiceExampleDeepRefResult() *ExampleServiceExampleDeepRefResult {
	return &ExampleServiceExampleDeepRefResult{}
}

func (p *ExampleServiceExampleDeepRefResult) InitDefault() {
}

var ExampleServiceExampleDeepRefResult_Success_DEFAULT *DeepRef

func (p *ExampleServiceExampleDeepRefResult) GetSuccess() (v *DeepRef) {
	if !p.IsSetSuccess() {
		return ExampleServiceExampleDeepRefResult_Success_DEFAULT
	}
	return p.Success
}

func (p *ExampleServiceExampleDeepRefResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *ExampleServiceExampleDeepRefResult) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceExampleDeepRefResult(%+v)", *p)
}

var fieldIDToName_ExampleServiceExampleDeepRefResult = map[int16]string{
	0: "success",
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

type ExampleServiceExampleToSnakeCaseArgs struct {
	Req *ExampleToSnakeCase `thrift:"req,1" json:"req"`
}

func NewExampleServiceExampleToSnakeCaseArgs() *ExampleServiceExampleToSnakeCaseArgs {
	return &ExampleServiceExampleToSnakeCaseArgs{}
}

func (p *ExampleServiceExampleToSnakeCaseArgs) InitDefault() {
}

var ExampleServiceExampleToSnakeCaseArgs_Req_DEFAULT *ExampleToSnakeCase

func (p *ExampleServiceExampleToSnakeCaseArgs) GetReq() (v *ExampleToSnakeCase) {
	if !p.IsSetReq() {
		return ExampleServiceExampleToSnakeCaseArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *ExampleServiceExampleToSnakeCaseArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *ExampleServiceExampleToSnakeCaseArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceExampleToSnakeCaseArgs(%+v)", *p)
}

var fieldIDToName_ExampleServiceExampleToSnakeCaseArgs = map[int16]string{
	1: "req",
}

type ExampleServiceExampleToSnakeCaseResult struct {
}

func NewExampleServiceExampleToSnakeCaseResult() *ExampleServiceExampleToSnakeCaseResult {
	return &ExampleServiceExampleToSnakeCaseResult{}
}

func (p *ExampleServiceExampleToSnakeCaseResult) InitDefault() {
}

func (p *ExampleServiceExampleToSnakeCaseResult) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceExampleToSnakeCaseResult(%+v)", *p)
}

var fieldIDToName_ExampleServiceExampleToSnakeCaseResult = map[int16]string{}

type ExampleServiceExampleDefaultValueArgs struct {
	Req *ExampleDefaultValue `thrift:"req,1" json:"req"`
}

func NewExampleServiceExampleDefaultValueArgs() *ExampleServiceExampleDefaultValueArgs {
	return &ExampleServiceExampleDefaultValueArgs{}
}

func (p *ExampleServiceExampleDefaultValueArgs) InitDefault() {
}

var ExampleServiceExampleDefaultValueArgs_Req_DEFAULT *ExampleDefaultValue

func (p *ExampleServiceExampleDefaultValueArgs) GetReq() (v *ExampleDefaultValue) {
	if !p.IsSetReq() {
		return ExampleServiceExampleDefaultValueArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *ExampleServiceExampleDefaultValueArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *ExampleServiceExampleDefaultValueArgs) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceExampleDefaultValueArgs(%+v)", *p)
}

var fieldIDToName_ExampleServiceExampleDefaultValueArgs = map[int16]string{
	1: "req",
}

type ExampleServiceExampleDefaultValueResult struct {
	Success *ExampleDefaultValue `thrift:"success,0,optional" json:"success,omitempty"`
}

func NewExampleServiceExampleDefaultValueResult() *ExampleServiceExampleDefaultValueResult {
	return &ExampleServiceExampleDefaultValueResult{}
}

func (p *ExampleServiceExampleDefaultValueResult) InitDefault() {
}

var ExampleServiceExampleDefaultValueResult_Success_DEFAULT *ExampleDefaultValue

func (p *ExampleServiceExampleDefaultValueResult) GetSuccess() (v *ExampleDefaultValue) {
	if !p.IsSetSuccess() {
		return ExampleServiceExampleDefaultValueResult_Success_DEFAULT
	}
	return p.Success
}

func (p *ExampleServiceExampleDefaultValueResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *ExampleServiceExampleDefaultValueResult) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ExampleServiceExampleDefaultValueResult(%+v)", *p)
}

var fieldIDToName_ExampleServiceExampleDefaultValueResult = map[int16]string{
	0: "success",
}

// exceptions of methods in ExampleService.
var (
	_ error = (*Exception)(nil)
)
