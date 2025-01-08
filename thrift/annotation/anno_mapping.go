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
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudwego/dynamicgo/internal/util"
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/cloudwego/thriftgo/parser"
)

type goTagMapper struct{}

func (m goTagMapper) Map(ctx context.Context, anns []parser.Annotation, desc interface{}, opts thrift.Options) (ret []parser.Annotation, next []parser.Annotation, err error) {
	for _, ann := range anns {
		out := make([]string, 0, len(ann.Values))
		for _, v := range ann.Values {
			out = append(out, util.SplitGoTags(v)...)
		}
		for _, v := range out {
			kv, err := util.SplitTagOptions(v)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid go.tag: %s", v)
			}
			switch kv[0] {
			case "json":
				if err := handleGoJSON(kv[1], &ret); err != nil {
					return nil, nil, err
				}
			}
		}
	}
	return
}

func handleGoJSON(name string, ret *[]parser.Annotation) error {
	*ret = append(*ret, parser.Annotation{
		Key:    APIKeyName,
		Values: []string{name},
	})
	return nil
}

type apiBodyMapper struct{}

func (m apiBodyMapper) Map(ctx context.Context, anns []parser.Annotation, desc interface{}, opts thrift.Options) (ret []parser.Annotation, next []parser.Annotation, err error) {
	if len(anns) > 1 {
		return nil, nil, errors.New("api.body must be unique")
	}
	for _, ann := range anns {
		if len(ann.Values) != 1 {
			return nil, nil, errors.New("api.body must have a value")
		}
		isRoot := ctx.Value(thrift.CtxKeyIsBodyRoot)
		// special fast-path: if the field is at body root, we don't need to add api.body to call GetMapBody()
		if opts.ApiBodyFastPath && isRoot != nil && isRoot.(bool) {
			ret = append(ret, parser.Annotation{
				Key:    APIKeyName,
				Values: []string{ann.Values[0]},
			})
			continue
		} else {
			ret = append(ret, parser.Annotation{
				Key:    "api.body",
				Values: []string{ann.Values[0]},
			})
		}
	}
	return
}

type sourceMapper struct{}

func (m sourceMapper) Map(ctx context.Context, anns []parser.Annotation, desc interface{}, opts thrift.Options) (ret []parser.Annotation, next []parser.Annotation, err error) {
	field, ok := desc.(*parser.Field)
	if !ok || field == nil {
		return nil, nil, errors.New("target must be used on field")
	}
	name := m.decideNameCase(field)
	for _, ann := range anns {
		if len(ann.Values) != 1 {
			return nil, nil, errors.New("source must have a value")
		}
		var val = strings.ToLower(ann.Values[0])
		var key = "api."
		switch val {
		case "query":
			key += "query"
		case "header":
			key += "header"
		case "body":
			key = "raw.body"
		case "cookie":
			key += "cookie"
		case "post":
			key += "form"
		case "path":
			key += "path"
		case "raw_uri":
			key += "raw_uri"
		case "raw_body":
			key += "raw_body"
			// case "body_dynamic":
			// 	key = "agw.body_dynamic"
		case "not_body_struct":
			key = "api.no_body_struct"
		default:
			continue
		}
		ret = append(ret, parser.Annotation{
			Key:    key,
			Values: []string{name},
		})
	}
	return
}

func (self sourceMapper) decideNameCase(field *parser.Field) string {
	name := field.Name
	flag := false
	// firstly check if there is api.keu annotation
	for _, ann := range FindAnnotations(field.Annotations, APIKeyName, "agw.key") {
		if len(ann.Values) == 1 && ann.Values[0] != "" {
			name = ann.Values[0]
			flag = true
			break
		}
	}
	// if not found, try use agw.to_snake or agw.to_lower_camel_case
	if !flag {
		anns := FindAnnotations(field.Annotations, NameCaseKeys...)
		if len(anns) > 2 || len(anns) == 0 {
			return name
		}
		var pre = ""
		if len(anns) == 2 {
			pre = anns[1].Key
		}
		c, _ := nameCaseMapper{}.decideNameCase(pre, anns[0].Key, anns[0].Values[0])
		name = util.ConvertNameCase(c, name)
	}
	return name
}

