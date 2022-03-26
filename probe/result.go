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

package probe

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/megaease/easeprobe/global"
	log "github.com/sirupsen/logrus"
)

// Format is the format of text
type Format int

// The format types
const (
	Makerdown Format = iota
	HTML
	Text
)

//ConfigDuration is the struct used for custom the time formation
type ConfigDuration struct {
	time.Duration
}

// UnmarshalJSON is Unmarshal the time
func (d *ConfigDuration) UnmarshalJSON(b []byte) (err error) {
	d.Duration, err = time.ParseDuration(strings.Trim(string(b), `"`))
	return
}

// MarshalJSON is marshal the time
func (d *ConfigDuration) MarshalJSON() (b []byte, err error) {
	return []byte(fmt.Sprintf(`"%s"`, d.Round(time.Millisecond))), nil
}

// Stat is the statistics of probe result
type Stat struct {
	Since    time.Time        `json:"since"`
	Total    int32            `json:"total"`
	Status   map[Status]int32 `json:"status"`
	UpTime   time.Duration    `json:"uptime"`
	DownTime time.Duration    `json:"downtime"`
}

// Result is the status of health check
type Result struct {
	Name           string         `json:"name"`
	Endpoint       string         `json:"endpoint"`
	StartTime      time.Time      `json:"time"`
	StartTimestamp int64          `json:"timestamp"`
	RoundTripTime  ConfigDuration `json:"rtt"`
	TimeFormat     string         `json:"-"`
	Status         Status         `json:"status"`
	PreStatus      Status         `json:"prestatus"`
	Message        string         `json:"message"`
	Stat           Stat           `json:"stat"`
}

// NewResult return a Result object
func NewResult() *Result {
	return &Result{
		Name:           "",
		Endpoint:       "",
		StartTime:      time.Now(),
		StartTimestamp: 0,
		RoundTripTime:  ConfigDuration{0},
		Status:         StatusInit,
		PreStatus:      StatusInit,
		Message:        "",
		Stat: Stat{
			Since:    time.Now(),
			Total:    0,
			Status:   map[Status]int32{},
			UpTime:   0,
			DownTime: 0,
		},
	}
}

// DoStat is the function do the statstics
func (r *Result) DoStat(d time.Duration) {
	r.Stat.Total++
	r.Stat.Status[r.Status]++
	if r.Status == StatusUp {
		r.Stat.UpTime += d
	} else {
		r.Stat.DownTime += d
	}
}

// JSON convert the object to JSON
func (r *Result) JSON() string {
	j, err := json.Marshal(&r)
	if err != nil {
		log.Errorf("error: %v", err)
		return ""
	}
	return string(j)
}

// JSONIndent convert the object to indent JSON
func (r *Result) JSONIndent() string {
	j, err := json.MarshalIndent(&r, "", "    ")
	if err != nil {
		log.Errorf("error: %v", err)
		return ""
	}
	return string(j)
}

// Title return the title for notification
func (r *Result) Title() string {
	t := "%s Recovery"
	if r.PreStatus == StatusInit {
		t = "Monitoring %s"
	}
	if r.Status != StatusUp {
		t = "%s Failure"
	}
	return fmt.Sprintf(t, r.Name)
}

// HTML convert the object to HTML
func (r *Result) HTML() string {
	html := HTMLHeader(r.Title()) + `
			<table style="font-size: 16px; line-height: 20px;">
				<tr>
					<td class="head right"><b> Service  Name </b></td>
					<td class="data">%s</td>
				</tr>
				<tr>
					<td class="head right"><b> Endpoint </b></td>
					<td class="data">%s</td>
				</tr>
				<tr>
					<td class="head right"><b> Status </b></td>
					<td class="data">%s %s</td>
				</tr>
				<tr>
					<td class="head right"><b> Probe Time </b></td>
					<td class="data">%s</td>
				</tr>
				<tr>
					<td class="head right"><b> Round Trip Time </b></td>
					<td class="data">%s</td>
				</tr>
				<tr>
					<td class="head right"><b> Message </b></td>
					<td class="data">%s</td>
				</tr>
			</table>
		` + HTMLFooter()

	rtt := r.RoundTripTime.Round(time.Millisecond)
	return fmt.Sprintf(html, r.Name, r.Endpoint, r.Status.Emoji(), r.Status.String(),
		r.StartTime.Format(r.TimeFormat), rtt, r.Message)
}

// SlackBlockJSON convert the object to Slack notification
// Go to https://app.slack.com/block-kit-builder to build the notification block
func (r *Result) SlackBlockJSON() string {

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
		r.Title(), r.Status.Emoji(), r.Endpoint, rtt, JSONEscape(r.Message))
	context := SlackTimeFormation(r.StartTime, " probed at ", r.TimeFormat)
	summary := fmt.Sprintf("%s %s - %s", r.Title(), r.Status.Emoji(), JSONEscape(r.Message))
	return fmt.Sprintf(json, summary, body, context)
}

