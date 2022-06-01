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
	"testing"

	"bou.ke/monkey"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

func TestLogLevel(t *testing.T) {
	l := LogLevel(log.DebugLevel)
	buf, err := yaml.Marshal(l)
	assert.Nil(t, err)
	assert.Equal(t, "debug\n", string(buf))

	err = yaml.Unmarshal([]byte("painc"), &l)
	assert.Nil(t, err)
	assert.Equal(t, LogLevel(log.PanicLevel), l)

	assert.Equal(t, l.GetLevel(), log.PanicLevel)
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
	defer monkey.UnpatchAll()

	l := NewLog()
	l.File = "test.log"
	l.SelfRotate = false
	l.InitLog(nil)
	assert.Equal(t, true, l.IsStdout)
	assert.Equal(t, os.Stdout, l.GetWriter())
}
