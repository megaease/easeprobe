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

package sms

import (
	"fmt"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/base"
	"github.com/megaease/easeprobe/report"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// NotifyConfig is the sms notification configuration
type NotifyConfig struct {
	base.DefaultNotify `yaml:",inline"`
	Apikey             string `yaml:"apikey"`
	Mobile             string `yaml:"mobile"`
	Sign               string `yaml:"sign"`
}

// Kind return the type of Notify
func (c *NotifyConfig) Kind() string {
	return c.MyKind
}

// Config configures the sms configuration
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.MyKind = "sms"
	c.Format = report.Markdown
	c.SendFunc = c.SendSms
	c.DefaultNotify.Config(gConf)
	log.Debugf("Notification [%s] - [%s] configuration: %+v", c.MyKind, c.Name, c)
	return nil
}

// SendSms is the wrapper for SendSmsNotification
func (c *NotifyConfig) SendSms(title, text string) error {
	return c.SendSmsNotification(text)
}

// SendSmsNotification will send the notification by sms.
func (c *NotifyConfig) SendSmsNotification(text string) error {
	api := "https://sms.yunpian.com/v2/sms/single_send.json"

	form := url.Values{}
	form.Add("apikey", c.Apikey)
	form.Add("mobile", c.Mobile)
	form.Add("text", c.Sign+text)

	log.Debugf("[%s] - API %s - Form %s", c.Kind(), api, form)
	req, err := http.NewRequest(http.MethodPost, api, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Close = true
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

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
		return fmt.Errorf("Error response from Sms [%d] - [%s]", resp.StatusCode, string(buf))
	}
	return nil
}
