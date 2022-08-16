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

package base

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/proxy"

	log "github.com/sirupsen/logrus"
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
	ProbeKind         string        `yaml:"-"`
	ProbeTag          string        `yaml:"-"`
	ProbeName         string        `yaml:"name"`
	ProbeChannels     []string      `yaml:"channels"`
	ProbeTimeout      time.Duration `yaml:"timeout,omitempty"`
	ProbeTimeInterval time.Duration `yaml:"interval,omitempty"`
	ProbeFunc         ProbeFuncType `yaml:"-"`
	ProbeResult       *probe.Result `yaml:"-"`
	metrics           *metrics      `yaml:"-"`
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

// Config default config
func (d *DefaultProbe) Config(gConf global.ProbeSettings,
	kind, tag, name, endpoint string, fn ProbeFuncType) error {

	d.ProbeKind = kind
	d.ProbeName = name
	d.ProbeTag = tag
	d.ProbeFunc = fn

	d.ProbeTimeout = gConf.NormalizeTimeOut(d.ProbeTimeout)
	d.ProbeTimeInterval = gConf.NormalizeInterval(d.ProbeTimeInterval)

	d.ProbeResult = probe.NewResultWithName(name)
	d.ProbeResult.Name = name
	d.ProbeResult.Endpoint = endpoint

	if len(d.ProbeChannels) == 0 {
		d.ProbeChannels = append(d.ProbeChannels, global.DefaultChannelName)
	}

	if len(d.ProbeTag) > 0 {
		log.Infof("Probe [%s / %s] - [%s] base options are configured!", d.ProbeKind, d.ProbeTag, d.ProbeName)
	} else {
		log.Infof("Probe [%s] - [%s] base options are configured!", d.ProbeKind, d.ProbeName)
	}

	d.metrics = newMetrics(kind, tag)

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

	status := probe.StatusUp
	title := "Success"
	if stat != true {
		status = probe.StatusDown
		title = "Error"
	}

	if len(d.ProbeTag) > 0 {
		d.ProbeResult.Message = fmt.Sprintf("%s (%s/%s): %s", title, d.ProbeKind, d.ProbeTag, msg)
		log.Debugf("[%s / %s / %s] - %s", d.ProbeKind, d.ProbeTag, d.ProbeName, msg)
	} else {
		d.ProbeResult.Message = fmt.Sprintf("%s (%s): %s", title, d.ProbeKind, msg)
		log.Debugf("[%s / %s] - %s", d.ProbeKind, d.ProbeName, msg)
	}

	d.ProbeResult.PreStatus = d.ProbeResult.Status
	d.ProbeResult.Status = status

	d.ExportMetrics()

	d.DownTimeCalculation(status)

	d.ProbeResult.DoStat(d.Interval())

	result := d.ProbeResult.Clone()
	return result
}

// ExportMetrics export the metrics
func (d *DefaultProbe) ExportMetrics() {
	d.metrics.Total.With(prometheus.Labels{
		"name":   d.ProbeName,
		"status": d.ProbeResult.Status.String(),
	}).Inc()

	d.metrics.Duration.With(prometheus.Labels{
		"name":   d.ProbeName,
		"status": d.ProbeResult.Status.String(),
	}).Set(float64(d.ProbeResult.RoundTripTime.Milliseconds()))

	status := ServiceUp // up
	if d.ProbeResult.Status != probe.StatusUp {
		status = ServiceDown // down
	}
	d.metrics.Status.With(prometheus.Labels{
		"name": d.ProbeName,
	}).Set(float64(status))

	d.metrics.SLA.With(prometheus.Labels{
		"name": d.ProbeName,
	}).Set(float64(d.ProbeResult.SLAPercent()))
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
