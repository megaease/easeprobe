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

func TestCommandLine(t *testing.T) {
	s := CommandLine("echo", []string{"hello", "world"})
	assert.Equal(t, "echo hello world", s)

	s = CommandLine("kubectl", []string{"get", "pod", "--all-namespaces", "-o", "json"})
	assert.Equal(t, "kubectl get pod --all-namespaces -o json", s)
}

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

func TestCheckEmpty(t *testing.T) {
	assert.Equal(t, "a", CheckEmpty("a"))
	assert.Equal(t, "empty", CheckEmpty("    "))
	assert.Equal(t, "empty", CheckEmpty("  \t"))
	assert.Equal(t, "empty", CheckEmpty("\n\r\t"))
	assert.Equal(t, "empty", CheckEmpty("  \n\r\t  "))
}
