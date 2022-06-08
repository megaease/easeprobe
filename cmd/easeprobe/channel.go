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

package main

import (
	"github.com/megaease/easeprobe/channel"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify"
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

func configChannels(probers []probe.Prober, notifiers []notify.Notify) {
	// set the probe channels
	for i := 0; i < len(probers); i++ {
		p := probers[i]
		if len(p.Channels()) <= 0 {
			channel.SetProber(global.DefaultChannelName, p)
			continue
		}
		for _, cName := range p.Channels() {
			channel.SetProber(cName, p)
		}
	}

	// set the notify to channels
	for i := 0; i < len(notifiers); i++ {
		n := notifiers[i]
		if len(n.Channels()) <= 0 {
			channel.SetNotify(global.DefaultChannelName, n)
			continue
		}
		for _, cName := range n.Channels() {
			channel.SetNotify(cName, n)
		}
	}

	// configure all of the channels
	channel.ConfigAllChannels()

	// check the channel configuration
	checkChannels()
}

func checkChannels() {
	log.Info("--------- Channel Configuration --------- ")
	// check the channel configuration
	for cName, ch := range channel.GetAllChannels() {

		log.Infof("Channel: %s\n", cName)
		log.Infof("   Probers:\n")
		for _, p := range ch.Probers {
			log.Infof("    - %s: %s\n", p.Kind(), p.Name())
		}
		if len(ch.Probers) <= 0 {
			log.Warnf("Channel(%s) has no probers!", cName)
		}

		log.Infof("   Notifiers:\n")
		for _, n := range ch.Notifiers {
			log.Infof("     - %s: %s\n", n.Kind(), n.Name())
		}
		if len(ch.Notifiers) <= 0 {
			log.Warnf("Channel(%s) has no notifiers!", cName)
		}
	}
}
