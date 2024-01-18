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

// Package wecom is the wecom notification package.
package wecom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/base"
	"github.com/megaease/easeprobe/report"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the slack notification configuration
type NotifyConfig struct {
	base.DefaultNotify `yaml:",inline"`
	WebhookURL         string `yaml:"webhook" json:"webhook" jsonschema:"required,format=url,title=Webhook URL,description=The Webhook URL of Wecom"`
}

// Config configures the slack notification
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.NotifyKind = "wecom"
	c.NotifyFormat = report.Markdown
	c.NotifySendFunc = c.SendWecom
	c.DefaultNotify.Config(gConf)
	log.Debugf("Notification [%s] - [%s] configuration: %+v", c.NotifyKind, c.NotifyName, c)
	return nil
}

// SendWecom is the wrapper of SendWecomNotification
func (c *NotifyConfig) SendWecom(title, msg string) error {
	return c.SendWecomNotification(msg)
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
	`, report.JSONEscape(msg))
	if !json.Valid([]byte(msgContent)) {
		log.Errorf("[%s / %s ] - %v, err: invalid json", c.Kind(), c.Name(), msgContent)
	}

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

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error response from Wecom with request body <%s> - code [%d] - msg [%s]", msgContent, resp.StatusCode, string(buf))
	}
	// It will be better to check response body.
	return nil
}
