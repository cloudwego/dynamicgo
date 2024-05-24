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
	"fmt"
	"strings"
	"testing"

	"github.com/cloudwego/dynamicgo/thrift"
	"github.com/stretchr/testify/require"
)

func TestNewTypedNode(t *testing.T) {
	type args struct {
		typ      thrift.Type
		et       thrift.Type
		kt       thrift.Type
		children []PathNode
	}
	tests := []struct {
		name    string
		args    args
		opts    Options
		wantRet interface{}
	}{
		{"empty map", args{thrift.MAP, thrift.DOUBLE, thrift.STRING, []PathNode{}}, Options{}, map[string]interface{}{}},
		{"empty list", args{thrift.LIST, thrift.DOUBLE, 0, []PathNode{}}, Options{}, []interface{}{}},
		{"empty struct", args{thrift.STRUCT, 0, 0, []PathNode{}}, Options{MapStructById: true}, map[thrift.FieldID]interface{}{}},
		{"empty string", args{thrift.STRING, 0, 0, []PathNode{}}, Options{}, ""},
		{"empty bool", args{thrift.BOOL, 0, 0, []PathNode{}}, Options{}, false},
		{"empty int8", args{thrift.I08, 0, 0, []PathNode{}}, Options{}, int(0)},
		{"empty int16", args{thrift.I16, 0, 0, []PathNode{}}, Options{}, int(0)},
		{"empty int32", args{thrift.I32, 0, 0, []PathNode{}}, Options{}, int(0)},
		{"empty int64", args{thrift.I64, 0, 0, []PathNode{}}, Options{}, int(0)},
		{"empty double", args{thrift.DOUBLE, 0, 0, []PathNode{}}, Options{}, float64(0)},
		{"bool", args{thrift.BOOL, 0, 0, []PathNode{{Node: NewNodeBool(true)}}}, Options{}, true},
		{"string", args{thrift.STRING, 0, 0, []PathNode{{Node: NewNodeString("1")}}}, Options{}, "1"},
		{"binary", args{thrift.STRING, 0, 0, []PathNode{{Node: NewNodeBinary([]byte{1})}}}, Options{CastStringAsBinary: true}, []byte{1}},
		{"byte", args{thrift.BYTE, 0, 0, []PathNode{{Node: NewNodeByte(1)}}}, Options{}, int(1)},
		{"i16", args{thrift.I16, 0, 0, []PathNode{{Node: NewNodeInt16(1)}}}, Options{}, int(1)},
		{"i32", args{thrift.I32, 0, 0, []PathNode{{Node: NewNodeInt32(1)}}}, Options{}, int(1)},
		{"i64", args{thrift.I64, 0, 0, []PathNode{{Node: NewNodeInt64(1)}}}, Options{}, int(1)},
		{"double", args{thrift.DOUBLE, 0, 0, []PathNode{{Node: NewNodeDouble(1)}}}, Options{}, float64(1)},
		{"list", args{thrift.LIST, thrift.BYTE, 0, []PathNode{{Path: NewPathIndex(0), Node: NewNodeByte(1)}}}, Options{}, []interface{}{int(1)}},
		{"set", args{thrift.SET, thrift.BYTE, 0, []PathNode{{Path: NewPathIndex(0), Node: NewNodeByte(1)}}}, Options{}, []interface{}{int(1)}},
		{"int map", args{thrift.MAP, thrift.BYTE, thrift.BYTE, []PathNode{{Path: NewPathIntKey(1), Node: NewNodeByte(1)}}}, Options{}, map[int]interface{}{int(1): int(1)}},
		{"string map", args{thrift.MAP, thrift.BYTE, thrift.STRING, []PathNode{{Path: NewPathStrKey("1"), Node: NewNodeByte(1)}}}, Options{}, map[string]interface{}{"1": int(1)}},
		{"any map + key list", args{thrift.MAP, thrift.BYTE, thrift.LIST, []PathNode{{Path: NewPathBinKey(NewNodeList([]interface{}{1, 2}).Raw()), Node: NewNodeByte(1)}}}, Options{}, map[interface{}]interface{}{&[]interface{}{int(1), int(2)}: int(1)}},
		{"any map + key map", args{thrift.MAP, thrift.BYTE, thrift.MAP, []PathNode{{Path: NewPathBinKey(NewNodeMap(map[interface{}]interface{}{1: 2}, &Options{}).Raw()), Node: NewNodeByte(1)}}}, Options{}, map[interface{}]interface{}{&map[int]interface{}{1: 2}: int(1)}},
		{"struct", args{thrift.STRUCT, 0, 0, []PathNode{{Path: NewPathFieldId(1), Node: NewNodeByte(1)}}}, Options{MapStructById: true}, map[thrift.FieldID]interface{}{thrift.FieldID(1): int(1)}},
		{"struct + struct", args{thrift.STRUCT, 0, 0, []PathNode{{Path: NewPathFieldId(1), Node: NewNodeStruct(map[thrift.FieldID]interface{}{1: 1}, &Options{})}}}, Options{MapStructById: true}, map[thrift.FieldID]interface{}{thrift.FieldID(1): map[thrift.FieldID]interface{}{thrift.FieldID(1): int(1)}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			println(tt.name)
			ret := NewTypedNode(tt.args.typ, tt.args.et, tt.args.kt, tt.args.children...)
			fmt.Printf("buf:%+v\n", ret.Raw())
			val, err := ret.Interface(&tt.opts)
			fmt.Printf("val:%#v\n", val)
			require.NoError(t, err)
			if strings.Contains(tt.name, "map + key") {
				em := tt.wantRet.(map[interface{}]interface{})
				gm := val.(map[interface{}]interface{})
				require.Equal(t, len(em), len(gm))
				var firstK, firstV interface{}
				for k, v := range em {
					firstK, firstV = k, v
					break
				}
				var firstKgot, firstVgot interface{}
				for k, v := range gm {
					firstKgot, firstVgot = k, v
					break
				}
				require.Equal(t, firstK, firstKgot)
				require.Equal(t, firstV, firstVgot)
			} else {
				require.Equal(t, tt.wantRet, val)
			}
		})
	}
}
