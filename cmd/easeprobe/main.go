package main

import (
	"flag"
	"io/ioutil"
	"os"
	"time"

	"github.com/megaease/easeprobe/notify"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/http"
	"github.com/megaease/easeprobe/probe/tcp"
	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

// Conf is Probe configuration
type Conf struct {
	HTTP   []http.HTTP   `yaml:"http"`
	TCP    []tcp.TCP     `yaml:"tcp"`
	Notify notify.Config `yaml:"notify"`
}

func initLog() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func readConf(conf *string) (Conf, error) {

	c := Conf{}

	y, err := ioutil.ReadFile(*conf)
	if err != nil {
		log.Errorf("error: %v ", err)
		return c, err
	}

	err = yaml.Unmarshal(y, &c)
	if err != nil {
		log.Errorf("error: %v\n", err)
		return c, err
	}
	log.Infoln("Load the configuration file successfully!")
	log.Debugf("%v\n", c)

	return c, err
}

// 1) all of probers send the result to notify channel
// 2)
func run(probers []probe.Prober, notifies []notify.Notify, done chan bool) {

	notifyChan := make(chan probe.Result)

	probeFn := func(p probe.Prober) {
		for {
			res := p.Probe()
			log.Infof("%s: %s\n", p.Kind(), res.JSON())
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
			}
		}
	}

}

func main() {

	initLog()

	yamlFile := flag.String("f", "config.yaml", "configuration file")
	flag.Parse()

	conf, err := readConf(yamlFile)
	if err != nil {
		log.Fatal("Fatal: Cannot read the YAML configuration file!")
		os.Exit(-1)
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
