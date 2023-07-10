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
	"reflect"
	"testing"
)

func TestSplitTagOptions(t *testing.T) {
	type args struct {
		tag string
	}
	tests := []struct {
		name    string
		args    args
		wantRet []string
		wantErr bool
	}{
		{
			name: "single",
			args: args{
				tag: `json:"k1,omitempty"`,
			},
			wantRet: []string{"json", "k1"},
			wantErr: false,
		},
		{
			name: "multi",
			args: args{
				tag: `json:"k1,omitempty" thrift:"k2,xxx"`,
			},
			wantRet: []string{"json", "k1", "thrift", "k2"},
			wantErr: false,
		},
		{
			name: "quoted",
			args: args{
				tag: `json:\"k1,omitempty\"`,
			},
			wantRet: []string{"json", "k1"},
			wantErr: false,
		},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRet, err := SplitTagOptions(tt.args.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitTagOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRet, tt.wantRet) {
				t.Errorf("SplitTagOptions() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}
