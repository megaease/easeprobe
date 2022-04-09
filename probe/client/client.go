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

package client

import (
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/client/conf"
	"github.com/megaease/easeprobe/probe/client/kafka"
	"github.com/megaease/easeprobe/probe/client/mongo"
	"github.com/megaease/easeprobe/probe/client/mysql"
	"github.com/megaease/easeprobe/probe/client/postgres"
	"github.com/megaease/easeprobe/probe/client/redis"
	"github.com/megaease/easeprobe/probe/client/zookeeper"
	log "github.com/sirupsen/logrus"
)

// Client implements the structure of client
type Client struct {
	//Embed structure
	conf.Options `yaml:",inline"`

	client conf.Driver `yaml:"-"`
}

// Config Client Config Object
func (c *Client) Config(gConf global.ProbeSettings) error {
	kind := "client"
	tag := c.DriverType.String()
	name := c.ProbeName
	c.DefaultOptions.Config(gConf, kind, tag, name, c.Host, c.DoProbe)
	c.configClientDriver()

	log.Debugf("[%s] configuration: %+v, %+v", c.ProbeKind, c, c.Result())
	return nil
}

func (c *Client) configClientDriver() {
	switch c.DriverType {
	case conf.MySQL:
		c.client = mysql.New(c.Options)
	case conf.Redis:
		c.client = redis.New(c.Options)
	case conf.Mongo:
		c.client = mongo.New(c.Options)
	case conf.Kafka:
		c.client = kafka.New(c.Options)
	case conf.PostgreSQL:
		c.client = postgres.New(c.Options)
	case conf.Zookeeper:
		c.client = zookeeper.New(c.Options)
	default:
		c.DriverType = conf.Unknown
	}

}

// DoProbe return the checking result
func (c *Client) DoProbe() (bool, string) {
	if c.DriverType == conf.Unknown {
		c.ProbeResult.PreStatus = probe.StatusUnknown
		c.ProbeResult.Status = probe.StatusUnknown
		return false, "Wrong Driver Type"
	}
	return c.client.Probe()
}
