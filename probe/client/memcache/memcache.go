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

// Package memcache is the native client probe for memcache
package memcache

import (
	"context"
	"fmt"
	"strings"

	MemcacheClient "github.com/bradfitz/gomemcache/memcache"
	"github.com/megaease/easeprobe/probe/client/conf"
	log "github.com/sirupsen/logrus"
)

// Kind is the type of driver
const Kind string = "Memcache"

// Memcache is the Memcache client
type Memcache struct {
	conf.Options `yaml:",inline"`
	Context      context.Context `yaml:"-" json:"-"`
}

// New create a Memcache client
func New(opt conf.Options) (*Memcache, error) {
	return &Memcache{
		Options: opt,
		Context: context.Background(),
	}, nil
}

// Kind return the name of client
func (m *Memcache) Kind() string {
	return Kind
}

// Probe do the health check
func (m *Memcache) Probe() (bool, string) {
	// TODO: Add SASL AUTH and protocol details
	mc := MemcacheClient.New(m.Host)
	mc.Timeout = m.Timeout()

	// Check if we need to query specific keys or not
	if len(m.Data) > 0 {
		keys := m.getDataKeys()

		// TODO: mc.GetMulti(ctx, keys)
		items, err := mc.GetMulti(keys)
		if err != nil {
			return false, err.Error()
		}

		if len(items) != len(m.Data) {
			return false, fmt.Sprintf("Number of fetched keys %d expected %d", len(items), len(m.Data))
		}

		return m.validateKeyValues(items)
	}

	log.Debugf("[%s / %s %s] Data empty, Pinging", m.ProbeKind, m.ProbeName, m.ProbeTag)
	err := mc.Ping()
	if err != nil {
		return false, err.Error()
	}

	return true, "Memcache key fetched Successfully!"
}

// Slice the keys only from the configuration file
func (m *Memcache) getDataKeys() []string {
	keys := make([]string, len(m.Data))
	i := 0
	for k := range m.Data {
		keys[i] = k
		i++
	}

	return keys
}

// Validate memcache items against configuration data
func (m *Memcache) validateKeyValues(items map[string]*MemcacheClient.Item) (bool, string) {
	// iterate the keys and confirm their values match
	for _, item := range items {
		log.Debugf("[%s / %s / %s] Got key: %s with value: %s", m.ProbeKind, m.ProbeName, m.ProbeTag, item.Key, string(item.Value))
		if strings.TrimSpace(m.Data[item.Key]) == "" {
			log.Debugf("[%s / %s / %s] Skipping value check for item %s", m.ProbeKind, m.ProbeName, m.ProbeTag, item.Key)
			continue
		}
		if string(item.Value) != m.Data[item.Key] {
			return false, fmt.Sprintf("Memcache value for key %s returned %s, expected %s", item.Key, string(item.Value), string(m.Data[item.Key]))
		}
	}
	return true, "Memcache key values match successfully"
}
