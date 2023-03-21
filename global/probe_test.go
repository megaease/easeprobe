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
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestProbe(t *testing.T) {
	p := ProbeSettings{}

	r := p.NormalizeTimeOut(0)
	assert.Equal(t, DefaultTimeOut, r)

	r = p.NormalizeTimeOut(10)
	assert.Equal(t, time.Duration(10), r)

	p.Timeout = 20
	r = p.NormalizeTimeOut(0)
	assert.Equal(t, time.Duration(20), r)

	r = p.NormalizeInterval(0)
	assert.Equal(t, DefaultProbeInterval, r)

	r = p.NormalizeInterval(10)
	assert.Equal(t, time.Duration(10), r)

	p.Interval = 20
	r = p.NormalizeInterval(0)
	assert.Equal(t, time.Duration(20), r)
}

func TestStatusChangeThresholdSettings(t *testing.T) {
	p := ProbeSettings{}

	r := p.NormalizeThreshold(StatusChangeThresholdSettings{})
	assert.Equal(t, StatusChangeThresholdSettings{
		Failure: DefaultStatusChangeThresholdSetting,
		Success: DefaultStatusChangeThresholdSetting,
	}, r)

	p.Failure = 2
	p.Success = 3

	r = p.NormalizeThreshold(StatusChangeThresholdSettings{
		Failure: 1,
	})
	assert.Equal(t, StatusChangeThresholdSettings{
		Failure: 1,
		Success: 3,
	}, r)

	r = p.NormalizeThreshold(StatusChangeThresholdSettings{
		Success: 2,
	})
	assert.Equal(t, StatusChangeThresholdSettings{
		Failure: 2,
		Success: 2,
	}, r)

	r = p.NormalizeThreshold(StatusChangeThresholdSettings{
		Failure: 5,
		Success: 6,
	})
	assert.Equal(t, StatusChangeThresholdSettings{
		Failure: 5,
		Success: 6,
	}, r)

	r = p.NormalizeThreshold(StatusChangeThresholdSettings{
		Failure: 0,
	})
	assert.Equal(t, StatusChangeThresholdSettings{
		Failure: 2,
		Success: 3,
	}, r)

	r = p.NormalizeThreshold(StatusChangeThresholdSettings{
		Success: -1,
	})
	assert.Equal(t, StatusChangeThresholdSettings{
		Failure: 2,
		Success: 3,
	}, r)

	p.Failure = -1
	r = p.NormalizeThreshold(StatusChangeThresholdSettings{
		Failure: 0,
	})
	assert.Equal(t, StatusChangeThresholdSettings{
		Failure: 1,
		Success: 3,
	}, r)
}

func TestNotificationStrategySettings(t *testing.T) {
	p := ProbeSettings{}
	n := p.NormalizeNotificationStrategy(NotificationStrategySettings{})
	assert.Equal(t, NotificationStrategySettings{
		Strategy: RegularStrategy,
		Factor:   DefaultNotificationFactor,
		MaxTimes: DefaultMaxNotificationTimes,
	}, n)

	p.Strategy = IncrementStrategy
	p.MaxTimes = 10
	n = p.NormalizeNotificationStrategy(NotificationStrategySettings{Strategy: ExponentialStrategy})
	assert.Equal(t, NotificationStrategySettings{
		Strategy: ExponentialStrategy,
		Factor:   DefaultNotificationFactor,
		MaxTimes: 10,
	}, n)

	p.Factor = -1
	n = p.NormalizeNotificationStrategy(NotificationStrategySettings{MaxTimes: 20})
	assert.Equal(t, NotificationStrategySettings{
		Strategy: IncrementStrategy,
		Factor:   DefaultNotificationFactor,
		MaxTimes: 20,
	}, n)

	p.Factor = 2
	n = p.NormalizeNotificationStrategy(NotificationStrategySettings{Factor: 3, MaxTimes: 20})
	assert.Equal(t, NotificationStrategySettings{
		Strategy: IncrementStrategy,
		Factor:   3,
		MaxTimes: 20,
	}, n)

	n = p.NormalizeNotificationStrategy(NotificationStrategySettings{Strategy: RegularStrategy, Factor: 1, MaxTimes: 5})
	assert.Equal(t, NotificationStrategySettings{
		Strategy: RegularStrategy,
		Factor:   1,
		MaxTimes: 5,
	}, n)

}

func testNotifyMarshalUnmarshal(t *testing.T, str string, ns IntervalStrategy, good bool,
	marshal func(in interface{}) ([]byte, error),
	unmarshal func(in []byte, out interface{}) (err error)) {

	var s IntervalStrategy
	err := unmarshal([]byte(str), &s)
	if good {
		assert.Nil(t, err)
		assert.Equal(t, ns, s)
	} else {
		assert.Error(t, err)
		assert.Equal(t, Unknown, s)
	}

	buf, err := marshal(ns)
	if good {
		assert.Nil(t, err)
		assert.Equal(t, str, string(buf))
	} else {
		assert.Error(t, err)
		assert.Nil(t, buf)
	}
}

func testNotifyYaml(t *testing.T, str string, ns IntervalStrategy, good bool) {
	testNotifyMarshalUnmarshal(t, str, ns, good, yaml.Marshal, yaml.Unmarshal)
}
func testNotifyJSON(t *testing.T, str string, ns IntervalStrategy, good bool) {
	testNotifyMarshalUnmarshal(t, str, ns, good, json.Marshal, json.Unmarshal)
}
func testNotifyYamlJSON(t *testing.T, str string, ns IntervalStrategy, good bool) {
	testNotifyYaml(t, str+"\n", ns, good)
	testNotifyJSON(t, `"`+str+`"`, ns, good)
}

func TestNotificationIntervalStrategy(t *testing.T) {

	assert.Equal(t, "regular", RegularStrategy.String())
	assert.Equal(t, "increment", IncrementStrategy.String())
	assert.Equal(t, "exponent", ExponentialStrategy.String())

	var s IntervalStrategy
	s.IntervalStrategy("regular")
	assert.Equal(t, RegularStrategy, s)
	s.IntervalStrategy("Regular")
	assert.Equal(t, RegularStrategy, s)
	s.IntervalStrategy("REGULAR")
	assert.Equal(t, RegularStrategy, s)
	s.IntervalStrategy("increment")
	assert.Equal(t, IncrementStrategy, s)
	s.IntervalStrategy("exponent")
	assert.Equal(t, ExponentialStrategy, s)
	s.IntervalStrategy("unknown")
	assert.Equal(t, UnknownStrategy, s)
	s.IntervalStrategy("bad")
	assert.Equal(t, UnknownStrategy, s)

	testNotifyYamlJSON(t, "regular", RegularStrategy, true)
	testNotifyYamlJSON(t, "increment", IncrementStrategy, true)
	testNotifyYamlJSON(t, "exponent", ExponentialStrategy, true)
}
