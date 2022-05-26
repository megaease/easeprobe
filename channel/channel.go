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
	log "github.com/sirupsen/logrus"
	"sync/atomic"
)

const kind = "channel"

// Channel implements a config for Channel
type Channel struct {
	Name      string                    `yaml:"name"`      // unique name
	Probers   map[string]*probe.Prober  `yaml:"probers"`   // probers
	Notifiers map[string]*notify.Notify `yaml:"notifiers"` // notifiers
	isWatch   int32                     `yaml:"-"`         // is watch
	done      chan bool                 `yaml:"-"`         // done channel
	channel   chan probe.Result         `yaml:"-"`         // notify channel

}

// NewEmpty creates a new empty Channel object with nil channel
// After setup the probers, You have to call Config() to create the channel
func NewEmpty(name string) *Channel {
	return &Channel{
		Name:      name,
		Probers:   map[string]*probe.Prober{},
		Notifiers: map[string]*notify.Notify{},
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
func (c *Channel) GetProber(name string) *probe.Prober {
	return c.Probers[name]
}

// SetProbers sets the Notify objects
func (c *Channel) SetProbers(probers []*probe.Prober) {
	for _, p := range probers {
		c.SetProber(p)
	}
}

// SetProber sets the Notify object
func (c *Channel) SetProber(p *probe.Prober) {
	if p == nil {
		return
	}
	c.Probers[(*p).Name()] = p
}

// GetNotify returns the Notify object
func (c *Channel) GetNotify(name string) *notify.Notify {
	return c.Notifiers[name]
}

// SetNotifiers sets the Notify objects
func (c *Channel) SetNotifiers(notifiers []*notify.Notify) {
	for _, n := range notifiers {
		c.SetNotify(n)
	}
}

// SetNotify sets the Notify object
func (c *Channel) SetNotify(n *notify.Notify) {
	if n == nil {
		return
	}
	c.Notifiers[(*n).GetName()] = n
}

var dryNotify bool

// SetDryNotify sets the global dry run flag
func SetDryNotify(dry bool) {
	dryNotify = dry
}

// WatchEvent watches the notification event
// Go through all of notification to notify the result.
func (c *Channel) WatchEvent() {
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
			// if the status has no change, no need notify
			if result.PreStatus == result.Status {
				log.Debugf("[%s / %s]: %s (%s) - Status no change [%s] == [%s], no notification.",
					kind, c.Name, result.Name, result.Endpoint, result.PreStatus, result.Status)
				continue
			}
			if result.PreStatus == probe.StatusInit && result.Status == probe.StatusUp {
				log.Debugf("[%s / %s]: %s (%s) - Initial Status [%s] == [%s], no notification.",
					kind, c.Name, result.Name, result.Endpoint, result.PreStatus, result.Status)
				continue
			}
			log.Infof("[%s / %s]: %s (%s) - Status changed [%s] ==> [%s]",
				kind, c.Name, result.Name, result.Endpoint, result.PreStatus, result.Status)
			for _, nRef := range c.Notifiers {
				n := *nRef
				if dryNotify {
					n.DryNotify(result)
				} else {
					go n.Notify(result)
				}
			}
		}
	}
}
