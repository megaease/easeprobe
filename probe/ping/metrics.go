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
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/metric"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	PacketsSent *prometheus.CounterVec
	PacketsRecv *prometheus.CounterVec
	PacketLoss  *prometheus.GaugeVec
	MinRtt      *prometheus.GaugeVec
	MaxRtt      *prometheus.GaugeVec
	AvgRtt      *prometheus.GaugeVec
	StdDevRtt   *prometheus.GaugeVec
}

// newMetrics create the metrics
func newMetrics(subsystem, name string) *metrics {
	namespace := global.GetEaseProbe().Name
	return &metrics{
		PacketsSent: metric.NewCounter(namespace, subsystem, name, "sent",
			"Total Package Sent", []string{"name"}),
		PacketsRecv: metric.NewCounter(namespace, subsystem, name, "recv",
			"Total Package Received", []string{"name"}),
		PacketLoss: metric.NewGauge(namespace, subsystem, name, "loss",
			"Package Loss Percentage", []string{"name"}),
		MinRtt: metric.NewGauge(namespace, subsystem, name, "min_rtt",
			"Minimum Round Trip Time", []string{"name"}),
		MaxRtt: metric.NewGauge(namespace, subsystem, name, "max_rtt",
			"Maximum Round Trip Time", []string{"name"}),
		AvgRtt: metric.NewGauge(namespace, subsystem, name, "avg_rtt",
			"Average Round Trip Time", []string{"name"}),
		StdDevRtt: metric.NewGauge(namespace, subsystem, name, "stddev_rtt",
			"Standard Deviation of Round Trip Time", []string{"name"}),
	}
}
