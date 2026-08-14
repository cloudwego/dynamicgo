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

package annotation

import (
	"fmt"
	"sync"
	"testing"

	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	InitAGWAnnos()
	m.Run()
}

func TestGoTagJSON(t *testing.T) {
	p, err := GetDescFromContent(`
	namespace go kitex.test.server
	struct Base {
		1: required string required_field
		999: required string double_quote (go.tag="json:\"test2,string\" protobuf:\"bytes,1,opt,name=test2\"")
		888: string no_quote (go.tag='protobuf:"bytes,1,opt,name=test3" json:"test3,string"')
		3: string pass 
	}

	service InboxService {
		string ExampleMethod(1: Base req)
	}
	`, "ExampleMethod")
	if err != nil {
		t.Fatal(err)
	}
	req := p.Request().Struct().Fields()[0].Type()
	assert.Equal(t, "test2", req.Struct().FieldById(999).Alias())
	assert.Equal(t, "test3", req.Struct().FieldById(888).Alias())

	_, err = GetDescFromContent(`
	namespace go kitex.test.server
	struct Base {
		1: required string required_field
		999: required string double_quote (go.tag="json:"test2,string" protobuf:\"bytes,1,opt,name=test2\"")
		3: string pass 
	}

	service InboxService {
		string ExampleMethod(1: Base req)
	}
	`, "ExampleMethod")
	require.Error(t, err)
	println(err.Error())

	_, err = GetDescFromContent(`
	namespace go kitex.test.server
	struct Base {
		1: required string required_field
		999: required string double_quote (go.tag='json:\"test2,string\" protobuf:\"bytes,1,opt,name=test2\"')
		3: string pass 
	}

	service InboxService {
		string ExampleMethod(1: Base req)
	}
	`, "ExampleMethod")
	require.NoError(t, err)
}

func TestGoTagMapperOptionsIsolation(t *testing.T) {
	content := `
	namespace go kitex.test.server
	struct Base {
		1: string Msg (go.tag = "json:\"message\"")
	}

	service InboxService {
		string ExampleMethod(1: Base req)
	}
	`

	opts := thrift.NewDefaultOptions()
	opts.RemoveAnnotationMapper(thrift.AnnoScopeField, "go.tag")
	p, err := GetDescFromContentWithOptions(content, "ExampleMethod", opts)
	require.NoError(t, err)
	req := p.Request().Struct().Fields()[0].Type()
	require.Equal(t, "Msg", req.Struct().FieldById(1).Alias())

	p, err = GetDescFromContent(content, "ExampleMethod")
	require.NoError(t, err)
	req = p.Request().Struct().Fields()[0].Type()
	require.Equal(t, "message", req.Struct().FieldById(1).Alias())

	opts = thrift.NewDefaultOptions()
	opts.RemoveAnnotationMapper(thrift.AnnoScopeField, "go.tag")
	opts.RegisterAnnotationMapper(thrift.AnnoScopeField, goTagMapper{}, "go.tag")
	p, err = GetDescFromContentWithOptions(content, "ExampleMethod", opts)
	require.NoError(t, err)
	req = p.Request().Struct().Fields()[0].Type()
	require.Equal(t, "message", req.Struct().FieldById(1).Alias())
}

func TestAnnotationMapperOptionsCopyOnWrite(t *testing.T) {
	content := `
	namespace go kitex.test.server
	struct Base {
		1: string Msg (go.tag = "json:\"message\"")
	}

	service InboxService {
		string ExampleMethod(1: Base req)
	}
	`

	base := thrift.NewDefaultOptions()
	base.RegisterAnnotationMapper(thrift.AnnoScopeField, goTagMapper{}, "go.tag")
	copied := base
	require.True(t, base == copied)

	const workers = 8
	var wg sync.WaitGroup
	errCh := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(disableGoTag bool) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				opts := base
				want := "message"
				if disableGoTag {
					opts.RemoveAnnotationMapper(thrift.AnnoScopeField, "go.tag")
					want = "Msg"
				} else {
					opts.RegisterAnnotationMapper(thrift.AnnoScopeField, goTagMapper{}, "go.tag")
				}

				p, err := GetDescFromContentWithOptions(content, "ExampleMethod", opts)
				if err != nil {
					errCh <- err
					return
				}
				req := p.Request().Struct().Fields()[0].Type()
				if got := req.Struct().FieldById(1).Alias(); got != want {
					errCh <- fmt.Errorf("unexpected field alias: got %q, want %q", got, want)
					return
				}
			}
		}(i%2 == 0)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}

	p, err := GetDescFromContentWithOptions(content, "ExampleMethod", base)
	require.NoError(t, err)
	req := p.Request().Struct().Fields()[0].Type()
	require.Equal(t, "message", req.Struct().FieldById(1).Alias())
}

