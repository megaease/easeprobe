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

package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

// Kind is the type of probe
const Kind string = "http"

// HTTP implements a config for HTTP.
type HTTP struct {
	Name            string            `yaml:"name"`
	URL             string            `yaml:"url"`
	ContentEncoding string            `yaml:"content_encoding,omitempty"`
	Method          string            `yaml:"method,omitempty"`
	Headers         map[string]string `yaml:"headers,omitempty"`
	Body            string            `yaml:"body,omitempty"`

	//Option - HTTP Basic Auth Credentials
	User string `yaml:"username,omitempty"`
	Pass string `yaml:"password,omitempty"`

	//Option - TLS Config
	global.TLS `yaml:",inline"`

	//Control Options
	Timeout      time.Duration `yaml:"timeout,omitempty"`
	TimeInterval time.Duration `yaml:"interval,omitempty"`

	client *http.Client  `yaml:"-"`
	result *probe.Result `yaml:"-"`
}

func checkHTTPMethod(m string) bool {

	methods := [...]string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "CONNECT", "OPTIONS", "TRACE"}
	for _, method := range methods {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

// Kind return the HTTP kind
func (h *HTTP) Kind() string {
	return Kind
}

// Interval get the interval
func (h *HTTP) Interval() time.Duration {
	return h.TimeInterval
}

// Result get the probe result
func (h *HTTP) Result() *probe.Result {
	return h.result
}

// Config HTTP Config Object
func (h *HTTP) Config(gConf global.ProbeSettings) error {

	h.Timeout = gConf.NormalizeTimeOut(h.Timeout)
	h.TimeInterval = gConf.NormalizeInterval(h.TimeInterval)

	tls, err := h.TLS.Config()
	if err != nil {
		log.Errorf("TLS configuration error - %s", err)
		return err
	}

	h.client = &http.Client{
		Timeout: h.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: tls,
		},
	}
	if !checkHTTPMethod(h.Method) {
		h.Method = "GET"
	}

	h.result = probe.NewResult()
	h.result.Name = h.Name
	h.result.Endpoint = h.URL
	h.result.PreStatus = probe.StatusInit
	h.result.TimeFormat = gConf.TimeFormat

	log.Debugf("[%s] configuration: %+v, %+v", h.Kind(), h, h.Result())
	return nil
}

// Probe return the checking result
func (h *HTTP) Probe() probe.Result {

	req, _ := http.NewRequest(h.Method, h.URL, bytes.NewBuffer([]byte(h.Body)))
	if len(h.User) > 0 && len(h.Pass) > 0 {
		req.SetBasicAuth(h.User, h.Pass)
	}
	if len(h.ContentEncoding) > 0 {
		req.Header.Set("Content-Type", h.ContentEncoding)
	}
	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}

	// client close the connection
	req.Close = true

	req.Header.Set("User-Agent", global.OrgProgVer)

	now := time.Now()
	h.result.StartTime = now
	h.result.StartTimestamp = now.UnixMilli()

	resp, err := h.client.Do(req)
	h.result.RoundTripTime.Duration = time.Since(now)
	status := probe.StatusUp
	if err != nil {
		h.result.Message = fmt.Sprintf("Error: %v", err)
		log.Errorf("error making get request: %v", err)
		status = probe.StatusDown
	} else {
		// Read the response body
		defer resp.Body.Close()
		response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Debugf("%s", string(response))
		}
		h.result.Message = fmt.Sprintf("Success: HTTP Status Code is %d", resp.StatusCode)
		if resp.StatusCode >= 500 {
			h.result.Message = fmt.Sprintf("Error: HTTP Status Code is %d", resp.StatusCode)
			status = probe.StatusDown
		}
	}

	h.result.PreStatus = h.result.Status
	h.result.Status = status

	h.result.DoStat(h.TimeInterval)

	return *h.result
}
