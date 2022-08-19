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

	"github.com/megaease/easeprobe/global"
)

// Status is the status of Probe
type Status int

// The status of a probe
const (
	StatusInit Status = iota
	StatusUp
	StatusDown
	StatusUnknown
	StatusBad
)

var (
	toString = map[Status]string{
		StatusInit:    "init",
		StatusUp:      "up",
		StatusDown:    "down",
		StatusUnknown: "unknown",
		StatusBad:     "bad",
	}

	toStatus = global.ReverseMap(toString)

	toEmoji = map[Status]string{
		StatusInit:    "🔎",
		StatusUp:      "✅",
		StatusDown:    "❌",
		StatusUnknown: "⛔️",
		StatusBad:     "🚫",
	}
)

// String convert the Status to string
func (s Status) String() string {
	if val, ok := toString[s]; ok {
		return val
	}
	return "unknown"
}

// Status convert the string to Status
func (s *Status) Status(status string) {
	if val, ok := toStatus[strings.ToLower(status)]; ok {
		*s = val
	} else {
		*s = StatusUnknown
	}
}

// Emoji convert the status to emoji
func (s *Status) Emoji() string {
	if val, ok := toEmoji[*s]; ok {
		return val
	}
	return "⛔️"
}

// UnmarshalYAML is Unmarshal the status
func (s *Status) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var status string
	*s = StatusUnknown
	if err := unmarshal(&status); err != nil {
		return err
	}
	if val, ok := toStatus[strings.ToLower(status)]; ok {
		*s = val
		return nil
	}
	return fmt.Errorf("Unknown status: %s", status)
}

// MarshalYAML is Marshal the status
func (s Status) MarshalYAML() (interface{}, error) {
	if val, ok := toString[s]; ok {
		return val, nil
	}
	return "unknown", fmt.Errorf("Unknown status: %s", s)
}

// UnmarshalJSON is Unmarshal the status
func (s *Status) UnmarshalJSON(b []byte) (err error) {
	var str string
	*s = StatusUnknown
	if err = json.Unmarshal(b, &str); err != nil {
		return err
	}
	if val, ok := toStatus[strings.ToLower(str)]; ok {
		*s = val
		return nil
	}
	return fmt.Errorf("Unknown status: %s", str)
}

// MarshalJSON is marshal the status
func (s Status) MarshalJSON() (b []byte, err error) {
	if val, ok := toString[s]; ok {
		return []byte(fmt.Sprintf(`"%s"`, val)), nil
	}
	return []byte("unknown"), fmt.Errorf("Unknown status: %s", s)
}
