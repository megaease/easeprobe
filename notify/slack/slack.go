package slack

import (
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the slack notification configuration
type NotifyConfig struct {
	URL string `yaml:"url"`
}

// Kind return the type of Notify
func (conf NotifyConfig) Kind() string {
	return "slack"
}

// Config configures the log files
func (conf NotifyConfig) Config() error {
	return nil
}

// Notify write the message into the file
func (conf NotifyConfig) Notify(result probe.Result) {
	log.Infoln("Slack got the notifcation...")
}
