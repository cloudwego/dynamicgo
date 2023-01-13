package thrift

import (
	"fmt"
	"sync"

	"github.com/cloudwego/dynamicgo/http"
)

// Router http router
type Router interface {
	// Handle register Route to Router
	Handle(rt Route)
	// Lookup FunctionDescriptor from HTTPRequest
	Lookup(req *http.HTTPRequest) (*FunctionDescriptor, error)
}

type router struct {
	trees      map[string]*node
	maxParams  uint16
	paramsPool sync.Pool
}

// NewRouter ...
func NewRouter() Router {
	return &router{}
}

func (r *router) getParams() *http.Params {
	ps, _ := r.paramsPool.Get().(*http.Params)
	ps.Params = ps.Params[0:0] // reset slice
	return ps
}

func (r *router) putParams(ps *http.Params) {
	if ps != nil {
		r.paramsPool.Put(ps)
	}
}

func (r *router) Handle(rt Route) {
	method := rt.Method()
	path := rt.Path()
	function := rt.Function()
	if method == "" {
		panic("method must not be empty")
	}
	if len(path) < 1 || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}
	if function == nil {
		panic("function descriptor must not be nil")
	}

	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root
	}

	root.addRoute(path, function)

	if paramsCount := countParams(path); paramsCount > r.maxParams {
		r.maxParams = paramsCount
	}

	// Lazy-init paramsPool alloc func
	if r.paramsPool.New == nil && r.maxParams > 0 {
		r.paramsPool.New = func() interface{} {
			ps := http.Params{
				Params:        make([]http.Param, 0, r.maxParams),
				RecycleParams: r.putParams,
			}
			return &ps
		}
	}
}

func (r *router) Lookup(req *http.HTTPRequest) (*FunctionDescriptor, error) {
	root, ok := r.trees[req.Method()]
	if !ok {
		return nil, fmt.Errorf("function lookup failed, no root with method=%s", req.Method())
	}
	fn, ps, _ := root.getValue(req.Path(), r.getParams, false)
	if fn == nil {
		r.putParams(ps)
		return nil, fmt.Errorf("function lookup failed, path=%s", req.Path())
	}
	if ps != nil {
		for _, param := range ps.Params {
			req.Params.Set(param.Key, param.Value)
		}
	}
	return fn, nil
}
