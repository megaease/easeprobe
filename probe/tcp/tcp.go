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

// Package tcp is the tcp probe package
package tcp

import (
	"fmt"
	"net"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/base"
	log "github.com/sirupsen/logrus"
)

// TCP implements a config for TCP
type TCP struct {
	base.DefaultProbe `yaml:",inline"`
	Host              string `yaml:"host" json:"host" jsonschema:"required,format=hostname,title=Host,description=The host to probe"`
	Proxy             string `yaml:"proxy" json:"proxy,omitempty" jsonschema:"format=hostname,title=Proxy,description=The proxy to use"`
}

// Config HTTP Config Object
func (t *TCP) Config(gConf global.ProbeSettings) error {
	kind := "tcp"
	tag := ""
	name := t.ProbeName
	t.DefaultProbe.Config(gConf, kind, tag, name, t.Host, t.DoProbe)

	log.Debugf("[%s / %s] configuration: %+v", t.ProbeKind, t.ProbeName, *t)
	return nil
}

// DoProbe return the checking result
func (t *TCP) DoProbe() (bool, string) {
	conn, err := t.GetProxyConnection(t.Proxy, t.Host)
	status := true
	message := ""
	if err != nil {
		message = fmt.Sprintf("Error: %v", err)
		log.Errorf("[%s / %s] error: %v", t.ProbeKind, t.ProbeName, err)
		status = false
	} else {
		message = "TCP Connection Established Successfully!"
		if tcpCon, ok := conn.(*net.TCPConn); ok {
			tcpCon.SetLinger(0)
		}
		defer conn.Close()
	}
	return status, message
}
