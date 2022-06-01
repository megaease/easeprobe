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

import (
	"io/ioutil"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

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

	e.Host = "xx.com:"
	check(t, fname, e.ParseHost(), e.Host, "xx.com:22")

	e.Host = ":22"
	check(t, fname, e.ParseHost(), e.Host, "localhost:22")
}

func TestSSHConfig(t *testing.T) {
	e := Endpoint{}
	e.Host = "example.com:22"
	e.User = "user"
	e.Password = "password"
	e.PrivateKey = "key"

	config, err := e.SSHConfig("ssh", "test", 30*time.Second)
	assert.Nil(t, config)
	assert.NotNil(t, err)

	monkey.Patch(ioutil.ReadFile, func(filename string) ([]byte, error) {
		return []byte(`
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAaAAAABNlY2RzYS
1zaGEyLW5pc3RwMjU2AAAACG5pc3RwMjU2AAAAQQR9WZPeBSvixkhjQOh9yCXXlEx5CN9M
yh94CJJ1rigf8693gc90HmahIR5oMGHwlqMoS7kKrRw+4KpxqsF7LGvxAAAAqJZtgRuWbY
EbAAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBH1Zk94FK+LGSGNA
6H3IJdeUTHkI30zKH3gIknWuKB/zr3eBz3QeZqEhHmgwYfCWoyhLuQqtHD7gqnGqwXssa/
EAAAAgBzKpRmMyXZ4jnSt3ARz0ul6R79AXAr5gQqDAmoFeEKwAAAAOYWpAYm93aWUubG9j
YWwBAg==
-----END OPENSSH PRIVATE KEY-----`), nil
	})
	config, err = e.SSHConfig("ssh", "test", 30*time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, config)
}
