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

package log

import (
	"bufio"
	"os"
	"strings"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/base"

	log "github.com/sirupsen/logrus"
)

// Network protocols
const (
	TCP = "tcp"
	UDP = "udp"
)

// Type is the log type
type Type int

// Log Type
const (
	FileLog = iota
	SysLog
)

// NotifyConfig is the configuration of the Notify
type NotifyConfig struct {
	base.DefaultNotify `yaml:",inline"`

	File    string `yaml:"file"`
	Host    string `yaml:"host"`
	Network string `yaml:"network"`
	Type    Type   `yaml:"-"`
	logger  *log.Logger
}

func (c *NotifyConfig) configLogFile() error {
	c.NotifyKind = "log"
	c.Type = FileLog
	file, err := os.OpenFile(c.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Errorf("[%s / %s] cannot open file: %s", c.Kind(), c.Name(), err)
		return err
	}
	c.logger.SetOutput(file)
	log.Infof("[%s / %s] - local log file(%s) configured", c.Kind(), c.Name(), c.File)
	return nil
}

// Config configures the log notification
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	if err := c.ConfigLog(); err != nil {
		return err
	}
	return c.DefaultNotify.Config(gConf)
}

// Log logs the message
func (c *NotifyConfig) Log(title, msg string) error {
	scanner := bufio.NewScanner(strings.NewReader(msg))
	for scanner.Scan() {
		line := scanner.Text()
		log.Debugf("[%s] %s", c.NotifyKind, line)
		c.logger.Info(line)
	}

	return scanner.Err()
}