type targetMapper struct{}

func (m targetMapper) Map(ctx context.Context, anns []parser.Annotation, desc interface{}, opts thrift.Options) (ret []parser.Annotation, next []parser.Annotation, err error) {
	field, ok := desc.(*parser.Field)
	if !ok || field == nil {
		return nil, nil, errors.New("target must be used on field")
	}
	name := sourceMapper{}.decideNameCase(field)
	// handle normal xxx.target
	for _, ann := range anns {
		if len(ann.Values) != 1 {
			return nil, nil, errors.New("target must have a value")
		}

		var val = strings.ToLower(ann.Values[0])
		var key = "api."
		switch val {
		case "header":
			key += "header"
		case "body":
			// directy set as body or write body content
			key = "raw.body"
		case "cookie":
			key += "cookie"
		case "http_code":
			key += "http_code"
		// case "body_dynamic":
		// 	key = "agw.body_dynamic"
		case "ignore":
			key = thrift.AnnoKeyDynamicGoDeprecated
		default:
			continue
		}
		ret = append(ret, parser.Annotation{
			Key:    key,
			Values: []string{name},
		})
	}
	return
}

func FindAnnotations(anns []*parser.Annotation, keys ...string) (ret []*parser.Annotation) {
	for _, ann := range anns {
		for _, key := range keys {
			if ann.Key == key {
				ret = append(ret, ann)
			}
		}
	}
	return
}

type nameCaseMapper struct{}

func (m nameCaseMapper) Map(ctx context.Context, anns []parser.Annotation, desc interface{}, opts thrift.Options) (ret []parser.Annotation, next []parser.Annotation, err error) {
	if len(anns) > 2 {
		return nil, nil, errors.New("name case must be unique")
	}
	if len(anns[0].Values) != 1 {
		return nil, nil, errors.New("name case must have one value")
	}

	// decide the name case based on both the previous and current annotations
	cur := anns[0].Key
	curval := anns[0].Values[0]
	var pre string
	if len(anns) == 2 {
		pre = anns[1].Key
	}
	final, pkg := m.decideNameCase(pre, cur, curval)

	var r parser.Annotation
	if f, ok := desc.(*parser.Field); ok {
		r.Key = APIKeyName
		r.Values = []string{util.ConvertNameCase(final, f.Name)}
		return []parser.Annotation{r}, nil, nil
	} else {
		r.Key = pkg + "." + m.caseToKey(final)
		r.Values = []string{"true"}
		return nil, []parser.Annotation{r}, nil
	}
}

func (m nameCaseMapper) keyToCase(key string) (meta.NameCase, string) {
	pkg, cur := util.SplitPrefix(key)
	switch cur {
	case "to_upper_camel_case":
		return meta.CaseUpperCamel, pkg
	case "to_lower_camel_case":
		return meta.CaseLowerCamel, pkg
	case "to_snake", "to_snake_case":
		return meta.CaseSnake, pkg
	default:
		return meta.CaseDefault, pkg
	}
}

func (m nameCaseMapper) caseToKey(c meta.NameCase) string {
	switch c {
	case meta.CaseUpperCamel:
		return "to_upper_camel_case"
	case meta.CaseLowerCamel:
		return "to_lower_camel_case"
	case meta.CaseSnake:
		return "to_snake"
	default:
		return ""
	}
}

func (m nameCaseMapper) valToEnable(val string) bool {
	var enable = true
	if val != "" {
		enable, _ = strconv.ParseBool(val)
	}
	return enable
}

// decideNameCase decides the final name case based on the previous and current annotations.
// current has higher priority than previous when it is true or is the same key.
// otherwise, the previous annotation is used.
func (m nameCaseMapper) decideNameCase(pre string, cur string, curval string) (meta.NameCase, string) {
	pc, _ := m.keyToCase(pre)
	cc, pkg := m.keyToCase(cur)
	enable := m.valToEnable(curval)
	if pc == cc {
		if enable {
			return cc, pkg
		}
		return meta.CaseDefault, pkg
	} else {
		if enable {
			return cc, pkg
		}
		return pc, pkg
	}
}
