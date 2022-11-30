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

package host

import (
	"fmt"
	"strings"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/metric"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Mem is the resource usage for memory and disk
type Mem struct {
	ResourceUsage `yaml:",inline"`

	Threshold float64 `yaml:"threshold"`
	metrics   *prometheus.GaugeVec
}

// Name returns the name of the metric
func (m *Mem) Name() string {
	return "mem"
}

// Command returns the command to get the memory usage
func (m *Mem) Command() string {
	return `free -m | awk 'NR==2{printf "%s %s %.2f\n", $3,$2,$3*100/$2 }'`
}

// OutputLines returns the lines of command output
func (m *Mem) OutputLines() int {
	return 1
}

// Config returns the config of the memory
func (m *Mem) Config(s *Server) {
	if s.Threshold.Mem == 0 {
		s.Threshold.Mem = DefaultMemThreshold
		log.Debugf("[%s / %s] Memory threshold is not set, using default value: %.2f", s.ProbeKind, s.ProbeName, s.Threshold.Mem)
	}
	m.SetThreshold(&s.Threshold)
	m.CreateMetrics(s.ProbeKind, s.ProbeTag)
}

// SetThreshold set the threshold of the memory
func (m *Mem) SetThreshold(t *Threshold) {
	m.Threshold = t.Mem
}

// Parse a string to a Memory struct
func (m *Mem) Parse(s []string) error {
	if len(s) < m.OutputLines() {
		return fmt.Errorf("invalid memory output")
	}
	mem := strings.Split(s[0], " ")
	if len(mem) < 3 {
		return fmt.Errorf("invalid memory output")
	}
	m.Used = int(strInt(mem[0]))
	m.Total = int(strInt(mem[1]))
	m.Usage = strFloat(mem[2])
	return nil
}

// UsageInfo returns the usage info of the memory
func (m *Mem) UsageInfo() string {
	return fmt.Sprintf("Memory: %.2f%%", m.Usage)
}

// CheckThreshold check the memory usage
func (m *Mem) CheckThreshold() (bool, string) {
	if m.Threshold > 0 && m.Threshold <= m.Usage/100 {
		return false, "Memory threshold alert!"
	}
	return true, ""
}

// CreateMetrics create the memory metrics
func (m *Mem) CreateMetrics(subsystem, name string) {
	namespace := global.GetEaseProbe().Name
	m.metrics = metric.NewGauge(namespace, subsystem, name, "memory",
		"Memory Usage", []string{"host", "state"})
}

// ExportMetrics export the memory metrics
func (m *Mem) ExportMetrics(name string) {
	m.metrics.With(prometheus.Labels{
		"host":  name,
		"state": "used",
	}).Set(float64(m.Used))

	m.metrics.With(prometheus.Labels{
		"host":  name,
		"state": "available",
	}).Set(float64(m.Total - m.Used))

	m.metrics.With(prometheus.Labels{
		"host":  name,
		"state": "total",
	}).Set(float64(m.Total))

	m.metrics.With(prometheus.Labels{
		"host":  name,
		"state": "usage",
	}).Set(m.Usage)
}
