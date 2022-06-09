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
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/megaease/easeprobe/channel"
	"github.com/megaease/easeprobe/conf"
	"github.com/megaease/easeprobe/daemon"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/web"

	log "github.com/sirupsen/logrus"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func main() {

	////////////////////////////////////////////////////////////////////////////
	//          Parse command line arguments and config file settings         //
	////////////////////////////////////////////////////////////////////////////

	dryNotify := flag.Bool("d", os.Getenv("PROBE_DRY") == "true", "dry notification mode")
	yamlFile := flag.String("f", getEnvOrDefault("PROBE_CONFIG", "config.yaml"), "configuration file")
	version := flag.Bool("v", false, "prints version")
	flag.Parse()

	if *version {
		fmt.Println(global.DefaultProg, global.Ver)
		os.Exit(0)
	}

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
		log.Info("Skipping PID file creation (pidfile empty).")
	}

	c.InitAllLogs()

	// if dry notification mode is specificed in command line, overwrite the configuration
	if *dryNotify {
		c.Settings.Notify.Dry = *dryNotify
	}

	if c.Settings.Notify.Dry {
		log.Infoln("Dry Notification Mode...")
	}

	////////////////////////////////////////////////////////////////////////////
	//                          Start the HTTP Server                         //
	////////////////////////////////////////////////////////////////////////////
	// if error happens, the EaseProbe will exit
	web.Server()

	////////////////////////////////////////////////////////////////////////////
	//                  Configure all of Probers and Notifiers                //
	////////////////////////////////////////////////////////////////////////////

	// Probers
	probers := c.AllProbers()
	// Notification
	notifies := c.AllNotifiers()
	// Configure the Probes
	probers = configProbers(probers)
	if len(probers) == 0 {
		log.Fatal("No probes configured, exiting...")
	}
	// Configure the Notifiers
	configNotifiers(notifies)
	// configure channels
	configChannels(probers, notifies)

	////////////////////////////////////////////////////////////////////////////
	//                          Start the EaseProbe                           //
	////////////////////////////////////////////////////////////////////////////

	// wait group for probers
	var wg sync.WaitGroup
	// the exit channel for all probers
	doneProbe := make(chan bool, len(probers))
	// the exit channel for saving the data
	doneSave := make(chan bool)

	// 1) SLA Data Save process
	probe.CleanData(probers) // remove the data not in probers
	go saveData(doneSave)    // save the data to file

	// 2) Start the Probers
	runProbers(probers, &wg, doneProbe)
	// 3) Start the Event Watching
	channel.WatchForAllEvents()

	// 4) Set probers into web server
	web.SetProbers(probers)

	// 5) Set the Cron Job for SLA Report
	if conf.Get().Settings.SLAReport.Schedule != conf.None {
		scheduleSLA(probers)
	} else {
		log.Info("No SLA Report would be sent!!")
	}

	////////////////////////////////////////////////////////////////////////////
	//                          Rotate the log file                           //
	////////////////////////////////////////////////////////////////////////////
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

	////////////////////////////////////////////////////////////////////////////
	//                              Graceful Shutdown                         //
	////////////////////////////////////////////////////////////////////////////
	done := make(chan os.Signal)
	signal.Notify(done, syscall.SIGTERM)
	select {
	case <-done:
		log.Infof("Received the exit signal, exiting...")
		for i := 0; i < len(probers); i++ {
			if probers[i].Result().Status != probe.StatusBad {
				doneProbe <- true
			}
		}
		wg.Wait()
		channel.AllDone()
		doneSave <- true
		doneRotate <- true
	}

	log.Info("Graceful Exit Successfully!")
}
