package slack

import (
	"bytes"
	"errors"
	"net/http"
	"time"

	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the slack notification configuration
type NotifyConfig struct {
	WebhookURL string `yaml:"webhook"`
}

// Kind return the type of Notify
func (conf NotifyConfig) Kind() string {
	return "slack"
}

// Config configures the log files
func (conf NotifyConfig) Config() error {
	return nil
}

// Notify write the message into the slack
func (conf NotifyConfig) Notify(result probe.Result) {
	log.Infoln("Slack got the notification...")
	json := result.SlackBlockJSON()
	err := conf.SendSlackNotification(json)
	if err != nil {
		log.Errorf("error %v\n%s", err, json)
	}
}

// NotifyStat write the all probe stat message to slack
func (conf NotifyConfig) NotifyStat(probers []probe.Prober) {
	log.Infoln("Slack  Sending the Statstics...")
	json := probe.StatSlackBlockJSON(probers)
	err := conf.SendSlackNotification(json)
	if err != nil {
		log.Errorf("error %v\n%s", err, json)
	}
}

// SendSlackNotification will post to an 'Incoming Webhook' url setup in Slack Apps. It accepts
// some text and the slack channel is saved within Slack.
func (conf NotifyConfig) SendSlackNotification(msg string) error {
	req, err := http.NewRequest(http.MethodPost, conf.WebhookURL, bytes.NewBuffer([]byte(msg)))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		return errors.New("Non-ok response returned from Slack " + buf.String())
	}
	return nil
}
