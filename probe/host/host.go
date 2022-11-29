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
	"strings"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/ssh"
	log "github.com/sirupsen/logrus"
)

// Server is the server of a host probe
type Server struct {
	ssh.Server `yaml:",inline"`
	Threshold  Threshold `yaml:"threshold,omitempty" json:"threshold,omitempty" jsonschema:"title=Threshold,description=the threshold of the probe for cpu/memory/disk"`
	Disks      []string  `yaml:"disks,omitempty" json:"disks,omitempty" jsonschema:"title=Disks,description=the disks to be monitored,example=[\"/\", \"/data\"]"`

	outputLines int        `yaml:"-" json:"-"`
	hostMetrics []IMetrics `yaml:"-" json:"-"`
	info        Info       `yaml:"-" json:"-"`
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

	s.ProbeKind = kind
	s.ProbeTag = tag
	s.ProbeName = name

	// put all of the metrics into the IMetrics slice
	s.hostMetrics = s.info.IMetrics()

	// Combine the commands and Config the metrics
	s.outputLines = 0
	s.Command = ""
	for _, m := range s.hostMetrics {
		m.Config(s)
		s.Command += m.Command() + "\n"
		s.outputLines += m.OutputLines()
		log.Debugf("[%s / %s] - metric [%s] configured!", s.ProbeKind, s.ProbeName, m.Name())
	}
	log.Debugf("[%s / %s]\n%s", s.ProbeKind, s.ProbeName, s.Command)

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
	s.ExportMetrics()
	return s.CheckThreshold(info)
}

// Usage return all of the resources usage
func (s *Server) Usage(info Info) string {
	usage := " ( "
	for i := 0; i < len(s.hostMetrics)-1; i++ {
		u := s.hostMetrics[i].UsageInfo()
		if u != "" {
			usage += u + " - "
		}
	}
	usage += s.hostMetrics[len(s.hostMetrics)-1].UsageInfo() + " )"
	return usage
}

// CheckThreshold check the threshold
func (s *Server) CheckThreshold(info Info) (bool, string) {
	status := true
	message := ""

	for _, metric := range s.hostMetrics {
		s, m := metric.CheckThreshold()
		if s == false {
			status = false
			message = addMessage(message, m)
		}
	}

	if message == "" {
		message = "Fine!"
	}

	return status, message + s.Usage(info)
}

// ParseHostInfo parse the host info
func (s *Server) ParseHostInfo(str string) (Info, error) {
	line := strings.Split(str, "\n")
	if len(line) < s.outputLines {
		return s.info, fmt.Errorf("invalid output lines")
	}

	idx := 0
	for _, m := range s.hostMetrics {
		strs := []string{}
		for i := 0; i < m.OutputLines(); i++ {
			strs = append(strs, line[idx])
			idx++
		}
		err := m.Parse(strs)
		if err != nil {
			return s.info, err
		}
	}

	return s.info, nil
}

// ExportMetrics export the metrics
func (s *Server) ExportMetrics() {
	for _, m := range s.hostMetrics {
		m.ExportMetrics(s.Name())
	}
}
