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
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func CreateTestResult() *Result {
	now := time.Now().UTC().Round(time.Millisecond)
	m := map[Status]int64{}
	m[StatusUp] = 50
	m[StatusDown] = 10

	r := &Result{
		Name:             "Test Name",
		Endpoint:         "http://example.com",
		StartTime:        now,
		StartTimestamp:   now.Unix(),
		RoundTripTime:    30 * time.Second,
		Status:           StatusUp,
		PreStatus:        StatusDown,
		Message:          "This is a test message",
		LatestDownTime:   now.Add(-20 * time.Hour),
		RecoveryDuration: 5 * time.Minute,
		Stat: Stat{
			Since:         now,
			Total:         1000,
			Status:        m,
			UpTime:        50 * time.Second,
			DownTime:      10 * time.Second,
			StatusCounter: *NewStatusCounter(2),
		},
	}
	return r
}

func TestResult(t *testing.T) {
	name := "Test Name"
	NewResultWithName(name)
	r := GetResultData(name)
	assert.Equal(t, name, r.Name)
	r1 := NewResultWithName(name)
	assert.Equal(t, r, r1)
	d := r.Clone()
	assert.Equal(t, name, d.Name)
}

func TestStatClone(t *testing.T) {
	s := Stat{
		Since: time.Now().UTC().Round(time.Millisecond),
		Total: 40,
		Status: map[Status]int64{
			StatusUp:   10,
			StatusDown: 30,
		},
		UpTime:   50 * time.Second,
		DownTime: 10 * time.Second,
	}

	d := s.Clone()
	if !reflect.DeepEqual(s, d) {
		t.Errorf("%v != %v", s, d)
	}
}

func TestResultMarshalJSON(t *testing.T) {

	r1 := CreateTestResult()
	r1.DoStat(30 * time.Minute)
	r1.PreStatus = StatusUp
	r1.Status = StatusDown
	r1.DoStat(60 * time.Minute)
	b, err := json.Marshal(r1)
	if err != nil {
		t.Error(err)
	}

	t.Log(string(b))

	r2 := &Result{}
	if err := json.Unmarshal(b, r2); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(r1, r2) {
		t.Fatalf("%v != %v", r1, r2)
	}
}

func TestResultMarshalYAML(t *testing.T) {

	r1 := CreateTestResult()
	b, err := yaml.Marshal(r1)
	if err != nil {
		t.Error(err)
	}

	t.Log(string(b))

	r2 := &Result{}
	if err := yaml.Unmarshal(b, r2); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(r1, r2) {
		t.Fatalf("%v != %v", r1, r2)
	}
}

func TestTitle(t *testing.T) {
	r := NewResult()
	r.Name = "Test Name"

	r.PreStatus = StatusInit
	r.Status = StatusUp
	expected := "Monitoring Test Name"
	if r.Title() != expected {
		t.Errorf("%s != %s", r.Title(), expected)
	}

	r.Status = StatusDown
	expected = "Test Name Failure"
	if r.Title() != expected {
		t.Errorf("%s != %s", r.Title(), expected)
	}

	r.PreStatus = StatusDown
	r.Status = StatusUp
	r.RecoveryDuration = 5 * time.Minute
	expected = "Test Name Recovery - ( 5m0s Downtime )"
	if r.Title() != expected {
		t.Errorf("%s != %s", r.Title(), expected)
	}
}

func TestDebug(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2022-01-01T00:00:00Z")
	r := NewResult()
	r.Name = "Test Name"
	r.Endpoint = "http://example.com"
	r.StartTime = now
	r.StartTimestamp = now.Unix()
	r.RoundTripTime = 30 * time.Second
	r.Status = StatusUp
	r.PreStatus = StatusDown
	r.Message = "This is a test message"
	r.LatestDownTime = now.Add(-20 * time.Hour)
	r.RecoveryDuration = 5 * time.Minute
	r.Stat.Since = now
	r.Stat.Total = 1000
	r.Stat.Status[StatusUp] = 50
	r.Stat.Status[StatusDown] = 10
	r.Stat.UpTime = 50 * time.Second
	r.Stat.DownTime = 10 * time.Second
	r.DoStat(30 * time.Minute)

	up := fmt.Sprintf("%d", StatusUp)
	down := fmt.Sprintf("%d", StatusDown)

	expected := `{"name":"Test Name","endpoint":"http://example.com","time":"2022-01-01T00:00:00Z","timestamp":1640995200,"rtt":30000000000,"status":"up","prestatus":"down","message":"This is a test message","latestdowntime":"2021-12-31T04:00:00Z","recoverytime":300000000000,"stat":{"since":"2022-01-01T00:00:00Z","total":1001,"status":{"1":51,"2":10},"uptime":1850000000000,"downtime":10000000000,"StatusHistory":[],"MaxLen":1,"CurrentStatus":true,"StatusCount":0,"alert":{"strategy":"regular","factor":1,"max":1,"notified":0,"failed":0,"next":1,"interval":0}}}`
	if r.DebugJSON() != expected {
		t.Errorf("%s != %s", r.DebugJSON(), expected)
	}

	expected = `{
    "name": "Test Name",
    "endpoint": "http://example.com",
    "time": "2022-01-01T00:00:00Z",
    "timestamp": 1640995200,
    "rtt": 30000000000,
    "status": "up",
    "prestatus": "down",
    "message": "This is a test message",
    "latestdowntime": "2021-12-31T04:00:00Z",
    "recoverytime": 300000000000,
    "stat": {
        "since": "2022-01-01T00:00:00Z",
        "total": 1001,
        "status": {
            "` + up + `": 51,
            "` + down + `": 10
        },
        "uptime": 1850000000000,
        "downtime": 10000000000,
        "StatusHistory": [],
        "MaxLen": 1,
        "CurrentStatus": true,
        "StatusCount": 0,
        "alert": {
            "strategy": "regular",
            "factor": 1,
            "max": 1,
            "notified": 0,
            "failed": 0,
            "next": 1,
            "interval": 0
        }
    }
}`

	str := r.DebugJSONIndent()
	if str != expected {
		t.Errorf("%s != %s", str, expected)
	}

	monkey.Patch(json.Marshal, func(v interface{}) ([]byte, error) {
		return nil, fmt.Errorf("marshal error")
	})
	monkey.Patch(json.MarshalIndent, func(v any, prefix string, indent string) ([]byte, error) {
		return nil, fmt.Errorf("marshal error")
	})

	str = r.DebugJSON()
	assert.Empty(t, str)
	str = r.DebugJSONIndent()
	assert.Empty(t, str)

	monkey.UnpatchAll()
}

func TestSLAPercent(t *testing.T) {
	r := NewResult()
	r.Stat.UpTime = 10 * time.Second
	r.Stat.DownTime = 10 * time.Second
	assert.Equal(t, float64(50), r.SLAPercent())

	r.Stat.UpTime = 0
	r.Stat.DownTime = 10 * time.Second
	assert.Equal(t, float64(0), r.SLAPercent())

	r.Stat.UpTime = 10 * time.Second
	r.Stat.DownTime = 0
	assert.Equal(t, float64(100), r.SLAPercent())

	r.Stat.UpTime = 0
	r.Stat.DownTime = 0
	r.Status = StatusDown
	assert.Equal(t, float64(0), r.SLAPercent())
	r.Status = StatusUp
	assert.Equal(t, float64(100), r.SLAPercent())
}
