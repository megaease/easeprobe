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

// Disks is the disk usage
type Disks struct {
	base.DefaultProbe
	Mount []string
	Usage []ResourceUsage

	Threshold float64
	metrics   *prometheus.GaugeVec
}

// Name returns the name of the metric
func (d *Disks) Name() string {
	return "disk"
}

// Command returns the command to get the cpu usage
func (d *Disks) Command() string {
	return `df -h ` + strings.Join(d.Mount, " ") + ` 2>/dev/null | awk '(NR>1){printf "%d %d %s %s\n", $3,$2,$5,$6}'`
}

// OutputLines returns the lines of command output
func (d *Disks) OutputLines() int {
	return len(d.Mount)
}

// Config returns the config of the cpu
func (d *Disks) Config(s *Server) {
	if len(s.Disks) == 0 {
		s.Disks = []string{"/"}
	}
	d.Mount = []string{}
	d.Mount = append(d.Mount, s.Disks...)
	d.Usage = make([]ResourceUsage, len(d.Mount))
	if s.Threshold.Disk == 0 {
		s.Threshold.Disk = DefaultDiskThreshold
		log.Debugf("[%s / %s] Disk threshold is not set, using default value: %.2f", s.ProbeKind, s.ProbeName, s.Threshold.Disk)
	}
	d.SetThreshold(&s.Threshold)
	d.CreateMetrics(s.ProbeKind, s.ProbeTag)
}

// SetThreshold set the threshold of the disk
func (d *Disks) SetThreshold(t *Threshold) {
	d.Threshold = t.Disk
}

// Parse a string to a CPU struct
func (d *Disks) Parse(s []string) error {
	if len(s) != d.OutputLines() {
		return fmt.Errorf("invalid disk output")
	}

	for i := 0; i < len(s); i++ {
		disk := strings.Split(s[i], " ")
		if len(disk) < 4 {
			return fmt.Errorf("invalid disk output")
		}
		d.Usage[i] = ResourceUsage{
			Used:  int(strInt(disk[0])),
			Total: int(strInt(disk[1])),
			Usage: strFloat(disk[2][:len(disk[2])-1]),
			Tag:   disk[3],
		}
	}
	return nil
}

// UsageInfo returns the usage info of the disks
func (d *Disks) UsageInfo() string {
	diskUsage := []string{}
	for _, disk := range d.Usage {
		diskUsage = append(diskUsage, fmt.Sprintf("`%s` %.2f%%", disk.Tag, disk.Usage))
	}
	return "Disk: " + strings.Join(diskUsage, ", ")
}

// CheckThreshold check the cpu usage
func (d *Disks) CheckThreshold() (bool, string) {
	lowDisks := []string{}
	for _, disk := range d.Usage {
		if d.Threshold > 0 && d.Threshold <= disk.Usage/100 {
			lowDisks = append(lowDisks, disk.Tag)
		}
	}
	if len(lowDisks) > 0 {
		return false, fmt.Sprintf("Disk Space threshold alert! - [%s]", strings.Join(lowDisks, ", "))
	}
	return true, ""
}

// CreateMetrics create the disk metrics
func (d *Disks) CreateMetrics(subsystem, name string) {
	namespace := global.GetEaseProbe().Name
	d.metrics = metric.NewGauge(namespace, subsystem, name, "disk",
		"Disk Usage", []string{"host", "disk", "state"}, d.Labels)
}

// ExportMetrics export the disk metrics
func (d *Disks) ExportMetrics(name string) {
	for _, disk := range d.Usage {
		d.metrics.With(metric.AddConstLabels(prometheus.Labels{
			"host":  name,
			"disk":  disk.Tag,
			"state": "used",
		}, d.Labels)).Set(float64(disk.Used))

		d.metrics.With(metric.AddConstLabels(prometheus.Labels{
			"host":  name,
			"disk":  disk.Tag,
			"state": "available",
		}, d.Labels)).Set(float64(disk.Total - disk.Used))

		d.metrics.With(metric.AddConstLabels(prometheus.Labels{
			"host":  name,
			"disk":  disk.Tag,
			"state": "total",
		}, d.Labels)).Set(float64(disk.Total))

		d.metrics.With(metric.AddConstLabels(prometheus.Labels{
			"host":  name,
			"disk":  disk.Tag,
			"state": "usage",
		}, d.Labels)).Set(disk.Usage)
	}
}
