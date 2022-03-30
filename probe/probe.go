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
func (s *Status) String() string {
	switch *s {
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

// Format is the format of text
type Format int

// The format types
const (
	MakerdownSocial Format = iota // *text* is bold
	Makerdown                     // **text** is bold
	HTML
	JSON
	Text
)

// String covert the Format to string
func (f *Format) String() string {
	switch *f {
	case MakerdownSocial:
		return "markdown-social"
	case Makerdown:
		return "markdown"
	case HTML:
		return "html"
	case JSON:
		return "json"
	default:
		return "text"
	}
}

// Format covert the string to Format
func (f *Format) Format(s string) {
	switch strings.ToLower(s) {
	case "markdown":
		*f = Makerdown
	case "markdown-social":
		*f = MakerdownSocial
	case "html":
		*f = HTML
	case "json":
		*f = JSON
	default:
		*f = Text
	}
}

// UnmarshalYAML is unmarshal the format
func (f *Format) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var format string
	if err := unmarshal(&format); err != nil {
		return err
	}
	f.Format(format)
	return nil
}
