// Code generated by Kitex v0.7.2. DO NOT EDIT.

package baselineservice

import (
	"context"
	baseline "github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/baseline"
	client "github.com/cloudwego/kitex/client"
	kitex "github.com/cloudwego/kitex/pkg/serviceinfo"
	streaming "github.com/cloudwego/kitex/pkg/streaming"
	proto "google.golang.org/protobuf/proto"
)

func serviceInfo() *kitex.ServiceInfo {
	return baselineServiceServiceInfo
}

var baselineServiceServiceInfo = NewServiceInfo()

func NewServiceInfo() *kitex.ServiceInfo {
	serviceName := "BaselineService"
	handlerType := (*baseline.BaselineService)(nil)
	methods := map[string]kitex.MethodInfo{
		"SimpleMethod":         kitex.NewMethodInfo(simpleMethodHandler, newSimpleMethodArgs, newSimpleMethodResult, false),
		"PartialSimpleMethod":  kitex.NewMethodInfo(partialSimpleMethodHandler, newPartialSimpleMethodArgs, newPartialSimpleMethodResult, false),
		"NestingMethod":        kitex.NewMethodInfo(nestingMethodHandler, newNestingMethodArgs, newNestingMethodResult, false),
		"PartialNestingMethod": kitex.NewMethodInfo(partialNestingMethodHandler, newPartialNestingMethodArgs, newPartialNestingMethodResult, false),
		"Nesting2Method":       kitex.NewMethodInfo(nesting2MethodHandler, newNesting2MethodArgs, newNesting2MethodResult, false),
	}
	extra := map[string]interface{}{
		"PackageName":     "pb3",
		"ServiceFilePath": ``,
	}
	svcInfo := &kitex.ServiceInfo{
		ServiceName:     serviceName,
		HandlerType:     handlerType,
		Methods:         methods,
		PayloadCodec:    kitex.Protobuf,
		KiteXGenVersion: "v0.7.2",
		Extra:           extra,
	}
	return svcInfo
}

func simpleMethodHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	switch s := arg.(type) {
	case *streaming.Args:
		st := s.Stream
		req := new(baseline.Simple)
		if err := st.RecvMsg(req); err != nil {
			return err
		}
		resp, err := handler.(baseline.BaselineService).SimpleMethod(ctx, req)
		if err != nil {
			return err
		}
		if err := st.SendMsg(resp); err != nil {
			return err
		}
	case *SimpleMethodArgs:
		success, err := handler.(baseline.BaselineService).SimpleMethod(ctx, s.Req)
		if err != nil {
			return err
		}
		realResult := result.(*SimpleMethodResult)
		realResult.Success = success
	}
	return nil
}
func newSimpleMethodArgs() interface{} {
	return &SimpleMethodArgs{}
}

func newSimpleMethodResult() interface{} {
	return &SimpleMethodResult{}
}

type SimpleMethodArgs struct {
	Req *baseline.Simple
}

func (p *SimpleMethodArgs) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new(baseline.Simple)
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *SimpleMethodArgs) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *SimpleMethodArgs) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}

func (p *SimpleMethodArgs) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
		return out, nil
	}
	return proto.Marshal(p.Req)
}

func (p *SimpleMethodArgs) Unmarshal(in []byte) error {
	msg := new(baseline.Simple)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Req = msg
	return nil
}

var SimpleMethodArgs_Req_DEFAULT *baseline.Simple

func (p *SimpleMethodArgs) GetReq() *baseline.Simple {
	if !p.IsSetReq() {
		return SimpleMethodArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *SimpleMethodArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *SimpleMethodArgs) GetFirstArgument() interface{} {
	return p.Req
}

type SimpleMethodResult struct {
	Success *baseline.Simple
}

var SimpleMethodResult_Success_DEFAULT *baseline.Simple

func (p *SimpleMethodResult) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new(baseline.Simple)
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *SimpleMethodResult) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *SimpleMethodResult) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}

func (p *SimpleMethodResult) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
		return out, nil
	}
	return proto.Marshal(p.Success)
}

func (p *SimpleMethodResult) Unmarshal(in []byte) error {
	msg := new(baseline.Simple)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Success = msg
	return nil
}

func (p *SimpleMethodResult) GetSuccess() *baseline.Simple {
	if !p.IsSetSuccess() {
		return SimpleMethodResult_Success_DEFAULT
	}
	return p.Success
}

func (p *SimpleMethodResult) SetSuccess(x interface{}) {
	p.Success = x.(*baseline.Simple)
}

