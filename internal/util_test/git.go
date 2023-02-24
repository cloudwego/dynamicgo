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

package util_test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// GetGitRoot gets the git root of this project.
func GetGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	sb := strings.Builder{}
	cmd.Stdout = &sb
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	str := sb.String()
	return str[:len(str)-1], nil
}

// MustGitPath is util for getting relative path to git root
func MustGitPath(relative string) string {
	ret, err := GetGitRoot()
	if err != nil {
		panic(err)
	}
	return ret + "/" + relative
}

// JsonEqual assert two json are equal
func JsonEqual(t *testing.T, expected, actual []byte) {
	var exp, act map[string]interface{}
	err := json.Unmarshal(expected, &exp)
	assert.Nil(t, err)
	err = json.Unmarshal(actual, &act)
	assert.Nil(t, err)
	assert.Equal(t, exp, act)
}
