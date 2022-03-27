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

package zookeeper

import (
	"crypto/tls"
	"github.com/go-zookeeper/zk"
	"github.com/megaease/easeprobe/probe/client/conf"
	log "github.com/sirupsen/logrus"
	"time"
)

// Kind is the type of driver
const Kind string = "Zookeeper"

// Zookeeper is the Zookeeper client
type Zookeeper struct {
	conf.Options `yaml:",inline"`
	tls          *tls.Config `yaml:"-"`
	ConnStr      string      `yaml:"conn_str"`
}

// New create a Redis client
func New(opt conf.Options) Zookeeper {
	var conn string

	tls, err := opt.TLS.Config()
	if err != nil {
		log.Errorf("[%s] %s - TLS Config error - %v", Kind, opt.Name, err)
	}

	return Zookeeper{
		Options: opt,
		tls:     tls,
		ConnStr: conn,
	}
}

// Kind return the name of client
func (r Zookeeper) Kind() string {
	return Kind
}

// Probe do the health check
func (r Zookeeper) Probe() (bool, string) {
	conn, _, err := zk.Connect([]string{r.Options.Host}, time.Second*5, zk.WithLogInfo(false))
	if err != nil {
		return false, err.Error()
	}
	defer conn.Close()

	return true, "Check Zookeeper Server Successfully!"
}
