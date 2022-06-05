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
	"math/rand"
	"testing"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/stretchr/testify/assert"
)

func newDummyResult(name string) probe.Result {

	var total int64 = 100
	up := rand.Int63n(total)
	down := rand.Int63n(total - up)

	return probe.Result{
		Name:             name,
		Endpoint:         "http://endpoint:8080",
		StartTime:        time.Now(),
		StartTimestamp:   time.Now().Unix(),
		RoundTripTime:    100 * time.Millisecond,
		Status:           probe.StatusUp,
		PreStatus:        probe.StatusDown,
		Message:          "dummy message",
		LatestDownTime:   time.Now(),
		RecoveryDuration: 20 * time.Second,
		Stat: probe.Stat{
			Since: time.Now(),
			Total: total,
			Status: map[probe.Status]int64{
				probe.StatusUp:      up,
				probe.StatusDown:    down,
				probe.StatusUnknown: total - up - down,
			},
			UpTime:   time.Duration(up) * time.Second,
			DownTime: time.Duration(down) * time.Second,
		},
		TimeFormat: time.RFC3339,
	}
}

func TestToText(t *testing.T) {
	global.InitEaseProbe("EaseProbe", "http://icon/url")
	r := newDummyResult("dummy")
	str := ToText(r)
	assert.NotEmpty(t, str)
	assert.Contains(t, str, r.Title())
	assert.Contains(t, str, r.Endpoint)
	assert.Contains(t, str, r.StartTime.Format(r.TimeFormat))
	assert.Contains(t, str, r.RoundTripTime.String())
	assert.Contains(t, str, r.Status.Emoji())
	assert.NotContains(t, str, r.PreStatus.Emoji())
	assert.Contains(t, str, r.Message)
	assert.Contains(t, str, r.LatestDownTime.Format(r.TimeFormat))
	assert.Contains(t, str, global.FooterString())
}

func checkResult(t *testing.T, r probe.Result, rDTO resultDTO) {
	assert.Equal(t, r.Title(), rDTO.Name)
	assert.Equal(t, r.Endpoint, rDTO.Endpoint)
	assert.Equal(t, r.StartTime.Format(r.TimeFormat), rDTO.StartTime.Format(r.TimeFormat))
	assert.Equal(t, r.StartTimestamp, rDTO.StartTimestamp)
	assert.Equal(t, r.RoundTripTime.Round(time.Millisecond), rDTO.RoundTripTime)
	assert.Equal(t, r.Status, rDTO.Status)
	assert.Equal(t, r.PreStatus, rDTO.PreStatus)
	assert.Equal(t, r.Message, rDTO.Message)
}
func TestResultToJSON(t *testing.T) {
	r := newDummyResult("dummy")
	str := ToJSON(r)
	var rDTO resultDTO
	err := json.Unmarshal([]byte(str), &rDTO)
	assert.Nil(t, err)
	checkResult(t, r, rDTO)
	str = ToJSONIndent(r)
	err = json.Unmarshal([]byte(str), &rDTO)
	assert.Nil(t, err)
	checkResult(t, r, rDTO)
}

func TestResultToHTML(t *testing.T) {
	global.InitEaseProbe("EaseProbe", "http://icon/url")
	r := newDummyResult("dummy")
	str := ToHTML(r)
	assert.NotEmpty(t, str)
	assert.Contains(t, str, "EaseProbe")
	assert.Contains(t, str, r.Title())
	assert.Contains(t, str, global.GetEaseProbe().IconURL)
}

func checkMarkdown(t *testing.T, str string, r probe.Result) {
	assert.NotEmpty(t, str)
	assert.Contains(t, str, "EaseProbe")
	assert.Contains(t, str, r.Title())
	assert.Contains(t, str, r.Status.Emoji())
	assert.Contains(t, str, r.Message)
	assert.Contains(t, str, r.StartTime.Format(r.TimeFormat))
	assert.Contains(t, str, r.RoundTripTime.String())
	assert.Contains(t, str, r.LatestDownTime.Format(r.TimeFormat))
	assert.Contains(t, str, global.FooterString())
}

func TestResultToMarkdown(t *testing.T) {
	global.InitEaseProbe("EaseProbe", "http://icon/url")
	r := newDummyResult("dummy")
	str := ToMarkdown(r)
	checkMarkdown(t, str, r)

	str = ToMarkdownSocial(r)
	checkMarkdown(t, str, r)
}

func TestResultToSlack(t *testing.T) {
	global.InitEaseProbe("EaseProbe", "http://icon/url")
	r := newDummyResult("dummy")
	str := ToSlack(r)
	assert.NotEmpty(t, str)
	assert.Contains(t, str, "EaseProbe")
	assert.Contains(t, str, r.Title())
	assert.Contains(t, str, r.Status.Emoji())
	assert.Contains(t, str, r.Message)
	context := SlackTimeFormation(r.StartTime, " probed at ", r.TimeFormat)
	assert.Contains(t, str, context)
	assert.Contains(t, str, r.RoundTripTime.String())
	assert.Contains(t, str, global.FooterString())
}

func TestResultToLark(t *testing.T) {
	global.InitEaseProbe("EaseProbe", "http://icon/url")
	r := newDummyResult("dummy")
	str := ToLark(r)
	assert.NotEmpty(t, str)
	assert.Contains(t, str, "EaseProbe")
	assert.Contains(t, str, r.Title())
	assert.Contains(t, str, r.Status.Emoji())
	assert.Contains(t, str, r.Message)
	assert.Contains(t, str, r.RoundTripTime.String())
	assert.Contains(t, str, global.FooterString())

	assert.Contains(t, str, "green")

	r.Status = probe.StatusDown
	str = ToLark(r)
	assert.Contains(t, str, "red")

	r.Status = probe.StatusUnknown
	str = ToLark(r)
	assert.Contains(t, str, "gray")

	r.Status = probe.StatusInit
	str = ToLark(r)
	assert.Contains(t, str, "blue")
}
