package probe

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestStatus(t *testing.T) {
	s := StatusUp
	assert.Equal(t, "up", s.String())
	s.Status("down")
	assert.Equal(t, StatusDown, s)
	assert.Equal(t, "❌", s.Emoji())
	s.Status("up")
	assert.Equal(t, StatusUp, s)
	assert.Equal(t, "✅", s.Emoji())

	err := yaml.Unmarshal([]byte("down"), &s)
	assert.Nil(t, err)
	assert.Equal(t, StatusDown, s)

	buf, err := yaml.Marshal(&s)
	assert.Nil(t, err)
	assert.Equal(t, "down\n", string(buf))

	buf, err = json.Marshal(s)
	assert.Nil(t, err)
	assert.Equal(t, "\"down\"", string(buf))

	err = yaml.Unmarshal([]byte("xxx"), &s)
	assert.Nil(t, err)
	assert.Equal(t, StatusUnknown, s)

	err = yaml.Unmarshal([]byte{1, 2}, &s)
	assert.NotNil(t, err)
}
