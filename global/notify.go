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

// NotifySettings is the global notification setting
type NotifySettings struct {
	TimeFormat string
	Timeout    time.Duration
	Retry      Retry
}

// NormalizeTimeOut return a normalized timeout value
func (n *NotifySettings) NormalizeTimeOut(t time.Duration) time.Duration {
	return normalize(n.Timeout, t, 0, DefaultTimeOut)
}

// NormalizeRetry return a normalized retry value
func (n *NotifySettings) NormalizeRetry(retry Retry) Retry {
	retry.Interval = normalize(n.Retry.Interval, retry.Interval, 0, DefaultRetryInterval)
	retry.Times = normalize(n.Retry.Times, retry.Times, 0, DefaultRetryTimes)
	return retry
}
