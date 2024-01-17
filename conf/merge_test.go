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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/op/go-logging.v1"
	"gopkg.in/yaml.v3"
)

// on Windows platform, TempDir RemoveAll cleanup failures with
// "The process cannot access the file because it is being used by another process."
// refer to https://github.com/golang/go/issues/51442
// using the workaround to remove the directory by a retry loop
func removeDir(dir string) {
	for {
		if err := os.RemoveAll(dir); err == nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func assertMerge(t *testing.T, into, from, expected string) {
	i, f, e := decode(into), decode(from), decode(expected)
	assert.NotNil(t, i, fmt.Sprintf("wrong yaml content: %v", into))
	assert.NotNil(t, f, fmt.Sprintf("wrong yaml content: %v", from))
	assert.NotNil(t, e, fmt.Sprintf("wrong yaml content: %v", expected))

	dir := t.TempDir()
	defer removeDir(dir)

	err := os.WriteFile(dir+"/config1.yaml", []byte(into), 0755)
	assert.Nil(t, err)
	err = os.WriteFile(dir+"/config2.yml", []byte(from), 0755)
	assert.Nil(t, err)

	actual, err := mergeYamlFiles(dir)
	assert.Nil(t, err)
	assert.Equal(t, decode(expected), decode(string(actual)))
}

func decode(s string) interface{} {
	var v interface{}
	yaml.NewDecoder(bytes.NewReader([]byte(s))).Decode(&v)
	return v
}

func TestMain(m *testing.M) {
	logging.SetLevel(logging.INFO, "yq-lib")
	exitCode := m.Run()
	os.Exit(exitCode)
}

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

func TestFailed(t *testing.T) {
	_, err := mergeYamlFiles("[]")
	assert.NotNil(t, err)

	dir := t.TempDir()
	defer removeDir(dir)
	err = os.WriteFile(dir+"/config.yaml", []byte("wrong yaml"), 0755)
	assert.Nil(t, err)

	_, err = mergeYamlFiles(dir)
	assert.NotNil(t, err)
}
