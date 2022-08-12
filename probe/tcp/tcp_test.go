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

package tcp

import (
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/proxy"
)

func TestTCP(t *testing.T) {
	global.InitEaseProbe("easeprobe", "http://icon")
	tcp := TCP{
		DefaultProbe: base.DefaultProbe{ProbeName: "dummy tcp"},
		Host:         "example.com:8888",
	}

	tcp.Config(global.ProbeSettings{})
	assert.Equal(t, "tcp", tcp.ProbeKind)

	monkey.Patch(net.DialTimeout, func(network, address string, timeout time.Duration) (net.Conn, error) {
		return &net.TCPConn{}, nil
	})
	var conn *net.TCPConn
	monkey.PatchInstanceMethod(reflect.TypeOf(conn), "Close", func(_ *net.TCPConn) error {
		return nil
	})

	s, m := tcp.DoProbe()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	monkey.Patch(net.DialTimeout, func(network, address string, timeout time.Duration) (net.Conn, error) {
		return nil, fmt.Errorf("tcp dial error")
	})
	s, m = tcp.DoProbe()
	assert.False(t, s)
	assert.Contains(t, m, "tcp dial error")
}

func TestTCPProxy(t *testing.T) {
	global.InitEaseProbe("easeprobe", "http://icon")
	tcp := TCP{
		DefaultProbe: base.DefaultProbe{ProbeName: "dummy tcp"},
		Host:         "example.com:8888",
	}
	tcp.Proxy = "http://\n\r"
	s, m := tcp.DoProbe()
	assert.False(t, s)
	assert.Contains(t, m, "Invalid proxy")

	tcp.Proxy = "sock:///localhost:1080"
	s, m = tcp.DoProbe()
	assert.False(t, s)
	assert.Contains(t, m, "Invalid proxy")

	monkey.Patch(proxy.SOCKS5, func(network string, address string, auth *proxy.Auth, forward proxy.Dialer) (proxy.Dialer, error) {
		return &net.Dialer{}, nil
	})
	var dialer *net.Dialer
	monkey.PatchInstanceMethod(reflect.TypeOf(dialer), "Dial", func(_ *net.Dialer, network, address string) (net.Conn, error) {
		return &net.TCPConn{}, nil
	})
	var conn *net.TCPConn
	monkey.PatchInstanceMethod(reflect.TypeOf(conn), "Close", func(_ *net.TCPConn) error {
		return nil
	})

	tcp.Proxy = "socks5://localhost:1080"
	s, m = tcp.DoProbe()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	monkey.UnpatchAll()
}
