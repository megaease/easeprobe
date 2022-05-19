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

package http

import (
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/metric"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics is the metrics for http probe
type Metrics struct {
	StatusCode *prometheus.CounterVec
	ContentLen *prometheus.GaugeVec
}

// NewMetrics create the HTTP metrics
func NewMetrics(subsystem, name string) *Metrics {
	namespace := global.GetEaseProbe().Name
	return &Metrics{
		StatusCode: metric.NewCounter(namespace, subsystem, name, "status_code",
			"HTTP Status Code", []string{"name", "status"}),
		ContentLen: metric.NewGauge(namespace, subsystem, name, "content_len",
			"HTTP Content Length", []string{"name", "status"}),
	}
}
