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

	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Endpoint is SSH Endpoint
type Endpoint struct {
	PrivateKey string      `yaml:"key"`
	Host       string      `yaml:"host"`
	User       string      `yaml:"username"`
	Password   string      `yaml:"password"`
	client     *ssh.Client `yaml:"-"`
}

// ParseHost check the host is configured the port or not
func (e *Endpoint) ParseHost() error {

	if strings.LastIndex(e.Host, ":") < 0 {
		e.Host = e.Host + ":22"
	}
	userIdx := strings.Index(e.Host, "@")
	if userIdx > 0 {
		e.User = e.Host[:userIdx]
		e.Host = e.Host[userIdx+1:]
	}
	_, _, err := net.SplitHostPort(e.Host)

	if err != nil {
		return err
	}
	return nil
}

// SSHConfig returns the ssh.ClientConfig
func (e *Endpoint) SSHConfig(kind, name string, timeout time.Duration) (*ssh.ClientConfig, error) {
	var Auth []ssh.AuthMethod

	if len(e.Password) > 0 {
		Auth = append(Auth, ssh.Password(e.Password))
	}

	if len(e.PrivateKey) > 0 {
		key, err := ioutil.ReadFile(e.PrivateKey)
		if err != nil {
			return nil, err
		}

		// Create the Signer for this private key.
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}

		Auth = append(Auth, ssh.PublicKeys(signer))
	}

	config := &ssh.ClientConfig{
		User:            e.User,
		Auth:            Auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	return config, nil
}
