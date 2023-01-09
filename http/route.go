package http

import (
	"net/http"

	"github.com/cloudwego/dynamicgo/thrift"
)

// Route the route annotation
type Route interface {
	Method() string
	Path() string
	Function() *thrift.FunctionDescriptor
}

// NewRoute route creator
type NewRoute func(value string, function *thrift.FunctionDescriptor) Route

// NewAPIGet ...
var NewAPIGet NewRoute = func(value string, function *thrift.FunctionDescriptor) Route {
	return &apiRoute{http.MethodGet, value, function}
}

// NewAPIPost ...
var NewAPIPost NewRoute = func(value string, function *thrift.FunctionDescriptor) Route {
	return &apiRoute{http.MethodPost, value, function}
}

// NewAPIPut ...
var NewAPIPut NewRoute = func(value string, function *thrift.FunctionDescriptor) Route {
	return &apiRoute{http.MethodPut, value, function}
}

// NewAPIDelete ...
var NewAPIDelete NewRoute = func(value string, function *thrift.FunctionDescriptor) Route {
	return &apiRoute{http.MethodDelete, value, function}
}

type apiRoute struct {
	method   string
	value    string
	function *thrift.FunctionDescriptor
}

func (r *apiRoute) Method() string {
	return r.method
}

func (r *apiRoute) Path() string {
	return r.value
}

func (r *apiRoute) Function() *thrift.FunctionDescriptor {
	return r.function
}
