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
	"context"
	"log"
	"os"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/base"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/report"
	"github.com/sirupsen/logrus"
)

// NotifyConfig is the configuration of the Notify
type NotifyConfig struct {
	base.DefaultNotify `yaml:",inline"`
	File               string `yaml:"file"`
}

// Kind return the type of Notify
func (c *NotifyConfig) Kind() string {
	return c.MyKind
}

// Config configures the log files
func (c *NotifyConfig) Config(global global.NotifySettings) error {
	c.MyKind = "log"
	c.Format = report.Text
	if c.Dry {
		logrus.Infof("Notification [%s] - [%s] is running on Dry mode!", c.MyKind, c.Name)
		log.SetOutput(os.Stdout)
		return nil
	}
	file, err := os.OpenFile(c.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		logrus.Errorf("error: %s", err)
		return err
	}
	log.SetOutput(file)

	logrus.Infof("Notification [%s] - [%s] is configured!", c.Kind(), c.Name)
	logrus.Debugf("Notification [%s] - [%s] configuration: %+v", c.Kind(), c.Name, c)
	return nil
}

// Notify write the message into the file
func (c *NotifyConfig) Notify(ctx context.Context, result probe.Result) {
	if c.Dry {
		c.DryNotify(result)
		return
	}
	log.Println(result.DebugJSON())
	logrus.Infof("Logged the notification for %s (%s)!", result.Name, result.Endpoint)
}

// NotifyStat write the stat message into the file
func (c *NotifyConfig) NotifyStat(ctx context.Context, probers []probe.Prober) {
	if c.Dry {
		c.DryNotifyStat(probers)
		return
	}
	logrus.Infoln("LogFile Sending the Statstics...")
	for _, p := range probers {
		log.Println(p.Result())
	}
	logrus.Infof("Logged the Statstics into %s!", c.File)
}
