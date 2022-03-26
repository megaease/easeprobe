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

package slack

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
	WebhookURL string        `yaml:"webhook"`
	Dry        bool          `yaml:"dry"`
	Timeout    time.Duration `yaml:"timeout"`
	Retry      global.Retry  `yaml:"retry"`
}

// Kind return the type of Notify
func (c *NotifyConfig) Kind() string {
	return "slack"
}

// Config configures the slack notification
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
	json := result.SlackBlockJSON()
	c.SendSlackNotificationWithRetry("Notification", json)
}

// NotifyStat write the all probe stat message to slack
func (c *NotifyConfig) NotifyStat(probers []probe.Prober) {
	if c.Dry {
		c.DryNotifyStat(probers)
		return
	}
	json := probe.StatSlackBlockJSON(probers)
	c.SendSlackNotificationWithRetry("SLA", json)

}

// DryNotify just log the notification message
func (c *NotifyConfig) DryNotify(result probe.Result) {
	log.Infof("[%s] - %s", c.Kind(), result.SlackBlockJSON())
}

// DryNotifyStat just log the notification message
func (c *NotifyConfig) DryNotifyStat(probers []probe.Prober) {
	log.Infof("[%s] - %s", c.Kind(), probe.StatSlackBlockJSON(probers))
}

// SendSlackNotificationWithRetry send the Slack notification with retry
func (c *NotifyConfig) SendSlackNotificationWithRetry(tag string, msg string) {

	for i := 0; i < c.Retry.Times; i++ {
		err := c.SendSlackNotification(msg)
		if err == nil {
			log.Infof("Successfully Sent the %s to Slack", tag)
			return
		}

		log.Debugf("[%s] - %s", c.Kind(), msg)
		log.Warnf("[%s] Retred to send %d/%d -  %v", c.Kind(), i+1, c.Retry.Times, err)

		// last time no need to sleep
		if i < c.Retry.Times-1 {
			time.Sleep(c.Retry.Interval)
		}

	}
	log.Errorf("[%s] Failed to sent the slack after %d retries!", c.Kind(), c.Retry.Times)

}

// SendSlackNotification will post to an 'Incoming Webhook' url setup in Slack Apps. It accepts
// some text and the slack channel is saved within Slack.
func (c *NotifyConfig) SendSlackNotification(msg string) error {
	req, err := http.NewRequest(http.MethodPost, c.WebhookURL, bytes.NewBuffer([]byte(msg)))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Close = true

	client := &http.Client{Timeout: 10 * time.Second}
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
		return fmt.Errorf("Error response from Slack [%d] - [%s]", resp.StatusCode, string(buf))
	}
	// if buf.String() != "ok" {
	// 	return errors.New("Non-ok response returned from Slack " + buf.String())
	// }
	return nil
}
