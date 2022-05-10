package conf

import (
	"strings"

	"github.com/megaease/easeprobe/global"
	"gopkg.in/natefinch/lumberjack.v2"

	log "github.com/sirupsen/logrus"
)

// LogLevel is the log level
type LogLevel log.Level

var levelToString  = map[LogLevel]string{
	LogLevel(log.DebugLevel): "debug",
	LogLevel(log.InfoLevel):  "info",
	LogLevel(log.WarnLevel):  "warn",
	LogLevel(log.ErrorLevel): "error",
	LogLevel(log.FatalLevel): "fatal",
	LogLevel(log.PanicLevel): "panic",
}

var stringToLevel = map[string]LogLevel{
	"debug": LogLevel(log.DebugLevel),
	"info":  LogLevel(log.InfoLevel),
	"warn":  LogLevel(log.WarnLevel),
	"error": LogLevel(log.ErrorLevel),
	"fatal": LogLevel(log.FatalLevel),
	"panic": LogLevel(log.PanicLevel),
}


// MarshalYAML is marshal the format
func (l *LogLevel) MarshalYAML() ([]byte, error) {
	return []byte(levelToString[*l]), nil
}

// UnmarshalYAML is unmarshal the debug level
func (l *LogLevel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var level string
	if err := unmarshal(&level); err != nil {
		return err
	}
	*l = stringToLevel[strings.ToLower(level)]
	return nil
}

// GetLevel return the log level
func (l *LogLevel) GetLevel() log.Level {
	return log.Level(*l)
}

// Log is the log settings
type Log struct {
	Level      LogLevel `yaml:"level"`
	File       string   `yaml:"file"`
	MaxSize    int      `yaml:"size"`
	MaxAge     int      `yaml:"age"`
	MaxBackups int      `yaml:"backups"`
	Compress   bool     `yaml:"compress"`
}

// NewLog create a new Log
func NewLog() Log {
	return Log{
		File:       "",
		Level:      LogLevel(log.InfoLevel),
		MaxSize:    global.DefaultMaxLogSize,
		MaxAge:     global.DefaultMaxLogAge,
		MaxBackups: global.DefaultMaxBackups,
		Compress:   true,
	}
}

// CheckDefault initialize the Log configuration
func (l *Log) CheckDefault() {
	if l.MaxAge == 0 {
		l.MaxAge = global.DefaultMaxLogAge
	}
	if l.MaxSize == 0 {
		l.MaxSize = global.DefaultMaxLogSize
	}
	if l.MaxBackups == 0 {
		l.MaxBackups = global.DefaultMaxBackups
	}
	if l.Level == 0 {
		l.Level = LogLevel(log.InfoLevel)
	}
}

// GetWriter return the log writer
func (l *Log) GetWriter() *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   l.File,
		MaxSize:    l.MaxSize, // megabytes
		MaxBackups: l.MaxBackups,
		MaxAge:     l.MaxAge, //days
		Compress:   l.Compress,
	}

}
