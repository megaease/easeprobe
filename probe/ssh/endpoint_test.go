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

package ssh

import "testing"

func check(t *testing.T, fname string, err error, result, expected string) {
	if err != nil {
		t.Errorf("%s error: %s", fname, err)
	}
	if result != expected {
		t.Errorf("%s result: %s, expected: %s", fname, result, expected)
	}
}

func TestParseHost(t *testing.T) {
	e := Endpoint{}

	const fname = "ParseHost"
	e.Host = "example.com:22"
	check(t, fname, e.ParseHost(), e.Host, "example.com:22")

	e.Host = "192.168.1.1"
	check(t, fname, e.ParseHost(), e.Host, "192.168.1.1:22")

	e.Host = "example.com:2222"
	check(t, fname, e.ParseHost(), e.Host, "example.com:2222")

	e.Host = "user@example.com"
	check(t, fname, e.ParseHost(), e.Host, "example.com:22")
	check(t, fname, nil, e.User, "user")
}
