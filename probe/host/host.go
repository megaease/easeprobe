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

// Package host is the host probe package
package host

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/ssh"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Default threshold
const (
	DefaultCPUThreshold  = 0.8
	DefaultMemThreshold  = 0.8
	DefaultDiskThreshold = 0.95
	DefaultLoadThreshold = 0.8
)

// Threshold is the threshold of a probe
type Threshold struct {
	CPU  float64            `yaml:"cpu,omitempty" json:"cpu,omitempty" jsonschema:"title=CPU threshold,description=CPU threshold (default: 0.8)"`
	Mem  float64            `yaml:"mem,omitempty" json:"mem,omitempty" jsonschema:"title=Memory threshold,description=Memory threshold (default: 0.8)"`
	Disk float64            `yaml:"disk,omitempty" json:"disk,omitempty" jsonschema:"title=Disk threshold,description=Disk threshold (default: 0.95)"`
	Load map[string]float64 `yaml:"load,omitempty" json:"load,omitempty" jsonschema:"title=Load average threshold,description=Load Average M1/M5/M15 threshold (default: 0.8)"`
}

func (t *Threshold) String() string {
	load := []string{}
	for _, v := range t.Load {
		load = append(load, fmt.Sprintf("%.2f", v))
	}

	return fmt.Sprintf("CPU: %.2f, Mem: %.2f, Disk: %.2f, Load: %s", t.CPU, t.Mem, t.Disk, strings.Join(load, "/"))
}

// Server is the server of a host probe
type Server struct {
	ssh.Server `yaml:",inline"`
	Threshold  Threshold `yaml:"threshold,omitempty" json:"threshold,omitempty" jsonschema:"title=Threshold,description=the threshold of the probe for cpu/memory/disk"`
	Disks      []string  `yaml:"disks,omitempty" json:"disks,omitempty" jsonschema:"title=Disks,description=the disks to be monitored,example=[\"/\", \"/data\"]"`
	metrics    *metrics  `yaml:"-" json:"-"`
}

// Host is the host probe configuration
type Host struct {
	Bastion *ssh.BastionMapType `yaml:"bastion,omitempty" json:"bastion,omitempty" jsonschema:"title=Bastion Servers,description=the bastion server for ssh login"`
	Servers []Server            `yaml:"servers" json:"servers" jsonschema:"required,title=Host Servers,description=the host servers to be monitored"`
}

// BastionMap is a map of bastion
var BastionMap ssh.BastionMapType

