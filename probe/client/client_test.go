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
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/megaease/easeprobe/probe/client/conf"
	"github.com/megaease/easeprobe/probe/client/kafka"
	"github.com/megaease/easeprobe/probe/client/memcache"
	"github.com/megaease/easeprobe/probe/client/mongo"
	"github.com/megaease/easeprobe/probe/client/mysql"
	"github.com/megaease/easeprobe/probe/client/postgres"
	"github.com/megaease/easeprobe/probe/client/redis"
	"github.com/megaease/easeprobe/probe/client/zookeeper"
	"github.com/stretchr/testify/assert"
)

func newDummyClient(driver conf.DriverType) Client {
	return Client{
		Options: conf.Options{
			DefaultProbe: base.DefaultProbe{
				ProbeName: "dummy_" + driver.String(),
			},
			Host:       "example.com:1234",
			DriverType: driver,
			Username:   "user",
			Password:   "pass",
			Data:       map[string]string{},
			TLS:        global.TLS{},
		},
		client: nil,
	}
}

func MockProbe[T any](c T) {
	p := &c
	monkey.PatchInstanceMethod(reflect.TypeOf(p), "Probe", func(_ *T) (bool, string) {
		return true, "Successfully"
	})
}

func TestClient(t *testing.T) {
	clients := []Client{
		newDummyClient(conf.MySQL),
		newDummyClient(conf.PostgreSQL),
		newDummyClient(conf.Redis),
		newDummyClient(conf.Mongo),
		newDummyClient(conf.Kafka),
		newDummyClient(conf.Zookeeper),
		newDummyClient(conf.Memcache),
	}

	for _, client := range clients {
		err := client.Config(global.ProbeSettings{})
		assert.Nil(t, err)
		assert.Equal(t, "client", client.ProbeKind)
		assert.Equal(t, client.DriverType.String(), client.ProbeTag)

		client.Host = "wronghost"
		err = client.Config(global.ProbeSettings{})
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Invalid Host")

		switch client.DriverType {
		case conf.MySQL:
			MockProbe(mysql.MySQL{})
		case conf.PostgreSQL:
			MockProbe(postgres.PostgreSQL{})
		case conf.Redis:
			MockProbe(redis.Redis{})
		case conf.Mongo:
			MockProbe(mongo.Mongo{})
		case conf.Kafka:
			MockProbe(kafka.Kafka{})
		case conf.Zookeeper:
			MockProbe(zookeeper.Zookeeper{})
		case conf.Memcache:
			MockProbe(memcache.Memcache{})
		}
		client.Host = "example.com:1234"
		err = client.Config(global.ProbeSettings{})
		assert.Nil(t, err)

		s, m := client.DoProbe()
		assert.True(t, s)
		assert.Contains(t, m, "Successfully")
	}

	u := newDummyClient(conf.Unknown)
	u.Config(global.ProbeSettings{})
	s, m := u.DoProbe()
	assert.False(t, s)
	assert.Contains(t, m, "Wrong Driver Type")

	monkey.UnpatchAll()
}

func TestFailed(t *testing.T) {

	c := newDummyClient(conf.Unknown)
	var cnf *conf.Options
	monkey.PatchInstanceMethod(reflect.TypeOf(cnf), "Check", func(_ *conf.Options) error {
		return nil
	})
	err := c.Config(global.ProbeSettings{})
	assert.NotNil(t, err)

}
