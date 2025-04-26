//go:build amd64 && !go1.25

/**
 * Copyright 2023 cloudwego Inc.
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
	"io"

	"github.com/cloudwego/dynamicgo/internal/native"
	"github.com/cloudwego/dynamicgo/internal/native/types"
)

// Skip skips over teh value for the given type using native C implementation.
func (p *BinaryProtocol) SkipNative(fieldType Type, maxDepth int) (err error) {
	if maxDepth >= types.TB_SKIP_STACK_SIZE {
		return errExceedDepthLimit
	}
	left := len(p.Buf) - p.Read
	if left <= 0 {
		return io.EOF
	}
	fsm := types.NewTStateMachine()
	ret := native.TBSkip(fsm, &p.Buf[p.Read], left, uint8(fieldType))
	if ret < 0 {
		return
	}
	p.Read += int(ret)
	types.FreeTStateMachine(fsm)
	return nil
}
