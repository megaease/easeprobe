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

	log "github.com/sirupsen/logrus"
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
func (s *Status) Status(status string) {
	switch strings.ToLower(status) {
	case "up":
		*s = StatusUp
	case "down":
		*s = StatusDown
	case "unknown":
		*s = StatusUnknown
	case "init":
		*s = StatusInit
	}
	*s = StatusUnknown
}

// Emoji convert the status to emoji
func (s *Status) Emoji() string {
	switch *s {
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
	s.Status(strings.ToLower(string(b)))
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
	Name             string         `json:"name"`
	Endpoint         string         `json:"endpoint"`
	StartTime        time.Time      `json:"time"`
	StartTimestamp   int64          `json:"timestamp"`
	RoundTripTime    ConfigDuration `json:"rtt"`
	Status           Status         `json:"status"`
	PreStatus        Status         `json:"prestatus"`
	Message          string         `json:"message"`
	LatestDownTime   time.Time      `json:"latestdowntime"`
	RecoveryDuration time.Duration  `json:"recoverytime"`
	Stat             Stat           `json:"stat"`

	TimeFormat string `json:"-"`
}

// NewResult return a Result object
func NewResult() *Result {
	return &Result{
		Name:             "",
		Endpoint:         "",
		StartTime:        time.Now(),
		StartTimestamp:   0,
		RoundTripTime:    ConfigDuration{0},
		Status:           StatusInit,
		PreStatus:        StatusInit,
		Message:          "",
		LatestDownTime:   time.Time{},
		RecoveryDuration: 0,
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

// Title return the title for notification
func (r *Result) Title() string {
	t := "%s"
	if r.PreStatus == StatusInit {
		t = "Monitoring %s"
	}
	if r.Status != StatusUp {
		t = "%s Failure"
	} else {
		t = "%s Recovery - ( " + r.RecoveryDuration.Round(time.Second).String() + " Downtime )"
	}
	return fmt.Sprintf(t, r.Name)
}

// DebugJSON convert the object to DebugJSON
func (r *Result) DebugJSON() string {
	j, err := json.Marshal(&r)
	if err != nil {
		log.Errorf("error: %v", err)
		return ""
	}
	return string(j)
}

// DebugJSONIndent convert the object to indent JSON
func (r *Result) DebugJSONIndent() string {
	j, err := json.MarshalIndent(&r, "", "    ")
	if err != nil {
		log.Errorf("error: %v", err)
		return ""
	}
	return string(j)
}
