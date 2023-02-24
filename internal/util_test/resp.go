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

type Resp struct {
	Body       []byte
	StatusCode int
	Header     map[string]string
	Cookie     map[string]string
}

func (r *Resp) SetStatusCode(code int) error {
	r.StatusCode = code
	return nil
}
func (r *Resp) SetHeader(key string, val string) error {
	if r.Header == nil {
		r.Header = map[string]string{}
	}
	r.Header[key] = val
	return nil
}
func (r *Resp) SetCookie(key string, val string) error {
	if r.Cookie == nil {
		r.Cookie = map[string]string{}
	}
	r.Cookie[key] = val
	return nil
}
func (r *Resp) SetRawBody(body []byte) error {
	r.Body = body
	return nil
}
