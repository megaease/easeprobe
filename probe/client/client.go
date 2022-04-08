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
	"fmt"
	"time"

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
	c.ProbeKind = "client"
	c.DefaultOptions.Config(gConf, c.Name, c.Host)
	c.configClientDriver()

	log.Debugf("[%s] configuration: %+v, %+v", c.Kind(), c, c.Result())
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

// Probe return the checking result
func (c *Client) Probe() probe.Result {
	if c.DriverType == conf.Unknown {
		c.ProbeResult.PreStatus = probe.StatusUnknown
		c.ProbeResult.Status = probe.StatusUnknown
		return *c.ProbeResult
	}

	now := time.Now()
	c.ProbeResult.StartTime = now
	c.ProbeResult.StartTimestamp = now.UnixMilli()

	stat, msg := c.client.Probe()

	c.ProbeResult.RoundTripTime.Duration = time.Since(now)

	status := probe.StatusUp
	c.ProbeResult.Message = fmt.Sprintf("%s client checked up successfully!", c.DriverType.String())

	if stat != true {
		c.ProbeResult.Message = fmt.Sprintf("Error (%s): %s", c.DriverType.String(), msg)
		log.Errorf("[%s / %s / %s] - %s", c.Kind(), c.client.Kind(), c.Name, msg)
		status = probe.StatusDown
	} else {
		log.Debugf("[%s / %s / %s] - %s", c.Kind(), c.client.Kind(), c.Name, msg)
	}

	c.ProbeResult.PreStatus = c.ProbeResult.Status
	c.ProbeResult.Status = status

	c.ProbeResult.DoStat(c.Interval())
	return *c.ProbeResult
}
