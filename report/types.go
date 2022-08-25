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

	"github.com/megaease/easeprobe/global"
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
	Log
	Slack
	Discord
	Lark
	SMS
	Shell
)

var fmtToStr = map[Format]string{
	Unknown:        "unknown",
	MarkdownSocial: "markdown-social",
	Markdown:       "markdown",
	HTML:           "html",
	JSON:           "json",
	Text:           "text",
	Log:            "log",
	Slack:          "slack",
	Discord:        "discord",
	Lark:           "lark",
	SMS:            "sms",
	Shell:          "shell",
}

var strToFmt = global.ReverseMap(fmtToStr)

// String covert the Format to string
func (f Format) String() string {
	return fmtToStr[f]
}

// Format covert the string to Format
func (f *Format) Format(s string) {
	*f = strToFmt[strings.ToLower(s)]
}

// MarshalYAML is marshal the format
func (f Format) MarshalYAML() (interface{}, error) {
	return global.EnumMarshalYaml(fmtToStr, f, "Format")
}

// UnmarshalYAML is unmarshal the format
func (f *Format) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return global.EnumUnmarshalYaml(unmarshal, strToFmt, f, Unknown, "Format")
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
	Log:            {ToLog, SLALog},
	JSON:           {ToJSON, SLAJSON},
	Markdown:       {ToMarkdown, SLAMarkdown},
	MarkdownSocial: {ToMarkdownSocial, SLAMarkdownSocial},
	HTML:           {ToHTML, SLAHTML},
	Slack:          {ToSlack, SLASlack},
	Lark:           {ToLark, SLALark},
	SMS:            {ToText, SLASummary},
	Shell:          {ToShell, SLAShell},
}
