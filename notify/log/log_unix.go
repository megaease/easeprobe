//go:build !windows

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
	"fmt"
	"log/syslog"
	"net"
	"strconv"
	"strings"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/report"

	log "github.com/sirupsen/logrus"
)

const (
	syslogIdentifier = "syslog"
)

func (c *NotifyConfig) checkNetworkProtocol() error {
	if strings.TrimSpace(c.Network) == "" {
		return fmt.Errorf("protocol is required")
	}
	if strings.TrimSpace(c.Host) == "" {
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

// IsSyslog returns true if the log is syslog
func (c *NotifyConfig) IsSyslog() bool {
	return strings.TrimSpace(c.File) == syslogIdentifier
}

// HasNetwork returns true if the log has network configuration
func (c *NotifyConfig) HasNetwork() (bool, error) {
	// if is not syslog, then return false
	if c.IsSyslog() == false {
		return false, nil
	}
	// if is syslog, but not configured network, then return false
	if strings.TrimSpace(c.Network) == "" || strings.TrimSpace(c.Host) == "" {
		return false, nil
	}
	// if the network is configured error, then return false
	if err := c.checkNetworkProtocol(); err != nil {
		return false, err
	}
	return true, nil
}

// ConfigLog configures the log
// Unix platform support syslog and log file notification
func (c *NotifyConfig) ConfigLog() error {
	c.NotifyKind = "log"
	c.NotifyFormat = report.Log
	c.NotifySendFunc = c.Log

	c.logger = log.New()

	isSyslog := c.IsSyslog()
	hasNetwork, err := c.HasNetwork()
	if err != nil {
		return err
	}
	// syslog && network configuration error
	if isSyslog == true && hasNetwork == true { // remote syslog
		c.NotifyKind = syslogIdentifier
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
	} else if isSyslog == true { // only for local syslog
		c.NotifyKind = syslogIdentifier
		c.Type = SysLog
		writer, err := syslog.New(syslog.LOG_NOTICE, global.GetEaseProbe().Name)
		if err != nil {
			log.Errorf("[%s] cannot open syslog: %s", c.NotifyKind, err)
			return err
		}
		c.logger.SetOutput(writer)
		log.Infof("[%s] %s - local syslog configured!", c.NotifyKind, c.NotifyName)
	} else { // just log file
		if err := c.configLogFile(); err != nil {
			return err
		}
	}
	c.logger.SetFormatter(&SysLogFormatter{
		Type: c.Type,
	})

	return nil
}
