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

package kafka

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/client/conf"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestKaf(t *testing.T) {
	conf := conf.Options{
		Host:       "example.com",
		DriverType: conf.Kafka,
		Username:   "root",
		Password:   "pass",
		TLS:        global.TLS{},
	}

	kaf, err := New(conf)
	assert.Equal(t, "Kafka", kaf.Kind())
	assert.Nil(t, err)

	var dialer *kafka.Dialer
	monkey.PatchInstanceMethod(reflect.TypeOf(dialer), "DialContext", func(_ *kafka.Dialer, _ context.Context, _ string, _ string) (*kafka.Conn, error) {
		return &kafka.Conn{}, nil
	})
	var conn *kafka.Conn
	monkey.PatchInstanceMethod(reflect.TypeOf(conn), "ReadPartitions", func(k *kafka.Conn, topics ...string) ([]kafka.Partition, error) {
		return []kafka.Partition{{Topic: "topic1", ID: 1}, {Topic: "topic2", ID: 2}}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(conn), "Close", func(_ *kafka.Conn) error {
		return nil
	})

	s, m := kaf.Probe()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	// TLS test
	kaf.Password = ""
	s, m = kaf.Probe()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	monkey.UnpatchAll()
}

func TestKafkaFailed(t *testing.T) {
	conf := conf.Options{
		Host:       "example.com",
		DriverType: conf.Kafka,
		TLS: global.TLS{
			CA:   "kafka.ca",
			Cert: "kafka.cert",
			Key:  "kafka.pem",
		},
	}

	kaf, err := New(conf)
	//TLS failed
	assert.Nil(t, kaf)
	assert.NotNil(t, err)

	conf.TLS = global.TLS{}
	kaf, err = New(conf)
	assert.NotNil(t, kaf)
	assert.Nil(t, err)
	assert.Equal(t, "Kafka", kaf.Kind())

	var dialer *kafka.Dialer
	monkey.PatchInstanceMethod(reflect.TypeOf(dialer), "DialContext", func(_ *kafka.Dialer, _ context.Context, _ string, _ string) (*kafka.Conn, error) {
		return &kafka.Conn{}, nil
	})
	var conn *kafka.Conn
	monkey.PatchInstanceMethod(reflect.TypeOf(conn), "ReadPartitions", func(k *kafka.Conn, topics ...string) ([]kafka.Partition, error) {
		return []kafka.Partition{}, fmt.Errorf("get topics error")
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(conn), "Close", func(_ *kafka.Conn) error {
		return nil
	})

	s, m := kaf.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "get topics error")

	monkey.PatchInstanceMethod(reflect.TypeOf(dialer), "DialContext", func(_ *kafka.Dialer, _ context.Context, _ string, _ string) (*kafka.Conn, error) {
		return nil, fmt.Errorf("connection error")
	})
	s, m = kaf.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "connection error")
}
