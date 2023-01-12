package thrift

import (
	"net/http"
)

// Route the route annotation
type Route interface {
	Method() string
	Path() string
	Function() *FunctionDescriptor
}

var annoRouteMap = map[string]NewRoute{
	"api.get":    NewAPIGet,
	"api.post":   NewAPIPost,
	"api.put":    NewAPIPut,
	"api.delete": NewAPIDelete,
}

func AnnoToRoute(annoKey string) (NewRoute, bool) {
	if nr, ok := annoRouteMap[annoKey]; ok {
		return nr, true
	}
	return nil, false
}

// NewRoute route creator
type NewRoute func(value string, function *FunctionDescriptor) Route

// NewAPIGet ...
var NewAPIGet NewRoute = func(value string, function *FunctionDescriptor) Route {
	return &apiRoute{http.MethodGet, value, function}
}

// NewAPIPost ...
var NewAPIPost NewRoute = func(value string, function *FunctionDescriptor) Route {
	return &apiRoute{http.MethodPost, value, function}
}

// NewAPIPut ...
var NewAPIPut NewRoute = func(value string, function *FunctionDescriptor) Route {
	return &apiRoute{http.MethodPut, value, function}
}

// NewAPIDelete ...
var NewAPIDelete NewRoute = func(value string, function *FunctionDescriptor) Route {
	return &apiRoute{http.MethodDelete, value, function}
}

type apiRoute struct {
	method   string
	value    string
	function *FunctionDescriptor
}

func (r *apiRoute) Method() string {
	return r.method
}

func (r *apiRoute) Path() string {
	return r.value
}

func (r *apiRoute) Function() *FunctionDescriptor {
	return r.function
}
