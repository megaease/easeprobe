package conf

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/megaease/easeprobe/notify"
	"github.com/megaease/easeprobe/probe/http"
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

// Settings is the EaseProbe configuration
type Settings struct {
	DefaultInvterval time.Duration `yaml:"interval"`
	LogFile          string        `yaml:"logfile"`
	LogLevel         LogLevel      `yaml:"loglevel"`
	DryNotify        bool          `yaml:"drynotify"`

	logfile *os.File `yaml:"-"`
}

// Conf is Probe configuration
type Conf struct {
	HTTP     []http.HTTP   `yaml:"http"`
	TCP      []tcp.TCP     `yaml:"tcp"`
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
			DefaultInvterval: time.Second * 60,
			LogFile:          "",
			LogLevel:         LogLevel{log.InfoLevel},
			DryNotify:        false,
			logfile:          nil,
		},
	}
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

	c.initLog()

	config = &c

	log.Infoln("Load the configuration file successfully!")
	log.Debugf("%v\n", c)
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
			log.Errorf("Error when opening log file: %v", err)
			log.Info("Using Standard Output as the log output...")
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