func (p *SimpleMethodResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *SimpleMethodResult) GetResult() interface{} {
	return p.Success
}

func partialSimpleMethodHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	switch s := arg.(type) {
	case *streaming.Args:
		st := s.Stream
		req := new(baseline.PartialSimple)
		if err := st.RecvMsg(req); err != nil {
			return err
		}
		resp, err := handler.(baseline.BaselineService).PartialSimpleMethod(ctx, req)
		if err != nil {
			return err
		}
		if err := st.SendMsg(resp); err != nil {
			return err
		}
	case *PartialSimpleMethodArgs:
		success, err := handler.(baseline.BaselineService).PartialSimpleMethod(ctx, s.Req)
		if err != nil {
			return err
		}
		realResult := result.(*PartialSimpleMethodResult)
		realResult.Success = success
	}
	return nil
}
func newPartialSimpleMethodArgs() interface{} {
	return &PartialSimpleMethodArgs{}
}

func newPartialSimpleMethodResult() interface{} {
	return &PartialSimpleMethodResult{}
}

type PartialSimpleMethodArgs struct {
	Req *baseline.PartialSimple
}

func (p *PartialSimpleMethodArgs) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new(baseline.PartialSimple)
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *PartialSimpleMethodArgs) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *PartialSimpleMethodArgs) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}

func (p *PartialSimpleMethodArgs) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
		return out, nil
	}
	return proto.Marshal(p.Req)
}

func (p *PartialSimpleMethodArgs) Unmarshal(in []byte) error {
	msg := new(baseline.PartialSimple)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Req = msg
	return nil
}

var PartialSimpleMethodArgs_Req_DEFAULT *baseline.PartialSimple

func (p *PartialSimpleMethodArgs) GetReq() *baseline.PartialSimple {
	if !p.IsSetReq() {
		return PartialSimpleMethodArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *PartialSimpleMethodArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *PartialSimpleMethodArgs) GetFirstArgument() interface{} {
	return p.Req
}

type PartialSimpleMethodResult struct {
	Success *baseline.PartialSimple
}

var PartialSimpleMethodResult_Success_DEFAULT *baseline.PartialSimple

func (p *PartialSimpleMethodResult) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new(baseline.PartialSimple)
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *PartialSimpleMethodResult) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *PartialSimpleMethodResult) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}

func (p *PartialSimpleMethodResult) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
		return out, nil
	}
	return proto.Marshal(p.Success)
}

func (p *PartialSimpleMethodResult) Unmarshal(in []byte) error {
	msg := new(baseline.PartialSimple)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Success = msg
	return nil
}

func (p *PartialSimpleMethodResult) GetSuccess() *baseline.PartialSimple {
	if !p.IsSetSuccess() {
		return PartialSimpleMethodResult_Success_DEFAULT
	}
	return p.Success
}

func (p *PartialSimpleMethodResult) SetSuccess(x interface{}) {
	p.Success = x.(*baseline.PartialSimple)
}

func (p *PartialSimpleMethodResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *PartialSimpleMethodResult) GetResult() interface{} {
	return p.Success
}

func nestingMethodHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	switch s := arg.(type) {
	case *streaming.Args:
		st := s.Stream
		req := new(baseline.Nesting)
		if err := st.RecvMsg(req); err != nil {
			return err
		}
		resp, err := handler.(baseline.BaselineService).NestingMethod(ctx, req)
		if err != nil {
			return err
		}
		if err := st.SendMsg(resp); err != nil {
			return err
		}
	case *NestingMethodArgs:
		success, err := handler.(baseline.BaselineService).NestingMethod(ctx, s.Req)
		if err != nil {
			return err
		}
		realResult := result.(*NestingMethodResult)
		realResult.Success = success
	}
	return nil
}
func newNestingMethodArgs() interface{} {
	return &NestingMethodArgs{}
}

func newNestingMethodResult() interface{} {
	return &NestingMethodResult{}
}

type NestingMethodArgs struct {
	Req *baseline.Nesting
}

func (p *NestingMethodArgs) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new(baseline.Nesting)
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *NestingMethodArgs) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *NestingMethodArgs) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}

func (p *NestingMethodArgs) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
		return out, nil
	}
	return proto.Marshal(p.Req)
}

func (p *NestingMethodArgs) Unmarshal(in []byte) error {
	msg := new(baseline.Nesting)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Req = msg
	return nil
}

var NestingMethodArgs_Req_DEFAULT *baseline.Nesting

