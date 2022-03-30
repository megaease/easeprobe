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
	"flag"
	"os"
	"time"

	"github.com/megaease/easeprobe/conf"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify"
	"github.com/megaease/easeprobe/probe"

	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
)

func main() {

	dryNotify := flag.Bool("d", false, "dry notification mode")
	yamlFile := flag.String("f", "config.yaml", "configuration file")
	flag.Parse()

	if _, err := os.Stat(*yamlFile); err != nil {
		log.Fatalf("Configuration file is not found! - %s", *yamlFile)
		os.Exit(-1)
	}

	conf, err := conf.New(yamlFile)
	if err != nil {
		log.Fatalln("Fatal: Cannot read the YAML configuration file!")
		os.Exit(-1)
	}
	defer conf.CloseLogFile()

	// if dry notification mode is specificed in command line, overwrite the configuration
	if *dryNotify {
		conf.Settings.Notify.Dry = *dryNotify
	}

	if conf.Settings.Notify.Dry {
		log.Infoln("Dry Notification Mode...")
	}

	// Probers
	probers := conf.AllProbers()

	// Notification
	notifies := conf.AllNotifiers()

	done := make(chan bool)
	run(probers, notifies, done)

}

// 1) all of probers send the result to notify channel
// 2) go through all of notification to notify the result.
// 3) send the SLA report
func run(probers []probe.Prober, notifies []notify.Notify, done chan bool) {

	dryNotify := conf.Get().Settings.Notify.Dry

	// Create the Notification Channel
	notifyChan := make(chan probe.Result)

	// Configure the Probes
	configProbers(probers, notifyChan)

	// Configure the Notifiers
	configNotifiers(notifies, notifyChan)

	// Set the Cron Job for SLA Report
	if conf.Get().Settings.SLAReport.Schedule != conf.None {
		scheduleSLA(probers, notifies)
	} else {
		log.Info("No SLA Report would be sent!!")
	}

	// Watching the Probe Event...
	for {
		select {
		case <-done:
			return
		case result := <-notifyChan:
			// if the status has no change, no need notify
			if result.PreStatus == result.Status {
				log.Debugf("%s (%s) - Status no change [%s] == [%s], no notification.",
					result.Name, result.Endpoint, result.PreStatus, result.Status)
				continue
			}
			if result.PreStatus == probe.StatusInit && result.Status == probe.StatusUp {
				log.Debugf("%s (%s) - Initial Status [%s] == [%s], no notification.",
					result.Name, result.Endpoint, result.PreStatus, result.Status)
				continue
			}
			log.Infof("%s (%s) - Status changed [%s] ==> [%s]",
				result.Name, result.Endpoint, result.PreStatus, result.Status)
			for _, n := range notifies {
				if dryNotify {
					n.DryNotify(result)
				} else {
					go n.Notify(result)
				}
			}
		}
	}

}

func configProbers(probers []probe.Prober, notifyChan chan probe.Result) {
	probeFn := func(p probe.Prober) {
		for {
			res := p.Probe()
			log.Debugf("%s: %s", p.Kind(), res.JSON())
			notifyChan <- res
			time.Sleep(p.Interval())
		}
	}
	gProbeConf := global.ProbeSettings{
		TimeFormat: conf.Get().Settings.TimeFormat,
		Interval:   conf.Get().Settings.Probe.Interval,
		Timeout:    conf.Get().Settings.Probe.Timeout,
	}
	log.Debugf("Global Probe Configuration: %+v", gProbeConf)
	for _, p := range probers {
		err := p.Config(gProbeConf)
		if err != nil {
			log.Errorf("error: %v", err)
			continue
		}
		log.Infof("Ready to monitor(%s): %s - %s", p.Kind(), p.Result().Name, p.Result().Endpoint)
		go probeFn(p)
	}
}

func configNotifiers(notifies []notify.Notify, notifyChan chan probe.Result) {
	gNotifyConf := global.NotifySettings{
		TimeFormat: conf.Get().Settings.TimeFormat,
		Retry:      conf.Get().Settings.Notify.Retry,
	}
	for _, n := range notifies {
		err := n.Config(gNotifyConf)
		if err != nil {
			log.Errorf("error: %v", err)
			continue
		}
		log.Infof("Successfully setup the notify channel: %s", n.Kind())
	}
}

func scheduleSLA(probers []probe.Prober, notifies []notify.Notify) {
	cron := gocron.NewScheduler(time.UTC)

	dryNotify := conf.Get().Settings.Notify.Dry

	SLAFn := func() {
		for _, n := range notifies {
			if dryNotify {
				n.DryNotifyStat(probers)
			} else {
				log.Debugf("[%s] notifying the SLA...", n.Kind())
				go n.NotifyStat(probers)
			}
		}
		_, t := cron.NextRun()
		log.Infof("Next Time to send the SLA Report - %s", t.Format(conf.Get().Settings.TimeFormat))
	}

	if conf.Get().Settings.SLAReport.Debug {
		cron.Every(1).Minute().Do(SLAFn)
		log.Infoln("Preparing to send the  SLA report in every minute...")
	} else {
		time := conf.Get().Settings.SLAReport.Time
		switch conf.Get().Settings.SLAReport.Schedule {
		case conf.Daily:
			cron.Every(1).Day().At(time).Do(SLAFn)
			log.Infof("Preparing to send the daily SLA report at %s UTC time...", time)
		case conf.Weekly:
			cron.Every(1).Day().Sunday().At(time).Do(SLAFn)
			log.Infof("Preparing to send the weekly SLA report on Sunday at %s UTC time...", time)
		case conf.Monthly:
			cron.Every(1).MonthLastDay().At(time).Do(SLAFn)
			log.Infof("Preparing to send the monthly SLA report in last day at %s UTC time...", time)
		default:
			cron.Every(1).Day().At("00:00").Do(SLAFn)
			log.Warnf("Bad Scheduling! Preparing to send the daily SLA report at 00:00 UTC time...")
		}
	}

	cron.StartAsync()
	_, t := cron.NextRun()
	log.Infof("Next Time to send the SLA Report - %s", t.Format(conf.Get().Settings.TimeFormat))
}
