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
	"math/rand"
	"testing"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/stretchr/testify/assert"
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

}
