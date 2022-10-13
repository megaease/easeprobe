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
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/stretchr/testify/assert"
)

func newDummyResult(name string) probe.Result {

	var total int64 = 100
	up := rand.Int63n(total)
	down := rand.Int63n(total - up)

	r := probe.Result{
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
	}
	probe.SetResultData(name, &r)
	return r
}

func TestToLog(t *testing.T) {
	global.InitEaseProbe("EaseProbe", "http://icon/url")
	r := newDummyResult("dummy")
	str := ToLog(r)
	assert.NotEmpty(t, str)
	assert.Contains(t, str, r.Title())
	assert.Contains(t, str, r.Endpoint)
	assert.Contains(t, str, FormatTime(r.StartTime))
	assert.Contains(t, str, r.RoundTripTime.String())
	assert.NotContains(t, str, r.PreStatus.Emoji())
	assert.Contains(t, str, r.Message)
	assert.Contains(t, str, FormatTime(r.LatestDownTime))
}

func TestToText(t *testing.T) {
	global.InitEaseProbe("EaseProbe", "http://icon/url")
	r := newDummyResult("dummy")
	str := ToText(r)
	assert.NotEmpty(t, str)
	assert.Contains(t, str, r.Title())
	assert.Contains(t, str, r.Endpoint)
	assert.Contains(t, str, FormatTime(r.StartTime))
	assert.Contains(t, str, r.RoundTripTime.String())
	assert.Contains(t, str, r.Status.Emoji())
	assert.NotContains(t, str, r.PreStatus.Emoji())
	assert.Contains(t, str, r.Message)
	assert.Contains(t, str, FormatTime(r.LatestDownTime))
	assert.Contains(t, str, global.FooterString())
}

func checkResult(t *testing.T, r probe.Result, rDTO resultDTO) {
	assert.Equal(t, r.Title(), rDTO.Name)
	assert.Equal(t, r.Endpoint, rDTO.Endpoint)
	assert.Equal(t, FormatTime(r.StartTime), FormatTime(rDTO.StartTime))
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

	monkey.Patch(json.Marshal, func(v interface{}) ([]byte, error) {
		return nil, fmt.Errorf("marshal error")
	})
	str = ToJSON(r)
	assert.Empty(t, str)

	monkey.Patch(json.MarshalIndent, func(interface{}, string, string) ([]byte, error) {
		return nil, fmt.Errorf("marshal error")
	})
	str = ToJSONIndent(r)
	assert.Empty(t, str)

	monkey.UnpatchAll()

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
	assert.Contains(t, str, FormatTime(r.StartTime))
	assert.Contains(t, str, r.RoundTripTime.String())
	assert.Contains(t, str, FormatTime(r.LatestDownTime))
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
	context := SlackTimeFormation(r.StartTime, " probed at ", global.GetTimeFormat())
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

func TestResultToShell(t *testing.T) {
	r := newDummyResult("dummy")
	str := ToShell(r)
	assert.NotEmpty(t, str)
	assert.Contains(t, str, r.Title())
	assert.Contains(t, str, r.Status.String())
	assert.Contains(t, str, r.Message)

	var envMap map[string]string
	err := json.Unmarshal([]byte(str), &envMap)
	assert.Nil(t, err)

	assert.Equal(t, "Status", envMap["EASEPROBE_TYPE"])
	assert.Equal(t, "dummy", envMap["EASEPROBE_NAME"])
	assert.Equal(t, r.Status.String(), envMap["EASEPROBE_STATUS"])
	assert.Equal(t, fmt.Sprintf("%d", r.StartTimestamp), envMap["EASEPROBE_TIMESTAMP"])
	assert.Equal(t, FormatTime(r.StartTime), envMap["EASEPROBE_TIME"])
	assert.Equal(t, fmt.Sprintf("%d", r.RoundTripTime.Round(time.Millisecond)), envMap["EASEPROBE_RTT"])
	assert.Equal(t, r.Message, envMap["EASEPROBE_MESSAGE"])

	assert.Equal(t, ToCSV(r), envMap["EASEPROBE_CSV"])
	assert.Equal(t, ToJSON(r), envMap["EASEPROBE_JSON"])

	csvReader := csv.NewReader(strings.NewReader(ToCSV(r)))
	data, err := csvReader.ReadAll()
	assert.Nil(t, err)
	assert.Equal(t, len(data), 2)
	assert.Equal(t, data[1][0], r.Title())
	assert.Equal(t, data[1][1], r.Name)
	assert.Equal(t, data[1][2], r.Endpoint)
	assert.Equal(t, data[1][3], r.Status.String())
	assert.Equal(t, data[1][4], r.PreStatus.String())
	assert.Equal(t, data[1][5], fmt.Sprintf("%d", r.RoundTripTime.Round(time.Millisecond)))
	assert.Equal(t, data[1][6], FormatTime(r.StartTime))
	assert.Equal(t, data[1][7], fmt.Sprintf("%d", r.StartTimestamp))
	assert.Equal(t, data[1][8], r.Message)

	var w *csv.Writer
	monkey.PatchInstanceMethod(reflect.TypeOf(w), "WriteAll", func(_ *csv.Writer, _ [][]string) error {
		return fmt.Errorf("error")
	})
	assert.Empty(t, ToCSV(r))

	monkey.Patch(json.Marshal, func(v any) ([]byte, error) {
		return nil, fmt.Errorf("error")
	})
	assert.Empty(t, ToShell(r))

	monkey.UnpatchAll()

}
