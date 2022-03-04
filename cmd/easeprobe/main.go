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
func run(probers []probe.Prober, notifies []notify.Notify, done chan bool) {

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
		go probeFn(p)
	}

	for _, n := range notifies {
		err := n.Config()
		if err != nil {
			log.Errorf("error: %v\n", err)
			continue
		}
	}

	statFn := func() {
		for _, n := range notifies {
			go n.NotifyStat(probers)
		}
	}
	cron := gocron.NewScheduler(time.UTC)
	cron.Every(1).Day().At("00:00").Do(statFn)
	//cron.Every(5).Minute().Do(statFn)
	cron.StartAsync()

	for {
		select {
		case <-done:
			return
		case result := <-notifyChan:
			// if the status has no change, no need notify
			if result.PreStatus == result.Status {
				log.Debugf("Status no change [%s] == [%s]\n", result.PreStatus, result.Status)
				continue
			}
			log.Infof("Status changed [%s] ==> [%s]\n", result.PreStatus, result.Status)
			for _, n := range notifies {
				go n.Notify(result)
				//log.Println(n)
			}
		}
	}

}

func main() {

	dryrun := flag.Bool("d", false, "dry run mode")
	yamlFile := flag.String("f", "config.yaml", "configuration file")
	flag.Parse()

	if _, err := os.Stat(*yamlFile); err != nil {
		log.Fatalf("Configuration file is not found! - %s", *yamlFile)
		os.Exit(-1)
	}

	conf, err := conf.NewConf(yamlFile)
	if err != nil {
		log.Fatal("Fatal: Cannot read the YAML configuration file!")
		os.Exit(-1)
	}
	defer conf.CloseLogFile()

	// if dry run mode is specificed in command line, overwrite the configuration
	if *dryrun {
		conf.Settings.Dryrun = *dryrun
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
