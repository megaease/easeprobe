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

package conf

import (
	"fmt"
	"strings"

	"github.com/megaease/easeprobe/notify/base"
)

// Provider Interface
type Provider interface {
	Kind() string
	Notify(string, string) error
}

// ProviderType is the sms provider
type ProviderType int

// The Provider Type of different sms provider
const (
	Unknown ProviderType = iota
	Yunpian
	Twilio
	Nexmo
)

// ProviderMap is the map of [provider, name]
var ProviderMap = map[ProviderType]string{
	Yunpian: "yunpian",
	Twilio:  "twilio",
	Nexmo:   "nexmo",
	Unknown: "unknown",
}

// Options implements the configuration for native client
type Options struct {
	base.DefaultNotify `yaml:",inline"`

	ProviderType ProviderType `yaml:"provider"`
	Mobile       string       `yaml:"mobile"`
}

// ProviderTypeMap is the map of provider [name, provider]
var ProviderTypeMap = reverseMap(ProviderMap)

func reverseMap(m map[ProviderType]string) map[string]ProviderType {
	n := make(map[string]ProviderType, len(m))
	for k, v := range m {
		n[v] = k
	}
	return n
}

// String convert the DriverType to string
func (d ProviderType) String() string {
	if val, ok := ProviderMap[d]; ok {
		return val
	}
	return ProviderMap[Unknown]
}

// ProviderType convert the string to ProviderType
func (d *ProviderType) ProviderType(name string) ProviderType {
	if val, ok := ProviderTypeMap[name]; ok {
		return val
	}
	return Unknown
}

// MarshalYAML is marshal the provider type
func (d *ProviderType) MarshalYAML() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalYAML is unmarshal the provider type
func (d *ProviderType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	*d = d.ProviderType(strings.ToLower(s))
	return nil
}

// UnmarshalJSON is Unmarshal the provider type
func (d *ProviderType) UnmarshalJSON(b []byte) (err error) {
	*d = d.ProviderType(strings.ToLower(string(b)))
	return nil
}

// MarshalJSON is marshal the provider
func (d *ProviderType) MarshalJSON() (b []byte, err error) {
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}
