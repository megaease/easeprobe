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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestSectionMerge(t *testing.T) {
	into := `
http:
  - name: "test 1"
    url: "http://localhost:8080"
    method: "GET"
tcp:
  - name: "test 2"
    host: "localhost:8080"`
	from := `
notify:
  slack:
    - name: "slack"
      webhook: "https://hooks.slack.com/services/xxxxxx"`
	expected := `
http:
  - name: "test 1"
    url: "http://localhost:8080"
    method: "GET"
tcp:
  - name: "test 2"
    host: "localhost:8080"
notify:
  slack:
    - name: "slack"
      webhook: "https://hooks.slack.com/services/xxxxxx"`

	assertMerge(t, into, from, expected)
}

func TestSameProbeMerge(t *testing.T) {
	into := `
http:
  - name: "test 1"
    url: "http://localhost:8080"
    method: "GET"`
	from := `
http:
  - name: "test 2"
    url: "http://localhost:8181"
    method: "GET"`
	expected := `
http:
  - name: "test 1"
    url: "http://localhost:8080"
    method: "GET"
  - name: "test 2"
    url: "http://localhost:8181"
    method: "GET"`

	assertMerge(t, into, from, expected)
}

func TestNotifyMerge(t *testing.T) {
	into := `
notify:
  slack:
    - name: slack
      webhook: "https://hooks.slack.com/services/xxxxxx"`
	from := `
notify:
  discord:
    - name: discord
      webhook: "https://discord.com/api/webhooks/xxxxxx"`
	expected := `
notify:
  slack:
    - name: slack
      webhook: "https://hooks.slack.com/services/xxxxxx"
  discord:
    - name: discord
      webhook: "https://discord.com/api/webhooks/xxxxxx"`

	assertMerge(t, into, from, expected)
}

func TestNotifyArrayMerge(t *testing.T) {
	into := `
notify:
  slack:
    - name: slack1
      webhook: "https://hooks.slack.com/services/xxxxxx"`
	from := `
notify:
  slack:
    - name: slack2
      webhook: "https://hooks.slack.com/services/xxxxxx"`
	expected := `
notify:
  slack:
    - name: slack1
      webhook: "https://hooks.slack.com/services/xxxxxx"
    - name: slack2
      webhook: "https://hooks.slack.com/services/xxxxxx"`

	assertMerge(t, into, from, expected)
}

func TestSettingsMerge(t *testing.T) {
	into := `
settings:
  name: easeprobe
  sla:
    schedule: "daily"
    time: "00:00"
  notify:
    retry:
      times: 5
      interval: 10`
	from := `
settings:
  name: easeprobe_from
  probe:
    timeout: 10s
    interval: 30s
  sla:
    schedule: "weekly"`
	expected := `
settings:
  name: easeprobe_from
  probe:
    timeout: 10s
    interval: 30s
  notify:
    retry:
      times: 5
      interval: 10
  sla:
    schedule: "weekly"
    time: "00:00"`

	assertMerge(t, into, from, expected)
}

func TestPathMerge(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "easeprobe_merge_test")
	os.Mkdir(dir, 0755)

	config1 := `
http:
  - name: name
    url: url`
	config2 := `
http:
  - name: name
    url: url`
	expected := `
http:
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
	assert.Equal(t, decode(expected), decode(string(r)))

	os.RemoveAll(dir)
}

func assertMerge(t *testing.T, into, from, expected string) {
	i, f, e := decode(into), decode(from), decode(expected)
	assert.NotNil(t, i, fmt.Sprintf("wrong yaml content: %v", into))
	assert.NotNil(t, f, fmt.Sprintf("wrong yaml content: %v", from))
	assert.NotNil(t, e, fmt.Sprintf("wrong yaml content: %v", expected))

	actual, err := merge(i, f)
	assert.Nil(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, e, actual)
}

func decode(s string) interface{} {
	var v interface{}
	yaml.NewDecoder(bytes.NewReader([]byte(format(s)))).Decode(&v)
	return v
}

func format(s string) string {
	return strings.ReplaceAll(s, "\t", "  ")
}
