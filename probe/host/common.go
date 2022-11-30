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

package host

import (
	"strconv"
	"strings"
)

// IMetrics is the interface of metrics
type IMetrics interface {
	Name() string                   // Name is the name of the metrics
	Command() string                // Command returns the command to get the metrics
	OutputLines() int               // OutputLines returns the lines of command output
	Config(s *Server)               // Config returns the config of the metrics
	SetThreshold(t *Threshold)      // SetThreshold sets the threshold of the metrics
	Parse(s []string) error         // Parse a string to a metrics struct
	UsageInfo() string              // UsageInfo returns the usage info of the metrics
	CheckThreshold() (bool, string) // CheckThreshold check the metrics usage
	CreateMetrics(kind, tag string) // CreateMetrics creates the metrics
	ExportMetrics(name string)      // ExportMetrics export the metrics
}

// ResourceUsage is the resource usage for cpu and memory
type ResourceUsage struct {
	Used  int     `yaml:"used"`
	Total int     `yaml:"total"`
	Usage float64 `yaml:"usage"`
	Tag   string  `yaml:"tag"`
}

func first(str string) string {
	return strings.Split(strings.TrimSpace(str), " ")[0]
}

func strFloat(str string) float64 {
	n, _ := strconv.ParseFloat(strings.TrimSpace(str), 32)
	return n
}

func strInt(str string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(str), 10, 32)
	return n
}

func addMessage(msg string, message string) string {
	if msg == "" || message == "" {
		return message
	}
	return msg + " | " + message
}
