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
