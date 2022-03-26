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
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the email notification configuration
type NotifyConfig struct {
	Server  string        `yaml:"server"`
	User    string        `yaml:"username"`
	Pass    string        `yaml:"password"`
	To      string        `yaml:"to"`
	Dry     bool          `yaml:"dry"`
	Timeout time.Duration `yaml:"timeout"`
	Retry   global.Retry  `yaml:"retry"`
}

// Kind return the type of Notify
func (c *NotifyConfig) Kind() string {
	return "email"
}

// Config configures the log files
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	if c.Dry {
		log.Infof("Notification %s is running on Dry mode!", c.Kind())
	}
	c.Timeout = gConf.NormalizeTimeOut(c.Timeout)
	c.Retry = gConf.NormalizeRetry(c.Retry)

	log.Infof("[%s] configuration: %+v", c.Kind(), c)
	return nil
}

// Notify send the result message to the email
func (c *NotifyConfig) Notify(result probe.Result) {
	if c.Dry {
		c.DryNotify(result)
		return
	}
	message := fmt.Sprintf("%s", result.HTML())
	c.SendMailWithRetry(result.Title(), message, "Notification")
}

// NotifyStat send the stat message into the email
func (c *NotifyConfig) NotifyStat(probers []probe.Prober) {
	if c.Dry {
		c.DryNotifyStat(probers)
		return
	}
	message := probe.StatHTML(probers)
	c.SendMailWithRetry("Overall SLA Report", message, "SLA")
}

// DryNotify just log the notification message
func (c *NotifyConfig) DryNotify(result probe.Result) {
	log.Infof("[%s] Dry Notify - %s", c.Kind(), result.HTML())
}

// DryNotifyStat just log the notification message
func (c *NotifyConfig) DryNotifyStat(probers []probe.Prober) {
	log.Infof("[%s] Dry Notify - %s", c.Kind(), probe.StatHTML(probers))
}

// SendMailWithRetry sends the email with retry if got error
func (c *NotifyConfig) SendMailWithRetry(title string, message string, tag string) {

	for i := 0; i < c.Retry.Times; i++ {
		err := c.SendMail(title, message)
		if err == nil {
			log.Infof("Successfully Sent %s to the email - %s", tag, title)
			return
		}

		log.Debugf("[%s] - %s", c.Kind(), message)
		log.Warnf("[%s] Retred to send %d/%d -  %v", c.Kind(), i+1, c.Retry.Times, err)
		// last time no need to sleep
		if i < c.Retry.Times-1 {
			time.Sleep(c.Retry.Interval)
		}
	}
	log.Errorf("[%s] Failed to sent the %s to email after %d retries!", c.Kind(), tag, c.Retry.Times)
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
