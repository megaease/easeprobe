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
	"errors"
	"net"
	"net/http/httptrace"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testTimeStart(t *testing.T, start *time.Time, fn func()) {
	var newTime time.Time
	assert.Equal(t, newTime, *start)
	fn()
	assert.Less(t, time.Since(*start), time.Minute)
}

func testTimeDone(t *testing.T, took *time.Duration, fns ...func()) {
	assert.Equal(t, time.Duration(0), *took)
	for _, fn := range fns {
		fn()
		assert.Less(t, *took, time.Minute)
	}
}

func TestTraceStart(t *testing.T) {
	// Create a new trace
	s := NewTraceStats("http", "tag", "test")
	assert.Equal(t, "http", s.kind)
	assert.Equal(t, "test", s.name)
	assert.Equal(t, "tag", s.tag)

	testTimeStart(t, &s.totalStartAt, func() { s.getConn("8080") })
	testTimeStart(t, &s.dnsStartAt, func() { s.dnsStart(httptrace.DNSStartInfo{}) })
	testTimeStart(t, &s.connStartAt, func() { s.connectStart("tcp", "8080") })
	testTimeStart(t, &s.sendStartAt, func() { s.wroteHeaderField("key", []string{"value"}) })
	testTimeStart(t, &s.tlsStartAt, func() { s.tlsStart() })
	testTimeStart(t, &s.transferStartAt, func() { s.gotFirstResponseByte() })
	testTimeStart(t, &s.waitStartAt, func() { s.wroteRequest(httptrace.WroteRequestInfo{}) })
}

func TestTraceDone(t *testing.T) {
	s := NewTraceStats("http", "tag", "test")

	s.connStartAt = time.Now()
	s.dnsStartAt = time.Now()
	s.sendStartAt = time.Now()
	s.tlsStartAt = time.Now()
	s.transferStartAt = time.Now()
	s.waitStartAt = time.Now()
	s.totalStartAt = time.Now()

	// Test DNS
	testTimeDone(t, &s.dnsTook,
		func() {
			s.dnsDone(httptrace.DNSDoneInfo{
				Addrs: []net.IPAddr{
					{IP: net.IPv4(1, 2, 3, 4)},
				}})
		},
		func() {
			s.dnsDone(httptrace.DNSDoneInfo{Err: &net.DNSError{}})
		},
	)

	// Test connection
	testTimeDone(t, &s.connTook,
		func() {
			s.connectDone("tcp", "8080", nil)
		},
		func() {
			s.connectDone("tcp", "8080", &net.OpError{})
		},
	)

	// Test TLS
	testTimeDone(t, &s.tlsTook,
		func() {
			s.tlsDone(tls.ConnectionState{}, nil)
		},
		func() {
			s.tlsDone(tls.ConnectionState{}, errors.New("test error"))
		},
	)

	// Test send
	testTimeDone(t, &s.sendTook,
		func() {
			s.wroteHeaders()
		},
	)

	// Test Wait
	testTimeDone(t, &s.waitTook,
		func() {
			s.wroteRequest(httptrace.WroteRequestInfo{})
		},
		func() {
			s.wroteRequest(httptrace.WroteRequestInfo{Err: errors.New("test error")})
		},
	)

	// Test transfer
	testTimeDone(t, &s.transferTook,
		func() {
			s.putIdleConn(nil)
		},
		func() {
			s.putIdleConn(errors.New("test error"))
		},
	)

	took := s.connTook
	s.gotConn(httptrace.GotConnInfo{})
	assert.Equal(t, took, s.connTook)

}
