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
	"time"

	"github.com/megaease/easeprobe/global"
	log "github.com/sirupsen/logrus"
)

// Stat is the statistics of probe result
type Stat struct {
	Since    time.Time        `json:"since" yaml:"since"`
	Total    int64            `json:"total" yaml:"total"`
	Status   map[Status]int64 `json:"status" yaml:"status"`
	UpTime   time.Duration    `json:"uptime" yaml:"uptime"`
	DownTime time.Duration    `json:"downtime" yaml:"downtime"`
	StatusCounter
	NotificationStrategyData `json:"alert" yaml:"alert"`
}

// Result is the status of health check
type Result struct {
	Name             string        `json:"name" yaml:"name"`
	Endpoint         string        `json:"endpoint" yaml:"endpoint"`
	StartTime        time.Time     `json:"time" yaml:"time"`
	StartTimestamp   int64         `json:"timestamp" yaml:"timestamp"`
	RoundTripTime    time.Duration `json:"rtt" yaml:"rtt"`
	Status           Status        `json:"status" yaml:"status"`
	PreStatus        Status        `json:"prestatus" yaml:"prestatus"`
	Message          string        `json:"message" yaml:"message"`
	LatestDownTime   time.Time     `json:"latestdowntime" yaml:"latestdowntime"`
	RecoveryDuration time.Duration `json:"recoverytime" yaml:"recoverytime"`
	Stat             Stat          `json:"stat" yaml:"stat"`
}

// NewResult return a Result object
func NewResult() *Result {
	return &Result{
		Name:             "",
		Endpoint:         "",
		StartTime:        time.Now().UTC(),
		StartTimestamp:   0,
		RoundTripTime:    0,
		Status:           StatusInit,
		PreStatus:        StatusInit,
		Message:          "",
		LatestDownTime:   time.Time{},
		RecoveryDuration: 0,
		Stat: Stat{
			Since: time.Now().UTC(),
			Total: 0,
			Status: map[Status]int64{
				StatusUp:   0,
				StatusDown: 0,
			},
			UpTime:        0,
			DownTime:      0,
			StatusCounter: *NewStatusCounter(global.DefaultStatusChangeThresholdSetting),
			NotificationStrategyData: *NewNotificationStrategyData(
				global.DefaultNotificationStrategy,
				global.DefaultMaxNotificationTimes,
				global.DefaultNotificationFactor,
			),
		},
	}
}

// NewResultWithName return a Result object with name
func NewResultWithName(name string) *Result {
	r := GetResultData(name)
	if r != nil {
		log.Infof("Found result from data file with name: %s", name)
		return r
	}
	log.Infof("Create result with name: %s", name)
	r = NewResult()
	r.Name = name
	SetResultData(name, r)
	return r
}

// Clone return a clone of the Result
func (r *Result) Clone() Result {
	dst := Result{}
	dst.Name = r.Name
	dst.Endpoint = r.Endpoint
	dst.StartTime = r.StartTime
	dst.StartTimestamp = r.StartTimestamp
	dst.RoundTripTime = r.RoundTripTime
	dst.Status = r.Status
	dst.PreStatus = r.PreStatus
	dst.Message = r.Message
	dst.LatestDownTime = r.LatestDownTime
	dst.RecoveryDuration = r.RecoveryDuration
	dst.Stat = r.Stat.Clone()
	return dst
}

// Clone return a clone of the Stat
func (s *Stat) Clone() Stat {
	dst := Stat{}
	dst.Since = s.Since
	dst.Total = s.Total
	dst.Status = make(map[Status]int64)
	for k, v := range s.Status {
		dst.Status[k] = v
	}
	dst.UpTime = s.UpTime
	dst.DownTime = s.DownTime
	dst.StatusCounter = s.StatusCounter.Clone()
	dst.NotificationStrategyData = s.NotificationStrategyData.Clone()
	return dst
}

// DoStat is the function do the statistics
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
	t := ""
	if r.PreStatus == StatusInit && r.Status == StatusUp {
		t = "Monitoring %s"
	} else if r.Status != StatusUp {
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

// SLAPercent calculate the SLAPercent
func (r *Result) SLAPercent() float64 {
	uptime := r.Stat.UpTime.Seconds()
	downtime := r.Stat.DownTime.Seconds()
	if uptime+downtime <= 0 {
		if r.Status == StatusUp {
			return 100
		}
		return 0
	}
	return uptime / (uptime + downtime) * 100
}
