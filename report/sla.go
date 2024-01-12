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

// Package report is the package for SLA report
package report

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
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
			SLA:      r.SLAPercent(),
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
		r := probe.GetResultData(p.Name())
		sla = append(sla, SLAObject(r))
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
		DurationStr(r.Stat.UpTime), DurationStr(r.Stat.DownTime), r.SLAPercent(),
		r.Stat.Total, SLAStatusText(r.Stat, Text),
		FormatTime(r.StartTime),
		r.Status.Emoji()+" "+r.Status.String(), JSONEscape(r.Message))
}

// SLAText return a full stat report
func SLAText(probers []probe.Prober) string {
	text := "[Overall SLA Report]\n\n"
	for _, p := range probers {
		r := probe.GetResultData(p.Name())
		text += SLATextSection(r) + "\n"
	}
	return text
}

// SLALogSection return the Log format string to stat
func SLALogSection(r *probe.Result) string {
	text := `name="%s"; endpoint="%s"; up="%s"; down="%s"; sla="%.2f%%"; total="%d(%s)"; latest_time="%s"; latest_status="%s"; message="%s"`
	return fmt.Sprintf(text, r.Name, r.Endpoint,
		DurationStr(r.Stat.UpTime), DurationStr(r.Stat.DownTime), r.SLAPercent(),
		r.Stat.Total, SLAStatusText(r.Stat, Log),
		FormatTime(r.StartTime),
		r.Status.String(), r.Message)
}

// SLALog return a full stat report with Log format
func SLALog(probers []probe.Prober) string {
	var text string
	n := len(probers)
	for i, p := range probers {
		r := probe.GetResultData(p.Name())
		text += fmt.Sprintf("SLA-Report-%d-%d %s\n", i+1, n, SLALogSection(r))
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
		DurationStr(r.Stat.UpTime), DurationStr(r.Stat.DownTime), r.SLAPercent(),
		r.Stat.Total, SLAStatusText(r.Stat, MarkdownSocial),
		FormatTime(r.StartTime),
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
		r := probe.GetResultData(p.Name())
		md += SLAMarkdownSection(r, f)
	}

	md += "\n> " + global.FooterString() + " at " + FormatTime(time.Now())
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
		r.SLAPercent(),
		r.Stat.Total, SLAStatusText(r.Stat, HTML),
		FormatTime(r.StartTime),
		r.Status.Emoji()+" "+r.Status.String(), JSONEscape(r.Message))
}

// SLAHTML return a full stat report
func SLAHTML(probers []probe.Prober) string {
	return SLAHTMLFilter(probers, nil)
}

// SLAHTMLFilter return a stat report with filter
func SLAHTMLFilter(probers []probe.Prober, filter *SLAFilter) string {
	html := HTMLHeader("Overall SLA Report")

	if filter == nil {
		filter = NewEmptyFilter()
	}

	probers = filter.Filter(probers)
	table := `<table style="font-size: 16px; line-height: 20px;">`
	for _, p := range probers {
		r := probe.GetResultData(p.Name())
		table += SLAHTMLSection(r)
	}
	table += `</table>`

	html = html + filter.HTML() + table

	html += HTMLFooter(FormatTime(time.Now()))
	return html
}

// SLASlackSection return the slack json format string to stat
func SLASlackSection(r *probe.Result) string {

	jsonMsg := `
			{
				"type": "mrkdwn",
				"text": "*%s* - %s` +
		`\n>*Availability*\n>\t` + " *Up*:  `%s`  *Down* `%s`  -  *SLA*: `%.2f %%`" +
		`\n>*Probe Times*\n>\t*Total* : %d ( %s )` +
		`\n>*Latest Probe*\n>\t%s | %s` +
		`\n>\t%s"` + `
			}`

	t := SlackTimeFormation(r.StartTime, "", global.GetTimeFormat())

	message := JSONEscape(r.Message)
	if r.Status != probe.StatusUp {
		message = "`" + message + "`"
	}

	output := fmt.Sprintf(jsonMsg, r.Name, JSONEscape(r.Endpoint),
		DurationStr(r.Stat.UpTime), DurationStr(r.Stat.DownTime), r.SLAPercent(),
		r.Stat.Total, SLAStatusText(r.Stat, MarkdownSocial),
		t, r.Status.Emoji()+" "+r.Status.String(), message)
	if !json.Valid([]byte(output)) {
		log.Errorf("SLASlackSection() for %s: invalid json: %s", r.Name, output)
	}
	return output
}

