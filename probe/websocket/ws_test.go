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
	"fmt"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/base"
)

type TestCase struct {
	URL     string
	Timeout time.Duration
	Headers map[string]string
	Want    bool
}

var (
	token = map[string]string{"Authorization": "token 123456"}
)

func TestWSPing(t *testing.T) {
	go func() {
		err := http.ListenAndServe(":18080", &Handler{})
		if err != nil {
			log.Fatal(err)
		}
	}()

	// wait for http server to start
	conn, err := net.Conn(nil), error(nil)
	for i := 0; i < 30; i++ {
		conn, err = net.DialTimeout("tcp", "127.0.0.1:18080", 1*time.Second)
		if err == nil {
			conn.Close()
			break
		}
		t.Log("waiting for http server to start...")
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		t.Fatalf("http server not started: %v", err)
	}

	testcases := []TestCase{
		{URL: "ws://127.0.0.1:18080/right", Timeout: 500 * time.Millisecond, Headers: token, Want: false},
		{URL: "ws://127.0.0.1:18080/right", Timeout: 2000 * time.Millisecond, Headers: token, Want: true},
		{URL: "ws://127.0.0.1:18080/right", Timeout: 2000 * time.Millisecond, Headers: nil, Want: false},
		{URL: "ws://127.0.0.1:18080/wrong", Timeout: 2000 * time.Millisecond, Headers: token, Want: false},
	}

	for i, test := range testcases {
		ws := WebSocket{
			DefaultProbe: base.DefaultProbe{
				ProbeTimeout: test.Timeout,
			},
			URL:     test.URL,
			Headers: test.Headers,
		}
		ws.Config(global.ProbeSettings{})
		ok, err := ws.DoProbe()
		assert.Equalf(t, test.Want, ok, fmt.Sprintf("case %d %s", i, err))
	}
}

type Handler struct {
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var u = websocket.Upgrader{} // use default options
	reqToken := r.Header.Get("Authorization")

	if r.URL.Path != "/right" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if reqToken != token["Authorization"] {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	c, err := u.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	c.SetPingHandler(func(appData string) error {
		log.Printf("ping from: %s", appData)
		time.Sleep(1 * time.Second)
		c.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(1*time.Second))
		return nil
	})

	c.SetCloseHandler(func(code int, text string) error {
		fmt.Printf("closed")
		return nil
	})

	for {
		// do nothing but trigger the loop to handle Ping/Pong message internally
		_, _, err := c.ReadMessage()
		if err != nil {
			break
		}
	}
}
