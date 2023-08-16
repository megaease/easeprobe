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

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/metric"
	"github.com/megaease/easeprobe/probe/base"
)

// CPU is the cpu usage
// "1.6 us,  1.6 sy,  3.2 ni, 91.9 id,  1.6 wa,  0.0 hi,  0.0 si,  0.0 st"
type CPU struct {
	base.DefaultProbe `yaml:",inline"`
	User              float64 `yaml:"user"`
	Sys               float64 `yaml:"sys"`
	Nice              float64 `yaml:"nice"`
	Idle              float64 `yaml:"idle"`
	Wait              float64 `yaml:"wait"`
	Hard              float64 `yaml:"hard"`
	Soft              float64 `yaml:"soft"`
	Steal             float64 `yaml:"steal"`

	Threshold float64 `yaml:"threshold"`
	metrics   *prometheus.GaugeVec
}

// Name returns the name of the metric
func (c *CPU) Name() string {
	return "cpu"
}

// Command returns the command to get the cpu usage
func (c *CPU) Command() string {
	return `top -b -n 1 | grep Cpu | awk -F ":" '{print $2}'`
}

// OutputLines returns the lines of command output
func (c *CPU) OutputLines() int {
	return 1
}

// Config returns the config of the cpu
func (c *CPU) Config(s *Server) {
	if s.Threshold.CPU == 0 {
		s.Threshold.CPU = DefaultCPUThreshold
		log.Debugf("[%s / %s] CPU threshold is not set, using default value: %.2f", s.ProbeKind, s.ProbeName, s.Threshold.CPU)
	}
	c.SetThreshold(&s.Threshold)
	c.CreateMetrics(s.ProbeKind, s.ProbeTag)
}

// SetThreshold set the cpu threshold
func (c *CPU) SetThreshold(t *Threshold) {
	c.Threshold = t.CPU
}

// Parse a string to a CPU struct
func (c *CPU) Parse(s []string) error {
	if len(s) < c.OutputLines() {
		return fmt.Errorf("invalid cpu output")
	}
	cpu := strings.Split(s[0], ",")
	if len(cpu) < 8 {
		return fmt.Errorf("invalid cpu output")
	}
	c.User = strFloat(first(cpu[0]))
	c.Sys = strFloat(first(cpu[1]))
	c.Nice = strFloat(first(cpu[2]))
	c.Idle = strFloat(first(cpu[3]))
	c.Wait = strFloat(first(cpu[4]))
	c.Hard = strFloat(first(cpu[5]))
	c.Soft = strFloat(first(cpu[6]))
	c.Steal = strFloat(first(cpu[7]))
	return nil
}

// UsageInfo returns the cpu usage info
func (c *CPU) UsageInfo() string {
	return fmt.Sprintf("CPU: %.2f%%", (100 - c.Idle))
}

// CheckThreshold check the cpu usage
func (c *CPU) CheckThreshold() (bool, string) {
	if c.Threshold > 0 && c.Threshold <= (100-c.Idle)/100 {
		return false, "CPU threshold alert!"
	}
	return true, ""
}

// CreateMetrics create the cpu metrics
func (c *CPU) CreateMetrics(subsystem, name string) {
	namespace := global.GetEaseProbe().Name
	c.metrics = metric.NewGauge(namespace, subsystem, name, "cpu",
		"CPU Usage", []string{"host", "state"}, c.Labels)
}

// ExportMetrics export the cpu metrics
func (c *CPU) ExportMetrics(name string) {
	// CPU metrics
	c.metrics.With(metric.AddConstLabels(prometheus.Labels{
		"host":  name,
		"state": "usage",
	}, c.Labels)).Set(100 - c.Idle)

	c.metrics.With(metric.AddConstLabels(prometheus.Labels{
		"host":  name,
		"state": "idle",
	}, c.Labels)).Set(c.Idle)

	c.metrics.With(metric.AddConstLabels(prometheus.Labels{
		"host":  name,
		"state": "user",
	}, c.Labels)).Set(c.User)

	c.metrics.With(metric.AddConstLabels(prometheus.Labels{
		"host":  name,
		"state": "sys",
	}, c.Labels)).Set(c.Sys)

	c.metrics.With(metric.AddConstLabels(prometheus.Labels{
		"host":  name,
		"state": "nice",
	}, c.Labels)).Set(c.Nice)

	c.metrics.With(metric.AddConstLabels(prometheus.Labels{
		"host":  name,
		"state": "wait",
	}, c.Labels)).Set(c.Wait)

	c.metrics.With(metric.AddConstLabels(prometheus.Labels{
		"host":  name,
		"state": "hard",
	}, c.Labels)).Set(c.Hard)

	c.metrics.With(metric.AddConstLabels(prometheus.Labels{
		"host":  name,
		"state": "soft",
	}, c.Labels)).Set(c.Soft)

	c.metrics.With(metric.AddConstLabels(prometheus.Labels{
		"host":  name,
		"state": "steal",
	}, c.Labels)).Set(c.Steal)
}
