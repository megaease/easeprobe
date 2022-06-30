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

// metrics is the metrics for http probe
type metrics struct {
	StatusCode       *prometheus.CounterVec
	ContentLen       *prometheus.GaugeVec
	DNSDuration      *prometheus.GaugeVec
	ConnectDuration  *prometheus.GaugeVec
	TLSDuration      *prometheus.GaugeVec
	SendDuration     *prometheus.GaugeVec
	WaitDuration     *prometheus.GaugeVec
	TransferDuration *prometheus.GaugeVec
	TotalDuration    *prometheus.GaugeVec
}

// newMetrics create the HTTP metrics
func newMetrics(subsystem, name string) *metrics {
	namespace := global.GetEaseProbe().Name
	return &metrics{
		StatusCode: metric.NewCounter(namespace, subsystem, name, "status_code",
			"HTTP Status Code", []string{"name", "status"}),
		ContentLen: metric.NewGauge(namespace, subsystem, name, "content_len",
			"HTTP Content Length", []string{"name", "status"}),
		DNSDuration: metric.NewGauge(namespace, subsystem, name, "dns_duration",
			"DNS Duration", []string{"name", "status"}),
		ConnectDuration: metric.NewGauge(namespace, subsystem, name, "connect_duration",
			"TCP Connection Duration", []string{"name", "status"}),
		TLSDuration: metric.NewGauge(namespace, subsystem, name, "tls_duration",
			"TLS Duration", []string{"name", "status"}),
		SendDuration: metric.NewGauge(namespace, subsystem, name, "send_duration",
			"Send Duration", []string{"name", "status"}),
		WaitDuration: metric.NewGauge(namespace, subsystem, name, "wait_duration",
			"Wait Duration", []string{"name", "status"}),
		TransferDuration: metric.NewGauge(namespace, subsystem, name, "transfer_duration",
			"Transfer Duration", []string{"name", "status"}),
		TotalDuration: metric.NewGauge(namespace, subsystem, name, "total_duration",
			"Total Duration", []string{"name", "status"}),
	}
}
