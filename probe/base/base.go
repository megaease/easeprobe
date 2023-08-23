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

// Package base is the base package for all probes
package base

import (
	"fmt"
	"math"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/metric"
	"github.com/megaease/easeprobe/probe"
)

// Probe Simple Status
const (
	ServiceUp   int = 1
	ServiceDown int = 0
)

// ProbeFuncType is the probe function type
type ProbeFuncType func() (bool, string)

// DefaultProbe is the default options for all probe
type DefaultProbe struct {
	ProbeKind                            string            `yaml:"-" json:"-"`
	ProbeTag                             string            `yaml:"-" json:"-"`
	ProbeName                            string            `yaml:"name" json:"name" jsonschema:"required,title=Probe Name,description=the name of probe must be unique"`
	ProbeChannels                        []string          `yaml:"channels" json:"channels,omitempty" jsonschema:"title=Probe Channels,description=the channels of probe message need to send to"`
	ProbeTimeout                         time.Duration     `yaml:"timeout,omitempty" json:"timeout,omitempty" jsonschema:"type=string,format=duration,title=Probe Timeout,description=the timeout of probe"`
	ProbeTimeInterval                    time.Duration     `yaml:"interval,omitempty" json:"interval,omitempty" jsonschema:"type=string,format=duration,title=Probe Interval,description=the interval of probe"`
	Labels                               prometheus.Labels `yaml:"labels,omitempty" json:"labels,omitempty" jsonschema:"title=Probe LabelMap,description=the labels of probe"`
	global.StatusChangeThresholdSettings `yaml:",inline" json:",inline"`
	global.NotificationStrategySettings  `yaml:"alert" json:"alert" jsonschema:"title=Probe Alert,description=the alert strategy of probe"`
	ProbeFunc                            ProbeFuncType `yaml:"-" json:"-"`
	ProbeResult                          *probe.Result `yaml:"-" json:"-"`
	metrics                              *metrics      `yaml:"-" json:"-"`
}

// LabelMap return the const metric labels  for a probe in the configuration.
func (d *DefaultProbe) LabelMap() prometheus.Labels {
	return d.Labels
}

// SetLabelMap set a set of new labels for a probe.
//
//	Note: This method takes effect before Probe.Config() only
func (d *DefaultProbe) SetLabelMap(labels prometheus.Labels) {
	d.Labels = labels
}

// Kind return the probe kind
func (d *DefaultProbe) Kind() string {
	return d.ProbeKind
}

// Name return the probe name
func (d *DefaultProbe) Name() string {
	return d.ProbeName
}

// Channels return the probe channels
func (d *DefaultProbe) Channels() []string {
	return d.ProbeChannels
}

// Timeout get the probe timeout
func (d *DefaultProbe) Timeout() time.Duration {
	return d.ProbeTimeout
}

// Interval get the probe interval
func (d *DefaultProbe) Interval() time.Duration {
	return d.ProbeTimeInterval
}

// Result get the probe result
func (d *DefaultProbe) Result() *probe.Result {
	return d.ProbeResult
}

// LogTitle return the log title
func (d *DefaultProbe) LogTitle() string {
	if len(d.ProbeTag) > 0 {
		return fmt.Sprintf("[%s / %s / %s]", d.ProbeKind, d.ProbeTag, d.ProbeName)
	}
	return fmt.Sprintf("[%s / %s]", d.ProbeKind, d.ProbeName)
}

// CheckStatusThreshold check the status threshold
func (d *DefaultProbe) CheckStatusThreshold() probe.Status {
	s := d.StatusChangeThresholdSettings
	c := d.ProbeResult.Stat.StatusCounter
	title := d.LogTitle()
	log.Debugf("%s - Status Threshold Checking - Current[%v], StatusCnt[%d], FailureThread[%d], SuccessThread[%d]",
		title, c.CurrentStatus, c.StatusCount, s.Failure, s.Success)

	if c.CurrentStatus == true && c.StatusCount >= s.Success {
		if d.ProbeResult.Status != probe.StatusUp {
			cnt := math.Max(float64(c.StatusCount), float64(s.Success))
			log.Infof("%s - Status is UP! Threshold reached for success [%d/%d]", title, int(cnt), s.Success)
		}
		return probe.StatusUp
	}
	if c.CurrentStatus == false && c.StatusCount >= s.Failure {
		if d.ProbeResult.Status != probe.StatusDown {
			cnt := math.Max(float64(c.StatusCount), float64(s.Failure))
			log.Infof("%s - Status is DOWN! Threshold reached for failure [%d/%d]", title, int(cnt), s.Failure)
		}
		return probe.StatusDown
	}
	if c.CurrentStatus == true {
		log.Infof("%s - Status unchanged [%s]! Threshold is not reached for success [%d/%d].",
			title, d.ProbeResult.PreStatus, c.StatusCount, s.Success)
	} else {
		log.Infof("%s - Status unchanged [%s]! Threshold is not reached for failure [%d/%d].",
			title, d.ProbeResult.PreStatus, c.StatusCount, s.Failure)
	}
	return d.ProbeResult.PreStatus
}

