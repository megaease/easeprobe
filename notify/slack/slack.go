package slack

import (
	"bytes"
	"errors"
	"net/http"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the slack notification configuration
type NotifyConfig struct {
	WebhookURL string       `yaml:"webhook"`
	Dry        bool         `yaml:"dry"`
	Retry      global.Retry `yaml:"retry"`
}

// Kind return the type of Notify
func (c *NotifyConfig) Kind() string {
	return "slack"
}

// Config configures the log files
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	if c.Dry {
		log.Infof("Notification %s is running on Dry mode!", c.Kind())
	}

	if c.Retry.Interval <= 0 {
		c.Retry.Interval = global.DefaultRetryInterval
		if gConf.Retry.Interval > 0 {
			c.Retry.Interval = gConf.Retry.Interval
		}
	}

	if c.Retry.Times <= 0 {
		c.Retry.Times = global.DefaultRetryTimes
		if gConf.Retry.Times >= 0 {
			c.Retry.Times = gConf.Retry.Times
		}
	}

	log.Infof("[%s] configuration: %+v", c.Kind(), c)
	return nil
}

// Notify write the message into the slack
func (c *NotifyConfig) Notify(result probe.Result) {
	if c.Dry {
		c.DryNotify(result)
		return
	}
	json := result.SlackBlockJSON()
	c.SendSlackNotificationWithRetry("Notification", json)
}

// NotifyStat write the all probe stat message to slack
func (c *NotifyConfig) NotifyStat(probers []probe.Prober) {
	if c.Dry {
		c.DryNotifyStat(probers)
		return
	}
	json := probe.StatSlackBlockJSON(probers)
	c.SendSlackNotificationWithRetry("SLA", json)

}

// DryNotify just log the notification message
func (c *NotifyConfig) DryNotify(result probe.Result) {
	log.Infof("[%s] - %s", c.Kind(), result.SlackBlockJSON())
}

// DryNotifyStat just log the notification message
func (c *NotifyConfig) DryNotifyStat(probers []probe.Prober) {
	log.Infof("[%s] - %s", c.Kind(), probe.StatSlackBlockJSON(probers))
}

// SendSlackNotificationWithRetry send the Slack notification with retry
func (c *NotifyConfig) SendSlackNotificationWithRetry(tag string, msg string) {

	for i := 0; i < c.Retry.Times; i++ {
		err := c.SendSlackNotification(msg)
		if err == nil {
			log.Infof("Successfully Sent the %s to Slack", tag)
			return
		}

		log.Debugf("[%s] - %s", c.Kind(), msg)
		log.Warnf("[%s] Retred to send %d/%d -  %v", c.Kind(), i+1, c.Retry.Times, err)

		// last time no need to sleep
		if i < c.Retry.Times-1 {
			time.Sleep(c.Retry.Interval)
		}

	}
	log.Errorf("[%s] Failed to sent the slack after %d retries!", c.Kind(), c.Retry.Times)

}

// SendSlackNotification will post to an 'Incoming Webhook' url setup in Slack Apps. It accepts
// some text and the slack channel is saved within Slack.
func (c *NotifyConfig) SendSlackNotification(msg string) error {
	req, err := http.NewRequest(http.MethodPost, c.WebhookURL, bytes.NewBuffer([]byte(msg)))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Close = true

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if resp.StatusCode != 200 {
		return errors.New(buf.String())
	}
	// if buf.String() != "ok" {
	// 	return errors.New("Non-ok response returned from Slack " + buf.String())
	// }
	return nil
}
