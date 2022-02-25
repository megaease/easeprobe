package notify

import (
	"github.com/megaease/easeprobe/notify/email"
	"github.com/megaease/easeprobe/notify/log"
	"github.com/megaease/easeprobe/notify/slack"
	"github.com/megaease/easeprobe/probe"
)

//Config is the notify configuration
type Config struct {
	Log   []log.NotifyConfig   `yaml:"log"`
	Email []email.NotifyConfig `yaml:"email"`
	Slack []slack.NotifyConfig `yaml:"slack"`
}

// Notify is the configuration of the Notify
type Notify interface {
	Kind() string
	Config() error 
	Notify(probe.Result)
}
