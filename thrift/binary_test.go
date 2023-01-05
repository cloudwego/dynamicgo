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

package thrift

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"testing"
	"time"

	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/skip"
	"github.com/stretchr/testify/require"
	thrift_iter "github.com/thrift-iterator/go"
	"github.com/thrift-iterator/go/general"
	"github.com/thrift-iterator/go/protocol"
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
	svc, err := NewDescritorFromPath("../testdata/idl/example.thrift")
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

func BenchmarkSkipNoCheck(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()

	b.Run("native", func(b *testing.B) {
		p := NewBinaryProtocol(data)
		err := p.SkipNative(desc.Type(), 512)
		require.Nil(b, err)
		require.Equal(b, len(data), p.Read)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.Read = 0
			_ = p.SkipNative(desc.Type(), 512)
		}
	})

	b.Run("go", func(b *testing.B) {
		p := NewBinaryProtocol(data)
		err := p.SkipGo(desc.Type(), 512)
		require.Nil(b, err)
		require.Equal(b, len(data), p.Read)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.Read = 0
			_ = p.SkipGo(desc.Type(), 512)
		}
	})
}

func BenchmarkSkipAndCheck(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()

	b.Run("native", func(b *testing.B) {
		p := NewBinaryProtocol(data)
		err := p.SkipNative(desc.Type(), 512)
		require.Nil(b, err)
		require.Equal(b, len(data), p.Read)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.Read = 0
			err := p.SkipNative(desc.Type(), 512)
			require.Nil(b, err)
			require.Equal(b, len(data), p.Read)
		}
	})

	b.Run("go", func(b *testing.B) {
		p := NewBinaryProtocol(data)
		err := p.SkipGo(desc.Type(), 512)
		require.Nil(b, err)
		require.Equal(b, len(data), p.Read)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.Read = 0
			err := p.SkipGo(desc.Type(), 512)
			require.Nil(b, err)
			require.Equal(b, len(data), p.Read)
		}
	})
}

func TestBinaryUnwrapReply(t *testing.T) {

	type args struct {
		proto meta.Encoding
		msg   general.Message
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "simple",
			args: args{
				proto: meta.EncodingThriftBinary,
				msg: general.Message{
					MessageHeader: protocol.MessageHeader{
						MessageName: "GetFavoriteMethod",
						MessageType: protocol.MessageTypeReply,
					},
					Arguments: general.Struct{
						1: general.Struct{
							1: int32(7749),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// pair of buffer
			inputBytes, err := thrift_iter.Marshal(tt.args.msg)
			if err != nil {
				t.Fatalf("failed to marshal input %s", err)
			}
			wantBytes, err := thrift_iter.Marshal(tt.args.msg.Arguments[1].(general.Struct))
			if err != nil {
				t.Fatalf("failed to marshal want bytes %s", err)
			}
			//
			_, _, _, _, got, err := UnwrapBinaryMessage(tt.args.proto, inputBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnwrapReply() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, wantBytes) {
				t.Errorf("UnwrapReply() = %s, want %s", hex.EncodeToString(got), hex.EncodeToString(wantBytes))
			}
			header, footer, err := GetBinaryMessageHeaderAndFooter(tt.args.msg.MessageHeader.MessageName, TMessageType(tt.args.msg.MessageHeader.MessageType), 1, 0)
			if err != nil {
				t.Fatalf("failed to wrap message %s", err)
			}
			header = append(header, got...)
			header = append(header, footer...)
			if !bytes.Equal(header, inputBytes) {
				t.Errorf("WrapBinaryMessage() = %s, want %s", hex.EncodeToString(header), hex.EncodeToString(inputBytes))
			}

		})
	}
}

func TestSkip(t *testing.T) {
	var MAX_STACKS = 1000
	r := require.New(t)

	obj := skip.NewTestListMap()
	obj.ListMap = make([]map[string]string, 10)
	for i := 0; i < 10; i++ {
		obj.ListMap[i] = map[string]string{strconv.Itoa(i): strconv.Itoa(i)}
	}
	data := make([]byte, obj.BLength())
	_ = obj.FastWriteNocopy(data, nil)

	p := BinaryProtocol{
		Buf: data,
	}
	println("Skip Go")
	e1 := p.SkipGo(STRUCT, MAX_STACKS)
	r.NoError(e1)
	r.Equal(len(data), p.Read)

	p.Read = 0
	println("Skip Native")
	e2 := p.SkipNative(STRUCT, MAX_STACKS)
	r.NoError(e2)
	r.Equal(len(data), p.Read)
}
