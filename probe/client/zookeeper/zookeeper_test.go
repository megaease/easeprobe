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
	"crypto/tls"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/client/conf"
	"github.com/stretchr/testify/assert"

	"github.com/go-zookeeper/zk"
)

func TestZooKeeper(t *testing.T) {
	conf := conf.Options{
		Host:       "127.0.0.0:2181",
		DriverType: conf.Zookeeper,
		Username:   "username",
		Password:   "password",
		TLS: global.TLS{
			CA:   "ca",
			Cert: "cert",
			Key:  "key",
		},
	}

	z, e := New(conf)
	assert.Nil(t, z)
	assert.NotNil(t, e)
	assert.Contains(t, e.Error(), "TLS Config Error")

	conf.TLS = global.TLS{}
	z, e = New(conf)
	assert.NotNil(t, z)
	assert.Nil(t, e)
	assert.Equal(t, "ZooKeeper", z.Kind())

	monkey.Patch(net.DialTimeout, func(network, address string, timeout time.Duration) (net.Conn, error) {
		return &net.TCPConn{}, nil
	})

	var conn *zk.Conn
	monkey.PatchInstanceMethod(reflect.TypeOf(conn), "Get", func(_ *zk.Conn, path string) ([]byte, *zk.Stat, error) {
		return []byte("test"), &zk.Stat{}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(conn), "Close", func(_ *zk.Conn) {
		return
	})

	monkey.Patch(zk.ConnectWithDialer, func(servers []string, sessionTimeout time.Duration, dialer zk.Dialer) (*zk.Conn, <-chan zk.Event, error) {
		return &zk.Conn{}, nil, nil
	})
	s, m := z.Probe()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	// TLS config success
	var tc *global.TLS
	monkey.PatchInstanceMethod(reflect.TypeOf(tc), "Config", func(_ *global.TLS) (*tls.Config, error) {
		return &tls.Config{}, nil
	})
	z, e = New(conf)
	assert.NotNil(t, z)
	assert.Nil(t, e)
	assert.NotNil(t, z.tls)

	s, m = z.Probe()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	// Get error
	monkey.PatchInstanceMethod(reflect.TypeOf(conn), "Get", func(_ *zk.Conn, path string) ([]byte, *zk.Stat, error) {
		return nil, nil, fmt.Errorf("get error")
	})
	s, m = z.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "get error")

	// Connect error
	monkey.Patch(zk.ConnectWithDialer, func(servers []string, sessionTimeout time.Duration, dialer zk.Dialer) (*zk.Conn, <-chan zk.Event, error) {
		return nil, nil, fmt.Errorf("connect error")
	})
	s, m = z.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "connect error")

	monkey.UnpatchAll()
}

func TestGetDialer(t *testing.T) {
	zConf := &Zookeeper{
		Options: conf.Options{
			Host:       "127.0.0.0:2181",
			DriverType: conf.Redis,
			Username:   "username",
			Password:   "password",
			TLS: global.TLS{
				CA:   "ca",
				Cert: "cert",
				Key:  "key",
			},
		},
		tls: &tls.Config{},
	}

	fn := getDialer(zConf)

	monkey.Patch(net.DialTimeout, func(network, address string, timeout time.Duration) (net.Conn, error) {
		return &net.TCPConn{}, nil
	})
	var tlsConn *tls.Conn
	monkey.PatchInstanceMethod(reflect.TypeOf(tlsConn), "Handshake", func(_ *tls.Conn) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(tlsConn), "Close", func(_ *tls.Conn) error {
		return nil
	})

	conn, err := fn("tcp", zConf.Host, time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, conn)

	monkey.PatchInstanceMethod(reflect.TypeOf(tlsConn), "Handshake", func(_ *tls.Conn) error {
		return fmt.Errorf("handshake error")
	})
	conn, err = fn("tcp", zConf.Host, time.Second)
	assert.Equal(t, "handshake error", err.Error())
	assert.Nil(t, conn)

	monkey.Patch(net.DialTimeout, func(network, address string, timeout time.Duration) (net.Conn, error) {
		return nil, fmt.Errorf("dial error")
	})
	conn, err = fn("tcp", zConf.Host, time.Second)
	assert.Equal(t, "dial error", err.Error())
	assert.Nil(t, conn)
}

func TestData(t *testing.T) {
	z := &Zookeeper{
		Options: conf.Options{
			Host:       "127.0.0.0:2181",
			DriverType: conf.Redis,
			Username:   "username",
			Password:   "password",
			Data: map[string]string{
				"test": "test",
			},
		},
	}

	monkey.Patch(getDialer, func(z *Zookeeper) func(string, string, time.Duration) (net.Conn, error) {
		return net.DialTimeout
	})
	monkey.Patch(zk.ConnectWithDialer, func(servers []string, sessionTimeout time.Duration, dialer zk.Dialer) (*zk.Conn, <-chan zk.Event, error) {
		return &zk.Conn{}, nil, nil
	})
	var conn *zk.Conn
	monkey.PatchInstanceMethod(reflect.TypeOf(conn), "Close", func(_ *zk.Conn) {
		return
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(conn), "Get", func(_ *zk.Conn, path string) ([]byte, *zk.Stat, error) {
		return []byte("test"), &zk.Stat{}, nil
	})

	s, m := z.Probe()
	assert.True(t, s)
	assert.Contains(t, m, "Successfully")

	z.Data = map[string]string{
		"test": "test1",
	}
	s, m = z.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "Data not match")

	monkey.PatchInstanceMethod(reflect.TypeOf(conn), "Get", func(_ *zk.Conn, path string) ([]byte, *zk.Stat, error) {
		return []byte(""), &zk.Stat{}, fmt.Errorf("get error")
	})
	s, m = z.Probe()
	assert.False(t, s)
	assert.Contains(t, m, "get error")

	monkey.UnpatchAll()

}
