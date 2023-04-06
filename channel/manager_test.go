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

package channel

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify"
	baseNotify "github.com/megaease/easeprobe/notify/base"
	"github.com/megaease/easeprobe/probe"
	baseProbe "github.com/megaease/easeprobe/probe/base"
	"github.com/megaease/easeprobe/report"
	"github.com/stretchr/testify/assert"
)

type dummyProber struct {
	baseProbe.DefaultProbe
}

func (d *dummyProber) Config(gConf global.ProbeSettings) error {
	d.ProbeTimeout = gConf.Timeout
	d.ProbeTimeInterval = gConf.Interval
	status := true
	d.DefaultProbe.Config(gConf, d.ProbeKind, d.ProbeTag, d.ProbeName, d.ProbeName,
		func() (bool, string) {
			switch d.ProbeName {
			case "dummy-X":
				return true, "OK"
			case "level-trigger":
				return false, "ERROR"
			default:
				status = !status
				return status, fmt.Sprintf("%s - %v", d.ProbeName, status)
			}
		})
	return nil
}

func newDummyProber(kind, tag, name string, channels []string) *dummyProber {
	return &dummyProber{
		DefaultProbe: baseProbe.DefaultProbe{
			ProbeKind:         kind,
			ProbeTag:          tag,
			ProbeName:         name,
			ProbeChannels:     channels,
			ProbeTimeout:      time.Second,
			ProbeTimeInterval: 5 * time.Second,
			ProbeFunc:         nil,
			ProbeResult:       &probe.Result{},
		},
	}

}

type dummyNotify struct {
	baseNotify.DefaultNotify
}

func newDummyNotify(kind, name string, channels []string) *dummyNotify {
	return &dummyNotify{
		DefaultNotify: baseNotify.DefaultNotify{
			NotifyKind:   kind,
			NotifyFormat: report.Text,
			NotifySendFunc: func(string, string) error {
				return nil
			},
			NotifyName:     name,
			NotifyChannels: channels,
			Dry:            false,
		},
	}
}

