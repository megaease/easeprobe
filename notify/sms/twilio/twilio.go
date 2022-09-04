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

// Package twilio is the twilio sms notification
package twilio

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/megaease/easeprobe/notify/sms/conf"
	log "github.com/sirupsen/logrus"
)

// Twilio is the Twilio sms provider
type Twilio struct {
	conf.Options `yaml:",inline"`
}

// New create a Twilio sms provider
func New(opt conf.Options) *Twilio {
	return &Twilio{
		Options: opt,
	}
}

// Notify return the type of Notify
func (c Twilio) Notify(title, text string) error {
	api := c.URL + c.Key + "/Messages.json"

	form := url.Values{}
	form.Add("From", c.From)
	form.Add("To", c.Mobile)
	form.Add("text", text)

	log.Debugf("[%s / %s] - API %s - Form %s", c.Kind(), c.Name(), api, form)
	req, err := http.NewRequest(http.MethodPost, api, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.Key, c.Secret)

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
		return fmt.Errorf("Error response from SMS [%d] - [%s]", resp.StatusCode, string(buf))
	}
	return nil
}