func (p *NestingMethodArgs) GetReq() *baseline.Nesting {
	if !p.IsSetReq() {
		return NestingMethodArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *NestingMethodArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *NestingMethodArgs) GetFirstArgument() interface{} {
	return p.Req
}

type NestingMethodResult struct {
	Success *baseline.Nesting
}

var NestingMethodResult_Success_DEFAULT *baseline.Nesting

func (p *NestingMethodResult) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new(baseline.Nesting)
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *NestingMethodResult) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *NestingMethodResult) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}

func (p *NestingMethodResult) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
		return out, nil
	}
	return proto.Marshal(p.Success)
}

func (p *NestingMethodResult) Unmarshal(in []byte) error {
	msg := new(baseline.Nesting)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Success = msg
	return nil
}

func (p *NestingMethodResult) GetSuccess() *baseline.Nesting {
	if !p.IsSetSuccess() {
		return NestingMethodResult_Success_DEFAULT
	}
	return p.Success
}

func (p *NestingMethodResult) SetSuccess(x interface{}) {
	p.Success = x.(*baseline.Nesting)
}

func (p *NestingMethodResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *NestingMethodResult) GetResult() interface{} {
	return p.Success
}

func partialNestingMethodHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	switch s := arg.(type) {
	case *streaming.Args:
		st := s.Stream
		req := new(baseline.PartialNesting)
		if err := st.RecvMsg(req); err != nil {
			return err
		}
		resp, err := handler.(baseline.BaselineService).PartialNestingMethod(ctx, req)
		if err != nil {
			return err
		}
		if err := st.SendMsg(resp); err != nil {
			return err
		}
	case *PartialNestingMethodArgs:
		success, err := handler.(baseline.BaselineService).PartialNestingMethod(ctx, s.Req)
		if err != nil {
			return err
		}
		realResult := result.(*PartialNestingMethodResult)
		realResult.Success = success
	}
	return nil
}
func newPartialNestingMethodArgs() interface{} {
	return &PartialNestingMethodArgs{}
}

func newPartialNestingMethodResult() interface{} {
	return &PartialNestingMethodResult{}
}

type PartialNestingMethodArgs struct {
	Req *baseline.PartialNesting
}

func (p *PartialNestingMethodArgs) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new(baseline.PartialNesting)
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *PartialNestingMethodArgs) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *PartialNestingMethodArgs) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}

func (p *PartialNestingMethodArgs) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
		return out, nil
	}
	return proto.Marshal(p.Req)
}

func (p *PartialNestingMethodArgs) Unmarshal(in []byte) error {
	msg := new(baseline.PartialNesting)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Req = msg
	return nil
}

var PartialNestingMethodArgs_Req_DEFAULT *baseline.PartialNesting

func (p *PartialNestingMethodArgs) GetReq() *baseline.PartialNesting {
	if !p.IsSetReq() {
		return PartialNestingMethodArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *PartialNestingMethodArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *PartialNestingMethodArgs) GetFirstArgument() interface{} {
	return p.Req
}

type PartialNestingMethodResult struct {
	Success *baseline.PartialNesting
}

var PartialNestingMethodResult_Success_DEFAULT *baseline.PartialNesting

func (p *PartialNestingMethodResult) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new(baseline.PartialNesting)
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *PartialNestingMethodResult) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *PartialNestingMethodResult) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}

func (p *PartialNestingMethodResult) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
		return out, nil
	}
	return proto.Marshal(p.Success)
}

func (p *PartialNestingMethodResult) Unmarshal(in []byte) error {
	msg := new(baseline.PartialNesting)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Success = msg
	return nil
}

func (p *PartialNestingMethodResult) GetSuccess() *baseline.PartialNesting {
	if !p.IsSetSuccess() {
		return PartialNestingMethodResult_Success_DEFAULT
	}
	return p.Success
}

func (p *PartialNestingMethodResult) SetSuccess(x interface{}) {
	p.Success = x.(*baseline.PartialNesting)
}

func (p *PartialNestingMethodResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *PartialNestingMethodResult) GetResult() interface{} {
	return p.Success
}

