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

package slack

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

// NotifyConfig is the slack notification configuration
type NotifyConfig struct {
	Name       string        `yaml:"name"`
	WebhookURL string        `yaml:"webhook"`
	Dry        bool          `yaml:"dry"`
	Timeout    time.Duration `yaml:"timeout"`
	Retry      global.Retry  `yaml:"retry"`
}

// Kind return the type of Notify
func (c *NotifyConfig) Kind() string {
	return "slack"
}

// Config configures the slack notification
func (c *NotifyConfig) Config(gConf global.NotifySettings) error {
	if c.Dry {
		log.Infof("Notification [%s] - [%s]  is running on Dry mode!", c.Kind(), c.Name)
	}

	c.Timeout = gConf.NormalizeTimeOut(c.Timeout)
	c.Retry = gConf.NormalizeRetry(c.Retry)

	log.Infof("[%s] configuration: %+v", c.Kind(), c)
	return nil
}

// Notify write the message into the slack
func (c *NotifyConfig) Notify(result probe.Result) {
	if c.Dry {
		c.DryNotify(result)
		return
	}
	json := BlockJSON(&result)
	c.SendSlackNotificationWithRetry("Notification", json)
}

// NotifyStat write the all probe stat message to slack
func (c *NotifyConfig) NotifyStat(probers []probe.Prober) {
	if c.Dry {
		c.DryNotifyStat(probers)
		return
	}
	json := StatSlackBlockJSON(probers)

	c.SendSlackNotificationWithRetry("SLA", json)

}

// DryNotify just log the notification message
func (c *NotifyConfig) DryNotify(result probe.Result) {
	log.Infof("[%s / %s] - %s", c.Kind(), c.Name, BlockJSON(&result))
}

// DryNotifyStat just log the notification message
func (c *NotifyConfig) DryNotifyStat(probers []probe.Prober) {
	log.Infof("[%s / %s] - %s", c.Kind(), c.Name, StatSlackBlockJSON(probers))
}

// SendSlackNotificationWithRetry send the Slack notification with retry
func (c *NotifyConfig) SendSlackNotificationWithRetry(tag string, msg string) {

	fn := func() error {
		log.Debugf("[%s - %s] - %s", c.Kind(), tag, msg)
		return c.SendSlackNotification(msg)
	}

	err := global.DoRetry(c.Kind(), c.Name, tag, c.Retry, fn)
	probe.LogSend(c.Kind(), c.Name, tag, "", err)
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

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error response from Slack [%d] - [%s]", resp.StatusCode, string(buf))
	}
	// if buf.String() != "ok" {
	// 	return errors.New("Non-ok response returned from Slack " + buf.String())
	// }
	return nil
}

// BlockJSON convert the object to Slack notification
// Go to https://app.slack.com/block-kit-builder to build the notification block
func BlockJSON(r *probe.Result) string {

	json := `
	{
		"channel": "Alert",
		"text": "%s",
		"blocks": [
			{
				"type": "section",
				"text": {
					"type": "mrkdwn",
					"text": "%s"
				}
			},
			{
				"type": "context",
				"elements": [
					{
						"type": "image",
						"image_url": "` + global.Icon + `",
						"alt_text": "` + global.OrgProg + `"
					},
					{
						"type": "mrkdwn",
						"text": "` + global.Prog + ` %s"
					}
				]
			}
		]
	}
	`
	rtt := r.RoundTripTime.Round(time.Millisecond)
	body := fmt.Sprintf("*%s*\\n>%s %s - ⏱ %s\n>%s",
		r.Title(), r.Status.Emoji(), r.Endpoint, rtt, probe.JSONEscape(r.Message))
	context := TimeFormation(r.StartTime, " probed at ", r.TimeFormat)
	summary := fmt.Sprintf("%s %s - %s", r.Title(), r.Status.Emoji(), probe.JSONEscape(r.Message))
	return fmt.Sprintf(json, summary, body, context)
}

// StatSlackBlockSectionJSON return the slack json format string to stat
func StatSlackBlockSectionJSON(r *probe.Result) string {

	json := `
			{
				"type": "mrkdwn",
				"text": "*%s* - %s` +
		`\n>*Availability*\n>\t` + " *Up*:  `%s`  *Down* `%s`  -  *SLA*: `%.2f %%`" +
		`\n>*Probe Times*\n>\t*Total* : %d ( %s )` +
		`\n>*Latest Probe*\n>\t%s | %s` +
		`\n>\t%s"` + `
			}`

	t := TimeFormation(r.StartTime, "", r.TimeFormat)

	message := probe.JSONEscape(r.Message)
	if r.Status != probe.StatusUp {
		message = "`" + message + "`"
	}

	return fmt.Sprintf(json, r.Name, r.Endpoint,
		probe.DurationStr(r.Stat.UpTime), probe.DurationStr(r.Stat.DownTime), r.SLA(),
		r.Stat.Total, probe.StatStatusText(r.Stat, probe.MakerdownSocial),
		t, r.Status.Emoji()+" "+r.Status.String(), message)
}

// StatSlackBlockJSON generate all probes stat message to slack block string
func StatSlackBlockJSON(probers []probe.Prober) string {
	sla := 0.0
	for _, p := range probers {
		sla += p.Result().SLA()
	}
	sla /= float64(len(probers))
	summary := fmt.Sprintf("Total %d Services, Average %.2f%% SLA", len(probers), sla)
	json := `{
		"channel": "Report",
		"text": "Daily Overall SLA Report - ` + summary + ` ",
		"blocks": [
		{
			"type": "header",
			"text": {
				"type": "plain_text",
				"text": "Overall SLA Report",
				"emoji": true
			}
		}`

	sectionHead := `
		{
		"type": "section",
		"fields": [`
	sectionFoot := `
				]
		}`
	total := len(probers)
	pageCnt := 10
	pages := total / pageCnt
	if total%pageCnt > 0 {
		pages++
	}

	for p := 0; p < pages; p++ {
		start := p * pageCnt
		end := (p + 1) * pageCnt
		if len(probers) < end {
			end = len(probers)
		}
		json += "," + sectionHead
		for i := start; i < end-1; i++ {
			json += StatSlackBlockSectionJSON(probers[i].Result()) + ","
		}
		json += StatSlackBlockSectionJSON(probers[end-1].Result())
		json += sectionFoot
	}

	context := `,
	{
		"type": "context",
		"elements": [
			{
				"type": "image",
				"image_url": "` + global.Icon + `",
				"alt_text": "` + global.OrgProg + `"
			},
			{
				"type": "mrkdwn",
				"text": "` + global.Prog + ` %s"
			}
		]
	}`

	timeFmt := "2006-01-02 15:04:05"
	if len(probers) > 0 {
		timeFmt = probers[len(probers)-1].Result().TimeFormat
	}
	time := TimeFormation(time.Now(), " reported at ", timeFmt)
	json += fmt.Sprintf(context, time)

	json += `]}`

	return json
}

// TimeFormation return the slack time formation
func TimeFormation(t time.Time, act string, format string) string {
	return fmt.Sprintf("<!date^%d^%s{date_num} {time_secs}|%s%s>",
		t.Unix(), act, act, t.UTC().Format(format))
}
