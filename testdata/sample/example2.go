/**
 * Copyright 2022 CloudWeGo Authors.
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

package sample

import (
	"math"

	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/base"
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example2"
)

var (
	Example2Obj   = GetExample2Req()
	Example2Super = GetExample2ReqSuper()
)

func GetBase() *base.Base {
	ret := base.NewBase()
	ret.LogID = "log_id"
	ret.TrafficEnv = &base.TrafficEnv{
		Open: true,
		Env:  "env",
	}
	ret.Extra = map[string]string{"a": "A", "c": "C", "b": "B"}
	return ret
}

func GetExample2Req() *example2.ExampleReq {
	obj := example2.NewExampleReq()
	obj.Msg = new(string)
	*obj.Msg = "hello"
	obj.Subfix = math.MaxFloat64
	obj.InnerBase = example2.NewInnerBase()
	obj.InnerBase.Binary = []byte{0xff, 0xff, 0xff, 0xff}
	obj.InnerBase.Bool = true
	obj.InnerBase.Byte = math.MaxInt8
	obj.InnerBase.Int16 = math.MinInt16
	obj.InnerBase.Int32 = math.MaxInt32
	obj.InnerBase.Int64 = math.MinInt64
	obj.InnerBase.Double = math.MaxFloat64
	obj.InnerBase.String_ = "你好"
	obj.InnerBase.ListInt32 = []int32{math.MinInt32, 0, math.MaxInt32}
	obj.InnerBase.SetInt32_ = []int32{math.MinInt32, 0, math.MaxInt32}
	obj.InnerBase.MapStringString = map[string]string{"a": "A", "c": "C", "b": "B"}
	obj.InnerBase.Foo = 1
	obj.InnerBase.MapInt8String = map[int8]string{1: "A", 2: "B", 3: "C"}
	obj.InnerBase.MapInt16String = map[int16]string{1: "A", 2: "B", 3: "C"}
	obj.InnerBase.MapInt32String = map[int32]string{1: "A", 2: "B", 3: "C"}
	obj.InnerBase.MapInt64String = map[int64]string{1: "A", 2: "B", 3: "C"}
	obj.InnerBase.MapDoubleString = map[float64]string{0.1: "A", 0.2: "B", 0.3: "C"}

	innerx := example2.NewInnerBase()
	innerx.Base = GetBase()
	obj.InnerBase.ListInnerBase = []*example2.InnerBase{
		innerx,
	}
	obj.InnerBase.MapInnerBaseInnerBase = map[*example2.InnerBase]*example2.InnerBase{
		innerx: innerx,
	}

	obj.InnerBase.Base = GetBase()
	obj.InnerBase.Base.LogID = "log_id_inner"
	obj.InnerBase.Base.TrafficEnv = &base.TrafficEnv{
		Open: true,
		Env:  "env_inner",
	}
	obj.Base = GetBase()
	return obj
}

func GetExample2ReqSuper() *example2.ExampleSuper {
	obj := example2.NewExampleSuper()
	obj.Ex1 = "ex"
	ex2 := "ex"
	obj.Ex2 = &ex2
	obj.Ex4 = "ex"
	// obj.MapStructStruct = map[*example2.BasePartial]*example2.BasePartial{
	// 	example2.NewBasePartial(): example2.NewBasePartial(),
	// }
	self := example2.NewSelfRef()
	self.Self = example2.NewSelfRef()
	obj.SelfRef = self

	c := GetExample2Req()
	obj.InnerBase = c.InnerBase
	obj.Msg = c.Msg
	obj.Base = c.Base
	obj.Subfix = c.Subfix
	return obj
}
