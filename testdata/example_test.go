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

package testdata

import (
	ejson "encoding/json"
	"os"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/base"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example2"
	"github.com/cloudwego/dynamicgo/testdata/sample"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
)

var (
	genData      = os.Getenv("DYNAMICGO_GEN_TESTDATA") != ""
	debugAsyncGC = os.Getenv("SONIC_NO_ASYNC_GC") == ""
)

func TestMain(m *testing.M) {
	go func() {
		if !debugAsyncGC {
			return
		}
		println("Begin GC looping...")
		for {
			runtime.GC()
			debug.FreeOSMemory()
		}
	}()
	time.Sleep(time.Millisecond)
	m.Run()
}

// const ExampleReqJSON = "{\"Msg\":\"中文\",\"Base\":{\"LogID\":\"a\",\"Caller\":\"b\",\"Addr\":\"c\",\"Client\":\"d\"},\"Subfix\":-0.000000000001}"

// const ExampleReqJSON2 = `{"A":null,"Msg":"hello","InnerBase":{"Bool":true,"Byte":127,"Int16":-32768,"Int32":2147483647,"Int64":-9223372036854775808,"Double":1.7976931348623157e308,"String":"你好","ListInt32":[-1,0,1],"MapStringString":{"c":"C","a":"A","b":"B"},"SetInt32":[-1,0,1],"Foo":1,"Base":{"LogID":"log_id_inner","Caller":"","Addr":"","Client":""}},"Base":{"LogID":"log_id","Caller":"","Addr":"","Client":""}}`

// func TestThriftEncoderExample2(t *testing.T) {
// 	svc, err := thrift.NewDescritorFromPath("../testdata/idl/example2.thrift")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	req := svc.Functions()["ExampleMethod"].Request().Struct().FieldByKey("req").Type()

// 	t.Run("load", func(t *testing.T) {
// 		// stru: converted from j2t
// 		root, err := json.NewSearcher(ExampleReqJSON2).GetByPath()
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		err = root.LoadAll()
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		enc := &json.ThriftEncoder{Proto: meta.ThriftBinary, Options: json.NewDefaultOptions()}
// 		out, err := enc.Encode(req, root)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		spew.Dump(out)
// 		stru := example2.NewExampleReq()
// 		ret, err := stru.FastRead(out)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		println(ret)

// 		// exp: unmarshaled from json
// 		exp := example2.NewExampleReq()
// 		if err := sonic.UnmarshalString(ExampleReqJSON2, exp); err != nil {
// 			t.Fatal(err)
// 		}
// 		assert.Equal(t, exp, stru)

// 		b := make([]byte, exp.BLength())
// 		if ret := exp.FastWriteNocopy(b, nil); ret < 0 {
// 			t.Fatal(ret)
// 		}
// 		// spew.Dump(b)
// 		// ioutil.WriteFile("./data/example2.bin", b, 0644)
// 		assert.Equal(t, len(b), len(out))
// 		// assert.Equal(t, b, out)
// 	})

// 	t.Run("raw", func(t *testing.T) {
// 		// stru: converted from j2t
// 		root, err := json.NewSearcher(ExampleReqJSON2).GetByPath()
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		enc := &json.ThriftEncoder{Proto: meta.ThriftBinary, Options: json.NewDefaultOptions()}
// 		out, err := enc.Encode(req, root)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		spew.Dump(out)
// 		stru := example2.NewExampleReq()
// 		ret, err := stru.FastRead(out)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		println(ret)

// 		// exp: unmarshaled from json
// 		exp := example2.NewExampleReq()
// 		if err := sonic.UnmarshalString(ExampleReqJSON2, exp); err != nil {
// 			t.Fatal(err)
// 		}
// 		//FIXME: since native.j2t_fsm_exec() does not support handling null optional field, we set it to nil manually
// 		stru.SetA(nil)
// 		assert.Equal(t, exp, stru)

// 		b := make([]byte, exp.BLength())
// 		if ret := exp.FastWriteNocopy(b, nil); ret < 0 {
// 			t.Fatal(ret)
// 		}
// 		// spew.Dump(b)
// 		// assert.Equal(t, len(b), len(out))
// 		// assert.Equal(t, b, out)
// 	})
// }

