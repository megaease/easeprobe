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
	"net"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

func createSSHConfig() *SSH {
	BastionMap = make(BastionMapType)
	BastionMap["aws"] = Endpoint{
		PrivateKey: "aws.key",
		Host:       "bastion.aws.com",
		User:       "root",
		Password:   "pass",
	}

	BastionMap["gcp"] = Endpoint{
		PrivateKey: "gcp.key",
		Host:       "bastion.gcp.com:2222",
		User:       "root",
		Password:   "pass",
	}

	BastionMap["empty"] = Endpoint{}
	BastionMap["error"] = Endpoint{Host: "asdf:asdf:22"}

	return &SSH{
		Bastion: &BastionMap,
		Servers: []Server{
			{
				DefaultOptions: base.DefaultOptions{
					ProbeName: "Server One",
				},
				Endpoint: Endpoint{
					PrivateKey: "server1.key",
					Host:       "server.example.com",
					User:       "ubuntu",
					Password:   "",
					client:     &ssh.Client{},
				},
				Command:    "test",
				Args:       []string{},
				Env:        []string{},
				Contain:    "good",
				NotContain: "bad",
				BastionID:  "aws",
			},
			{
				DefaultOptions: base.DefaultOptions{
					ProbeName: "Server Two",
				},
				Endpoint: Endpoint{
					PrivateKey: "server2.key",
					Host:       "server.example.com:2222",
					User:       "ubuntu",
					Password:   "",
					client:     &ssh.Client{},
				},
				Command:   "test",
				Args:      []string{},
				Env:       []string{},
				BastionID: "none",
			},
		},
	}
}

func TestSSH(t *testing.T) {
	_ssh := createSSHConfig()
	_ssh.Bastion.ParseAllBastionHost()

	assert.Equal(t, 3, len(*_ssh.Bastion))
	assert.Equal(t, "bastion.aws.com:22", BastionMap["aws"].Host)
	assert.Equal(t, "bastion.gcp.com:2222", BastionMap["gcp"].Host)
	assert.Equal(t, "localhost:22", BastionMap["empty"].Host)

	global.InitEaseProbe("EaseProbeTest", "none")
	gConf := global.ProbeSettings{}
	for i := 0; i < len(_ssh.Servers); i++ {
		s := &_ssh.Servers[i]
		err := s.Config(gConf)
		assert.Nil(t, err)
		if s.BastionID == "none" {
			assert.Nil(t, s.bastion)
		}
	}

	monkey.Patch(ssh.Dial, func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		c, _, _, _ := ssh.NewClientConn(nil, "", config)
		return &ssh.Client{Conn: c}, nil
	})
	monkey.Patch(ssh.NewClient, func(c ssh.Conn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request) *ssh.Client {
		c, _, _, _ = ssh.NewClientConn(nil, "", nil)
		return &ssh.Client{Conn: c}
	})
	var client *ssh.Client
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Dial", func(c *ssh.Client, n, a string) (net.Conn, error) {
		return &net.TCPConn{}, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Close", func(c *ssh.Client) error {
		return nil
	})

	monkey.Patch(ssh.NewClientConn, func(c net.Conn, addr string, config *ssh.ClientConfig) (ssh.Conn, <-chan ssh.NewChannel, <-chan *ssh.Request, error) {
		return &ssh.Client{}, nil, nil, nil
	})

	monkey.PatchInstanceMethod(reflect.TypeOf(client), "NewSession", func(c *ssh.Client) (*ssh.Session, error) {
		return &ssh.Session{}, nil
	})

	var ss *ssh.Session
	monkey.PatchInstanceMethod(reflect.TypeOf(ss), "Run", func(s *ssh.Session, cmd string) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(ss), "Close", func(c *ssh.Session) error {
		return nil
	})

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

	for _, s := range _ssh.Servers {
		status, _ := s.DoProbe()
		switch s.ProbeName {
		case "Server One":
			assert.Equal(t, false, status)
		case "Server Two":
			assert.Equal(t, true, status)
		}
	}

	monkey.UnpatchAll()
	status, _ := _ssh.Servers[1].DoProbe()
	assert.Equal(t, false, status)
}
