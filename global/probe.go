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

import (
	"strings"
	"time"
)

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
	NotificationStrategySettings
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

// NormalizeNotificationStrategy return a normalized notification strategy value
func (p *ProbeSettings) NormalizeNotificationStrategy(t NotificationStrategySettings) NotificationStrategySettings {
	return NotificationStrategySettings{
		Strategy: normalize(p.Strategy, t.Strategy, UnknownStrategy, RegularStrategy),
		Factor:   normalize(p.Factor, t.Factor, 0, DefaultNotificationFactor),
		MaxTimes: normalize(p.MaxTimes, t.MaxTimes, 0, DefaultMaxNotificationTimes),
	}
}

// ----------------------------------------------------------------------------------------
//                          Notification Interval Strategy
// ----------------------------------------------------------------------------------------

// IntervalStrategy is the notification strategy
type IntervalStrategy int

// The notification strategy enum
const (
	UnknownStrategy     IntervalStrategy = iota
	RegularStrategy                      // the same period of time between each notification
	IncrementStrategy                    // the period of time between each notification is increased by a fixed value
	ExponentialStrategy                  // the period of time between each notification is increased exponentially
)

var (
	toString = map[IntervalStrategy]string{
		UnknownStrategy:     "unknown",
		RegularStrategy:     "regular",
		IncrementStrategy:   "increment",
		ExponentialStrategy: "exponent",
	}
	toIntervalStrategy = ReverseMap(toString)
)

// String returns the string value of the NotificationIntervalStrategy
func (n IntervalStrategy) String() string {
	return toString[n]
}

// IntervalStrategy returns the IntervalStrategy value of the string
func (n *IntervalStrategy) IntervalStrategy(s string) {
	if val, ok := toIntervalStrategy[strings.ToLower(s)]; ok {
		*n = val
	} else {
		*n = UnknownStrategy
	}
}

// UnmarshalYAML is unmarshal the IntervalStrategy.
func (n *IntervalStrategy) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return EnumUnmarshalYaml(unmarshal, toIntervalStrategy, n, RegularStrategy, "IntervalStrategy")
}

// MarshalYAML is marshal the IntervalStrategy.
func (n IntervalStrategy) MarshalYAML() (interface{}, error) {
	return EnumMarshalYaml(toString, n, "IntervalStrategy")
}

// UnmarshalJSON is unmarshal the NotificationIntervalStrategy.
func (n *IntervalStrategy) UnmarshalJSON(data []byte) error {
	return EnumUnmarshalJSON(data, toIntervalStrategy, n, RegularStrategy, "IntervalStrategy")
}

// MarshalJSON is marshal the NotificationIntervalStrategy.
func (n IntervalStrategy) MarshalJSON() ([]byte, error) {
	return EnumMarshalJSON(toString, n, "IntervalStrategy")
}

// NotificationStrategySettings is the notification strategy settings
type NotificationStrategySettings struct {
	Strategy IntervalStrategy `yaml:"strategy" json:"strategy" jsonschema:"title=Alert Interval Strategy,description=the notification interval strategy such as regular,increment,exponent,default=regular"`
	Factor   int              `yaml:"factor" json:"factor" jsonschema:"title=Factor,description=the factor to increase the interval, it must be greater than 0,default=1"`
	MaxTimes int              `yaml:"max" json:"max" jsonschema:"title=Max Times,description=the max times to send notification,default=1"`
}
