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
	"io"
	"os"
	"strings"

	"github.com/megaease/easeprobe/global"
	"gopkg.in/natefinch/lumberjack.v2"

	log "github.com/sirupsen/logrus"
)

// LogLevel is the log level
type LogLevel log.Level

var levelToString = map[LogLevel]string{
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
func (l LogLevel) MarshalYAML() (interface{}, error) {
	return levelToString[l], nil
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
	Level      LogLevel    `yaml:"level"`
	File       string      `yaml:"file"`
	SelfRotate bool        `yaml:"self_rotate"`
	MaxSize    int         `yaml:"size"`
	MaxAge     int         `yaml:"age"`
	MaxBackups int         `yaml:"backups"`
	Compress   bool        `yaml:"compress"`
	Writer     io.Writer   `yaml:"-"`
	Logger     *log.Logger `yaml:"-"`
	IsStdout   bool        `yaml:"-"`
}

// NewLog create a new Log
func NewLog() Log {
	return Log{
		Level:      LogLevel(log.InfoLevel),
		File:       "",
		SelfRotate: true,
		MaxSize:    global.DefaultMaxLogSize,
		MaxAge:     global.DefaultMaxLogAge,
		MaxBackups: global.DefaultMaxBackups,
		Compress:   true,
		Writer:     nil,
		Logger:     nil,
		IsStdout:   true,
	}
}

// InitLog initialize the log
func (l *Log) InitLog(logger *log.Logger) {
	l.Logger = logger
	l.CheckDefault()
	if l.File != "" {
		l.File = global.MakeDirectory(l.File)
	}
	l.Open()
	l.ConfigureLogger()
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

// Open open the log file
func (l *Log) Open() {
	// using stdout if no log file
	if l.File == "" {
		l.IsStdout = true
		l.Writer = os.Stdout
		return
	}
	// using lumberjack if self rotate
	if l.SelfRotate == true {
		log.Debugf("[Log] Self Rotate log file %s", l.File)
		l.IsStdout = false
		l.Writer = &lumberjack.Logger{
			Filename:   l.File,
			MaxSize:    l.MaxSize, // megabytes
			MaxBackups: l.MaxBackups,
			MaxAge:     l.MaxAge, //days
			Compress:   l.Compress,
		}
		return
	}
	// using log file if not self rotate
	f, err := os.OpenFile(l.File, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0640)
	if err != nil {
		log.Warnf("[Log] Cannot open log file: %v", err)
		log.Infoln("[Log] Using Standard Output as the log output...")
		l.IsStdout = true
		l.Writer = os.Stdout
		return
	}
	l.IsStdout = false
	l.Writer = f
}

// Close close the log file
func (l *Log) Close() {
	if l.Writer == nil || l.IsStdout {
		return
	}
	if f, ok := l.Writer.(*os.File); ok {
		f.Close()
	}
}

// GetWriter return the log writer
func (l *Log) GetWriter() io.Writer {
	if l.Writer == nil {
		l.Open()
	}
	return (io.Writer)(l.Writer)
}

// Rotate rotate the log file
func (l *Log) Rotate() {
	if l.Writer == nil || l.IsStdout == true {
		return
	}
	if lumberjackLogger, ok := l.Writer.(*lumberjack.Logger); ok {
		// self rotate
		if err := lumberjackLogger.Rotate(); err != nil {
			log.Errorf("[Log] Rotate log file failed: %s", err)
		}
	} else if fileLogger, ok := l.Writer.(*os.File); ok {
		// rotate managed by outside program (e.g. logrotate)
		// just close and open current log file
		if err := fileLogger.Close(); err != nil {
			log.Errorf("[Log] Close log file failed: %s", err)
		}
		l.Open()            // open another writer
		l.ConfigureLogger() // set the new logger writer.
	}

}

// ConfigureLogger configure the logger
func (l *Log) ConfigureLogger() {
	if l.Logger != nil {
		l.Logger.SetOutput(l.Writer)
		l.Logger.SetLevel(l.Level.GetLevel())
		l.Logger.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	} else { //system-wide log
		log.SetOutput(l.Writer)
		log.SetLevel(l.Level.GetLevel())
		log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	}
}

// LogInfo log info
func (l *Log) LogInfo(name string) {
	logger := log.New()
	rotate := "Third-Party Rotate (e.g. logrotate)"
	if l.SelfRotate {
		rotate = "Self-Rotate"
	}
	if l.File != "" {
		logger.Infof("%s Log File [%s] - %s", name, l.File, rotate)
	} else {
		logger.Infof("%s Log File [Stdout] - %s ", name, rotate)
	}
}
