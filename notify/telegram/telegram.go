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

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/base"
	"github.com/megaease/easeprobe/report"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the telegram notification configuration
type NotifyConfig struct {
	base.DefaultNotify `yaml:",inline"`
	Token              string `yaml:"token"`
	ChatID             string `yaml:"chat_id"`
}

// Config configures the telegram configuration
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.NotifyKind = "telegram"
	c.NotifyFormat = report.Markdown
	c.NotifySendFunc = c.SendTelegram
	c.DefaultNotify.Config(gConf)
	log.Debugf("Notification [%s] - [%s] configuration: %+v", c.NotifyKind, c.NotifyName, c)
	return nil
}

// SendTelegram is the wrapper for SendTelegramNotification
func (c *NotifyConfig) SendTelegram(title, text string) error {
	return c.SendTelegramNotification(text)
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
