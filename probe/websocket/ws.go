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

package websocket

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/base"
)

// WebSocket implements a Config for a websocket prober.
type WebSocket struct {
	base.DefaultProbe `yaml:",inline"`
	URL               string            `yaml:"url" json:"url" jsonschema:"format=uri,title=WebSocket URL,description=WebSocket URL to probe"`
	Proxy             string            `yaml:"proxy" json:"proxy,omitempty" jsonschema:"format=url,title=Proxy Server,description=proxy to use for the HTTP request"`
	Headers           map[string]string `yaml:"headers,omitempty" json:"headers,omitempty" jsonschema:"title=HTTP Headers,description=HTTP headers for the initial HTTP request"`

	proxy *url.URL
}

// Config Websocket config Object
func (h *WebSocket) Config(gConf global.ProbeSettings) error {
	kind := "websocket"
	tag := ""
	name := h.ProbeName
	h.DefaultProbe.Config(gConf, kind, tag, name, h.URL, h.DoProbe)

	url, err := url.Parse(h.URL)
	if err != nil {
		return err
	}

	if url.Scheme != "ws" && url.Scheme != "wss" {
		return fmt.Errorf(`the scheme should be "ws" or "wss", but got: %s`, url.Scheme)
	}

	if h.Proxy != "" {
		h.proxy, err = url.Parse(h.Proxy)
		if err != nil {
			return err
		}
	}

	return nil
}

// DoProbe return the checking result
func (h *WebSocket) DoProbe() (bool, string) {
	wsHeader := make(http.Header)
	for k, v := range h.Headers {
		wsHeader.Set(k, v)
	}

	begin := time.Now()
	remaining := h.ProbeTimeout

	var dial = websocket.DefaultDialer
	dial.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	dial.HandshakeTimeout = remaining
	if h.proxy != nil {
		dial.Proxy = func(request *http.Request) (*url.URL, error) {
			return h.proxy, nil
		}
	}

	ws, _, err := dial.Dial(h.URL, wsHeader)
	if err != nil {
		return false, err.Error()
	}

	defer ws.Close()

	pingPongChan := make(chan struct{})
	ws.SetPongHandler(func(appData string) error {
		pingPongChan <- struct{}{}
		return nil
	})

	// doing nothing but trigger the read message loop to receive the
	// Pong and Close messages from the server. exit after ws.Close()
	go func() {
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				break
			}
		}
	}()

	remaining = h.ProbeTimeout - time.Since(begin)
	err = ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(remaining))
	if err != nil {
		return false, err.Error()
	}

	remaining = h.ProbeTimeout - time.Since(begin)
	t := time.NewTimer(remaining)
	defer t.Stop()

	select {
	case <-t.C:
		return false, "ping timeout"
	case <-pingPongChan:
		remaining = h.ProbeTimeout - time.Since(begin)
		closeCode := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
		// try to do a graceful close, but do not care the result
		err := ws.WriteControl(websocket.CloseMessage, closeCode, time.Now().Add(remaining))
		if err != nil {
			log.Error(err)
		}

		return true, ""
	}
}
