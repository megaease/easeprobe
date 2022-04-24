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
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"

	log "github.com/sirupsen/logrus"
)

// ProbeFuncType is the probe function type
type ProbeFuncType func() (bool, string)

// DefaultOptions is the default options for all probe
type DefaultOptions struct {
	ProbeKind         string        `yaml:"-"`
	ProbeTag          string        `yaml:"-"`
	ProbeName         string        `yaml:"name"`
	ProbeTimeout      time.Duration `yaml:"timeout,omitempty"`
	ProbeTimeInterval time.Duration `yaml:"interval,omitempty"`
	ProbeFunc         ProbeFuncType `yaml:"-"`
	ProbeResult       *probe.Result `yaml:"-"`
}

// Kind return the probe kind
func (d *DefaultOptions) Kind() string {
	return d.ProbeKind
}

// Name return the probe name
func (d *DefaultOptions) Name() string {
	return d.ProbeName
}

// Timeout get the probe timeout
func (d *DefaultOptions) Timeout() time.Duration {
	return d.ProbeTimeout
}

// Interval get the probe interval
func (d *DefaultOptions) Interval() time.Duration {
	return d.ProbeTimeInterval
}

// Result get the probe result
func (d *DefaultOptions) Result() *probe.Result {
	return d.ProbeResult
}

// Config default config
func (d *DefaultOptions) Config(gConf global.ProbeSettings,
	kind, tag, name, endpoint string, fn ProbeFuncType) error {

	d.ProbeKind = kind
	d.ProbeName = name
	d.ProbeTag = tag
	d.ProbeFunc = fn

	d.ProbeTimeout = gConf.NormalizeTimeOut(d.ProbeTimeout)
	d.ProbeTimeInterval = gConf.NormalizeInterval(d.ProbeTimeInterval)

	d.ProbeResult = probe.NewResult()
	d.ProbeResult.Name = name
	d.ProbeResult.Endpoint = endpoint
	d.ProbeResult.PreStatus = probe.StatusInit
	d.ProbeResult.TimeFormat = gConf.TimeFormat

	if len(d.ProbeTag) > 0 {
		log.Infof("Probe [%s / %s] - [%s] base options are configured!", d.ProbeKind, d.ProbeTag, d.ProbeName)
	} else {
		log.Infof("Probe [%s] - [%s] base options are configured!", d.ProbeKind, d.ProbeName)
	}
	return nil
}

// Probe return the checking result
func (d *DefaultOptions) Probe() probe.Result {
	if d.ProbeFunc == nil {
		return *d.ProbeResult
	}

	now := time.Now()
	d.ProbeResult.StartTime = now
	d.ProbeResult.StartTimestamp = now.UnixMilli()

	stat, msg := d.ProbeFunc()

	d.ProbeResult.RoundTripTime.Duration = time.Since(now)

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

	d.DownTimeCalculation(status)

	d.ProbeResult.DoStat(d.Interval())
	return *d.ProbeResult
}

// DownTimeCalculation calculate the down time
func (d *DefaultOptions) DownTimeCalculation(status probe.Status) {

	// Status from UP to DOWN - Failure
	if d.ProbeResult.PreStatus != probe.StatusDown && status == probe.StatusDown {
		d.ProbeResult.LatestDownTime = time.Now()
	}

	// Status from DOWN to UP - Recovery
	if d.ProbeResult.PreStatus == probe.StatusDown && status == probe.StatusUp {
		d.ProbeResult.RecoveryDuration = time.Since(d.ProbeResult.LatestDownTime)
	}
}
