//go:build amd64

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
	"bytes"
	"context"
	"encoding/json"
	stdh "net/http"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	sjson "github.com/bytedance/sonic/ast"
	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/internal/native"
	"github.com/cloudwego/dynamicgo/internal/native/types"
	"github.com/cloudwego/dynamicgo/internal/rt"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example3"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	desc := getExampleDesc()

	t.Run("ERR_NULL_REQUIRED", func(t *testing.T) {
		desc := getNullDesc()
		data := getNullErrData()
		cv := NewBinaryConv(conv.Options{})
		ctx := context.Background()
		_, err := cv.Do(ctx, desc, data)
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrMissRequiredField, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "missing required field 3"))
		cv2 := NewBinaryConv(conv.Options{
			WriteRequireField: true,
		})
		_, err2 := cv2.Do(context.Background(), desc, data)
		require.NoError(t, err2)
		// require.True(t, strings.Contains(msg, "near "+strconv.Itoa(strings.Index(string(data), `"Null3": null`)+13)))
	})

	t.Run("INVALID_CHAR", func(t *testing.T) {
		data := `{xx}`
		cv := NewBinaryConv(conv.Options{})
		ctx := context.Background()
		_, err := cv.Do(ctx, desc, []byte(data))
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrRead, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "invalid char 'x' for state J2T_OBJ_0"))
		// require.True(t, strings.Contains(msg, "near 2"))
	})

	t.Run("ERR_INVALID_NUMBER_FMT", func(t *testing.T) {
		desc := getExampleInt2FloatDesc()
		data := []byte(`{"Float64":1.x1}`)
		cv := NewBinaryConv(conv.Options{
			EnableValueMapping: true,
		})
		ctx := context.Background()
		_, err := cv.Do(ctx, desc, data)
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrConvert, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "unexpected number type"))
		// require.True(t, strings.Contains(msg, "near 41"))
	})

	t.Run("ERR_UNSUPPORT_THRIFT_TYPE", func(t *testing.T) {
		desc := getErrorExampleDesc()
		data := []byte(`{"MapInnerBaseInnerBase":{"a":"a"}`)
		cv := NewBinaryConv(conv.Options{})
		ctx := context.Background()
		_, err := cv.Do(ctx, desc, data)
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrUnsupportedType, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "unsupported thrift type STRUCT"))
		// require.True(t, strings.Contains(msg, "near 32"))
	})

	t.Run("ERR_DISMATCH_TYPE", func(t *testing.T) {
		data := getExampleData()
		n, err := sjson.NewSearcher(string(data)).GetByPath()
		require.Nil(t, err)
		exist, err := n.Set("code_code", sjson.NewString("1.1"))
		require.True(t, exist)
		require.Nil(t, err)
		data, err = n.MarshalJSON()
		require.Nil(t, err)
		cv := NewBinaryConv(conv.Options{})
		ctx := context.Background()
		_, err = cv.Do(ctx, desc, data)
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrDismatchType, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "expect type I64 but got type 11"))
	})

	t.Run("ERR_UNKNOWN_FIELD", func(t *testing.T) {
		desc := getErrorExampleDesc()
		data := []byte(`{"UnknownField":"1"}`)
		cv := NewBinaryConv(conv.Options{
			DisallowUnknownField: true,
		})
		ctx := context.Background()
		_, err := cv.Do(ctx, desc, data)
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrUnknownField, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "unknown field 'UnknownField'"))
	})

	t.Run("ERR_UNKNOWN_FIELD", func(t *testing.T) {
		desc := getErrorExampleDesc()
		data := []byte(`{"UnknownField":"1"}`)
		cv := NewBinaryConv(conv.Options{
			DisallowUnknownField: true,
		})
		ctx := context.Background()
		_, err := cv.Do(ctx, desc, data)
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrUnknownField, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "unknown field 'UnknownField'"))
	})

	t.Run("ERR_DECODE_BASE64", func(t *testing.T) {
		desc := getErrorExampleDesc()
		data := []byte(`{"Base64":"xxx"}`)
		cv := NewBinaryConv(conv.Options{})
		ctx := context.Background()
		_, err := cv.Do(ctx, desc, data)
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrRead, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "decode base64 error: "))
	})

	t.Run("ERR_RECURSE_EXCEED_MAX", func(t *testing.T) {
		desc := getExampleInt2FloatDesc()
		src := []byte(`{}`)
		cv := NewBinaryConv(conv.Options{})
		ctx := context.Background()
		buf := make([]byte, 0, 1)
		mock := MockConv{
			sp:        types.MAX_RECURSE + 1,
			reqsCache: 1,
			keyCache:  1,
		}
		err := mock.do(&cv, ctx, src, desc, &buf, nil, true)
		require.Error(t, err)
		msg := err.Error()
		require.Equal(t, meta.ErrStackOverflow, err.(meta.Error).Code.Behavior())
		require.True(t, strings.Contains(msg, "stack "+strconv.Itoa(types.MAX_RECURSE+1)+" overflow"))
	})
}

