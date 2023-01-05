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

package http

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONBody(t *testing.T) {
	var data = []byte(`{"name":"foo","age":18}`)
	req, err := http.NewRequest("POST", "http://localhost:8080", bytes.NewBuffer(data))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	r, err := NewHTTPRequestFromStdReq(req)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, `foo`, r.MapBody("name"))
	require.Equal(t, "18", r.MapBody("age"))
}

func TestPostFormBody(t *testing.T) {
	var data = url.Values{
		"name": {"{\"foo\":\"bar\"}", "bar"},
		"age":  {"18"},
	}.Encode()
	req, err := http.NewRequest("POST", "http://localhost:8080", strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r, err := NewHTTPRequestFromStdReq(req)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, "{\"foo\":\"bar\"}", r.MapBody("name"))
	require.Equal(t, "{\"foo\":\"bar\"}", r.FormValue("name"))
	require.Equal(t, "18", r.MapBody("age"))
	require.Equal(t, "18", r.FormValue("age"))
}
