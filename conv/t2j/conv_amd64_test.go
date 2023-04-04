//go:build amd64 && go1.16
// +build amd64,go1.16

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

package t2j

import (
	"context"
	"testing"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestConvThrift2HTTP(t *testing.T) {
	desc := thrift.FnResponse(thrift.GetFnDescFromFile("testdata/idl/example3.thrift", "ExampleMethod", thrift.Options{}))
	expJSON := getExmaple3JSON()
	cv := NewBinaryConv(conv.Options{
		// MapRecurseDepth:    conv.DefaultMaxDepth,
		EnableValueMapping: true,
		EnableHttpMapping:  true,
		OmitHttpMappingErrors: true,
	})
	in := getExample3Data()
	ctx := context.Background()
	resp := http.NewHTTPResponse()
	ctx = context.WithValue(ctx, conv.CtxKeyHTTPResponse, resp)
	out, err := cv.Do(ctx, desc, in)
	if err != nil {
		t.Fatal(err)
	}
	println(expJSON, string(out))
	chechHelper(t, expJSON, string(out), testOptions{
		JSConv: true,
	})
	// buf, err := io.ReadAll(resp.Response.Body)
	// require.Nil(t, err)
	// chechHelper(t, expJSON, string(buf))
	assert.Equal(t, "true", resp.Header["Heeader"][0])
	if len(resp.Header) != 2 {
		spew.Dump(resp.Header)
	}
	assert.Equal(t, 2, len(resp.Header))
	assert.Equal(t, "cookie", resp.Response.Cookies()[0].Name)
	assert.Equal(t, "-1e-7", resp.Response.Cookies()[0].Value)
	assert.Equal(t, "inner_string", resp.Response.Cookies()[1].Name)
	assert.Equal(t, "hello", resp.Response.Cookies()[1].Value)

	ctx = context.Background()
	resp = http.NewHTTPResponse()
	ctx = context.WithValue(ctx, conv.CtxKeyHTTPResponse, resp)
	out, err = cv.Do(ctx, desc, in)
	if err != nil {
		t.Fatal(err)
	}
	chechHelper(t, expJSON, string(out), testOptions{
		JSConv: true,
	})
	// buf, err = io.ReadAll(resp.Response.Body)
	// require.Nil(t, err)
	// chechHelper(t, expJSON, string(buf))
	assert.Equal(t, "true", resp.Header["Heeader"][0])
	assert.Equal(t, 2, len(resp.Header))
	assert.Equal(t, "cookie", resp.Response.Cookies()[0].Name)
	assert.Equal(t, "-1e-7", resp.Response.Cookies()[0].Value)
	assert.Equal(t, "inner_string", resp.Response.Cookies()[1].Name)
	assert.Equal(t, "hello", resp.Response.Cookies()[1].Value)
}
