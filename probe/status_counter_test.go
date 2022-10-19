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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStatusCounter(t *testing.T) {
	const Len = 3
	s := NewStatusCounter(Len)
	assert.Equal(t, Len, s.MaxLen)
	assert.True(t, s.CurrentStatus)

	for i := 1; i <= Len+2; i++ {
		s.AppendStatus(false, "failure")
		assert.False(t, s.CurrentStatus)
		if i <= Len {
			assert.Equal(t, i, s.StatusCount)
		} else {
			assert.Equal(t, Len, s.StatusCount)
		}
	}

	for i := 1; i <= Len+2; i++ {
		s.AppendStatus(true, "success")
		assert.True(t, s.CurrentStatus)
		if i <= Len {
			assert.Equal(t, i, s.StatusCount)
		} else {
			assert.Equal(t, Len, s.StatusCount)
		}
	}

	s1 := s.Clone()
	assert.True(t, reflect.DeepEqual(s, &s1))

	s1.SetMaxLen(2)
	assert.Equal(t, 2, s1.MaxLen)
	assert.Equal(t, 2, len(s1.StatusHistory))
}
