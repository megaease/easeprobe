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
	"strconv"
	"strings"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/ssh"
	log "github.com/sirupsen/logrus"
)

// Default threshold
const (
	DefaultCPUThreshold  = 0.8
	DefaultMemThreshold  = 0.8
	DefaultDiskThreshold = 0.95
)

// Threshold is the threshold of a probe
type Threshold struct {
	CPU  float64 `yaml:"cpu"`
	Mem  float64 `yaml:"mem"`
	Disk float64 `yaml:"disk"`
}

func (t *Threshold) String() string {
	return fmt.Sprintf("CPU: %.2f, Mem: %.2f, Disk: %.2f", t.CPU, t.Mem, t.Disk)
}

// Server is the server of a host probe
type Server struct {
	ssh.Server `yaml:",inline"`
	Threshold  Threshold `yaml:"threshold"`
}

// Host is the host probe configuration
type Host struct {
	Bastion *ssh.BastionMapType `yaml:"bastion"`
	Servers []Server            `yaml:"servers"`
}

// BastionMap is a map of bastion
var BastionMap ssh.BastionMapType

// Config is the host probe configuration
func (s *Server) Config(gConf global.ProbeSettings) error {
	kind := "host"
	tag := "server"
	name := s.ProbeName

	// The following commands are:
	// 1. retrive the hostname: 	`hostname``
	// 2. retrive the os:  			`awk -F= '/^NAME/{print $2}' /etc/os-release | tr -d '\"'`
	// 3. retrive the memory usage:	`free -m | awk 'NR==2{printf "%s %s %.2f\n", $3,$2,$3*100/$2 }'`
	//    output: used(MB) total(MB) usage(%), example: 19379 31654 61.22
	// 4. retrive the cpu core:		`grep -c ^processor /proc/cpuinfo;`
	// 5. retrive the cpu usage:	`top -b -n 1 | grep Cpu | awk -F ":" '{print $2}'`
	//    output example: 1.6 us,  0.0 sy,  0.0 ni, 98.4 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st
	// 6. retrive the disk usage	`df -h / 2>/dev/null | awk '$NF=="/"{printf "%d %d %s\n", $3,$2,$5}'`
	//    output: used(GB) total(GB) usage(%), example: 40 970 5%

	s.Command = `hostname;
	awk -F= '/^NAME/{print $2}' /etc/os-release | tr -d '\"';
	free -m | awk 'NR==2{printf "%s %s %.2f\n", $3,$2,$3*100/$2 }';
	grep -c ^processor /proc/cpuinfo;
	top -b -n 1 | grep Cpu | awk -F ":" '{print $2}';
	df -h / 2>/dev/null | awk '$NF=="/"{printf "%d %d %s\n", $3,$2,$5}'`

	if s.Threshold.CPU == 0 {
		s.Threshold.CPU = DefaultCPUThreshold
	}
	if s.Threshold.Mem == 0 {
		s.Threshold.Mem = DefaultMemThreshold
	}
	if s.Threshold.Disk == 0 {
		s.Threshold.Disk = DefaultDiskThreshold
	}

	endpoint := s.Threshold.String()
	return s.Configure(gConf, kind, tag, name, endpoint, &BastionMap, s.DoProbe)
}

// DoProbe return the checking result
func (s *Server) DoProbe() (bool, string) {

	output, err := s.RunSSHCmd()

	if err != nil {
		log.Errorf("[%s / %s] %v", s.ProbeKind, s.ProbeName, err)
		return false, err.Error() + " - " + output
	}

	log.Debugf("[%s / %s] - %s", s.ProbeKind, s.ProbeName, probe.CommandLine(s.Command, s.Args))
	log.Debugf("[%s / %s] - %s", s.ProbeKind, s.ProbeName, probe.CheckEmpty(string(output)))

	info, err := s.ParseHostInfo(string(output))
	if err != nil {
		log.Errorf("[%s / %s] %v", s.ProbeKind, s.ProbeName, err)
		return false, fmt.Sprintf("Prase the output failed: %v", err)
	}
	log.Debugf("[%s / %s] - %+v", s.ProbeKind, s.ProbeName, info)
	return s.CheckThreshold(info)
}

// CheckThreshold check the threshold
func (s *Server) CheckThreshold(info Info) (bool, string) {
	status := true
	message := ""
	usage := fmt.Sprintf(" ( CPU: %.2f%% - ", (100 - info.CPU.Idle))
	usage += fmt.Sprintf("Memory: %.2f%% - ", info.Memory.Usage)
	usage += fmt.Sprintf("Disk: %.2f%% )", info.Disk.Usage)

	if s.Threshold.CPU > 0 && s.Threshold.CPU <= (100-info.CPU.Idle)/100 {
		status = false
		message += "CPU Busy!"
	}
	if s.Threshold.Mem > 0 && s.Threshold.Mem <= info.Memory.Usage/100 {
		status = false
		if message != "" {
			message += " | "
		}
		message += "Memory Shortage!"
	}
	if s.Threshold.Disk > 0 && s.Threshold.Disk <= info.Disk.Usage/100 {
		status = false
		if message != "" {
			message += " | "
		}
		message += "Disk Full!"
	}

	if message == "" {
		message = "Fine!"
	}

	return status, message + usage
}

// Usage is the resource usage for memory and disk
type Usage struct {
	Used  int     `yaml:"used"`
	Total int     `yaml:"total"`
	Usage float64 `yaml:"usage"`
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
	HostName string `yaml:"hostname"`
	OS       string `yaml:"os"`
	Core     int64  `yaml:"core"`
	CPU      CPU    `yaml:"cpu"`
	Memory   Usage  `yaml:"memory"`
	Disk     Usage  `yaml:"disk"`
}

// ParseHostInfo parse the host info
func (s *Server) ParseHostInfo(str string) (Info, error) {
	info := Info{}
	line := strings.Split(str, "\n")
	if len(line) < 5 {
		return info, fmt.Errorf("invalid output")
	}

	info.HostName = line[0]
	info.OS = line[1]

	mem := strings.Split(line[2], " ")
	if len(mem) < 3 {
		return info, fmt.Errorf("invalid memory output")
	}
	info.Memory.Used = int(strInt(mem[0]))
	info.Memory.Total = int(strInt(mem[1]))
	info.Memory.Usage = strFloat(mem[2])

	info.Core = strInt(line[3])
	if err := info.CPU.Parse(line[4]); err != nil {
		return info, err
	}

	disk := strings.Split(line[5], " ")
	if len(disk) < 3 {
		return info, fmt.Errorf("invalid disk output")
	}
	info.Disk.Used = int(strInt(disk[0]))
	info.Disk.Total = int(strInt(disk[1]))
	info.Disk.Usage = strFloat(disk[2][:len(disk[2])-1])

	return info, nil
}

func strFloat(str string) float64 {
	n, _ := strconv.ParseFloat(str, 32)
	return n
}

func strInt(str string) int64 {
	n, _ := strconv.ParseInt(str, 10, 32)
	return n
}
