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
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func format(s string) string {
	s = strings.ReplaceAll(s, "\t", "  ")
	return s
}

func clean(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\t", "")
	return s
}

func mergeString(intoStr, fromStr string) (string, error) {
	var into, from interface{}
	yaml.NewDecoder(bytes.NewReader([]byte(format(intoStr)))).Decode(&into)
	yaml.NewDecoder(bytes.NewReader([]byte(format(fromStr)))).Decode(&from)
	if r, err := merge(into, from); err == nil {
		buf := &bytes.Buffer{}
		yaml.NewEncoder(buf).Encode(r)
		return buf.String(), nil
	} else {
		return "", err
	}
}

func assertMerge(t *testing.T, into, from, expected string) {
	actual, err := mergeString(into, from)
	assert.Nil(t, err)
	assert.Equal(t, clean(expected), clean(actual))
}

func TestMergeArray(t *testing.T) {
	into := `http:
	- name: name
		url: url`
	from := `http:
	- name: name
		url: url`
	expected := `http:
	- name: name
		url: url
	- name: name
		url: url`
	assertMerge(t, into, from, expected)

	into = `http:
  sub:
    - name: name
      url: url`
	from = `http:
  sub:
    - name: name
      url: url`
	expected = `http:
  sub:
    - name: name
      url: url
    - name: name
      url: url`
	assertMerge(t, into, from, expected)
}

func TestMergeMapping(t *testing.T) {
	into := `key: v1`
	from := `key: v2`
	expected := `key: v2`
	assertMerge(t, into, from, expected)

	into = `key1: v1`
	from = `key2: v2`
	expected = `key1: v1
  key2: v2`
	assertMerge(t, into, from, expected)

	into = `key:
  sub1: 1`
	from = `key:
  sub2: 2`
	expected = `key:
	sub1: 1
	sub2: 2`
	assertMerge(t, into, from, expected)
}

func TestMergePath(t *testing.T) {
	dir := os.TempDir() + "easeprobe_merge_test"
	os.Mkdir(dir, 0755)

	config1 := `http:
  - name: name
		url: url`
	config2 := `http:
	- name: name
		url: url`
	expected := `http:
	- name: name
		url: url
	- name: name
		url: url`
	os.Chdir(dir)
	err := os.WriteFile("config1.yaml", []byte(format(config1)), 0755)
	assert.Nil(t, err)
	err = os.WriteFile("config2.yaml", []byte(format(config2)), 0755)
	assert.Nil(t, err)

	r, err := mergeYamlFiles(dir)
	assert.Nil(t, err)
	assert.Equal(t, clean(expected), clean(string(r)))

	os.RemoveAll(dir)
}
