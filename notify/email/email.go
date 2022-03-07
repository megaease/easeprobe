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
	Dry    bool   `yaml:"dry"`
}

// Kind return the type of Notify
func (c NotifyConfig) Kind() string {
	return "email"
}

// Config configures the log files
func (c NotifyConfig) Config() error {
	if c.Dry {
		log.Infof("Notification %s is running on Dry mode!", c.Kind())
	}
	return nil
}

// Notify send the result message to the email
func (c NotifyConfig) Notify(result probe.Result) {
	if c.Dry {
		c.DryNotify(result)
		return
	}
	message := fmt.Sprintf("%s", result.HTML())

	if err := c.SendMail(result.Title(), message); err != nil {
		log.Errorln(err)
	}
	log.Infof("Sent the email notification for %s (%s)!", result.Name, result.Endpoint)
}

// NotifyStat send the stat message into the email
func (c NotifyConfig) NotifyStat(probers []probe.Prober) {
	if c.Dry {
		c.DryNotifyStat(probers)
		return
	}
	message := probe.StatHTML(probers)

	if err := c.SendMail("Overall SLA Report", message); err != nil {
		log.Errorln(err)
	}
	log.Infoln("Sent the Statstics to Email Successfully!")

}

// DryNotify just log the notification message
func (c NotifyConfig) DryNotify(result probe.Result) {
	log.Infoln(result.HTML())
}

// DryNotifyStat just log the notification message
func (c NotifyConfig) DryNotifyStat(probers []probe.Prober) {
	log.Infoln(probe.StatHTML(probers))
}

// SendMail sends the email
func (c NotifyConfig) SendMail(subject string, message string) error {

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
