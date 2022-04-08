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

package tcp

import (
	"fmt"
	"net"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/base"
	log "github.com/sirupsen/logrus"
)

// TCP implements a config for TCP
type TCP struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`

	base.DefaultOptions `yaml:",inline"`
}

// Config HTTP Config Object
func (t *TCP) Config(gConf global.ProbeSettings) error {
	t.ProbeKind = "tcp"
	t.DefaultOptions.Config(gConf, t.Name, t.Host)

	log.Debugf("[%s] configuration: %+v, %+v", t.Kind(), t, t.Result())
	return nil
}

// Probe return the checking result
func (t *TCP) Probe() probe.Result {

	now := time.Now()
	t.ProbeResult.StartTime = now
	t.ProbeResult.StartTimestamp = now.UnixMilli()

	conn, err := net.DialTimeout("tcp", t.Host, t.Timeout())
	t.ProbeResult.RoundTripTime.Duration = time.Since(now)
	status := probe.StatusUp
	if err != nil {
		t.ProbeResult.Message = fmt.Sprintf("Error: %v", err)
		log.Errorf("error: %v", err)
		status = probe.StatusDown
	} else {
		t.ProbeResult.Message = "TCP Connection Established Successfully!"
		conn.Close()
	}
	t.ProbeResult.PreStatus = t.ProbeResult.Status
	t.ProbeResult.Status = status

	t.ProbeResult.DoStat(t.Interval())

	return *t.ProbeResult
}
