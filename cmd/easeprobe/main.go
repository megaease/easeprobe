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

// 1) all of probers send the result to notify channel
// 2) go through all of notification to notify the result.
// 3) send the SLA report
func run(probers []probe.Prober, notifies []notify.Notify, done chan bool) {

	dryNotify := conf.Get().Settings.Notify.Dry

	notifyChan := make(chan probe.Result)

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

	cron := gocron.NewScheduler(time.UTC)

	statFn := func() {
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

	if conf.Get().Settings.Debug {
		cron.Every(1).Minute().Do(statFn)
		log.Infoln("Preparing to send the  SLA report in every minute...")
	} else {
		cron.Every(1).Day().At("00:00").Do(statFn)
		log.Infoln("Preparing to send the daily SLA report at 00:00 UTC time...")
	}

	cron.StartAsync()

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
	var probers []probe.Prober

	for i := 0; i < len(conf.HTTP); i++ {
		probers = append(probers, &conf.HTTP[i])
	}

	for i := 0; i < len(conf.TCP); i++ {
		probers = append(probers, &conf.TCP[i])
	}

	// Notification
	var notifies []notify.Notify

	for i := 0; i < len(conf.Notify.Log); i++ {
		notifies = append(notifies, &conf.Notify.Log[i])
	}

	for i := 0; i < len(conf.Notify.Email); i++ {
		notifies = append(notifies, &conf.Notify.Email[i])
	}

	for i := 0; i < len(conf.Notify.Slack); i++ {
		notifies = append(notifies, &conf.Notify.Slack[i])
	}

	for i := 0; i < len(conf.Notify.Discord); i++ {
		notifies = append(notifies, &conf.Notify.Discord[i])
	}

	done := make(chan bool)
	run(probers, notifies, done)

}
