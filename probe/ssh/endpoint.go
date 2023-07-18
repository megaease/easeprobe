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
	"net"
	"os"

	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Endpoint is SSH Endpoint
type Endpoint struct {
	PrivateKey string      `yaml:"key" json:"key,omitempty" jsonschema:"title=Private Key,description=the private key file path for ssh login"`
	Passphrase string      `yaml:"passphrase" json:"passphrase,omitempty" jsonschema:"title=Passphrase,description=the passphrase for ssh private key"`
	Host       string      `yaml:"host" json:"host" jsonschema:"required,format=hostname,title=Host,description=the host for ssh probe"`
	User       string      `yaml:"username" json:"username,omitempty" jsonschema:"title=User,description=the username for ssh probe"`
	Password   string      `yaml:"password" json:"password,omitempty" jsonschema:"title=Password,description=the password for ssh probe"`
	client     *ssh.Client `yaml:"-" json:"-"`
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
	host, port, err := net.SplitHostPort(e.Host)

	if err != nil {
		return err
	}

	if host == "" {
		host = "localhost"
	}

	if port == "" {
		port = "22"
	}
	e.Host = host + ":" + port
	return nil
}

// SSHConfig returns the ssh.ClientConfig
func (e *Endpoint) SSHConfig(kind, name string, timeout time.Duration) (*ssh.ClientConfig, error) {
	var Auth []ssh.AuthMethod

	if len(e.Password) > 0 {
		Auth = append(Auth, ssh.Password(e.Password))
	}

	if len(e.PrivateKey) > 0 {
		key, err := os.ReadFile(e.PrivateKey)
		if err != nil {
			return nil, err
		}

		// Create the Signer for this private key.
		var signer ssh.Signer

		if len(e.Passphrase) > 0 {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(e.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(key)
		}

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
