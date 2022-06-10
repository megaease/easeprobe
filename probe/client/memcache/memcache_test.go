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
	"runtime"
	"testing"

	"github.com/megaease/easeprobe/probe/client/conf"
	"github.com/stretchr/testify/assert"
)

func TestMemcache(t *testing.T) {
	conf := conf.Options{
		Host:       "localhost:12345",
		DriverType: conf.Memcache,
		Data:       map[string]string{"sysconfig:event_active": "1"},
	}

	m := New(conf)
	assert.Equal(t, "Memcache", m.Kind())

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

}
