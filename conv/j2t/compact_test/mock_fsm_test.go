package j2t

import (
	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/internal/native/types"
)

type MockConv struct {
	sp        int
	reqsCache int
	keyCache  int
	dcap      int
	fc        int
}

// func (mock *MockConv) Do(self *BinaryConv, ctx context.Context, desc *thrift.TypeDescriptor, jbytes []byte) (tbytes []byte, err error) {
// 	buf := conv.NewBytes()

// 	var req http.RequestGetter
// 	if self.opts.EnableHttpMapping {
// 		reqi := ctx.Value(conv.CtxKeyHTTPRequest)
// 		if reqi != nil {
// 			reqi, ok := reqi.(http.RequestGetter)
// 			if !ok {
// 				return nil, newError(meta.ErrInvalidParam, "invalid http.RequestGetter", nil)
// 			}
// 			req = reqi
// 		} else {
// 			return nil, newError(meta.ErrInvalidParam, "EnableHttpMapping but no http response in context", nil)
// 		}
// 	}

// 	err = mock.do(self, ctx, jbytes, desc, buf, req, true)

// 	if err == nil && len(*buf) > 0 {
// 		tbytes = make([]byte, len(*buf))
// 		copy(tbytes, *buf)
// 	}

// 	conv.FreeBytes(buf)
// 	return
// }

// func (mock MockConv) do(self *j2t.CompactConv, ctx context.Context, src []byte, desc *thrift.TypeDescriptor, buf *[]byte, req http.RequestGetter, top bool) (err error) {
// 	flags := toFlags(self.opts)
// 	jp := rt.Mem2Str(src)
// 	tmp := make([]byte, 0, mock.dcap)
// 	fsm := &types.J2TStateMachine{
// 		SP: mock.sp,
// 		JT: types.JsonState{
// 			Dbuf: *(**byte)(unsafe.Pointer(&tmp)),
// 			Dcap: mock.dcap,
// 		},
// 		ReqsCache:  make([]byte, 0, mock.reqsCache),
// 		KeyCache:   make([]byte, 0, mock.keyCache),
// 		SM:         types.StateMachine{},
// 		VT:         [types.MAX_RECURSE]types.J2TState{},
// 		FieldCache: make([]int32, 0, mock.fc),
// 	}
// 	fsm.Init(0, unsafe.Pointer(desc))
// 	if mock.sp != 0 {
// 		fsm.SP = mock.sp
// 	}

// exec:
// 	ret := native.J2T_FSM_TB(fsm, buf, &jp, flags)
// 	if ret != 0 {
// 		cont, e := self.handleError(ctx, fsm, buf, src, req, ret, top)
// 		if cont && e == nil {
// 			goto exec
// 		}
// 		err = e
// 		goto ret
// 	}

// ret:
// 	types.FreeJ2TStateMachine(fsm)
// 	runtime.KeepAlive(desc)
// 	return
// }

func toFlags(opts conv.Options) (flags uint64) {
	if opts.WriteDefaultField {
		flags |= types.F_WRITE_DEFAULT
	}
	if !opts.DisallowUnknownField {
		flags |= types.F_ALLOW_UNKNOWN
	}
	if opts.EnableValueMapping {
		flags |= types.F_VALUE_MAPPING
	}
	if opts.EnableHttpMapping {
		flags |= types.F_HTTP_MAPPING
	}
	if opts.String2Int64 {
		flags |= types.F_STRING_INT
	}
	if opts.WriteRequireField {
		flags |= types.F_WRITE_REQUIRE
	}
	if opts.NoBase64Binary {
		flags |= types.F_NO_BASE64
	}
	if opts.WriteOptionalField {
		flags |= types.F_WRITE_OPTIONAL
	}
	if opts.ReadHttpValueFallback {
		flags |= types.F_TRACE_BACK
	}
	return
}
