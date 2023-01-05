/**
 * Copyright 2022 CloudWeGo Authors.
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

package util_test

// Req is a zero value useful request object
type Req struct {
	BodyArr   []byte
	HeaderMap map[string]string
	CookieMap map[string]string
	QueryMap  map[string]string
	ParamMap  map[string]string
	MethodStr string
	PathStr   string
	HostStr   string
	UriStr    string
}

func (r Req) Header(k string) string {
	if r.HeaderMap == nil {
		return ""
	}
	return r.HeaderMap[k]
}

func (r Req) Cookie(k string) string {
	if r.CookieMap == nil {
		return ""
	}
	return r.CookieMap[k]
}

func (r Req) Query(k string) string {
	if r.QueryMap == nil {
		return ""
	}
	return r.QueryMap[k]
}

func (r Req) Param(k string) string {
	if r.ParamMap == nil {
		return ""
	}
	return r.ParamMap[k]
}

func (r Req) Body() []byte {
	return r.BodyArr
}

func (r Req) PostForm(key string) string {
	return ""
}

func (r Req) MapBody(key string) string {
	return ""
}

func (r Req) Method() string {
	return r.MethodStr
}

func (r Req) Path() string {
	return r.PathStr
}

func (r Req) Host() string {
	return r.HostStr
}

func (r Req) Uri() string {
	return r.UriStr
}
