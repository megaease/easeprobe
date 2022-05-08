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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	log "github.com/sirupsen/logrus"
)

// Availability is the Availability JSON structure
type Availability struct {
	UpTime   time.Duration `json:"up"`
	DownTime time.Duration `json:"down"`
	SLA      float64       `json:"sla"`
}

// Summary is the Summary JSON structure
type Summary struct {
	Total int64 `json:"total"`
	Up    int64 `json:"up"`
	Down  int64 `json:"down"`
}

// LatestProbe is the LatestProbe JSON structure
type LatestProbe struct {
	Time    time.Time    `json:"time"`
	Status  probe.Status `json:"status"`
	Message string       `json:"message"`
}

// SLA is the SLA JSON structure
type SLA struct {
	Name         string       `json:"name"`
	Endpoint     string       `json:"endpoint"`
	Availability Availability `json:"sla"`
	ProbeTimes   Summary      `json:"probe_summary"`
	LatestProbe  LatestProbe  `json:"latest_probe"`
}

// SLAObject covert the result to SLA struct
func SLAObject(r *probe.Result) SLA {
	return SLA{
		Name:     r.Name,
		Endpoint: r.Endpoint,
		Availability: Availability{
			UpTime:   r.Stat.UpTime,
			DownTime: r.Stat.DownTime,
			SLA:      SLAPercent(r),
		},
		ProbeTimes: Summary{
			Total: r.Stat.Total,
			Up:    r.Stat.Status[probe.StatusUp],
			Down:  r.Stat.Status[probe.StatusDown] + r.Stat.Status[probe.StatusUnknown],
		},
		LatestProbe: LatestProbe{
			Time:    r.StartTime,
			Status:  r.Status,
			Message: r.Message,
		},
	}

}

// SLAPercent calculate the SLAPercent
func SLAPercent(r *probe.Result) float64 {
	uptime := r.Stat.UpTime.Seconds()
	downtime := r.Stat.DownTime.Seconds()
	if uptime+downtime <= 0 {
		if r.Status == probe.StatusUp {
			return 100
		}
		return 0
	}
	return uptime / (uptime + downtime) * 100
}

// SLAJSONSection return the JSON format string to stat
func SLAJSONSection(r *probe.Result) string {
	sla := SLAObject(r)
	j, err := json.Marshal(&sla)
	if err != nil {
		log.Errorf("error: %v", err)
		return ""
	}
	return string(j)
}

// SLAJSON return a full stat report
func SLAJSON(probers []probe.Prober) string {
	var sla []SLA
	for _, p := range probers {
		sla = append(sla, SLAObject(p.Result()))
	}
	j, err := json.Marshal(&sla)
	if err != nil {
		log.Errorf("error: %v", err)
		return ""
	}
	return string(j)
}

// SLATextSection return the Text format string to stat
func SLATextSection(r *probe.Result) string {
	text := "Name: %s - %s, \n\tAvailability: Up - %s, Down - %s, SLA: %.2f%%\n\tProbe-Times: Total: %d ( %s ), \n\tLatest-Probe:%s - %s, Message:%s"
	return fmt.Sprintf(text, r.Name, r.Endpoint,
		DurationStr(r.Stat.UpTime), DurationStr(r.Stat.DownTime), SLAPercent(r),
		r.Stat.Total, SLAStatusText(r.Stat, Text),
		time.Now().UTC().Format(r.TimeFormat),
		r.Status.Emoji()+" "+r.Status.String(), JSONEscape(r.Message))
}

// SLAText return a full stat report
func SLAText(probers []probe.Prober) string {
	text := "[Overall SLA Report]\n\n"
	for _, p := range probers {
		text += SLATextSection(p.Result()) + "\n\b"
	}
	return text
}

// SLAMarkdownSection return the Markdown format string to stat
func SLAMarkdownSection(r *probe.Result, f Format) string {

	text := "\n**%s** - %s\n"
	if f == MarkdownSocial {
		text = "\n*%s* - %s\n"
	}

	text += "- Availability: Up - `%s`, Down - `%s`, SLA: `%.2f%%` \n" +
		"- Probe-Times: Total: `%d` ( %s ) \n" +
		"- Latest-Probe: %s - %s \n" +
		"  ```%s```\n"

	return fmt.Sprintf(text, r.Name, r.Endpoint,
		DurationStr(r.Stat.UpTime), DurationStr(r.Stat.DownTime), SLAPercent(r),
		r.Stat.Total, SLAStatusText(r.Stat, MarkdownSocial),
		time.Now().UTC().Format(r.TimeFormat),
		r.Status.Emoji()+" "+r.Status.String(), r.Message)
}

// SLAMarkdown return a full stat report with Markdown format
func SLAMarkdown(probers []probe.Prober) string {
	return slaMarkdown(probers, Markdown)
}

// SLAMarkdownSocial return a full stat report with social markdown
func SLAMarkdownSocial(probers []probe.Prober) string {
	return slaMarkdown(probers, MarkdownSocial)
}

func slaMarkdown(probers []probe.Prober, f Format) string {
	md := "**Overall SLA Report**\n"
	if f == MarkdownSocial {
		md = "*Overall SLA Report*\n"
	}
	for _, p := range probers {
		md += SLAMarkdownSection(p.Result(), f)
	}
	return md
}

