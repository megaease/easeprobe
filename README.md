# EaseProbe
[![Go Report Card](https://goreportcard.com/badge/github.com/megaease/easeprobe)](https://goreportcard.com/report/github.com/megaease/easeprobe)

EaseProbe is a simple, standalone, and lightWeight tool that can do health/status checking, written in Go.

![](docs/overview.png)

<h2>Table of Contents</h2>

- [EaseProbe](#easeprobe)
  - [1. Overview](#1-overview)
    - [1.1 Probe](#11-probe)
    - [1.2 Notification](#12-notification)
    - [1.3 Report](#13-report)
    - [1.4 Administration](#14-administration)
  - [2. Getting Started](#2-getting-started)
    - [2.1 Build](#21-build)
    - [2.2 Configure](#22-configure)
    - [2.3 Run](#23-run)
  - [3. Configuration](#3-configuration)
    - [3.1 HTTP Probe Configuration](#31-http-probe-configuration)
    - [3.2 TCP Probe Configuration](#32-tcp-probe-configuration)
    - [3.3 Shell Command Probe Configuration](#33-shell-command-probe-configuration)
    - [3.4 SSH Command Probe Configuration](#34-ssh-command-probe-configuration)
    - [3.5 Host Resource Usage Probe Configuration](#35-host-resource-usage-probe-configuration)
    - [3.6 Native Client Probe](#36-native-client-probe)
    - [3.7 Notification Configuration](#37-notification-configuration)
    - [3.8 Global Setting Configuration](#38-global-setting-configuration)
  - [4. Community](#4-community)
  - [5. License](#5-license)

## 1. Overview

EaseProbe would do three kinds of work - **Probe**, **Notify**, and **Report**.

### 1.1 Probe

Ease Probe supports the following probing methods: **HTTP**, **TCP**, **Shell Command**, **SSH Command**,  **Host Resource Usage**, and **Native Client**.

> **Notes**:
>
> The prober name is a unique ID, DO NOT use the same name for a different prober.

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

- **Host**. Run an SSH command on a remote host and check the CPU, Memory, and Disk usage. ( [Host Load Probe](#35-host-resource-usage-probe-configuration) )

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

- **Client**. Currently, support the following native client. Support the mTLS. ( [Native Client Probe](#36-native-client-probe) )
  - **MySQL**. Connect to the MySQL server and run the `SHOW STATUS` SQL.
  - **Redis**. Connect to the Redis server and run the `PING` command.
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

Ease Probe supports the following notifications:

- **Slack**. Using Webhook for notification
- **Discord**. Using Webhook for notification
- **Telegram**. Using Telegram Bot for notification
- **Email**. Support multiple email addresses.
- **AWS SNS**. Support AWS Simple Notification Service.
- **WeChat Work**. Support Enterprise WeChat Work notification.
- **DingTalk**. Support the DingTalk notification.
- **Lark**. Support the Lark(Feishu) notification.
- **Log File**. Write the notification into a log file
- **SMS**. Support SMS notification with multiple SMS service providers - [Twilio](https://www.twilio.com/sms), [Vonage(Nexmo)](https://developer.vonage.com/messaging/sms/overview), [YunPain](https://www.yunpian.com/doc/en/domestic/list.html)

**Note**:

- The notification is **Edge-Triggered Mode**, only notified while the status is changed.

```YAML
# Notification Configuration
notify:
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
```

Check the  [Notification Configuration](#37-notification-configuration) to see how to configure it.

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

  Refer to the [Global Setting Configuration](#38-global-setting-configuration) to see how to configure the access log.


- **SLA Data Persistence**. Save the SLA statistics data on the disk.

  The SLA data would be persisted in `$CWD/data/data.yaml` by default. If you want to configure the path, you can do it in the `settings` section.

  Whenever EaseProbe starts,  load the data if found, and remove the probers that are not in configuration if the configuration changes.

  > **Note**:
  >
  > **The prober's name is a unique ID, so if multiple probes with the same name, the data would conflict, and the behavior is unknown.**

  ```YAML
  settings:
    sla:
      # SLA data persistence file path.
      # The default location is `$CWD/data/data.yaml`
      data: /path/to/data/file.yaml
  ```

For more information, please check the [Global Setting Configuration](#38-global-setting-configuration)


### 1.4 Administration

There are some administration configuration options:

**1) PID file**

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

**2) Log file Rotation**

  There are two types of log file: **Application Log** and **HTTP Access Log**. 

  Both Application Log and HTTP Access Log would be StdOut by default.  They all can be configured by:

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

  If `self_rotate` is `false`, EaseProbe would not rotate the log, and the log file would be rotated by the 3rd-party tool. e.g. `logrotate`.

  EaseProbe accepts the `HUP` signal to rotate the log.


## 2. Getting Started

### 2.1 Build

Compiler `Go 1.18+` (Generics Programming Support)

Use `make` to make the binary file. the target is under the `build/bin` directory

```shell
$ make
```
### 2.2 Configure

Read the [Configuration Guide](#3-configuration) to learn how to configure EaseProbe.

Create the configuration file - `$CWD/config.ymal`.

### 2.3 Run

Running the following command for the local test

```shell
$ build/bin/easeprobe -f config.yaml
```
* `-f` configuration file or URL. Can also be achieved by setting the environment variable `PROBE_CONFIG`
* `-d` dry run. Can also be achieved by setting the environment variable `PROBE_DRY`

## 3. Configuration
EaseProbe can be configured by supplying a yaml file or URL to fetch configuration settings from.
By default EaseProbe will look for its `config.yaml` on the current folder, this can be changed by supplying the `-f` parameter.

```shell
easeprobe -f path/to/config.yaml
easeprobe -f https://example.com/config
```

The following environment variables can be used to fine-tune the request to the configuration file
* `HTTP_AUTHORIZATION`
* `HTTP_TIMEOUT`

The following example configurations illustrate the EaseProbe supported features.

**Notes**: All probes have the following options:

- `timeout` - the maximum time to wait for the probe to complete. default : `30s`.
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
    # HTTP successful response code range, default is [0, 499].
    success_code:
      - [200,206] # the code >=200 and <= 206
      - [300,308] # the code >=300 and <= 308
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

### 3.5 Host Resource Usage Probe Configuration

Support the host probe, the configuration example as below.

The feature probe the CPU, Memory, and Disk usage, if one of them exceeds the threshold, then mark the host as status down.

> Note:
> - The thresholds are **OR** conditions, if one of them exceeds the threshold, then mark the host as status down.
> - The Host needs remote server have the following command: `top`, `df`, `free`, `awk`, `grep`, `tr`, and `hostname` (check the [source code](./probe/host/host.go) to see how it works).
> - The disk usage only check the root disk.

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

### 3.6 Native Client Probe

```YAML
# Native Client Probe
client:
  - name: Redis Native Client (local)
    driver: "redis"  # driver is redis
    host: "localhost:6379"  # server and port
    password: "abc123" # password
    # mTLS
    ca: /path/to/file.ca
    cert: /path/to/file.crt
    key: /path/to/file.key

  - name: MySQL Native Client (local)
    driver: "mysql"
    host: "localhost:3306"
    username: "root"
    password: "pass"

  - name: MongoDB Native Client (local)
    driver: "mongo"
    host: "localhost:27017"
    username: "admin"
    password: "abc123"
    timeout: 5s

  - name: Kafka Native Client (local)
    driver: "kafka"
    host: "localhost:9093"
    # mTLS
    ca: /path/to/file.ca
    cert: /path/to/file.crt
    key: /path/to/file.key

  - name: PostgreSQL Native Client (local)
    driver: "postgres"
    host: "localhost:5432"
    username: "postgres"
    password: "pass"

  - name: Zookeeper Native Client (local)
    driver: "zookeeper"
    host: "localhost:2181"
    timeout: 5s
    # mTLS
    ca: /path/to/file.ca
    cert: /path/to/file.crt
    key: /path/to/file.key
```


### 3.7 Notification Configuration

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
  # Notify by sms using yunpian  https://www.yunpian.com/official/document/sms/zh_cn/domestic_single_send
  sms:
    - name: "sms alert service - yunpian"
      provider: "yunpian"
      key: xxxxxxxxxxxx # yunpian apikey 
      mobile: 123456789,987654321 # mobile phone number, multi phone number joint by `,`
      sign: "xxxxx" # get this from yunpian

```

**Notes**: All of the notifications can have the following optional configuration.

```YAML
  dry: true # dry notification, print the Discord JSON in log(STDOUT)
  timeout: 20s # the timeout send out notification, default: 30s
  retry: # somehow the network is not good and needs to retry.
    times: 3 # default: 3
    interval: 10s # default: 5s
```


### 3.8 Global Setting Configuration

```YAML
# Global settings for all probes and notifiers.
settings:

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
  timeformat: "2006-01-02 15:04:05 UTC"

```

## 4. Community

- [Join Slack Workspace](https://join.slack.com/t/openmegaease/shared_invite/zt-upo7v306-lYPHvVwKnvwlqR0Zl2vveA) for requirement, issue, and development.
- [MegaEase on Twitter](https://twitter.com/megaease)

## 5. License

EaseProbe is under the Apache 2.0 license. See the [LICENSE](./LICENSE) file for details.
