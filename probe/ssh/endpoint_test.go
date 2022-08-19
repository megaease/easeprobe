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

	e.Passphrase = "123"
	monkey.Patch(ioutil.ReadFile, func(filename string) ([]byte, error) {
		return []byte(`
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAACmFlczI1Ni1jdHIAAAAGYmNyeXB0AAAAGAAAABBdkMia9n
8kS8QhkzdcB1fcAAAAEAAAAAEAAACXAAAAB3NzaC1yc2EAAAADAQABAAAAgQCxg/1PL60U
IQGBfMopFoMjdcujc/AvyycubRQiH2L8MVnPk7q7whRXMVHjlKX7zOnVzOMUMAwUov+rge
mLZtYVxMuGomHwD0Lc2wrHdmwgk+3fuTPX8VBZOEmErISZHi1q/9JxzZFybmXXkIfNk6FZ
CGXKWJ2dNBo2UxOKCxumqwAAAhAnEomH6JMqjVcnxxtUcvX+wtADoAEDh3Iw/HWK4OPHfQ
CzuW0rJGiuOkDs24QQxHgskqPEWwgC/aTfYMX9u2jPo8w2ta2NpiP4RsQ3cVoL32i25np4
JuYo4P/3Od3j82DnyV8NVwtN0m4EazCmY+aVM3iT/3VfhmTI4o/e0ZO/sMF8sxZTn8yUZa
W3MLSN1j4Nc5sKnOdTF8bQslJYWL2z9Q48E+y7XTkHGMZmmRfv4EKGHu4Nle/cRsLS5WSX
/eIUhUUI9FF4HD+Wmi+RTTBV1PdpaL6O4For6ot0i+CYotxe72++JmO51KCpKeVbWQiUgE
EyXgVoFPqg47SKpLKneFysRH9P0rEwVjx7Q2tYc3XeOENcs0gtgS7HHtNFzO1LRNWtZ8HA
v5v32pULoiuYtbgFeywmTGciLohsDtHfPKz0VHBdecjlYj8y/96BQ79V504QXt8j1dVdZ2
bAh6YI/rkK6kmTfZ7DvS8+ZDw/dJYln/31hiBqomkpsm64IPdcWtAGI8LqfIhYcfz60MqT
i/TNs5nhmqQDEcwD/Sq0YR2ItMAx7W5j1EZ72zpHa9KTjm+z/JKXlA6Cm2WN/pnO06gShZ
DX4y9QvXoaUiQ5vB63voiqRTBT4hTYBDVi2G7NEjIczNs9S8JQM5Mg52mZsdH77g6ChUPp
1gmeAwg2IKBY2y+HzQ/5xub5KjGDG6E=
-----END OPENSSH PRIVATE KEY-----
`), nil
	})
	config, err = e.SSHConfig("ssh", "test", 30*time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, config)

	monkey.Patch(ioutil.ReadFile, func(filename string) ([]byte, error) {
		return []byte(``), nil
	})
	config, err = e.SSHConfig("ssh", "test", 30*time.Second)
	assert.Nil(t, config)
	assert.NotNil(t, err)
}
