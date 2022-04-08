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
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/base"
	log "github.com/sirupsen/logrus"
)

// Shell implements a config for shell command (os.Exec)
type Shell struct {
	Name       string   `yaml:"name"`
	Command    string   `yaml:"cmd"`
	Args       []string `yaml:"args,omitempty"`
	Env        []string `yaml:"env,omitempty"`
	Contain    string   `yaml:"contain,omitempty"`
	NotContain string   `yaml:"not_contain,omitempty"`

	base.DefaultOptions `yaml:",inline"`
}

// Config Shell Config Object
func (s *Shell) Config(gConf global.ProbeSettings) error {
	s.ProbeKind = "shell"
	s.DefaultOptions.Config(gConf, s.Name, s.CommandLine())

	log.Debugf("[%s] configuration: %+v, %+v", s.Kind(), s, s.Result())
	return nil
}

// Probe return the checking result
func (s *Shell) Probe() probe.Result {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, e := range s.Env {
		v := strings.Split(e, "=")
		os.Setenv(v[0], v[1])
	}

	now := time.Now()
	s.Result().StartTime = now
	s.Result().StartTimestamp = now.UnixMilli()

	cmd := exec.CommandContext(ctx, s.Command, s.Args...)
	output, err := cmd.CombinedOutput()

	s.Result().RoundTripTime.Duration = time.Since(now)

	outputFmt := func(output []byte) string {
		s := string(output)
		if len(strings.TrimSpace(s)) <= 0 {
			return "empty"
		}
		return s
	}

	status := probe.StatusUp
	s.ProbeResult.Message = "Shell Command has been Run Successfully!"

	if err != nil {
		exitCode := 0
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}

		s.ProbeResult.Message = fmt.Sprintf("Error: %v, ExitCode(%d), Output:%s", err, exitCode, outputFmt(output))
		log.Errorf(s.ProbeResult.Message)
		status = probe.StatusDown
	}
	log.Debugf("[%s] - %s", s.Kind(), s.CommandLine())
	log.Debugf("[%s] - %s", s.Kind(), outputFmt(output))

	if err := s.CheckOutput(output); err != nil {
		log.Errorf("[%s] - %v", s.Kind(), err)
		s.ProbeResult.Message = fmt.Sprintf("Error: %v", err)
		status = probe.StatusDown
	}

	s.ProbeResult.PreStatus = s.ProbeResult.Status
	s.ProbeResult.Status = status

	s.ProbeResult.DoStat(s.Interval())
	return *s.ProbeResult
}

// CheckOutput checks the output text,
// - if it contains a configured string then return nil
// - if it does not contain a configured string then return nil
func (s *Shell) CheckOutput(output []byte) error {

	str := string(output)
	if len(s.Contain) > 0 && !strings.Contains(str, s.Contain) {

		return fmt.Errorf("the output does not contain [%s]", s.Contain)
	}

	if len(s.NotContain) > 0 && strings.Contains(str, s.NotContain) {
		return fmt.Errorf("the output contains [%s]", s.NotContain)

	}
	return nil
}

// CommandLine will return the whole command line which includes command and all arguments
func (s *Shell) CommandLine() string {
	result := s.Command
	for _, arg := range s.Args {
		result += " " + arg
	}
	return result
}
