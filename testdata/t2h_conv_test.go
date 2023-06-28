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
	"github.com/cloudwego/dynamicgo/conv/t2j"
	"github.com/cloudwego/dynamicgo/internal/util_test"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/stretchr/testify/assert"
)

func TestThrift2HTTPConv_Do(t *testing.T) {
	svc, err := thrift.NewDescritorFromPath(context.Background(), util_test.MustGitPath("testdata/idl/example4.thrift"))
	if err != nil {
		t.Fatal(err)
	}
	fn, err := svc.LookupFunctionByMethod("GetFavoriteMethod")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		conv    *t2j.HTTPConv
		tbytes  []byte
		wantErr bool
		want    util_test.Resp
	}{
		{
			name: "example4",
			conv: t2j.NewHTTPConv(meta.EncodingThriftBinary, fn),
			tbytes: util_test.MustDecodeHex(
				t,
				"80010002000000114765744661766f726974654d6574686f64000000130c00000c000108000100001e450f00020b00000002"+
					"0000000568656c6c6f00000005776f726c64000c00ff080002000000000b000100000000000000",
			),
			want: util_test.Resp{
				Body: []byte(`{"favorite":{"Uid":7749,"Stuffs":["hello","world"]},` +
					`"BaseResp":{"StatusCode":0,"StatusMessage":""}}`),
				StatusCode: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &util_test.Resp{}
			if err := tt.conv.Do(context.Background(), resp, tt.tbytes, conv.Options{}); (err != nil) != tt.wantErr {
				t.Errorf("HTTPConv.Do() error = %v, wantErr %v", err, tt.wantErr)
			}
			util_test.JsonEqual(t, tt.want.Body, resp.Body)
			assert.Equal(t, tt.want.StatusCode, resp.StatusCode)
			assert.Equal(t, tt.want.Header, resp.Header)
			assert.Equal(t, tt.want.Cookie, resp.Cookie)

			resp = &util_test.Resp{}
			buf := make([]byte, 0, 1)
			if err := tt.conv.DoInto(context.Background(), resp, tt.tbytes, &buf, conv.Options{}); (err != nil) != tt.wantErr {
				t.Errorf("HTTPConv.Do() error = %v, wantErr %v", err, tt.wantErr)
			}
			util_test.JsonEqual(t, tt.want.Body, resp.Body)
			assert.Equal(t, tt.want.StatusCode, resp.StatusCode)
			assert.Equal(t, tt.want.Header, resp.Header)
			assert.Equal(t, tt.want.Cookie, resp.Cookie)
		})
	}
}
