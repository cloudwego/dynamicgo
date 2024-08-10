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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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

func TestBinaryProtocol_WriteAny_ReadAny(t *testing.T) {
	type args struct {
		val         interface{}
		sliceAsSet  bool
		strAsBinary bool
		byteAsInt8  bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    interface{}
	}{
		{"bool", args{true, false, false, false}, false, true},
		{"byte", args{byte(1), false, false, false}, false, byte(1)},
		{"byte", args{byte(1), false, false, true}, false, int8(1)},
		{"i16", args{int16(1), false, false, false}, false, int16(1)},
		{"i32", args{int32(1), false, false, false}, false, int32(1)},
		{"i64", args{int64(1), false, false, false}, false, int64(1)},
		{"int", args{1, false, false, false}, false, int64(1)},
		{"f32", args{float32(1.0), false, false, false}, false, float64(1.0)},
		{"f64", args{1.0, false, false, false}, false, float64(1.0)},
		{"string", args{"1", false, false, false}, false, "1"},
		{"string2binary", args{"1", false, true, false}, false, []byte{'1'}},
		{"binary2string", args{[]byte{1}, false, false, false}, false, string("\x01")},
		{"binary2binary", args{[]byte{1}, false, true, false}, false, []byte{1}},
		{"list", args{[]interface{}{int32(1)}, false, false, false}, false, []interface{}{int32(1)}},
		{"set", args{[]interface{}{int64(1)}, true, false, false}, false, []interface{}{int64(1)}},
		{"int map", args{map[int]interface{}{1: byte(1)}, false, false, false}, false, map[int]interface{}{1: byte(1)}},
		{"int map error", args{map[int64]interface{}{1: byte(1)}, false, false, true}, false, map[int]interface{}{1: int8(1)}},
		{"int map empty", args{map[int64]interface{}{}, false, false, false}, true, nil},
		{"string map", args{map[string]interface{}{"1": "1"}, false, false, true}, false, map[string]interface{}{"1": "1"}},
		{"string map error", args{map[string]interface{}{"1": []int{1}}, false, false, true}, true, nil},
		{"string map empty", args{map[string]interface{}{}, false, false, true}, true, nil},
		{"any map", args{map[interface{}]interface{}{1.1: "1"}, false, false, false}, false, map[interface{}]interface{}{1.1: "1"}},
		{"any map + key list", args{map[interface{}]interface{}{&[]interface{}{1}: "1"}, false, false, false}, false, map[interface{}]interface{}{&[]interface{}{int64(1)}: "1"}},
		{"any map + val list", args{map[interface{}]interface{}{1.1: []interface{}{"1"}}, false, true, false}, false, map[interface{}]interface{}{1.1: []interface{}{[]byte{'1'}}}},
		{"any map empty", args{map[interface{}]interface{}{}, false, false, false}, true, nil},
		{"struct", args{map[FieldID]interface{}{FieldID(1): 1.1}, false, false, false}, false, map[FieldID]interface{}{FieldID(1): 1.1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			println("case:", tt.name)
			p := &BinaryProtocol{}
			typ, err := p.WriteAny(tt.args.val, tt.args.sliceAsSet)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BinaryProtocol.WriteAny() error = %v, wantErr %v", err, tt.wantErr)
			}
			fmt.Printf("buf:%+v\n", p.RawBuf())
			got, err := p.ReadAny(typ, tt.args.strAsBinary, tt.args.byteAsInt8)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BinaryProtocol.ReadAny() error = %v, wantErr %v", err, tt.wantErr)
			}
			fmt.Printf("got:%#v\n", got)
			if strings.Contains(tt.name, "any map + key") {
				em := tt.want.(map[interface{}]interface{})
				gm := got.(map[interface{}]interface{})
				require.Equal(t, len(em), len(gm))
				var firstK, firstV interface{}
				for k, v := range em {
					firstK, firstV = k, v
					break
				}
				var firstKgot, firstVgot interface{}
				for k, v := range gm {
					firstKgot, firstVgot = k, v
					break
				}
				require.Equal(t, firstK, firstKgot)
				require.Equal(t, firstV, firstVgot)
			} else {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func TestBinaryUnwrap(t *testing.T) {
	b := BinaryProtocol{
		Buf:  nil,
		Read: 0,
	}
	var i32 int32 = -2147418110
	var methodName = "DummyNew"
	b.WriteI32(i32)
	b.WriteString(methodName)
	b.WriteI32(1)
	b.WriteByte(byte(STOP))

	name, rType, _, _, bs, err := UnwrapBinaryMessage(b.Buf)

	require.Equal(t, bs, []byte{})
	require.Equal(t, name, methodName)
	require.Equal(t, rType, REPLY)
	var nilErr error
	require.Equal(t, err, nilErr)

}
