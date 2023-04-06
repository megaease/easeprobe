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

// Package channel implements the channel interface
package channel

import (
	"sync"
	"sync/atomic"

	"github.com/megaease/easeprobe/notify"
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

const kind = "channel"

// Channel implements a config for Channel
type Channel struct {
	Name      string                   `yaml:"name"`      // unique name
	Probers   map[string]probe.Prober  `yaml:"probers"`   // probers
	Notifiers map[string]notify.Notify `yaml:"notifiers"` // notifiers
	isWatch   int32                    `yaml:"-"`         // is watch
	done      chan bool                `yaml:"-"`         // done channel
	channel   chan probe.Result        `yaml:"-"`         // notify channel

}

// NewEmpty creates a new empty Channel object with nil channel
// After setup the probers, You have to call Config() to create the channel
func NewEmpty(name string) *Channel {
	return &Channel{
		Name:      name,
		Probers:   map[string]probe.Prober{},
		Notifiers: map[string]notify.Notify{},
		isWatch:   0,
		done:      nil,
		channel:   nil,
	}
}

// Config configures the channel
func (c *Channel) Config() {
	c.done = make(chan bool)
	c.channel = make(chan probe.Result, len(c.Probers))
}

// Done returns the done channel
func (c *Channel) Done() chan bool {
	return c.done
}

// Channel returns the notification channel
func (c *Channel) Channel() chan probe.Result {
	return c.channel
}

// Send sends the result to the channel
func (c *Channel) Send(result probe.Result) {
	c.channel <- result
}

// GetProber returns the Notify object
func (c *Channel) GetProber(name string) probe.Prober {
	return c.Probers[name]
}

// SetProbers sets the Notify objects
func (c *Channel) SetProbers(probers []probe.Prober) {
	for _, p := range probers {
		c.SetProber(p)
	}
}

// SetProber sets the Notify object
func (c *Channel) SetProber(p probe.Prober) {
	if p == nil {
		return
	}
	if _, ok := c.Probers[p.Name()]; ok {
		log.Errorf("Prober [%s - %s] name is duplicated, ignored!", p.Kind(), p.Name())
		return
	}
	c.Probers[p.Name()] = p
}

// GetNotify returns the Notify object
func (c *Channel) GetNotify(name string) notify.Notify {
	return c.Notifiers[name]
}

// SetNotifiers sets the Notify objects
func (c *Channel) SetNotifiers(notifiers []notify.Notify) {
	for _, n := range notifiers {
		c.SetNotify(n)
	}
}

// SetNotify sets the Notify object
func (c *Channel) SetNotify(n notify.Notify) {
	if n == nil {
		return
	}
	if _, ok := c.Notifiers[n.Name()]; ok {
		log.Errorf("Notifier [%s - %s] name is duplicated, ignored!", n.Kind(), n.Name())
		return
	}
	c.Notifiers[n.Name()] = n
}

// WatchEvent watches the notification event
// Go through all of notification to notify the result.
func (c *Channel) WatchEvent(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	// check if the channel is watching
	if atomic.CompareAndSwapInt32(&(c.isWatch), 0, 1) == false {
		log.Warnf("[%s/ %s]: Channel is already watching!", kind, c.Name)
		return
	}

	// set the channel is not watching
	defer func() {
		atomic.StoreInt32(&(c.isWatch), 0)
	}()

	// Watching the Probe Event...
	for {
		select {
		case <-c.done:
			log.Infof("[%s / %s]: Received the done signal, channel exiting...", kind, c.Name)
			return
		case result := <-c.channel:
			// if it is the first time, and the status is UP, no need notify
			if result.PreStatus == probe.StatusInit && result.Status == probe.StatusUp {
				log.Debugf("[%s / %s]: %s (%s) - Initial Status [%s] == [%s], no notification.",
					kind, c.Name, result.Name, result.Endpoint, result.PreStatus, result.Status)
				continue
			}

			// if the status has no change for UP or Init, no need notify
			if result.PreStatus == result.Status && (result.Status == probe.StatusUp || result.Status == probe.StatusInit) {
				log.Debugf("[%s / %s]: %s (%s) - Status no change [%s] == [%s], no notification.",
					kind, c.Name, result.Name, result.Endpoint, result.PreStatus, result.Status)
				continue
			}

			nsd := &result.Stat.NotificationStrategyData
			// if the status changed to UP, reset the notification strategy
			if result.Status == probe.StatusUp {
				nsd.Reset()
			}

			// if the status is DOWN, check the notification strategy
			if result.Status == probe.StatusDown {
				if result.Stat.NotificationStrategyData.NeedToSendNotification() == false {
					log.Debugf("[%s / %s]: %s (%s) - Don't meet the notification condition [max=%d, notified=%d, failed=%d, next=%d], no notification.",
						kind, c.Name, result.Name, result.Endpoint, nsd.MaxTimes, nsd.Notified, nsd.Failed, nsd.Next)
					continue
				}
			}

			if result.PreStatus != result.Status {
				log.Infof("[%s / %s]: %s (%s) - Status changed [%s] ==> [%s], sending notification...",
					kind, c.Name, result.Name, result.Endpoint, result.PreStatus, result.Status)
			} else {
				log.Debugf("[%s / %s]: %s (%s) - Meet the notification condition [max=%d, notified=%d, failed=%d, next=%d], sending notification...",
					kind, c.Name, result.Name, result.Endpoint, nsd.MaxTimes, nsd.Notified, nsd.Failed, nsd.Next)
			}

			for _, n := range c.Notifiers {
				if IsDryNotify() == true {
					n.DryNotify(result)
				} else {
					go n.Notify(result)
				}
			}
		}
	}
}
