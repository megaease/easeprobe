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

package email

import (
	"errors"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/report"
	"github.com/stretchr/testify/assert"
	"gopkg.in/gomail.v2"
)

type MyWriteCloser struct{}

func (mwc *MyWriteCloser) Close() error {
	return nil
}
func (mwc *MyWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}
func assertError(t *testing.T, err error, msg string) {
	assert.Error(t, err)
	assert.Equal(t, msg, err.Error())
}

func TestEmail(t *testing.T) {
	conf := &NotifyConfig{}
	conf.From = "easeprobe@megaease.com"
	err := conf.Config(global.NotifySettings{})
	assert.NoError(t, err)
	assert.Equal(t, report.HTML, conf.NotifyFormat)
	assert.Equal(t, "email", conf.Kind())

	err = conf.SendMail("title", "message")
	assert.Error(t, err, "missing port")

	conf.Server = "smtp.example.com:xx"
	err = conf.SendMail("title", "message")
	assert.Error(t, err, "invalid syntax")

	monkey.Patch(gomail.NewDialer, func(_ string, _ int, _, _ string) *gomail.Dialer {
		return &gomail.Dialer{}
	})

	var d *gomail.Dialer
	monkey.PatchInstanceMethod(reflect.TypeOf(d), "DialAndSend", func(_ *gomail.Dialer, _ ...*gomail.Message) error {
		return errors.New("send error")
	})
	conf.Server = "smtp.example.com:25"
	err = conf.SendMail("title", "message")
	assert.Error(t, err, "send error")

	conf.To = "test@emai.com;user@email.com"
	err = conf.SendMail("title", "message")
	assert.Error(t, err, "send error")

	monkey.PatchInstanceMethod(reflect.TypeOf(d), "DialAndSend", func(_ *gomail.Dialer, _ ...*gomail.Message) error {
		return nil
	})
	err = conf.SendMail("title", "message")
	assert.NoError(t, err)

	conf.To = "test@emai.com;user@email.com"
	err = conf.SendMail("title", "message")
	assert.NoError(t, err)

	monkey.UnpatchAll()
}
