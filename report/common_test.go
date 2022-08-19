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

package report

import (
	"encoding/json"
	"fmt"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func TestDurationStr(t *testing.T) {
	var expected, result string

	expected = "0s"
	result = DurationStr(0)
	assert.Equal(t, expected, result)

	expected = "10ns"
	result = DurationStr(-10)
	assert.Equal(t, expected, result)

	expected = "10ms"
	result = DurationStr(10 * 1000 * 1000)
	assert.Equal(t, expected, result)

	expected = "10s"
	result = DurationStr(10 * 1000 * 1000 * 1000)
	assert.Equal(t, expected, result)

	expected = "10m0s"
	result = DurationStr(10 * 60 * 1000 * 1000 * 1000)
	assert.Equal(t, expected, result)

	expected = "10h0m0s"
	result = DurationStr(10 * 60 * 60 * 1000 * 1000 * 1000)
	assert.Equal(t, expected, result)

	expected = "10d"
	result = DurationStr(10 * 24 * 60 * 60 * 1000 * 1000 * 1000)
	assert.Equal(t, expected, result)

	expected = "10d10h10m10s"
	result = DurationStr(10*24*60*60*1000*1000*1000 + 10*60*60*1000*1000*1000 + 10*60*1000*1000*1000 + 10*1000*1000*1000)
	assert.Equal(t, expected, result)
}

func TestJSONEscape(t *testing.T) {
	var expected, result string

	expected = `\"\"`
	result = JSONEscape(`""`)
	assert.Equal(t, expected, result)

	expected = `hello`
	result = JSONEscape(`hello`)
	assert.Equal(t, expected, result)

	expected = `hello\\nworld`
	result = JSONEscape(`hello\nworld`)
	assert.Equal(t, expected, result)

	monkey.Patch(json.Marshal, func(v interface{}) ([]byte, error) {
		return nil, fmt.Errorf("error")
	})
	expected = `{hello}`
	result = JSONEscape(`{hello}`)
	assert.Equal(t, expected, result)
	monkey.UnpatchAll()
}

func TestAutoRefreshJS(t *testing.T) {
	result := AutoRefreshJS("1000")
	assert.Contains(t, result, "setInterval('autoRefresh()', 1000);")
}

func TestLogSend(t *testing.T) {
	LogSend("client", "MYSQL Test", "mysql", "Hello World", nil)
	LogSend("client", "MYSQL Test", "mysql", "", fmt.Errorf("error"))
}
