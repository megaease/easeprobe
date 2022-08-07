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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/base"
	"github.com/megaease/easeprobe/report"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the dingtalk notification configuration
type NotifyConfig struct {
	base.DefaultNotify `yaml:",inline"`
	WebhookURL         string `yaml:"webhook"`
	SignSecret         string `yaml:"secret"`
}

// Config configures the dingtalk notification
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.NotifyKind = "dingtalk"
	c.NotifyFormat = report.Markdown
	c.NotifySendFunc = c.SendDingtalkNotification
	c.DefaultNotify.Config(gConf)
	log.Debugf("Notification [%s] - [%s] configuration: %+v", c.NotifyKind, c.NotifyName, c)
	return nil
}

// SendDingtalkNotification will post to an 'Robot Webhook' url in Dingtalk Apps. It accepts
// some text and the Dingtalk robot will send it in group.
func (c *NotifyConfig) SendDingtalkNotification(title, msg string) error {

	title = "**" + title + "**"
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
	req, err := http.NewRequest(http.MethodPost, addSign(c.WebhookURL, c.SignSecret), bytes.NewBuffer([]byte(msgContent)))
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

// add sign for url by secret
func addSign(webhookURL string, secret string) string {
	if secret != "" {
		timestamp := time.Now().UnixMilli()
		stringToSign := fmt.Sprint(timestamp, "\n", secret)
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(stringToSign))
		sign := url.QueryEscape(base64.StdEncoding.EncodeToString(h.Sum(nil)))
		return fmt.Sprint(webhookURL, "&timestamp=", timestamp, "&sign="+sign)
	}
	return webhookURL
}
