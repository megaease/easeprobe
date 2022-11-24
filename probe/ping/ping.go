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

// Package ping is the ping probe package
package ping

import (
	"fmt"

	"github.com/go-ping/ping"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Ping implements a config for ping
type Ping struct {
	base.DefaultProbe `yaml:",inline"`
	Host              string  `yaml:"host" json:"host" jsonschema:"required,title=Host,description=The host to ping"`
	Count             int     `yaml:"count" json:"count" jsonschema:"title=Count,description=The number of ping packets to send,minimum=1,default=3"`
	LostThreshold     float64 `yaml:"lost" json:"lost" jsonschema:"title=Lost Threshold,description=The threshold of packet loss,minimum=0,maximum=1,default=0"`

	metrics *metrics `yaml:"-" json:"-"`
}

// DefaultPingCount is the default ping count
const DefaultPingCount = 3

// DefaultLostThreshold is the default lost threshold - 0% lost
const DefaultLostThreshold = 0.0

// Config Ping Config Object
func (p *Ping) Config(gConf global.ProbeSettings) error {
	kind := "ping"
	tag := ""
	name := p.ProbeName
	p.DefaultProbe.Config(gConf, kind, tag, name, p.Host, p.DoProbe)

	if p.Count <= 0 {
		log.Debugf("[%s / %s] ping count is not set, use default value: %d", p.ProbeKind, p.ProbeName, DefaultPingCount)
		p.Count = DefaultPingCount
	}

	if p.LostThreshold < 0 || p.LostThreshold > 1 {
		log.Debugf("[%s / %s] lost threshold is not set, use default value: %f", p.ProbeKind, p.ProbeName, DefaultLostThreshold)
		p.LostThreshold = DefaultLostThreshold
	}

	p.metrics = newMetrics(kind, tag)

	log.Debugf("[%s / %s] configuration: %+v", p.ProbeKind, p.ProbeName, p)
	return nil
}

// DoProbe return the checking result
func (p *Ping) DoProbe() (bool, string) {
	pinger, err := ping.NewPinger(p.Host)
	if err != nil {
		return false, err.Error()
	}
	pinger.Timeout = p.ProbeTimeout
	pinger.Count = p.Count
	err = pinger.Run() // Blocks until finished.
	if err != nil {
		return false, err.Error()
	}
	stats := pinger.Statistics() // get send/receive/rtt stats
	p.ExportMetrics(stats)

	stat := ""
	stat += fmt.Sprintf("%d sent, %d received, %v%% loss, RTT min/avg/max/stddev = %v/%v/%v/%v",
		stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss,
		stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)

	log.Debugf("[%s / %s] --- %s ping statistics ---", p.ProbeKind, p.ProbeName, stats.Addr)
	log.Debugf("[%s / %s] %d packets transmitted, %d packets received, %v%% packet loss\n",
		p.ProbeKind, p.ProbeName, stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
	log.Debugf("[%s / %s] round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
		p.ProbeKind, p.ProbeName, stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)

	result := true
	message := "Ping Succeeded!"
	// if half of the packets are lost, return false
	if stats.PacketLoss > p.LostThreshold*100 {
		result = false
		message = "Ping Failed!"
	}
	return result, fmt.Sprintf("%s: %d/%d ( %s )", message, stats.PacketsRecv, stats.PacketsSent, stat)
}

// ExportMetrics export Ping metrics
func (p *Ping) ExportMetrics(stats *ping.Statistics) {
	p.metrics.PacketsSent.With(prometheus.Labels{
		"name": p.ProbeName,
	}).Add(float64(stats.PacketsSent))

	p.metrics.PacketsRecv.With(prometheus.Labels{
		"name": p.ProbeName,
	}).Add(float64(stats.PacketsRecv))

	p.metrics.PacketLoss.With(prometheus.Labels{
		"name": p.ProbeName,
	}).Set(stats.PacketLoss)

	p.metrics.MaxRtt.With(prometheus.Labels{
		"name": p.ProbeName,
	}).Set(float64(stats.MaxRtt.Milliseconds()))

	p.metrics.MinRtt.With(prometheus.Labels{
		"name": p.ProbeName,
	}).Set(float64(stats.MinRtt.Milliseconds()))

	p.metrics.AvgRtt.With(prometheus.Labels{
		"name": p.ProbeName,
	}).Set(float64(stats.AvgRtt.Milliseconds()))

	p.metrics.StdDevRtt.With(prometheus.Labels{
		"name": p.ProbeName,
	}).Set(float64(stats.StdDevRtt.Milliseconds()))
}
