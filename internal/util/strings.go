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

package util

import (
	"strings"

	"github.com/cloudwego/dynamicgo/meta"
	"github.com/fatih/structtag"
	"github.com/iancoleman/strcase"
)

func ConvertNameCase(opt meta.NameCase, name string) string {
	switch opt {
	case meta.CaseUpperCamel:
		return ToUpperCamelCase(name)
	case meta.CaseSnake:
		return ToSnakeCase(name)
	case meta.CaseLowerCamel:
		return ToLowerCamelCase(name)
	default:
		return name
	}
}

// ToSnakeCase converts a camelCase string to snake_case.
func ToSnakeCase(name string) string {
	return strcase.ToSnake(name)
}

// ToCamelCase converts a snake_case string to camelCase.
func ToUpperCamelCase(name string) string {
	return strcase.ToCamel(name)
}

// LowerFirstLetter
func ToLowerCamelCase(name string) string {
	return strcase.ToLowerCamel(name)
}

func SplitTagOptions(tag string) (ret []string, err error) {
	tags, err := structtag.Parse(tag)
	if err != nil {
		// try replace `\"`
		tag = strings.Replace(tag, `\"`, `"`, -1)
		tags, err = structtag.Parse(tag)
		if err != nil {
			return nil, err
		}
	}
	// kv := strings.Trim(tag,  "\n\b\t\r")
	// i := strings.Index(kv, ":")
	// if i <= 0 {
	// 	return nil, fmt.Errorf("invalid go tag: %s", tag)
	// }
	// ret = append(ret, kv[:i])
	// val := strings.Trim(kv[i+1:], "\n\b\t\r")

	// v, err := strconv.Unquote(kv[i+1:])
	// if err != nil {
	// 	return ret, err
	// }
	// vs := strings.Split(v, ",")
	// ret = append(ret, vs...)
	for _, t := range tags.Tags() {
		ret = append(ret, t.Key)
		ret = append(ret, t.Name)
	}
	return ret, nil
}

func SplitGoTags(input string) []string {
	out := make([]string, 0, 4)
	ns := len(input)

	flag := false
	prev := 0
	i := 0
	for i = 0; i < ns; i++ {
		c := input[i]
		if c == '"' {
			flag = !flag
		}
		if !flag && c == ' ' {
			if prev < i {
				out = append(out, input[prev:i])
			}
			prev = i + 1
		}
	}
	if i != 0 && prev < i {
		out = append(out, input[prev:i])
	}

	return out
}

func SplitPrefix(t string) (pkg, name string) {
	idx := strings.Index(t, ".")
	if idx == -1 {
		return "", t
	}
	return t[:idx], t[idx+1:]
}

func SplitSubfix(t string) (typ, val string) {
	idx := strings.LastIndex(t, ".")
	if idx == -1 {
		return "", t
	}
	return t[:idx], t[idx+1:]
}

func SplitSubfix2(t string) (typ, val string) {
	idx := strings.IndexAny(t, ".")
	if idx == -1 {
		return "", t
	}
	return t[:idx], t[idx+1:]
}
