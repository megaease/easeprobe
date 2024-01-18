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

// Package teams is the teams notification
package teams

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

// NotifyConfig is the teams notification configuration
type NotifyConfig struct {
	base.DefaultNotify `yaml:",inline"`
	WebhookURL         string `yaml:"webhook"  json:"webhook" jsonschema:"required,format=url,title=Webhook URL,description=The Microsoft Teams Robot Webhook URL"`
}

// Config configures the teams notification
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.NotifyKind = "teams"
	c.NotifyFormat = report.MarkdownSocial
	c.NotifySendFunc = c.SendTeamsMessage
	c.DefaultNotify.Config(gConf)

	log.Debugf("Notification [%s] - [%s] configuration: %+v", c.NotifyKind, c.NotifyName, c)
	return nil
}

type messageCard struct {
	Type    string `json:"@type"`
	Context string `json:"@context"`
	Title   string `json:"title,omitempty"`
	Text    string `json:"text,omitempty"`
}

// SendTeamsMessage sends the message to the teams channel
func (c *NotifyConfig) SendTeamsMessage(title, msg string) error {

	json, err := json.Marshal(messageCard{
		Type:    "MessageCard",
		Context: "https://schema.org/extensions",
		Title:   title,
		Text:    msg,
	})

	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.WebhookURL, bytes.NewBuffer([]byte(json)))
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

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 && string(buf) != "1" {
		return fmt.Errorf("error response from Teams Webhook with request body <%s> - code [%d] - msg [%s]", json, resp.StatusCode, string(buf))
	}
	return nil
}
