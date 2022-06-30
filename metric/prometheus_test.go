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

package metric

import (
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func TestGetName(t *testing.T) {
	var expected, result string

	expected = ""
	result = GetName(" ", "\n", "")
	assert.Equal(t, expected, result)
	assert.Equal(t, ValidMetricName(result), false)

	expected = "namespace_name_metric"
	result = GetName("namespace", "", "name", "metric")
	assert.Equal(t, expected, result)
	assert.Equal(t, ValidMetricName(result), true)

	expected = "name_metric"
	result = GetName("", "", "name", "metric")
	assert.Equal(t, expected, result)
	assert.Equal(t, ValidMetricName(result), true)

	expected = "namespace_subsystem_name"
	result = GetName("namespace", "subsystem", "name", "")
	assert.Equal(t, expected, result)
	assert.Equal(t, ValidMetricName(result), true)

	expected = "namespace_subsystem_name_metric"
	result = GetName("namespace", "subsystem", "name", "metric")
	assert.Equal(t, expected, result)
	assert.Equal(t, ValidMetricName(result), true)

	expected = "namespace_subsystemtest_name_metric"
	result = GetName("name@!$space", "sub-system(test)", "name", "metric")
	assert.Equal(t, expected, result)
	assert.Equal(t, ValidMetricName(result), true)

	expected = "namespace_subsystem:test_name_metric3"
	result = GetName("namespace", "subsystem:test", "123name", "metric3")
	assert.Equal(t, expected, result)
	assert.Equal(t, ValidMetricName(result), true)
}

func TestNewMetrics(t *testing.T) {
	NewCounter("namespace", "subsystem", "counter", "metric",
		"help", []string{"label1", "label2"})
	assert.NotNil(t, GetName("namespace_subsystem_counter_metric"))
	assert.NotNil(t, Counter("namespace_subsystem_counter_metric"))

	NewGauge("namespace", "subsystem", "gauge", "metric",
		"help", []string{"label1", "label2"})
	assert.NotNil(t, GetName("namespace_subsystem_gauge_metric"))
	assert.NotNil(t, Gauge("namespace_subsystem_gauge_metric"))
}

func TestName(t *testing.T) {

	assert.False(t, ValidMetricName(""))
	assert.False(t, ValidMetricName(" "))
	assert.False(t, ValidMetricName("\n"))
	assert.False(t, ValidMetricName("5name"))
	assert.False(t, ValidMetricName("name%"))
	assert.False(t, ValidMetricName("hello-world"))
	assert.False(t, ValidMetricName("hello-world@"))

	assert.True(t, ValidMetricName("name5"))
	assert.True(t, ValidMetricName(":name"))
	assert.True(t, ValidMetricName("hello_world:name"))
	assert.True(t, ValidMetricName("_hello_world:name"))
	assert.True(t, ValidMetricName(":_hello_world:name"))
	assert.True(t, ValidMetricName("namespace_name_metric"))

	assert.False(t, ValidLabelName(""))
	assert.False(t, ValidLabelName(" "))
	assert.False(t, ValidLabelName("\n"))
	assert.False(t, ValidLabelName("5name"))
	assert.False(t, ValidLabelName("name%"))
	assert.False(t, ValidLabelName("hello-world"))
	assert.False(t, ValidLabelName("hello-world@"))

	assert.True(t, ValidLabelName("_name5"))
	assert.True(t, ValidLabelName("name_"))
	assert.True(t, ValidLabelName("name5"))
	assert.True(t, ValidLabelName("hello_world"))
	assert.True(t, ValidLabelName("_hello_world_"))
	assert.True(t, ValidLabelName("_hello_world_1_"))
}

func TestDuplicateName(t *testing.T) {
	counter1 := NewCounter("namespace", "subsystem", "counter", "metric",
		"help", []string{})
	counter2 := NewCounter("namespace", "subsystem", "counter", "metric",
		"help", []string{})
	assert.Equal(t, counter1, counter2)

	gauge1 := NewGauge("namespace", "subsystem", "gauge", "metric",
		"help", []string{})
	gauge2 := NewGauge("namespace", "subsystem", "gauge", "metric",
		"help", []string{})
	assert.Equal(t, gauge1, gauge2)
}

func TestInvalidName(t *testing.T) {

	//label errors
	counter := NewCounter("namespace", "subsystem", "counter", "metric",
		"help", []string{"label-1", "label:2"})
	assert.Nil(t, counter)

	gauge := NewGauge("namespace", "subsystem", "gauge", "metric",
		"help", []string{"label-1", "label:2"})
	assert.Nil(t, gauge)

	monkey.Patch(ValidMetricName, func(name string) bool {
		return false
	})
	counter = NewCounter("namespace", "subsystem", "counter", "metric",
		"help", []string{})
	assert.Nil(t, counter)

	monkey.UnpatchAll()
}
