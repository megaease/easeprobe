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

package telegram

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the telegram notification configuration
type NotifyConfig struct {
	Token   string        `yaml:"token"`
	ChatID  string        `yaml:"chat_id"`
	Dry     bool          `yaml:"dry"`
	Timeout time.Duration `yaml:"timeout"`
	Retry   global.Retry  `yaml:"retry"`
}

// Kind return the type of Notify
func (c *NotifyConfig) Kind() string {
	return "telegram"
}

// Config configures the telegram configuration
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	if c.Dry {
		log.Infof("Notification %s is running on Dry mode!", c.Kind())
	}

	c.Timeout = gConf.NormalizeTimeOut(c.Timeout)
	c.Retry = gConf.NormalizeRetry(c.Retry)

	log.Infof("[%s] configuration: %+v", c.Kind(), c)
	return nil
}

// Notify write the message into the slack
func (c *NotifyConfig) Notify(result probe.Result) {
	if c.Dry {
		c.DryNotify(result)
		return
	}
	c.SendTelegramNotificationWithRetry("Notification", result.Markdown())
}

// NotifyStat write the all probe stat message to slack
func (c *NotifyConfig) NotifyStat(probers []probe.Prober) {
	if c.Dry {
		c.DryNotifyStat(probers)
		return
	}

	c.SendTelegramNotificationWithRetry("SLA", probe.StatMarkDown(probers))

}

// DryNotify just log the notification message
func (c *NotifyConfig) DryNotify(result probe.Result) {
	log.Infof("[%s] - %s", c.Kind(), result.Markdown())
}

// DryNotifyStat just log the notification message
func (c *NotifyConfig) DryNotifyStat(probers []probe.Prober) {
	log.Infof("[%s] - %s", c.Kind(), probe.StatMarkDown(probers))
}

// SendTelegramNotificationWithRetry send the telegram notification with retry
func (c *NotifyConfig) SendTelegramNotificationWithRetry(tag string, text string) {

	fn := func() error {
		log.Debugf("[%s - %s] - %s", c.Kind(), tag, text)
		return c.SendTelegramNotification(text)
	}
	err := global.DoRetry(c.Kind(), tag, c.Retry, fn)
	if err != nil {
		log.Errorf("[%s - %s] - failed to send! (%v)", c.Kind(), tag, err)
	} else {
		log.Infof("[%s - %s] - successfully sent! (%v)", c.Kind(), tag)
	}
}

// SendTelegramNotification will send the notification to telegram.
func (c *NotifyConfig) SendTelegramNotification(text string) error {
	api := "https://api.telegram.org/bot" + c.Token +
		"/sendMessage?&chat_id=" + c.ChatID +
		"&parse_mode=markdown" +
		"&text=" + url.QueryEscape(text)
	log.Debugf("[%s] - API %s", c.Kind(), api)
	req, err := http.NewRequest(http.MethodPost, api, nil)
	if err != nil {
		return err
	}
	req.Close = true
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error response from Telegram [%d] - [%s]", resp.StatusCode, string(buf))
	}
	return nil
}
