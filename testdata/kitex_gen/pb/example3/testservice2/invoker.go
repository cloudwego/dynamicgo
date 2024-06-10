// Code generated by Kitex v0.9.1. DO NOT EDIT.

package testservice2

import (
	example3 "github.com/cloudwego/dynamicgo/testdata/kitex_gen/pb/example3"
	server "github.com/cloudwego/kitex/server"
)

// NewInvoker creates a server.Invoker with the given handler and options.
func NewInvoker(handler example3.TestService2, opts ...server.Option) server.Invoker {
	var options []server.Option

	options = append(options, opts...)

	s := server.NewInvoker(options...)
	if err := s.RegisterService(serviceInfo(), handler); err != nil {
		panic(err)
	}
	if err := s.Init(); err != nil {
		panic(err)
	}
	return s
}