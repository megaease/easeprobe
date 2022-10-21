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
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

func saveData(doneSave chan bool, saveChannel chan probe.Result) {
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
		case res := <-saveChannel:
			probe.SetResultData(res.Name, &res)
		case <-doneSave:
			save()
			log.Info("Received the exit signal. Saving data, process is exiting...")
			return
		case <-interval.C:
			log.Debugf("SaveData - %s Interval is up, Saving data to file...", c.Settings.Probe.Interval)
			save()
			interval.Reset(c.Settings.Probe.Interval)
		}
	}
}

func scheduleSLA(probers []probe.Prober) {
	cron := gocron.NewScheduler(global.GetEaseProbe().TimeLoc)

	dryNotify := conf.Get().Settings.Notify.Dry

	notifies := channel.GetNotifiers(conf.Get().Settings.SLAReport.Channels)
	if len(notifies) == 0 {
		log.Warnf("No notify settings found for SLA report...")
		return
	}

	log.Debugf("--------- SLA Report Notifies ---------")
	for _, n := range notifies {
		log.Debugf("  - %s : %s", n.Kind(), n.Name())
	}

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
		log.Infof("Next SLA report will be sent at %s", t.Format(conf.Get().Settings.TimeFormat))
	}

	time := conf.Get().Settings.SLAReport.Time
	switch conf.Get().Settings.SLAReport.Schedule {
	case conf.Minutely:
		cron.Cron("* * * * *").Do(SLAFn)
		log.Infoln("Scheduling every minute SLA reports...")
	case conf.Hourly:
		cron.Cron("0 * * * *").Do(SLAFn)
		log.Infof("Scheduling hourly SLA reports...")
	case conf.Daily:
		cron.Every(1).Day().At(time).Do(SLAFn)
		log.Infof("Scheduling daily SLA reports at %s UTC time...", time)
	case conf.Weekly:
		cron.Every(1).Day().Sunday().At(time).Do(SLAFn)
		log.Infof("Scheduling weekly SLA reports on Sunday at %s UTC time...", time)
	case conf.Monthly:
		cron.Every(1).MonthLastDay().At(time).Do(SLAFn)
		log.Infof("Scheduling monthly SLA reports for last day of the month at %s UTC time...", time)
	default:
		cron.Every(1).Day().At("00:00").Do(SLAFn)
		log.Warnf("Bad Scheduling! Setting daily SLA reports to be sent at 00:00 UTC...")
	}

	cron.StartAsync()

	_, t := cron.NextRun()
	log.Infof("The SLA report will be schedule at %s", t.Format(conf.Get().Settings.TimeFormat))
}
