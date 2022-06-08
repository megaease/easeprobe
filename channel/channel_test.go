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

package channel

import (
	"testing"

	"github.com/megaease/easeprobe/notify"
	"github.com/megaease/easeprobe/probe"
	"github.com/stretchr/testify/assert"
)

func TestChannel(t *testing.T) {
	ch := NewEmpty("test")

	ch.Config()
	ch.SetNotify(nil)
	ch.SetProber(nil)

	assert.NotNil(t, ch.Done())
	assert.NotNil(t, ch.Channel())

	probers := []probe.Prober{
		newDummyProber("http", "XY", "dummy-XY", []string{"X", "Y"}),
		newDummyProber("http", "X", "dummy-X", []string{"X"}),
	}
	ch.SetProbers(probers)
	assert.Equal(t, 2, len(ch.Probers))
	assert.Equal(t, "http", ch.GetProber("dummy-XY").Kind())

	notifiers := []notify.Notify{
		newDummyNotify("email", "dummy-XY", []string{"X", "Y"}),
		newDummyNotify("email", "dummy-X", []string{"X"}),
	}
	ch.SetNotifiers(notifiers)
	assert.Equal(t, 2, len(ch.Notifiers))
	assert.Equal(t, "email", ch.GetNotify("dummy-XY").Kind())

	// test duplicate name
	n := newDummyNotify("discord", "dummy-XY", []string{"X", "Y"})
	ch.SetNotify(n)
	assert.NotEqual(t, "discord", ch.GetNotify("dummy-XY").Kind())
	assert.Equal(t, "email", ch.GetNotify("dummy-XY").Kind())

	p := newDummyProber("ssh", "XY", "dummy-XY", []string{"X", "Y"})
	ch.SetProber(p)
	assert.NotEqual(t, "ssh", ch.GetProber("dummy-XY").Kind())
	assert.Equal(t, "http", ch.GetProber("dummy-XY").Kind())

}
