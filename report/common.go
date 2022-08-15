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
	"time"

	"github.com/megaease/easeprobe/global"
	log "github.com/sirupsen/logrus"
)

// DurationStr convert the curation to string
func DurationStr(d time.Duration) string {

	const day = time.Minute * 60 * 24

	if d < 0 {
		d *= -1
	}

	if d < day {
		return d.String()
	}

	n := d / day
	d -= n * day

	if d == 0 {
		return fmt.Sprintf("%dd", n)
	}

	return fmt.Sprintf("%dd%s", n, d)
}

// HTMLHeader return the HTML head
func HTMLHeader(title string) string {
	return `
	<html>
	<head>
		<style>
		 .head {
			background: #2442bf;
			font-weight: 900;
			color: #fff;
			padding: 6px 12px;
		 }
		 .head a:link, .head a:visited {
			color: #ff9;
			text-decoration: none;
		  }
		  .head a:hover, .head a:active {
			text-decoration: underline;
		  }
		 .data {
			background: #f6f6f6;
			padding: 6px 12px;
			color: #3b3b3b;
		 }
		 .right{
			text-align: right;
		 }
		 .center{
			text-align: center;
		 }
		</style>
		<title>` + title + `</title>
	</head>
	<body style="font-family: Montserrat, sans-serif;">
		<h1 style="font-weight: normal; letter-spacing: -1px;color: #3b3b3b;">` + title + `</h1>`
}

// HTMLFooter return the HTML footer
func HTMLFooter(time string) string {
	footer := "<img src=\"" + global.GetEaseProbe().IconURL + "\" width=16 height=16> "
	footer += global.FooterString() + " at " + time
	return `
	<p>` + footer + `</p>
	</body>
	</html>`
}

// AutoRefreshJS return the auto refresh javascript
func AutoRefreshJS(time string) string {
	return `<script>
    function autoRefresh() {
        window.location = window.location.href;
    }
    setInterval('autoRefresh()', ` + time + `);
	</script>`
}

// LogSend is helper function to log the send logs.
func LogSend(kind, name, tag, message string, err error) {
	if len(message) <= 0 {
		message = " " + message + " "
	}
	if err != nil {
		log.Errorf("[%s / %s / %s] - %s(%v) - failed to send! ", kind, name, tag, message, err)
	} else {
		log.Infof("[%s / %s / %s] - %s - successfully sent!", kind, name, tag, message)
	}
}

// SlackTimeFormation return the slack time formation
func SlackTimeFormation(t time.Time, act string, format string) string {
	return fmt.Sprintf("<!date^%d^%s{date_num} {time_secs}|%s%s>",
		t.Unix(), act, act, t.UTC().Format(format))
}

// JSONEscape escape the string
func JSONEscape(str string) string {
	b, err := json.Marshal(str)
	if err != nil {
		return str
	}
	s := string(b)
	return s[1 : len(s)-1]
}

// FormatTime format the time with time zone and time format
func FormatTime(t time.Time) string {
	tf := global.GetTimeFormat()
	tz := global.GetTimeLocation()
	return t.In(tz).Format(tf)
}
