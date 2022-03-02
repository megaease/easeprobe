package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the email notification configuration
type NotifyConfig struct {
	Server string `yaml:"server"`
	User   string `yaml:"username"`
	Pass   string `yaml:"password"`
	To     string `yaml:"to"`
}

// Kind return the type of Notify
func (conf NotifyConfig) Kind() string {
	return "email"
}

// Config configures the log files
func (conf NotifyConfig) Config() error {
	return nil
}

// Notify write the message into the file
func (conf NotifyConfig) Notify(result probe.Result) {
	log.Infoln("Email got the notification...")

	mesage := fmt.Sprintf("%s", result.JSONIndent())
	if err := conf.SendMail("EaseProbe Notification", mesage); err != nil {
		log.Errorln(err)
	}
}

// SendMail sends the email
func (conf NotifyConfig) SendMail(subject string, message string) error {

	host, _, err := net.SplitHostPort(conf.Server)
	if err != nil {
		return err
	}

	email := "Notification" + "<" + conf.User + ">"
	header := make(map[string]string)
	header["From"] = email
	header["To"] = conf.To
	header["Subject"] = subject
	header["Content-Type"] = "text/html; charset=UTF-8"

	body := ""
	for k, v := range header {
		body += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	body += "\r\n" + message

	auth := smtp.PlainAuth("", conf.User, conf.Pass, host)

	conn, err := tls.Dial("tcp", conf.Server, nil)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer c.Close()

	// Auth
	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				log.Errorln(err)
				return err
			}
		}
	}

	// To && From
	if err = c.Mail(conf.User); err != nil {
		return err
	}

	// support "," and ";"
	split := func(r rune) bool {
		return r == ';' || r == ','
	}
	for _, addr := range strings.FieldsFunc(conf.To, split) {

		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	// Data
	w, err := c.Data()
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

	return c.Quit()
}