func TestAnnotationMapperConcurrentDefaultMutationAndParse(t *testing.T) {
	t.Cleanup(func() {
		thrift.RegisterAnnotationMapper(thrift.AnnoScopeField, goTagMapper{}, "go.tag")
	})
	content := `
	namespace go kitex.test.server
	struct Base {
		1: string Msg (go.tag = "json:\"message\"")
	}

	service InboxService {
		string ExampleMethod(1: Base req)
	}
	`

	var wg sync.WaitGroup
	errCh := make(chan error, 512)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(disableGoTag bool) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				opts := thrift.NewDefaultOptions()
				if disableGoTag {
					opts.RemoveAnnotationMapper(thrift.AnnoScopeField, "go.tag")
				}
				_, err := GetDescFromContentWithOptions(content, "ExampleMethod", opts)
				if err != nil {
					errCh <- err
					return
				}
			}
		}(i%2 == 0)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			thrift.RemoveAnnotationMapper(thrift.AnnoScopeField, "go.tag")
			thrift.RegisterAnnotationMapper(thrift.AnnoScopeField, goTagMapper{}, "go.tag")
		}
	}()

	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}
}

func TestApiKey(t *testing.T) {
	p, err := GetDescFromContent(`
	namespace go kitex.test.server

	struct Base {
		1: string FieldName1 (agw.source="query", api.key="test1")
		2: string FieldName2 (agw.target="header")
		3: string FieldName3 (agw.source="query", go.tag="json:\"test3\"")
	} (agw.to_snake="true")

	service InboxService {
		string ExampleMethod(1: Base req)
	}
	`, "ExampleMethod")
	require.NoError(t, err)
	req := p.Request().Struct().Fields()[0].Type()
	require.Equal(t, "test1", req.Struct().FieldById(1).Alias())
	require.Equal(t, "test1", req.Struct().FieldById(1).HTTPMappings()[0].(apiQuery).value)
	require.Equal(t, "field_name_2", req.Struct().FieldById(2).Alias())
	require.Equal(t, "field_name_2", req.Struct().FieldById(2).HTTPMappings()[0].(apiHeader).value)
	require.Equal(t, "test3", req.Struct().FieldById(3).Alias())
	require.Equal(t, "field_name_3", req.Struct().FieldById(3).HTTPMappings()[0].(apiQuery).value)
}

func TestIgnoreField(t *testing.T) {
	p, err := GetDescFromContent(`
	namespace go kitex.test.server

	struct Base {
		0: string DefaultField
		1: string FieldName1 (agw.target="ignore")
		2: string FieldName2 (agw.source="ignore")
	}

	service InboxService {
		string ExampleMethod(1: Base req)
	}
	`, "ExampleMethod")
	require.NoError(t, err)
	req := p.Request().Struct().Fields()[0].Type()
	require.Equal(t, (*thrift.FieldDescriptor)(nil), req.Struct().FieldById(1))
	require.Equal(t, (*thrift.FieldDescriptor)(nil), req.Struct().FieldByKey("FieldName1"))
	require.NotNil(t, req.Struct().FieldById(2))
	require.NotNil(t, req.Struct().FieldByKey("FieldName2"))
}

func TestApiNoneField(t *testing.T) {
	p, err := GetDescFromContent(`
	namespace go kitex.test.server

	struct ExampleStruct {
		0: string FieldName1
		1: string FieldName2 (api.none="")
	}

	struct Base {
		0: string DefaultField
		1: string FieldName1 (api.none="")
		2: ExampleStruct ExampleStruct
	}

	service InboxService {
		Base ExampleMethod(1: Base req)
	}
	`, "ExampleMethod")
	require.NoError(t, err)
	req := p.Request().Struct().Fields()[0].Type()
	resp := p.Response().Struct().Fields()[0].Type()
	require.NotNil(t, req.Struct().FieldById(1))
	require.NotNil(t, req.Struct().FieldByKey("FieldName1"))
	require.NotNil(t, req.Struct().FieldById(2).Type().Struct().FieldById(1))
	require.NotNil(t, req.Struct().FieldByKey("ExampleStruct").Type().Struct().FieldByKey("FieldName2"))
	require.Equal(t, (*thrift.FieldDescriptor)(nil), resp.Struct().FieldById(1))
	require.Equal(t, (*thrift.FieldDescriptor)(nil), resp.Struct().FieldByKey("FieldName1"))
	require.Equal(t, (*thrift.FieldDescriptor)(nil), resp.Struct().FieldById(2).Type().Struct().FieldById(1))
	require.Equal(t, (*thrift.FieldDescriptor)(nil), resp.Struct().FieldByKey("ExampleStruct").Type().Struct().FieldByKey("FieldName2"))
}
