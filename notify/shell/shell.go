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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/base"
	"github.com/megaease/easeprobe/report"

	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the config for shell notify
type NotifyConfig struct {
	base.DefaultNotify `yaml:",inline"`

	Cmd  string   `yaml:"cmd" json:"cmd" jsonschema:"required,title=Command,description=the command to run"`
	Args []string `yaml:"args,omitempty" json:"args,omitempty" jsonschema:"title=Arguments,description=the arguments for command"`
	Env  []string `yaml:"env,omitempty" json:"env,omitempty" jsonschema:"title=Environment variables,description=the environment variables for command"`

	CleanEnv bool `yaml:"clean_env,omitempty" json:"clean_env,omitempty" jsonschema:"title=Clean Environment Variables,description=set it to true to keep the environment variables of the current process"`
}

// Config is the config for shell probe
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.NotifyKind = "shell"
	c.NotifyFormat = report.Shell
	c.NotifySendFunc = c.RunShell
	c.DefaultNotify.Config(gConf)

	return nil
}

// RunShell is the shell for shell notify
func (c *NotifyConfig) RunShell(title, msg string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	var envMap map[string]string
	err := json.Unmarshal([]byte(msg), &envMap)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, c.Cmd, c.Args...)
	var env []string
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Stdin = strings.NewReader(envMap["EASEPROBE_CSV"])
	if c.CleanEnv == false {
		cmd.Env = append(os.Environ(), env...)
	} else {
		log.Infof("[%s / %s] clean the environment variables", c.NotifyKind, c.NotifyName)
		cmd.Env = env
	}
	cmd.Env = append(cmd.Env, c.Env...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	log.Debugf("[%s / %s] - %s", c.NotifyKind, c.NotifyName, global.CommandLine(c.Cmd, c.Args))
	log.Debugf("input: \n%s", msg)
	log.Debugf("output:\n%s", string(output))
	return nil
}
