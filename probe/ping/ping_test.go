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

package ping

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/base"
	ping "github.com/prometheus-community/pro-bing"
	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	ping := Ping{
		DefaultProbe: base.DefaultProbe{ProbeName: "dummy ping"},
		Host:         "127.0.0.1",
	}
	ping.Config(global.ProbeSettings{})
	assert.Equal(t, "ping", ping.ProbeKind)
	assert.Equal(t, DefaultPingCount, ping.Count)
	assert.Equal(t, DefaultLostThreshold, ping.LostThreshold)

	ping = Ping{
		DefaultProbe:  base.DefaultProbe{ProbeName: "dummy ping"},
		Host:          "127.0.0.1",
		Count:         0,
		LostThreshold: -1,
	}
	ping.Config(global.ProbeSettings{})
	assert.Equal(t, "ping", ping.ProbeKind)
	assert.Equal(t, DefaultPingCount, ping.Count)
	assert.Equal(t, DefaultLostThreshold, ping.LostThreshold)

	if runtime.GOOS == "windows" {
		ping.Privileged = true
	}
	s, m := ping.DoProbe()
	assert.True(t, s)
	assert.Contains(t, m, "Succeeded")
}

func TestPingWithInvalidHost(t *testing.T) {

	p := Ping{
		DefaultProbe:  base.DefaultProbe{ProbeName: "dummy ping"},
		Host:          "127.0.0.1",
		LostThreshold: 0.5,
	}
	p.Config(global.ProbeSettings{})

	var pinger *ping.Pinger
	monkey.PatchInstanceMethod(reflect.TypeOf(pinger), "Statistics", func(_ *ping.Pinger) *ping.Statistics {
		return &ping.Statistics{PacketLoss: 51}
	})
	if runtime.GOOS == "windows" {
		p.Privileged = true
	}
	s, m := p.DoProbe()
	assert.False(t, s)
	assert.Contains(t, m, "Failed")

	monkey.PatchInstanceMethod(reflect.TypeOf(pinger), "Run", func(_ *ping.Pinger) error {
		return fmt.Errorf("ping error")
	})
	s, m = p.DoProbe()
	assert.False(t, s)
	assert.Contains(t, m, "ping error")

	p = Ping{
		DefaultProbe: base.DefaultProbe{ProbeName: "dummy ping"},
		Host:         "unknown",
	}
	err := p.Config(global.ProbeSettings{})
	assert.Error(t, err)
	if runtime.GOOS == "windows" {
		p.Privileged = true
	}
	s, m = p.DoProbe()
	assert.False(t, s)
	assert.Contains(t, m, "lookup unknown")
}
