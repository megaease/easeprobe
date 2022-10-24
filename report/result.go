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

package report

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"

	log "github.com/sirupsen/logrus"
)

// ToLog convert the result object to Log format
func ToLog(r probe.Result) string {
	tpl := `title="%s"; name="%s"; status="%s"; endpoint="%s"; rtt="%s"; time="%s"; message="%s"`
	rtt := r.RoundTripTime.Round(time.Millisecond)
	return fmt.Sprintf(tpl,
		r.Title(), r.Name, r.Status.String(), r.Endpoint, rtt, FormatTime(r.StartTime), r.Message)
}

// ToText convert the result object to ToText
func ToText(r probe.Result) string {
	tpl := "[%s] %s\n%s - ⏱ %s\n%s\n%s"
	rtt := r.RoundTripTime.Round(time.Millisecond)
	return fmt.Sprintf(tpl,
		r.Title(), r.Status.Emoji(), r.Endpoint, rtt, r.Message,
		global.FooterString()+" at "+FormatTime(r.StartTime))
}

// resultDTO only for JSON format notification
type resultDTO struct {
	Name           string        `json:"name"`
	Endpoint       string        `json:"endpoint"`
	StartTime      time.Time     `json:"time"`
	StartTimestamp int64         `json:"timestamp"`
	RoundTripTime  time.Duration `json:"rtt"`
	Status         probe.Status  `json:"status"`
	PreStatus      probe.Status  `json:"prestatus"`
	Message        string        `json:"message"`
}

// ToJSON convert the result object to ToJSON
func ToJSON(r probe.Result) string {
	ro := resultDTO{
		Name:           r.Title(),
		Endpoint:       r.Endpoint,
		StartTime:      r.StartTime,
		StartTimestamp: r.StartTimestamp,
		RoundTripTime:  r.RoundTripTime,
		Status:         r.Status,
		PreStatus:      r.PreStatus,
		Message:        r.Message,
	}
	j, err := json.Marshal(&ro)
	if err != nil {
		log.Errorf("error: %v", err)
		return ""
	}
	return string(j)
}

// ToJSONIndent convert the object to indent JSON
func ToJSONIndent(r probe.Result) string {
	ro := resultDTO{
		Name:           r.Title(),
		Endpoint:       r.Endpoint,
		StartTime:      r.StartTime,
		StartTimestamp: r.StartTimestamp,
		RoundTripTime:  r.RoundTripTime,
		Status:         r.Status,
		PreStatus:      r.PreStatus,
		Message:        r.Message,
	}
	j, err := json.MarshalIndent(&ro, "", "    ")
	if err != nil {
		log.Errorf("error: %v", err)
		return ""
	}
	return string(j)
}

