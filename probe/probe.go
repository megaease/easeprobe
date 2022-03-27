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
	"fmt"
	"strings"
	"time"

	"github.com/megaease/easeprobe/global"
)

// Prober Interface
type Prober interface {
	Kind() string
	Config(global.ProbeSettings) error
	Probe() Result
	Interval() time.Duration
	Result() *Result
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
	*s = s.Status(strings.ToLower(string(b)))
	return nil
}

// MarshalJSON is marshal the status
func (s *Status) MarshalJSON() (b []byte, err error) {
	return []byte(fmt.Sprintf(`"%s"`, s.String())), nil
}
