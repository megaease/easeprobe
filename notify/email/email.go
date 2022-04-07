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
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify/base"
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the email notification configuration
type NotifyConfig struct {
	base.DefaultNotify `yaml:",inline"`
	Server             string `yaml:"server"`
	User               string `yaml:"username"`
	Pass               string `yaml:"password"`
	To                 string `yaml:"to"`
}

// Kind return the type of Notify
func (c *NotifyConfig) Kind() string {
	return c.MyKind
}

// Config configures the log files
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	c.MyKind = "email"
	c.Format = probe.HTML
	c.SendFunc = c.SendMail
	c.DefaultNotify.Config(gConf)
	log.Debugf("Notification [%s] - [%s] configuration: %+v", c.MyKind, c.Name, c)
	return nil
}

// SendMail sends the email
func (c *NotifyConfig) SendMail(subject string, message string) error {

	host, _, err := net.SplitHostPort(c.Server)
	if err != nil {
		return err
	}

	email := "Notification" + "<" + c.User + ">"
	header := make(map[string]string)
	header["From"] = email
	header["To"] = c.To
	header["Subject"] = subject
	header["Content-Type"] = "text/html; charset=UTF-8"

	body := ""
	for k, v := range header {
		body += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	body += "\r\n" + message

	auth := smtp.PlainAuth("", c.User, c.Pass, host)

	conn, err := tls.Dial("tcp", c.Server, nil)
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	// Auth
	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err = client.Auth(auth); err != nil {
				log.Errorln(err)
				return err
			}
		}
	}

	// To && From
	if err = client.Mail(c.User); err != nil {
		return err
	}

	// support "," and ";"
	split := func(r rune) bool {
		return r == ';' || r == ','
	}
	for _, addr := range strings.FieldsFunc(c.To, split) {

		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	// Data
	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(body))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}
