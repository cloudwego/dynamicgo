// Code generated by Kitex v0.7.2. DO NOT EDIT.

package testservice2

import (
	"context"
	example2 "github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example2"
	client "github.com/cloudwego/kitex/client"
	callopt "github.com/cloudwego/kitex/client/callopt"
)

// Client is designed to provide IDL-compatible methods with call-option parameter for kitex framework.
type Client interface {
	ExampleMethod(ctx context.Context, Req *example2.ExampleReq, callOptions ...callopt.Option) (r *example2.ExampleResp, err error)
	ExamplePartialMethod(ctx context.Context, Req *example2.ExampleReqPartial, callOptions ...callopt.Option) (r *example2.A, err error)
	ExamplePartialMethod2(ctx context.Context, Req *example2.ExampleReqPartial, callOptions ...callopt.Option) (r *example2.ExampleRespPartial, err error)
	ExampleSuperMethod(ctx context.Context, Req *example2.ExampleSuper, callOptions ...callopt.Option) (r *example2.A, err error)
	Int2FloatMethod(ctx context.Context, Req *example2.ExampleInt2Float, callOptions ...callopt.Option) (r *example2.ExampleInt2Float, err error)
	Foo(ctx context.Context, Req *example2.A, callOptions ...callopt.Option) (r *example2.A, err error)
	Ping(ctx context.Context, Req *example2.A, callOptions ...callopt.Option) (r *example2.PingResponse, err error)
	Oneway(ctx context.Context, Req *example2.OnewayRequest, callOptions ...callopt.Option) (r *example2.VoidResponse, err error)
	Void(ctx context.Context, Req *example2.VoidRequest, callOptions ...callopt.Option) (r *example2.VoidResponse, err error)
}

// NewClient creates a client for the service defined in IDL.
func NewClient(destService string, opts ...client.Option) (Client, error) {
	var options []client.Option
	options = append(options, client.WithDestService(destService))

	options = append(options, opts...)

	kc, err := client.NewClient(serviceInfo(), options...)
	if err != nil {
		return nil, err
	}
	return &kTestService2Client{
		kClient: newServiceClient(kc),
	}, nil
}

// MustNewClient creates a client for the service defined in IDL. It panics if any error occurs.
func MustNewClient(destService string, opts ...client.Option) Client {
	kc, err := NewClient(destService, opts...)
	if err != nil {
		panic(err)
	}
	return kc
}

type kTestService2Client struct {
	*kClient
}

func (p *kTestService2Client) ExampleMethod(ctx context.Context, Req *example2.ExampleReq, callOptions ...callopt.Option) (r *example2.ExampleResp, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.ExampleMethod(ctx, Req)
}

func (p *kTestService2Client) ExamplePartialMethod(ctx context.Context, Req *example2.ExampleReqPartial, callOptions ...callopt.Option) (r *example2.A, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.ExamplePartialMethod(ctx, Req)
}

func (p *kTestService2Client) ExamplePartialMethod2(ctx context.Context, Req *example2.ExampleReqPartial, callOptions ...callopt.Option) (r *example2.ExampleRespPartial, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.ExamplePartialMethod2(ctx, Req)
}

func (p *kTestService2Client) ExampleSuperMethod(ctx context.Context, Req *example2.ExampleSuper, callOptions ...callopt.Option) (r *example2.A, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.ExampleSuperMethod(ctx, Req)
}

func (p *kTestService2Client) Int2FloatMethod(ctx context.Context, Req *example2.ExampleInt2Float, callOptions ...callopt.Option) (r *example2.ExampleInt2Float, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.Int2FloatMethod(ctx, Req)
}

func (p *kTestService2Client) Foo(ctx context.Context, Req *example2.A, callOptions ...callopt.Option) (r *example2.A, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.Foo(ctx, Req)
}

func (p *kTestService2Client) Ping(ctx context.Context, Req *example2.A, callOptions ...callopt.Option) (r *example2.PingResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.Ping(ctx, Req)
}

func (p *kTestService2Client) Oneway(ctx context.Context, Req *example2.OnewayRequest, callOptions ...callopt.Option) (r *example2.VoidResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.Oneway(ctx, Req)
}

func (p *kTestService2Client) Void(ctx context.Context, Req *example2.VoidRequest, callOptions ...callopt.Option) (r *example2.VoidResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.Void(ctx, Req)
}