type MockConv struct {
	sp         int
	reqsCache  int
	keyCache   int
	fieldCache int
	panic      int
}

func (mock *MockConv) Do(self *BinaryConv, ctx context.Context, desc *thrift.TypeDescriptor, jbytes []byte) (tbytes []byte, err error) {
	buf := conv.NewBytes()

	var req http.RequestGetter
	if self.opts.EnableHttpMapping {
		reqi := ctx.Value(conv.CtxKeyHTTPRequest)
		if reqi != nil {
			reqi, ok := reqi.(http.RequestGetter)
			if !ok {
				return nil, newError(meta.ErrInvalidParam, "invalid http.RequestGetter", nil)
			}
			req = reqi
		} else {
			return nil, newError(meta.ErrInvalidParam, "EnableHttpMapping but no http response in context", nil)
		}
	}

	err = mock.do(self, ctx, jbytes, desc, buf, req, true)

	if err == nil && len(*buf) > 0 {
		tbytes = make([]byte, len(*buf))
		copy(tbytes, *buf)
	}

	conv.FreeBytes(buf)
	return
}

func (mock MockConv) do(self *BinaryConv, ctx context.Context, src []byte, desc *thrift.TypeDescriptor, buf *[]byte, req http.RequestGetter, top bool) (err error) {
	flags := toFlags(self.opts)
	jp := rt.Mem2Str(src)
	// tmp := make([]byte, 0, mock.dcap)
	fsm := types.NewJ2TStateMachine()
	if mock.reqsCache != 0 {
		fsm.ReqsCache = make([]byte, 0, mock.reqsCache)
	}
	if mock.keyCache != 0 {
		fsm.KeyCache = make([]byte, 0, mock.keyCache)
	}
	if mock.fieldCache != 0 {
		fsm.FieldCache = make([]int32, 0, mock.fieldCache)
	}
	// fsm := &types.J2TStateMachine{
	// 	SP: mock.sp,
	// 	JT: types.JsonState{
	// 		Dbuf: *(**byte)(unsafe.Pointer(&tmp)),
	// 		Dcap: mock.dcap,
	// 	},
	// 	ReqsCache:  make([]byte, 0, mock.reqsCache),
	// 	KeyCache:   make([]byte, 0, mock.keyCache),
	// 	SM:         types.StateMachine{},
	// 	VT:         [types.MAX_RECURSE]types.J2TState{},
	// 	FieldCache: make([]int32, 0, mock.fc),
	// }
	fsm.Init(0, unsafe.Pointer(desc))
	if mock.sp != 0 {
		fsm.SP = mock.sp
	}
	var ret uint64
	defer func() {
		if msg := recover(); msg != nil {
			panic(makePanicMsg(msg, src, desc, buf, req, self.flags, fsm, ret))
		}
	}()

exec:
	mock.panic -= 1
	if mock.panic == 0 {
		panic("test!")
	}
	ret = native.J2T_FSM(fsm, buf, &jp, flags)
	if ret != 0 {
		cont, e := self.handleError(ctx, fsm, buf, src, req, ret, top)
		if cont && e == nil {
			goto exec
		}
		err = e
		goto ret
	}

ret:
	types.FreeJ2TStateMachine(fsm)
	runtime.KeepAlive(desc)
	return
}

