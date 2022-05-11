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
	Level      LogLevel           `yaml:"level"`
	File       string             `yaml:"file"`
	MaxSize    int                `yaml:"size"`
	MaxAge     int                `yaml:"age"`
	MaxBackups int                `yaml:"backups"`
	Compress   bool               `yaml:"compress"`
	Logger     *lumberjack.Logger `yaml:"-"`
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
	if l.Logger == nil {
		l.Logger = &lumberjack.Logger{
			Filename:   l.File,
			MaxSize:    l.MaxSize, // megabytes
			MaxBackups: l.MaxBackups,
			MaxAge:     l.MaxAge, //days
			Compress:   l.Compress,
		}
	}
	return l.Logger
}

//Rotate rotate the log file
func (l *Log) Rotate() {
	if l.Logger != nil {
		if err := l.Logger.Rotate(); err != nil {
			log.Errorf("Rotate log file failed: %s", err)
		}
	}
}
