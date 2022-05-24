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

package channel

import (
	"github.com/megaease/easeprobe/notify"
	"github.com/megaease/easeprobe/probe"
)

var channel = make(map[string]*Channel)

// GetAllChannels returns all channels
func GetAllChannels() map[string]*Channel {
	return channel
}

// GetChannel returns the channel
func GetChannel(name string) *Channel {
	return channel[name]
}

// SetChannel sets the channel
func SetChannel(name string, buffer int) {
	ch := GetChannel(name)
	if ch == nil {
		channel[name] = NewChannel(name, buffer)
	}
}

// SetProbers sets the probers
func SetProbers(channel string, probers []*probe.Prober) {
	for _, p := range probers {
		SetProber(channel, p)
	}
}

// SetProber sets the prober
func SetProber(channel string, p *probe.Prober) {
	ch := GetChannel(channel)
	if ch == nil {
		SetChannel(channel, 0)
		ch = GetChannel(channel)
	}
	ch.SetProber(p)
}

// SetNotifiers set a notify to the channel
func SetNotifiers(channel string, notifiers []*notify.Notify) {
	for _, n := range notifiers {
		SetNotify(channel, n)
	}
}

// SetNotify set a notify to the channel
func SetNotify(channel string, n *notify.Notify) {
	ch := GetChannel(channel)
	if ch == nil {
		SetChannel(channel, 0)
		ch = GetChannel(channel)
	}
	ch.SetNotify(n)
}

// GetNotifiers returns all of the notifiers by a list of channel names
func GetNotifiers(channel []string) map[string]*notify.Notify {
	notifiers := make(map[string]*notify.Notify)

	for _, c := range channel {
		ch := GetChannel(c)
		if ch == nil {
			continue
		}
		for _, n := range ch.Notifiers {
			notifiers[(*n).GetName()] = n
		}
	}
	return notifiers
}

// WatchForAllEvents watch the event for all channels
func WatchForAllEvents() {
	for _, c := range channel {
		go c.WatchEvent()
	}
}

// AllDone send the done signal to all channels
func AllDone() {
	for _, c := range channel {
		c.Done() <- true
	}
}
