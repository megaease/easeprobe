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

package global

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNotify(t *testing.T) {
	n := NotifySettings{
		Timeout: 0,
		Retry: Retry{
			Times:    0,
			Interval: 0,
		},
	}

	r := n.NormalizeTimeOut(0)
	assert.Equal(t, DefaultTimeOut, r)

	r = n.NormalizeTimeOut(10)
	assert.Equal(t, time.Duration(10), r)

	n.Timeout = 20
	r = n.NormalizeTimeOut(0)
	assert.Equal(t, time.Duration(20), r)

	retry := n.NormalizeRetry(Retry{Times: 10, Interval: 0})
	assert.Equal(t, Retry{Times: 10, Interval: DefaultRetryInterval}, retry)

	retry = n.NormalizeRetry(Retry{Times: 0, Interval: 10})
	assert.Equal(t, Retry{Times: DefaultRetryTimes, Interval: 10}, retry)

	retry = n.NormalizeRetry(Retry{Times: 10, Interval: 10})
	assert.Equal(t, Retry{Times: 10, Interval: 10}, retry)

	n.Retry.Times = 20
	retry = n.NormalizeRetry(Retry{Times: 0, Interval: 0})
	assert.Equal(t, Retry{Times: 20, Interval: DefaultRetryInterval}, retry)

	n.Retry.Interval = 20
	retry = n.NormalizeRetry(Retry{Times: 0, Interval: 0})
	assert.Equal(t, Retry{Times: 20, Interval: 20}, retry)
}
