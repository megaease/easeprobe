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

package base

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/proxy"
)

var (
	status = false
	msg    = map[bool]string{
		true:  "success",
		false: "failed",
	}
	pStatus = map[bool]probe.Status{
		true:  probe.StatusUp,
		false: probe.StatusDown,
	}
)

type dummyProber struct {
	DefaultProbe
}

func (d *dummyProber) Config(g global.ProbeSettings) error {
	d.DefaultProbe.Config(g, d.ProbeKind, d.ProbeTag, d.ProbeName, "endpoint", d.DoProbe)
	return nil
}
func (d *dummyProber) DoProbe() (bool, string) {
	status = rand.Int()%2 == 0
	return status, msg[status]
}

func newDummyProber(name string) *dummyProber {
	return &dummyProber{
		DefaultProbe: DefaultProbe{
			ProbeKind:         "dummy",
			ProbeTag:          "tag",
			ProbeName:         name,
			ProbeTimeout:      0,
			ProbeTimeInterval: 0,
			ProbeFunc:         nil,
			ProbeResult:       &probe.Result{},
		},
	}
}

func TestBase(t *testing.T) {
	global.InitEaseProbe("DummyProbe", "icon")
	p := newDummyProber("probe")
	p.Config(global.ProbeSettings{})
	assert.Equal(t, "dummy", p.Kind())
	assert.Equal(t, "probe", p.Name())
	assert.Equal(t, []string{global.DefaultChannelName}, p.Channels())
	assert.Equal(t, global.DefaultTimeOut, p.Timeout())
	assert.Equal(t, global.DefaultProbeInterval, p.Interval())
	assert.Equal(t, probe.StatusInit, p.Result().Status)

	p.ProbeTag = ""
	p.Config(global.ProbeSettings{})
	assert.Equal(t, "dummy", p.Kind())
	assert.Equal(t, "probe", p.Name())

	for i := 0; i < 10; i++ {
		p.Probe()
		preStatus := p.Result().Status
		assert.Equal(t, pStatus[status], preStatus)
		assert.Contains(t, p.Result().Message, msg[status])

		p.ProbeTag = "tag"
		p.Probe()
		assert.Equal(t, preStatus, p.Result().PreStatus)
		assert.Equal(t, pStatus[status], p.Result().Status)
	}

	p.ProbeFunc = nil
	r := p.Probe()
	assert.Equal(t, *p.ProbeResult, r)
}

func TestProxyConnection(t *testing.T) {
	p := newDummyProber("probe")
	p.Config(global.ProbeSettings{})

	os.Setenv("ALL_PROXY", "socks5://localhost:1080")

	conn, err := p.GetProxyConnection("sock://localhost:8080", "host:80")
	assert.NotNil(t, err)
	assert.Nil(t, conn)

	conn, err = p.GetProxyConnection("sock5://\n\r", "host:80")
	assert.NotNil(t, err)
	assert.Nil(t, conn)

	monkey.Patch(proxy.SOCKS5, func(network string, address string, auth *proxy.Auth, forward proxy.Dialer) (proxy.Dialer, error) {
		return &net.Dialer{}, nil
	})
	var dialer *net.Dialer
	monkey.PatchInstanceMethod(reflect.TypeOf(dialer), "Dial", func(_ *net.Dialer, network, address string) (net.Conn, error) {
		return &net.TCPConn{}, nil
	})

	conn, err = p.GetProxyConnection("", "host:80")
	assert.Nil(t, err)
	assert.NotNil(t, conn)

	conn, err = p.GetProxyConnection("socks5://localhost:8080", "host:80")
	assert.Nil(t, err)
	assert.NotNil(t, conn)

	monkey.Patch(proxy.FromEnvironment, func() proxy.Dialer {
		return proxy.Direct
	})

	monkey.Patch(net.DialTimeout, func(string, string, time.Duration) (net.Conn, error) {
		return &net.TCPConn{}, nil
	})
	conn, err = p.GetProxyConnection("", "host:80")
	assert.Nil(t, err)
	assert.NotNil(t, conn)

	monkey.UnpatchAll()
}

func TestStatusThreshold(t *testing.T) {
	p := newDummyProber("probe")
	p.StatusChangeThresholdSettings = global.StatusChangeThresholdSettings{
		Failure: 2,
		Success: 3,
	}
	p.Config(global.ProbeSettings{
		StatusChangeThresholdSettings: global.StatusChangeThresholdSettings{
			Failure: 2,
			Success: 1,
		},
	})
	assert.Equal(t, 3, p.ProbeResult.Stat.MaxLen)
	assert.Equal(t, 2, p.StatusChangeThresholdSettings.Failure)
	assert.Equal(t, 3, p.StatusChangeThresholdSettings.Success)

	p.ProbeResult.Status = probe.StatusInit

	cnt := 0
	p.ProbeFunc = func() (bool, string) {
		cnt++
		return true, fmt.Sprintf("success - %d", cnt)
	}

	n := p.ProbeResult.Stat.MaxLen + 2
	for i := 1; i <= n; i++ {
		p.Probe()
		if i < p.StatusChangeThresholdSettings.Success {

			assert.Equal(t, probe.StatusInit, p.Result().Status)
		} else {
			assert.Equal(t, probe.StatusUp, p.Result().Status)
		}
		if i < p.ProbeResult.Stat.MaxLen {
			assert.Equal(t, i, p.ProbeResult.Stat.StatusCount)
		} else {
			assert.Equal(t, p.ProbeResult.Stat.MaxLen, p.ProbeResult.Stat.StatusCount)
		}
	}

	cnt = 0
	p.ProbeFunc = func() (bool, string) {
		cnt++
		return false, fmt.Sprintf("failure - %d", cnt)
	}

	for i := 1; i <= n; i++ {
		p.Probe()
		if i < p.StatusChangeThresholdSettings.Failure {
			assert.Equal(t, probe.StatusUp, p.Result().Status)
		} else {
			assert.Equal(t, probe.StatusDown, p.Result().Status)
		}
		if i < p.ProbeResult.Stat.MaxLen {
			assert.Equal(t, i, p.ProbeResult.Stat.StatusCount)
		} else {
			assert.Equal(t, p.ProbeResult.Stat.MaxLen, p.ProbeResult.Stat.StatusCount)
		}
	}
}

func TestStatusThreshold2(t *testing.T) {
	p := newDummyProber("probe")
	p.Failure = 2
	p.Success = 1
	p.Config(global.ProbeSettings{})
	p.ProbeResult.Status = probe.StatusUp

	p.ProbeFunc = func() (bool, string) {
		return true, fmt.Sprintf("success")
	}
	p.Probe()
	assert.Equal(t, probe.StatusUp, p.Result().Status)

	p.ProbeFunc = func() (bool, string) {
		return false, fmt.Sprintf("failure")
	}
	p.Probe()
	assert.Equal(t, probe.StatusUp, p.Result().Status)
	p.Probe()
	assert.Equal(t, probe.StatusDown, p.Result().Status)
	p.Probe()
	assert.Equal(t, probe.StatusDown, p.Result().Status)

	p.ProbeFunc = func() (bool, string) {
		return true, fmt.Sprintf("success")
	}
	p.Probe()
	assert.Equal(t, probe.StatusUp, p.Result().Status)
}
