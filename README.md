# EaseProbe
[![Go Report Card](https://goreportcard.com/badge/github.com/megaease/easeprobe)](https://goreportcard.com/report/github.com/megaease/easeprobe)
[![codecov](https://codecov.io/gh/megaease/easeprobe/branch/main/graph/badge.svg?token=L7SR8X6SRN)](https://codecov.io/gh/megaease/easeprobe)
[![Build](https://github.com/megaease/easeprobe/actions/workflows/test.yaml/badge.svg)](https://github.com/megaease/easeprobe/actions/workflows/test.yaml)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/megaease/easeprobe)](https://github.com/megaease/easeprobe/blob/main/go.mod)
[![Join MegaEase Slack](https://img.shields.io/badge/slack-megaease-brightgreen?logo=slack)](https://join.slack.com/t/openmegaease/shared_invite/zt-upo7v306-lYPHvVwKnvwlqR0Zl2vveA)


EaseProbe is a simple, standalone, and lightWeight tool that can do health/status checking, written in Go.

![](docs/overview.png)

<h2>Table of Contents</h2>

- [EaseProbe](#easeprobe)
  - [1. Overview](#1-overview)
    - [1.1 Probe](#11-probe)
    - [1.2 Notification](#12-notification)
    - [1.3 Report](#13-report)
    - [1.4 Channel](#14-channel)
    - [1.5 Administration](#15-administration)
    - [1.6 Prometheus Metrics Exporter](#16-prometheus-metrics-exporter)
  - [2. Getting Started](#2-getting-started)
    - [2.1 Build](#21-build)
    - [2.2 Configure](#22-configure)
    - [2.3 Run](#23-run)
  - [3. Configuration](#3-configuration)
    - [3.1 HTTP Probe Configuration](#31-http-probe-configuration)
    - [3.2 TCP Probe Configuration](#32-tcp-probe-configuration)
    - [3.3 Shell Command Probe Configuration](#33-shell-command-probe-configuration)
    - [3.4 SSH Command Probe Configuration](#34-ssh-command-probe-configuration)
    - [3.5 TLS Probe Configuration](#35-tls-probe-configuration)
    - [3.6 Host Resource Usage Probe Configuration](#36-host-resource-usage-probe-configuration)
    - [3.7 Native Client Probe Configuration](#37-native-client-probe-configuration)
    - [3.8 Notification Configuration](#38-notification-configuration)
    - [3.9 Global Setting Configuration](#39-global-setting-configuration)
  - [4. Benchmark](#4-benchmark)
  - [5. Contributing](#5-contributing)
  - [6. Community](#6-community)
  - [7. License](#7-license)

## 1. Overview

EaseProbe is designed to do three kinds of work - **Probe**, **Notify**, and **Report**.

### 1.1 Probe

EaseProbe supports the following probing methods: **HTTP**, **TCP**, **Shell Command**, **SSH Command**,  **Host Resource Usage**, and **Native Client**.

Each probe is identified by the method it supports (eg `http`), a unique name (across all probes in the configuration file) and the method specific parameters.

On application startup, the configured probes are scheduled for their initial fire up based on the following criteria:
* Less than or equal to 60 total probers exist: the delay between initial prober fire-up is `1 second`
* More than 60 total probers exist: the startup is scheduled based on the following equation `timeGap = DefaultProbeInterval / numProbes`

> **Note**:
>
> **If multiple probes using the same name then this could lead to corruption of the metrics data and/or the behavior of the application in non-deterministic way.**


- **HTTP**. Checking the HTTP status code, Support mTLS, HTTP Basic Auth, and can set the Request Header/Body. ( [HTTP Probe Configuration](#31-http-probe-configuration) )

  ```YAML
  http:
    # Some of the Software support the HTTP Query
    - name: ElasticSearch
      url: http://elasticsearch.server:9200
    - name: Prometheus
      url: http://prometheus:9090/graph
  ```

- **TCP**. Just simply check whether the TCP connection can be established or not. ( [TCP Probe Configuration](#32-tcp-probe-configuration) )

  ```YAML
  tcp:
    - name: Kafka
      host: kafka.server:9093
  ```

- **Shell**. Run a Shell command and check the result. ( [Shell Command Probe Configuration](#33-shell-command-probe-configuration) )

  ```YAML
  shell:
    # run redis-cli ping and check the "PONG"
    - name: Redis (Local)
      cmd: "redis-cli"
      args:
        - "-h"
        - "127.0.0.1"
        - "ping"
      env:
        # set the `REDISCLI_AUTH` environment variable for redis password
        - "REDISCLI_AUTH=abc123"
      # check the command output, if does not contain the PONG, mark the status down
      contain : "PONG"
  ```

- **SSH**. Run a remote command via SSH and check the result. Support the bastion/jump server  ([SSH Command Probe Configuration](#34-ssh-command-probe-configuration))

  ```YAML
  ssh:
    servers:
      - name : ServerX
        host: ubuntu@172.10.1.1:22
        password: xxxxxxx
        key: /Users/user/.ssh/id_rsa
        cmd: "ps auxwe | grep easeprobe | grep -v grep"
        contain: easeprobe
  ```

- **TLS**. Ping the remote endpoint, can probe for revoked or expired certificates ( [TLS Probe Configuration](#35-tls-probe-configuration) )

  ```YAML
  tls:
    - name: expired test
      host: expired.badssl.com:443
  ```

- **Host**. Run an SSH command on a remote host and check the CPU, Memory, and Disk usage. ( [Host Load Probe](#36-host-resource-usage-probe-configuration) )

  ```yaml
  host:
    servers:
      - name : server
        host: ubuntu@172.20.2.202:22
        key: /path/to/server.pem
        threshold:
          cpu: 0.80  # cpu usage  80%
          mem: 0.70  # memory usage 70%
          disk: 0.90  # disk usage 90%
  ```

- **Client**. Currently, support the following native client. Support the mTLS. ( refer to: [Native Client Probe Configuration](#37-native-client-probe-configuration) )
  - **MySQL**. Connect to the MySQL server and run the `SHOW STATUS` SQL.
  - **Redis**. Connect to the Redis server and run the `PING` command.
  - **Memcache**. Connect to a Memcache server and run the `version` command or check based on key/value checks.
  - **MongoDB**. Connect to MongoDB server and just ping server.
  - **Kafka**. Connect to Kafka server and list all topics.
  - **PostgreSQL**. Connect to PostgreSQL server and run `SELECT 1` SQL.
  - **Zookeeper**. Connect to Zookeeper server and run `get /` command.

  ```YAML
  client:
    - name: Kafka Native Client (local)
      driver: "kafka"
      host: "localhost:9093"
      # mTLS
      ca: /path/to/file.ca
      cert: /path/to/file.crt
      key: /path/to/file.key
  ```


### 1.2 Notification

EaseProbe supports the following notifications:

- **Slack**. Using Webhook for notification
- **Discord**. Using Webhook for notification
- **Telegram**. Using Telegram Bot for notification
- **Teams**. Support the [Microsoft Teams](https://docs.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/connectors-using?tabs=cURL#setting-up-a-custom-incoming-webhook) notification.
- **Email**. Support multiple email addresses.
- **AWS SNS**. Support AWS Simple Notification Service.
- **WeChat Work**. Support Enterprise WeChat Work notification.
- **DingTalk**. Support the DingTalk notification.
- **Lark**. Support the Lark(Feishu) notification.
- **SMS**. Support SMS notification with multiple SMS service providers - [Twilio](https://www.twilio.com/sms), [Vonage(Nexmo)](https://developer.vonage.com/messaging/sms/overview), [YunPain](https://www.yunpian.com/doc/en/domestic/list.html)
- **Log**. Write the notification into a log file or syslog.
- **Shell**. Run a shell command to notify the result. (see [example](resources/scripts/notify/notify.sh))

> **Note**:
>
> The notification is **Edge-Triggered Mode**, this means that these notifications are triggered when the status changes.
>
> Windows platforms do not support syslog as notification method.

```YAML
# Notification Configuration
notify:
  log:
    - name: log file # local log file
      file: /var/log/easeprobe.log
    - name: Remote syslog # syslog (!!! Not For Windows !!!)
      file: syslog # <-- must be "syslog" keyword
      host: 127.0.0.1:514 # remote syslog server - optional
      network: udp #remote syslog network [tcp, udp] - optional
  slack:
    - name: "MegaEase#Alert"
      webhook: "https://hooks.slack.com/services/........../....../....../"
  discord:
    - name: "MegaEase#Alert"
      webhook: "https://discord.com/api/webhooks/...../....../"
  telegram:
    - name: "MegaEase Alert Group"
      token: 1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZ # Bot Token
      chat_id: -123456789 # Channel / Group ID
  email:
    - name: "DevOps Mailing List"
      server: smtp.email.example.com:465
      username: user@example.com
      password: ********
      to: "user1@example.com;user2@example.com"
  aws_sns:
    - name: AWS SNS
      region: us-west-2
      arn: arn:aws:sns:us-west-2:298305261856:xxxxx
      endpoint: https://sns.us-west-2.amazonaws.com
      credential:
        id: AWSXXXXXXXID
        key: XXXXXXXX/YYYYYYY
  wecom:
    - name: "wecom alert service"
      webhook: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=589f9674-a2aa-xxxxxxxx-16bb6c43034a" # wecom robot webhook
  dingtalk:
    - name: "dingtalk alert service"
      webhook: "https://oapi.dingtalk.com/robot/send?access_token=xxxx"
  lark:
    - name: "lark alert service"
      webhook: "https://open.feishu.cn/open-apis/bot/v2/hook/d5366199-xxxx-xxxx-bd81-a57d1dd95de4"
  sms:
    - name: "sms alert service"
      provider: "yunpian"
      key: xxxxxxxxxxxx # yunpian apikey
      mobile: 123456789,987654321 # mobile phone number, multiple phone number joint by `,`
      sign: "xxxxxxxx" # need to register; usually brand name
  teams:
      - name: "teams alert service"
        webhook: "https://outlook.office365.com/webhook/a1269812-6d10-44b1-abc5-b84f93580ba0@9e7b80c7-d1eb-4b52-8582-76f921e416d9/IncomingWebhook/3fdd6767bae44ac58e5995547d66a4e4/f332c8d9-3397-4ac5-957b-b8e3fc465a8c" # see https://docs.microsoft.com/en-us/outlook/actionable-messages/send-via-connectors
  shell: # EaseProbe set the environment variables -
         # (see the example: resources/scripts/notify/notify.sh)
    - name: "shell alert service"
      cmd: "/bin/bash"
      args:
        - "-c"
        - "/path/to/script.sh"
      env: # set the env to the notification command
        - "EASEPROBE=1"
```

Check the  [Notification Configuration](#38-notification-configuration) to see how to configure it.

### 1.3 Report

- **SLA Report Notify**. EaseProbe would send the daily, weekly, or monthly SLA report.

  ```YAML
  settings:
    # SLA Report schedule
    sla:
      #  daily, weekly (Sunday), monthly (Last Day), none
      schedule: "weekly"
      # UTC time, the format is 'hour:min:sec'
      time: "23:59"
  ```

- **SLA Live Report**. You can query the SLA Live Report

  The EaseProbe would listen on the `0.0.0.0:8181` port by default. And you can access the Live SLA report by the following URL:

  - HTML: `http://localhost:8181/`
  - JSON: `http://localhost:8181/api/v1/sla`

You can use the following URL query options for both HTML and JSON:
  - `refresh`: (_HTML only_) refresh the page every given seconds (ex, `?refresh=30s` refreshes the page every 30 seconds)
  - `pg` & `sz`: page number and page size (ex, `?pg=2&sz=10` shows the second page with 10 probers), default page size is `100`
  - `name`: filter the probers that contain the value of name (ex, `?name=probe1` list the probers which name containing `probe1`)
  - `kind`: filter the probers with the kind (ex, `?kind=http` list the probers with kind `http`)
  - `ep`: filter the probers with the endpoint (ex, `?ep=example.com` list the probers which endpoint containing  `example.com`)
  - `msg`: filter the probers with the message (ex, `?msg=example` list the probers which message containing `example`)
  - `status`: filter the probers with specific status, accepted values `up` or `down` (ex. `?status=up` list only probers with status `up`).
  - `gte`: filter the probers with SLA greater than or equal to the given percentage (ex. `?gte=50` filter only hosts with SLA percentage `>= 50%`)
  - `lte`:filter the probers with SLA less than or equal to the given percentage (ex. `?lte=90` filter only hosts with SLA percentage `<= 90%` )

  Refer to the [Global Setting Configuration](#39-global-setting-configuration) to see how to configure the access log.


- **SLA Data Persistence**. Save the SLA statistics data on the disk.

  The SLA data would be persisted in `$CWD/data/data.yaml` by default. If you want to configure the path, you can do it in the `settings` section.

  When EaseProbe starts, it looks for the location of `data.yaml` and if found, loads the file and removes any probes that are no longer present in the configuration file. Setting a value of `"-"` for `data:` disables SLA persistence (eg `data: "-"`).

  ```YAML
  settings:
    sla:
      # SLA data persistence file path.
      # The default location is `$CWD/data/data.yaml`
      data: /path/to/data/file.yaml
  ```

For more information, please check the [Global Setting Configuration](#39-global-setting-configuration)


### 1.4 Channel

The Channel is used for connecting the Probers and the Notifiers. It can be configured for every Prober and Notifier.

This feature could help you group the Probers and Notifiers into a logical group.

> **Note**:
>
> If no Channel is defined on a probe or notify entry, then the default channel will be used. The default channel name is `__EaseProbe_Channel__`
>
> EaseProbe versions prior to v1.5.0, do not have support for the `channel` feature

For example:

```YAML
http:
   - name: probe A
     channels : [ Dev_Channel, Manager_Channel ]
shell:
   - name: probe B
     channels: [ Ops_Channel ]
notify:
   - discord: Discord
     channels: [ Dev_Channel, Ops_Channel ]
   - email: Gmail
     channels: [ Mgmt_Channel ]
```

Then, we will have the following diagram

```
┌───────┐          ┌──────────────┐
│Probe B├─────────►│ Mgmt_Channel ├────┐
└───────┘          └──────────────┘    │
                                       │
                                       │
                   ┌─────────────┐     │   ┌─────────┐
            ┌─────►│ Dev_Channel ├─────▼───► Discord │
            │      └─────────────┘         └─────────┘
┌───────┐   │
│Probe A├───┤
└───────┘   │
            │      ┌────────────┐          ┌─────────┐
            └─────►│ QA_Channel ├──────────►  Gmail  │
                   └────────────┘          └─────────┘
```

### 1.5 Administration

There are some administration configuration options:

**PID file**

  The EaseProbe would create a PID file (default `$CWD/easeprobe.pid`) when it starts. it can be configured by:

  ```YAML
  settings:
    pid: /var/run/easeprobe.pid
  ```

  - If the file already exists, EaseProbe would overwrite it.
  - If the file cannot be written, EaseProbe would exit with an error.

  If you want to disable the PID file, you can configure the pid file to "".

  ```YAML
  settings:
    pid: "" # EaseProbe won't create a PID file
  ```

**Log file Rotation**

  There are two types of the log files: **Application Log** and **HTTP Access Log**.

  Both application and HTTP access logs will be displayed on StdOut by default. Both can be be configured by the `log:` directive such as:

  ```YAML
  log:
    file: /path/to/log/file
    self_rotate: true # default: true
  ```

  If `self_rotate` is `true`, EaseProbe would rotate the log automatically, and the following options are available:

  ```YAML
    size: 10 # max size of log file. default: 10M
    age: 7 # max age days of log file. default: 7 days
    backups: 5 # max backup log files. default: 5
    compress: true # compress. default: true
  ```

  If `self_rotate` is `false`, EaseProbe will not rotate the log, and the log file will have to be rotated by a 3rd-party tool (such as `logrotate`) or manually by the administrator.

  ```shell
  mv /path/to/easeprobe.log /path/to/easeprobe.log.0
  kill -HUP `cat /path/to/easeprobe.pid`
  ```

  EaseProbe accepts the `HUP` signal to rotate the log.

### 1.6 Prometheus Metrics Exporter

EaseProbe supports Prometheus metrics exporter. The Prometheus endpoint is `http://localhost:8181/metrics` by default.

Currently, All of the Probers support the following metrics:

  - `total`: the total number of probes
  - `duration`: Probe duration in milliseconds
  - `status`: Probe status
  - `SLA`: Probe SLA percentage

And for the different Probers, the following metrics are available:

- HTTP Probe
  - `status_code`: HTTP status code
  - `content_len`: HTTP content length
  - `dns_duration`: DNS duration in milliseconds
  - `connect_duration`: TCP connection duration in milliseconds
  - `tls_duration`: TLS handshake duration in milliseconds
  - `send_duration`: HTTP send duration in milliseconds
  - `wait_duration`: HTTP wait duration in milliseconds
  - `transfer_duration`: HTTP transfer duration in milliseconds
  - `total_duration`: HTTP total duration in milliseconds

- TLS Probe
  - `earliest_cert_expiry`: last TLS chain expiry in timestamp seconds
  - `last_chain_expiry_timestamp_seconds`: earliest TLS cert expiry in Unix time

- Shell & SSH Probe
  - `exit_code`: exit code of the command
  - `output_len`: length of the output

- Host Probe
  - `cpu`: CPU usage in percentage
  - `memory`: memory usage in percentage
  - `disk`: disk usage in percentage


The following snapshot is the Grafana panel for host CPU metrics

![](./docs/grafana.demo.png)

Refer to the [Global Setting Configuration](#39-global-setting-configuration) for further details on how to configure the HTTP server.

## 2. Getting Started

You can get started with EaseProbe, by any of the following methods:
* Download the release for your platform from https://github.com/megaease/easeprobe/releases
* Use the available EaseProbe docker image `docker run -it megaease/easeprobe`
* Build `easeprobe` from sources

### 2.1 Build

Compiler `Go 1.18+` (Generics Programming Support)

Use `make` to build and produce the `easeprobe` binary file. The executable is produced under the `build/bin` directory

```shell
$ make
```
### 2.2 Configure

Read the [Configuration Guide](#3-configuration) to learn how to configure EaseProbe.

Create the configuration file - `$CWD/config.yaml`.

The following is an example of simple configuration file to get started:

```YAML
http: # http probes
  - name: EaseProbe Github
    url: https://github.com/megaease/easeprobe
notify:
  log:
    - name: log file # local log file
      file: /var/log/easeprobe.log
settings:
  probe:
    timeout: 30s # the time out for all probes
    interval: 1m # probe every minute for all probes
```

### 2.3 Run

Running the following command for the local test

```shell
$ build/bin/easeprobe -f config.yaml
```
* `-f` configuration file or URL. Can also be achieved by setting the environment variable `PROBE_CONFIG`
* `-d` dry run. Can also be achieved by setting the environment variable `PROBE_DRY`

## 3. Configuration

EaseProbe can be configured by supplying a YAML file or URL to fetch configuration settings from.

By default, EaseProbe will look for its `config.yaml` on the current folder. This behavior can be changed by supplying the `-f` parameter.

```shell
easeprobe -f path/to/config.yaml
easeprobe -f https://example.com/config
```

The following environment variables can be used to fine-tune the request to the configuration file

* `HTTP_AUTHORIZATION`
* `HTTP_TIMEOUT`

And the configuration file should be versioned, the version should be aligned with the EaseProbe binary version.

```yaml
version: v1.5.0
```

The following example configurations illustrate the EaseProbe supported features.

**Note**:   All probes have the following options:

- `timeout` - the maximum time to wait for the probe to complete. default: `30s`.
- `interval` - the interval time to run the probe. default: `1m`.


### 3.1 HTTP Probe Configuration

```YAML
# HTTP Probe Configuration

http:
  # A Website
  - name: MegaEase Website (Global)
    url: https://megaease.com

  # Some of the Software support the HTTP Query
  - name: ElasticSearch
    url: http://elasticsearch.server:9200
  - name: Eureka
    url: http://eureka.server:8761
  - name: Prometheus
    url: http://prometheus:9090/graph

  # Spring Boot Application with Actuator Heath API
  - name: EaseService-Governance
    url: http://easeservice-mgmt-governance:38012/actuator/health
  - name: EaseService-Control
    url: http://easeservice-mgmt-control:38013/actuator/health
  - name: EaseService-Mesh
    url: http://easeservice-mgmt-mesh:38013/actuator/health

  # A completed HTTP Probe configuration
  - name: Special Website
    url: https://megaease.cn
    # Request Method
    method: GET
    # Request Header
    headers:
      X-head-one: xxxxxx
      X-head-two: yyyyyy
      X-head-THREE: zzzzzzX-
    content_encoding: text/json
    # Request Body
    body: '{ "FirstName": "Mega", "LastName" : "Ease", "UserName" : "megaease", "Email" : "user@example.com"}'
    # HTTP Basic Auth
    username: username
    password: password
    # mTLS
    ca: /path/to/file.ca
    cert: /path/to/file.crt
    key: /path/to/file.key
    # TLS
    insecure: true # skip any security checks, useful for self-signed and expired certs. default: false
    # HTTP successful response code range, default is [0, 499].
    success_code:
      - [200,206] # the code >=200 and <= 206
      - [300,308] # the code >=300 and <= 308
    # Response Checking
    contain: "success" # response body must contain this string, if not the probe is considered failed.
    not_contain: "failure" # response body must NOT contain this string, if it does the probe is considered failed.
    # configuration
    timeout: 10s # default is 30 seconds

```
### 3.2 TCP Probe Configuration

```YAML
# TCP Probe Configuration
tcp:
  - name: SSH Service
    host: example.com:22
    timeout: 10s # default is 30 seconds
    interval: 2m # default is 60 seconds

  - name: Kafka
    host: kafka.server:9093
```

### 3.3 Shell Command Probe Configuration

The shell command probe is used to execute a shell command and check the output.

The following example shows how to configure the shell command probe.

```YAML
# Shell Probe Configuration
shell:
  # A proxy curl shell script
  - name: Google Service
    cmd: "./resources/probe/scripts/proxy.curl.sh"
    args:
      - "socks5://127.0.0.1:1085"
      - "www.google.com"

  # run redis-cli ping and check the "PONG"
  - name: Redis (Local)
    cmd: "redis-cli"
    args:
      - "-h"
      - "127.0.0.1"
      - "ping"
    clean_env: true # Do not pass the OS environment variables to the command
                    # default: false
    env:
      # set the `REDISCLI_AUTH` environment variable for redis password
      - "REDISCLI_AUTH=abc123"
    # check the command output, if does not contain the PONG, mark the status down
    contain : "PONG"

  # Run Zookeeper command `stat` to check the zookeeper status
  - name: Zookeeper (Local)
    cmd: "/bin/sh"
    args:
      - "-c"
      - "echo stat | nc 127.0.0.1 2181"
    contain: "Mode:"
```

### 3.4 SSH Command Probe Configuration

SSH probe is similar to Shell probe.
- Support Password and Private key authentication.
- Support the Bastion host tunnel.

The `host` supports the following configuration
- `example.com`
- `example.com:22`
- `user@example.com:22`

The following are examples of SSH probe configuration.

```YAML
# SSH Probe Configuration
ssh:
  # SSH bastion host configuration
  bastion:
    aws: # bastion host ID      ◄──────────────────────────────┐
      host: aws.basition.com:22 #                              │
      username: ubuntu # login user                            │
      key: /path/to/aws/basion/key.pem # private key file      │
    gcp: # bastion host ID                                     │
      host: ubuntu@gcp.basition.com:22 # bastion host          │
      key: /path/to/gcp/basion/key.pem # private key file      │
  # SSH Probe configuration                                    │
  servers:   #                                                 │
    # run redis-cli ping and check the "PONG"                  │
    - name: Redis (AWS) # Name                                 │
      bastion: aws  # bastion host id ------------------------─┘
      host: 172.20.2.202:22
      username: ubuntu  # SSH Login username
      password: xxxxx   # SSH Login password
      key: /path/to/private.key # SSH login private file
      cmd: "redis-cli"
      args:
        - "-h"
        - "127.0.0.1"
        - "ping"
      env:
        # set the `REDISCLI_AUTH` environment variable for redis password
        - "REDISCLI_AUTH=abc123"
      # check the command output, if does not contain the PONG, mark the status down
      contain : "PONG"

    # Check the process status of `Kafka`
    - name:  Kafka (GCP)
      bastion: gcp         #  ◄------ bastion host id
      host: 172.10.1.100:22
      username: ubuntu
      key: /path/to/private.key
      cmd: "ps -ef | grep kafka"
```

### 3.5 TLS Probe Configuration

TLS ping to remote endpoint, can probe for revoked or expired certificates

  ```YAML
  tls:
    - name: expired test
      host: expired.badssl.com:443
      insecure_skip_verify: true # dont check cert validity
      expire_skip_verify: true # dont check cert expire date
      alert_expire_before: 168h # alert if cert expire date is before X, the value is a Duration, see https://pkg.go.dev/time#ParseDuration. example: 1h, 1m, 1s. expire_skip_verify must be false to use this feature.
      # root_ca_pem_path: /path/to/root/ca.pem # ignore if root_ca_pem is present
      # root_ca_pem: |
      #   -----BEGIN CERTIFICATE-----
    - name: untrust test
      host: untrusted-root.badssl.com:443
      # insecure_skip_verify: true # dont check cert validity
      # expire_skip_verify: true # dont check cert expire date
      # root_ca_pem_path: /path/to/root/ca.pem # ignore if root_ca_pem is present
      # root_ca_pem: |
      #   -----BEGIN CERTIFICATE-----
  ```

### 3.6 Host Resource Usage Probe Configuration

The host resource usage probe allows for collecting information and alerting when certain resource utilization thresholds are exceeded.

The resources currently monitored include CPU, memory and disk utilization. The probe status is considered as `down` when any value exceeds its defined threshold.

> **Note**:
> - The host running easerprobe needs the following commands to be installed on the remote system that will be monitored: `top`, `df`, `free`, `awk`, `grep`, `tr`, and `hostname` (check the [source code](./probe/host/host.go) for more details on this works and/or modify its behavior).
> - The disk usage check is limited to the root filesystem only with the following command `df -h /`.

```yaml
host:
  bastion: # bastion server configuration
    aws: # bastion host ID      ◄──────────────────┐
      host: ubuntu@example.com # bastion host      │
      key: /path/to/bastion.pem # private key file │
  # Servers List                                   │
  servers: #                                       │
    - name : aws server   #                        │
      bastion: aws #  <-- bastion server id ------─┘
      host: ubuntu@172.20.2.202:22
      key: /path/to/server.pem
      disks: # [optional] Check multiple disks. if not present, only check `/` by default
        - /
        - /data
      threshold:
        cpu: 0.80  # cpu usage  80%
        mem: 0.70  # memory usage 70%
        disk: 0.90  # disk usage 90%

    # Using the default threshold
    # cpu 80%, mem 80% and disk 95%
    - name : My VPS
      host: user@example.com:22
      key: /Users/user/.ssh/id_rsa
```

### 3.7 Native Client Probe Configuration

```YAML
# Native Client Probe
client:
  - name: Redis Native Client (local)
    driver: "redis"  # driver is redis
    host: "localhost:6379"  # server and port
    password: "abc123" # password
    data:         # Optional
      key: val    # Check that `key` exists and its value is `val`
    # mTLS - Optional
    ca: /path/to/file.ca
    cert: /path/to/file.crt
    key: /path/to/file.key

  - name: MySQL Native Client (local)
    driver: "mysql"
    host: "localhost:3306"
    username: "root"
    password: "pass"
    data: # Optional, check the specific column value in the table
      #  Usage: "database:table:column:primary_key:value" : "expected_value"
      #         transfer to : "SELECT column FROM database.table WHERE primary_key = value"
      #         the `value` for `primary_key` must be int
      "test:product:name:id:1" : "EaseProbe" # select name from test.product where id = 1
      "test:employee:age:id:2" : 45          # select age from test.employee where id = 2
    # mTLS - Optional
    ca: /path/to/file.ca
    cert: /path/to/file.crt
    key: /path/to/file.key

  - name: MongoDB Native Client (local)
    driver: "mongo"
    host: "localhost:27017"
    username: "admin"
    password: "abc123"
    timeout: 5s
    data: # Optional, find the specific value in the table
      #  Usage: "database:collection" : "{JSON}"
      "test.employee" : '{"name":"Hao Chen"}' # find the employee with name "Hao Chen"
      "test.product" : '{"name":"EaseProbe"}' # find the product with name "EaseProbe"

  - name: Memcache Native Client (local)
    driver: "memcache"
    host: "localhost:11211"
    timeout: 5s
    data:         # Optional
      key: val    # Check that key exists and its value is val
      "namespace:key": val # Namespaced keys enclosed in "

  - name: Kafka Native Client (local)
    driver: "kafka"
    host: "localhost:9093"
    # mTLS - Optional
    ca: /path/to/file.ca
    cert: /path/to/file.crt
    key: /path/to/file.key

  - name: PostgreSQL Native Client (local)
    driver: "postgres"
    host: "localhost:5432"
    username: "postgres"
    password: "pass"
    data: # Optional, check the specific column value in the table
      #  Usage: "database:table:column:primary_key:value" : "expected_value"
      #         transfer to : "SELECT column FROM table WHERE primary_key = value"
      #         the `value` for `primary_key` must be int
      "test:product:name:id:1" : "EaseProbe" # select name from product where id = 1
      "test:employee:age:id:2" : 45          # select age from employee where id = 2
    # mTLS - Optional
    ca: /path/to/file.ca
    cert: /path/to/file.crt
    key: /path/to/file.key

  - name: Zookeeper Native Client (local)
    driver: "zookeeper"
    host: "localhost:2181"
    timeout: 5s
    # mTLS
    ca: /path/to/file.ca
    cert: /path/to/file.crt
    key: /path/to/file.key
```


### 3.8 Notification Configuration

```YAML
# Notification Configuration
notify:
  # Notify to Slack Channel
  slack:
    - name: "Organization #Alert"
      webhook: "https://hooks.slack.com/services/........../....../....../"
      # dry: true   # dry notification, print the Slack JSON in log(STDOUT)
  telegram:
    - name: "Group Name"
      token: 1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZ # Bot Token
      chat_id: -123456789 # Group ID
    - name: "Channel Name"
      token: 1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZ # Bot Token
      chat_id: -1001234567890 # Channel ID
  # Notify to Discord Text Channel
  discord:
    - name: "Server #Alert"
      webhook: "https://discord.com/api/webhooks/...../....../"
      # the avatar and thumbnail setting for notify block
      avatar: "https://img.icons8.com/ios/72/appointment-reminders--v1.png"
      thumbnail: "https://freeiconshop.com/wp-content/uploads/edd/notification-flat.png"
      # dry: true # dry notification, print the Discord JSON in log(STDOUT)
      retry: # something the network is not good need to retry.
        times: 3
        interval: 10s
  # Notify to email addresses
  email:
    - name: "XXX Mail List"
      server: smtp.email.example.com:465
      username: user@example.com
      password: ********
      to: "user1@example.com;user2@example.com"
      from: "from@example.com" # Optional
      # dry: true # dry notification, print the Email HTML in log(STDOUT)
  # Notify to AWS Simple Notification Service
  aws_sns:
    - name: AWS SNS
      region: us-west-2 # AWS Region
      arn: arn:aws:sns:us-west-2:298305261856:xxxxx # SNS ARN
      endpoint: https://sns.us-west-2.amazonaws.com # SNS Endpoint
      credential: # AWS Access Credential
        id: AWSXXXXXXXID  # AWS Access Key ID
        key: XXXXXXXX/YYYYYYY # AWS Access Key Secret
  # Notify to Wecom(WeChatwork) robot.
  wecom:
    - name: "wecom alert service"
      webhook: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=589f9674-a2aa-xxxxxxxx-16bb6c43034a" # wecom robot webhook
  # Notify to Dingtalk
  dingtalk:
    - name: "dingtalk alert service"
      webhook: "https://oapi.dingtalk.com/robot/send?access_token=xxxx"
  # Notify to Lark
  lark:
    - name: "lark alert service"
      webhook: "https://open.feishu.cn/open-apis/bot/v2/hook/d5366199-xxxx-xxxx-bd81-a57d1dd95de4"
  # Notify to a local log file
  log:
    - name: "Local Log"
      file: "/tmp/easeprobe.log"
      dry: true
    - name: Remote syslog # syslog (!!! Not For Windows !!!)
      file: syslog # <-- must be "syslog" keyword
      host: 127.0.0.1:514 # remote syslog server - optional
      network: udp #remote syslog network [tcp, udp] - optional
  # Notify by sms using yunpian  https://www.yunpian.com/official/document/sms/zh_cn/domestic_single_send
  sms:
    - name: "sms alert service - yunpian"
      provider: "yunpian"
      key: xxxxxxxxxxxx # yunpian apikey
      mobile: 123456789,987654321 # mobile phone number, multi phone number joint by `,`
      sign: "xxxxx" # get this from yunpian

  # EaseProbe set the following environment variables
  #  - EASEPROBE_TYPE: "Status" or "SLA"
  #  - EASEPROBE_NAME: probe name
  #  - EASEPROBE_STATUS: "up" or "down"
  #  - EASEPROBE_RTT: round trip time in milliseconds
  #  - EASEPROBE_TIMESTAMP: timestamp of probe time
  #  - EASEPROBE_MESSAGE: probe message
  # and offer two formats of string
  #  - EASEPROBE_JSON: the JSON format
  #  - EASEPROBE_CSV: the CSV format
  # The CVS format would be set for STDIN for the shell command.
  # (see the example: resources/scripts/notify/notify.sh)
  shell:
    - name: "shell alert service"
      cmd: "/bin/bash"
      args:
        - "-c"
        - "/path/to/script.sh"
      clean_env: true # Do not pass the OS environment variables to the command
                      # default: false
      env: # set the env to the notification command
        - "EASEPROBE=1"
        - "KEY=Value"
```

**Note**: All of the notifications support the following optional configuration parameters.

```YAML
  dry: true # dry notification, print the Discord JSON in log(STDOUT)
  timeout: 20s # the timeout send out notification, default: 30s
  retry: # somehow the network is not good and needs to retry.
    times: 3 # default: 3
    interval: 10s # default: 5s
```


### 3.9 Global Setting Configuration

```YAML
# Global settings for all probes and notifiers.
settings:

  # The customized name and icon
  name: "EaseProbe" # the name of the probe: default: "EaseProbe"
  icon: "https://path/to/icon.png" # the icon of the probe. default: "https://megaease.com/favicon.png"
  # Daemon settings

  # pid file path,  default: $CWD/easeprobe.pid,
  # if set to "", will not create pid file.
  pid: /var/run/easeprobe.pid

  # A HTTP Server configuration
  http:
    ip: 127.0.0.1 # the IP address of the server. default:"0.0.0.0"
    port: 8181 # the port of the server. default: 8181
    refresh: 5s # the auto-refresh interval of the server. default: the minimum value of the probes' interval.
    log:
      file: /path/to/access.log # access log file. default: Stdout
      # Log Rotate Configuration (optional)
      self_rotate: true # true: self rotate log file. default: true
                        # false: managed by outside  (e.g logrotate)
                        #        the blow settings will be ignored.
      size: 10 # max of access log file size. default: 10m
      age: 7 #  max of access log file age. default: 7 days
      backups: 5 # max of access log file backups. default: 5
      compress: true # compress the access log file. default: true

  # SLA Report schedule
  sla:
    #  daily, weekly (Sunday), monthly (Last Day), none
    schedule : "daily"
    # UTC time, the format is 'hour:min:sec'
    time: "23:59"
    # debug mode
    # - true: send the SLA report every minute
    # - false: send the SLA report in schedule
    debug: false
    # SLA data persistence file path.
    # The default location is `$CWD/data/data.yaml`
    data: /path/to/data/file.yaml
    # Use the following to disable SLA data persistence
    # data: "-"
    backups: 5 # max of SLA data file backups. default: 5
               # if set to a negative value, keep all backup files

  notify:
    # dry: true # Global settings for dry run
    retry: # Global settings for retry
      times: 5
      interval: 10s

  probe:
    timeout: 30s # the time out for all probes
    interval: 1m # probe every minute for all probes

  # easeprobe program running log file.
  log:
    file: "/path/to/easeprobe.log" # default: stdout
    # Log Level Configuration
    # can be: panic, fatal, error, warn, info, debug.
    level: "debug"
    # Log Rotate Configuration (optional)
    self_rotate: true # true: self rotate log file. default: true
                        # false: managed by outside  (e.g logrotate)
                        #        the blow settings will be ignored.
    size: 10 # max of access log file size. default: 10m
    age: 7 #  max of access log file age. default: 7 days
    backups: 5 # max of access log file backups. default: 5
    compress: true # compress the access log file. default: true

  # Date format
  # Date
  #  - January 2, 2006
  #  - 01/02/06
  #  - Jan-02-06
  #
  # Time
  #   - 15:04:05
  #   - 3:04:05 PM
  #
  # Date Time
  #   - Jan _2 15:04:05                   (Timestamp)
  #   - Jan _2 15:04:05.000000            (with microseconds)
  #   - 2006-01-02T15:04:05-0700          (ISO 8601 (RFC 3339))
  #   - 2006-01-02 15:04:05
  #   - 02 Jan 06 15:04 MST               (RFC 822)
  #   - 02 Jan 06 15:04 -0700             (with numeric zone)
  #   - Mon, 02 Jan 2006 15:04:05 MST     (RFC 1123)
  #   - Mon, 02 Jan 2006 15:04:05 -0700   (with numeric zone)
  timeformat: "2006-01-02 15:04:05 Z07:00"
  # check the following link to see the time zone list
  # https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
  timezone: "America/New_York" #  default: UTC
```

## 4. Benchmark

Refer to - [Benchmark Report](./docs/Benchmark.md)

## 5. Contributing

If you're interested in contributing to the project, please spare a moment to read our [CONTRIBUTING Guide](./docs/CONTRIBUTING.md)


## 6. Community

- [Join Slack Workspace](https://join.slack.com/t/openmegaease/shared_invite/zt-upo7v306-lYPHvVwKnvwlqR0Zl2vveA) for requirement, issue, and development.
- [MegaEase on Twitter](https://twitter.com/megaease)

## 7. License

EaseProbe is under the Apache 2.0 license. See the [LICENSE](./LICENSE) file for details.