// ToHTML convert the object to ToHTML
func ToHTML(r probe.Result) string {
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
		` + HTMLFooter(FormatTime(r.StartTime))

	rtt := r.RoundTripTime.Round(time.Millisecond)
	return fmt.Sprintf(html, r.Name, r.Endpoint, r.Status.Emoji(), r.Status.String(),
		FormatTime(r.StartTime), rtt, r.Message)
}

// ToMarkdown convert the object to ToMarkdown
func ToMarkdown(r probe.Result) string {
	return markdown(r, Markdown)
}

// ToMarkdownSocial convert the object to Markdown
func ToMarkdownSocial(r probe.Result) string {
	return markdown(r, MarkdownSocial)
}

func markdown(r probe.Result, f Format) string {
	tpl := "**%s** %s\n%s - ⏱ %s\n%s\n> %s"
	if f == MarkdownSocial {
		tpl = "*%s* %s\n%s - ⏱ %s\n%s\n> %s"
	}
	rtt := r.RoundTripTime.Round(time.Millisecond)
	return fmt.Sprintf(tpl,
		r.Title(), r.Status.Emoji(), r.Endpoint, rtt, r.Message,
		global.FooterString()+" at "+FormatTime(r.StartTime))
}

// ToSlack convert the object to ToSlack notification
// Go to https://app.slack.com/block-kit-builder to build the notification block
func ToSlack(r probe.Result) string {

	json := `
	{
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
						"image_url": "` + global.GetEaseProbe().IconURL + `",
						"alt_text": "` + global.OrgProg + `"
					},
					{
						"type": "mrkdwn",
						"text": "` + global.FooterString() + ` %s"
					}
				]
			}
		]
	}
	`
	rtt := r.RoundTripTime.Round(time.Millisecond)
	body := fmt.Sprintf("*%s*\\n>%s %s - ⏱ %s\n>%s",
		r.Title(), r.Status.Emoji(), r.Endpoint, rtt, JSONEscape(r.Message))
	context := SlackTimeFormation(r.StartTime, " probed at ", global.GetTimeFormat())
	summary := fmt.Sprintf("%s %s - %s", r.Title(), r.Status.Emoji(), JSONEscape(r.Message))
	return fmt.Sprintf(json, summary, body, context)
}

// ToLark convert the object to Lark notification
// Go to https://open.feishu.cn/document/ukTMukTMukTM/ucTM5YjL3ETO24yNxkjN#4996824a to build the notification block
func ToLark(r probe.Result) string {
	json := `
	{
		"msg_type": "interactive",
		"card": {
			"config": {
				"wide_screen_mode": true
			},
			"header": {
				"template": "%s",
				"title": {
				"content": "%s",
				"tag": "plain_text"
				}
			},
			"elements": [
				{
					"tag": "div",
					"text": {
						"content": "%s",
						"tag": "lark_md"
					}
				},
				{
					"tag": "hr"
				},
				{
					"tag": "note",
					"elements": [
						{
							"tag": "plain_text",
							"content": "%s"
						}
					]
				}
			]
		}
	}`

	headerColor := "gray"
	switch r.Status {
	case probe.StatusUp:
		headerColor = "green"
	case probe.StatusDown:
		headerColor = "red"
	case probe.StatusUnknown:
		headerColor = "gray"
	case probe.StatusInit:
		headerColor = "blue"
	}

	title := fmt.Sprintf("%s %s", r.Title(), r.Status.Emoji())
	rtt := r.RoundTripTime.Round(time.Millisecond)
	content := fmt.Sprintf("%s - ⏱ %s\\n%s", r.Endpoint, rtt, JSONEscape(r.Message))
	footer := global.FooterString() + " probed at " + FormatTime(r.StartTime)
	return fmt.Sprintf(json, headerColor, title, content, footer)
}

// ToCSV convert the object to CSV
func ToCSV(r probe.Result) string {
	rtt := fmt.Sprintf("%d", r.RoundTripTime.Round(time.Millisecond))
	time := FormatTime(r.StartTime)
	timestamp := fmt.Sprintf("%d", r.StartTimestamp)
	data := [][]string{
		{"Title", "Name", "Endpoint", "Status", "PreStatus", "RoundTripTime", "Time", "Timestamp", "Message"},
		{r.Title(), r.Name, r.Endpoint, r.Status.String(), r.PreStatus.String(), rtt, time, timestamp, r.Message},
	}

	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)

	if err := w.WriteAll(data); err != nil {
		log.Errorf("ToCSV(): Failed to write to csv buffer: %v", err)
		return ""
	}
	return buf.String()
}

// ToShell convert the result object to shell variables
func ToShell(r probe.Result) string {
	env := make(map[string]string)

	// set the notify type variable
	env["EASEPROBE_TYPE"] = "Status"

	// set individual variables
	env["EASEPROBE_TITLE"] = r.Title()
	env["EASEPROBE_NAME"] = r.Name
	env["EASEPROBE_ENDPOINT"] = r.Endpoint
	env["EASEPROBE_STATUS"] = r.Status.String()
	env["EASEPROBE_TIMESTAMP"] = fmt.Sprintf("%d", r.StartTimestamp)
	env["EASEPROBE_TIME"] = FormatTime(r.StartTime)
	env["EASEPROBE_RTT"] = fmt.Sprintf("%d", r.RoundTripTime.Round(time.Millisecond))
	env["EASEPROBE_MESSAGE"] = r.Message

	// set JSON and CVS format
	env["EASEPROBE_JSON"] = ToJSON(r)
	env["EASEPROBE_CSV"] = ToCSV(r)

	buf, err := json.Marshal(env)
	if err != nil {
		log.Errorf("ToShell(): Failed to marshal env to json: %s", err)
		return ""
	}
	return string(buf)
}
