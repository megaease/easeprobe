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

// Package email is the email notification package
package email

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/base"
	"github.com/megaease/easeprobe/report"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

// NotifyConfig is the email notification configuration
type NotifyConfig struct {
	base.DefaultNotify `yaml:",inline"`
	Server             string `yaml:"server" json:"server" jsonschema:"required,format=hostname,title=SMTP Server,description=SMTP server with port,example=\"smtp.example.com:465\""`
	User               string `yaml:"username" json:"username" jsonschema:"required,title=SMTP Username,description=SMTP username,example=\"name@example.com\""`
	Pass               string `yaml:"password" json:"password" jsonschema:"required,title=SMTP Password,description=SMTP password,example=\"password\""`
	To                 string `yaml:"to" json:"to" jsonschema:"required,title=To,description=Email address to send,example=\"usera@example.com;userb@example.com\""`
	From               string `yaml:"from,omitempty" json:"from,omitempty" jsonschema:"title=From,description=Email address from,example=\"from@example.com\""`
}

// Config configures the log files
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.NotifyKind = "email"
	c.NotifyFormat = report.HTML
	c.NotifySendFunc = c.SendMail
	c.DefaultNotify.Config(gConf)
	log.Debugf("Notification [%s] - [%s] configuration: %+v", c.NotifyKind, c.NotifyName, c)
	return nil
}

// SendMail sends the email
func (c *NotifyConfig) SendMail(subject string, message string) error {

	host, p, err := net.SplitHostPort(c.Server)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(p)
	if err != nil {
		return err
	}

	email := "Notification" + "<" + c.User + ">"
	if c.From != "" {
		email = c.From
	}

	split := func(r rune) bool {
		return r == ';' || r == ','
	}

	recipients := strings.FieldsFunc(c.To, split)

	m := gomail.NewMessage()
	m.SetHeader("From", email)
	m.SetHeader("To", recipients...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html; charset=UTF-8", message)

	d := gomail.NewDialer(host, port, c.User, c.Pass)
	err = d.DialAndSend(m)
	if err != nil {
		return fmt.Errorf("[%s / %s] - Error response from mail with body <%s>, %v", c.Kind(), c.Name(), message, err)
	}

	return nil
}
