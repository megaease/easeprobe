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

// Package conf is the configuration package for native client
package conf

import (
	"fmt"
	"net"
	"strconv"

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
	Memcache
	Kafka
	Mongo
	PostgreSQL
	Zookeeper
)

// DriverMap is the map of [driver, name]
var DriverMap = map[DriverType]string{
	MySQL:      "mysql",
	Redis:      "redis",
	Memcache:   "memcache",
	Kafka:      "kafka",
	Mongo:      "mongo",
	PostgreSQL: "postgres",
	Zookeeper:  "zookeeper",
	Unknown:    "unknown",
}

// Options implements the configuration for native client
type Options struct {
	base.DefaultProbe `yaml:",inline"`

	Host       string            `yaml:"host" json:"host" jsonschema:"required,format=hostname,title=Host,description=The host of the client,example=10.1.1.1:9000"`
	DriverType DriverType        `yaml:"driver" json:"driver" jsonschema:"required,type=string,enum=mysql,enum=redis,enum=memcache,enum=kafka,enum=mongo,enum=postgres,enum=zookeeper,title=Driver,description=The driver of the client,example=mysql"`
	Username   string            `yaml:"username,omitempty" json:"username,omitempty" jsonschema:"title=Username,description=The username of the client,example=root"`
	Password   string            `yaml:"password,omitempty" json:"password,omitempty" jsonschema:"title=Password,description=The password of the client,example=123456"`
	Data       map[string]string `yaml:"data,omitempty" json:"data,omitempty" jsonschema:"title=Data,description=The data of the client,example={\"key\":\"value\"}"`

	//TLS
	global.TLS `yaml:",inline"`
}

// Check do the configuration check
func (d *Options) Check() error {
	_, port, err := net.SplitHostPort(d.Host)
	if err != nil {
		return fmt.Errorf("Invalid Host: %s. %v", d.Host, err)
	}
	if p, err := strconv.Atoi(port); err != nil || p < 1 || p > 65535 {
		return fmt.Errorf("Invalid Port: %s", port)
	}
	if d.DriverType == Unknown {
		return fmt.Errorf("Unknown driver")
	}
	return nil
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
		*d = val
		return val
	}
	return Unknown
}

// MarshalYAML is marshal the driver type
func (d DriverType) MarshalYAML() (interface{}, error) {
	return global.EnumMarshalYaml(DriverMap, d, "Client Driver")
}

// UnmarshalYAML is unmarshal the driver type
func (d *DriverType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return global.EnumUnmarshalYaml(unmarshal, DriverTypeMap, d, Unknown, "Client Driver")
}

// MarshalJSON is marshal the driver
func (d DriverType) MarshalJSON() (b []byte, err error) {
	return global.EnumMarshalJSON(DriverMap, d, "Client Driver")
}

// UnmarshalJSON is Unmarshal the driver type
func (d *DriverType) UnmarshalJSON(b []byte) (err error) {
	return global.EnumUnmarshalJSON(b, DriverTypeMap, d, Unknown, "Client Driver")
}
