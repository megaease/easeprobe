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

	"github.com/megaease/easeprobe/channel"
	"github.com/megaease/easeprobe/conf"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

func configProbers(probers []probe.Prober) {
	gProbeConf := global.ProbeSettings{
		TimeFormat: conf.Get().Settings.TimeFormat,
		Interval:   conf.Get().Settings.Probe.Interval,
		Timeout:    conf.Get().Settings.Probe.Timeout,
	}
	log.Debugf("Global Probe Configuration: %+v", gProbeConf)
	for i := 0; i < len(probers); i++ {
		p := probers[i]
		if err := p.Config(gProbeConf); err != nil {
			p.Result().Status = probe.StatusBad
			p.Result().Message = "Bad Configuration: " + err.Error()
			log.Errorf("Bad Probe Configuration: %v", err)
			continue
		}

		if len(p.Result().Message) <= 0 {
			p.Result().Message = "Good Configuration!"
		}
	}
}

func runProbers(probers []probe.Prober, wg *sync.WaitGroup, done chan bool) {
	probeFn := func(p probe.Prober) {
		wg.Add(1)
		defer wg.Done()

		interval := time.NewTimer(p.Interval())
		defer interval.Stop()
		for {
			res := p.Probe()
			log.Debugf("%s: %s", p.Kind(), res.DebugJSON())
			// send the notification to all channels
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
		go probeFn(p)
	}
}