func nesting2MethodHandler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	switch s := arg.(type) {
	case *streaming.Args:
		st := s.Stream
		req := new(baseline.Nesting2)
		if err := st.RecvMsg(req); err != nil {
			return err
		}
		resp, err := handler.(baseline.BaselineService).Nesting2Method(ctx, req)
		if err != nil {
			return err
		}
		if err := st.SendMsg(resp); err != nil {
			return err
		}
	case *Nesting2MethodArgs:
		success, err := handler.(baseline.BaselineService).Nesting2Method(ctx, s.Req)
		if err != nil {
			return err
		}
		realResult := result.(*Nesting2MethodResult)
		realResult.Success = success
	}
	return nil
}
func newNesting2MethodArgs() interface{} {
	return &Nesting2MethodArgs{}
}

func newNesting2MethodResult() interface{} {
	return &Nesting2MethodResult{}
}

type Nesting2MethodArgs struct {
	Req *baseline.Nesting2
}

func (p *Nesting2MethodArgs) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new(baseline.Nesting2)
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *Nesting2MethodArgs) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *Nesting2MethodArgs) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}

func (p *Nesting2MethodArgs) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
		return out, nil
	}
	return proto.Marshal(p.Req)
}

func (p *Nesting2MethodArgs) Unmarshal(in []byte) error {
	msg := new(baseline.Nesting2)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Req = msg
	return nil
}

var Nesting2MethodArgs_Req_DEFAULT *baseline.Nesting2

func (p *Nesting2MethodArgs) GetReq() *baseline.Nesting2 {
	if !p.IsSetReq() {
		return Nesting2MethodArgs_Req_DEFAULT
	}
	return p.Req
}

func (p *Nesting2MethodArgs) IsSetReq() bool {
	return p.Req != nil
}

func (p *Nesting2MethodArgs) GetFirstArgument() interface{} {
	return p.Req
}

type Nesting2MethodResult struct {
	Success *baseline.Nesting2
}

var Nesting2MethodResult_Success_DEFAULT *baseline.Nesting2

func (p *Nesting2MethodResult) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new(baseline.Nesting2)
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *Nesting2MethodResult) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *Nesting2MethodResult) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}

func (p *Nesting2MethodResult) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
		return out, nil
	}
	return proto.Marshal(p.Success)
}

func (p *Nesting2MethodResult) Unmarshal(in []byte) error {
	msg := new(baseline.Nesting2)
	if err := proto.Unmarshal(in, msg); err != nil {
		return err
	}
	p.Success = msg
	return nil
}

func (p *Nesting2MethodResult) GetSuccess() *baseline.Nesting2 {
	if !p.IsSetSuccess() {
		return Nesting2MethodResult_Success_DEFAULT
	}
	return p.Success
}

func (p *Nesting2MethodResult) SetSuccess(x interface{}) {
	p.Success = x.(*baseline.Nesting2)
}

func (p *Nesting2MethodResult) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *Nesting2MethodResult) GetResult() interface{} {
	return p.Success
}

type kClient struct {
	c client.Client
}

func newServiceClient(c client.Client) *kClient {
	return &kClient{
		c: c,
	}
}

func (p *kClient) SimpleMethod(ctx context.Context, Req *baseline.Simple) (r *baseline.Simple, err error) {
	var _args SimpleMethodArgs
	_args.Req = Req
	var _result SimpleMethodResult
	if err = p.c.Call(ctx, "SimpleMethod", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
}

func (p *kClient) PartialSimpleMethod(ctx context.Context, Req *baseline.PartialSimple) (r *baseline.PartialSimple, err error) {
	var _args PartialSimpleMethodArgs
	_args.Req = Req
	var _result PartialSimpleMethodResult
	if err = p.c.Call(ctx, "PartialSimpleMethod", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
}

func (p *kClient) NestingMethod(ctx context.Context, Req *baseline.Nesting) (r *baseline.Nesting, err error) {
	var _args NestingMethodArgs
	_args.Req = Req
	var _result NestingMethodResult
	if err = p.c.Call(ctx, "NestingMethod", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
}

func (p *kClient) PartialNestingMethod(ctx context.Context, Req *baseline.PartialNesting) (r *baseline.PartialNesting, err error) {
	var _args PartialNestingMethodArgs
	_args.Req = Req
	var _result PartialNestingMethodResult
	if err = p.c.Call(ctx, "PartialNestingMethod", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
}

func (p *kClient) Nesting2Method(ctx context.Context, Req *baseline.Nesting2) (r *baseline.Nesting2, err error) {
	var _args Nesting2MethodArgs
	_args.Req = Req
	var _result Nesting2MethodResult
	if err = p.c.Call(ctx, "Nesting2Method", &_args, &_result); err != nil {
		return
	}
	return _result.GetSuccess(), nil
}
