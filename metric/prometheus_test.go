package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetName(t *testing.T) {
	var expected, result string

	expected = ""
	result = GetName("", "", "")
	assert.Equal(t, expected, result)

	expected = "namespace_name_metric"
	result = GetName("namespace", "", "name", "metric")
	assert.Equal(t, expected, result)

	expected = "name_metric"
	result = GetName("", "", "name", "metric")
	assert.Equal(t, expected, result)

	expected = "namespace_subsystem_name"
	result = GetName("namespace", "subsystem", "name", "")
	assert.Equal(t, expected, result)

	expected = "namespace_subsystem_name_metric"
	result = GetName("namespace", "subsystem", "name", "metric")
	assert.Equal(t, expected, result)

	expected = "namespace_subsystemtest_name_metric"
	result = GetName("name@!$space", "subsystem(test)", "name", "metric")
	assert.Equal(t, expected, result)

}
