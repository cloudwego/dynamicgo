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

package testdata

import (
	"context"
	"testing"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/conv/j2t"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/util_test"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	thrift_iter "github.com/thrift-iterator/go"
	"github.com/thrift-iterator/go/general"
	"github.com/thrift-iterator/go/protocol"
)

func getFnDescByPathName(t *testing.T, filePath, fnName string) *thrift.FunctionDescriptor {
	svc, err := thrift.NewDescritorFromPath(context.Background(), util_test.MustGitPath(filePath))
	if err != nil {
		t.Fatal(err)
	}
	fn, err := svc.LookupFunctionByMethod(fnName)
	if err != nil {
		t.Fatal(err)
	}
	return fn
}

func TestHTTP2ThriftConv_Do(t *testing.T) {
	t.Parallel()
	getFavoriteFn := getFnDescByPathName(t, "testdata/idl/example4.thrift", "GetFavoriteMethod")

	tests := []struct {
		name    string
		proto   meta.Encoding
		fn      *thrift.FunctionDescriptor
		req     http.RequestGetter
		want    general.Message
		wantErr bool
	}{
		{
			name:  "exmple4",
			proto: meta.EncodingThriftBinary,
			fn:    getFavoriteFn,
			req: util_test.Req{
				BodyArr:  []byte(`{"Id":1234,"Base":{"Uid":123,"Stuffs":[]}}`),
				QueryMap: map[string]string{"id": "7749"},
			},
			want: general.Message{
				MessageHeader: protocol.MessageHeader{
					MessageName: "GetFavoriteMethod",
					MessageType: protocol.MessageTypeCall,
				},
				Arguments: general.Struct{
					1: general.Struct{
						1: int32(7749),
						255: general.Struct{
							1: int32(123),
							2: general.List(nil),
						},
					},
				},
			},
		},
		{
			name:  "exmple4empty",
			proto: meta.EncodingThriftBinary,
			fn:    getFavoriteFn,
			req: util_test.Req{
				BodyArr:  []byte(``),
				QueryMap: map[string]string{"id": "7749"},
			},
			want: general.Message{
				MessageHeader: protocol.MessageHeader{
					MessageName: "GetFavoriteMethod",
					MessageType: protocol.MessageTypeCall,
				},
				Arguments: general.Struct{
					1: general.Struct{
						1: int32(7749),
						255: general.Struct{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			convIns := j2t.NewHTTPConv(tt.proto, tt.fn)
			gotTbytes := make([]byte, 0, 1)
			err := convIns.DoInto(context.Background(), tt.req, &gotTbytes, conv.Options{
				WriteRequireField:            true,
				ReadHttpValueFallback:        true,
				EnableHttpMapping:            true,
			})
			spew.Dump(gotTbytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPConv.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := general.Message{}
			err = thrift_iter.Unmarshal(gotTbytes, &got)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
