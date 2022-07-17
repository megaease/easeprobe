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
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestDirverType(t *testing.T) {
	assert.Equal(t, "mysql", MySQL.String())
	assert.Equal(t, "redis", Redis.String())
	assert.Equal(t, "memcache", Memcache.String())
	assert.Equal(t, "mongo", Mongo.String())

	d := Unknown
	assert.Equal(t, MySQL, d.DriverType("mysql"))
	assert.Equal(t, Redis, d.DriverType("redis"))
	assert.Equal(t, Memcache, d.DriverType("memcache"))

	d = d.DriverType("postgres")
	buf, err := yaml.Marshal(d)
	assert.Nil(t, err)
	assert.Equal(t, "postgres\n", string(buf))

	err = yaml.Unmarshal([]byte("zookeeper"), &d)
	assert.Nil(t, err)
	assert.Equal(t, Zookeeper, d)

	err = yaml.Unmarshal([]byte("xxx"), &d)
	assert.Nil(t, err)
	assert.Equal(t, Unknown, d)

	d = MySQL
	buf, err = json.Marshal(d)
	assert.Nil(t, err)
	assert.Equal(t, "\"mysql\"", string(buf))

	err = json.Unmarshal([]byte("\"mongo\""), &d)
	assert.Nil(t, err)
	assert.Equal(t, Mongo, d)

	d = Memcache
	buf, err = json.Marshal(d)
	assert.Nil(t, err)
	assert.Equal(t, "\"memcache\"", string(buf))

	d = 10
	assert.Equal(t, "unknown", d.String())

}

func TestOptionsCheck(t *testing.T) {
	opts := Options{
		Host:       "localhost:3306",
		DriverType: MySQL,
	}
	err := opts.Check()
	assert.Nil(t, err)

	opts.Host = "127.0.0.1:3306"
	err = opts.Check()
	assert.Nil(t, err)

	opts.Host = "localhost:3306"
	opts.DriverType = Unknown
	err = opts.Check()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unknown driver")

	opts.Host = "localhost"
	err = opts.Check()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Host")

	opts.Host = "localhost:3306:1234"
	err = opts.Check()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Host")

	opts.Host = "10.10.10.1:asdf"
	err = opts.Check()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Port")

	opts.Host = "10.10.10.1:123456"
	err = opts.Check()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Port")
}
