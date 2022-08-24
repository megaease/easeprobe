//go:build !windows

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
	"log/syslog"
	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/stretchr/testify/assert"
)

func TestLocalSyslog(t *testing.T) {
	conf := &NotifyConfig{}
	conf.File = "syslog"
	err := conf.Config(global.NotifySettings{})
	assert.NoError(t, err)

	err = conf.Log("title", "message")
	assert.NoError(t, err)

	monkey.Patch(syslog.New, func(priority syslog.Priority, tag string) (*syslog.Writer, error) {
		return nil, errors.New("new syslog error")
	})
	err = conf.Config(global.NotifySettings{})
	assertError(t, err, "new syslog error", false)

	monkey.UnpatchAll()
}

func TestNetworkSyslog(t *testing.T) {
	conf := &NotifyConfig{}
	conf.NotifyName = "dummy"
	conf.File = "syslog"
	conf.Network = "tcp"
	conf.Host = "localhost:514"

	monkey.Patch(syslog.Dial, func(_, _ string, _ syslog.Priority, _ string) (*syslog.Writer, error) {
		return &syslog.Writer{}, nil
	})

	err := conf.Config(global.NotifySettings{})
	assert.NoError(t, err)

	monkey.Patch(syslog.Dial, func(_, _ string, _ syslog.Priority, _ string) (*syslog.Writer, error) {
		return nil, errors.New("dial syslog error")
	})
	err = conf.Config(global.NotifySettings{})
	assertError(t, err, "dial syslog error", false)

	conf.Host = "localhost:port"
	err = conf.Config(global.NotifySettings{})
	assertError(t, err, "invalid port", true)

	conf.Host = "localhost"
	err = conf.Config(global.NotifySettings{})
	assertError(t, err, "invalid host", true)

	conf.Network = "unknown"
	err = conf.Config(global.NotifySettings{})
	assertError(t, err, "invalid protocol", true)

	conf.Host = ""
	err = conf.checkNetworkProtocol()
	assertError(t, err, "host is required", true)
	conf.Network = ""
	err = conf.checkNetworkProtocol()
	assertError(t, err, "protocol is required", true)

	monkey.UnpatchAll()
}