// Config is the host probe configuration
func (s *Server) Config(gConf global.ProbeSettings) error {
	kind := "host"
	tag := "server"
	name := s.ProbeName

	// The following commands are:
	// 1. retrieve the hostname: 	`hostname``
	// 2. retrieve the os:  			`awk -F= '/^NAME/{print $2}' /etc/os-release | tr -d '\"'`
	// 3. retrieve the memory usage:	`free -m | awk 'NR==2{printf "%s %s %.2f\n", $3,$2,$3*100/$2 }'`
	//    output: used(MB) total(MB) usage(%), example: 19379 31654 61.22
	// 4. retrieve the cpu core:		`grep -c ^processor /proc/cpuinfo;`
	// 5. retrieve the cpu usage:	`top -b -n 1 | grep Cpu | awk -F ":" '{print $2}'`
	//    output example: 1.6 us,  0.0 sy,  0.0 ni, 98.4 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st
	// 6. retrieve the load average:	`cat /proc/loadavg | awk '{print $1,$2,$3}'`
	// 7. retrieve the disk usage	`df -h / 2>/dev/null | awk '(NR>1){printf "%d %d %s %s\n", $3,$2,$5,$6}'`
	//    output: used(GB) total(GB) usage(%) disk, example: 40 970 5% /

	s.Command = `hostname;
	awk -F= '/^NAME/{print $2}' /etc/os-release | tr -d '\"';
	free -m | awk 'NR==2{printf "%s %s %.2f\n", $3,$2,$3*100/$2 }';
	grep -c ^processor /proc/cpuinfo;
	top -b -n 1 | grep Cpu | awk -F ":" '{print $2}';
	cat /proc/loadavg | awk '{print $1,$2,$3}';` + "\n"

	if len(s.Disks) == 0 {
		s.Disks = []string{"/"}
	}

	s.Command += "\t" + `df -h ` + strings.Join(s.Disks, " ") + ` 2>/dev/null | awk '(NR>1){printf "%d %d %s %s\n", $3,$2,$5,$6}'`

	if s.Threshold.CPU == 0 {
		s.Threshold.CPU = DefaultCPUThreshold
		log.Debugf("[%s / %s] CPU threshold is not set, use default value: %.2f", s.ProbeKind, s.ProbeName, s.Threshold.CPU)
	}
	if s.Threshold.Mem == 0 {
		s.Threshold.Mem = DefaultMemThreshold
		log.Debugf("[%s / %s] Memory threshold is not set, use default value: %.2f", s.ProbeKind, s.ProbeName, s.Threshold.Mem)
	}
	if s.Threshold.Disk == 0 {
		s.Threshold.Disk = DefaultDiskThreshold
		log.Debugf("[%s / %s] Disk threshold is not set, use default value: %.2f", s.ProbeKind, s.ProbeName, s.Threshold.Disk)
	}
	if s.Threshold.Load == nil {
		s.Threshold.Load = make(map[string]float64)
		s.Threshold.Load["m1"] = DefaultLoadThreshold
		s.Threshold.Load["m5"] = DefaultLoadThreshold
		s.Threshold.Load["m15"] = DefaultLoadThreshold
		log.Debugf("[%s / %s] All of load average threshold is not set, use default value: %.2f", s.ProbeKind, s.ProbeName, DefaultLoadThreshold)
	} else {
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
	}

	s.metrics = newMetrics(kind, tag)

	endpoint := s.Threshold.String()
	err := s.Configure(gConf, kind, tag, name, endpoint, &BastionMap, s.DoProbe)
	log.Debugf("[%s / %s] configuration: %+v", s.ProbeKind, s.ProbeName, *s)
	return err
}

// DoProbe return the checking result
func (s *Server) DoProbe() (bool, string) {

	output, err := s.RunSSHCmd()

	if err != nil {
		log.Errorf("[%s / %s] %v", s.ProbeKind, s.ProbeName, err)
		return false, err.Error() + " - " + output
	}

	log.Debugf("[%s / %s] - %s", s.ProbeKind, s.ProbeName, global.CommandLine(s.Command, s.Args))
	log.Debugf("[%s / %s] - %s", s.ProbeKind, s.ProbeName, probe.CheckEmpty(string(output)))

	info, err := s.ParseHostInfo(string(output))
	if err != nil {
		log.Errorf("[%s / %s] %v", s.ProbeKind, s.ProbeName, err)
		return false, fmt.Sprintf("Prase the output failed: %v", err)
	}
	log.Debugf("[%s / %s] - %+v", s.ProbeKind, s.ProbeName, info)
	s.ExportMetrics(&info)
	return s.CheckThreshold(info)
}

// Usage return all of the resources usage
func (s *Server) Usage(info Info) string {
	usage := fmt.Sprintf(" ( CPU: %.2f%% - ", (100 - info.CPU.Idle))
	usage += fmt.Sprintf("Memory: %.2f%% - ", info.Memory.Usage)
	diskUsage := []string{}
	for _, disk := range info.Disks {
		diskUsage = append(diskUsage, fmt.Sprintf("`%s` %.2f%%", disk.Tag, disk.Usage))
	}
	usage += "Disk: " + strings.Join(diskUsage, ", ")
	loadAvg := []string{}
	for _, load := range info.Load {
		loadAvg = append(loadAvg, fmt.Sprintf("%.2f", load))
	}
	usage += " - "
	usage += "Load: " + strings.Join(loadAvg, "/") + " )"
	return usage
}

func addMessage(msg string, message string) string {
	if msg == "" {
		return message
	}
	return msg + " | " + message
}

