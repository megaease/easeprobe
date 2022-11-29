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
				Disks: []string{"/", "/data"},
			},
		},
	}
}

var hostInfo = `t01
Ubuntu
4
  71.6 us,  1.7 sy,  0.2 ni, 26.8 id,  0.3 wa,  0.4 hi,  0.5 si,  0.6 st
4407 15718 28.04
58 97 60% /
20 80 20% /data
4
0.00 0.03 0.10
`

func TestHostInfo(t *testing.T) {
	host := newHost(t)
	host.Servers[0].Config(global.ProbeSettings{})
	info, err := host.Servers[0].ParseHostInfo(hostInfo)
	assert.Nil(t, err)
	assert.Equal(t, "t01", info.HostName)
	assert.Equal(t, "Ubuntu", info.OS)
	assert.Equal(t, int64(4), info.Core)
	assert.Equal(t, "28.04", fmt.Sprintf("%.2f", info.Memory.Usage))
	assert.Equal(t, "71.60", fmt.Sprintf("%.2f", info.CPU.User))
	assert.Equal(t, "1.70", fmt.Sprintf("%.2f", info.CPU.Sys))
	assert.Equal(t, "0.20", fmt.Sprintf("%.2f", info.CPU.Nice))
	assert.Equal(t, "26.80", fmt.Sprintf("%.2f", info.CPU.Idle))
	assert.Equal(t, "0.30", fmt.Sprintf("%.2f", info.CPU.Wait))
	assert.Equal(t, "0.40", fmt.Sprintf("%.2f", info.CPU.Hard))
	assert.Equal(t, "0.50", fmt.Sprintf("%.2f", info.CPU.Soft))
	assert.Equal(t, "0.60", fmt.Sprintf("%.2f", info.CPU.Steal))
	assert.Equal(t, "0.00", fmt.Sprintf("%.2f", info.Load.Metrics["m1"]))
	assert.Equal(t, "0.03", fmt.Sprintf("%.2f", info.Load.Metrics["m5"]))
	assert.Equal(t, "0.10", fmt.Sprintf("%.2f", info.Load.Metrics["m15"]))
	assert.Equal(t, 4407, info.Memory.Used)
	assert.Equal(t, 15718, info.Memory.Total)
	assert.Equal(t, 58, info.Disks.Usage[0].Used)
	assert.Equal(t, 97, info.Disks.Usage[0].Total)
	assert.Equal(t, "60.00", fmt.Sprintf("%.2f", info.Disks.Usage[0].Usage))
	assert.Equal(t, "/", info.Disks.Usage[0].Tag)
	assert.Equal(t, 20, info.Disks.Usage[1].Used)
	assert.Equal(t, 80, info.Disks.Usage[1].Total)
	assert.Equal(t, "20.00", fmt.Sprintf("%.2f", info.Disks.Usage[1].Usage))
	assert.Equal(t, "/data", info.Disks.Usage[1].Tag)
}

