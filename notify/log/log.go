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
}

// Kind return the type of Notify
func (conf NotifyConfig) Kind() string {
	return "log"
}

// Config configures the log files
func (conf NotifyConfig) Config() error {
	file, err := os.OpenFile(conf.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		logrus.Errorf("error: %s\n", err)
		return err
	}
	log.SetOutput(file)
	return nil
}

// Notify write the message into the file
func (conf NotifyConfig) Notify(result probe.Result) {
	logrus.Infoln("LogFile got the notification...")
	log.Println(result.String())
}
