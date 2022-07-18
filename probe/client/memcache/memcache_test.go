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

package memcache

import (
	"reflect"
	"runtime"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/megaease/easeprobe/probe/client/conf"
	"github.com/stretchr/testify/assert"
)

func TestMemcache(t *testing.T) {
	conf := conf.Options{
		Host:       "localhost:12345",
		DriverType: conf.Memcache,
		Data:       map[string]string{"sysconfig:event_active": "1"},
	}

	m, e := New(conf)
	assert.Equal(t, "Memcache", m.Kind())
	assert.Nil(t, e)

	// since memcached is not running
	// confirm that we get error
	s, errmsg := m.Probe()
	assert.False(t, s)
	if runtime.GOOS == "windows" {
		assert.Contains(t, errmsg, "connect timeout")
	} else {
		assert.Contains(t, errmsg, "connection refused")
	}

	conf.Data = map[string]string{"sysconfig:event_active": "1"}
	assert.Equal(t, len(m.getDataKeys()), len(conf.Data))
	assert.True(t, len(m.getDataKeys()) > 0)

	// mock the memcached server
	var mc *memcache.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "GetMulti", func(*memcache.Client, []string) (map[string]*memcache.Item, error) {
		return map[string]*memcache.Item{
			"sysconfig:event_active": {
				Key:        "",
				Value:      []byte("1"),
				Flags:      0,
				Expiration: 0,
			},
		}, nil
	})

	s, msg := m.Probe()
	assert.True(t, s)
	assert.Contains(t, msg, "successfully")

	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "GetMulti", func(*memcache.Client, []string) (map[string]*memcache.Item, error) {
		return map[string]*memcache.Item{
			"sysconfig:event_active": {
				Key:        "sysconfig:event_active",
				Value:      []byte("2"),
				Flags:      0,
				Expiration: 0,
			},
		}, nil
	})
	s, msg = m.Probe()
	assert.False(t, s)
	assert.Contains(t, msg, "expected")

	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "GetMulti", func(*memcache.Client, []string) (map[string]*memcache.Item, error) {
		return map[string]*memcache.Item{
			"sysconfig:event_active": {
				Key:        "sysconfig:event_active",
				Value:      []byte("1"),
				Flags:      0,
				Expiration: 0,
			},
		}, nil
	})
	s, msg = m.Probe()
	assert.True(t, s)
	assert.Contains(t, msg, "successfully")

	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "GetMulti", func(*memcache.Client, []string) (map[string]*memcache.Item, error) {
		return map[string]*memcache.Item{}, nil
	})
	s, msg = m.Probe()
	assert.False(t, s)
	assert.Contains(t, msg, "expected")

	m.Data = map[string]string{}
	m.ProbeTimeout = time.Second
	s, msg = m.Probe()
	assert.False(t, s)

	monkey.PatchInstanceMethod(reflect.TypeOf(mc), "Ping", func(*memcache.Client) error {
		return nil
	})

	s, msg = m.Probe()
	assert.True(t, s)
	assert.Contains(t, msg, "Successfully")

	monkey.UnpatchAll()

}