// SLASlack generate all probes stat message to slack block string
func SLASlack(probers []probe.Prober) string {
	summary := SLASummary(probers)
	json := `{
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
			r := probe.GetResultData(probers[i].Name())
			json += SLASlackSection(r) + ","
		}
		r := probe.GetResultData(probers[end-1].Name())
		json += SLASlackSection(r)
		json += sectionFoot
	}

	context := `,
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
	}`

	time := SlackTimeFormation(time.Now(), " reported at ", global.GetTimeFormat())
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
	case Log:
		format = "%s:%d "
	}

	// sort status
	var statusKeys []int
	for statusKey, _ := range s.Status {
		statusKeys = append(statusKeys, int(statusKey))
	}
	sort.Ints(statusKeys)
	for _, k := range statusKeys {
		status += fmt.Sprintf(format, probe.Status(k).String(), s.Status[probe.Status(k)])
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
		DurationStr(r.Stat.UpTime), DurationStr(r.Stat.DownTime), r.SLAPercent(),
		r.Stat.Total, JSONEscape(SLAStatusText(r.Stat, Lark)),
		FormatTime(r.StartTime),
		r.Status.Emoji()+" "+r.Status.String(), JSONEscape(r.Message))
}

// SLALark return a full stat report
func SLALark(probers []probe.Prober) string {
	jsonMsg := `
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
							"content": "` + global.FooterString() + `"
						}
					]
				}
			]
		}
	}`

	title := "Overall SLA Report"
	sections := []string{}
	for _, p := range probers {
		r := probe.GetResultData(p.Name())
		sections = append(sections, SLALarkSection(r))
	}

	elements := strings.Join(sections, "")
	s := fmt.Sprintf(jsonMsg, title, elements)
	if !json.Valid([]byte(s)) {
		log.Errorf("SLALark(): invalid json: %s", s)
	}

	fmt.Printf("SLA: %s\n", s)
	return s
}

// SLASummary return a summary stat report
func SLASummary(probers []probe.Prober) string {
	sla := 0.0
	for _, p := range probers {
		r := probe.GetResultData(p.Name())
		sla += r.SLAPercent()
	}
	sla /= float64(len(probers))
	summary := fmt.Sprintf("Total %d Services, Average %.2f%% SLA", len(probers), sla)
	summary += "\n" + global.FooterString()
	return summary
}

// SLACSVSection set the CSV format for SLA
func SLACSVSection(r *probe.Result) []string {
	return []string{
		// Name, Endpoint,
		r.Name, r.Endpoint,
		// UpTime, DownTime, SLA
		DurationStr(r.Stat.UpTime), DurationStr(r.Stat.DownTime), fmt.Sprintf("%.2f%%", r.SLAPercent()),
		// ProbeSummary - Total( Up, Down)
		fmt.Sprintf("%d(%s)", r.Stat.Total, SLAStatusText(r.Stat, Text)),
		// LatestProbe, LatestStatus
		FormatTime(r.StartTime), r.Status.String(),
		// Message
		r.Message,
	}
}

// SLACSV return a full stat report with CSV format
func SLACSV(probers []probe.Prober) string {
	data := [][]string{
		{"Name", "Endpoint", "UpTime", "DownTime", "SLA", "ProbeSummary", "LatestProbe", "LatestStatus", "Message"},
	}

	for _, p := range probers {
		r := probe.GetResultData(p.Name())
		data = append(data, SLACSVSection(r))
	}

	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)

	if err := w.WriteAll(data); err != nil {
		log.Errorf("SLACSV(): Failed to write to csv buffer: %v", err)
		return ""
	}

	return buf.String()
}

// SLAShell set the environment for SLA
func SLAShell(probers []probe.Prober) string {
	env := make(map[string]string)

	env["EASEPROBE_TYPE"] = "SLA"
	env["EASEPROBE_JSON"] = SLAJSON(probers)
	env["EASEPROBE_CSV"] = SLACSV(probers)

	buf, err := json.Marshal(env)
	if err != nil {
		log.Errorf("SLAShell(): Failed to marshal env to json: %s", err)
		return ""
	}
	return string(buf)
}
