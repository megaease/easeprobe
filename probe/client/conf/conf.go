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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/base"
)

// Driver Interface
type Driver interface {
	Kind() string
	Probe() (bool, string)
}

// DriverType is the client driver
type DriverType int

// The Driver Type of native client
const (
	Unknown DriverType = iota
	MySQL
	Redis
	Kafka
	Mongo
	PostgreSQL
	Zookeeper
)

// DriverMap is the map of [driver, name]
var DriverMap = map[DriverType]string{
	MySQL:      "mysql",
	Redis:      "redis",
	Kafka:      "kafka",
	Mongo:      "mongo",
	PostgreSQL: "postgres",
	Zookeeper:  "zookeeper",
	Unknown:    "unknown",
}

// Options implements the configuration for native client
type Options struct {
	base.DefaultProbe `yaml:",inline"`

	Host       string     `yaml:"host"`
	DriverType DriverType `yaml:"driver"`
	Username   string     `yaml:"username"`
	Password   string     `yaml:"password"`

	//TLS
	global.TLS `yaml:",inline"`
}

// DriverTypeMap is the map of driver [name, driver]
var DriverTypeMap = reverseMap(DriverMap)

func reverseMap(m map[DriverType]string) map[string]DriverType {
	n := make(map[string]DriverType, len(m))
	for k, v := range m {
		n[v] = k
	}
	return n
}

// String convert the DriverType to string
func (d DriverType) String() string {
	if val, ok := DriverMap[d]; ok {
		return val
	}
	return DriverMap[Unknown]
}

// DriverType convert the string to DriverType
func (d *DriverType) DriverType(name string) DriverType {
	if val, ok := DriverTypeMap[name]; ok {
		return val
	}
	return Unknown
}

// MarshalYAML is marshal the driver type
func (d DriverType) MarshalYAML() (interface{}, error) {
	return d.String(), nil
}

// UnmarshalYAML is unmarshal the driver type
func (d *DriverType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	*d = d.DriverType(strings.ToLower(s))
	return nil
}

// UnmarshalJSON is Unmarshal the driver type
func (d *DriverType) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return err
	}
	*d = d.DriverType(strings.ToLower(s))
	return nil
}

// MarshalJSON is marshal the driver
func (d DriverType) MarshalJSON() (b []byte, err error) {
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}