// CheckThreshold check the threshold
func (s *Server) CheckThreshold(info Info) (bool, string) {
	status := true
	message := ""

	if s.Threshold.CPU > 0 && s.Threshold.CPU <= (100-info.CPU.Idle)/100 {
		status = false
		message = addMessage(message, "CPU Busy!")
	}
	if s.Threshold.Mem > 0 && s.Threshold.Mem <= info.Memory.Usage/100 {
		status = false
		message = addMessage(message, "Memory Shortage!")
	}
	lowDisks := []string{}
	for _, disk := range info.Disks {
		if s.Threshold.Disk > 0 && s.Threshold.Disk <= disk.Usage/100 {
			lowDisks = append(lowDisks, disk.Tag)
		}
	}
	if len(lowDisks) > 0 {
		status = false
		message = addMessage(message, fmt.Sprintf("Disk Space Low! - [%s]", strings.Join(lowDisks, ", ")))
	}

	for k, v := range info.Load {
		// normalize the load average to 1 cpu core
		if v/float64(info.Core) > s.Threshold.Load[k] {
			status = false
			message = addMessage(message, fmt.Sprintf("Load Average %s High! - %.2f", k, v))
		}
	}

	if message == "" {
		message = "Fine!"
	}

	return status, message + s.Usage(info)
}

// Usage is the resource usage for memory and disk
type Usage struct {
	Used  int     `yaml:"used"`
	Total int     `yaml:"total"`
	Usage float64 `yaml:"usage"`
	Tag   string  `yaml:"tag"`
}

// CPU is the cpu usage
// "1.6 us,  1.6 sy,  3.2 ni, 91.9 id,  1.6 wa,  0.0 hi,  0.0 si,  0.0 st"
type CPU struct {
	User  float64 `yaml:"user"`
	Sys   float64 `yaml:"sys"`
	Nice  float64 `yaml:"nice"`
	Idle  float64 `yaml:"idle"`
	Wait  float64 `yaml:"wait"`
	Hard  float64 `yaml:"hard"`
	Soft  float64 `yaml:"soft"`
	Steal float64 `yaml:"steal"`
}

// Parse a string to a CPU struct
func (c *CPU) Parse(s string) error {
	cpu := strings.Split(s, ",")
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

func first(str string) string {
	return strings.Split(strings.TrimSpace(str), " ")[0]
}

// Info is the host probe information
type Info struct {
	HostName string             `yaml:"hostname"`
	OS       string             `yaml:"os"`
	Core     int64              `yaml:"core"`
	CPU      CPU                `yaml:"cpu"`
	Memory   Usage              `yaml:"memory"`
	Load     map[string]float64 `yaml:"load"`
	Disks    []Usage            `yaml:"disks"`
}

type lineCnt int

const (
	hostName lineCnt = iota
	osName
	memUsage
	cpuCore
	cpuUsage
	loadAvg
	diskUsage
)

// ParseHostInfo parse the host info
func (s *Server) ParseHostInfo(str string) (Info, error) {
	info := Info{
		HostName: "",
		OS:       "",
		Core:     0,
		CPU:      CPU{},
		Memory:   Usage{},
		Load:     map[string]float64{},
		Disks:    []Usage{},
	}
	line := strings.Split(str, "\n")
	if len(line) < int(diskUsage) {
		return info, fmt.Errorf("invalid output")
	}

	info.HostName = line[hostName]
	info.OS = line[osName]

	mem := strings.Split(line[memUsage], " ")
	if len(mem) < 3 {
		return info, fmt.Errorf("invalid memory output")
	}
	info.Memory.Used = int(strInt(mem[0]))
	info.Memory.Total = int(strInt(mem[1]))
	info.Memory.Usage = strFloat(mem[2])

	info.Core = strInt(line[cpuCore])
	if err := info.CPU.Parse(line[cpuUsage]); err != nil {
		return info, err
	}

	load := strings.Split(line[loadAvg], " ")
	if len(load) < 3 {
		return info, fmt.Errorf("invalid load average output")
	}
	info.Load["m1"] = strFloat(load[0])
	info.Load["m5"] = strFloat(load[1])
	info.Load["m15"] = strFloat(load[2])

	for i := int(diskUsage); i < len(line); i++ {
		if strings.TrimSpace(line[i]) == "" {
			break
		}
		disk := strings.Split(line[i], " ")
		if len(disk) < 4 {
			return info, fmt.Errorf("invalid disk output")
		}
		info.Disks = append(info.Disks, Usage{
			Used:  int(strInt(disk[0])),
			Total: int(strInt(disk[1])),
			Usage: strFloat(disk[2][:len(disk[2])-1]),
			Tag:   disk[3],
		})
	}

	return info, nil
}

func strFloat(str string) float64 {
	n, _ := strconv.ParseFloat(strings.TrimSpace(str), 32)
	return n
}

func strInt(str string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(str), 10, 32)
	return n
}

