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

package generic

import (
	"os"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example2"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/stretchr/testify/require"
)

var (
	debugAsyncGC = os.Getenv("SONIC_NO_ASYNC_GC") == ""
)

func TestMain(m *testing.M) {
	go func() {
		if !debugAsyncGC {
			return
		}
		println("Begin GC looping...")
		for {
			runtime.GC()
			debug.FreeOSMemory()
		}
	}()
	time.Sleep(time.Millisecond)
	m.Run()
}

func BenchmarkSetOne_DynamicGo(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()
	v := NewValue(desc, data)
	d := desc.Struct().FieldByKey("Base").Type().Struct().FieldByKey("Extra").Type().Elem()
	p := thrift.NewBinaryProtocolBuffer()
	exp := "中文"
	p.WriteString(exp)
	buf := p.Buf
	vv := NewValue(d, buf)
	ps := []Path{NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("b")}
	_, err2 := v.SetByPath(vv, ps...)
	require.Nil(b, err2)
	s2 := v.GetByPath(ps...)
	require.Empty(b, s2.Error())
	f2, _ := s2.String()
	require.Equal(b, exp, f2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = v.SetByPath(vv, ps...)
	}
}

func BenchmarkSetMany_DynamicGo(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()
	d1 := desc.Struct().FieldByKey("Msg").Type()
	d2 := desc.Struct().FieldByKey("Subfix").Type()
	v := NewValue(desc, data)
	p := thrift.NewBinaryProtocolBuffer()

	e1 := "test1"
	p.WriteString(e1)
	v1 := NewValue(d1, []byte(string(p.Buf)))
	p.Buf = p.Buf[:0]
	e2 := float64(-255.0001)
	p.WriteDouble(e2)
	v2 := NewValue(d2, []byte(string(p.Buf)))
	p.Buf = p.Buf[:0]
	v3 := v.GetByPath(NewPathFieldName("Base"))
	ps := []PathNode{
		{
			Path: NewPathFieldId(1),
			Node: v1.Node,
		},
		{
			Path: NewPathFieldId(32767),
			Node: v2.Node,
		},
		{
			Path: NewPathFieldId(255),
			Node: v3.Node,
		},
	}
	opts := &Options{}
	err := v.SetMany(ps, opts)
	require.Nil(b, err)
	exp := example2.NewExampleReq()
	if _, err := exp.FastRead(v.Raw()); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.SetMany(ps, opts)
	}
}

func BenchmarkGetOne_DynamicGo(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()
	v := NewValue(desc, data)
	vv, err := v.FieldByName("Base").FieldByName("Extra").GetByStr("c").String()
	require.Nil(b, err)
	require.Equal(b, "C", vv)
	vv, err = v.Field(255).Field(6).GetByStr("c").String()
	require.Nil(b, err)
	require.Equal(b, "C", vv)
	vv, err = v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("c")).String()
	require.Nil(b, err)
	require.Equal(b, "C", vv)
	b.Run("ByName", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = v.FieldByName("Base").FieldByName("Extra").GetByStr("c").String()
		}
	})
	b.Run("ById", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = v.Field(255).Field(6).GetByStr("c").String()
		}
	})
	b.Run("ByPath", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = v.GetByPath(NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("c")).String()
		}
	})
}

func BenchmarkGetMany_DynamicGo(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()
	v := NewValue(desc, data)

	b.Run("ByName", func(b *testing.B) {
		tree := GetSampleTree(v)
		opts := Options{}
		if err := tree.Assgin(true, &opts); err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tree.Assgin(true, &opts)
		}
	})
	b.Run("ById", func(b *testing.B) {
		tree := GetSampleTreeById(v)
		opts := Options{}
		if err := tree.Assgin(true, &opts); err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = tree.Assgin(true, &opts)
		}
	})
}

func BenchmarkUnmarshalAll_KitexFast(b *testing.B) {
	data := getExampleData()
	exp := example2.NewExampleReq()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = exp.FastRead(data)
		_ = exp.Base.Client
	}
}

func BenchmarkMarshalMany_DynamicGo(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()
	v := NewValue(desc, data)

	b.Run("ByName", func(b *testing.B) {
		tree := GetSampleTree(v)
		opts := Options{}
		if err := tree.Assgin(true, &opts); err != nil {
			b.Fatal(err)
		}
		_, err := tree.Marshal(&opts)
		if err != nil {
			b.Fatal(err)
		}
		// b.SetBytes(int64(len(out)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tree.Marshal(&opts)
		}
	})

	b.Run("ById", func(b *testing.B) {
		tree := GetSampleTreeById(v)
		opts := Options{}
		if err := tree.Assgin(true, &opts); err != nil {
			b.Fatal(err)
		}
		_, err := tree.Marshal(&opts)
		if err != nil {
			b.Fatal(err)
		}
		// b.SetBytes(int64(len(out)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = tree.Marshal(&opts)
		}
	})
}

func BenchmarkMarshalTo_DynamicGo(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()
	partial := getExamplePartialDesc()

	exp := example2.NewExampleReq()
	v := NewValue(desc, data)
	_, err := exp.FastRead(data)
	require.Nil(b, err)
	opts := Options{
		WriteDefault: true,
	}
	_, err = v.MarshalTo(partial, &opts)
	require.Nil(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = v.MarshalTo(partial, &opts)
	}
}

func BenchmarkMarshalAll_KitexFast(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()
	v := NewValue(desc, data)
	tree := GetSampleTreeById(v)
	opts := Options{}
	if err := tree.Assgin(true, &opts); err != nil {
		b.Fatal(err)
	}
	_, err := tree.Marshal(&opts)
	if err != nil {
		b.Fatal(err)
	}

	exp := example2.NewExampleReq()
	_, err = exp.FastRead(data)
	require.Nil(b, err)
	buf := make([]byte, exp.BLength())
	if exp.FastWriteNocopy(buf, nil) <= 0 {
		b.Fatal(buf)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, exp.BLength())
		_ = exp.FastWriteNocopy(buf, nil)
	}
}

func BenchmarkGetAll_DynamicGo(b *testing.B) {
	desc := getExampleDesc()
	data := getExampleData()
	v := NewValue(desc, data)
	opts := Options{}
	children := make([]PathNode, 0, DefaultNodeSliceCap)
	p := PathNode{
		Node: v.Node,
		Next: children,
	}
	err := v.Children(&children, false, &opts)
	require.Nil(b, err)

	b.Run("skip", func(b *testing.B) {
		opts := Options{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.ResetValue()
			err = v.Children(&p.Next, false, &opts)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("load_all", func(b *testing.B) {
		opts := Options{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.ResetValue()
			_ = v.Children(&p.Next, true, &opts)
		}
	})

	b.Run("only_struct", func(b *testing.B) {
		opts := Options{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			p.ResetValue()
			_ = v.Children(&p.Next, true, &opts)
		}
	})
}
