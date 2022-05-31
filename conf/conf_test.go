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

package conf

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/megaease/easeprobe/probe/client"
	clientConf "github.com/megaease/easeprobe/probe/client/conf"
	"github.com/megaease/easeprobe/probe/host"
	httpProbe "github.com/megaease/easeprobe/probe/http"
	"github.com/megaease/easeprobe/probe/shell"
	"github.com/megaease/easeprobe/probe/ssh"
	"github.com/megaease/easeprobe/probe/tcp"
	"github.com/stretchr/testify/assert"
)

func testisExternalURL(url string, expects bool, t *testing.T) {
	if got := isExternalURL(url); got != expects {
		t.Errorf("isExternalURL(\"%s\") = %v, expected %v", url, got, expects)
	}
}

func TestPathAndURL(t *testing.T) {
	testisExternalURL("/tmp", false, t)
	testisExternalURL("//tmp", false, t)
	testisExternalURL("file:///tmp", false, t)
	testisExternalURL("http://", false, t)
	testisExternalURL("https://", false, t)
	testisExternalURL("hTtP://", false, t)
	testisExternalURL("http", false, t)
	testisExternalURL("https", false, t)
	testisExternalURL("ftp", false, t)
	testisExternalURL("hTtP://127.0.0.1", true, t)
	testisExternalURL("localhost", false, t)
	testisExternalURL("ftp://127.0.0.1", false, t)
}

func TestGetYamlFileFromFile(t *testing.T) {
	if _, err := getYamlFileFromFile("/tmp/nonexistent"); err == nil {
		t.Errorf("getYamlFileFromFile(\"/tmp/nonexistent\") = nil, expected error")
	}

	tmpfile, err := ioutil.TempFile("", "invalid*.yaml")
	if err != nil {
		t.Errorf("TempFile(\"invalid*.yaml\") %v", err)
	}

	defer os.Remove(tmpfile.Name()) // clean up

	// test empty file
	data, err := getYamlFileFromFile(tmpfile.Name())
	if err != nil {
		t.Errorf("getYamlFileFromFile(\"%s\") = %v, expected nil", tmpfile.Name(), err)
	}

	//confirm we read empty data
	if string(data) != "" {
		t.Errorf("getYamlFileFromFile(\"%s\") got data %s, expected nil", tmpfile.Name(), data)
	}
}

const confVer = "version: 0.1.0\n"

const confHTTP = `
http:
  - name: dummy
    url: http://localhost:12345/dummy
    channels:
      - "telegram#Dev"
  - name: Local mTLS test
    url: https://localhost:8443/hello
    ca: ../mTLS/certs/ca.crt
    cert: ../mTLS/certs/client.b.crt
    key: ../mTLS/certs/client.b.key
  - name: MegaCloud
    url: https://cloud.megaease.cn/
    timeout: 2m
    interval: 30s
    channels:
      - "telegram#Dev"
`

func checkHTTPProbe(t *testing.T, probe httpProbe.HTTP) {
	switch probe.ProbeName {
	case "dummy":
		assert.Equal(t, probe.URL, "http://localhost:12345/dummy")
		assert.Equal(t, probe.Channels(), []string{"telegram#Dev"})
	case "Local mTLS test":
		assert.Equal(t, probe.URL, "https://localhost:8443/hello")
		assert.Equal(t, probe.CA, "../mTLS/certs/ca.crt")
		assert.Equal(t, probe.Cert, "../mTLS/certs/client.b.crt")
		assert.Equal(t, probe.Key, "../mTLS/certs/client.b.key")
	case "MegaCloud":
		assert.Equal(t, probe.URL, "https://cloud.megaease.cn/")
		assert.Equal(t, probe.ProbeTimeout, 2*time.Minute)
		assert.Equal(t, probe.ProbeTimeInterval, 30*time.Second)
		assert.Equal(t, probe.Channels(), []string{"telegram#Dev"})
	default:
		t.Errorf("unexpected probe name %s", probe.ProbeName)
	}
}

const confTCP = `
tcp:
  - name: Example SSH
    host: example.com:22
    timeout: 10s
    interval: 2m
  - name: Example HTTP
    host: example.com:80
`

