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
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/megaease/easeprobe/conf"
	"github.com/megaease/easeprobe/daemon"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/web"

	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func main() {

	dryNotify := flag.Bool("d", os.Getenv("PROBE_DRY") == "true", "dry notification mode")
	yamlFile := flag.String("f", getEnvOrDefault("PROBE_CONFIG", "config.yaml"), "configuration file")
	flag.Parse()

	c, err := conf.New(yamlFile)
	if err != nil {
		log.Fatalln("Fatal: Cannot read the YAML configuration file!")
		os.Exit(-1)
	}

	// Create the pid file if the file name is not empty
	if len(strings.TrimSpace(c.Settings.PIDFile)) > 0 {
		d, err := daemon.NewPIDFile(c.Settings.PIDFile)
		if err != nil {
			log.Fatalf("Fatal: Cannot create the PID file: %s!", err)
			os.Exit(-1)
		}
		log.Infof("Successfully created the PID file: %s", d.PIDFile)
		defer d.RemovePIDFile()
	} else {
		log.Info("No PID file is created.")
	}

	c.InitAllLogs()

	// if dry notification mode is specificed in command line, overwrite the configuration
	if *dryNotify {
		c.Settings.Notify.Dry = *dryNotify
	}

	if c.Settings.Notify.Dry {
		log.Infoln("Dry Notification Mode...")
	}

	// Probers
	probers := c.AllProbers()

	// Notification
	notifies := c.AllNotifiers()

	// wait group for probers
	var wg sync.WaitGroup
	// the exit channel for all probers
	doneProbe := make(chan bool, len(probers))
	// the exit channel for saving the data
	doneSave := make(chan bool)
	// the exit channel for watching the event
	doneWatch := make(chan bool)
	// Create the Notification Channel
	notifyChan := make(chan probe.Result)

	// Configure the Probes
	configProbers(probers, notifyChan, &wg, doneProbe)
	probe.CleanData(probers) // remove the data not in probers
	go saveData(doneSave)    // save the data to file

	// Configure the Notifiers
	configNotifiers(notifies)
	// Start the Event Watching
	go watchEvent(notifyChan, notifies, doneWatch)

	// Start the HTTP Server
	web.SetProbers(probers)
	web.Server()

	// Set the Cron Job for SLA Report
	if conf.Get().Settings.SLAReport.Schedule != conf.None {
		scheduleSLA(probers, notifies)
	} else {
		log.Info("No SLA Report would be sent!!")
	}

	// Graceful Shutdown
	done := make(chan os.Signal)
	signal.Notify(done, syscall.SIGTERM)
	signal.Notify(done, syscall.SIGINT)

	// Rotate the log file
	rotateLog := make(chan os.Signal, 1)
	doneRotate := make(chan bool, 1)
	signal.Notify(rotateLog, syscall.SIGHUP)
	go func() {
		for {
			c := conf.Get()
			select {
			case <-doneRotate:
				log.Info("Received the exit signal, Rotating log file process exiting...")
				c.Settings.Log.Close()
				c.Settings.HTTPServer.AccessLog.Close()
				return
			case <-rotateLog:
				log.Info("Received SIGHUP, rotating the log file...")
				c.Settings.Log.Rotate()
				c.Settings.HTTPServer.AccessLog.Rotate()
			}
		}
	}()

	select {
	case <-done:
		log.Infof("Received the exit signal, exiting...")
		for i := 0; i < len(probers); i++ {
			if probers[i].Result().Status != probe.StatusBad {
				doneProbe <- true
			}
		}
		wg.Wait()
		doneWatch <- true
		doneSave <- true
		doneRotate <- true
	}

	log.Info("Graceful Exit Successfully!")
}

func saveData(doneSave chan bool) {
	c := conf.Get()
	file := c.Settings.SLAReport.DataFile
	save := func() {
		if err := probe.SaveDataToFile(file); err != nil {
			log.Errorf("Failed to save the SLA data to file: %v", err)
		} else {
			log.Debugf("Successfully save the SLA data to file: %s", file)
		}
	}
	save()
	for {
		select {
		case <-doneSave:
			save()
			log.Info("Received the exit signal, Saving data process is exiting...")
			return
		case <-time.After(c.Settings.Probe.Interval):
			save()
		}
	}

}

// 1) all of probers send the result to notify channel
// 2) go through all of notification to notify the result.
// 3) send the SLA report
func watchEvent(notifyChan chan probe.Result, notifiers []notify.Notify, doneWatch chan bool) {

	dryNotify := conf.Get().Settings.Notify.Dry

	// Watching the Probe Event...
	for {
		select {
		case <-doneWatch:
			log.Infof("Received the done signal, event watching exiting...")
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
			for _, n := range notifiers {
				if dryNotify {
					n.DryNotify(result)
				} else {
					go n.Notify(result)
				}
			}
		}
	}

}

func configProbers(probers []probe.Prober, notifyChan chan probe.Result, wg *sync.WaitGroup, done chan bool) {

	probeFn := func(p probe.Prober) {
		wg.Add(1)
		defer wg.Done()
		for {
			res := p.Probe()
			log.Debugf("%s: %s", p.Kind(), res.DebugJSON())
			notifyChan <- res

			select {
			case <-done:
				log.Infof("%s / %s - Received the done signal, exiting...", p.Kind(), p.Name())
				return
			case <-time.After(p.Interval()):
			}
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
			p.Result().Status = probe.StatusBad
			p.Result().Message = "Bad Configuration: " + err.Error()
			log.Errorf("Bad Probe Configuration: %v", err)
			continue
		}

		if len(p.Result().Message) <= 0 {
			p.Result().Message = "Good Configuration!"
		}
		log.Infof("Ready to monitor(%s): %s - %s", p.Kind(), p.Result().Name, p.Result().Endpoint)
		go probeFn(p)
	}
}

func configNotifiers(notifies []notify.Notify) {
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
