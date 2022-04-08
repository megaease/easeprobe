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
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
)

// DefaultOptions is the default options for all probe
type DefaultOptions struct {
	ProbeKind         string        `yaml:"kind"`
	ProbeTimeout      time.Duration `yaml:"timeout,omitempty"`
	ProbeTimeInterval time.Duration `yaml:"interval,omitempty"`
	ProbeResult       *probe.Result `yaml:"-"`
}

// Config default config
func (d *DefaultOptions) Config(gConf global.ProbeSettings, name, endpoint string) error {
	d.ProbeTimeout = gConf.NormalizeTimeOut(d.ProbeTimeout)
	d.ProbeTimeInterval = gConf.NormalizeInterval(d.ProbeTimeInterval)

	d.ProbeResult = probe.NewResult()
	d.ProbeResult.Name = name
	d.ProbeResult.Endpoint = endpoint
	d.ProbeResult.PreStatus = probe.StatusInit
	d.ProbeResult.TimeFormat = gConf.TimeFormat
	return nil
}

// Kind return the probe kind
func (d *DefaultOptions) Kind() string {
	return d.ProbeKind
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
