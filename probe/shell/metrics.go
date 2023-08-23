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
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/metric"
	"github.com/prometheus/client_golang/prometheus"
)

// metrics is the metrics for shell probe
type metrics struct {
	ExitCode  *prometheus.CounterVec
	OutputLen *prometheus.GaugeVec
}

// newMetrics create the shell metrics
func newMetrics(subsystem, name string, constLabels prometheus.Labels) *metrics {
	namespace := global.GetEaseProbe().Name
	return &metrics{
		ExitCode: metric.NewCounter(namespace, subsystem, name, "exit_code",
			"Exit Code", []string{"name", "exit", "endpoint"}, constLabels),
		OutputLen: metric.NewGauge(namespace, subsystem, name, "output_len",
			"Output Length", []string{"name", "exit", "endpoint"}, constLabels),
	}
}
