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

// Package yunpian is the yunpian sms notification.
package yunpian

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/megaease/easeprobe/notify/sms/conf"
	log "github.com/sirupsen/logrus"
)

// Yunpian is the Yunpian sms provider
type Yunpian struct {
	conf.Options `yaml:",inline"`
}

// New create a yunpian sms provider
func New(opt conf.Options) *Yunpian {
	return &Yunpian{
		Options: opt,
	}
}

// Notify return the type of Notify
func (c Yunpian) Notify(title, text string) error {
	api := c.URL

	form := url.Values{}
	form.Add("apikey", c.Key)
	form.Add("mobile", c.Mobile)
	form.Add("text", c.Sign+text)

	log.Debugf("[%s / %s] - API %s - Form %s", c.Kind(), c.Name(), api, form)
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

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error response from SMS with request body <%s> [%d] - [%s]", form.Encode(), resp.StatusCode, string(buf))
	}
	return nil
}
