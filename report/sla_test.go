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
	"math/rand"
	"testing"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/stretchr/testify/assert"
)

type ProbeFuncType func() (bool, string)

type dummyProber struct {
	base.DefaultOptions
}

func (d *dummyProber) Config(g global.ProbeSettings) error {
	d.DefaultOptions.Config(g, d.ProbeKind, d.ProbeTag, d.ProbeName, "endpoint", d.DoProbe)
	return nil
}
func (d *dummyProber) DoProbe() (bool, string) {
	return rand.Int()%2 == 0, "hello world"
}

var probes = []probe.Prober{
	newDummyProber("probe1"),
	newDummyProber("probe2"),
	newDummyProber("probe3"),
	newDummyProber("probe4"),
}

func newDummyProber(name string) probe.Prober {
	r := newDummyResult(name)
	return &dummyProber{
		DefaultOptions: base.DefaultOptions{
			ProbeKind:   "dummy",
			ProbeTag:    "tag",
			ProbeName:   name,
			ProbeResult: &r,
		},
	}
}

func TestSLA(t *testing.T) {
	global.InitEaseProbe("DummyProbe", "icon")
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