func TestManager(t *testing.T) {

	name := "test"
	SetNotify(name, newDummyNotify("email", "dummy", []string{"test"}))
	nm := GetNotifiers([]string{"nil-channel"})
	assert.Equal(t, len(nm), 0)
	nm = GetNotifiers([]string{"test"})
	assert.Equal(t, len(nm), 1)
	assert.Equal(t, nm["dummy"].Name(), "dummy")

	SetProber(name, newDummyProber("http", "", "dummy", []string{"test"}))
	test := GetChannel(name)
	assert.NotNil(t, test)

	probers := []probe.Prober{
		newDummyProber("http", "XY", "dummy-XY", []string{"X", "Y"}),
		newDummyProber("http", "X", "dummy-X", []string{"X"}),
		newDummyProber("http", "Y", "dummy-Y", []string{"Y"}),
		newDummyProber("http", "ALL", "dummy-ALL", []string{"X", "Y", "test"}),
	}
	SetProbers(probers)
	x := GetChannel("X")
	assert.NotNil(t, x)
	assert.NotNil(t, x.GetProber("dummy-X"))
	assert.NotNil(t, x.GetProber("dummy-XY"))
	assert.Equal(t, "dummy-X", x.GetProber("dummy-X").Name())
	assert.Equal(t, "dummy-XY", x.GetProber("dummy-XY").Name())
	assert.Equal(t, "dummy-ALL", x.GetProber("dummy-ALL").Name())

	y := GetChannel("Y")
	assert.NotNil(t, y)
	assert.Nil(t, y.GetProber("dummy-X"))
	assert.NotNil(t, y.GetProber("dummy-Y"))
	assert.NotNil(t, y.GetProber("dummy-XY"))
	assert.Equal(t, "dummy-Y", y.GetProber("dummy-Y").Name())
	assert.Equal(t, "dummy-XY", y.GetProber("dummy-XY").Name())
	assert.Equal(t, "dummy-ALL", y.GetProber("dummy-ALL").Name())

	assert.Equal(t, "dummy-ALL", test.GetProber("dummy-ALL").Name())

	notifiers := []notify.Notify{
		newDummyNotify("email", "dummy-XY", []string{"X", "Y"}),
		newDummyNotify("email", "dummy-X", []string{"X"}),
	}
	SetNotifiers(notifiers)
	assert.NotNil(t, x.GetNotify("dummy-X"))
	assert.NotNil(t, x.GetNotify("dummy-XY"))
	assert.Equal(t, "dummy-XY", x.GetNotify("dummy-XY").Name())
	assert.Equal(t, "dummy-X", x.GetNotify("dummy-X").Name())

	assert.Nil(t, y.GetNotify("dummy-X"))
	assert.NotNil(t, y.GetNotify("dummy-XY"))
	assert.Equal(t, "dummy-XY", y.GetNotify("dummy-XY").Name())

	chs := GetAllChannels()
	assert.Equal(t, 3, len(chs))
	assert.NotNil(t, "test", chs["test"])
	assert.NotNil(t, "X", chs["X"])
	assert.NotNil(t, "Y", chs["Y"])

	global.InitEaseProbe("DummyProbe", "")

	gProbeConf := global.ProbeSettings{}
	test.GetProber("dummy").Config(gProbeConf)
	x.GetProber("dummy-X").Config(gProbeConf)
	x.GetProber("dummy-XY").Config(gProbeConf)
	y.GetProber("dummy-Y").Config(gProbeConf)

	gNotifyConf := global.NotifySettings{}
	test.GetNotify("dummy").Config(gNotifyConf)
	x.GetNotify("dummy-X").Config(gNotifyConf)
	y.GetNotify("dummy-XY").Config(gNotifyConf)

	for _, ch := range chs {
		assert.Nil(t, ch.channel)
		assert.Nil(t, ch.done)
	}
	ConfigAllChannels()
	for _, ch := range chs {
		assert.NotNil(t, ch.channel)
		assert.NotNil(t, ch.done)
	}

	for _, ch := range chs {
		for _, p := range ch.Probers {
			res := p.Probe()
			assert.NotNil(t, res)
			ch.Send(res)
		}
	}

	WatchForAllEvents()
	time.Sleep(200 * time.Millisecond)
	nGoroutine := runtime.NumGoroutine()
	WatchForAllEvents() // only one watch goroutine for each channel
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, nGoroutine, runtime.NumGoroutine())

	// test the dry notification
	SetDryNotify(true)
	assert.True(t, IsDryNotify())
	for _, ch := range chs {
		for _, p := range ch.Probers {
			res := p.Probe()
			assert.NotNil(t, res)
			ch.Send(res)
		}
	}

	AllDone()
}

func TestLevelTrigger(t *testing.T) {
	name := "test"
	SetNotify(name, newDummyNotify("email", "level-trigger", []string{"test"}))
	SetProber(name, newDummyProber("http", "", "level-trigger", []string{"test"}))
	test := GetChannel(name)
	assert.NotNil(t, test)

	gProbeConf := global.ProbeSettings{
		NotificationStrategySettings: global.NotificationStrategySettings{
			Strategy: global.RegularStrategy,
			Factor:   1,
			MaxTimes: 3,
		},
	}
	test.GetProber("level-trigger").Config(gProbeConf)

	gNotifyConf := global.NotifySettings{}
	test.GetNotify("level-trigger").Config(gNotifyConf)

	ConfigAllChannels()
	WatchForAllEvents()

	p := test.GetProber("level-trigger")

	for i := 0; i < 5; i++ {
		res := p.Probe()
		assert.NotNil(t, res)
		assert.Equal(t, probe.StatusDown, res.Status)
		test.Send(res)
		time.Sleep(100 * time.Millisecond)
	}

	AllDone()
}
