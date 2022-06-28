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

package http

import (
	"crypto/tls"
	"net/http/httptrace"
	"time"

	log "github.com/sirupsen/logrus"
)

// TraceStats is the stats for a http request
type TraceStats struct {
	kind string
	tag  string
	name string

	connStartAt     time.Time
	connTook        time.Duration
	dnsStartAt      time.Time
	dnsTook         time.Duration
	sendStartAt     time.Time
	sendTook        time.Duration
	tlsStartAt      time.Time
	tlsTook         time.Duration
	totalStartAt    time.Time
	totalTook       time.Duration
	transferStartAt time.Time
	transferTook    time.Duration
	waitStartAt     time.Time
	waitTook        time.Duration

	clientTrace *httptrace.ClientTrace
}

func (s *TraceStats) getConn(hostPort string) {
	s.totalStartAt = time.Now()
}

func (s *TraceStats) dnsStart(info httptrace.DNSStartInfo) {
	s.dnsStartAt = time.Now()
}

func (s *TraceStats) dnsDone(info httptrace.DNSDoneInfo) {
	s.dnsTook = time.Since(s.dnsStartAt)
	if info.Err != nil {
		return
	}
	log.Debugf("[%s %s %s] - DNS resolve time %.3fms", s.kind, s.tag, s.name, toMS(s.dnsTook))
	for _, addr := range info.Addrs {
		log.Debugf("[%s %s %s]   %s\n", s.kind, s.tag, s.name, &addr.IP)
	}
}

func (s *TraceStats) connectStart(network, addr string) {
	s.connStartAt = time.Now()
}

func (s *TraceStats) connectDone(network, addr string, err error) {
	s.connTook = time.Since(s.connStartAt)
	if err != nil {
		return
	}
	log.Debugf("[%s %s %s] - %s connection created to %s. time: %.3fms\n",
		s.kind, s.tag, s.name, network, addr, toMS(s.connTook))
}

func (s *TraceStats) tlsStart() {
	s.tlsStartAt = time.Now()
}

func (s *TraceStats) tlsDone(cs tls.ConnectionState, err error) {
	s.tlsTook = time.Since(s.tlsStartAt)
	if err != nil {
		return
	}
	log.Debugf("[%s %s %s] - tls negotiated to %q, time: %.3fms",
		s.kind, s.tag, s.name, cs.ServerName, toMS(s.tlsTook))
}

func (s TraceStats) gotConn(info httptrace.GotConnInfo) {
	log.Debugf("[%s %s %s] - connection established. reused: %t idle: %t idle time: %dms\n",
		s.kind, s.tag, s.name, info.Reused, info.WasIdle, info.IdleTime.Milliseconds())
}

func (s *TraceStats) wroteHeaderField(key string, value []string) {
	s.sendStartAt = time.Now()
}

func (s *TraceStats) wroteHeaders() {
	s.sendTook = time.Since(s.sendStartAt)
	log.Debugf("[%s %s %s] - headers written, time: %.3fms",
		s.kind, s.tag, s.name, toMS(s.sendTook))
}

func (s *TraceStats) wroteRequest(info httptrace.WroteRequestInfo) {
	s.waitStartAt = time.Now()
	if info.Err != nil {
		return
	}
}

func (s *TraceStats) gotFirstResponseByte() {
	s.waitTook = time.Since(s.waitStartAt)
	s.transferStartAt = time.Now()
	log.Debugf("[%s %s %s] - got first response byte, time: %.3fms",
		s.kind, s.tag, s.name, toMS(s.waitTook))
}

func (s *TraceStats) putIdleConn(err error) {
	s.totalTook = time.Since(s.totalStartAt)
	s.transferTook = time.Since(s.transferStartAt)
	if err != nil {
		return
	}
	log.Debugf("[%s %s %s] - put conn idle, transfer: %.3f, total: %.3fms",
		s.kind, s.tag, s.name, toMS(s.transferTook), toMS(s.totalTook))
}

func toMS(t time.Duration) float64 {
	return float64(t.Nanoseconds()) / 1000000.0
}

// NewTraceStats returns a new traceSTats.
func NewTraceStats(kind, tag, name string) *TraceStats {
	s := &TraceStats{
		kind: kind,
		name: name,
		tag:  tag,
	}

	s.clientTrace = &httptrace.ClientTrace{
		GetConn:              s.getConn,
		DNSStart:             s.dnsStart,
		DNSDone:              s.dnsDone,
		ConnectStart:         s.connectStart,
		ConnectDone:          s.connectDone,
		TLSHandshakeStart:    s.tlsStart,
		TLSHandshakeDone:     s.tlsDone,
		GotConn:              s.gotConn,
		WroteHeaderField:     s.wroteHeaderField,
		WroteHeaders:         s.wroteHeaders,
		WroteRequest:         s.wroteRequest,
		GotFirstResponseByte: s.gotFirstResponseByte,
		PutIdleConn:          s.putIdleConn,
	}
	return s
}
