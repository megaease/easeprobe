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

package shell

import (
	"context"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/stretchr/testify/assert"
)

func createShell() *Shell {
	return &Shell{
		DefaultProbe: base.DefaultProbe{ProbeName: "dummy shell"},
		Command:      "dummy command",
		Args:         []string{"arg1", "arg2"},
		Env:          []string{"env1=value1", "env2=value2"},
		TextChecker: probe.TextChecker{
			Contain:    "good",
			NotContain: "bad",
		},
	}
}

func TestTextCheckerConfig(t *testing.T) {
	s := createShell()
	s.TextChecker = probe.TextChecker{
		Contain:    "",
		NotContain: "",
		RegExp:     true,
	}

	err := s.Config(global.ProbeSettings{})
	assert.NoError(t, err)

	s.Contain = `[a-zA-z]\d+`
	err = s.Config(global.ProbeSettings{})
	assert.NoError(t, err)
	assert.Equal(t, `[a-zA-z]\d+`, s.TextChecker.Contain)

	s.NotContain = `(?=.*word1)(?=.*word2)`
	err = s.Config(global.ProbeSettings{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or unsupported Perl syntax")
}

func TestShell(t *testing.T) {
	s := createShell()
	s.Config(global.ProbeSettings{})
	assert.Equal(t, "shell", s.ProbeKind)

	monkey.Patch(exec.CommandContext, func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return &exec.Cmd{}
	})
	var cmd *exec.Cmd
	monkey.PatchInstanceMethod(reflect.TypeOf(cmd), "CombinedOutput", func(_ *exec.Cmd) ([]byte, error) {
		return []byte("good"), nil
	})

	status, message := s.DoProbe()
	assert.True(t, status)
	assert.Contains(t, message, "Successfully")

	// contains the bad output
	s.Contain = ""
	s.NotContain = "bad"
	monkey.PatchInstanceMethod(reflect.TypeOf(cmd), "CombinedOutput", func(_ *exec.Cmd) ([]byte, error) {
		return []byte("bad"), nil
	})
	status, message = s.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "bad")

	// run command error
	monkey.UnpatchAll()
	//no exit code
	status, message = s.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "ExitCode(null)")

	s.Command = "sh"
	s.Args = []string{"-c", "notfound"}
	status, message = s.DoProbe()
	assert.False(t, status)
	assert.NotContains(t, message, "ExitCode(null)")
}

func TestEnv(t *testing.T) {
	s := &Shell{
		DefaultProbe: base.DefaultProbe{ProbeName: "dummy shell"},
		Command:      "env",
		Args:         []string{},
		Env:          []string{},
	}

	s.Config(global.ProbeSettings{})

	err := os.Setenv("EASEPROBE", "1")
	assert.Nil(t, err)
	s.Contain = "EASEPROBE=1"
	status, message := s.DoProbe()
	assert.True(t, status)
	assert.Contains(t, message, "Successfully")

	s.CleanEnv = true
	s.Env = []string{"env1=value1"}
	s.Contain = "env1=value1"
	s.NotContain = "EASEPROBE=1"
	status, message = s.DoProbe()
	assert.True(t, status)
	assert.Contains(t, message, "Successfully")
}
