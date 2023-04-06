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

package probe

import (
	"testing"

	"github.com/megaease/easeprobe/global"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestNotificationStrategy(t *testing.T) {
	strategy := NewNotificationStrategyData(global.RegularStrategy, 3)

	buf, err := yaml.Marshal(strategy)
	assert.Nil(t, err)
	str := string(buf)
	assert.Contains(t, str, "strategy: regular")
	assert.Contains(t, str, "max: 3")
	assert.Contains(t, str, "next: 1")
	assert.Contains(t, str, "failed: 0")
	assert.Contains(t, str, "step: 1")

	err = yaml.Unmarshal(buf, &strategy)
	assert.Nil(t, err)

	assert.Equal(t, global.RegularStrategy, strategy.Strategy)
	assert.Equal(t, 3, strategy.MaxTimes)
	assert.Equal(t, 0, strategy.FailedTimes)
	assert.Equal(t, 1, strategy.Next)
}

func testNotificationStrategy(t *testing.T, strategy global.IntervalStrategy, probeTimes int, notify []int) {
	s := NewNotificationStrategyData(strategy, len(notify))
	j := 0
	for i := 1; i <= probeTimes; i++ {
		send := s.NeedToSendNotification()
		assert.Equal(t, notify[j] == i, send)
		if send && j < len(notify)-1 {
			j++
		}
	}
}

func TestRegularNotificationStrategy(t *testing.T) {
	probeTimes := 10
	notify := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	testNotificationStrategy(t, global.RegularStrategy, probeTimes, notify)

	notify = []int{1, 2, 3}
	testNotificationStrategy(t, global.RegularStrategy, probeTimes, notify)
}

func TestIncrementNotificationStrategy(t *testing.T) {
	probeTimes := 66
	notify := []int{1, 3, 6, 10, 15, 21, 28, 36, 45, 55, 66}
	testNotificationStrategy(t, global.IncrementStrategy, probeTimes, notify)

	notify = []int{1, 3, 6, 10, 15}
	testNotificationStrategy(t, global.IncrementStrategy, probeTimes, notify)
}

func TestExponentiationNotificationStrategy(t *testing.T) {
	probeTimes := 64
	notify := []int{1, 2, 4, 8, 16, 32, 64}
	testNotificationStrategy(t, global.ExponentialStrategy, probeTimes, notify)

	notify = []int{1, 2, 4, 8, 16}
	testNotificationStrategy(t, global.ExponentialStrategy, probeTimes, notify)
}

func TestIllegalNotificationStrategy(t *testing.T) {
	s := NewNotificationStrategyData(global.IntervalStrategy(100), 3)
	notify := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for i := 1; i <= 10; i++ {
		send := s.NeedToSendNotification()
		if s.IsExceedMaxTimes() {
			assert.False(t, send)
		} else {
			assert.Equal(t, notify[i-1] == i, send)
		}
	}
}

func TestNotificationReset(t *testing.T) {
	s := NewNotificationStrategyData(global.IncrementStrategy, 3)
	notify := []int{1, 3, 6}
	test := func() {
		j := 0
		for i := 1; i <= 5; i++ {
			send := s.NeedToSendNotification()
			assert.Equal(t, notify[j] == i, send)
			if send && j < len(notify)-1 {
				j++
			}
		}
	}

	test()
	s.Reset()
	test()

}
