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

package conv

import (
	"sync"
)

// ContextKey is the key type for context arguments
type ContextKey struct {
	_ bool
}

var (
	// CtxKeyHTTPResponse is the key for http.ResponseSetter in context
	CtxKeyHTTPResponse = &ContextKey{}
	// CtxKeyHTTPRequest is the key for http.RequestGetter in context
	CtxKeyHTTPRequest = &ContextKey{}
	// CtxKeyThriftRespBase is the key for base.Base in context
	CtxKeyThriftRespBase = &ContextKey{}
	// CtxKeyThriftReqBase is the key for base.BaseResp in context
	CtxKeyThriftReqBase = &ContextKey{}
	// CtxKeyConvOptions is the key for Options in context
	CtxKeyConvOptions = &ContextKey{}
)

var (
	// DefaultBufferSize is the default buffer size for conversion
	DefaultBufferSize = 4096
	// DefaultHttpValueBufferSize is the default buffer size for copying a json value
	DefaulHttpValueBufferSizeForJSON = 1024
	// DefaultHttpValueBufferSize is the default buffer size for copying a http value
	DefaulHttpValueBufferSizeForScalar = 64
)

type Options struct {

	// EnableValueMapping indicates if value mapping (api.js_conv...) should be enabled
	EnableValueMapping bool
	// EnableHttpMapping indicates if http mapping (api.query|api.header...) should be enabled
	EnableHttpMapping bool

	// EnableThriftBase indicates if thrift/base should be recognized and mapping to/from context
	// NOTICE: To enable this option, user must parse IDL with thrift.Options.EnableThriftBase;
	// after enable:
	//   -For j2t: user SHOULD pre-set *base.Base with key conv.CtxKeyThriftReqBase in context,
	//   and SHOULDN'T pass this field in JSON (Undocumented Behavior)
	//   -For t2j: user SHOULD pre-set *base.BaseResponse with key conv.CtxKeyThriftRespBase in context,
	//   and this field WON'T appear in JSON
	EnableThriftBase bool

	// String2Int64 indicates if string value can be converted to **Int8/Int16/Int32/Int64/Float64**,
	String2Int64 bool

	// Int642String indicates if a **Int64** field can be converted to string
	Int642String bool

	// NoBase64Binary indicates if base64 string should be Encode/Decode as []byte
	NoBase64Binary bool
	// ByteAsUint8 indicates if byte should be conv as uint8 (default is int8), this only works for t2j now
	ByteAsUint8 bool

	// WriteOptionalField indicates if optional-requireness fields should be written when not given
	WriteOptionalField bool
	// WriteDefaultField indicates if default-requireness fields should be written if
	WriteDefaultField bool
	// WriteRequireField indicates if required-requireness fields should be written empty value if
	// not found
	WriteRequireField bool
	// DisallowUnknownField indicates if unknown fields should be skipped
	DisallowUnknownField bool

	// ReadHttpValueFallback indicates if http-annotated fields should fallback to http body after reading from non-body parts (header,cookie...) failed
	ReadHttpValueFallback bool
	// WriteHttpValueFallback indicates if http-annotated fields should fallback to http body after writing to non-body parts (header,cookie...) failed
	WriteHttpValueFallback bool
	// TracebackRequredOrRootFields indicates if required-requireness
	// or root-level fields should be seeking on http-values when reading failed from current layer of json.
	// this option is only used in j2t now.
	TracebackRequredOrRootFields bool
	// OmitHttpMappingErrors indicates to omit http-mapping failing errors.
	// If there are more-than-one HTTP annotations on the field, dynamicgo will try to mapping next annotation source (from left to right) until succeed.
	OmitHttpMappingErrors bool

	// NoCopyString indicates if string-kind http values should be copied or just referenced (if possible)
	NoCopyString bool
	// UseNativeSkip indicates if using thrift.SkipNative() or thrift.SkipGo()
	UseNativeSkip bool

	// ConvertException indicates that it returns error for thrift exception fields when doing BinaryConv t2j
	ConvertException bool

	// UseKitexHttpEncoding indicating using kitex's text encoding to output complex http values
	UseKitexHttpEncoding bool
}

var bufPool = sync.Pool{
	New: func() interface{} {
		ret := make([]byte, 0, DefaultBufferSize)
		return &ret
	},
}

// NewBytes returns a new byte slice from pool
func NewBytes() *[]byte {
	return bufPool.Get().(*[]byte)
}

// FreeBytes returns a byte slice to pool and reset it
func FreeBytes(b *[]byte) {
	*b = (*b)[:0]
	bufPool.Put(b)
}
