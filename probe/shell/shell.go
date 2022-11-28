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

// Package shell is the shell probe package
package shell

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Shell implements a config for shell command (os.Exec)
type Shell struct {
	base.DefaultProbe `yaml:",inline"`
	Command           string   `yaml:"cmd" json:"cmd" jsonschema:"required,title=Command Line,description=Command Line"`
	Args              []string `yaml:"args,omitempty" json:"args,omitempty" jsonschema:"title=Command Line Arguments,description=Command Line Arguments"`
	Env               []string `yaml:"env,omitempty" json:"env,omitempty" jsonschema:"title=Environment Variables,description=Environment Variables,example=[\"PATH=/usr/local/bin\"]"`
	CleanEnv          bool     `yaml:"clean_env,omitempty" json:"clean_env,omitempty" jsonschema:"title=Clean Environment,description=set it to true to keep the environment variables of the current process"`

	// Output Text Checker
	probe.TextChecker `yaml:",inline"`

	exitCode  int `yaml:"-" json:"-"`
	outputLen int `yaml:"-" json:"-"`

	metrics *metrics `yaml:"-" json:"-"`
}

// Config Shell Config Object
func (s *Shell) Config(gConf global.ProbeSettings) error {
	kind := "shell"
	tag := ""
	name := s.ProbeName
	s.DefaultProbe.Config(gConf, kind, tag, name,
		global.CommandLine(s.Command, s.Args), s.DoProbe)

	if err := s.TextChecker.Config(); err != nil {
		return err
	}

	s.metrics = newMetrics(kind, tag)

	log.Debugf("[%s / %s] configuration: %+v", s.ProbeKind, s.ProbeName, *s)
	return nil
}

// DoProbe return the checking result
func (s *Shell) DoProbe() (bool, string) {

	ctx, cancel := context.WithTimeout(context.Background(), s.ProbeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.Command, s.Args...)
	if s.CleanEnv == false {
		cmd.Env = append(os.Environ(), s.Env...)
	} else {
		log.Infof("[%s / %s] clean the environment variables", s.ProbeKind, s.ProbeName)
		cmd.Env = s.Env
	}
	output, err := cmd.CombinedOutput()

	status := true
	message := "Shell Command has been Run Successfully!"

	s.exitCode = 0
	s.outputLen = len(output)
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			s.exitCode = exitError.ExitCode()
			message = fmt.Sprintf("Error: %v, ExitCode(%d), Output:%s",
				err, s.exitCode, probe.CheckEmpty(string(output)))
		} else {
			message = fmt.Sprintf("Error: %v, ExitCode(null), Output:%s",
				err, probe.CheckEmpty(string(output)))
		}
		log.Errorf("[%s / %s] - %s", s.ProbeKind, s.ProbeName, message)
		status = false
	}
	log.Debugf("[%s / %s] - %s", s.ProbeKind, s.ProbeName, global.CommandLine(s.Command, s.Args))
	log.Debugf("[%s / %s] - %s", s.ProbeKind, s.ProbeName, probe.CheckEmpty(string(output)))

	s.ExportMetrics()

	log.Debugf("[%s / %s] - %s", s.ProbeKind, s.ProbeName, s.TextChecker.String())
	if err := s.Check(string(output)); err != nil {
		log.Errorf("[%s / %s] - %v", s.ProbeKind, s.ProbeName, err)
		message = fmt.Sprintf("Error: %v", err)
		status = false
	}

	return status, message
}

// ExportMetrics export shell metrics
func (s *Shell) ExportMetrics() {
	s.metrics.ExitCode.With(prometheus.Labels{
		"name": s.ProbeName,
		"exit": fmt.Sprintf("%d", s.exitCode),
	}).Inc()

	s.metrics.OutputLen.With(prometheus.Labels{
		"name": s.ProbeName,
		"exit": fmt.Sprintf("%d", s.exitCode),
	}).Set(float64(s.outputLen))
}
