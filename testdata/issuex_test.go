/**
 * Copyright 2023 ByteDance Inc.
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
	"runtime"
	// "encoding/json"
	// "strconv"
	"strings"
	"sync"
	"testing"

	"github.com/cloudwego/dynamicgo/conv"
	"github.com/cloudwego/dynamicgo/conv/j2t"
	"github.com/cloudwego/dynamicgo/http"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/stretchr/testify/require"
)

type SubMessage struct {
	Id    *int64  `thrift:"id,1,optional" json:"id,omitempty"`
	Value *string `thrift:"value,2,optional" json:"value,omitempty"`
}

type Message struct {
	Id          *int64        `thrift:"id,1,optional" json:"id,omitempty"`
	Value       *string       `thrift:"value,2,optional" json:"value,omitempty"`
	SubMessages []*SubMessage `thrift:"subMessages,3,optional" json:"subMessages,omitempty"`
}

type ObjReq struct {
	Action   string                 `thrift:"action,1,required" json:"action"`
	Msg      string                 `thrift:"msg,2,required" json:"msg"`
	MsgMap   map[string]*SubMessage `thrift:"msgMap,3,required" json:"msgMap"`
	SubMsgs  []*SubMessage          `thrift:"subMsgs,4,required" json:"subMsgs"`
	MsgSet   []*Message             `thrift:"msgSet,5,optional" json:"msgSet,omitempty"`
	FlagMsg  *Message               `thrift:"flagMsg,6,required" json:"flagMsg"`
	MockCost *string                `thrift:"mockCost,7,optional" json:"mockCost,omitempty"`
}

func getReqValue(size int) string {
	// req := &ObjReq{
	// 	Action:  "",
	// 	Msg:     "",
	// 	MsgMap:  map[string]*SubMessage{},
	// 	SubMsgs: []*SubMessage{},
	// 	MsgSet:  []*Message{},
	// 	FlagMsg: &Message{},
	// }

	// for i := 0; i < size; i++ {
	// 	req.MsgMap[strconv.Itoa(i)] = getSubMessage(int64(i))
	// 	req.SubMsgs = append(req.SubMsgs, getSubMessage(int64(i)))
	// 	req.MsgSet = append(req.MsgSet, getMessage(int64(i)))
	// 	req.FlagMsg = getMessage(int64(i))
	// }

	// data, _ := json.Marshal(req)
	// return string(data)
	return `{"action":"","msg":"","msgMap":{"0":{"id":0,"value":"hello"},"1":{"id":1,"value":"hello"},"10":{"id":10,"value":"hello"},"11":{"id":11,"value":"hello"},"12":{"id":12,"value":"hello"},"13":{"id":13,"value":"hello"},"14":{"id":14,"value":"hello"},"15":{"id":15,"value":"hello"},"16":{"id":16,"value":"hello"},"17":{"id":17,"value":"hello"},"18":{"id":18,"value":"hello"},"19":{"id":19,"value":"hello"},"2":{"id":2,"value":"hello"},"20":{"id":20,"value":"hello"},"21":{"id":21,"value":"hello"},"22":{"id":22,"value":"hello"},"23":{"id":23,"value":"hello"},"24":{"id":24,"value":"hello"},"25":{"id":25,"value":"hello"},"26":{"id":26,"value":"hello"},"27":{"id":27,"value":"hello"},"28":{"id":28,"value":"hello"},"29":{"id":29,"value":"hello"},"3":{"id":3,"value":"hello"},"30":{"id":30,"value":"hello"},"31":{"id":31,"value":"hello"},"32":{"id":32,"value":"hello"},"33":{"id":33,"value":"hello"},"34":{"id":34,"value":"hello"},"35":{"id":35,"value":"hello"},"36":{"id":36,"value":"hello"},"37":{"id":37,"value":"hello"},"38":{"id":38,"value":"hello"},"39":{"id":39,"value":"hello"},"4":{"id":4,"value":"hello"},"40":{"id":40,"value":"hello"},"41":{"id":41,"value":"hello"},"42":{"id":42,"value":"hello"},"43":{"id":43,"value":"hello"},"44":{"id":44,"value":"hello"},"45":{"id":45,"value":"hello"},"46":{"id":46,"value":"hello"},"47":{"id":47,"value":"hello"},"48":{"id":48,"value":"hello"},"49":{"id":49,"value":"hello"},"5":{"id":5,"value":"hello"},"50":{"id":50,"value":"hello"},"51":{"id":51,"value":"hello"},"52":{"id":52,"value":"hello"},"53":{"id":53,"value":"hello"},"54":{"id":54,"value":"hello"},"55":{"id":55,"value":"hello"},"56":{"id":56,"value":"hello"},"57":{"id":57,"value":"hello"},"58":{"id":58,"value":"hello"},"59":{"id":59,"value":"hello"},"6":{"id":6,"value":"hello"},"60":{"id":60,"value":"hello"},"61":{"id":61,"value":"hello"},"62":{"id":62,"value":"hello"},"63":{"id":63,"value":"hello"},"64":{"id":64,"value":"hello"},"65":{"id":65,"value":"hello"},"66":{"id":66,"value":"hello"},"67":{"id":67,"value":"hello"},"68":{"id":68,"value":"hello"},"69":{"id":69,"value":"hello"},"7":{"id":7,"value":"hello"},"70":{"id":70,"value":"hello"},"71":{"id":71,"value":"hello"},"72":{"id":72,"value":"hello"},"73":{"id":73,"value":"hello"},"74":{"id":74,"value":"hello"},"75":{"id":75,"value":"hello"},"76":{"id":76,"value":"hello"},"77":{"id":77,"value":"hello"},"78":{"id":78,"value":"hello"},"79":{"id":79,"value":"hello"},"8":{"id":8,"value":"hello"},"80":{"id":80,"value":"hello"},"81":{"id":81,"value":"hello"},"82":{"id":82,"value":"hello"},"83":{"id":83,"value":"hello"},"84":{"id":84,"value":"hello"},"85":{"id":85,"value":"hello"},"86":{"id":86,"value":"hello"},"87":{"id":87,"value":"hello"},"88":{"id":88,"value":"hello"},"89":{"id":89,"value":"hello"},"9":{"id":9,"value":"hello"},"90":{"id":90,"value":"hello"},"91":{"id":91,"value":"hello"},"92":{"id":92,"value":"hello"},"93":{"id":93,"value":"hello"},"94":{"id":94,"value":"hello"},"95":{"id":95,"value":"hello"},"96":{"id":96,"value":"hello"},"97":{"id":97,"value":"hello"},"98":{"id":98,"value":"hello"},"99":{"id":99,"value":"hello"}},"subMsgs":[{"id":0,"value":"hello"},{"id":1,"value":"hello"},{"id":2,"value":"hello"},{"id":3,"value":"hello"},{"id":4,"value":"hello"},{"id":5,"value":"hello"},{"id":6,"value":"hello"},{"id":7,"value":"hello"},{"id":8,"value":"hello"},{"id":9,"value":"hello"},{"id":10,"value":"hello"},{"id":11,"value":"hello"},{"id":12,"value":"hello"},{"id":13,"value":"hello"},{"id":14,"value":"hello"},{"id":15,"value":"hello"},{"id":16,"value":"hello"},{"id":17,"value":"hello"},{"id":18,"value":"hello"},{"id":19,"value":"hello"},{"id":20,"value":"hello"},{"id":21,"value":"hello"},{"id":22,"value":"hello"},{"id":23,"value":"hello"},{"id":24,"value":"hello"},{"id":25,"value":"hello"},{"id":26,"value":"hello"},{"id":27,"value":"hello"},{"id":28,"value":"hello"},{"id":29,"value":"hello"},{"id":30,"value":"hello"},{"id":31,"value":"hello"},{"id":32,"value":"hello"},{"id":33,"value":"hello"},{"id":34,"value":"hello"},{"id":35,"value":"hello"},{"id":36,"value":"hello"},{"id":37,"value":"hello"},{"id":38,"value":"hello"},{"id":39,"value":"hello"},{"id":40,"value":"hello"},{"id":41,"value":"hello"},{"id":42,"value":"hello"},{"id":43,"value":"hello"},{"id":44,"value":"hello"},{"id":45,"value":"hello"},{"id":46,"value":"hello"},{"id":47,"value":"hello"},{"id":48,"value":"hello"},{"id":49,"value":"hello"},{"id":50,"value":"hello"},{"id":51,"value":"hello"},{"id":52,"value":"hello"},{"id":53,"value":"hello"},{"id":54,"value":"hello"},{"id":55,"value":"hello"},{"id":56,"value":"hello"},{"id":57,"value":"hello"},{"id":58,"value":"hello"},{"id":59,"value":"hello"},{"id":60,"value":"hello"},{"id":61,"value":"hello"},{"id":62,"value":"hello"},{"id":63,"value":"hello"},{"id":64,"value":"hello"},{"id":65,"value":"hello"},{"id":66,"value":"hello"},{"id":67,"value":"hello"},{"id":68,"value":"hello"},{"id":69,"value":"hello"},{"id":70,"value":"hello"},{"id":71,"value":"hello"},{"id":72,"value":"hello"},{"id":73,"value":"hello"},{"id":74,"value":"hello"},{"id":75,"value":"hello"},{"id":76,"value":"hello"},{"id":77,"value":"hello"},{"id":78,"value":"hello"},{"id":79,"value":"hello"},{"id":80,"value":"hello"},{"id":81,"value":"hello"},{"id":82,"value":"hello"},{"id":83,"value":"hello"},{"id":84,"value":"hello"},{"id":85,"value":"hello"},{"id":86,"value":"hello"},{"id":87,"value":"hello"},{"id":88,"value":"hello"},{"id":89,"value":"hello"},{"id":90,"value":"hello"},{"id":91,"value":"hello"},{"id":92,"value":"hello"},{"id":93,"value":"hello"},{"id":94,"value":"hello"},{"id":95,"value":"hello"},{"id":96,"value":"hello"},{"id":97,"value":"hello"},{"id":98,"value":"hello"},{"id":99,"value":"hello"}],"msgSet":[{"id":0,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":1,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":2,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":3,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":4,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":5,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":6,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":7,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":8,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":9,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":10,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":11,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":12,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":13,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":14,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":15,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":16,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":17,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":18,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":19,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":20,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":21,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":22,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":23,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":24,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":25,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":26,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":27,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":28,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":29,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":30,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":31,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":32,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":33,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":34,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":35,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":36,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":37,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":38,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":39,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":40,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":41,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":42,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":43,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":44,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":45,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":46,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":47,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":48,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":49,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":50,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":51,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":52,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":53,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":54,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":55,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":56,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":57,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":58,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":59,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":60,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":61,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":62,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":63,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":64,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":65,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":66,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":67,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":68,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":69,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":70,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":71,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":72,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":73,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":74,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":75,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":76,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":77,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":78,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":79,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":80,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":81,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":82,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":83,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":84,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":85,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":86,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":87,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":88,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":89,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":90,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":91,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":92,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":93,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":94,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":95,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":96,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":97,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":98,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]},{"id":99,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]}],"flagMsg":{"id":99,"value":"hello","subMessages":[{"id":1,"value":"hello"},{"id":2,"value":"hello"}]}}`
}

func getSubMessage(i int64) *SubMessage {
	value := "hello"
	return &SubMessage{
		Id:    &i,
		Value: &value,
	}
}

func getMessage(i int64) *Message {
	value := "hello"
	ret := &Message{
		Id:          &i,
		Value:       &value,
		SubMessages: []*SubMessage{},
	}
	ret.SubMessages = append(ret.SubMessages, getSubMessage(1))
	ret.SubMessages = append(ret.SubMessages, getSubMessage(2))
	return ret
}

func getReqDesc(name string, file string) *thrift.TypeDescriptor {
	opts := thrift.Options{}
	svc, err := opts.NewDescritorFromPath(context.Background(), "idl/echo.thrift")
	if err != nil {
		panic(err)
	}
	return svc.Functions()[name].Request().Struct().Fields()[0].Type()
}
func TestRuntimeVersion(t *testing.T) {
	println(runtime.Version())
}

func TestXxx(t *testing.T) {
	con := 1000
	wg := sync.WaitGroup{}
	desc := getReqDesc("TestObj", "idl/echo.thrift")
	for i := 0; i < con; i++ {
		wg.Add(1)
		go func(i int) {
			jd := getReqValue(100)
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic: %d\n%s", i, jd)
				}
			}()
			defer wg.Done()
			println(i)
			cv := j2t.BinaryConv{}
			ctx := context.Background()
			req, _ := http.NewHTTPRequestFromUrl("POST", "/echo", strings.NewReader(jd), http.Param{"action", "action"})
			req.Header.Set("msg", "msg")
			ctx = context.WithValue(ctx, conv.CtxKeyHTTPRequest, req)
			_, err := cv.Do(ctx, desc, []byte(jd))
			if err != nil {
				t.Fatalf("error: %d\n%s", i, jd)
			}
			require.NoError(t, err)
		}(i)
	}
	wg.Wait()
}