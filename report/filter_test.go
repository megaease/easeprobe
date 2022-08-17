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

package report

import (
	"fmt"
	"testing"

	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/stretchr/testify/assert"
)

func initData() []probe.Prober {
	var probes = []probe.Prober{
		&dummyProber{
			DefaultProbe: base.DefaultProbe{
				ProbeKind: "http",
				ProbeName: "probe_001",
				ProbeResult: &probe.Result{
					Endpoint: "http://probe_001",
				},
			},
		},
		&dummyProber{
			DefaultProbe: base.DefaultProbe{
				ProbeKind: "tcp",
				ProbeName: "probe_002",
				ProbeResult: &probe.Result{
					Endpoint: "probe_002:443",
				},
			},
		},
		&dummyProber{
			DefaultProbe: base.DefaultProbe{
				ProbeKind: "http",
				ProbeName: "probe_003",
				ProbeResult: &probe.Result{
					Endpoint: "https://probe_003:443",
				},
			},
		},
		&dummyProber{
			DefaultProbe: base.DefaultProbe{
				ProbeKind: "tcp",
				ProbeName: "probe_004",
				ProbeResult: &probe.Result{
					Endpoint: "probe_004:443",
				},
			},
		},
	}
	probes[0].Result().Status = probe.StatusUp
	probes[1].Result().Status = probe.StatusDown
	probes[2].Result().Status = probe.StatusUp
	probes[3].Result().Status = probe.StatusDown

	for i := 0; i < len(probes); i++ {
		probes[i].Result().Message = fmt.Sprintf("%s %s %s", probes[i].Kind(), probes[i].Name(), "message")
	}

	// 80% SLA
	probes[0].Result().Stat.UpTime = 80
	probes[0].Result().Stat.DownTime = 20

	// 60% SLA
	probes[1].Result().Stat.UpTime = 60
	probes[1].Result().Stat.DownTime = 40

	// 40% SLA
	probes[2].Result().Stat.UpTime = 40
	probes[2].Result().Stat.DownTime = 60

	// 20% SLA
	probes[3].Result().Stat.UpTime = 20
	probes[3].Result().Stat.DownTime = 80

	for _, p := range probes {
		probe.SetResultData(p.Name(), p.Result())
	}

	return probes
}
func TestFilter(t *testing.T) {
	_probes := initData()
	filter := NewEmptyFilter()

	probes := filter.Filter(_probes)
	assert.Equal(t, len(_probes), len(probes))
	assert.Equal(t, _probes, probes)

	filter.PageSize = -1
	err := filter.Check()
	assert.NotNil(t, err)

	filter.PageNum = -1
	err = filter.Check()
	assert.NotNil(t, err)

	filter = NewEmptyFilter()
	// sla >= 60  && sla <= 40
	filter.SLAGreater = 60
	filter.SLALess = 40
	err = filter.Check()
	assert.NotNil(t, err)
	// sla >= 60  && sla <= 200
	filter.SLALess = 200
	err = filter.Check()
	assert.NotNil(t, err)
	// sla >= -100  && sla <= 60
	filter.SLALess = 60
	filter.SLAGreater = -100
	err = filter.Check()
	assert.NotNil(t, err)
	//sla >= 40  && sla <= 60
	filter.SLAGreater = 40
	err = filter.Check()
	assert.Nil(t, err)

	// sla >= 60  && sla <= 40
	probes = filter.Filter(_probes)
	assert.Equal(t, 2, len(probes))
	assert.Equal(t, _probes[1], probes[0])
	assert.Equal(t, _probes[2], probes[1])

	html := filter.HTML()
	assert.Contains(t, html, "<b>SLA</b>: 40.00% - 60.00%")

	// filter by message
	filter = NewEmptyFilter()
	filter.Message = "http"
	probes = filter.Filter(_probes)
	assert.Equal(t, 2, len(probes))
	assert.Equal(t, _probes[0], probes[0])
	assert.Equal(t, _probes[2], probes[1])

	html = filter.HTML()
	assert.Contains(t, html, "<b>Message</b>: http")

	// filter by status
	filter = NewEmptyFilter()
	s := probe.StatusUp
	filter.Status = &s
	probes = filter.Filter(_probes)
	assert.Equal(t, 2, len(probes))
	assert.Equal(t, _probes[0], probes[0])
	assert.Equal(t, _probes[2], probes[1])

	html = filter.HTML()
	assert.Contains(t, html, "<b>Status</b>: up")

	// filter by endpoint
	filter = NewEmptyFilter()
	filter.Endpoint = ":443"
	probes = filter.Filter(_probes)
	assert.Equal(t, 3, len(probes))
	assert.Equal(t, _probes[1], probes[0])
	assert.Equal(t, _probes[2], probes[1])
	assert.Equal(t, _probes[3], probes[2])

	html = filter.HTML()
	assert.Contains(t, html, "<b>Endpoint</b>: :443")

	// filter by kind
	filter = NewEmptyFilter()
	filter.Kind = "tcp"
	probes = filter.Filter(_probes)
	assert.Equal(t, 2, len(probes))
	assert.Equal(t, _probes[1], probes[0])
	assert.Equal(t, _probes[3], probes[1])

	html = filter.HTML()
	assert.Contains(t, html, "<b>Kind</b>: tcp")

	// filter by name
	filter = NewEmptyFilter()
	filter.Name = "probe_001"
	probes = filter.Filter(_probes)
	assert.Equal(t, 1, len(probes))
	assert.Equal(t, _probes[0], probes[0])

	html = filter.HTML()
	assert.Contains(t, html, "<b>Name</b>: probe_001")

	// filter by name and kind
	filter = NewEmptyFilter()
	filter.Name = "probe"
	filter.Kind = "http"
	probes = filter.Filter(_probes)
	assert.Equal(t, 2, len(probes))
	assert.Equal(t, _probes[0], probes[0])
	assert.Equal(t, _probes[2], probes[1])

	// filter by status and kind
	filter = NewEmptyFilter()
	filter.Status = &s
	filter.Kind = "tcp"
	probes = filter.Filter(_probes)
	assert.Equal(t, 0, len(probes))

	filter.Kind = "http"
	probes = filter.Filter(_probes)
	assert.Equal(t, 2, len(probes))
	assert.Equal(t, _probes[0], probes[0])
	assert.Equal(t, _probes[2], probes[1])

	// filter by status and sla
	filter = NewEmptyFilter()
	filter.Status = &s
	filter.SLAGreater = 20
	filter.SLALess = 100
	probes = filter.Filter(_probes)
	assert.Equal(t, 2, len(probes))
	assert.Equal(t, _probes[0], probes[0])
	assert.Equal(t, _probes[2], probes[1])

	// filter by status and sla and endpoint
	filter.Endpoint = ":443"
	probes = filter.Filter(_probes)
	assert.Equal(t, 1, len(probes))
	assert.Equal(t, _probes[2], probes[0])
}

func NewFilterWithPage(pg, sz int) *SLAFilter {
	filter := NewEmptyFilter()
	filter.PageNum = pg
	filter.PageSize = sz
	return filter
}
func TestPage(t *testing.T) {
	_probes := initData()
	filter := NewEmptyFilter()
	probes := filter.Filter(_probes)
	assert.Equal(t, _probes, probes)

	filter = NewFilterWithPage(1, 2)
	probes = filter.Filter(_probes)
	assert.Equal(t, 2, len(probes))
	assert.Equal(t, _probes[0], probes[0])
	assert.Equal(t, _probes[1], probes[1])

	filter = NewFilterWithPage(2, 2)
	probes = filter.Filter(_probes)
	assert.Equal(t, 2, len(probes))
	assert.Equal(t, _probes[2], probes[0])
	assert.Equal(t, _probes[3], probes[1])

	filter = NewFilterWithPage(2, 3)
	probes = filter.Filter(_probes)
	assert.Equal(t, 1, len(probes))
	assert.Equal(t, _probes[3], probes[0])
}
