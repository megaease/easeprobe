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

package dingtalk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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
	return "dingtalk"
}

// Config configures the dingtalk notification
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	if c.Dry {
		log.Infof("Notification [%s] - [%s]  is running on Dry mode!", c.Kind(), c.Name)
	}

	c.Timeout = gConf.NormalizeTimeOut(c.Timeout)
	c.Retry = gConf.NormalizeRetry(c.Retry)

	log.Infof("[%s] configuration: %+v", c.Kind(), c)
	return nil
}

// Notify write the message into the dingtalk
func (c *NotifyConfig) Notify(result probe.Result) {
	if c.Dry {
		c.DryNotify(result)
		return
	}
	c.SendDingtalkNotificationWithRetry("Notification", result.Title(), result.Markdown())
}

// NotifyStat write the all probe stat message to dingtalk
func (c *NotifyConfig) NotifyStat(probers []probe.Prober) {
	if c.Dry {
		c.DryNotifyStat(probers)
		return
	}
	md := probe.StatMarkDown(probers)
	title := strings.Split(md, "\n")[0]
	c.SendDingtalkNotificationWithRetry("SLA", title, md)
}

// DryNotify just log the notification message
func (c *NotifyConfig) DryNotify(result probe.Result) {
	log.Infof("[%s / %s] - %s", c.Kind(), c.Name, result.Markdown())
}

// DryNotifyStat just log the notification message
func (c *NotifyConfig) DryNotifyStat(probers []probe.Prober) {
	log.Infof("[%s / %s] - %s", c.Kind(), c.Name, probe.StatMarkDown(probers))
}

// SendDingtalkNotificationWithRetry send the dingtalk notification with retry
func (c *NotifyConfig) SendDingtalkNotificationWithRetry(tag string, title, msg string) {

	fn := func() error {
		log.Debugf("[%s - %s] - %s", c.Kind(), tag, msg)
		return c.SendDingtalkNotification(title, msg)
	}

	err := global.DoRetry(c.Kind(), c.Name, tag, c.Retry, fn)
	probe.LogSend(c.Kind(), c.Name, tag, "", err)
}

// SendDingtalkNotification will post to an 'Robot Webhook' url in Dingtalk Apps. It accepts
// some text and the Dingtalk robot will send it in group.
func (c *NotifyConfig) SendDingtalkNotification(title, msg string) error {

	// It will be better to escape the msg.
	msgContent := fmt.Sprintf(`
	{
		"msgtype": "markdown",
		"markdown": {
			"title": "%s",
			"text": "%s" 
		}
	}
	`, title, msg)
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
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	ret := make(map[string]interface{})
	err = json.Unmarshal(buf, &ret)
	if err != nil {
		return fmt.Errorf("Error response from Dingtalk [%d] - [%s]", resp.StatusCode, string(buf))
	}
	if ret["errmsg"] != "ok" {
		return fmt.Errorf("Error response from Dingtalk [%d] - [%s]", ret["errcode"], string(buf))
	}
	return nil
}
