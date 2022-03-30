/*
 * Copyright (c) 2022, MegaEase
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/client"
	"github.com/megaease/easeprobe/probe/http"
	"github.com/megaease/easeprobe/probe/shell"
	"github.com/megaease/easeprobe/probe/tcp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var config *Conf

// Get return the global configuration
func Get() *Conf {
	return config
}

// LogLevel is the log level
type LogLevel struct {
	Level log.Level
}

// UnmarshalYAML is unmarshal the debug level
func (l *LogLevel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var level string
	if err := unmarshal(&level); err != nil {
		return err
	}
	switch strings.ToLower(level) {
	case "debug":
		l.Level = log.DebugLevel
	case "info":
		l.Level = log.InfoLevel
	case "warn":
		l.Level = log.WarnLevel
	case "error":
		l.Level = log.ErrorLevel
	case "fatal":
		l.Level = log.FatalLevel
	case "panic":
		l.Level = log.PanicLevel
	}
	return nil
}

// Schedule is the schedule.
type Schedule int

//
const (
	Hourly Schedule = iota
	Daily
	Weekly
	Monthly
	None
)

// UnmarshalYAML is unmarshal the debug level
func (s *Schedule) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var level string
	if err := unmarshal(&level); err != nil {
		return err
	}
	switch strings.ToLower(level) {
	case "hourly":
		*s = Hourly
	case "daily":
		*s = Daily
	case "weekly":
		*s = Weekly
	case "monthly":
		*s = Monthly
	default:
		*s = None
	}
	return nil
}

// Notify is the settings of notification
type Notify struct {
	Retry global.Retry `yaml:"retry"`
	Dry   bool         `yaml:"dry"`
}

// Probe is the settings of prober
type Probe struct {
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}

// SLAReport is the settings for SLA report
type SLAReport struct {
	Schedule Schedule `yaml:"schedule"`
	Time     string   `yaml:"time"`
	Debug    bool     `yaml:"debug"`
}

// Settings is the EaseProbe configuration
type Settings struct {
	LogFile    string    `yaml:"logfile"`
	LogLevel   LogLevel  `yaml:"loglevel"`
	TimeFormat string    `yaml:"timeformat"`
	Probe      Probe     `yaml:"probe"`
	Notify     Notify    `yaml:"notify"`
	SLAReport  SLAReport `yaml:"sla"`
	logfile    *os.File  `yaml:"-"`
}

// Conf is Probe configuration
type Conf struct {
	HTTP     []http.HTTP     `yaml:"http"`
	TCP      []tcp.TCP       `yaml:"tcp"`
	Shell    []shell.Shell   `yaml:"shell"`
	Client   []client.Client `yaml:"client"`
	Notify   notify.Config   `yaml:"notify"`
	Settings Settings        `yaml:"settings"`
}

// New read the configuration from yaml
func New(conf *string) (Conf, error) {
	c := Conf{
		HTTP:   []http.HTTP{},
		TCP:    []tcp.TCP{},
		Shell:  []shell.Shell{},
		Client: []client.Client{},
		Notify: notify.Config{},
		Settings: Settings{
			LogFile:    "",
			LogLevel:   LogLevel{log.InfoLevel},
			TimeFormat: "2006-01-02 15:04:05 UTC",
			Probe: Probe{
				Interval: time.Second * 60,
				Timeout:  time.Second * 10,
			},
			Notify: Notify{
				Retry: global.Retry{
					Times:    3,
					Interval: time.Second * 5,
				},
				Dry: false,
			},
			SLAReport: SLAReport{
				Schedule: Daily,
				Time:     "00:00",
				Debug:    false,
			},
			logfile: nil,
		},
	}
	y, err := ioutil.ReadFile(*conf)
	if err != nil {
		log.Errorf("error: %v ", err)
		return c, err
	}

	err = yaml.Unmarshal(y, &c)
	if err != nil {
		log.Errorf("error: %v", err)
		return c, err
	}

	c.initLog()

	config = &c

	log.Infoln("Load the configuration file successfully!")
	if log.GetLevel() >= log.DebugLevel {
		s, err := json.MarshalIndent(c, "", "  ")
		if err != nil {
			log.Debugf("%+v", c)
		} else {
			log.Debugf("%s", string(s))
		}
	}
	return c, err
}

func (conf *Conf) initLog() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	if conf == nil {
		log.SetOutput(os.Stdout)
		log.SetLevel(log.InfoLevel)
	} else {
		// open a file
		f, err := os.OpenFile(conf.Settings.LogFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0660)
		if err != nil {
			log.Warnf("Cannot open log file: %v", err)
			log.Infoln("Using Standard Output as the log output...")
			log.SetOutput(os.Stdout)
		} else {
			conf.Settings.logfile = f
			log.SetOutput(f)
		}
		log.SetLevel(conf.Settings.LogLevel.Level)
	}
}

// CloseLogFile close the log file
func (conf *Conf) CloseLogFile() {
	if conf.Settings.logfile != nil {
		conf.Settings.logfile.Close()
	}
}

// AllProbers return all probers
func (conf *Conf) AllProbers() []probe.Prober {
	// Probers
	var probers []probe.Prober

	for i := 0; i < len(conf.HTTP); i++ {
		probers = append(probers, &conf.HTTP[i])
	}

	for i := 0; i < len(conf.TCP); i++ {
		probers = append(probers, &conf.TCP[i])
	}

	for i := 0; i < len(conf.Shell); i++ {
		probers = append(probers, &conf.Shell[i])
	}

	for i := 0; i < len(conf.Client); i++ {
		probers = append(probers, &conf.Client[i])
	}
	return probers
}

// AllNotifiers return all notifiers
func (conf *Conf) AllNotifiers() []notify.Notify {
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

	for i := 0; i < len(conf.Notify.Telegram); i++ {
		notifies = append(notifies, &conf.Notify.Telegram[i])
	}

	for i := 0; i < len(conf.Notify.AwsSNS); i++ {
		notifies = append(notifies, &conf.Notify.AwsSNS[i])
	}

	return notifies
}
