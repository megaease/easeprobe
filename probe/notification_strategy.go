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

	// the current notified times
	Notified int `yaml:"notified" json:"notified"`
	// the current continuous failed rounds
	Failed int `yaml:"failed" json:"failed"`
	// the next round will be notified
	Next int `yaml:"next" json:"next"`
	// the Interval is the interval between two notifications
	Interval int `yaml:"interval" json:"interval"`
	// the flag to indicate whether the notification is sent
	IsSent bool `yaml:"-" json:"-"`
}

// NewNotificationStrategyData returns a new NotificationStrategy
func NewNotificationStrategyData(strategy global.IntervalStrategy, maxTimes int, factor int) *NotificationStrategyData {
	n := &NotificationStrategyData{
		NotificationStrategySettings: global.NotificationStrategySettings{
			Strategy: strategy,
			Factor:   factor,
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
		Notified:                     n.Notified,
		Failed:                       n.Failed,
		Next:                         n.Next,
		Interval:                     n.Interval,
		IsSent:                       n.IsSent,
	}
}

// Reset resets the current times
func (n *NotificationStrategyData) Reset() {
	n.Failed = 0
	n.Notified = 0
	n.Next = 1
	n.Interval = 0
	n.IsSent = false
}

// IsExceedMaxTimes returns true if the current times is equal to the max times
func (n *NotificationStrategyData) IsExceedMaxTimes() bool {
	return n.Notified > n.MaxTimes
}

// NextNotification returns the next notification times
func (n *NotificationStrategyData) NextNotification() {
	switch n.Strategy {

	// the interval is fixed and same as the probe interval，
	//       interval = factor
	// if the factor = 1, then the alert would be sent at 1, 2，3，4，5，6，7...
	//          		  the interval is 1, 1, 1, 1, 1, 1, 1...
	// if the factor = 2, then the alert would be sent at 1, 3, 5, 7, 9, 11, 13...
	// 				  	  the interval is 2, 2, 2, 2, 2, 2, 2...
	// if the factor = 3, then the alert would be sent at 1, 4, 7, 10, 13, 16, 19...
	// 				      the interval is 3, 3, 3, 3, 3, 3, 3...
	case global.RegularStrategy:
		n.Interval = n.Factor

	// the interval is increased linearly.
	//     interval = factor * ( failed times - 1 ) + 1
	// if the factor = 1, then the alert would be sent at 1, 2, 4, 7, 11, 16, 22, 29, 37...
	// 	                  the interval is 1, 2, 3, 4, 5, 6, 7, 8, 9...
	// if the factor = 2, then the alert would be sent at 1, 3, 7, 13, 21, 31, 43, 57, 73...
	// 	                  the interval is 2, 4, 6, 8, 10, 12, 14, 16, 18...
	// if the factor = 3, then the alert would be sent at 1, 4, 10, 19, 31, 46, 64, 85, 109...
	// 	                  the interval is 3, 6, 9, 12, 15, 18, 21, 24, 27...
	case global.IncrementStrategy:
		n.Interval += n.Factor

	// the interval is increased exponentially.
	//    interval =   failed times + factor * ( failed times - 1 )
	// if the factor = 1, then the alert would be sent at 1, 2, 4, 8, 16, 32, 64, 128, 256...
	// 	                  the interval is 1, 2, 4, 8, 16, 32, 64, 128...
	// if the factor = 2, then the alert would be sent at 1, 3, 9, 27, 81, 243, 729, 2187, 6561...
	// 	                  the interval is 2, 6, 18, 54, 162, 486, 1458, 4374, 13122...
	// if the factor = 3, then the alert would be sent at 1, 4, 16, 64, 256, 1024...
	// 	                  the interval is 3, 12, 48, 192, 768...
	case global.ExponentialStrategy:
		n.Interval = n.Failed * n.Factor

	// the regular strategy and the factor is 1 is used by default
	default:
		n.Interval = 1
	}

	// the next alert round will be current round + the interval
	n.Next = n.Failed + n.Interval
}

// ProcessStatus processes the probe status
func (n *NotificationStrategyData) ProcessStatus(status bool) {
	n.IsSent = false
	if status == true {
		n.Reset()
		return
	}
	n.Failed++
	// not meet the next notification round
	if n.Failed < n.Next {
		return
	}
	// meet the next notification round
	n.Notified++

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