// ExportMetrics export the metrics
func (s *Server) ExportMetrics(info *Info) {
	s.ExportCPUMetrics(info)
	s.ExportMemoryMetrics(info)
	s.ExportDiskMetrics(info)
	s.ExportLoadMetrics(info)
}

// ExportCPUMetrics export the cpu metrics
func (s *Server) ExportCPUMetrics(info *Info) {
	// CPU metrics
	s.metrics.CPU.With(prometheus.Labels{
		"host":  s.Name(),
		"state": "usage",
	}).Set(100 - info.CPU.Idle)

	s.metrics.CPU.With(prometheus.Labels{
		"host":  s.Name(),
		"state": "idle",
	}).Set(info.CPU.Idle)

	s.metrics.CPU.With(prometheus.Labels{
		"host":  s.Name(),
		"state": "user",
	}).Set(info.CPU.User)

	s.metrics.CPU.With(prometheus.Labels{
		"host":  s.Name(),
		"state": "sys",
	}).Set(info.CPU.Sys)

	s.metrics.CPU.With(prometheus.Labels{
		"host":  s.Name(),
		"state": "nice",
	}).Set(info.CPU.Nice)

	s.metrics.CPU.With(prometheus.Labels{
		"host":  s.Name(),
		"state": "wait",
	}).Set(info.CPU.Wait)

	s.metrics.CPU.With(prometheus.Labels{
		"host":  s.Name(),
		"state": "hard",
	}).Set(info.CPU.Hard)

	s.metrics.CPU.With(prometheus.Labels{
		"host":  s.Name(),
		"state": "soft",
	}).Set(info.CPU.Soft)

	s.metrics.CPU.With(prometheus.Labels{
		"host":  s.Name(),
		"state": "steal",
	}).Set(info.CPU.Steal)
}

// ExportMemoryMetrics export the memory metrics
func (s *Server) ExportMemoryMetrics(info *Info) {
	s.metrics.Memory.With(prometheus.Labels{
		"host":  s.Name(),
		"state": "used",
	}).Set(float64(info.Memory.Used))

	s.metrics.Memory.With(prometheus.Labels{
		"host":  s.Name(),
		"state": "available",
	}).Set(float64(info.Memory.Total - info.Memory.Used))

	s.metrics.Memory.With(prometheus.Labels{
		"host":  s.Name(),
		"state": "total",
	}).Set(float64(info.Memory.Total))

	s.metrics.Memory.With(prometheus.Labels{
		"host":  s.Name(),
		"state": "usage",
	}).Set(info.Memory.Usage)
}

// ExportDiskMetrics export the disk metrics
func (s *Server) ExportDiskMetrics(info *Info) {
	for _, disk := range info.Disks {
		s.metrics.Disk.With(prometheus.Labels{
			"host":  s.Name(),
			"disk":  disk.Tag,
			"state": "used",
		}).Set(float64(disk.Used))

		s.metrics.Disk.With(prometheus.Labels{
			"host":  s.Name(),
			"disk":  disk.Tag,
			"state": "available",
		}).Set(float64(disk.Total - disk.Used))

		s.metrics.Disk.With(prometheus.Labels{
			"host":  s.Name(),
			"disk":  disk.Tag,
			"state": "total",
		}).Set(float64(disk.Total))

		s.metrics.Disk.With(prometheus.Labels{
			"host":  s.Name(),
			"disk":  disk.Tag,
			"state": "usage",
		}).Set(disk.Usage)
	}
}

// ExportLoadMetrics export the load metrics
func (s *Server) ExportLoadMetrics(info *Info) {
	s.metrics.Load.With(prometheus.Labels{
		"host":  s.Name(),
		"state": "m1",
	}).Set(info.Load["m1"])

	s.metrics.Load.With(prometheus.Labels{
		"host":  s.Name(),
		"state": "m5",
	}).Set(info.Load["m5"])

	s.metrics.Load.With(prometheus.Labels{
		"host":  s.Name(),
		"state": "m15",
	}).Set(info.Load["m15"])
}
