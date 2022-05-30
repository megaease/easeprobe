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
