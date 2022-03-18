package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify"
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
	switch level {
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

// Settings is the EaseProbe configuration
type Settings struct {
	LogFile    string   `yaml:"logfile"`
	LogLevel   LogLevel `yaml:"loglevel"`
	Debug      bool     `yaml:"debug"`
	TimeFormat string   `yaml:"timeformat"`
	Probe      Probe    `yaml:"probe"`
	Notify     Notify   `yaml:"notify"`
	logfile    *os.File `yaml:"-"`
}

// Conf is Probe configuration
type Conf struct {
	HTTP     []http.HTTP   `yaml:"http"`
	TCP      []tcp.TCP     `yaml:"tcp"`
	Shell    []shell.Shell `yaml:"shell"`
	Notify   notify.Config `yaml:"notify"`
	Settings Settings      `yaml:"settings"`
}

// New read the configuration from yaml
func New(conf *string) (Conf, error) {
	c := Conf{
		HTTP:   []http.HTTP{},
		TCP:    []tcp.TCP{},
		Notify: notify.Config{},
		Settings: Settings{
			LogFile:    "",
			LogLevel:   LogLevel{log.InfoLevel},
			Debug:      false,
			TimeFormat: "2006-01-02 15:04:05 UTC",
			Probe: Probe{
				Interval: time.Second * 60,
			},
			Notify: Notify{
				Retry: global.Retry{
					Times:    3,
					Interval: time.Second * 5,
				},
				Dry: false,
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
			log.Warnf("Error when opening log file: %v", err)
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
