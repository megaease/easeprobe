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
	"errors"
	"io/ioutil"
	"net"
	"reflect"
	"testing"
	"time"

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
				DefaultProbe: base.DefaultProbe{
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
				DefaultProbe: base.DefaultProbe{
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

	bastion := Endpoint{
		PrivateKey: "my.key",
		Host:       "bastion.my.com",
		User:       "ubuntu",
		Password:   "pass",
	}
	_ssh.Servers[0].SetBastion(&bastion)
	assert.Equal(t, "bastion.my.com:22", _ssh.Servers[0].bastion.Host)
	assert.Equal(t, "my.key", _ssh.Servers[0].bastion.PrivateKey)
	assert.Equal(t, "ubuntu", _ssh.Servers[0].bastion.User)
	assert.Equal(t, "pass", _ssh.Servers[0].bastion.Password)

	bastion1 := Endpoint{
		Host: "test:12:23",
	}
	_ssh.Servers[0].SetBastion(&bastion1)
	assert.Equal(t, "bastion.my.com:22", _ssh.Servers[0].bastion.Host)

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

	// NewClientConn failed
	monkey.Patch(ssh.NewClientConn, func(c net.Conn, addr string, config *ssh.ClientConfig) (ssh.Conn, <-chan ssh.NewChannel, <-chan *ssh.Request, error) {
		return &ssh.Client{}, nil, nil, errors.New("NewClientConn failed")
	})
	for _, s := range _ssh.Servers {
		if s.bastion != nil {
			status, message := s.DoProbe()
			assert.Equal(t, false, status)
			assert.Contains(t, message, "NewClientConn failed")
		}
	}

	// ssh Client Dial failed
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "Dial", func(c *ssh.Client, n, a string) (net.Conn, error) {
		return nil, errors.New("ssh Client Dial failed")
	})
	for _, s := range _ssh.Servers {
		if s.bastion != nil {
			status, message := s.DoProbe()
			assert.Equal(t, false, status)
			assert.Contains(t, message, "ssh Client Dial failed")
		}
	}

	// ssh Dial failed
	monkey.Patch(ssh.Dial, func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		return nil, errors.New("ssh Dial failed")
	})
	for _, s := range _ssh.Servers {
		status, message := s.DoProbe()
		assert.Equal(t, false, status)
		assert.Contains(t, message, "ssh Dial failed")
	}

	// SSHConfig failed
	var ed *Endpoint
	monkey.PatchInstanceMethod(reflect.TypeOf(ed), "SSHConfig", func(e *Endpoint, _, _ string, _ time.Duration) (*ssh.ClientConfig, error) {
		return nil, errors.New("SSHConfig failed")
	})
	for _, s := range _ssh.Servers {
		status, message := s.DoProbe()
		assert.Equal(t, false, status)
		assert.Contains(t, message, "SSHConfig failed")
	}

	monkey.UnpatchAll()
	status, _ := _ssh.Servers[1].DoProbe()
	assert.Equal(t, false, status)
}
