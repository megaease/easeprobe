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

// ProbeSettings is the global probe setting
type ProbeSettings struct {
	TimeFormat string
	Interval   time.Duration
	Timeout    time.Duration
}

// NormalizeTimeOut return a normalized timeout value
func (p *ProbeSettings) NormalizeTimeOut(t time.Duration) time.Duration {
	return normalize(p.Timeout, t, 0, DefaultTimeOut)
}

// NormalizeInterval return a normalized time interval value
func (p *ProbeSettings) NormalizeInterval(t time.Duration) time.Duration {
	return normalize(p.Interval, t, 0, DefaultProbeInterval)
}