func TestStateMachineOOM(t *testing.T) {
	desc := getExampleInt2FloatDesc()
	src := []byte(`{"\u4e2d\u6587":"\u4e2d\u6587", "Int32":222}`)
	println(((uint8)([]byte(`中文`)[5])))
	f := desc.Struct().FieldByKey("中文")
	require.NotNil(t, f)
	cv := NewBinaryConv(conv.Options{})
	ctx := context.Background()
	buf := make([]byte, 0, 1)
	mock := MockConv{
		sp:        0,
		reqsCache: 1,
		keyCache:  1,
	}
	err := mock.do(&cv, ctx, src, desc, &buf, nil, true)
	require.Nil(t, err)
	exp := example3.NewExampleInt2Float()
	err = json.Unmarshal(src, exp)
	require.NoError(t, err)
	act := example3.NewExampleInt2Float()
	_, err = act.FastRead(buf)
	require.Nil(t, err)
	require.Equal(t, exp, act)

	t.Run("field cache OOM", func(t *testing.T) {
		desc := getExampleDesc()
		data := []byte(`{}`)
		hr, err := stdh.NewRequest("POST", "http://localhost:8888/root?Msg=a&Subfix=1", bytes.NewBuffer(nil))
		hr.Header.Set("Base", `{"LogID":"c"}`)
		hr.Header.Set("Path", `b`)
		// NOTICE: optional field will be ignored
		hr.Header.Set("Extra", `{"x":"y"}`)
		require.NoError(t, err)
		req := &http.HTTPRequest{
			Request: hr,
			BodyMap: map[string]string{
				"InnerBase": `{"Bool":true}`,
			},
		}
		edata := []byte(`{"Base":{"LogID":"c"},"Subfix":1,"Path":"b","InnerBase":{"Bool":true}}`)
		exp := example3.NewExampleReq()
		exp.RawUri = req.GetUri()
		err = json.Unmarshal(edata, exp)
		require.Nil(t, err)
		exp.RawUri = req.GetUri()
		mock := MockConv{
			sp:         1,
			reqsCache:  1,
			keyCache:   1,
			fieldCache: 0,
		}
		cv := NewBinaryConv(conv.Options{
			EnableHttpMapping:            true,
			WriteRequireField:            true,
			ReadHttpValueFallback:        true,
			TracebackRequredOrRootFields: true,
		})
		ctx := context.Background()
		ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
		out, err := mock.Do(&cv, ctx, desc, data)
		require.NoError(t, err)
		act := example3.NewExampleReq()
		_, err = act.FastRead(out)
		require.Nil(t, err)
		require.Equal(t, exp, act)
	})
}

func TestPanicRecover(t *testing.T) {
	desc := getExampleDesc()
	data := getExampleData()
	exp := example3.NewExampleReq()
	err := json.Unmarshal(data, exp)
	require.Nil(t, err)
	req := getExampleReq(exp, true, data)
	cv := NewBinaryConv(conv.Options{
		EnableHttpMapping: true,
	})
	ctx := context.Background()
	ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
	buf := make([]byte, 0, 1)
	mock := MockConv{
		panic: 2,
	}
	defer func() {
		if v := recover(); v == nil {
			t.Fatal("not panic")
		} else {
			t.Log(v)
		}
	}()
	_ = mock.do(&cv, ctx, data, desc, &buf, req, true)
}
