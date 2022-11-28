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
)

// Load is the load average of the host
type Load struct {
	Core      int64              `json:"core"`
	Metrics   map[string]float64 `json:"metrics"`
	Threshold map[string]float64
}

// Command returns the command to get the cpu usage
func (l *Load) Command() string {
	return `grep -c ^processor /proc/cpuinfo;` + "\n" +
		`cat /proc/loadavg | awk '{print $1,$2,$3}';`
}

// OutputLines returns the lines of command output
func (l *Load) OutputLines() int {
	return 2
}

// Config returns the config of the cpu
func (l *Load) Config(s *Server) error {
	l.Metrics = make(map[string]float64)
	if s.Threshold.Load == nil {
		s.Threshold.Load = make(map[string]float64)
		s.Threshold.Load["m1"] = DefaultLoadThreshold
		s.Threshold.Load["m5"] = DefaultLoadThreshold
		s.Threshold.Load["m15"] = DefaultLoadThreshold
		log.Debugf("[%s / %s] All of load average threshold is not set, use default value: %.2f", s.ProbeKind, s.ProbeName, DefaultLoadThreshold)
		l.SetThreshold(&s.Threshold)
		return nil
	}

	for k, v := range s.Threshold.Load {
		s.Threshold.Load[strings.ToLower(k)] = v
	}
	if _, ok := s.Threshold.Load["m1"]; !ok {
		s.Threshold.Load["m1"] = DefaultLoadThreshold
		log.Debugf("[%s / %s] Load average threshold for m1 is not set, use default value: %.2f", s.ProbeKind, s.ProbeName, s.Threshold.Load["m1"])
	}
	if _, ok := s.Threshold.Load["m5"]; !ok {
		s.Threshold.Load["m5"] = DefaultLoadThreshold
		log.Debugf("[%s / %s] Load average threshold for m5 is not set, use default value: %.2f", s.ProbeKind, s.ProbeName, s.Threshold.Load["m5"])
	}
	if _, ok := s.Threshold.Load["m15"]; !ok {
		s.Threshold.Load["m15"] = DefaultLoadThreshold
		log.Debugf("[%s / %s] Load average threshold for m15 is not set, use default value: %.2f", s.ProbeKind, s.ProbeName, s.Threshold.Load["m15"])
	}
	l.SetThreshold(&s.Threshold)
	return nil
}

// SetThreshold  set the threshold of the load
func (l *Load) SetThreshold(t *Threshold) {
	l.Threshold = t.Load
}

// Parse a string to a CPU struct
func (l *Load) Parse(s []string) error {
	if len(s) < l.OutputLines() {
		return fmt.Errorf("invalid load average output")
	}
	l.Core = strInt(s[0])
	load := strings.Split(s[1], " ")
	if len(load) < 3 {
		return fmt.Errorf("invalid load average output")
	}
	l.Metrics["m1"] = strFloat(load[0])
	l.Metrics["m5"] = strFloat(load[1])
	l.Metrics["m15"] = strFloat(load[2])
	return nil
}

// UsageInfo returns the usage info of the load
func (l *Load) UsageInfo() string {
	loadAvg := []string{}
	for _, load := range l.Metrics {
		loadAvg = append(loadAvg, fmt.Sprintf("%.2f", load))
	}
	return "Load: " + strings.Join(loadAvg, "/")
}

// CheckThreshold check the cpu usage
func (l *Load) CheckThreshold() (bool, string) {
	for k, v := range l.Metrics {
		// normalize the load average to 1 cpu core
		if v/float64(l.Core) > l.Threshold[k] {
			return false, fmt.Sprintf("Load Average %s High! - %.2f", k, v)
		}
	}
	return true, ""
}

// ExportMetrics export the cpu metrics
func (l *Load) ExportMetrics(name string, g *prometheus.GaugeVec) {
	g.With(prometheus.Labels{
		"host":  name,
		"state": "m1",
	}).Set(l.Metrics["m1"])

	g.With(prometheus.Labels{
		"host":  name,
		"state": "m5",
	}).Set(l.Metrics["m5"])

	g.With(prometheus.Labels{
		"host":  name,
		"state": "m15",
	}).Set(l.Metrics["m15"])
}
