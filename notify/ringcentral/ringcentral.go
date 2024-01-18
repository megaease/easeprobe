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

// Package ringcentral is the ringcentral notification package.
package ringcentral

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

// NotifyConfig is the ringcentral notification configuration
type NotifyConfig struct {
	base.DefaultNotify `yaml:",inline"`
	WebhookURL         string `yaml:"webhook" json:"webhook" jsonschema:"required,format=uri,title=Webhook URL,description=The RingCentral webhook URL"`
}

// Config configures the ringcentral notification
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.NotifyKind = "ringcentral"
	c.NotifyFormat = report.Text
	c.NotifySendFunc = c.SendRingCentral
	c.DefaultNotify.Config(gConf)
	log.Debugf("Notification [%s] - [%s] configuration: %+v", c.NotifyKind, c.NotifyName, c)
	return nil
}

// SendRingCentral will compose a message with title and msg, then post it to the 'Incoming Webhook' URL configured in RingCentral App.
func (c *NotifyConfig) SendRingCentral(title, msg string) error {
	msgContent := fmt.Sprintf(`
	{
		"attachments":[
		   {
			  "$schema":"http://adaptivecards.io/schemas/adaptive-card.json",
			  "type":"AdaptiveCard",
			  "version":"1.0",
			  "body":[
				 {
					"type":"TextBlock",
					"text":"%s",
					"weight":"bolder",
					"size":"medium",
					"wrap":true
				 },
				 {
					"type":"TextBlock",
					"text":"%s",
					"wrap":true
				 }
			  ]
		   }
		]
	 }
	`, report.JSONEscape(title), report.JSONEscape(msg))
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
		return fmt.Errorf("Error response from RingCentral with request body <%s> - code [%d] - msg [%s]", msgContent, resp.StatusCode, string(buf))
	}
	if string(buf) != "{\"status\":\"OK\"}" {
		return fmt.Errorf("Non-ok response returned from RingCentral with request body <%s>, %s", msgContent, string(buf))
	}
	return nil
}
