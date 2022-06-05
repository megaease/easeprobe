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

package global

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// EaseProbe is the information of the program
type EaseProbe struct {
	Name    string `yaml:"name"`
	IconURL string `yaml:"icon"`
	Version string `yaml:"version"`
	Host    string `yaml:"host"`
}

var easeProbe *EaseProbe

// InitEaseProbe the EaseProbe
func InitEaseProbe(name, icon string) {
	host, err := os.Hostname()
	if err != nil {
		log.Errorf("Get Hostname Failed: %s", err)
		host = "unknown"
	}
	easeProbe = &EaseProbe{
		Name:    name,
		IconURL: icon,
		Version: Ver,
		Host:    host,
	}
}

// GetEaseProbe return the EaseProbe
func GetEaseProbe() *EaseProbe {
	if easeProbe == nil {
		InitEaseProbe(DefaultProg, DefaultIconURL)
	}
	return easeProbe
}

// FooterString return the footer string
// e.g. "EaseProbe v1.0.0 @ localhost"
func FooterString() string {
	return easeProbe.Name + " " + easeProbe.Version + " @ " + easeProbe.Host
}
