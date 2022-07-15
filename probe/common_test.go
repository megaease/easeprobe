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

func TestCheckOutput(t *testing.T) {
	err := CheckOutput("hello", "good", "easeprobe hello world")
	assert.Nil(t, err)

	err = CheckOutput("hello", "world", "easeprobe hello world")
	assert.NotNil(t, err)

	err = CheckOutput("hello", "world", "easeprobe hello world")
	assert.NotNil(t, err)

	err = CheckOutput("good", "bad", "easeprobe hello world")
	assert.NotNil(t, err)
}

func TestCheckOutputRegExp(t *testing.T) {
	reg := `word[0-9]+`
	err := CheckOutputRegExp(reg, "", "word word10 word")
	assert.Nil(t, err)
	err = CheckOutputRegExp("", reg, "word word word")
	assert.Nil(t, err)

	time := "[0-9]?[0-9]:[0-9][0-9]"
	err = CheckOutputRegExp(time, "", "easeprobe hello world 1234")
	assert.NotNil(t, err)
	err = CheckOutputRegExp(time, "", "easeprobe hello world 12:34")
	assert.Nil(t, err)

	html := `<\/?[\w\s]*>|<.+[\W]>`
	err = CheckOutputRegExp(html, "", "<p>test hello world </p>")
	assert.Nil(t, err)
	err = CheckOutputRegExp("hello", html, "text test hello world")
	assert.Nil(t, err)

	or := `word1|word2`
	err = CheckOutputRegExp(or, "", "word1 easeprobe word2")
	assert.Nil(t, err)
	err = CheckOutputRegExp(or, "", "word2 easeprobe word1")
	assert.Nil(t, err)
	err = CheckOutputRegExp("", or, " easeprobe word1")
	assert.NotNil(t, err)

	unsupported := "(?=.*word1)(?=.*word2)"
	err = CheckOutputRegExp(unsupported, "", "word1 word2")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid or unsupported Perl syntax")
	err = CheckOutputRegExp("", unsupported, "word1 word2")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid or unsupported Perl syntax")

}

func TestTextChecker(t *testing.T) {
	checker := TextChecker{
		Contain:    "hello",
		NotContain: "",
		RegExp:     false,
	}
	assert.Nil(t, checker.Check("hello world"))
	assert.Contains(t, checker.String(), "Text Mode")

	checker = TextChecker{
		Contain:    "[0-9]+$",
		NotContain: "",
		RegExp:     true,
	}
	assert.Nil(t, checker.Check("hello world 2022"))
	assert.Contains(t, checker.String(), "RegExp Mode")

	checker = TextChecker{
		Contain:    "",
		NotContain: `<\/?[\w\s]*>|<.+[\W]>`,
		RegExp:     true,
	}
	assert.NotNil(t, checker.Check("<p>test hello world </p>"))
}

func TestCheckEmpty(t *testing.T) {
	assert.Equal(t, "a", CheckEmpty("a"))
	assert.Equal(t, "empty", CheckEmpty("    "))
	assert.Equal(t, "empty", CheckEmpty("  \t"))
	assert.Equal(t, "empty", CheckEmpty("\n\r\t"))
	assert.Equal(t, "empty", CheckEmpty("  \n\r\t  "))
}
