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
	"io/ioutil"
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
	require.Nil(t, ioutil.WriteFile("./data/example.bin", out, 0644))
	out, err := ejson.Marshal(obj)
	require.Nil(t, err)
	require.Nil(t, ioutil.WriteFile("./data/example.json", out, 0644))
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
	require.Nil(t, ioutil.WriteFile("./data/example2.bin", out, 0644))
	// out, err := ejson.Marshal(obj)
	// require.Nil(t, err)
	// require.Nil(t, ioutil.WriteFile("./data/example2.json", out, 0644))
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
	require.Nil(t, ioutil.WriteFile("./data/example2super.bin", out, 0644))
	// out, err := ejson.Marshal(obj)
	// require.Nil(t, err)
	// require.Nil(t, ioutil.WriteFile("./data/example2super.json", out, 0644))
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
	require.Nil(t, ioutil.WriteFile("./data/example3req.bin", out, 0644))
	out, err := jsoniter.Marshal(obj)
	require.NoError(t, err)
	require.Nil(t, ioutil.WriteFile("./data/example3req.json", out, 0644))
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
	require.Nil(t, ioutil.WriteFile("./data/example3resp.bin", out, 0644))
	out, err := ejson.Marshal(obj)
	require.Nil(t, err)
	require.Nil(t, ioutil.WriteFile("./data/example3resp.json", out, 0644))
}
