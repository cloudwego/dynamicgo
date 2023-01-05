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
	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example3"
)

func GetEmptyInnerBase3() *example3.InnerBase {
	return &example3.InnerBase{
		ListInt32:          []int32{},
		MapStringString:    map[string]string{},
		SetInt32_:          []int32{},
		MapInt32String:     map[int32]string{},
		Binary:             []byte{},
		MapInt8String:      map[int8]string{},
		MapInt16String:     map[int16]string{},
		MapInt64String:     map[int64]string{},
		ListInnerBase:      []*example3.InnerBase{},
		MapStringInnerBase: map[string]*example3.InnerBase{},
		Base:               &base.Base{},
	}
}

func GetInnerBase3() *example3.InnerBase {
	in := example3.NewInnerBase()
	in.Binary = []byte{0xff, 0xff, 0xff, 0xff}
	in.Bool = true
	in.Byte = math.MaxInt8
	in.Int16 = math.MinInt16
	in.Int32 = math.MaxInt32
	in.Int64 = math.MinInt64
	in.Double = math.MaxFloat64
	in.String_ = "hello"
	in.ListInt32 = []int32{math.MinInt32, 0, math.MaxInt32}
	in.SetInt32_ = []int32{math.MinInt32, 0, math.MaxInt32}
	in.MapStringString = map[string]string{"a": "A", "c": "C", "b": "B"}
	in.Foo = 1
	in.MapInt8String = map[int8]string{1: "A", 2: "B", 3: "C"}
	in.MapInt16String = map[int16]string{1: "A", 2: "B", 3: "C"}
	in.MapInt32String = map[int32]string{1: "A", 2: "B", 3: "C"}
	in.MapInt64String = map[int64]string{1: "A", 2: "B", 3: "C"}
	in.InnerQuery = "中文"

	innerx := GetEmptyInnerBase3()
	innerx.Base = GetBase()
	in.ListInnerBase = []*example3.InnerBase{
		innerx,
	}
	in.MapStringInnerBase = map[string]*example3.InnerBase{
		"innerx": innerx,
	}

	in.Base = GetBase()
	in.Base.LogID = "log_id_inner"
	in.Base.TrafficEnv = &base.TrafficEnv{
		Open: true,
		Env:  "env_inner",
	}
	return in
}

func GetExample3Req() *example3.ExampleReq {
	obj := example3.NewExampleReq()
	obj.Msg = new(string)
	*obj.Msg = "hello"
	obj.Subfix = math.MaxFloat64
	obj.InnerBase = GetInnerBase3()
	obj.Base = GetBase()
	obj.Code = 1024
	return obj
}

func GetExample3Resp() *example3.ExampleResp {
	obj := example3.NewExampleResp()
	obj.Code = 1024
	obj.Msg = "中文"
	obj.Status = 202
	f := float64(-0.0000001)
	obj.Cookie = &f
	b := bool(true)
	obj.Header = &b
	obj.Subfix = 0
	obj.InnerBase = GetInnerBase3()
	obj.BaseResp = &base.BaseResp{
		StatusMessage: "a",
		StatusCode:    2048,
	}
	return obj
}
