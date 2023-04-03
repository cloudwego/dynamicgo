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
	"encoding/json"
	"testing"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example3"
	"github.com/cloudwego/dynamicgo/testdata/sample"
	"github.com/stretchr/testify/require"
)

func TestCompactWriteDefault(t *testing.T) {
	desc := getExampleDesc()
	data := []byte(`{"Path":"<>"}`)
	exp := example3.NewExampleReq()
	data2 := []byte(`{"Path":"<>","Base":{}}`)
	exp.InnerBase = sample.GetEmptyInnerBase3()
	err := json.Unmarshal(data2, exp)
	require.Nil(t, err)
	req := getExampleReq(exp, false, data)
	cv := NewBinaryConv(conv.Options{
		WriteDefaultField: true,
		EnableHttpMapping: true,
	})
	ctx := context.Background()
	ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
	out, err := cv.Do(ctx, desc, data)
	require.Nil(t, err)
	act := example3.NewExampleReq()
	_, err = act.FastRead(out)
	require.Nil(t, err)
	require.Equal(t, exp, act)
}
