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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	httpClient "net/http"
	"os"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/notify"
	"github.com/megaease/easeprobe/notify/discord"
	"github.com/megaease/easeprobe/notify/email"
	"github.com/megaease/easeprobe/notify/slack"
	"github.com/megaease/easeprobe/notify/telegram"
	"github.com/megaease/easeprobe/probe/client"
	clientConf "github.com/megaease/easeprobe/probe/client/conf"
	"github.com/megaease/easeprobe/probe/host"
	httpProbe "github.com/megaease/easeprobe/probe/http"
	"github.com/megaease/easeprobe/probe/shell"
	"github.com/megaease/easeprobe/probe/ssh"
	"github.com/megaease/easeprobe/probe/tcp"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
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

func testScheduleYaml(t *testing.T, name string, sch Schedule, good bool) {
	var s Schedule
	err := yaml.Unmarshal([]byte(name), &s)
	if good {
		assert.Nil(t, err)
		assert.Equal(t, sch, s)
	} else {
		assert.NotNil(t, err)
	}

	buf, err := yaml.Marshal(sch)
	if good {
		assert.Nil(t, err)
		assert.Equal(t, name+"\n", string(buf))
	} else {
		assert.NotNil(t, err)
	}
}
func TestScheduleYaml(t *testing.T) {
	testScheduleYaml(t, "minutely", Minutely, true)
	testScheduleYaml(t, "hourly", Hourly, true)
	testScheduleYaml(t, "daily", Daily, true)
	testScheduleYaml(t, "weekly", Weekly, true)
	testScheduleYaml(t, "monthly", Monthly, true)
	testScheduleYaml(t, "none", None, true)
	testScheduleYaml(t, "yearly", 100, false)
	testScheduleYaml(t, "- bad", 100, false)
}

