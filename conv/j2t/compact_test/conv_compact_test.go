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

package j2t

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"

	stdh "net/http"

	xthrift "github.com/apache/thrift/lib/go/thrift"
	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/conv/j2t"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example3"
	"github.com/cloudwego/dynamicgo/testdata/sample"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	thrift_iter "github.com/thrift-iterator/go"
)

const (
	// exampleIDLPath = "../../testdata/idl/example3.thrift"
	// cur
	exampleIDLPath = "../../../testdata/idl/example3.thrift"
)

var tiCompact = thrift_iter.Config{Protocol: thrift_iter.ProtocolCompact, StaticCodegen: false}.Froze()

func getExampleDesc() *thrift.TypeDescriptor {
	opts := thrift.Options{}
	svc, err := opts.NewDescritorFromPath(context.Background(), exampleIDLPath)
	if err != nil {
		panic(err)
	}
	return svc.Functions()["ExampleMethod"].Request().Struct().FieldById(1).Type()
}

func getExampleReq(exp *example3.ExampleReq, setIs bool, body []byte) *http.HTTPRequest {
	f := -1.00001
	x := true
	q := []string{"1", "2", "3"}
	p := "<>"
	is := "abcd"
	uv := url.Values{
		"query": []string{strings.Join(q, ",")},
	}
	if setIs {
		uv.Add("inner_query", is)
		exp.InnerBase.InnerQuery = is
		exp.InnerBase.ListInnerBase[0].InnerQuery = is
		exp.InnerBase.MapStringInnerBase["innerx"].InnerQuery = is
	}

	uri := "http://localhost:8888/root?" + uv.Encode()
	hr, err := stdh.NewRequest("POST", uri, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	hr.Header.Set("Content-Type", "application/json")
	req, err := http.NewHTTPRequestFromStdReq(hr)
	if err != nil {
		panic(err)
	}
	req.Params.Set("path", p)
	req.Request.Header.Add("heeader", strconv.FormatBool(x))
	req.AddCookie(&stdh.Cookie{Name: "cookie", Value: strconv.FormatFloat(f, 'f', -1, 64)})
	if setIs {
		req.Request.Header.Add("inner_string", is)
		exp.InnerBase.ListInnerBase[0].String_ = is
		exp.InnerBase.MapStringInnerBase["innerx"].String_ = is
		exp.InnerBase.String_ = is
	}
	exp.Path = p
	exp.Query = q
	exp.Header = &x
	exp.Cookie = &f
	exp.RawUri = uri
	return req
}

func TestCompactWriteDefault(t *testing.T) {
	desc := getExampleDesc()
	data := []byte(`{"Path":"<>"}`)
	exp := example3.NewExampleReq()
	data2 := []byte(`{"Path":"<>","Base":{}}`)
	exp.InnerBase = sample.GetEmptyInnerBase3()
	err := json.Unmarshal(data2, exp)
	require.Nil(t, err)
	req := getExampleReq(exp, false, data)

	cv := j2t.NewCompactConv(conv.Options{
		WriteDefaultField: true,
		EnableHttpMapping: true,
	})
	protoF := xthrift.NewTCompactProtocol

	// cv := j2t.NewBinaryConv(conv.Options{
	// 	WriteDefaultField: true,
	// 	EnableHttpMapping: true,
	// })
	// protoF := func(trans xthrift.TTransport) *xthrift.TBinaryProtocol {
	// 	return xthrift.NewTBinaryProtocol(trans, true, true)
	// }

	ctx := context.Background()
	ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
	out, err := cv.Do(ctx, desc, data)

	fmt.Printf("%+#v\n", out)
	require.Nil(t, err)
	act := example3.NewExampleReq()

	var oout bytes.Buffer
	pout := protoF(xthrift.NewStreamTransportW(&oout))
	err = act.Write(pout)
	require.NoError(t, err)
	pout.Flush(nil)
	fmt.Printf("encode: %+#v\n", oout.Bytes())

	// p := thrift.NewCompactProtocol(out)
	pin := protoF(xthrift.NewStreamTransportR(bytes.NewReader(out)))
	err = act.Read(pin)
	// _, err = act.FastRead(out)

	spew.Dump(act)
	require.NoError(t, err)
	require.Equal(t, exp, act)
}
