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

package thrift

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"runtime/debug"
	"testing"
	"time"
)

var (
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

func getExampleDesc() *TypeDescriptor {
	svc, err := NewDescritorFromPath(context.Background(), "../testdata/idl/example.thrift")
	if err != nil {
		panic(err)
	}
	return svc.Functions()["ExampleMethod"].Request().Struct().FieldByKey("req").Type()
}

func getExampleData() []byte {
	out, err := ioutil.ReadFile("../testdata/data/example.bin")
	if err != nil {
		panic(err)
	}
	return out
}

func TestBinaryProtocol_ReadAnyWithDesc(t *testing.T) {
	p1, err := NewDescritorFromPath(context.Background(), "../testdata/idl/example3.thrift")
	if err != nil {
		panic(err)
	}
	exp3partial := p1.Functions()["PartialMethod"].Response().Struct().FieldById(0).Type()
	data, err := ioutil.ReadFile("../testdata/data/example3.bin")
	if err != nil {
		panic(err)
	}

	p := NewBinaryProtocol(data)
	v, err := p.ReadAnyWithDesc(exp3partial, false, false, false, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", v)
	p = NewBinaryProtocolBuffer()
	err = p.WriteAnyWithDesc(exp3partial, v, true, true, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%x", p.RawBuf())
	v, err = p.ReadAnyWithDesc(exp3partial, false, false, false, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", v)
}