func TestGetYamlFileFromFile(t *testing.T) {
	if _, err := getYamlFileFromFile("/tmp/nonexistent"); err == nil {
		t.Errorf("getYamlFileFromFile(\"/tmp/nonexistent\") = nil, expected error")
	}

	tmpfile, err := os.CreateTemp("", "invalid*.yaml")
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

func TestGetYamlFileFromPath(t *testing.T) {
	tmpDir := t.TempDir()
	defer os.RemoveAll(tmpDir)

	// test empty dir
	_, err := getYamlFileFromFile(tmpDir)
	assert.NotNil(t, err)
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
  - name: Env Variables
    url: $WEB_SITE
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
	case "Env Variables":
		assert.Equal(t, probe.URL, os.Getenv("WEB_SITE"))
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
      from: "send@email.com"
      to: "test@email.com"
      dry: true
`

func checkNotify(t *testing.T, n notify.Config) {
	for _, s := range n.Slack {
		checkSlackNotify(t, s)
	}
	for _, d := range n.Discord {
		checkDiscordNotify(t, d)
	}
	for _, tg := range n.Telegram {
		checkTelegramNotify(t, tg)
	}
	for _, e := range n.Email {
		checkEmailNotify(t, e)
	}
}

func checkSlackNotify(t *testing.T, n slack.NotifyConfig) {
	switch n.NotifyName {
	case "Slack":
		assert.Equal(t, n.WebhookURL, "https://hooks.slack.com/services/xxx")
		assert.Equal(t, n.Channels(), []string{"slack", "general"})
		assert.Equal(t, n.Dry, false)
	default:
		t.Errorf("unexpected notify name %s", n.NotifyName)
	}
}

func checkDiscordNotify(t *testing.T, n discord.NotifyConfig) {
	switch n.NotifyName {
	case "Discord":
		assert.Equal(t, n.WebhookURL, "https://discord.com/api/webhooks/xxx")
		assert.Equal(t, n.Avatar, "https://img.icons8.com/ios/72/appointment-reminders--v1.png")
		assert.Equal(t, n.Thumbnail, "https://freeiconshop.com/wp-content/uploads/edd/notification-flat.png")
		assert.Equal(t, n.Dry, true)
		assert.Equal(t, n.Retry.Times, 3)
		assert.Equal(t, n.Retry.Interval, 10*time.Second)
	case "Test":
		assert.Equal(t, n.WebhookURL, "https://discord.com/api/webhooks/xxx")
		assert.Equal(t, n.Avatar, "https://img.icons8.com/ios/72/appointment-reminders--v1.png")
		assert.Equal(t, n.Thumbnail, "https://freeiconshop.com/wp-content/uploads/edd/notification-flat.png")
		assert.Equal(t, n.Dry, false)
		assert.Equal(t, n.Retry.Times, 3)
		assert.Equal(t, n.Retry.Interval, 10*time.Second)
		assert.Equal(t, n.Channels(), []string{"general"})

	default:
		t.Errorf("unexpected notify name %s", n.NotifyName)
	}
}

func checkTelegramNotify(t *testing.T, n telegram.NotifyConfig) {
	switch n.NotifyName {
	case "Dev Channel":
		assert.Equal(t, n.Token, "123456:xxxxx")
		assert.Equal(t, n.ChatID, "-1001343458903")
		assert.Equal(t, n.Channels(), []string{"telegram#Dev", "test"})
		assert.Equal(t, n.Dry, false)
	case "Ops Group":
		assert.Equal(t, n.Token, "123456:yyyyyy")
		assert.Equal(t, n.ChatID, "-12310934")
		assert.Equal(t, n.Dry, true)
		assert.Equal(t, n.Retry.Times, 3)
		assert.Equal(t, n.Retry.Interval, 8*time.Second)
	default:
		t.Errorf("unexpected notify name %s", n.NotifyName)
	}
}

func checkEmailNotify(t *testing.T, n email.NotifyConfig) {
	switch n.NotifyName {
	case "email":
		assert.Equal(t, n.Server, "smtp.email.com:465")
		assert.Equal(t, n.User, "noreply@megaease.com")
		assert.Equal(t, n.Pass, "xxx")
		assert.Equal(t, n.From, "send@email.com")
		assert.Equal(t, n.To, "test@email.com")
		assert.Equal(t, n.Dry, true)
	default:
		t.Errorf("unexpected notify name %s", n.NotifyName)
	}
}

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
  timeformat: "2006-01-02 15:04:05 UTC"
`

func checkSettings(t *testing.T, s Settings) {
	assert.Equal(t, s.Name, "EaseProbeBot")
	assert.Equal(t, s.IconURL, "https://upload.wikimedia.org/wikipedia/commons/2/2d/Etcher-icon.png")
	assert.Equal(t, s.HTTPServer.IP, "127.0.0.1")
	assert.Equal(t, s.HTTPServer.Port, "8181")
	assert.Equal(t, s.HTTPServer.AutoRefreshTime, 5*time.Second)
	assert.Equal(t, s.HTTPServer.AccessLog.File, "/tmp/log/easeprobe.access.log")
	assert.Equal(t, s.HTTPServer.AccessLog.SelfRotate, true)
	assert.Equal(t, s.SLAReport.Schedule, Weekly)
	assert.Equal(t, s.SLAReport.Time, "23:59")
	assert.Equal(t, s.SLAReport.Backups, 20)
	assert.Equal(t, s.SLAReport.Channels, []string{"general"})
	assert.Equal(t, s.Notify.Dry, true)
	assert.Equal(t, s.Notify.Retry.Times, 5)
	assert.Equal(t, s.Notify.Retry.Interval, 10*time.Second)
	assert.Equal(t, s.Probe.Interval, 15*time.Second)
	assert.Equal(t, s.Log.Level, LogLevel(log.DebugLevel))
	assert.Equal(t, s.Log.MaxSize, 1)
	assert.Equal(t, s.TimeFormat, "2006-01-02 15:04:05 UTC")
}

const confYAML = confVer + confHTTP + confSSH + confHost + confClient + confNotify + confSettings

func writeConfig(file, content string) error {
	return os.WriteFile(file, []byte(content), 0644)
}

func httpServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(confYAML))
	})
	mux.HandleFunc("/modified", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(confYAML + "  \n  \n"))
	})

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		panic(err)
	}
}

