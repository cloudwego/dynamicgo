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

import (
	"context"
	"fmt"

	"github.com/cloudwego/dynamicgo/internal/util_test"
)

// GetFnDescFromFile get a fucntion descriptor from idl path (relative to your git root) and
// the function name
func GetFnDescFromFile(filePath, fnName string, opts Options) *FunctionDescriptor {
	svc, err := opts.NewDescritorFromPath(context.Background(), util_test.MustGitPath(filePath))
	if err != nil {
		panic(fmt.Errorf("%s:%s", util_test.MustGitPath(filePath), err))
	}
	fn, err := svc.LookupFunctionByMethod(fnName)
	if err != nil {
		panic(err)
	}
	return fn
}

// FnResponse get the normal response type
func FnResponse(fn *FunctionDescriptor) *TypeDescriptor {
	// let-it-fail: it panic when something is nil
	return fn.Response().Struct().FieldById(0).Type()
}

// FnWholeResponse get the normal response type
func FnWholeResponse(fn *FunctionDescriptor) *TypeDescriptor {
	// let-it-fail: it panic when something is nil
	return fn.Response()
}

// FnRequest
// We assume the request only have one argument and the only argument it the type we want.
func FnRequest(fn *FunctionDescriptor) *TypeDescriptor {
	// let-it-fail: it panic when something is nil
	return fn.Request().Struct().Fields()[0].Type()
}
