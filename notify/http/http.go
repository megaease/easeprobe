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

// Package http is the HTTP notification
package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/base"
	"github.com/megaease/easeprobe/report"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the HTTP notification configuration
type NotifyConfig struct {
	base.DefaultNotify `yaml:",inline"`

	URL           string   `yaml:"url" json:"url,omitempty" jsonschema:"title=HTTP URL,description=The HTTP endpoint to send notifications"`
	SuccessStatus int      `yaml:"success_status" json:"success_status,omitempty" jsonschema:"title=Success Status,description=The success status code of the HTTP request"`
	Headers       []Header `yaml:"headers" json:"headers,omitempty" jsonschema:"title=HTTP Headers,description=Custom headers for the HTTP request"`
}

// Config configures the HTTP notification
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.NotifyKind = "http"
	c.NotifyFormat = report.Markdown
	c.NotifySendFunc = c.SendHTTP
	c.DefaultNotify.Config(gConf)
	log.Debugf("Notification [%s] - [%s] configuration: %+v", c.NotifyKind, c.NotifyName, c)
	return nil
}

// SendHTTP sends the HTTP notification
func (c *NotifyConfig) SendHTTP(title, text string) error {
	req, err := http.NewRequest(http.MethodPost, c.URL, bytes.NewBuffer([]byte(text)))
	if err != nil {
		return err
	}
	req.Close = true
	for _, h := range c.Headers {
		req.Header.Set(h.Name, h.Value)
	}
	req.Header.Set("User-Agent", "EaseProbe")

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Debugf("[%s] - [%s] sending notification to %s", c.Kind(), c.Name(), c.URL)

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != c.SuccessStatus {
		return fmt.Errorf("Error response from HTTP - code [%d] - msg [%s]", resp.StatusCode, string(buf))
	}
	return nil
}