func TestHost(t *testing.T) {
	host := newHost(t)
	for i := 0; i < len(host.Servers); i++ {
		server := &host.Servers[i]
		server.Config(global.ProbeSettings{})
		assert.Equal(t, "host", server.ProbeKind)
		assert.Equal(t, "server", server.ProbeTag)
	}

	localHostInfo := hostInfo
	var s *ssh.Server
	monkey.PatchInstanceMethod(reflect.TypeOf(s), "RunSSHCmd", func(_ *ssh.Server) (string, error) {
		return localHostInfo, nil
	})

	server := &host.Servers[0]
	server.Config(global.ProbeSettings{})
	status, message := server.DoProbe()
	assert.True(t, status)
	assert.Contains(t, message, "Fine")

	server.Threshold.CPU = 0.5
	server.Config(global.ProbeSettings{})
	status, message = server.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "CPU threshold alert!")

	server.Threshold.Mem = 0.2
	server.Config(global.ProbeSettings{})
	status, message = server.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "Memory threshold alert!")

	server.Threshold.Disk = 0.2
	server.Config(global.ProbeSettings{})
	status, message = server.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "Disk Space threshold alert!")

	// default disk test
	server.Disks = []string{}
	server.Config(global.ProbeSettings{})
	assert.Equal(t, 1, len(server.Disks))
	assert.Equal(t, "/", server.Disks[0])
	assert.Equal(t, 1, len(server.info.Disks.Mount))
	assert.Equal(t, "/", server.info.Disks.Mount[0])

	// invalid disk format
	localHostInfo = `t01
	Ubuntu
	4
	  71.6 us,  1.7 sy,  0.2 ni, 26.8 id,  0.3 wa,  0.4 hi,  0.5 si,  0.6 st
	4407 15718 28.04
	0.00 0.03 0.10
	58
	4
	0.00 0.03 0.10`
	status, message = server.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "invalid disk output")

	// invalid load average format
	localHostInfo = `t01
	Ubuntu
	4
	  71.6 us,  1.7 sy,  0.2 ni, 26.8 id,  0.3 wa,  0.4 hi,  0.5 si,  0.6 st
	4407 15718 28.04
	58 97 60% /
	20 80 20% /data
	4
	0.00 0.03`
	status, message = server.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "invalid load average output")

	// invalid memory format
	localHostInfo = `t01
	Ubuntu
	4
	71.6 us,  1.7 sy,  0.2 ni, 26.8 id,  0.3 wa,  0.4 hi,  0.5 si,  0.6 st
	4407 15718
	0.00 0.03 0.10
	58 97 60%
	4
	0.00 0.03 0.10`
	status, message = server.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "invalid memory output")

	// invalid cpu format
	localHostInfo = `t01
	Ubuntu
	4
	71.6 us,  1.7 sy,  0.2 ni, 26.8 id
	4407 15718 28.04
	0.00 0.03 0.10
	58 97 60%
	4
	0.00 0.03 0.10`
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

	monkey.UnpatchAll()
}

func TestLoad(t *testing.T) {
	host := newHost(t)
	server := &host.Servers[0]
	server.Config(global.ProbeSettings{})
	assert.Equal(t, DefaultLoadThreshold, server.Threshold.Load["m1"])
	assert.Equal(t, DefaultLoadThreshold, server.Threshold.Load["m5"])
	assert.Equal(t, DefaultLoadThreshold, server.Threshold.Load["m15"])

	host = newHost(t)
	server = &host.Servers[0]
	server.Threshold.Load = map[string]float64{
		"xxx": 0.1,
	}
	server.Config(global.ProbeSettings{})
	assert.Equal(t, DefaultLoadThreshold, server.Threshold.Load["m1"])
	assert.Equal(t, DefaultLoadThreshold, server.Threshold.Load["m5"])
	assert.Equal(t, DefaultLoadThreshold, server.Threshold.Load["m15"])

	host = newHost(t)
	server = &host.Servers[0]
	server.Threshold.Load = map[string]float64{
		"M1":  0.1,
		"m5":  0.2,
		"M15": 0.3,
	}
	server.Config(global.ProbeSettings{})
	assert.Equal(t, 0.1, server.Threshold.Load["m1"])
	assert.Equal(t, 0.2, server.Threshold.Load["m5"])
	assert.Equal(t, 0.3, server.Threshold.Load["m15"])

	localHostInfo := hostInfo
	var s *ssh.Server
	monkey.PatchInstanceMethod(reflect.TypeOf(s), "RunSSHCmd", func(_ *ssh.Server) (string, error) {
		return localHostInfo, nil
	})

	status, message := server.DoProbe()
	assert.True(t, status)
	assert.Contains(t, message, "Fine")

	localHostInfo = `t01
	Ubuntu
	4
	71.6 us,  1.7 sy,  0.2 ni, 26.8 id,  0.3 wa,  0.4 hi,  0.5 si,  0.6 st
	4407 15718 28.04
	58 97 60% /
	20 80 20% /data
	4
	0.4 0.03 0.10`

	status, message = server.DoProbe()
	assert.False(t, status)
	assert.Contains(t, message, "Load Average threshold m1 alert!")

	monkey.UnpatchAll()
}

func TestBadParse(t *testing.T) {
	info := Info{}
	err := info.Basic.Parse([]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid basic output")

	err = info.CPU.Parse([]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cpu output")

	err = info.Memory.Parse([]string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid memory output")

	info.Disks.Mount = []string{"/", "/data"}
	err = info.Disks.Parse([]string{"58 97 60% /"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid disk output")

	err = info.Load.Parse([]string{"0.4 0.03 0.10"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid load average output")
}
