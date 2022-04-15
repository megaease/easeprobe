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
	"strings"

	"github.com/megaease/easeprobe/probe"
)

// Format is the format of text
type Format int

// The format types
const (
	Unknown        Format = iota
	MarkdownSocial        // *text* is bold
	Markdown              // **text** is bold
	HTML
	JSON
	Text
	Slack
	Discord
	Lark
)

// String covert the Format to string
func (f Format) String() string {
	switch f {
	case MarkdownSocial:
		return "markdown-social"
	case Markdown:
		return "markdown"
	case HTML:
		return "html"
	case JSON:
		return "json"
	case Slack:
		return "slack"
	case Discord:
		return "discord"
	case Lark:
		return "lark"
	default:
		return "unknown"
	}
}

// Format covert the string to Format
func (f *Format) Format(s string) {
	switch strings.ToLower(s) {
	case "markdown":
		*f = Markdown
	case "markdown-social":
		*f = MarkdownSocial
	case "html":
		*f = HTML
	case "json":
		*f = JSON
	case "slack":
		*f = Slack
	case "discrod":
		*f = Discord
	case "lark":
		*f = Lark
	default:
		*f = Unknown
	}
}

// MarshalYAML is marshal the format
func (f *Format) MarshalYAML() ([]byte, error) {
	return []byte(f.String()), nil
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

// FormatFuncType is the format function
type FormatFuncType func(probe.Result) string

// StatFormatFuncType is the format function for Stat
type StatFormatFuncType func([]probe.Prober) string

// FormatFuncStruct is the format function struct
type FormatFuncStruct struct {
	ResultFn FormatFuncType
	StatFn   StatFormatFuncType
}

// FormatFuncs is the format function map
var FormatFuncs = map[Format]FormatFuncStruct{
	Unknown:        {ToText, SLAText},
	Text:           {ToText, SLAText},
	JSON:           {ToJSON, SLAJSON},
	Markdown:       {ToMarkdown, SLAMarkdown},
	MarkdownSocial: {ToMarkdownSocial, SLAMarkdownSocial},
	HTML:           {ToHTML, SLAHTML},
	Slack:          {ToSlack, SLASlack},
	Lark:           {ToLark, SLALark},
}