func TestConfig(t *testing.T) {
	file := "./config.yaml"
	err := writeConfig(file, confYAML)
	assert.Nil(t, err)

	// bad config
	os.Setenv("WEB_SITE", "\n - x::")
	_, err = New(&file)
	assert.NotNil(t, err)

	os.Setenv("WEB_SITE", "https://easeprobe.com")
	monkey.Patch(yaml.Marshal, func(v interface{}) ([]byte, error) {
		return nil, errors.New("marshal error")
	})
	_, err = New(&file)
	assert.Nil(t, err)
	monkey.UnpatchAll()

	_, err = New(&file)
	assert.Nil(t, err)
	conf := Get()

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

	checkNotify(t, conf.Notify)
	checkSettings(t, conf.Settings)

	conf.InitAllLogs()
	probers := conf.AllProbers()
	assert.Equal(t, 9, len(probers))
	notifiers := conf.AllNotifiers()
	assert.Equal(t, 6, len(notifiers))

	go httpServer("65535")
	url := "http://localhost:65535"
	os.Setenv("HTTP_AUTHORIZATION", "Basic dXNlcm5hbWU6cGFzc3dvcmQ=")
	os.Setenv("HTTP_TIMEOUT", "10")
	httpConf, err := New(&url)
	assert.Nil(t, err)
	assert.Equal(t, "EaseProbeBot", httpConf.Settings.Name)
	assert.Equal(t, "0.1.0", httpConf.Version)

	// test config modification
	assert.False(t, IsConfigModified(url))
	assert.False(t, IsConfigModified(url))
	url += "/modified"
	assert.True(t, IsConfigModified(url))

	probers = conf.AllProbers()
	assert.Equal(t, 9, len(probers))
	notifiers = conf.AllNotifiers()
	assert.Equal(t, 6, len(notifiers))

	os.RemoveAll(file)
	os.RemoveAll("data")

	// error test
	url = "http://localhost:65534"
	_, err = New(&url)
	assert.NotNil(t, err)

	os.Setenv("HTTP_TIMEOUT", "invalid")
	_, err = New(&url)
	assert.NotNil(t, err)

	monkey.Patch(httpClient.NewRequest, func(method, url string, body io.Reader) (*http.Request, error) {
		return nil, errors.New("error")
	})
	url = "http://localhost"
	_, err = New(&url)
	assert.NotNil(t, err)

	monkey.UnpatchAll()
}

func TestInitData(t *testing.T) {
	c := Conf{}

	// data file disabled
	c.Settings.SLAReport.DataFile = "-"
	c.initData()
	assert.NoFileExists(t, c.Settings.SLAReport.DataFile)

	// default data file will be used
	c.Settings.SLAReport.DataFile = ""
	c.initData()
	assert.Equal(t, global.DefaultDataFile, c.Settings.SLAReport.DataFile)
	assert.DirExists(t, "data")
	os.RemoveAll("data")

	c.Settings.SLAReport.DataFile = "mydata/sla.yaml"
	c.initData()
	assert.Equal(t, "mydata/sla.yaml", c.Settings.SLAReport.DataFile)
	assert.DirExists(t, "mydata")

	os.WriteFile("mydata/sla.yaml", []byte("key : value"), 0644)
	c.initData()
	os.RemoveAll("mydata")

	monkey.Patch(os.MkdirAll, func(path string, perm os.FileMode) error {
		return fmt.Errorf("MkdirAll")
	})
	c.initData()
	assert.NoDirExists(t, "mydata")
	monkey.Unpatch(os.MkdirAll)
}

func TestEmptyNotifies(t *testing.T) {
	myConf := confVer
	file := "./config.yaml"
	err := writeConfig(file, myConf)
	assert.Nil(t, err)

	conf, err := New(&file)
	assert.Nil(t, err)
	notifiers := conf.AllNotifiers()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(notifiers))

	os.RemoveAll(file)
	os.RemoveAll("data")
}

func TestEmptyProbes(t *testing.T) {
	myConf := confVer
	file := "./config.yaml"
	err := writeConfig(file, myConf)
	assert.Nil(t, err)

	conf, err := New(&file)
	assert.Nil(t, err)
	probers := conf.AllProbers()
	assert.Equal(t, 0, len(probers))

	os.RemoveAll(file)
	os.RemoveAll("data")
}

func TestJSONSchema(t *testing.T) {
	schema, err := JSONSchema()
	assert.Nil(t, err)
	assert.NotEmpty(t, schema)

	monkey.Patch(json.MarshalIndent, func(v interface{}, prefix, indent string) ([]byte, error) {
		return nil, fmt.Errorf("error")
	})
	schema, err = JSONSchema()
	assert.NotNil(t, err)
	assert.Empty(t, schema)
	monkey.UnpatchAll()
}

func TestFileConfigModificaiton(t *testing.T) {
	file := "./config.yaml"
	err := writeConfig(file, confYAML)
	assert.Nil(t, err)
	ResetPreviousYAMLFile()
	assert.False(t, IsConfigModified(file))
	assert.False(t, IsConfigModified(file))

	err = writeConfig(file, confYAML+"  \n\n")
	assert.Nil(t, err)
	assert.True(t, IsConfigModified(file))
	assert.False(t, IsConfigModified(file))

	err = writeConfig(file, confYAML+"\ninvalid")
	assert.Nil(t, err)
	assert.False(t, IsConfigModified(file))

	os.RemoveAll(file)
	assert.False(t, IsConfigModified(file))

}
