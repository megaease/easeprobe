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

package log

import (
	"errors"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/report"
	"github.com/stretchr/testify/assert"
)

func assertError(t *testing.T, err error, msg string, contain bool) {
	assert.Error(t, err)
	if contain {
		assert.Contains(t, err.Error(), msg)
	} else {
		assert.Equal(t, msg, err.Error())
	}
}

func TestLogFile(t *testing.T) {
	conf := &NotifyConfig{}
	conf.NotifyName = "test"
	conf.File = "notify.log"
	err := conf.Config(global.NotifySettings{})
	assert.NoError(t, err)
	assert.Equal(t, report.Log, conf.NotifyFormat)
	assert.Equal(t, "log", conf.Kind())

	err = conf.Log("title", "message")
	assert.NoError(t, err)

	os.RemoveAll(conf.File)

	monkey.Patch(os.OpenFile, func(_ string, _ int, _ os.FileMode) (*os.File, error) {
		return nil, errors.New("open file error")
	})
	err = conf.Config(global.NotifySettings{})
	assertError(t, err, "open file error", false)

	monkey.UnpatchAll()
}
