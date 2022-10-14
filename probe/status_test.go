/*
 * Copyright (c) 2022, MegaEase
 * All rights reserved.
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

package probe

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func testMarshalUnmarshal(t *testing.T, str string, status Status, good bool,
	marshal func(in interface{}) ([]byte, error),
	unmarshal func(in []byte, out interface{}) (err error)) {

	var s Status
	err := unmarshal([]byte(str), &s)
	if good {
		assert.Nil(t, err)
		assert.Equal(t, status, s)
	} else {
		assert.Error(t, err)
		assert.Equal(t, StatusUnknown, s)
	}

	buf, err := marshal(status)
	if good {
		assert.Nil(t, err)
		assert.Equal(t, str, string(buf))
	} else {
		assert.Error(t, err)
		assert.Nil(t, buf)
	}
}

func testYamlJSON(t *testing.T, str string, status Status, good bool) {
	testYaml(t, str+"\n", status, good)
	testJSON(t, `"`+str+`"`, status, good)
}
func testYaml(t *testing.T, str string, status Status, good bool) {
	testMarshalUnmarshal(t, str, status, good, yaml.Marshal, yaml.Unmarshal)
}
func testJSON(t *testing.T, str string, status Status, good bool) {
	testMarshalUnmarshal(t, str, status, good, json.Marshal, json.Unmarshal)
}

func TestYamlJSON(t *testing.T) {
	testYamlJSON(t, "init", StatusInit, true)
	testYamlJSON(t, "up", StatusUp, true)
	testYamlJSON(t, "down", StatusDown, true)
	testYamlJSON(t, "unknown", StatusUnknown, true)
	testYamlJSON(t, "bad", StatusBad, true)

	testYamlJSON(t, "xxx", 10, false)

	testYaml(t, "- xxx", 10, false)
	testJSON(t, `{"x":"y"}`, 10, false)
}

func TestStatus(t *testing.T) {
	s := StatusUp
	assert.Equal(t, "up", s.String())
	s.Status("down")
	assert.Equal(t, StatusDown, s)
	assert.Equal(t, "❌", s.Emoji())
	s.Status("up")
	assert.Equal(t, StatusUp, s)
	assert.Equal(t, "✅", s.Emoji())
	s.Status("xxx")
	assert.Equal(t, StatusUnknown, s)
	s = 10
	assert.Equal(t, "⛔️", s.Emoji())

	err := yaml.Unmarshal([]byte("down"), &s)
	assert.Nil(t, err)
	assert.Equal(t, StatusDown, s)

	buf, err := yaml.Marshal(&s)
	assert.Nil(t, err)
	assert.Equal(t, "down\n", string(buf))

	buf, err = json.Marshal(s)
	assert.Nil(t, err)
	assert.Equal(t, "\"down\"", string(buf))

	err = yaml.Unmarshal([]byte("xxx"), &s)
	assert.Error(t, err)
	assert.Equal(t, StatusUnknown, s)

	err = yaml.Unmarshal([]byte{1, 2}, &s)
	assert.NotNil(t, err)
}

func TestStatusTitle(t *testing.T) {
	s := StatusInit
	assert.Equal(t, "Initialization", s.Title())

	s = StatusUp
	assert.Equal(t, "Success", s.Title())

	s = StatusDown
	assert.Equal(t, "Error", s.Title())

	s = StatusUnknown
	assert.Equal(t, "Unknown", s.Title())

	s = StatusBad
	assert.Equal(t, "Bad", s.Title())

	s = -1
	assert.Equal(t, "Unknown", s.Title())
}
