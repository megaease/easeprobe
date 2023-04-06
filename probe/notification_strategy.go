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

package probe

import "github.com/megaease/easeprobe/global"

// NotificationStrategyData is the notification strategy
type NotificationStrategyData struct {
	global.NotificationStrategySettings `yaml:",inline" json:",inline"`

	// the current notification times
	NotifyTimes int `yaml:"times" json:"times"`
	// the current round and the next notification round
	FailedTimes int `yaml:"failed" json:"failed"`
	Next        int `yaml:"next" json:"next"`
	// the Step to the next notification round
	Step int `yaml:"step" json:"step"`
	// the flag to indicate whether the notification is sent
	IsSent bool `yaml:"-" json:"-"`
}

// NewNotificationStrategyData returns a new NotificationStrategy
func NewNotificationStrategyData(strategy global.IntervalStrategy, maxTimes int) *NotificationStrategyData {
	n := &NotificationStrategyData{
		NotificationStrategySettings: global.NotificationStrategySettings{
			Strategy: strategy,
			MaxTimes: maxTimes,
		},
	}
	n.Reset()
	return n
}

// Clone returns a new NotificationStrategyData
func (n *NotificationStrategyData) Clone() NotificationStrategyData {
	return NotificationStrategyData{
		NotificationStrategySettings: n.NotificationStrategySettings,
		NotifyTimes:                  n.NotifyTimes,
		FailedTimes:                  n.FailedTimes,
		Next:                         n.Next,
		Step:                         n.Step,
		IsSent:                       n.IsSent,
	}
}

// Reset resets the current times
func (n *NotificationStrategyData) Reset() {
	n.FailedTimes = 0
	n.NotifyTimes = 0
	n.Next = 1
	n.Step = 0
	n.IsSent = false
}

// IsExceedMaxTimes returns true if the current times is equal to the max times
func (n *NotificationStrategyData) IsExceedMaxTimes() bool {
	return n.NotifyTimes > n.MaxTimes
}

// NextNotification returns the next notification times
func (n *NotificationStrategyData) NextNotification() {
	switch n.Strategy {
	case global.RegularStrategy:
		// Next time is the same as the probe interval， 1, 2，3，4，5，6，7...
		n.Step = 1
	case global.IncrementStrategy:
		// Next time is increased linearly.  1, 2, 4, 7, 11, 16, 22, 29, 37...
		n.Step++
	case global.ExponentialStrategy:
		// Next time is increased exponentially, 1, 2, 4, 8, 16, 32, 64...
		n.Step = n.FailedTimes
	default:
		n.Step = 1
	}
	n.Next = n.FailedTimes + n.Step
}

// ProcessStatus processes the probe status
func (n *NotificationStrategyData) ProcessStatus(status bool) {
	n.IsSent = false
	if status == true {
		n.Reset()
		return
	}
	n.FailedTimes++
	// not meet the next notification round
	if n.FailedTimes != n.Next {
		return
	}
	// meet the next notification round
	n.NotifyTimes++

	// check if exceed the max times
	if n.IsExceedMaxTimes() {
		return
	}

	// update the next notification round
	n.NextNotification()

	// set the flag to indicate that the notification is sent
	n.IsSent = true
}

// NeedToSendNotification returns true if the notification should be sent
func (n *NotificationStrategyData) NeedToSendNotification() bool {
	return n.IsSent
}