// StatText return the Text format string to stat
func (r *Result) StatText() string {
	text := "Name: %s - %s, Availability: Up - %s, Down - %s, SLA: %.2f%%, Probe-Times: Total: %d ( %s ), Last-Probe:%s - %s, Message:%s"
	return fmt.Sprintf(text, r.Name, r.Endpoint,
		DurationStr(r.Stat.UpTime), DurationStr(r.Stat.DownTime), r.SLA(),
		r.Stat.Total, StatStatusText(r.Stat, Text),
		time.Now().UTC().Format(r.TimeFormat),
		r.Status.Emoji()+" "+r.Status.String(), JSONEscape(r.Message))
}

// StatHTMLSection return the HTML format string to stat
func (r *Result) StatHTMLSection() string {

	html := `
	<tr>
		<td class="head" colspan="3"><b>%s</b> - %s<td>
	</tr>
	<tr>
		<td class="data"><b>Availability</b><br><b>Uptime: </b>%s,  <b>Downtime: </b>%s  </td>
		<td class="data"><b>SLA<b><br>%.2f%%</td>
		<td class="data"><b>Probe-Times</b><br><b>Total</b>: %d ( %s )</td>
	</tr>
	<tr>
		<td  class="data" colspan="3"><b>Last Probe</b>: %s - %s<br>%s<td>
	</tr>
	`
	return fmt.Sprintf(html, r.Name, r.Endpoint,
		DurationStr(r.Stat.UpTime), DurationStr(r.Stat.DownTime),
		r.SLA(),
		r.Stat.Total, StatStatusText(r.Stat, HTML),
		time.Now().UTC().Format(r.TimeFormat),
		r.Status.Emoji()+" "+r.Status.String(), JSONEscape(r.Message))
}

// StatHTML return a full stat report
func StatHTML(probers []Prober) string {
	html := HTMLHeader("Overall SLA Report")

	html += `<table style="font-size: 16px; line-height: 20px;">`
	for _, p := range probers {
		html += p.Result().StatHTMLSection()
	}
	html += `</table>`

	html += HTMLFooter()
	return html
}

// StatSlackBlockSectionJSON return the slack json format string to stat
func (r *Result) StatSlackBlockSectionJSON() string {

	json := `
			{
				"type": "mrkdwn",
				"text": "*%s* - %s` +
		`\n>*Availability*\n>\t` + " *Up*:  `%s`  *Down* `%s`  -  *SLA*: `%.2f %%`" +
		`\n>*Probe Times*\n>\t*Total* : %d ( %s )` +
		`\n>*Lastest Probe*\n>\t%s | %s` +
		`\n>\t%s"` + `
			}`

	t := SlackTimeFormation(r.StartTime, "", r.TimeFormat)

	message := JSONEscape(r.Message)
	if r.Status != StatusUp {
		message = "`" + message + "`"
	}

	return fmt.Sprintf(json, r.Name, r.Endpoint,
		DurationStr(r.Stat.UpTime), DurationStr(r.Stat.DownTime), r.SLA(),
		r.Stat.Total, StatStatusText(r.Stat, Makerdown),
		t, r.Status.Emoji()+" "+r.Status.String(), message)
}

// StatSlackBlockJSON generate all probes stat message to slack block string
func StatSlackBlockJSON(probers []Prober) string {
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
			json += probers[i].Result().StatSlackBlockSectionJSON() + ","
		}
		json += probers[end-1].Result().StatSlackBlockSectionJSON()
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

	time := SlackTimeFormation(time.Now(), " reported at ", probers[len(probers)-1].Result().TimeFormat)
	json += fmt.Sprintf(context, time)

	json += `]}`

	return json
}

//JSONEscape escape the string
func JSONEscape(str string) string {
	b, err := json.Marshal(str)
	if err != nil {
		return str
	}
	s := string(b)
	return s[1 : len(s)-1]
}

// SLA calculate the SLA
func (r *Result) SLA() float64 {
	uptime := r.Stat.UpTime.Seconds()
	downtime := r.Stat.DownTime.Seconds()
	if uptime+downtime <= 0 {
		if r.Status == StatusUp {
			return 100
		}
		return 0
	}
	return uptime / (uptime + downtime) * 100
}

// StatStatusText return the string of status statices
func StatStatusText(s Stat, t Format) string {
	status := ""
	format := "%s : %d \t"
	switch t {
	case Makerdown:
		format = "**%s** : `%d` \t"
	case HTML:
		format = "<b>%s</b> : %d \t"
	}
	for k, v := range s.Status {
		status += fmt.Sprintf(format, k.String(), v)
	}
	return strings.TrimSpace(status)
}

// SlackTimeFormation return the slack time formation
func SlackTimeFormation(t time.Time, act string, format string) string {
	return fmt.Sprintf("<!date^%d^%s{date_num} {time_secs}|%s%s>",
		t.Unix(), act, act, t.UTC().Format(format))
}
