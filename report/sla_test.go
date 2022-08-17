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
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/stretchr/testify/assert"
)

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

func TestSLA(t *testing.T) {
	global.InitEaseProbe("DummyProbe", "icon")
	probes := getProbers()
	probes[0].Result().Status = probe.StatusDown
	probe.SetResultData(probes[0].Name(), probes[0].Result())
	for f, fn := range FormatFuncs {
		sla := fn.StatFn(probes)
		assert.NotEmpty(t, sla)
		if f == SMS {
			assert.Contains(t, sla, "Total "+fmt.Sprintf("%d", len(probes))+" Services")
			continue
		}
		for _, p := range probes {
			assert.Contains(t, sla, p.Name())
		}
	}
}

func TestSLAJSONSection(t *testing.T) {
	p := newDummyProber("probe1")
	sla := SLAJSONSection(p.Result())
	assert.NotEmpty(t, sla)
	assert.Contains(t, sla, "\"name\":\"probe1\"")
	assert.Contains(t, sla, "\"status\":\"up\"")

	monkey.Patch(json.Marshal, func(v any) ([]byte, error) {
		return nil, fmt.Errorf("error")
	})
	sla = SLAJSONSection(p.Result())
	assert.Empty(t, sla)

	sla = SLAJSON([]probe.Prober{p})
	assert.Empty(t, sla)

	monkey.UnpatchAll()
}

func TestSLAStatusText(t *testing.T) {
	p := newDummyProber("probe1")
	str := SLAStatusText(p.Probe().Stat, MarkdownSocial)
	assert.Contains(t, str, "`")
	str = SLAStatusText(p.Probe().Stat, Markdown)
	assert.Contains(t, str, "`")
	assert.Contains(t, str, "**")
	str = SLAStatusText(p.Probe().Stat, HTML)
	assert.Contains(t, str, "<b>")
	str = SLAStatusText(p.Probe().Stat, Log)
	assert.NotContains(t, str, "<b>")
	assert.NotContains(t, str, "`")
	assert.NotContains(t, str, "**")
}

func TestFailed(t *testing.T) {
	probes := getProbers()
	var w *csv.Writer
	monkey.PatchInstanceMethod(reflect.TypeOf(w), "WriteAll", func(_ *csv.Writer, _ [][]string) error {
		return fmt.Errorf("error")
	})
	sla := SLACSV(probes)
	assert.Empty(t, sla)

	monkey.Patch(json.Marshal, func(v any) ([]byte, error) {
		return nil, fmt.Errorf("error")
	})
	sla = SLAShell(probes)
	assert.Empty(t, sla)

	monkey.UnpatchAll()
}

func TestSLAFilter(t *testing.T) {
	probes := getProbers()
	probes[0].Result().Status = probe.StatusUp
	probes[1].Result().Status = probe.StatusDown
	probes[2].Result().Status = probe.StatusUp
	probes[3].Result().Status = probe.StatusDown
	setResultData(probes)

	html := SLAHTMLFilter(probes, nil)
	for _, p := range probes {
		assert.Contains(t, html, p.Name())
	}

	filter := NewEmptyFilter()
	html = SLAHTMLFilter(probes, filter)
	for _, p := range probes {
		assert.Contains(t, html, p.Name())
	}

	status := probe.StatusUp
	filter.Status = &status
	html = SLAHTMLFilter(probes, filter)
	assert.Contains(t, html, probes[0].Name())
	assert.NotContains(t, html, probes[1].Name())
	assert.Contains(t, html, probes[2].Name())
	assert.NotContains(t, html, probes[3].Name())

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

	setResultData(probes)

	// sla between 50 - 90, status is up
	filter.SLAGreater = 50
	filter.SLALess = 90
	html = SLAHTMLFilter(probes, filter)
	assert.Contains(t, html, probes[0].Name())
	assert.NotContains(t, html, probes[1].Name())
	assert.NotContains(t, html, probes[2].Name())
	assert.NotContains(t, html, probes[3].Name())

}
