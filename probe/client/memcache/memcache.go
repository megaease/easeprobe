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
	"context"

	MemcacheClient "github.com/bradfitz/gomemcache/memcache"
	"github.com/megaease/easeprobe/probe/client/conf"
)

// Kind is the type of driver
const Kind string = "Memcache"

// Memcache is the Memcache client
type Memcache struct {
	conf.Options `yaml:",inline"`
	Key          string          `yaml:",inline`
	Value        string          `yaml:",inline`
	Context      context.Context `yaml:"-"`
}

// New create a Memcache client
func New(opt conf.Options) Memcache {

	return Memcache{
		Options: opt,
		Context: context.Background(),
	}
}

// Kind return the name of client
func (r Memcache) Kind() string {
	return Kind
}

// Probe do the health check
func (r Memcache) Probe() (bool, string) {

	mc := MemcacheClient.New(r.Host)
	//	ctx, cancel := context.WithTimeout(r.Context, r.Timeout())
	//	defer cancel()

	it, err := mc.Get("sysconfig:event_active")
	if err != nil {
		return false, err.Error()
	}

	if string(it.Value) != "1" {
		return false, "Memcache value returned do not much"
	}

	return true, "Memcache key fetched Successfully!"

}
