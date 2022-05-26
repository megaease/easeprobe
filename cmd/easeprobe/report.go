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
	"time"

	"github.com/go-co-op/gocron"
	"github.com/megaease/easeprobe/channel"
	"github.com/megaease/easeprobe/conf"
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

func saveData(doneSave chan bool) {
	c := conf.Get()
	file := c.Settings.SLAReport.DataFile
	save := func() {
		if err := probe.SaveDataToFile(file); err != nil {
			log.Errorf("Failed to save the SLA data to file(%s): %v", file, err)
		} else {
			log.Debugf("Successfully save the SLA data to file: %s", file)
		}
	}

	// if DataFile is explicitly disabled redefine as empty
	if c.Settings.SLAReport.DataFile == "-" {
		save = func() {}
	}

	// save data to file when start the EaseProbe
	save()

	interval := time.NewTimer(c.Settings.Probe.Interval)
	defer interval.Stop()
	for {
		select {
		case <-doneSave:
			save()
			log.Info("Received the exit signal, Saving data process is exiting...")
			return
		case <-interval.C:
			log.Debugf("SaveData - %s Interval is up, Saving data to file...", c.Settings.Probe.Interval)
			save()
			interval.Reset(c.Settings.Probe.Interval)
		}
	}
}

func scheduleSLA(probers []probe.Prober) {
	cron := gocron.NewScheduler(time.UTC)

	dryNotify := conf.Get().Settings.Notify.Dry

	notifies := channel.GetNotifiers(conf.Get().Settings.SLAReport.Channels)
	if len(notifies) == 0 {
		log.Warnf("No notify found for SLA report...")
		return
	}

	log.Debugf("--------- SLA Report Notifies ---------")
	for _, n := range notifies {
		log.Debugf("  - %s : %s", (*n).Kind(), (*n).GetName())
	}

	SLAFn := func() {
		for _, nRef := range notifies {
			n := *nRef
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