func checkTCPProbe(t *testing.T, probe tcp.TCP) {
	switch probe.ProbeName {
	case "Example SSH":
		assert.Equal(t, probe.Host, "example.com:22")
		assert.Equal(t, probe.ProbeTimeout, 10*time.Second)
		assert.Equal(t, probe.ProbeTimeInterval, 2*time.Minute)
	case "Example HTTP":
		assert.Equal(t, probe.Host, "example.com:80")
	default:
		t.Errorf("unexpected probe name %s", probe.ProbeName)
	}
}

const confShell = `
shell:
  - name: Google Service
    cmd: "./resources/scripts/probe/proxy.curl.sh"
    args:
      - "socks5://127.0.0.1:1085"
      - "www.google.com"
    timeout: 20s
    interval: 1m
    channels:
        - "telegram#Dev"
  - name: Redis (Local)
    cmd: "redis-cli"
    args:
      - "-h"
      - "127.0.0.1"
      - "-p"
      - 6379
      - "ping"
    env:
      - "REDISCLI_AUTH=abc123"
    contain: "PONG"
	channels:
        - "slack"
`

func checkShellProbe(t *testing.T, probe shell.Shell) {
	switch probe.ProbeName {
	case "Google Service":
		assert.Equal(t, probe.Command, "./resources/scripts/probe/proxy.curl.sh")
		assert.Equal(t, probe.Args, []string{"socks5://127.0.0.1:1085", "www.google.com"})
		assert.Equal(t, probe.ProbeTimeout, 20*time.Second)
		assert.Equal(t, probe.ProbeTimeInterval, 1*time.Minute)
		assert.Equal(t, probe.Channels(), []string{"telegram#Dev"})
	case "Redis (Local)":
		assert.Equal(t, probe.Command, "redis-cli")
		assert.Equal(t, probe.Args, []string{"-h", "127.0.0.1", "-p", "6379", "ping"})
		assert.Equal(t, probe.Env, []string{"REDISCLI_AUTH=abc123"})
		assert.Equal(t, probe.Contain, "PONG")
		assert.Equal(t, probe.Channels(), []string{"slack"})
	default:
		t.Errorf("unexpected probe name %s", probe.ProbeName)
	}
}

const confSSH = `
ssh:
  bastion:
    aws:
      host: ubuntu@one.aws.server
      #username: ubuntu
      key: /home/chenhao/.ssh/pem/my.pem
    gcp:
      host: one.gcp.server:22
      username: ubuntu
      key: /home/chenhao/.ssh/pem/my.pem

  servers:
    - name: AWS Server
      bastion: aws
      host: 172.20.2.202:22
      username: ubuntu
      password: xxxx
      key: /home/chenhao/.ssh/pem/my.pem
      cmd: "env;hostname;ifconfig"
      env:
        - "EASEPROBE=1"

    - name: GCP Server
      host: 10.1.1.1:22
      bastion: gcp
      key: /home/chenhao/.ssh/pem/my.pem
      cmd: "env;hostname;ifconfig"
      env:
        - "EASEPROBE=1"
      contain: EASEPROBE
`

func checkSSHProbe(t *testing.T, probe ssh.SSH) {
	m := probe.Bastion
	aws := (*m)["aws"]
	assert.Equal(t, aws.Host, "one.aws.server:22")
	assert.Equal(t, aws.User, "ubuntu")
	assert.Equal(t, aws.PrivateKey, "/home/chenhao/.ssh/pem/my.pem")
	gcp := (*m)["gcp"]
	assert.Equal(t, gcp.Host, "one.gcp.server:22")
	assert.Equal(t, gcp.User, "ubuntu")
	assert.Equal(t, gcp.PrivateKey, "/home/chenhao/.ssh/pem/my.pem")

	for _, server := range probe.Servers {
		checkSSHServerProbe(t, server)
	}
}

