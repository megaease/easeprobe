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

// Package sms contains the sms implementation.
package sms

import (
	"errors"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/sms/conf"
	"github.com/megaease/easeprobe/notify/sms/nexmo"
	"github.com/megaease/easeprobe/notify/sms/twilio"
	"github.com/megaease/easeprobe/notify/sms/yunpian"
	"github.com/megaease/easeprobe/report"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig implements the structure of Sms
type NotifyConfig struct {
	//Embed structure
	conf.Options `yaml:",inline"`

	Provider conf.Provider `yaml:"-" json:"-"`
}

// Config Sms Config Object
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.NotifyKind = conf.ProviderMap[c.ProviderType]
	c.NotifyFormat = report.SMS
	c.DefaultNotify.Config(gConf)
	c.configSMSDriver()
	c.NotifySendFunc = c.DoNotify

	log.Debugf("Notification [%s] - [%s] configuration: %+v", c.NotifyKind, c.NotifyName, c)
	return nil
}

func (c *NotifyConfig) configSMSDriver() {
	switch c.ProviderType {
	case conf.Yunpian:
		c.Provider = yunpian.New(c.Options)
	case conf.Twilio:
		c.Provider = twilio.New(c.Options)
	case conf.Nexmo:
		c.Provider = nexmo.New(c.Options)
	default:
		c.ProviderType = conf.Unknown
	}

}

// DoNotify return the notify function
func (c *NotifyConfig) DoNotify(title, text string) error {
	if c.ProviderType == conf.Unknown {
		return errors.New("wrong Provider type")
	}
	return c.Provider.Notify(title, text)
}