// SLAHTMLSection return the HTML format string to stat
func SLAHTMLSection(r *probe.Result) string {

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
		<td  class="data" colspan="3"><b>Latest Probe</b>: %s - %s<br>%s<td>
	</tr>
	`
	return fmt.Sprintf(html, r.Name, r.Endpoint,
		DurationStr(r.Stat.UpTime), DurationStr(r.Stat.DownTime),
		SLAPercent(r),
		r.Stat.Total, SLAStatusText(r.Stat, HTML),
		time.Now().UTC().Format(r.TimeFormat),
		r.Status.Emoji()+" "+r.Status.String(), JSONEscape(r.Message))
}

// SLAHTML return a full stat report
func SLAHTML(probers []probe.Prober) string {
	html := HTMLHeader("Overall SLA Report")

	html += `<table style="font-size: 16px; line-height: 20px;">`
	for _, p := range probers {
		html += SLAHTMLSection(p.Result())
	}
	html += `</table>`

	html += HTMLFooter()
	return html
}

// SLASlackSection return the slack json format string to stat
func SLASlackSection(r *probe.Result) string {

	json := `
			{
				"type": "mrkdwn",
				"text": "*%s* - %s` +
		`\n>*Availability*\n>\t` + " *Up*:  `%s`  *Down* `%s`  -  *SLA*: `%.2f %%`" +
		`\n>*Probe Times*\n>\t*Total* : %d ( %s )` +
		`\n>*Latest Probe*\n>\t%s | %s` +
		`\n>\t%s"` + `
			}`

	t := SlackTimeFormation(r.StartTime, "", r.TimeFormat)

	message := JSONEscape(r.Message)
	if r.Status != probe.StatusUp {
		message = "`" + message + "`"
	}

	return fmt.Sprintf(json, r.Name, JSONEscape(r.Endpoint),
		DurationStr(r.Stat.UpTime), DurationStr(r.Stat.DownTime), SLAPercent(r),
		r.Stat.Total, SLAStatusText(r.Stat, MarkdownSocial),
		t, r.Status.Emoji()+" "+r.Status.String(), message)
}

// SLASlack generate all probes stat message to slack block string
func SLASlack(probers []probe.Prober) string {
	summary := SLASummary(probers)
	json := `{
		"channel": "Report",
		"text": "Overall SLA Report - ` + summary + ` ",
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
			json += SLASlackSection(probers[i].Result()) + ","
		}
		json += SLASlackSection(probers[end-1].Result())
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
	time := SlackTimeFormation(time.Now(), " reported at ", timeFmt)
	json += fmt.Sprintf(context, time)

	json += `]}`

	return json
}

// SLAStatusText return the string of status statices
func SLAStatusText(s probe.Stat, t Format) string {
	status := ""
	format := "%s : %d \t"
	switch t {
	case MarkdownSocial:
		format = "%s : `%d` \t"
	case Markdown:
		format = "**%s** : `%d` \t"
	case HTML:
		format = "<b>%s</b> : %d \t"
	}
	for k, v := range s.Status {
		status += fmt.Sprintf(format, k.String(), v)
	}
	return strings.TrimSpace(status)
}

// SLALarkSection return the Text format string to stat
func SLALarkSection(r *probe.Result) string {
	text := `
	{
		"tag": "hr"
	}, {
		"tag": "div",
		"text": {
		  "content": "**Name:** %s - %s\n**Availability:** Up - %s, Down - %s\n**SLA:** %.2f%%\n**Probe-Times:** Total: %d ( %s )\n**Latest-Probe:** %s - %s\n**Message:**%s",
		  "tag": "lark_md"
		}
	},`
	return fmt.Sprintf(text, r.Name, r.Endpoint,
		DurationStr(r.Stat.UpTime), DurationStr(r.Stat.DownTime), SLAPercent(r),
		r.Stat.Total, SLAStatusText(r.Stat, Lark),
		time.Now().UTC().Format(r.TimeFormat),
		r.Status.Emoji()+" "+r.Status.String(), JSONEscape(r.Message))
}

// SLALark return a full stat report
func SLALark(probers []probe.Prober) string {
	json := `
	{
		"msg_type": "interactive",
		"card": {
			"header": {
				"template": "blue",
				"title": {
				"content": "%s",
				"tag": "plain_text"
				}
			},
			"config": {
				"wide_screen_mode": true
			},
			"elements": [%s
				{
					"tag": "hr"
				}, {
					"tag": "note",
					"elements": [
						{
							"tag": "plain_text",
							"content": global.Prog 
						}
					]
				}
			]
		}
	}`

	title := "Overall SLA Report"
	sections := []string{}
	for _, p := range probers {
		sections = append(sections, SLALarkSection(p.Result()))
	}

	elements := strings.Join(sections, "")
	s := fmt.Sprintf(json, title, elements)
	fmt.Printf("SLA: %s\n", s)
	return s
}

// SLASummary return a summary stat report
func SLASummary(probers []probe.Prober) string {
	sla := 0.0
	for _, p := range probers {
		sla += SLAPercent(p.Result())
	}
	sla /= float64(len(probers))
	summary := fmt.Sprintf("Total %d Services, Average %.2f%% SLA", len(probers), sla)
	return summary
}
