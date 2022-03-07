package log

import (
	"log"
	"os"

	"github.com/megaease/easeprobe/probe"
	"github.com/sirupsen/logrus"
)

// NotifyConfig is the configuration of the Notify
type NotifyConfig struct {
	File string `yaml:"file"`
	Dry  bool   `yaml:"dry"`
}

// Kind return the type of Notify
func (c NotifyConfig) Kind() string {
	return "log"
}

// Config configures the log files
func (c NotifyConfig) Config() error {
	if c.Dry {
		logrus.Infof("Notification %s is running on Dry mode!", c.Kind())
		log.SetOutput(os.Stdout)
		return nil
	}
	file, err := os.OpenFile(c.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		logrus.Errorf("error: %s", err)
		return err
	}
	log.SetOutput(file)
	return nil
}

// Notify write the message into the file
func (c NotifyConfig) Notify(result probe.Result) {
	if c.Dry {
		c.DryNotify(result)
		return
	}
	log.Println(result.JSON())
	logrus.Infof("Logged the notification for %s (%s)!", result.Name, result.Endpoint)
}

// NotifyStat write the stat message into the file
func (c NotifyConfig) NotifyStat(probers []probe.Prober) {
	if c.Dry {
		c.DryNotifyStat(probers)
		return
	}
	logrus.Infoln("LogFile Sending the Statstics...")
	for _, p := range probers {
		log.Println(p.Result())
	}
	logrus.Infoln("Logged the Statstics into %s!", c.File)
}

// DryNotify just log the notification message
func (c NotifyConfig) DryNotify(result probe.Result) {
	logrus.Infoln(result.HTML())
}

// DryNotifyStat just log the notification message
func (c NotifyConfig) DryNotifyStat(probers []probe.Prober) {
	logrus.Infoln(probe.StatHTML(probers))
}
