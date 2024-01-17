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
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/megaease/easeprobe/channel"
	"github.com/megaease/easeprobe/conf"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
)

func configProbers(probers []probe.Prober) []probe.Prober {
	conf.MergeConstLabels(probers)

	gProbeConf := global.ProbeSettings{
		Interval:                      conf.Get().Settings.Probe.Interval,
		Timeout:                       conf.Get().Settings.Probe.Timeout,
		StatusChangeThresholdSettings: conf.Get().Settings.Probe.StatusChangeThresholdSettings,
		NotificationStrategySettings:  conf.Get().Settings.Probe.NotificationStrategySettings,
	}
	log.Debugf("Global Probe Configuration: %+v", gProbeConf)

	validProbers := []probe.Prober{}
	for i := 0; i < len(probers); i++ {
		p := probers[i]
		if err := p.Config(gProbeConf); err != nil {
			p.Result().Status = probe.StatusBad
			p.Result().Message = "Bad Configuration: " + err.Error()
			log.Errorf("Bad Probe Configuration for prober %s %s: %v", p.Kind(), p.Name(), err)
			continue
		}

		if len(p.Result().Message) <= 0 {
			p.Result().Message = "Good Configuration!"
		}
		validProbers = append(validProbers, p)
	}

	return validProbers
}

func runProbers(probers []probe.Prober, wg *sync.WaitGroup, done chan bool, saveChannel chan probe.Result) {
	// we need to run all probers in equally distributed time, not at the same time.
	timeGap := global.DefaultProbeInterval / time.Duration(len(probers))
	// if less than or equal to 60 probers, use 1 second instead
	if time.Duration(len(probers))*time.Second <= time.Minute {
		timeGap = time.Second
	}
	log.Debugf("Start Time Gap: %v = %v / %d", timeGap, global.DefaultProbeInterval, len(probers))

	probeFn := func(p probe.Prober, index int) {
		wg.Add(1)
		defer wg.Done()

		// Sleep a round time to avoid all probers start at the same time.
		t := time.Duration(index) * timeGap
		log.Debugf("[%s / %s] Delay %v = (%d * %v) seconds to start the probe work",
			p.Kind(), p.Name(), t, index, timeGap)
		time.Sleep(t)

		interval := time.NewTimer(p.Interval())
		defer interval.Stop()
		for {
			res := p.Probe()
			log.Debugf("%s: %s", p.Kind(), res.DebugJSON())
			// send the result to the persistent channel
			saveChannel <- res
			// send the result to all channels
			for _, cName := range p.Channels() {
				if ch := channel.GetChannel(cName); ch != nil {
					ch.Send(res)
				}
			}

			select {
			case <-done:
				log.Infof("%s / %s - Received the done signal, exiting...", p.Kind(), p.Name())
				return
			case <-interval.C:
				interval.Reset(p.Interval())
				log.Debugf("%s / %s - %s Interval is up, continue...", p.Kind(), p.Name(), p.Interval())
			}
		}
	}

	for i := 0; i < len(probers); i++ {
		p := probers[i]
		if p.Result().Status == probe.StatusBad {
			continue
		}
		log.Infof("Ready to monitor(%s): %s - %s", p.Kind(), p.Result().Name, p.Result().Endpoint)
		go probeFn(p, i)
	}
}
