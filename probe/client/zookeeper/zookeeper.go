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

package zookeeper

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/megaease/easeprobe/probe/client/conf"
	log "github.com/sirupsen/logrus"
)

// Kind is the type of driver
const Kind string = "ZooKeeper"

// Zookeeper is the Zookeeper client
type Zookeeper struct {
	conf.Options `yaml:",inline"`
	tls          *tls.Config     `yaml:"-"`
	Context      context.Context `yaml:"conn_str"`
}

// New create a Zookeeper client
func New(opt conf.Options) Zookeeper {
	tls, err := opt.TLS.Config()
	if err != nil {
		log.Errorf("[%s] %s - TLS Config error - %v", Kind, opt.ProbeName, err)
	}

	return Zookeeper{
		Options: opt,
		tls:     tls,
		Context: context.Background(),
	}
}

// Kind return the name of client
func (z Zookeeper) Kind() string {
	return Kind
}

// Probe do the health check
func (z Zookeeper) Probe() (bool, string) {
	var (
		conn *zk.Conn
		err  error
	)

	dialer := getDialer(z)
	conn, _, err = zk.ConnectWithDialer([]string{z.Host}, z.Timeout(), dialer)
	if err != nil {
		return false, err.Error()
	}
	defer conn.Close()

	_, _, err = conn.Get("/")
	if err != nil {
		return false, err.Error()
	}

	return true, "Check Zookeeper Server Successfully!"
}

func getDialer(z Zookeeper) func(network string, address string, _ time.Duration) (net.Conn, error) {
	if z.tls == nil {
		return net.DialTimeout
	}

	return func(network, address string, _ time.Duration) (net.Conn, error) {
		tlsConfig := &tls.Config{
			Certificates:       z.tls.Certificates,
			RootCAs:            z.tls.RootCAs,
			InsecureSkipVerify: true,
		}

		ipConn, err := net.DialTimeout(network, z.Host, z.Timeout())
		if err != nil {
			return nil, err
		}

		tlsConn := tls.Client(ipConn, tlsConfig)
		err = tlsConn.Handshake()
		if err != nil {
			return nil, err
		}

		return tlsConn, nil
	}
}
