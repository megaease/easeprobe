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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckText(t *testing.T) {
	tc := TextChecker{}

	tc.Contain = "hello"
	tc.NotContain = "bad"
	err := tc.Check("easeprobe hello world")
	assert.Nil(t, err)

	tc.Contain = "hello"
	tc.NotContain = "world"
	err = tc.Check("easeprobe hello world")
	assert.NotNil(t, err)

	tc.Contain = ""
	tc.NotContain = "world"
	err = tc.Check("easeprobe hello world")
	assert.NotNil(t, err)

	tc.Contain = "hello"
	tc.NotContain = ""
	err = tc.Check("easeprobe hello world")
	assert.Nil(t, err)

	tc.Contain = "good"
	tc.NotContain = ""
	err = tc.Check("easeprobe hello world")
	assert.NotNil(t, err)

	tc.Contain = ""
	tc.NotContain = "bad"
	err = tc.Check("easeprobe hello world")
	assert.Nil(t, err)

	tc.Contain = "good"
	tc.NotContain = "bad"
	err = tc.Check("easeprobe hello world")
	assert.NotNil(t, err)
}

func testRegExpHelper(t *testing.T, regExp string, str string, match bool) {
	tc := TextChecker{RegExp: true}
	tc.Contain = regExp
	tc.Config()
	if match {
		assert.Nil(t, tc.CheckRegExp(str))
	} else {
		assert.NotNil(t, tc.CheckRegExp(str))
	}

	tc.Contain = ""
	tc.NotContain = regExp
	tc.Config()
	if match {
		assert.NotNil(t, tc.CheckRegExp(str))
	} else {
		assert.Nil(t, tc.CheckRegExp(str))
	}
}

func TestCheckRegExp(t *testing.T) {

	word := `word[0-9]+`
	testRegExpHelper(t, word, "word word10 word", true)
	testRegExpHelper(t, word, "word word word", false)

	time := "[0-9]?[0-9]:[0-9][0-9]"
	testRegExpHelper(t, time, "easeprobe hello world 12:34", true)
	testRegExpHelper(t, time, "easeprobe hello world 1234", false)

	html := `<\/?[\w\s]*>|<.+[\W]>`
	testRegExpHelper(t, html, "<p>test hello world </p>", true)
	testRegExpHelper(t, html, "test hello world", false)

	or := `word1|word2`
	testRegExpHelper(t, or, "word1 easeprobe word2", true)
	testRegExpHelper(t, or, "word2 easeprobe word1", true)
	testRegExpHelper(t, or, "word3 easeprobe word1", true)
	testRegExpHelper(t, or, "word2 easeprobe word3", true)
	testRegExpHelper(t, or, "word easeprobe word3", false)
	testRegExpHelper(t, or, "word easeprobe hello world", false)

	unsupported := "(?=.*word1)(?=.*word2)"
	tc := TextChecker{RegExp: true}
	tc.Contain = unsupported
	err := tc.Config()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid or unsupported Perl syntax")

	tc.Contain = ""
	tc.NotContain = unsupported
	err = tc.Config()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid or unsupported Perl syntax")

}

func TestTextChecker(t *testing.T) {
	checker := TextChecker{
		Contain:    "hello",
		NotContain: "",
		RegExp:     false,
	}
	checker.Config()
	assert.Nil(t, checker.Check("hello world"))
	assert.Contains(t, checker.String(), "Text Mode")

	checker = TextChecker{
		Contain:    "[0-9]+$",
		NotContain: "",
		RegExp:     true,
	}
	checker.Config()
	assert.Nil(t, checker.Check("hello world 2022"))
	assert.Contains(t, checker.String(), "RegExp Mode")

	checker = TextChecker{
		Contain:    "",
		NotContain: `<\/?[\w\s]*>|<.+[\W]>`,
		RegExp:     true,
	}
	checker.Config()
	assert.NotNil(t, checker.Check("<p>test hello world </p>"))
}

func TestCheckEmpty(t *testing.T) {
	assert.Equal(t, "a", CheckEmpty("a"))
	assert.Equal(t, "empty", CheckEmpty("    "))
	assert.Equal(t, "empty", CheckEmpty("  \t"))
	assert.Equal(t, "empty", CheckEmpty("\n\r\t"))
	assert.Equal(t, "empty", CheckEmpty("  \n\r\t  "))
}