// Config default config
func (d *DefaultProbe) Config(gConf global.ProbeSettings,
	kind, tag, name, endpoint string, fn ProbeFuncType) error {

	d.ProbeKind = kind
	d.ProbeName = name
	d.ProbeTag = tag
	d.ProbeFunc = fn

	d.ProbeTimeout = gConf.NormalizeTimeOut(d.ProbeTimeout)
	d.ProbeTimeInterval = gConf.NormalizeInterval(d.ProbeTimeInterval)
	d.StatusChangeThresholdSettings = gConf.NormalizeThreshold(d.StatusChangeThresholdSettings)
	d.NotificationStrategySettings = gConf.NormalizeNotificationStrategy(d.NotificationStrategySettings)

	d.ProbeResult = probe.NewResultWithName(name)
	d.ProbeResult.Name = name
	d.ProbeResult.Endpoint = endpoint

	// update the notification strategy settings
	d.ProbeResult.Stat.NotificationStrategyData.Strategy = d.NotificationStrategySettings.Strategy
	d.ProbeResult.Stat.NotificationStrategyData.Factor = d.NotificationStrategySettings.Factor
	d.ProbeResult.Stat.NotificationStrategyData.MaxTimes = d.NotificationStrategySettings.MaxTimes

	// Set the new length of the status counter
	maxLen := d.StatusChangeThresholdSettings.Failure
	if d.StatusChangeThresholdSettings.Success > maxLen {
		maxLen = d.StatusChangeThresholdSettings.Success
	}
	d.ProbeResult.Stat.StatusCounter.SetMaxLen(maxLen)

	// if there no channels, use the default channel
	if len(d.ProbeChannels) == 0 {
		d.ProbeChannels = append(d.ProbeChannels, global.DefaultChannelName)
	}

	log.Infof("Probe %s base options are configured!", d.LogTitle())

	if d.Failure > 1 || d.Success > 1 {
		log.Infof("Probe %s Status Threshold are configured! failure[%d], success[%d]", d.LogTitle(), d.Failure, d.Success)
	}

	d.metrics = newMetrics(kind, tag, d.Labels)

	return nil
}

// Probe return the checking result
func (d *DefaultProbe) Probe() probe.Result {
	if d.ProbeFunc == nil {
		return *d.ProbeResult
	}

	now := time.Now().UTC()
	d.ProbeResult.StartTime = now
	d.ProbeResult.StartTimestamp = now.UnixMilli()

	stat, msg := d.ProbeFunc()

	d.ProbeResult.RoundTripTime = time.Since(now)

	// check the status threshold
	d.ProbeResult.Stat.StatusCounter.AppendStatus(stat, msg)
	status := d.CheckStatusThreshold()
	title := status.Title()

	// process the notification strategy
	d.ProbeResult.Stat.NotificationStrategyData.ProcessStatus(status == probe.StatusUp)

	if len(d.ProbeTag) > 0 {
		d.ProbeResult.Message = fmt.Sprintf("%s (%s/%s): %s", title, d.ProbeKind, d.ProbeTag, msg)
	} else {
		d.ProbeResult.Message = fmt.Sprintf("%s (%s): %s", title, d.ProbeKind, msg)
	}

	log.Debugf("%s - %s", d.LogTitle(), msg)

	d.ProbeResult.PreStatus = d.ProbeResult.Status
	d.ProbeResult.Status = status

	d.DownTimeCalculation(status)

	d.ProbeResult.DoStat(d.Interval())

	d.ExportMetrics()

	result := d.ProbeResult.Clone()
	return result
}

