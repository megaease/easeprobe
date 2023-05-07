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

// Package main is the entry point for the easeprobe command.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

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

func showVersion() {

	var v = global.Ver

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		v = fmt.Sprintf("%v %v", global.DefaultProg, v)
		fmt.Println(v)
		return
	}

	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			v = fmt.Sprintf("%v %v", v, s.Value[:9])
		case "vcs.time":
			v = fmt.Sprintf("%v %v", v, s.Value)
		}
	}

	v = fmt.Sprintf("%v %v %v", global.DefaultProg, v, bi.GoVersion)
	fmt.Println(v)
}

func main() {
	////////////////////////////////////////////////////////////////////////////
	//          Parse command line arguments and config file settings         //
	////////////////////////////////////////////////////////////////////////////

	dryNotify := flag.Bool("d", os.Getenv("PROBE_DRY") == "true", "dry notification mode")
	yamlFile := flag.String("f", getEnvOrDefault("PROBE_CONFIG", "config.yaml"), "configuration file")
	jsonSchema := flag.Bool("j", false, "show JSON schema")
	version := flag.Bool("v", false, "prints version")
	flag.Parse()

	if *version {
		showVersion()
		os.Exit(0)
	}

	if *jsonSchema {
		schema, err := conf.JSONSchema()
		if err != nil {
			log.Fatalf("failed to show JSON schema: %v", err)
		}
		fmt.Println(schema)
		os.Exit(0)
	}

	c, err := conf.New(yamlFile)
	if err != nil {
		log.Errorln("Fatal: Cannot read the YAML configuration file!")
		os.Exit(-1)
	}

	// Create the pid file if the file name is not empty
	c.Settings.PIDFile = strings.TrimSpace(c.Settings.PIDFile)
	if len(c.Settings.PIDFile) > 0 && c.Settings.PIDFile != "-" {
		d, err := daemon.NewPIDFile(c.Settings.PIDFile)
		if err != nil {
			log.Errorf("Fatal: Cannot create the PID file: %s!", err)
			os.Exit(-1)
		}
		log.Infof("Successfully created the PID file: %s", d.PIDFile)
		defer d.RemovePIDFile()
	} else {
		if len(c.Settings.PIDFile) == 0 {
			log.Info("Skipping PID file creation (pid file is empty).")
		} else {
			log.Info("Skipping PID file creation (pid file is set to '-').")
		}
	}

	c.InitAllLogs()

	// if dry notification mode is specified in command line, overwrite the configuration
	if *dryNotify {
		c.Settings.Notify.Dry = *dryNotify
		log.Infoln("Dry Notification Mode...")
	}
	// set the dry notify flag to channel
	channel.SetDryNotify(c.Settings.Notify.Dry)

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
	notifies = configNotifiers(notifies)
	if len(notifies) == 0 {
		log.Fatal("No notifies configured, exiting...")
	}

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
	// the channel for saving the probe result data
	saveChannel := make(chan probe.Result, len(probers))

	// 1) SLA Data Save process
	probe.CleanData(probers)           // remove the data not in probers
	go saveData(doneSave, saveChannel) // save the data to file

	// 2) Start the Probers
	runProbers(probers, &wg, doneProbe, saveChannel)
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
	//                         Graceful Shutdown / Re-Run                     //
	////////////////////////////////////////////////////////////////////////////
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM)

	// the graceful shutdown process
	exit := func() {
		web.Shutdown()
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

	// the graceful restart process
	reRun := func() {
		exit()
		p, e := os.StartProcess(os.Args[0], os.Args, &os.ProcAttr{
			Env:   os.Environ(),
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		})
		if e != nil {
			log.Errorf("!!! FAILED TO RESTART THE EASEPROBE: %v !!!", e)
			return
		}
		log.Infof("!!! RESTART THE EASEPROBE SUCCESSFULLY - PID=[%d] !!!", p.Pid)
	}

	// Monitor the configuration file
	monConf := make(chan bool, 1)
	go monitorYAMLFile(*yamlFile, monConf)

	// wait for the exit and restart signal
	select {
	case <-done:
		log.Info("!!! RECEIVED THE SIGTERM EXIT SIGNAL, EXITING... !!!")
		exit()
	case <-monConf:
		log.Info("!!! RECEIVED THE RESTART EVENT, RESTARTING... !!!")
		reRun()
	}

	log.Info("Graceful Exit Successfully!")
}

func monitorYAMLFile(path string, monConf chan bool) {
	for {
		if conf.IsConfigModified(path) {
			log.Infof("The configuration file [%s] has been modified, restarting...", path)
			monConf <- true
			break
		}
		log.Debugf("The configuration file [%s] has not been modified", path)
		time.Sleep(global.DefaultConfigFileCheckInterval)
	}
}
