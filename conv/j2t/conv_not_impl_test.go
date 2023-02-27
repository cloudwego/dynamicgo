//go:build !amd64
// +build !amd64

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
	"context"
	"testing"

	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/stretchr/testify/require"
)

func TestConvJSON2Thrift(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()
	cv := NewBinaryConv(conv.Options{})
	ctx := context.Background()
	out, err := cv.Do(ctx, desc, data)
	require.Equal(t, meta.ErrorNotSupport, err)
	require.Nil(t, out)
}

func (mock MockConv) do(self *BinaryConv, ctx context.Context, src []byte, desc *thrift.TypeDescriptor, buf *[]byte, req http.RequestGetter, top bool) (err error) {
	return meta.ErrorNotSupport
}
