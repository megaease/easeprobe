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

package host

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/megaease/easeprobe/probe/ssh"
	"github.com/stretchr/testify/assert"
)

func newHost(t *testing.T) *Host {
	return &Host{
		Servers: []Server{
			{
				Server: ssh.Server{
					DefaultProbe: base.DefaultProbe{ProbeName: "dummy host"},
					Endpoint:     ssh.Endpoint{Host: "server:22"},
				},
			},
		},
	}
}

var hostInfo = `t01
Ubuntu
4407 15718 28.04
4
  71.6 us,  1.7 sy,  0.2 ni, 26.8 id,  0.3 wa,  0.4 hi,  0.5 si,  0.6 st
58 97 60% /
20 80 20% /data`

func TestHostInfo(t *testing.T) {
	host := newHost(t)
	info, err := host.Servers[0].ParseHostInfo(hostInfo)
	assert.Nil(t, err)
	assert.Equal(t, "t01", info.HostName)
	assert.Equal(t, "Ubuntu", info.OS)
	assert.Equal(t, 4407, info.Memory.Used)
	assert.Equal(t, 15718, info.Memory.Total)
	assert.Equal(t, "28.04", fmt.Sprintf("%.2f", info.Memory.Usage))
	assert.Equal(t, int64(4), info.Core)
	assert.Equal(t, "71.60", fmt.Sprintf("%.2f", info.CPU.User))
	assert.Equal(t, "1.70", fmt.Sprintf("%.2f", info.CPU.Sys))
	assert.Equal(t, "0.20", fmt.Sprintf("%.2f", info.CPU.Nice))
	assert.Equal(t, "26.80", fmt.Sprintf("%.2f", info.CPU.Idle))
	assert.Equal(t, "0.30", fmt.Sprintf("%.2f", info.CPU.Wait))
	assert.Equal(t, "0.40", fmt.Sprintf("%.2f", info.CPU.Hard))
	assert.Equal(t, "0.50", fmt.Sprintf("%.2f", info.CPU.Soft))
	assert.Equal(t, "0.60", fmt.Sprintf("%.2f", info.CPU.Steal))
	assert.Equal(t, 58, info.Disks[0].Used)
	assert.Equal(t, 97, info.Disks[0].Total)
	assert.Equal(t, "60.00", fmt.Sprintf("%.2f", info.Disks[0].Usage))
	assert.Equal(t, "/", info.Disks[0].Tag)
	assert.Equal(t, 20, info.Disks[1].Used)
	assert.Equal(t, 80, info.Disks[1].Total)
	assert.Equal(t, "20.00", fmt.Sprintf("%.2f", info.Disks[1].Usage))
	assert.Equal(t, "/data", info.Disks[1].Tag)
}

func TestHost(t *testing.T) {
	host := newHost(t)
	for i := 0; i < len(host.Servers); i++ {
		server := &host.Servers[i]
		server.Config(global.ProbeSettings{})
		assert.Equal(t, "host", server.ProbeKind)
		assert.Equal(t, "server", server.ProbeTag)
	}

	var s *ssh.Server
	monkey.PatchInstanceMethod(reflect.TypeOf(s), "RunSSHCmd", func(_ *ssh.Server) (string, error) {
		return hostInfo, nil
	})

	server := &host.Servers[0]
	status, message := server.DoProbe()
	assert.True(t, status)
	assert.Contains(t, message, "Fine")

	server.Threshold.CPU = 0.5
	status, message = server.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "CPU Busy!")

	server.Threshold.Mem = 0.2
	status, message = server.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "Memory Shortage!")

	server.Threshold.Disk = 0.2
	status, message = server.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "Disk Full!")

	// invalid disk format
	hostInfo = `t01
	Ubuntu
	4407 15718 28.04
	4
	  71.6 us,  1.7 sy,  0.2 ni, 26.8 id,  0.3 wa,  0.4 hi,  0.5 si,  0.6 st
	58`
	status, message = server.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "invalid disk output")

	// invalid memory format
	hostInfo = `t01
	Ubuntu
	4407 15718
	4
	71.6 us,  1.7 sy,  0.2 ni, 26.8 id,  0.3 wa,  0.4 hi,  0.5 si,  0.6 st
	58 97 60%`
	status, message = server.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "invalid memory output")

	// invalid cpu format
	hostInfo = `t01
	Ubuntu
	4407 15718 28.04
	4
	71.6 us,  1.7 sy,  0.2 ni, 26.8 id
	58 97 60%`
	status, message = server.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "invalid cpu output")

	// bad format
	monkey.PatchInstanceMethod(reflect.TypeOf(s), "RunSSHCmd", func(_ *ssh.Server) (string, error) {
		return "", nil
	})
	status, message = server.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "invalid output")

	// run ssh failed
	monkey.PatchInstanceMethod(reflect.TypeOf(s), "RunSSHCmd", func(_ *ssh.Server) (string, error) {
		return "", errors.New("ssh error")
	})
	status, message = server.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "ssh error")

}
