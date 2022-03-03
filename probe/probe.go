package probe

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// Status is the status of Probe
type Status int

// The status of a probe
const (
	StatusUp Status = iota
	StatusDown
	StatusUnknown
	StatusInit
)

// String convert the Status to string
func (s Status) String() string {
	switch s {
	case StatusUp:
		return "up"
	case StatusDown:
		return "down"
	case StatusUnknown:
		return "unknown"
	case StatusInit:
		return "init"
	}
	return "unknown"
}

//Status convert the string to Status
func (s *Status) Status(status string) Status {
	switch status {
	case "up":
		return StatusUp
	case "down":
		return StatusDown
	case "unknown":
		return StatusUnknown
	case "init":
		return StatusInit
	}
	return StatusUnknown
}

// Emoji convert the status to emoji
func (s Status) Emoji() string {
	switch s {
	case StatusUp:
		return "✅"
	case StatusDown:
		return "❌"
	case StatusUnknown:
		return "⛔️"
	case StatusInit:
		return "🔎"
	}
	return "⛔️"
}

// UnmarshalJSON is Unmarshal the status
func (s *Status) UnmarshalJSON(b []byte) (err error) {
	*s = s.Status(string(b))
	return nil
}

// MarshalJSON is marshal the status
func (s *Status) MarshalJSON() (b []byte, err error) {
	return []byte(fmt.Sprintf(`"%s"`, s.String())), nil
}

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

// Prober Interface
type Prober interface {
	Kind() string
	Config() error
	Probe() Result
	Interval() time.Duration
	Result() *Result
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
		log.Printf("error: %v\n", err)
		return ""
	}
	return string(j)
}

// JSONIndent convert the object to indent JSON
func (r *Result) JSONIndent() string {
	j, err := json.MarshalIndent(&r, "", "    ")
	if err != nil {
		log.Printf("error: %v\n", err)
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
	html := `
		<html>
		<head>
			<style>
			 .head {
				background: #2980b9;
				font-weight: 900;
    			color: #ffffff;
				padding: 6px 12px;
				text-align: right;
			 }
			 .data {
				background: #f6f6f6;
				padding: 6px 12px;
				color: #3b3b3b;
			 }
			</style>
		</head>
		<body style="font-family: Montserrat, sans-serif;">
			<h1 style="font-weight: normal; letter-spacing: -1px;color: #3b3b3b;">%s</h1>
			<table style="font-size: 16px; line-height: 20px;">
				<tr>
					<td class="head"><b> Service  Name </b></td>
					<td class="data">%s</td>
				</tr>
				<tr>
					<td class="head"><b> Endpoint </b></td>
					<td class="data">%s</td>
				</tr>
				<tr>
					<td class="head"><b> Status </b></td>
					<td class="data">%s - %s</td>
				</tr>
				<tr>
					<td class="head"><b> Probe Time </b></td>
					<td class="data">%s</td>
				</tr>
				<tr>
					<td class="head"><b> Round Trip Time </b></td>
					<td class="data">%s</td>
				</tr>
				<tr>
					<td class="head"><b> Message </b></td>
					<td class="data">%s</td>
				</tr>
			</table>
		</body>
		</html>`

	title := r.Title()
	rtt := r.RoundTripTime.Round(time.Millisecond)

	return fmt.Sprintf(html, title, r.Name, r.Endpoint, r.Status.Emoji(), r.Status.String(), r.StartTime, rtt, r.Message)
}

// SlackBlockJSON convert the object to Slack notification
// Go to https://app.slack.com/block-kit-builder to build the notification block
func (r *Result) SlackBlockJSON() string {

	json := `
	{
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
						"image_url": "https://megaease.cn/favicon.png",
						"alt_text": "MegaEase EaseProbe"
					},
					{
						"type": "mrkdwn",
						"text": "EaseProbe %s"
					}
				]
			}
		]
	}
	`
	rtt := r.RoundTripTime.Round(time.Millisecond)
	body := fmt.Sprintf("*%s*\\n>%s %s - ⏱ %s\n>%s", r.Title(), r.Status.Emoji(), r.Endpoint, rtt, JSONEscape(r.Message))
	context := fmt.Sprintf("<!date^%d^probed at {date_num} {time_secs} | probed at %s>",
		r.StartTime.Unix(), r.StartTime.UTC().Format(time.UnixDate))
	return fmt.Sprintf(json, body, context)
}

// StatHTMLSection return the HTML format string to stat
func (r *Result) StatHTMLSection() string {
	return ""
}

// StatSlackBlockSectionJSON return the slack json format string to stat
func (r *Result) StatSlackBlockSectionJSON() string {

	json := `
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*%s* - %s"
			}
		},
		{
			"type": "section",
			"fields": [
				{
					"type": "mrkdwn",
					"text": "*Lastest Probe Status & Time*\n>%s | %s"
				},
				{
					"type": "mrkdwn",
					"text": "*Latest Probe Message*\n>%s"
				},
				{
					"type": "mrkdwn",
					"text": "*Availability*\n>✅ *%s*  ❌ *%s*   -   *SLA*: ` + "`%.2f %%`" + `"
				},
				{
					"type": "mrkdwn",
					"text": "*Probe Times*\n>*Total* : %d ( %s )"
				}
			]
		}`

	status := ""
	for k, v := range r.Stat.Status {
		status += fmt.Sprintf("*%s* : %d \t", k.String(), v)
	}

	t := fmt.Sprintf("<!date^%d^{date_num} {time_secs} |  %s>",
		r.StartTime.Unix(), r.StartTime.UTC().Format(time.UnixDate))

	return fmt.Sprintf(json, r.Name, r.Endpoint,
		t, r.Status.Emoji()+" "+r.Status.String(), JSONEscape(r.Message),
		r.Stat.UpTime.Round(time.Second), r.Stat.DownTime.Round(time.Second),
		(r.Stat.UpTime.Seconds()/(r.Stat.UpTime.Seconds()+r.Stat.DownTime.Seconds()))*100,
		r.Stat.Total, strings.TrimSpace(status))
}

// StatSlackBlockJSON generate all probes stat message to slack block string
func StatSlackBlockJSON(probers []Prober) string {
	json := `{"blocks": [
		{
			"type": "header",
			"text": {
				"type": "plain_text",
				"text": "Overall SLA Report",
				"emoji": true
			}
		}`
	for i := 0; i < len(probers)-1; i++ {
		json += "," + probers[i].Result().StatSlackBlockSectionJSON()
		json += `,
		{
			"type": "divider"
		}`
	}
	if len(probers) > 0 {
		json += "," + probers[len(probers)-1].Result().StatSlackBlockSectionJSON()
		context := `,
		{
			"type": "context",
			"elements": [
				{
					"type": "image",
					"image_url": "https://megaease.cn/favicon.png",
					"alt_text": "MegaEase EaseProbe"
				},
				{
					"type": "mrkdwn",
					"text": "EaseProbe %s"
				}
			]
		}`
		time := fmt.Sprintf("<!date^%d^report at {date_num} {time_secs} | report at %s>",
			time.Now().Unix(), time.Now().UTC().Format(time.UnixDate))
		json += fmt.Sprintf(context, time)
	}

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
