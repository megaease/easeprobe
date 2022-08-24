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

package shell

import (
	"math/rand"
	"testing"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/report"
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

func TestShell(t *testing.T) {
	conf := &NotifyConfig{}
	conf.Cmd = "echo"
	err := conf.Config(global.NotifySettings{})
	assert.NoError(t, err)
	assert.Equal(t, "shell", conf.Kind())
	assert.Equal(t, report.Shell, conf.NotifyFormat)

	r := newDummyResult("dummy")
	msg := report.ToShell(r)
	err = conf.RunShell("title", msg)
	assert.NoError(t, err)

	conf.CleanEnv = true
	err = conf.RunShell("title", msg)
	assert.NoError(t, err)

	conf.Cmd = "bad-command"
	err = conf.RunShell("title", msg)
	assert.Error(t, err)

	err = conf.RunShell("title", "{bad:json:format}")
	assert.Error(t, err)
}