// func TestThriftEncoderExample0(t *testing.T) {
// 	svc, err := thrift.NewDescritorFromPath("../testdata/idl/example.thrift")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	req := svc.Functions()["ExampleMethod"].Request().Struct().FieldByKey("req").Type()

// 	root, err := json.NewSearcher(ExampleReqJSON).GetByPath()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = root.LoadAll()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	enc := &json.ThriftEncoder{Proto: meta.ThriftBinary, Options: json.NewDefaultOptions()}
// 	out, err := enc.Encode(req, root)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	spew.Dump(out)

// 	stru := example.NewExampleReq()
// 	_, err = stru.FastRead(out)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	struc := example.NewExampleReq()
// 	if err := sonic.UnmarshalString(ExampleReqJSON, struc); err != nil {
// 		t.Fatal(err)
// 	}
// 	assert.Equal(t, struc, stru)

// 	b := make([]byte, struc.BLength())
// 	if ret := struc.FastWriteNocopy(b, nil); ret < 0 {
// 		t.Fatal(ret)
// 	}
// 	spew.Dump(b)
// 	assert.Equal(t, len(b), len(out))

// 	enc.WriteDefault = false
// 	out, err = enc.Encode(req, root)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	spew.Dump(out)
// 	stru = example.NewExampleReq()
// 	_, err = stru.FastRead(out)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

func TestGenExample(t *testing.T) {
	obj := example2.NewExampleReq()
	msg := "中文"
	obj.Msg = &msg
	obj.Subfix = -0.000000000001
	obj.Base = base.NewBase()
	obj.Base.LogID = "a"
	obj.Base.Caller = "b"
	obj.Base.Addr = "c"
	obj.Base.Client = "d"
	out := make([]byte, obj.BLength())
	ret := obj.FastWriteNocopy(out, nil)
	if ret <= 0 {
		t.Fatal(ret)
	}
	if !genData {
		return
	}
	require.Nil(t, os.WriteFile("./data/example.bin", out, 0644))
	out, err := ejson.Marshal(obj)
	require.Nil(t, err)
	require.Nil(t, os.WriteFile("./data/example.json", out, 0644))
}

func TestGenExample2(t *testing.T) {
	obj := sample.GetExample3Req()
	out := make([]byte, obj.BLength())
	ret := obj.FastWriteNocopy(out, nil)
	if ret <= 0 {
		t.Fatal(ret)
	}
	if !genData {
		return
	}
	require.Nil(t, os.WriteFile("./data/example2.bin", out, 0644))
	// out, err := ejson.Marshal(obj)
	// require.Nil(t, err)
	// require.Nil(t, os.WriteFile("./data/example2.json", out, 0644))
}

func TestGenExample2Super(t *testing.T) {
	obj := sample.GetExample2ReqSuper()
	out := make([]byte, obj.BLength())
	ret := obj.FastWriteNocopy(out, nil)
	if ret <= 0 {
		t.Fatal(ret)
	}
	if !genData {
		return
	}
	require.Nil(t, os.WriteFile("./data/example2super.bin", out, 0644))
	// out, err := ejson.Marshal(obj)
	// require.Nil(t, err)
	// require.Nil(t, os.WriteFile("./data/example2super.json", out, 0644))
}

func TestGenExample3Req(t *testing.T) {
	obj := sample.GetExample3Req()
	out := make([]byte, obj.BLength())
	ret := obj.FastWriteNocopy(out, nil)
	if ret <= 0 {
		t.Fatal(ret)
	}
	if !genData {
		return
	}
	require.Nil(t, os.WriteFile("./data/example3req.bin", out, 0644))
	out, err := jsoniter.Marshal(obj)
	require.NoError(t, err)
	require.Nil(t, os.WriteFile("./data/example3req.json", out, 0644))
}

func TestGenExample3Resp(t *testing.T) {
	obj := sample.GetExample3Resp()
	out := make([]byte, obj.BLength())
	ret := obj.FastWriteNocopy(out, nil)
	if ret <= 0 {
		t.Fatal(ret)
	}
	if !genData {
		return
	}
	require.Nil(t, os.WriteFile("./data/example3resp.bin", out, 0644))
	out, err := ejson.Marshal(obj)
	require.Nil(t, err)
	require.Nil(t, os.WriteFile("./data/example3resp.json", out, 0644))
}
