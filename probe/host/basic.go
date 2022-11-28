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

	"github.com/prometheus/client_golang/prometheus"
)

// Basic is the basic information of the host
type Basic struct {
	HostName string `yaml:"hostname"`
	OS       string `yaml:"os"`
	Core     int64  `yaml:"core"`
}

// Command returns the command to get the cpu usage
func (b *Basic) Command() string {
	return `hostname;` + "\n" +
		`awk -F= '/^NAME/{print $2}' /etc/os-release | tr -d '\"';` + "\n" +
		`grep -c ^processor /proc/cpuinfo;`
}

// OutputLines returns the lines of command output
func (b *Basic) OutputLines() int {
	return 3
}

// Config returns the config of the basic info
func (b *Basic) Config(s *Server) error {
	b.SetThreshold(&s.Threshold)
	return nil
}

// SetThreshold set the basic threshold
func (b *Basic) SetThreshold(t *Threshold) {
}

// Parse a string to a CPU struct
func (b *Basic) Parse(s []string) error {
	if len(s) < b.OutputLines() {
		return fmt.Errorf("invalid basic output")
	}
	b.HostName = s[0]
	b.OS = s[1]
	b.Core = strInt(s[2])
	return nil
}

// UsageInfo returns the usage info of the cpu
func (b *Basic) UsageInfo() string {
	return ""
}

// CheckThreshold check the cpu usage
func (b *Basic) CheckThreshold() (bool, string) {
	return true, ""
}

// ExportMetrics export the cpu metrics
func (b *Basic) ExportMetrics(name string, g *prometheus.GaugeVec) {
	g.With(prometheus.Labels{
		"host":  name,
		"state": "cpu_core",
	}).Set(float64(b.Core))
}
