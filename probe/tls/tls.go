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

type TLS struct {
	base.DefaultOptions `yaml:",inline"`
	Host                string `yaml:"host"`
	InsecureSkipVerify  bool   `yaml:"insecure_skip_verify"`

	RootCAPemPath string `yaml:"root_ca_pem_path"`
	RootCaPem     string `yaml:"root_ca_pem"`
	rootCAs       *x509.CertPool

	ExpireSkipVerify bool `yaml:"expire_skip_verify"`

	metrics *metrics
}

// Config HTTP Config Object
func (t *TLS) Config(gConf global.ProbeSettings) error {
	kind := "tls"
	tag := ""
	name := t.ProbeName
	t.DefaultOptions.Config(gConf, kind, tag, name, t.Host, t.DoProbe)

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

	log.Debugf("[%s] configuration: %+v, %+v", t.ProbeKind, t, t.Result())
	return nil
}

// DoProbe return the checking result
func (t *TLS) DoProbe() (bool, string) {
	addr := t.Host
	conn, err := net.DialTimeout("tcp", addr, t.Timeout())
	if err != nil {
		log.Errorf("tcp dial error: %v", err)
		return false, fmt.Sprintf("tcp dial error: %v", err)
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
		log.Errorf("tls handshake error: %v", err)
		return false, fmt.Sprintf("tls handshake error: %v", err)
	}

	if !t.ExpireSkipVerify {
		for _, cert := range tconn.ConnectionState().PeerCertificates {
			valid := true
			valid = valid && !time.Now().After(cert.NotAfter)
			valid = valid && !time.Now().Before(cert.NotBefore)

			if !valid {
				log.Errorf("host %v cert expired", t.Host)
				return false, "certificate is expired or not yet valid"
			}
		}
	}

	state := tconn.ConnectionState()

	t.metrics.EarliestCertExpiry.With(prometheus.Labels{}).Set(float64(getEarliestCertExpiry(&state).Unix()))
	t.metrics.LastChainExpiryTimestampSeconds.With(prometheus.Labels{}).Set(float64(getLastChainExpiry(&state).Unix()))

	return true, "TLS Endpoint Verified Successfully!"
}
