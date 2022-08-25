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

package conf

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func testMarshalUnmarshal(t *testing.T, str string, pt ProviderType, good bool,
	marshal func(in interface{}) ([]byte, error),
	unmarshal func(in []byte, out interface{}) (err error)) {

	var s ProviderType
	err := unmarshal([]byte(str), &s)
	if good {
		assert.Nil(t, err)
		assert.Equal(t, pt, s)
	} else {
		assert.Error(t, err)
		assert.Equal(t, Unknown, s)
	}

	buf, err := marshal(pt)
	if good {
		assert.Nil(t, err)
		assert.Equal(t, str, string(buf))
	} else {
		assert.Error(t, err)
		assert.Nil(t, buf)
	}
}
func testYamlJSON(t *testing.T, str string, pt ProviderType, good bool) {
	testYaml(t, str+"\n", pt, good)
	testJSON(t, `"`+str+`"`, pt, good)
}
func testYaml(t *testing.T, str string, pt ProviderType, good bool) {
	testMarshalUnmarshal(t, str, pt, good, yaml.Marshal, yaml.Unmarshal)
}
func testJSON(t *testing.T, str string, pt ProviderType, good bool) {
	testMarshalUnmarshal(t, str, pt, good, json.Marshal, json.Unmarshal)
}

func TestSMSProviderType(t *testing.T) {
	testYamlJSON(t, "yunpian", Yunpian, true)
	testYamlJSON(t, "twilio", Twilio, true)
	testYamlJSON(t, "nexmo", Nexmo, true)
	testYamlJSON(t, "unknown", Unknown, true)
	testYamlJSON(t, "bad", 10, false)

	testYaml(t, "- xxx", 10, false)
	testJSON(t, `{"x":"y"}`, 10, false)

	p := ProviderType(10)
	assert.Equal(t, "unknown", p.String())
	p = Yunpian
	assert.Equal(t, "yunpian", p.String())

	p = p.ProviderType("bad")
	assert.Equal(t, Unknown, p)
	p = p.ProviderType("nexmo")
	assert.Equal(t, Nexmo, p)
}
