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
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"bou.ke/monkey"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

func testLogYaml(t *testing.T, name string, level LogLevel, good bool) {
	var l LogLevel
	err := yaml.Unmarshal([]byte(name), &l)
	if good {
		assert.Nil(t, err)
		assert.Equal(t, level, l)
	} else {
		assert.NotNil(t, err)
		assert.Equal(t, LogLevel(log.PanicLevel), l)
	}

	buf, err := yaml.Marshal(level)
	if good {
		assert.Nil(t, err)
		assert.Equal(t, name+"\n", string(buf))
	} else {
		assert.NotNil(t, err)
		assert.Nil(t, buf)
	}

}

func TestLogLevelYaml(t *testing.T) {
	testLogYaml(t, "debug", LogLevel(log.DebugLevel), true)
	testLogYaml(t, "info", LogLevel(log.InfoLevel), true)
	testLogYaml(t, "warn", LogLevel(log.WarnLevel), true)
	testLogYaml(t, "error", LogLevel(log.ErrorLevel), true)
	testLogYaml(t, "fatal", LogLevel(log.FatalLevel), true)
	testLogYaml(t, "panic", LogLevel(log.PanicLevel), true)
	testLogYaml(t, "none", LogLevel(log.Level(100)), false)
	testLogYaml(t, "- none", LogLevel(log.Level(100)), false)
}

func TestLogLevel(t *testing.T) {
	l := LogLevel(log.DebugLevel)
	buf, err := yaml.Marshal(l)
	assert.Nil(t, err)
	assert.Equal(t, "debug\n", string(buf))

	err = yaml.Unmarshal([]byte("panic"), &l)
	assert.Nil(t, err)
	assert.Equal(t, LogLevel(log.PanicLevel), l)

	assert.Equal(t, l.GetLevel(), log.PanicLevel)

	// test error yaml format
	err = yaml.Unmarshal([]byte("- log::error"), &l)
	assert.NotNil(t, err)
	assert.Equal(t, log.PanicLevel, l.GetLevel())
}

func testLogs(t *testing.T, name string, new bool, app bool, selfRotate bool) {
	var l Log
	if new {
		l = NewLog()
	}

	l.SelfRotate = selfRotate

	file := name + ".log"
	l.File = file

	if app {
		l.InitLog(log.New())
		l.Logger.Info(name)
	} else {
		l.InitLog(nil)
		log.Info(name)
	}
	assert.FileExists(t, file)

	l.LogInfo(name)
	l.Rotate()

	match, err := filepath.Glob(name + "-*")
	assert.Nil(t, err)
	if l.SelfRotate {
		_, ok := l.GetWriter().(*lumberjack.Logger)
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(match[0]), 1)
	} else {
		_, ok := l.GetWriter().(*os.File)
		assert.True(t, ok)
		assert.Equal(t, 0, len(match))
	}

	l.Close()
	os.RemoveAll(file)
	for _, f := range match {
		os.RemoveAll(f)
	}
}

func TestAppLog(t *testing.T) {
	testLogs(t, "test", true, true, true)
}

func TestLog(t *testing.T) {
	testLogs(t, "easeprobe", true, false, true)
}

func TestNonSelfRotateLog(t *testing.T) {
	testLogs(t, "my", false, false, false)
}

func TestOpenLogFail(t *testing.T) {
	monkey.Patch(os.OpenFile, func(name string, flag int, perm os.FileMode) (*os.File, error) {
		return nil, fmt.Errorf("error")
	})

	file := "failed"

	l := NewLog()
	l.File = file + ".log"
	l.SelfRotate = false
	l.InitLog(nil)
	assert.Equal(t, true, l.IsStdout)
	assert.Equal(t, os.Stdout, l.GetWriter())

	l.Close()
	l.Writer = nil
	l.Rotate()

	w := l.GetWriter()
	assert.Equal(t, os.Stdout, w)

	monkey.UnpatchAll()

	// test rotate error - log file
	l = NewLog()
	l.File = file + ".log"
	l.SelfRotate = false
	l.InitLog(nil)
	assert.Equal(t, false, l.IsStdout)

	var fp *os.File
	monkey.PatchInstanceMethod(reflect.TypeOf(fp), "Close", func(_ *os.File) error {
		return fmt.Errorf("error")
	})

	l.Rotate()
	files, _ := filepath.Glob(file + "*")
	fmt.Println(files)
	assert.Equal(t, 1, len(files))
	l.Close()
	os.Remove(l.File)

	// test rotate error - lumberjackLogger
	l = NewLog()
	l.File = file + ".log"
	l.SelfRotate = true
	l.InitLog(nil)
	assert.Equal(t, false, l.IsStdout)

	var lum *lumberjack.Logger
	monkey.PatchInstanceMethod(reflect.TypeOf(lum), "Rotate", func(_ *lumberjack.Logger) error {
		return fmt.Errorf("error")
	})
	l.Rotate()
	files, _ = filepath.Glob(file + "*")
	fmt.Println(files)
	assert.Equal(t, 1, len(files))
	l.Close()
	os.Remove(l.File)

	monkey.UnpatchAll()
}
