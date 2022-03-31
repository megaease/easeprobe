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

package wecom

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the slack notification configuration
type NotifyConfig struct {
	Name       string        `yaml:"name"`
	WebhookURL string        `yaml:"webhook"`
	Dry        bool          `yaml:"dry"`
	Timeout    time.Duration `yaml:"timeout"`
	Retry      global.Retry  `yaml:"retry"`
}

// Kind return the type of Notify
func (c *NotifyConfig) Kind() string {
	return "wecom"
}

// Config configures the slack notification
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	if c.Dry {
		log.Infof("Notification [%s] - [%s]  is running on Dry mode!", c.Kind(), c.Name)
	}

	c.Timeout = gConf.NormalizeTimeOut(c.Timeout)
	c.Retry = gConf.NormalizeRetry(c.Retry)

	log.Infof("[%s] configuration: %+v", c.Kind(), c)
	return nil
}

// Notify write the message into the wecom
func (c *NotifyConfig) Notify(result probe.Result) {
	if c.Dry {
		c.DryNotify(result)
		return
	}
	c.SendWecomNotificationWithRetry("Notification", result.Markdown())
}

// NotifyStat write the all probe stat message to wecom
func (c *NotifyConfig) NotifyStat(probers []probe.Prober) {
	if c.Dry {
		c.DryNotifyStat(probers)
		return
	}
	c.SendWecomNotificationWithRetry("SLA", probe.StatMarkDown(probers))

}

// DryNotify just log the notification message
func (c *NotifyConfig) DryNotify(result probe.Result) {
	log.Infof("[%s / %s] - %s", c.Kind(), c.Name, result.Markdown())
}

// DryNotifyStat just log the notification message
func (c *NotifyConfig) DryNotifyStat(probers []probe.Prober) {
	log.Infof("[%s / %s] - %s", c.Kind(), c.Name, probe.StatMarkDown(probers))
}

// SendWecomNotificationWithRetry send the wecom notification with retry
func (c *NotifyConfig) SendWecomNotificationWithRetry(tag string, msg string) {

	fn := func() error {
		log.Debugf("[%s - %s] - %s", c.Kind(), tag, msg)
		return c.SendWecomNotification(msg)
	}

	err := global.DoRetry(c.Kind(), c.Name, tag, c.Retry, fn)
	probe.LogSend(c.Kind(), c.Name, tag, "", err)
}

// SendWecomNotification will post to an 'Robot Webhook' url in Wecom Apps. It accepts
// some text and the Wecom robot will send it in group.
// https://developer.work.weixin.qq.com/document/path/91770
func (c *NotifyConfig) SendWecomNotification(msg string) error {

	// It will be better to escape the msg.
	msgContent := fmt.Sprintf(`
	{
		"msgtype": "markdown",
		"markdown": {
			"content": "%s" 
		}
	}
	`, msg)
	req, err := http.NewRequest(http.MethodPost, c.WebhookURL, bytes.NewBuffer([]byte(msgContent)))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Close = true

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
		return fmt.Errorf("Error response from Wecom [%d] - [%s]", resp.StatusCode, string(buf))
	}
	// It will be better to check response body.
	return nil
}
