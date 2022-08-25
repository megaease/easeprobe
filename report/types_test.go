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
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func testFormat(t *testing.T, except Format, str string) {
	var result Format
	result.Format(str)
	assert.Equal(t, except, result)

	s := result.String()
	assert.Equal(t, str, s)
}

func TestFormat(t *testing.T) {
	testFormat(t, MarkdownSocial, "markdown-social")
	testFormat(t, Markdown, "markdown")
	testFormat(t, HTML, "html")
	testFormat(t, JSON, "json")
	testFormat(t, Text, "text")
	testFormat(t, Slack, "slack")
	testFormat(t, Discord, "discord")
	testFormat(t, Lark, "lark")
	testFormat(t, SMS, "sms")
	testFormat(t, Unknown, "unknown")
}

func testFormatYAML(t *testing.T, except Format, str string) {
	var result Format
	buf, err := yaml.Marshal(except)
	assert.Nil(t, err)
	assert.Equal(t, str, string(buf))

	assert.Nil(t, yaml.Unmarshal(buf, &result))
	assert.Equal(t, except, result)
}

func TestFormatYAML(t *testing.T) {
	testFormatYAML(t, MarkdownSocial, "markdown-social\n")
	testFormatYAML(t, Markdown, "markdown\n")
	testFormatYAML(t, HTML, "html\n")
	testFormatYAML(t, JSON, "json\n")
	testFormatYAML(t, Text, "text\n")
	testFormatYAML(t, Slack, "slack\n")
	testFormatYAML(t, Discord, "discord\n")
	testFormatYAML(t, Lark, "lark\n")
	testFormatYAML(t, SMS, "sms\n")
	testFormatYAML(t, Unknown, "unknown\n")
}

func TestBadFormat(t *testing.T) {
	buf := ([]byte)("- asdf::")
	var f Format
	err := yaml.Unmarshal(buf, &f)
	assert.NotNil(t, err)
}
