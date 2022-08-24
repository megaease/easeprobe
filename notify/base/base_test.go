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
	"bytes"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/megaease/easeprobe/report"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func newDummyResult(name string) probe.Result {
	var total int64 = 100
	up := rand.Int63n(total)
	down := rand.Int63n(total - up)

	r := probe.Result{
		Name:             name,
		Endpoint:         "http://endpoint:8080",
		StartTime:        time.Now(),
		StartTimestamp:   time.Now().Unix(),
		RoundTripTime:    100 * time.Millisecond,
		Status:           probe.StatusUp,
		PreStatus:        probe.StatusDown,
		Message:          "dummy message",
		LatestDownTime:   time.Now(),
		RecoveryDuration: 20 * time.Second,
		Stat: probe.Stat{
			Since: time.Now(),
			Total: total,
			Status: map[probe.Status]int64{
				probe.StatusUp:      up,
				probe.StatusDown:    down,
				probe.StatusUnknown: total - up - down,
			},
			UpTime:   time.Duration(up) * time.Second,
			DownTime: time.Duration(down) * time.Second,
		},
	}
	probe.SetResultData(name, &r)
	return r
}

type ProbeFuncType func() (bool, string)

type dummyProber struct {
	base.DefaultProbe
}

func (d *dummyProber) Config(g global.ProbeSettings) error {
	d.DefaultProbe.Config(g, d.ProbeKind, d.ProbeTag, d.ProbeName, "endpoint", d.DoProbe)
	return nil
}
func (d *dummyProber) DoProbe() (bool, string) {
	return rand.Int()%2 == 0, "hello world"
}

func getProbers() []probe.Prober {
	ps := []probe.Prober{
		newDummyProber("probe1"),
		newDummyProber("probe2"),
		newDummyProber("probe3"),
		newDummyProber("probe4"),
	}
	setResultData(ps)
	return ps
}
func newDummyProber(name string) probe.Prober {
	r := newDummyResult(name)
	return &dummyProber{
		DefaultProbe: base.DefaultProbe{
			ProbeKind:   "dummy",
			ProbeTag:    "tag",
			ProbeName:   name,
			ProbeResult: &r,
		},
	}
}
func setResultData(probes []probe.Prober) {
	for _, p := range probes {
		probe.SetResultData(p.Name(), p.Result())
	}
}

func TestDefaultNotify(t *testing.T) {
	d := DefaultNotify{
		NotifyKind:   "TestKind",
		NotifyFormat: report.Markdown,
		NotifySendFunc: func(string, string) error {
			return nil
		},
		NotifyName:     "TestName",
		NotifyChannels: []string{},
		Dry:            false,
		Timeout:        time.Second * 10,
	}

	d.Config(global.NotifySettings{})

	assert.Equal(t, d.Kind(), "TestKind")
	assert.Equal(t, d.Name(), "TestName")
	assert.Equal(t, d.Channels(), []string{global.DefaultChannelName})
	assert.Equal(t, d.Timeout, time.Second*10)
	assert.Equal(t, d.Retry, global.Retry{
		Times:    global.DefaultRetryTimes,
		Interval: global.DefaultRetryInterval,
	})

	// Dry Notify
	d.Dry = true
	d.Config(global.NotifySettings{})
	r := newDummyResult("dummy")

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	d.Notify(r)
	assert.Contains(t, buf.String(), "[TestKind / TestName / dry_notify]")
	assert.Contains(t, buf.String(), "**dummy Recovery - ( 20s Downtime )** ✅")

	p := getProbers()
	buf.Reset()
	d.NotifyStat(p)
	assert.Contains(t, buf.String(), "[TestKind / TestName / dry_notify]")
	assert.Contains(t, buf.String(), "**Overall SLA Report**")

	// Live Notify
	d.Dry = false

	buf.Reset()
	d.NotifyStat(p)
	assert.Contains(t, buf.String(), "[TestKind / TestName / SLA] - Overall SLA Report - successfully sent!")

	// Nil Notify function
	d.NotifySendFunc = nil
	buf.Reset()
	d.Notify(r)
	assert.Contains(t, buf.String(), "SendFunc is nil")

	logrus.SetOutput(os.Stdout)
}
