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

// Package client is the native client probe package
package client

import (
	"fmt"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/client/conf"
	"github.com/megaease/easeprobe/probe/client/kafka"
	"github.com/megaease/easeprobe/probe/client/memcache"
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

	client conf.Driver `yaml:"-" json:"-"`
}

// Config Client Config Object
func (c *Client) Config(gConf global.ProbeSettings) error {
	kind := "client"
	tag := c.DriverType.String()
	name := c.ProbeName
	c.DefaultProbe.Config(gConf, kind, tag, name, c.Host, c.DoProbe)
	if err := c.Check(); err != nil {
		return err
	}
	if err := c.configClientDriver(); err != nil {
		return err
	}
	log.Debugf("[%s / %s / %s ] configuration: %+v", c.ProbeKind, c.ProbeTag, c.ProbeName, *c)
	return nil
}

func (c *Client) configClientDriver() (err error) {
	switch c.DriverType {
	case conf.MySQL:
		c.client, err = mysql.New(c.Options)
	case conf.Redis:
		c.client, err = redis.New(c.Options)
	case conf.Memcache:
		c.client, err = memcache.New(c.Options)
	case conf.Mongo:
		c.client, err = mongo.New(c.Options)
	case conf.Kafka:
		c.client, err = kafka.New(c.Options)
	case conf.PostgreSQL:
		c.client, err = postgres.New(c.Options)
	case conf.Zookeeper:
		c.client, err = zookeeper.New(c.Options)
	default:
		c.DriverType = conf.Unknown
		err = fmt.Errorf("Unknown Driver Type")
	}
	return
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