func checkSSHServerProbe(t *testing.T, probe ssh.Server) {
	switch probe.ProbeName {
	case "AWS Server":
		assert.Equal(t, probe.Host, "172.20.2.202:22")
		assert.Equal(t, probe.BastionID, "aws")
		assert.Equal(t, probe.User, "ubuntu")
		assert.Equal(t, probe.Password, "xxxx")
		assert.Equal(t, probe.PrivateKey, "/home/chenhao/.ssh/pem/my.pem")
		assert.Equal(t, probe.Command, "env;hostname;ifconfig")
		assert.Equal(t, probe.Env, []string{"EASEPROBE=1"})
	case "GCP Server":
		assert.Equal(t, probe.Host, "10.1.1.1:22")
		assert.Equal(t, probe.BastionID, "gcp")
		assert.Equal(t, probe.PrivateKey, "/home/chenhao/.ssh/pem/my.pem")
		assert.Equal(t, probe.Command, "env;hostname;ifconfig")
		assert.Equal(t, probe.Env, []string{"EASEPROBE=1"})
		assert.Equal(t, probe.Contain, "EASEPROBE")
	default:
		t.Errorf("unexpected probe name %s", probe.ProbeName)
	}
}

const confHost = `
host:
  bastion:
    aws:
      host: ubuntu@one.aws.server
      key: /home/chenhao/.ssh/pem/my.pem

  servers:
    - name: AWS Server
      bastion: aws
      host: ubuntu@172.20.2.125
      key: /home/chenhao/.ssh/pem/my.pem
      threshold:
        cpu: 0.75
        mem: 0.70
        disk: 0.90
      channels:
        - general
        - test
`

func checkHostProbe(t *testing.T, probe host.Host) {
	m := probe.Bastion
	aws := (*m)["aws"]
	assert.Equal(t, aws.Host, "one.aws.server:22")
	assert.Equal(t, aws.User, "ubuntu")
	assert.Equal(t, aws.PrivateKey, "/home/chenhao/.ssh/pem/my.pem")
	for _, server := range probe.Servers {
		checkHostServerProbe(t, server)
	}
}

func checkHostServerProbe(t *testing.T, probe host.Server) {
	switch probe.ProbeName {
	case "AWS Server":
		assert.Equal(t, probe.Host, "ubuntu@172.20.2.125")
		assert.Equal(t, probe.BastionID, "aws")
		assert.Equal(t, probe.PrivateKey, "/home/chenhao/.ssh/pem/my.pem")
		assert.Equal(t, probe.Threshold.CPU, 0.75)
		assert.Equal(t, probe.Threshold.Mem, 0.70)
		assert.Equal(t, probe.Threshold.Disk, 0.90)
		assert.Equal(t, probe.Channels(), []string{"general", "test"})
	default:
		t.Errorf("unexpected probe name %s", probe.ProbeName)
	}
}

const confClient = `
client:
  - name: Redis Native Client (local)
    driver: "redis"
    host: "localhost:6379"
    password: "abc123"
    channels:
      - test
  - name: MySQL Native Client (local)
    driver: "mysql"
    host: "localhost:3306"
    username: "root"
    password: "pass"
    ca: /home/chenhao/Github/mTLS/certs/ca.crt
    cert: /home/chenhao/Github/mTLS/certs/client.b.crt
    key: /home/chenhao/Github/mTLS/certs/client.b.key
`

func checkClientProbe(t *testing.T, probe client.Client) {
	switch probe.ProbeName {
	case "Redis Native Client (local)":
		assert.Equal(t, probe.DriverType, clientConf.Redis)
		assert.Equal(t, probe.Host, "localhost:6379")
		assert.Equal(t, probe.Password, "abc123")
		assert.Equal(t, probe.Channels(), []string{"test"})
	case "MySQL Native Client (local)":
		assert.Equal(t, probe.DriverType, clientConf.MySQL)
		assert.Equal(t, probe.Host, "localhost:3306")
		assert.Equal(t, probe.Username, "root")
		assert.Equal(t, probe.Password, "pass")
		assert.Equal(t, probe.CA, "/home/chenhao/Github/mTLS/certs/ca.crt")
		assert.Equal(t, probe.Cert, "/home/chenhao/Github/mTLS/certs/client.b.crt")
		assert.Equal(t, probe.Key, "/home/chenhao/Github/mTLS/certs/client.b.key")
	default:
		t.Errorf("unexpected probe name %s", probe.ProbeName)
	}
}

