/**
 * Copyright 2023 CloudWeGo Authors.
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

func (r Req) GetHeader(k string) string {
	if r.HeaderMap == nil {
		return ""
	}
	return r.HeaderMap[k]
}

func (r Req) GetCookie(k string) string {
	if r.CookieMap == nil {
		return ""
	}
	return r.CookieMap[k]
}

func (r Req) GetQuery(k string) string {
	if r.QueryMap == nil {
		return ""
	}
	return r.QueryMap[k]
}

func (r Req) GetParam(k string) string {
	if r.ParamMap == nil {
		return ""
	}
	return r.ParamMap[k]
}

func (r Req) GetBody() []byte {
	return r.BodyArr
}

func (r Req) GetPostForm(key string) string {
	return ""
}

func (r Req) GetMapBody(key string) string {
	return ""
}

func (r Req) GetMethod() string {
	return r.MethodStr
}

func (r Req) GetPath() string {
	return r.PathStr
}

func (r Req) GetHost() string {
	return r.HostStr
}

func (r Req) GetUri() string {
	return r.UriStr
}
