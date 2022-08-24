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
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/smtp"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/report"
	"github.com/stretchr/testify/assert"
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

	monkey.Patch(tls.Dial, func(_, _ string, _ *tls.Config) (*tls.Conn, error) {
		return &tls.Conn{}, nil
	})
	monkey.Patch(smtp.NewClient, func(_ net.Conn, _ string) (*smtp.Client, error) {
		return &smtp.Client{}, nil
	})
	var client *smtp.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Close", func(_ *smtp.Client) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Extension", func(_ *smtp.Client, _ string) (bool, string) {
		return true, ""
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Auth", func(_ *smtp.Client, _ smtp.Auth) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Mail", func(_ *smtp.Client, _ string) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Rcpt", func(_ *smtp.Client, _ string) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Data", func(_ *smtp.Client) (io.WriteCloser, error) {
		return &MyWriteCloser{}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Quit", func(_ *smtp.Client) error {
		return nil
	})

	conf.Server = "smtp.example.com:25"
	err = conf.SendMail("title", "message")
	assert.NoError(t, err)

	var w *MyWriteCloser
	monkey.PatchInstanceMethod(reflect.TypeOf(w), "Close", func(_ *MyWriteCloser) error {
		return errors.New("close error")
	})
	err = conf.SendMail("title", "message")
	assertError(t, err, "close error")

	monkey.PatchInstanceMethod(reflect.TypeOf(w), "Write", func(_ *MyWriteCloser, data []byte) (int, error) {
		return 0, errors.New("write error")
	})
	err = conf.SendMail("title", "message")
	assertError(t, err, "write error")

	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Data", func(_ *smtp.Client) (io.WriteCloser, error) {
		return nil, errors.New("data error")
	})
	err = conf.SendMail("title", "message")
	assertError(t, err, "data error")

	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Rcpt", func(_ *smtp.Client, _ string) error {
		return errors.New("rcpt error")
	})
	conf.To = "test@emai.com;user@email.com"
	err = conf.SendMail("title", "message")
	assertError(t, err, "rcpt error")

	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Mail", func(_ *smtp.Client, _ string) error {
		return errors.New("mail error")
	})
	err = conf.SendMail("title", "message")
	assertError(t, err, "mail error")

	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Auth", func(_ *smtp.Client, _ smtp.Auth) error {
		return errors.New("auth error")
	})
	err = conf.SendMail("title", "message")
	assertError(t, err, "auth error")

	monkey.Patch(smtp.NewClient, func(_ net.Conn, _ string) (*smtp.Client, error) {
		return nil, errors.New("new client error")
	})
	err = conf.SendMail("title", "message")
	assertError(t, err, "new client error")

	monkey.Patch(tls.Dial, func(_, _ string, _ *tls.Config) (*tls.Conn, error) {
		return nil, errors.New("dial error")
	})
	err = conf.SendMail("title", "message")
	assertError(t, err, "dial error")

	monkey.UnpatchAll()
}
