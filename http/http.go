/*
 * Copyright 2021 CloudWeGo Authors
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

package http

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/bytedance/sonic/ast"
	"github.com/cloudwego/dynamicgo/internal/rt"
)

// Endpoint a http endpoint.
type Endpoint struct {
	Method, Path string
}

var annoMethoMap = map[string]string{
	"api.get":     http.MethodGet,
	"api.post":    http.MethodPost,
	"api.put":     http.MethodPut,
	"api.delete":  http.MethodDelete,
	"api.patch":   http.MethodPatch,
	"api.head":    http.MethodHead,
	"api.options": http.MethodOptions,
	"api.connect": http.MethodConnect,
	"api.trace":   http.MethodTrace,
}

// AnnoToMethod maps annotation to corresponding http method
func AnnoToMethod(annoKey string) string {
	return annoMethoMap[annoKey]
}

// Param in url path
//
// e.g. /user/:id + /user/123 => Param{Key: "id", Value: "123"}
type Param struct {
	Key   string
	Value string
}

// RequestGetter is a interface for getting request parameters
type RequestGetter interface {
	// GetMethod returns the http method.
	GetMethod() string
	// GetHost returns the host.
	GetHost() string
	// GetUri returns entire uri.
	GetUri() string
	// Header returns the value of the header with the given key.
	GetHeader(string) string
	// Cookie returns the value of the cookie with the given key.
	GetCookie(string) string
	// Query returns the value of the query with the given key.
	GetQuery(string) string
	// Param returns the value of the url-path param with the given key.
	GetParam(string) string
	// PostForm returns the value of the post-form body with the given key.
	GetPostForm(string) string
	// MapBody returns the value of body with the given key.
	GetMapBody(string) string
	// Body returns the raw body in bytes.
	GetBody() []byte
}

// Request is a implementation of RequestGetter.
// It wraps http.Request.
type HTTPRequest struct {
	*http.Request
	rawBody []byte
	Params  Params
	BodyMap interface{}
}

func (h *HTTPRequest) String() string {
	return fmt.Sprintf(`URL: %s
Headers: %v
Body: %s
BodyMap: %v`, h.URL.String(), h.Header, string(h.rawBody), h.BodyMap)
}

// NewHTTPRequest creates a new HTTPRequest.
func NewHTTPRequest() *HTTPRequest {
	return &HTTPRequest{}
}

// NewHTTPRequestFromUrl creates a new HTTPRequest from url, body and url-path param.
func NewHTTPRequestFromUrl(method, url string, body io.Reader, params ...Param) (*HTTPRequest, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	ret := &HTTPRequest{
		Request: req,
	}
	for _, p := range params {
		ret.Params.Set(p.Key, p.Value)
	}
	return ret, nil
}

var (
	// DefaultJsonPairSize is the default size of json.Pair slice.
	DefaultJsonPairSize = 16
)

var jsonPairsPool = sync.Pool{
	New: func() interface{} {
		ret := make([]ast.Pair, DefaultJsonPairSize)
		return &ret
	},
}

type jsonCache struct {
	root ast.Node
	m   map[string]string
}

// NewHTTPRequestFromStdReq creates a new HTTPRequest from http.Request.
// It will check the content-type of the request and parse the body if the type one of following:
//   - application/json
//   - application/x-www-form-urlencoded
func NewHTTPRequestFromStdReq(req *http.Request, params ...Param) (ret *HTTPRequest, err error) {
	ret = &HTTPRequest{}
	ret.Request = req
	for _, p := range params {
		ret.Params.Set(p.Key, p.Value)
	}

	switch req.Header.Get(HeaderContentType) {
	case "application/json":
		{
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			ret.rawBody = body
			if len(body) == 0 {
				return ret, nil
			}
			node := ast.NewRaw(rt.Mem2Str(body))
			cache := &jsonCache{
				root: node,
				m:    make(map[string]string),
			}
			ret.BodyMap = cache
			// parser := json.NewParser(rt.Mem2Str(body))
			// parser.Set(true, false)
			// vs := jsonPairsPool.New().(*[]json.Pair)
			// if err := parser.DecodeObject(vs); err != nil {
			// 	return nil, err
			// }
			// // TODO: reuse map memory?
			// ret.BodyMap = make(map[string]string, len(*vs))
			// for _, kv := range *vs {
			// 	js, _ := kv.Value.Raw()
			// 	ret.BodyMap[kv.Key] = js
			// }
			// (*vs) = (*vs)[:0]
			// jsonPairsPool.Put(vs)
		}
	case "application/x-www-form-urlencoded":
		if err := req.ParseForm(); err != nil {
			return nil, err
		}
		ret.BodyMap = req.PostForm
	}
	return ret, nil
}

// Header implements RequestGetter.Header.
func (self HTTPRequest) GetHeader(key string) string {
	return self.Request.Header.Get(key)
}

// Cookie implements RequestGetter.Cookie.
func (self HTTPRequest) GetCookie(key string) string {
	if c, err := self.Request.Cookie(key); err == nil {
		return c.Value
	}
	return ""
}

// Query implements RequestGetter.Query.
func (self HTTPRequest) GetQuery(key string) string {
	return self.Request.URL.Query().Get(key)
}

// Body implements RequestGetter.Body.
func (self HTTPRequest) GetBody() []byte {
	if self.rawBody != nil {
		return self.rawBody
	}
	buf, err := ioutil.ReadAll(self.Request.Body)
	if err != nil {
		return nil
	}
	return buf
}

// Method implements RequestGetter.Method.
func (self HTTPRequest) GetMethod() string {
	return self.Request.Method
}

// Path implements RequestGetter.Path.
func (self HTTPRequest) GetPath() string {
	return self.Request.URL.Path
}

// Host implements RequestGetter.Host.
func (self HTTPRequest) GetHost() string {
	return self.Request.URL.Host
}

// Param implements RequestGetter.Param.
func (self HTTPRequest) GetParam(key string) string {
	return self.Params.ByName(key)
}

// MapBody implements RequestGetter.MapBody.
func (self *HTTPRequest) GetMapBody(key string) string {
	if self.BodyMap == nil && self.Request != nil {
		v, err := NewHTTPRequestFromStdReq(self.Request)
		if err != nil || v.BodyMap == nil {
			return ""
		}
		self.BodyMap = v.BodyMap
	}
	switch t := self.BodyMap.(type) {
	case *jsonCache:
		// fast path
		if v, ok := t.m[key]; ok {
			return v
		}
		// slow path
		v := t.root.Get(key)
		if v.Check() != nil {
			return ""
		}
		j, e := v.Raw()
		if e != nil {
			return ""
		}
		if v.Type() == ast.V_STRING {
			j, e = v.String()
			if e != nil {
				return ""
			}
		}
		t.m[key] = j
		return j
	case map[string]string:
		return t[key]
	case url.Values:
		return t.Get(key)
	default:
		return ""
	}
}

// PostForm implements RequestGetter.PostForm.
func (self HTTPRequest) GetPostForm(key string) string {
	return self.Request.PostFormValue(key)
}

// Uri implements RequestGetter.Uri.
func (self HTTPRequest) GetUri() string {
	return self.Request.URL.String()
}

// ResponseSetter is a interface for setting response parameters
type ResponseSetter interface {
	// SetStatusCode sets the status code of the response
	SetStatusCode(int) error
	// SetHeader sets the header of the response
	SetHeader(string, string) error
	// SetCookie sets the cookie of the response
	SetCookie(string, string) error
	// SetRawBody sets the raw body of the response
	SetRawBody([]byte) error
}

// HTTPResponse is an implementation of ResponseSetter
type HTTPResponse struct {
	*http.Response
}

// NewHTTPResponse creates a new HTTPResponse
func NewHTTPResponse() *HTTPResponse {
	return &HTTPResponse{
		Response: &http.Response{
			Header: make(http.Header),
		},
	}
}

// SetStatusCode implements ResponseSetter.SetStatusCode
func (self HTTPResponse) SetStatusCode(code int) error {
	self.Response.StatusCode = code
	return nil
}

// SetHeader implements ResponseSetter.SetHeader
func (self HTTPResponse) SetHeader(key string, val string) error {
	self.Response.Header.Set(key, val)
	return nil
}

// SetCookie implements ResponseSetter.SetCookie
func (self HTTPResponse) SetCookie(key string, val string) error {
	c := &http.Cookie{Name: key, Value: val}
	self.Response.Header.Add(HeaderSetCookie, c.String())
	return nil
}

type wrapBody struct {
	io.Reader
}

func (wrapBody) Close() error { return nil }

func (self HTTPResponse) SetRawBody(body []byte) error {
	self.Response.Body = wrapBody{bytes.NewReader(body)}
	return nil
}

const (
	// HeaderContentType is the key of Content-Type header
	HeaderContentType = "Content-Type"
	// HeaderSetCookie is the key of Set-Cookie header
	HeaderSetCookie = "Set-Cookie"
)

// Http url-path params
type Params struct {
	params   []Param
	recycle  func(*Params)
	recycled bool
}

// Recycle the Params
func (ps *Params) Recycle() {
	if ps.recycled {
		return
	}
	ps.recycled = true
	ps.recycle(ps)
}

// ByName search Param by given name
func (ps *Params) ByName(name string) string {
	for _, p := range ps.params {
		if p.Key == name {
			return p.Value
		}
	}
	return ""
}

// Set set Param by given name and value, return true if Param exists
func (ps *Params) Set(name string, val string) bool {
	for i, p := range ps.params {
		if p.Key == name {
			ps.params[i].Value = val
			return true
		}
	}
	ps.params = append(ps.params, Param{Key: name, Value: val})
	return false
}
