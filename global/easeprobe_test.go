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
	"fmt"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func TestEaseProbe(t *testing.T) {
	e := GetEaseProbe()
	assert.Equal(t, DefaultProg, e.Name)
	assert.Equal(t, DefaultIconURL, e.IconURL)

	InitEaseProbe("test", "icon")
	e = GetEaseProbe()
	assert.Equal(t, "test", e.Name)
	assert.Equal(t, "icon", e.IconURL)
	assert.Equal(t, Ver, e.Version)

	h, err := os.Hostname()
	if err != nil {
		h = "unknown"
	}
	assert.Equal(t, h, e.Host)

	str := "test " + Ver + " @ " + h
	assert.Equal(t, str, FooterString())

}

// If you use VSCode run the test,
// make sure add the following test flag in settings.json
//	    "go.testFlags": ["-gcflags=-l"],
func TestEaseProbeFail(t *testing.T) {
	monkey.Patch(os.Hostname, func() (string, error) {
		return "", fmt.Errorf("error")
	})
	InitEaseProbe("test", "icon")
	e := GetEaseProbe()
	assert.Equal(t, "unknown", e.Host)

	monkey.Unpatch(os.Hostname)
}
