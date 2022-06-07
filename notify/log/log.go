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
	"fmt"
	"log/syslog"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/base"
	"github.com/megaease/easeprobe/report"
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

func (c *NotifyConfig) checkNetworkProtocol() error {
	if len(c.Network) == 0 {
		return fmt.Errorf("protocol is required")
	}
	if len(c.Host) == 0 {
		return fmt.Errorf("host is required")
	}
	if c.Network != TCP && c.Network != UDP {
		return fmt.Errorf("[%s] invalid protocol: %s", c.NotifyKind, c.Network)
	}
	_, port, err := net.SplitHostPort(c.Host)
	if err != nil {
		return fmt.Errorf("[%s] invalid host: %s", c.NotifyKind, c.Host)
	}
	_, err = strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("[%s] invalid port: %s", c.NotifyKind, port)
	}
	return nil
}

// Config configures the log files
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.NotifyKind = "log"
	c.NotifyFormat = report.Log
	c.NotifySendFunc = c.Log

	c.logger = log.New()

	isSyslog := (strings.TrimSpace(c.File) == "syslog")
	hasNetwork := false
	if isSyslog == true {
		if len(c.Network) == 0 || len(c.Host) == 0 {
			hasNetwork = false
		} else {
			err := c.checkNetworkProtocol()
			if err != nil {
				return err
			}
			hasNetwork = true
		}
	}

	// sysylog && network configuration error
	if isSyslog && hasNetwork == true { // remote syslog
		c.NotifyKind = "syslog"
		c.Type = SysLog
		if err := c.checkNetworkProtocol(); err != nil {
			return err
		}
		writer, err := syslog.Dial(c.Network, c.Host, syslog.LOG_NOTICE, global.GetEaseProbe().Name)
		if err != nil {
			log.Errorf("[%s] cannot dial syslog network: %s", c.NotifyKind, err)
			return err
		}
		c.logger.SetOutput(writer)
		log.Infof("[%s] %s - remote syslog (%s:%s) configured", c.NotifyKind, c.NotifyName, c.Network, c.Host)
	} else if isSyslog { // only for local syslog
		c.NotifyKind = "syslog"
		c.Type = SysLog
		writer, err := syslog.New(syslog.LOG_NOTICE, global.GetEaseProbe().Name)
		if err != nil {
			log.Errorf("[%s] cannot open syslog: %s", c.NotifyKind, err)
			return err
		}
		c.logger.SetOutput(writer)
		log.Info("[%s] %s - local syslog configured", c.NotifyKind, c.NotifyName)
	} else { // just log file
		c.NotifyKind = "log"
		c.Type = FileLog
		file, err := os.OpenFile(c.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Errorf("[%s] cannot open file: %s", c.NotifyKind, err)
			return err
		}
		c.logger.SetOutput(file)
		log.Infof("[%s] %s - local log file(%s) configured", c.NotifyKind, c.NotifyName, c.File)
	}
	c.logger.SetFormatter(&SysLogFormatter{
		Type: c.Type,
	})
	c.DefaultNotify.Config(gConf)
	return nil
}

// Log logs the message
func (c *NotifyConfig) Log(title, msg string) error {
	scanner := bufio.NewScanner(strings.NewReader(msg))
	for scanner.Scan() {
		line := scanner.Text()
		log.Debugf("[%s] %s", c.NotifyKind, line)
		c.logger.Info(line)
	}

	return nil
}
