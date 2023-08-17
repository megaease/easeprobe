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

package conf

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/megaease/easeprobe/probe/http"
	"github.com/megaease/easeprobe/probe/shell"
	"github.com/megaease/easeprobe/probe/tcp"
)

func TestMergeConstLabels(t *testing.T) {

	ps := []probe.Prober{
		&http.HTTP{
			DefaultProbe: base.DefaultProbe{
				Labels: prometheus.Labels{"service": "service_a"},
			},
		},
		&tcp.TCP{
			DefaultProbe: base.DefaultProbe{
				Labels: prometheus.Labels{"host": "host_b"},
			},
		},
		&shell.Shell{},
	}

	MergeConstLabels(ps)

	assert.Equal(t, 2, len(ps[0].LabelMap()))
	assert.Equal(t, "service_a", ps[0].LabelMap()["service"])
	assert.Equal(t, "", ps[0].LabelMap()["host"])

	assert.Equal(t, 2, len(ps[1].LabelMap()))
	assert.Equal(t, "", ps[1].LabelMap()["service"])
	assert.Equal(t, "host_b", ps[1].LabelMap()["host"])

	assert.Equal(t, 2, len(ps[2].LabelMap()))
	assert.Equal(t, "", ps[2].LabelMap()["service"])
	assert.Equal(t, "", ps[2].LabelMap()["host"])
}
