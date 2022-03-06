package main

import (
	"flag"
	"os"
	"time"

	"github.com/megaease/easeprobe/conf"
	"github.com/megaease/easeprobe/notify"
	"github.com/megaease/easeprobe/probe"

	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
)

// 1) all of probers send the result to notify channel
// 2) go through all of notification to notify the result.
// 3) send the SLA report
func run(probers []probe.Prober, notifies []notify.Notify, done chan bool) {

	dryNotify := conf.Get().Settings.DryNotify

	notifyChan := make(chan probe.Result)

	probeFn := func(p probe.Prober) {
		for {
			res := p.Probe()
			log.Debugf("%s: %s\n", p.Kind(), res.JSON())
			notifyChan <- res
			time.Sleep(p.Interval())
		}
	}

	for _, p := range probers {
		err := p.Config()
		if err != nil {
			log.Errorf("error: %v\n", err)
			continue
		}
		p.Result().TimeFormat = conf.Get().Settings.TimeFormat
		go probeFn(p)
	}

	for _, n := range notifies {
		err := n.Config()
		if err != nil {
			log.Errorf("error: %v\n", err)
			continue
		}
	}

	cron := gocron.NewScheduler(time.UTC)

	statFn := func() {
		for _, n := range notifies {
			if dryNotify {
				n.DryNotifyStat(probers)
			} else {
				go n.NotifyStat(probers)
			}
		}
		_, t := cron.NextRun()
		log.Infof("Next Time to send the SLA Report - %s\n", t.Format("2006-01-02 15:04:05 UTC"))
	}

	if dryNotify {
		cron.Every(1).Minute().Do(statFn)
	} else {
		cron.Every(1).Day().At("00:00").Do(statFn)
	}

	cron.StartAsync()

	for {
		select {
		case <-done:
			return
		case result := <-notifyChan:
			// if the status has no change, no need notify
			if result.PreStatus == result.Status {
				log.Debugf("%s (%s) - Status no change [%s] == [%s]\n, no notification.\n",
					result.Name, result.Endpoint, result.PreStatus, result.Status)
				continue
			}
			if result.PreStatus == probe.StatusInit && result.Status == probe.StatusUp {
				log.Debugf("%s (%s) - Initial Status [%s] == [%s], no notification.\n",
					result.Name, result.Endpoint, result.PreStatus, result.Status)
				continue
			}
			log.Infof("%s (%s) - Status changed [%s] ==> [%s]\n",
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
		conf.Settings.DryNotify = *dryNotify
	}

	if conf.Settings.DryNotify {
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
		notifies = append(notifies, conf.Notify.Log[i])
	}

	for i := 0; i < len(conf.Notify.Email); i++ {
		notifies = append(notifies, conf.Notify.Email[i])
	}

	for i := 0; i < len(conf.Notify.Slack); i++ {
		notifies = append(notifies, conf.Notify.Slack[i])
	}

	done := make(chan bool)
	run(probers, notifies, done)

}
