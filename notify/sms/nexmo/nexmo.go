package nexmo

import (
	"github.com/megaease/easeprobe/notify/sms/conf"
)

// Kind is the type of Provider
const Kind string = "Nexmo"

type Nexmo struct {
	conf.Options `yaml:",inline"`

	ApiKey    string `yaml:"apikey"`
	ApiSecret string `yaml:"api_secret"`
	From      string `yaml:"from"`
}

// New create a Nexmo sms provider
func New(opt conf.Options) *Nexmo {
	return &Nexmo{
		Options: opt,
	}
}

// Kind return the type of Notify
func (c Nexmo) Kind() string {
	return Kind
}

// Notify return the type of Notify
func (c Nexmo) Notify(title, text string) error {
	//todo
	return nil
}
