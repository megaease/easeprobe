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
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/metric"
	"github.com/prometheus/client_golang/prometheus"
)

// metrics is the probe metrics
type metrics struct {
	TotalCnt  *prometheus.GaugeVec
	TotalTime *prometheus.GaugeVec
	Duration  *prometheus.GaugeVec
	Status    *prometheus.GaugeVec
	SLA       *prometheus.GaugeVec
}

// newMetrics create the metrics
func newMetrics(subsystem, name string) *metrics {
	namespace := global.GetEaseProbe().Name
	return &metrics{
		TotalCnt: metric.NewGauge(namespace, subsystem, name, "total",
			"Total Probed Counts", []string{"name", "status"}),
		TotalTime: metric.NewGauge(namespace, subsystem, name, "total_time",
			"Total Time(Seconds) of Status", []string{"name", "status"}),
		Duration: metric.NewGauge(namespace, subsystem, name, "duration",
			"Probe Duration", []string{"name", "status"}),
		Status: metric.NewGauge(namespace, subsystem, name, "status",
			"Probe Status", []string{"name"}),
		SLA: metric.NewGauge(namespace, subsystem, name, "sla",
			"Probe SLA", []string{"name"}),
	}
}
