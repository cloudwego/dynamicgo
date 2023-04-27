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

import "github.com/cloudwego/dynamicgo/meta"

func UnwrapMessage(proto meta.Encoding, buf []byte) (name string, callType TMessageType, seqID int32, structID FieldID, body []byte, err error) {
	switch proto {
	case meta.EncodingThriftBinary:
		return UnwrapBinaryMessage(proto, buf)
	case meta.EncodingThriftCompact:
		return UnwrapCompactMessage(proto, buf)
	default:
		err = errNotImplemented
		return
	}
}
