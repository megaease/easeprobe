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

// Package tls is the tls probe package
package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe/base"
	"github.com/prometheus/client_golang/prometheus"

	log "github.com/sirupsen/logrus"
)

// TLS implements a config for TLS
type TLS struct {
	base.DefaultProbe  `yaml:",inline"`
	Host               string `yaml:"host" json:"host" jsonschema:"required,format=hostname,title=Host,description=The host to probe"`
	Proxy              string `yaml:"proxy" json:"proxy,omitempty" jsonschema:"format=hostname,title=Proxy,description=The proxy to use for the TLS connection"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify" json:"insecure_skip_verify,omitempty" jsonschema:"title=Insecure Skip Verify,description=Whether to skip verifying the certificate chain and host name"`

	RootCAPemPath string         `yaml:"root_ca_pem_path" json:"root_ca_pem_path,omitempty" jsonschema:"title=Root CA PEM Path,description=The path to the root CA PEM file"`
	RootCaPem     string         `yaml:"root_ca_pem" json:"root_ca_pem,omitempty" jsonschema:"title=Root CA PEM,description=The root CA PEM"`
	rootCAs       *x509.CertPool `yaml:"-" json:"-"`

	ExpireSkipVerify  bool          `yaml:"expire_skip_verify" json:"expire_skip_verify,omitempty" jsonschema:"title=Expire Skip Verify,description=Whether to skip verifying the certificate expire time"`
	AlertExpireBefore time.Duration `yaml:"alert_expire_before" json:"alert_expire_before,omitempty" jsonschema:"title=Alert Expire Before,description=The alert expire before time"`

	metrics *metrics
}

// Config HTTP Config Object
func (t *TLS) Config(gConf global.ProbeSettings) error {
	kind := "tls"
	tag := ""
	name := t.ProbeName
	t.DefaultProbe.Config(gConf, kind, tag, name, t.Host, t.DoProbe)

	rootCaPem := []byte(t.RootCaPem)

	if len(rootCaPem) == 0 && t.RootCAPemPath != "" {
		var err error
		rootCaPem, err = ioutil.ReadFile(t.RootCAPemPath)
		if err != nil {
			return err
		}
	}

	if len(rootCaPem) > 0 {
		t.rootCAs = x509.NewCertPool()
		if !(t.rootCAs.AppendCertsFromPEM(rootCaPem)) {
			return fmt.Errorf("cannot parse root ca pem")
		}
	}

	t.metrics = newMetrics(kind, tag)

	log.Debugf("[%s / %s] configuration: %+v", t.ProbeKind, t.ProbeName, *t)
	return nil
}

// DoProbe return the checking result
func (t *TLS) DoProbe() (bool, string) {
	addr := t.Host
	conn, err := t.GetProxyConnection(t.Proxy, addr)
	if err != nil {
		log.Errorf("[%s / %s] tcp dial error: %v", t.ProbeKind, t.ProbeName, err)
		return false, fmt.Sprintf("tcp dial error: %v", err)
	}
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetLinger(0)
	}
	defer conn.Close()

	colonPos := strings.LastIndex(addr, ":")
	if colonPos == -1 {
		colonPos = len(addr)
	}
	hostname := addr[:colonPos]

	tconn := tls.Client(conn, &tls.Config{
		InsecureSkipVerify: t.InsecureSkipVerify,
		RootCAs:            t.rootCAs,
		ServerName:         hostname,
	})

	ctx, cancel := context.WithTimeout(context.Background(), t.Timeout())
	defer cancel()
	err = tconn.HandshakeContext(ctx)
	if err != nil {
		log.Errorf("[%s / %s] tls handshake error: %v", t.ProbeKind, t.ProbeName, err)
		return false, fmt.Sprintf("tls handshake error: %v", err)
	}

	if !t.ExpireSkipVerify {
		for _, cert := range tconn.ConnectionState().PeerCertificates {
			valid := true
			valid = valid && !time.Now().After(cert.NotAfter)
			valid = valid && !time.Now().Before(cert.NotBefore)

			if !valid {
				log.Errorf("[%s / %s] host %v cert expired", t.ProbeKind, t.ProbeName, t.Host)
				return false, "certificate is expired or not yet valid"
			}

			if t.AlertExpireBefore > 0 {
				durLeft := time.Until(cert.NotAfter)
				if durLeft < t.AlertExpireBefore {
					return false, fmt.Sprintf("certificate is expiring in %v", durLeft)
				}
			}
		}
	}

	state := tconn.ConnectionState()

	t.metrics.EarliestCertExpiry.With(prometheus.Labels{}).Set(float64(getEarliestCertExpiry(&state).Unix()))
	t.metrics.LastChainExpiryTimestampSeconds.With(prometheus.Labels{}).Set(float64(getLastChainExpiry(&state).Unix()))

	return true, "TLS Endpoint Verified Successfully!"
}
