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

func testMarshalUnmarshal(
	t *testing.T, str string, driver DriverType, good bool,
	marshal func(in interface{}) ([]byte, error),
	unmarshal func(in []byte, out interface{}) (err error)) {

	var d DriverType
	err := unmarshal([]byte(str), &d)
	if good {
		assert.Nil(t, err)
		assert.Equal(t, driver, d)
	} else {
		assert.Error(t, err)
		assert.Equal(t, Unknown, d)
	}

	buf, err := marshal(driver)
	if good {
		assert.Nil(t, err)
		assert.Equal(t, str, string(buf))
	} else {
		assert.Error(t, err)
		assert.Nil(t, buf)
	}
}

func testYamlJSON(t *testing.T, str string, drive DriverType, good bool) {
	testYaml(t, str+"\n", drive, good)
	testJSON(t, `"`+str+`"`, drive, good)
}
func testYaml(t *testing.T, str string, drive DriverType, good bool) {
	testMarshalUnmarshal(t, str, drive, good, yaml.Marshal, yaml.Unmarshal)
}
func testJSON(t *testing.T, str string, drive DriverType, good bool) {
	testMarshalUnmarshal(t, str, drive, good, json.Marshal, json.Unmarshal)
}

func TestDriverType(t *testing.T) {
	assert.Equal(t, "mysql", MySQL.String())
	assert.Equal(t, "redis", Redis.String())
	assert.Equal(t, "memcache", Memcache.String())
	assert.Equal(t, "mongo", Mongo.String())

	d := Unknown
	assert.Equal(t, MySQL, d.DriverType("mysql"))
	assert.Equal(t, Redis, d.DriverType("redis"))
	assert.Equal(t, Memcache, d.DriverType("memcache"))

	d = 10
	assert.Equal(t, "unknown", d.String())

	testYamlJSON(t, "mysql", MySQL, true)
	testYamlJSON(t, "redis", Redis, true)
	testYamlJSON(t, "memcache", Memcache, true)
	testYamlJSON(t, "kafka", Kafka, true)
	testYamlJSON(t, "mongo", Mongo, true)
	testYamlJSON(t, "postgres", PostgreSQL, true)
	testYamlJSON(t, "zookeeper", Zookeeper, true)
	testYamlJSON(t, "unknown", Unknown, false)

	testJSON(t, "", Unknown, false)
	testJSON(t, `{"x":"y"}`, Unknown, false)
	testJSON(t, `"xyz"`, Unknown, false)
	testYaml(t, "- mysql::", Unknown, false)

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
