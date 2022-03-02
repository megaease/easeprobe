package probe

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// Prober Interface
type Prober interface {
	Kind() string
	Config() error
	Probe() Result
	Interval() time.Duration
}

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
}

// NewResult return a Result object
func NewResult() *Result {
	return &Result{
		Name:          "",
		Endpoint:      "",
		StartTime:     time.Now(),
		RoundTripTime: ConfigDuration{0},
		Status:        StatusInit,
		PreStatus:     StatusInit,
		Message:       "",
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
	body := fmt.Sprintf("*%s*\\n>%s %s - ⏱ %s", r.Title(), r.Status.Emoji(), r.Endpoint, rtt)
	context := fmt.Sprintf("<!date^%d^probed at {date_num} {time_secs} | probed at %s>", r.StartTime.Unix(), r.StartTime)
	return fmt.Sprintf(json, body, context)
}
