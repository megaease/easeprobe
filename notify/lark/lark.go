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

package lark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/base"
	"github.com/megaease/easeprobe/report"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the slack notification configuration
type NotifyConfig struct {
	base.DefaultNotify `yaml:",inline"`
	WebhookURL         string `yaml:"webhook"`
}

// Config configures the slack notification
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.NotifyKind = "lark"
	c.NotifyFormat = report.Lark
	c.NotifySendFunc = c.SendLark
	c.DefaultNotify.Config(gConf)
	log.Debugf("Notification [%s] - [%s] configuration: %+v", c.NotifyKind, c.NotifyName, c)
	return nil
}

// SendLark is the wrapper for SendLarkNotification
func (c *NotifyConfig) SendLark(title, msg string) error {
	return c.SendLarkNotification(msg)
}

// SendLarkNotification will post to an 'Robot Webhook' url in Lark Apps. It accepts
// some text and the Lark robot will send it in group.
func (c *NotifyConfig) SendLarkNotification(msg string) error {
	req, err := http.NewRequest(http.MethodPost, c.WebhookURL, bytes.NewBuffer([]byte(msg)))
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
		return fmt.Errorf("Error response from Lark [%d] - [%s]", resp.StatusCode, string(buf))
	}
	// Server returns {"Extra":null,"StatusCode":0,"StatusMessage":"success"} on success
	// otherwise it returns {"code":9499,"msg":"Bad Request","data":{}}
	if statusCode, ok := ret["StatusCode"].(float64); !ok || statusCode != 0 {
		return fmt.Errorf("Error response from Lark [%d] - [%s]", int(ret["code"].(float64)), ret["msg"])
	}
	return nil
}
