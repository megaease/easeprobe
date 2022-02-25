package email

import (
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the email notification configuration
type NotifyConfig struct {
	Host string `yaml:"host"`
	User string `yaml:"username"`
	Pass string `yaml:"password"`
}

// Kind return the type of Notify
func (conf NotifyConfig) Kind() string {
	return "email"
}

// Config configures the log files
func (conf NotifyConfig) Config() error {
	return nil
}

// Notify write the message into the file
func (conf NotifyConfig) Notify(result probe.Result) {
	log.Infoln("Email got the notifcation...")
}
