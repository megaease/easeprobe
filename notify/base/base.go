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

package base

import (
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/report"
	log "github.com/sirupsen/logrus"
)

// DefaultNotify is the base struct of the Notify
type DefaultNotify struct {
	MyKind   string                     `yaml:"-"`
	Format   report.Format              `yaml:"-"`
	SendFunc func(string, string) error `yaml:"-"`

	Name    string        `yaml:"name"`
	Dry     bool          `yaml:"dry"`
	Timeout time.Duration `yaml:"timeout"`
	Retry   global.Retry  `yaml:"retry"`
}

// Config is the default configuration for notification
func (c *DefaultNotify) Config(gConf global.NotifySettings) error {
	mode := "Live"
	if c.Dry {
		mode = "Dry"
	}
	log.Infof("Notification [%s] - [%s] is running on %s mode!", c.MyKind, c.Name, mode)
	c.Timeout = gConf.NormalizeTimeOut(c.Timeout)
	c.Retry = gConf.NormalizeRetry(c.Retry)

	log.Infof("Notification [%s] - [%s] is configured!", c.MyKind, c.Name)
	return nil
}

// Notify send the result message to the email
func (c *DefaultNotify) Notify(result probe.Result) {
	if c.Dry {
		c.DryNotify(result)
		return
	}
	title := result.Title()
	message := report.FormatFuncs[c.Format].ResultFn(result)

	c.SendWithRetry(title, message, "Notification")
}

// NotifyStat send the stat message into the email
func (c *DefaultNotify) NotifyStat(probers []probe.Prober) {
	if c.Dry {
		c.DryNotifyStat(probers)
		return
	}
	title := "Overall SLA Report"
	message := report.FormatFuncs[c.Format].StatFn(probers)
	c.SendWithRetry(title, message, "SLA")
}

// SendWithRetry sends the notification with retry if got error
func (c *DefaultNotify) SendWithRetry(title string, message string, tag string) {
	fn := func() error {
		log.Debugf("[%s - %s] - %s", c.MyKind, tag, title)
		if c.SendFunc == nil {
			log.Errorf("[%s - %s] - %s SendFunc is nil", c.MyKind, tag, title)
		}
		return c.SendFunc(title, message)
	}
	err := global.DoRetry(c.MyKind, c.Name, tag, c.Retry, fn)
	report.LogSend(c.MyKind, c.Name, tag, title, err)
}

// DryNotify just log the notification message
func (c *DefaultNotify) DryNotify(result probe.Result) {
	log.Infof("[%s / %s] - %s", c.MyKind, c.Name,
		report.FormatFuncs[c.Format].ResultFn(result))
}

// DryNotifyStat just log the notification message
func (c *DefaultNotify) DryNotifyStat(probers []probe.Prober) {
	log.Infof("[%s / %s] - %s", c.MyKind, c.Name,
		report.FormatFuncs[c.Format].StatFn(probers))
}
