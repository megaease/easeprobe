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