// ExportMetrics export the metrics
func (d *DefaultProbe) ExportMetrics() {
	cnt := int64(0)
	time := time.Duration(0)

	if d.ProbeResult.Status == probe.StatusUp {
		cnt = d.ProbeResult.Stat.Status[probe.StatusUp]
		time = d.ProbeResult.Stat.UpTime
	} else {
		cnt = d.ProbeResult.Stat.Status[probe.StatusDown]
		time = d.ProbeResult.Stat.DownTime
	}

	// Add endpoint label according to ProbeKind(tcp/http/ping/host/...)
	d.metrics.TotalCnt.With(metric.AddConstLabels(prometheus.Labels{
		"name":     d.ProbeName,
		"status":   d.ProbeResult.Status.String(),
		"endpoint": d.ProbeResult.Endpoint,
	}, d.Labels)).Set(float64(cnt))

	d.metrics.TotalTime.With(metric.AddConstLabels(prometheus.Labels{
		"name":     d.ProbeName,
		"status":   d.ProbeResult.Status.String(),
		"endpoint": d.ProbeResult.Endpoint,
	}, d.Labels)).Set(float64(time.Seconds()))

	d.metrics.Duration.With(metric.AddConstLabels(prometheus.Labels{
		"name":     d.ProbeName,
		"status":   d.ProbeResult.Status.String(),
		"endpoint": d.ProbeResult.Endpoint,
	}, d.Labels)).Set(float64(d.ProbeResult.RoundTripTime.Milliseconds()))

	status := ServiceUp // up
	if d.ProbeResult.Status != probe.StatusUp {
		status = ServiceDown // down
	}
	d.metrics.Status.With(metric.AddConstLabels(prometheus.Labels{
		"name":     d.ProbeName,
		"endpoint": d.ProbeResult.Endpoint,
	}, d.Labels)).Set(float64(status))

	d.metrics.SLA.With(metric.AddConstLabels(prometheus.Labels{
		"name":     d.ProbeName,
		"endpoint": d.ProbeResult.Endpoint,
	}, d.Labels)).Set(float64(d.ProbeResult.SLAPercent()))
}

// DownTimeCalculation calculate the down time
func (d *DefaultProbe) DownTimeCalculation(status probe.Status) {

	// Status from UP to DOWN - Failure
	if d.ProbeResult.PreStatus != probe.StatusDown && status == probe.StatusDown {
		d.ProbeResult.LatestDownTime = time.Now().UTC()
	}

	// Status from DOWN to UP - Recovery
	if d.ProbeResult.PreStatus == probe.StatusDown && status == probe.StatusUp {
		d.ProbeResult.RecoveryDuration = time.Since(d.ProbeResult.LatestDownTime)
	}
}

// GetProxyConnection return the proxy connection
func (d *DefaultProbe) GetProxyConnection(socks5 string, host string) (net.Conn, error) {
	proxyDialer := proxy.FromEnvironment()
	env := true
	if socks5 != "" {
		log.Debugf("[%s / %s] - Proxy Setting found - %s", d.ProbeKind, d.ProbeName, socks5)
		proxyURL, err := url.Parse(socks5)
		if err != nil {
			log.Errorf("[%s / %s] Invalid proxy: %s", d.ProbeKind, d.ProbeName, socks5)
			return nil, fmt.Errorf("Invalid proxy: %s, %v", socks5, err)
		}
		proxyDialer, err = proxy.FromURL(proxyURL, &net.Dialer{Timeout: d.ProbeTimeout})
		if err != nil {
			log.Errorf("[%s / %s] Invalid proxy: %s", d.ProbeKind, d.ProbeName, socks5)
			return nil, fmt.Errorf("Invalid proxy: %s, %v", socks5, err)
		}
		env = false
	}

	if proxyDialer != proxy.Direct {
		if env {
			names := []string{"ALL_PROXY", "all_proxy"}
			for _, n := range names {
				socks5 = os.Getenv(n)
				if socks5 != "" {
					break
				}
			}
		}

		log.Debugf("[%s / %s] - Using the proxy server [%s] for connection", d.ProbeKind, d.ProbeName, socks5)
		return proxyDialer.Dial("tcp", host)
	}
	return net.DialTimeout("tcp", host, d.ProbeTimeout)
}
