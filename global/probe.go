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

package global

import "time"

// StatusChangeThresholdSettings is the settings for probe threshold
type StatusChangeThresholdSettings struct {
	// the failures threshold such as 2, 5
	Failure int `yaml:"failure,omitempty" json:"failure,omitempty" jsonschema:"title=Failure Threshold,description=the failures threshold to change the status such as 3,default=1"`
	// the success threshold such as 2, 5
	Success int `yaml:"success,omitempty" json:"success,omitempty" jsonschema:"title=Success Threshold,description=the success threshold to change the status such as 2,default=1"`
}

// ProbeSettings is the global probe setting
type ProbeSettings struct {
	Interval time.Duration
	Timeout  time.Duration
	StatusChangeThresholdSettings
}

// NormalizeTimeOut return a normalized timeout value
func (p *ProbeSettings) NormalizeTimeOut(t time.Duration) time.Duration {
	return normalize(p.Timeout, t, 0, DefaultTimeOut)
}

// NormalizeInterval return a normalized time interval value
func (p *ProbeSettings) NormalizeInterval(t time.Duration) time.Duration {
	return normalize(p.Interval, t, 0, DefaultProbeInterval)
}

// NormalizeThreshold return a normalized threshold value
func (p *ProbeSettings) NormalizeThreshold(t StatusChangeThresholdSettings) StatusChangeThresholdSettings {
	return StatusChangeThresholdSettings{
		Failure: normalize(p.Failure, t.Failure, 0, DefaultStatusChangeThresholdSetting),
		Success: normalize(p.Success, t.Success, 0, DefaultStatusChangeThresholdSetting),
	}
}