const confNotify = `
notify:
  slack:
    - name: "Slack"
      webhook: "https://hooks.slack.com/services/xxx"
      channels:
        - "slack"
        - "general"
      #dry: true
  discord:
    - name: "Discord"
      webhook: "https://discord.com/api/webhooks/xxx"
      avatar: "https://img.icons8.com/ios/72/appointment-reminders--v1.png"
      thumbnail: "https://freeiconshop.com/wp-content/uploads/edd/notification-flat.png"
      dry: true
      retry:
        times: 3
        interval: 10s
    - name: "Test"
      webhook: "https://discord.com/api/webhooks/xxx"
      avatar: "https://img.icons8.com/ios/72/appointment-reminders--v1.png"
      thumbnail: "https://freeiconshop.com/wp-content/uploads/edd/notification-flat.png"
      retry:
        times: 3
        interval: 10s
      channels:
        - "general"
  telegram:
    - name: "Dev Channel"
      token: 123456:xxxxx
      chat_id: -1001343458903
      channels:
        - "telegram#Dev"
        - "test"
    - name: "Ops Group"
      token: 123456:yyyyyy
      #chat_id: -234234
      chat_id: -12310934
      dry: true
      retry:
        times: 3
        interval: 8s
  email:
    - name: "email"
      server: smtp.email.com:465
      username: noreply@megaease.com
      password: xxx
      to: "test@email.com"
      dry: true
`

const confSettings = `
settings:
  name: "EaseProbeBot"
  icon: https://upload.wikimedia.org/wikipedia/commons/2/2d/Etcher-icon.png
  http:
    ip: 127.0.0.1
    port: 8181
    refresh: 5s
    log:
      file: /tmp/log/easeprobe.access.log
      self_rotate: true
  sla:
    schedule: "weekly"
    time: "23:59"
    debug: true
    backups: 20
    channels:
      - general
  notify:
    dry: true
    retry:
      times: 5
      interval: 10s
  probe:
    interval: 15s
  log:
    level: debug
    size: 1
  #backups: 10
  debug: true
  timeformat: "2006-01-02 15:04:05 UTC"
`

const confYAML = confVer + confHTTP + confSSH + confHost + confClient + confNotify + confSettings

func writeConfig(file string) error {
	return ioutil.WriteFile(file, []byte(confYAML), 0644)
}

func httpServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(confYAML))
	})
	http.ListenAndServe(":8080", nil)
}

func TestConfig(t *testing.T) {
	file := "./config.yaml"
	err := writeConfig(file)
	assert.Nil(t, err)

	conf, err := New(&file)
	assert.Nil(t, err)

	assert.Equal(t, "EaseProbeBot", conf.Settings.Name)
	assert.Equal(t, "0.1.0", conf.Version)

	for _, v := range conf.HTTP {
		checkHTTPProbe(t, v)
	}
	for _, v := range conf.TCP {
		checkTCPProbe(t, v)
	}
	for _, v := range conf.Shell {
		checkShellProbe(t, v)
	}
	for _, v := range conf.Client {
		checkClientProbe(t, v)
	}
	checkSSHProbe(t, conf.SSH)
	checkHostProbe(t, conf.Host)

	conf.InitAllLogs()
	probers := conf.AllProbers()
	assert.Equal(t, 8, len(probers))
	notifiers := conf.AllNotifiers()
	assert.Equal(t, 6, len(notifiers))

	go httpServer()
	url := "http://localhost:8080"
	httpConf, err := New(&url)
	assert.Nil(t, err)
	assert.Equal(t, "EaseProbeBot", httpConf.Settings.Name)
	assert.Equal(t, "0.1.0", httpConf.Version)

	probers = conf.AllProbers()
	assert.Equal(t, 8, len(probers))
	notifiers = conf.AllNotifiers()
	assert.Equal(t, 6, len(notifiers))

	os.RemoveAll(file)
	os.RemoveAll("data")
}
