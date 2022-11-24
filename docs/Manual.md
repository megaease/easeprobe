<h1>User Manual</h1>

EaseProbe is a simple, standalone, and lightweight tool that can do health/status checking, written in Go.

![](./overview.png)

<h1>Outline</h1>

- [1. Probe](#1-probe)
  - [1.1 Overview](#11-overview)
  - [1.2 Initial Fire Up](#12-initial-fire-up)
- [2. Notification](#2-notification)
  - [2.1 Slack](#21-slack)
  - [2.2 Discord](#22-discord)
  - [2.3 Telegram](#23-telegram)
  - [2.4 Teams](#24-teams)
  - [2.5 Email](#25-email)
  - [2.6 AWS SNS](#26-aws-sns)
  - [2.7 WeChat Work](#27-wechat-work)
  - [2.8 DingTalk](#28-dingtalk)
  - [2.9 Lark](#29-lark)
  - [2.10 SMS](#210-sms)
  - [2.11 Log](#211-log)
  - [2.12 Shell](#212-shell)
  - [2.13 RingCentral](#213-ringcentral)
- [3. Report](#3-report)
  - [3.1 SLA Report Notification](#31-sla-report-notification)
  - [3.2 SLA Live Report](#32-sla-live-report)
  - [3.3 SLA Data Persistence](#33-sla-data-persistence)
- [4. Channel](#4-channel)
  - [4.1 Overview](#41-overview)
  - [4.2 Examples](#42-examples)
- [5. Administration](#5-administration)
  - [5.1 PID file](#51-pid-file)
  - [5.2 Log file Rotation](#52-log-file-rotation)
- [6. Prometheus Metrics Exporter](#6-prometheus-metrics-exporter)
- [7. Configuration](#7-configuration)
  - [7.1 HTTP Probe Configuration](#71-http-probe-configuration)
    - [7.1.1 Basic HTTP Configuration](#711-basic-http-configuration)
    - [7.1.2 Complete HTTP Configuration](#712-complete-http-configuration)
    - [7.1.3 Expression Evaluation](#713-expression-evaluation)
  - [7.2 TCP Probe Configuration](#72-tcp-probe-configuration)
  - [7.3 Ping Probe Configuration](#73-ping-probe-configuration)
  - [7.4 Shell Command Probe Configuration](#74-shell-command-probe-configuration)
  - [7.5 SSH Command Probe Configuration](#75-ssh-command-probe-configuration)
  - [7.6 TLS Probe Configuration](#76-tls-probe-configuration)
  - [7.7 Host Resource Usage Probe Configuration](#77-host-resource-usage-probe-configuration)
  - [7.8 Native Client Probe Configuration](#78-native-client-probe-configuration)
  - [7.9 Notification Configuration](#79-notification-configuration)
  - [7.10 Global Setting Configuration](#710-global-setting-configuration)
- [8. Tools](#8-tools)
  - [8.1 EaseProbe JSON Schema](#81-easeprobe-json-schema)



# 1. Probe

## 1.1 Overview

EaseProbe supports the following probing methods: **HTTP**, **TCP**, **TLS**, **Shell Command**, **SSH Command**, **Host Resource Usage**, and **Native Client**.

Each probe is identified by the method it supports (eg `http`), a unique name (across all probes in the configuration file) and the method specific parameters.
- **HTTP**. Checking the HTTP status code, Support mTLS, HTTP Basic Auth, and can set the Request Header/Body. ( [HTTP Probe Configuration](#71-http-probe-configuration) )
- **TCP**. Just simply check whether the TCP connection can be established or not. ( [TCP Probe Configuration](#72-tcp-probe-configuration) )
- **Ping**. Just simply check whether can be pinged or not. ( [Ping Probe Configuration](#73-ping-probe-configuration) )
- **Shell**. Run a Shell command and check the result. ( [Shell Command Probe Configuration](#74-shell-command-probe-configuration) )
- **SSH**. Run a remote command via SSH and check the result. Support the bastion/jump server  ([SSH Command Probe Configuration](#75-ssh-command-probe-configuration))
- **TLS**. Ping the remote endpoint, can probe for revoked or expired certificates ( [TLS Probe Configuration](#76-tls-probe-configuration) )
- **Host**. Run an SSH command on a remote host and check the CPU, Memory, and Disk usage. ( [Host Load Probe](#77-host-resource-usage-probe-configuration) )
- **Client**. Currently, support the following native client. Support the mTLS. ( refer to: [Native Client Probe Configuration](#78-native-client-probe-configuration) )
  - **MySQL**. Connect to the MySQL server and run the `SHOW STATUS` SQL.
  - **Redis**. Connect to the Redis server and run the `PING` command.
  - **Memcache**. Connect to a Memcache server and run the `version` command or check based on key/value checks.
  - **MongoDB**. Connect to MongoDB server and just ping server.
  - **Kafka**. Connect to Kafka server and list all topics.
  - **PostgreSQL**. Connect to PostgreSQL server and run `SELECT 1` SQL.
  - **Zookeeper**. Connect to Zookeeper server and run `get /` command.

  Most of the clients support the additional validity check of data pulled from the service (such as checking a redis or memcache key for specific values). Check the documentation of the corresponding client for details on how to enable.

## 1.2 Initial Fire Up

On application startup, the configured probes are scheduled for their initial fire up based on the following criteria:

-  Less than or equal to 60 total probers exist: the delay between initial prober fire-up is `1 second`
-  More than 60 total probers exist: the startup is scheduled based on the following equation `timeGap = DefaultProbeInterval / numProbes`

> **Note**:
>
> **If multiple probes using the same name then this could lead to corruption of the metrics data and/or the behavior of the application in non-deterministic way.**


# 2. Notification
EaseProbe supports a variety of notifications. The notifications are **Edge-Triggered**, this means that these notifications are triggered when the status changes.

Each notification is identified by the delivery it supports (eg `slack`), a unique name (across all notifies in the configuration file) and (optionally) the notify specific parameters.

And please be aware that the following configuration:

1) Setting the environment variables `$HTTP_PROXY` & `$HTTPS_PROXY` allows for configuring the proxy settings for all HTTP related webhook notifications such as discord, slack, telegram etc.

    ```shell
    export HTTPS_PROXY=socks5://127.0.0.1:1080
    ```

2) All of the notifications support the `dry`, `timeout`, and `retry` optional configuration parameters. For example:

    ```YAML
    notify:
      - name: "slack"
        webhook: "https://hooks.slack.com/services/xxxxxx"
        dry: true # dry notification, print the Discord JSON in log(STDOUT)
        timeout: 20s # the timeout send out notification, default: 30s
        retry: # somehow the network is not good and needs to retry.
          times: 3 # default: 3
          interval: 10s # default: 5s
    ```
3) We can configure the general notification settings in the `notify` section of the configuration file.

    The following configuration is effective for all notification, unless the notification has its own configuration.

    ```yaml
    settings:
      notify:
        dry: true # Global settings for dry run, default is false
        retry: # the retry setting to send the notification
          times: 5 # retry times, default is 3
          interval: 10s # retry interval, default is 5s
    ```

For a complete list of examples using all the notifications please check the [Notification Configuration](#79-notification-configuration) section.

## 2.1 Slack
This notification method utilizes the Slack webhooks to deliver status updates as messages.

The plugin supports the following parameters:
 - `name`: A unique name for this notification endpoint
 - `webhook`: The URL for the webhook

Example:
```YAML
# Notification Configuration
notify:
  slack:
    - name: "MegaEase#Alert"
      webhook: "https://hooks.slack.com/services/........../....../....../"
```

For more details you can visit the [Slack Webhooks API](https://api.slack.com/messaging/webhooks)

## 2.2 Discord
This notification method utilizes the Discord webhooks to deliver status updates as messages.

The plugin supports the following parameters:
 - `name`: A unique name for this notification endpoint
 - `webhook`: The URL for the webhook

Example:
```YAML
# Notification Configuration
notify:
  discord:
    - name: "MegaEase#Alert"
      webhook: "https://discord.com/api/webhooks/...../....../"
```

For more details you can visit the [Discord Intro to Webhooks](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks)

## 2.3 Telegram
This notification method utilizes a Telegram Bot to deliver status updates as messages to a Channel or Group.

The plugin supports the following parameters:
 - `name`: A unique name for this notification endpoint
 - `token`: The authorization token for the bot
 - `chat_id`: The Telegram Channel or Group ID

Example:
```YAML
# Notification Configuration
notify:
  telegram:
    - name: "MegaEase Alert Group"
      token: 1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZ # Bot Token
      chat_id: -123456789 # Channel / Group ID
```

## 2.4 Teams
This notification method utilizes the Microsoft Teams connectors to deliver status updates as messages.

Support the  notification.

The plugin supports the following parameters:
 - `name`: A unique name for this notification endpoint
 - `webhook`: The URL for the webhook

Example:
```YAML
# Notification Configuration
notify:
  teams:
      - name: "teams alert service"
        webhook: "https://outlook.office365.com/webhook/a1269812-6d10-44b1-abc5-b84f93580ba0@9e7b80c7-d1eb-4b52-8582-76f921e416d9/IncomingWebhook/3fdd6767bae44ac58e5995547d66a4e4/f332c8d9-3397-4ac5-957b-b8e3fc465a8c"
```
For more details you can visit:
   - [Microsoft Teams Create and send messages](https://docs.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/connectors-using?tabs=cURL#setting-up-a-custom-incoming-webhook)
   - [Microsoft Teams actionable messages](https://docs.microsoft.com/en-us/outlook/actionable-messages/send-via-connectors)

## 2.5 Email
This notification method utilizes an SMTP server to deliver status updates as mail messages.

The plugin supports the following parameters:
 - `name`: A unique name for this notification endpoint
 - `server`: The server hostname or IP and port that accepts SMTP messages
 - `username`: A username to authenticate to the remote mail server
 - `passord`: The password
 - `to`: List of email addresses, separated by **`;`**, for the notification messages to be send

Example:
```YAML
# Notification Configuration
notify:
  email:
    - name: "DevOps Mailing List"
      server: smtp.email.example.com:465
      username: user@example.com
      password: ********
      to: "user1@example.com;user2@example.com"
```

## 2.6 AWS SNS
Support AWS Simple Notification Service.

The plugin supports the following parameters:
 - `name`: A unique name for this notification endpoint
 - `region`: The region to use
 - `arn`: The ARN
 - `endpoint`: The endpoint URL
 - `credential`: Credential section
   - `id`: The AWS ID
   - `key`: The AWS authorization key

Example:
```YAML
# Notification Configuration
notify:
  aws_sns:
    - name: AWS SNS
      region: us-west-2
      arn: arn:aws:sns:us-west-2:298305261856:xxxxx
      endpoint: https://sns.us-west-2.amazonaws.com
      credential:
        id: AWSXXXXXXXID
        key: XXXXXXXX/YYYYYYY
```

## 2.7 WeChat Work
Support Enterprise WeChat Work notification.

The plugin supports the following parameters:
 - `name`: A unique name for this notification endpoint
 - `webhook`: The URL for the webhook

Example:
```YAML
# Notification Configuration
notify:
  wecom:
    - name: "wecom alert service"
      webhook: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=589f9674-a2aa-xxxxxxxx-16bb6c43034a" # wecom robot webhook
```

## 2.8 DingTalk
Support the DingTalk notification.

The plugin supports the following parameters:
 - `name`: A unique name for this notification endpoint
 - `webhook`: The URL for the webhook
 - `secret`: Optionally a secret key to use

Example:
```YAML
# Notification Configuration
notify:
  dingtalk:
    - name: "dingtalk alert service"
      webhook: "https://oapi.dingtalk.com/robot/send?access_token=xxxx"
      secret: "" # sign secret if set
```

## 2.9 Lark
Support the Lark (Feishu) notification.

The plugin supports the following parameters:
 - `name`: A unique name for this notification endpoint
 - `webhook`: The URL for the webhook

Example:
```YAML
# Notification Configuration
notify:
  lark:
    - name: "lark alert service"
      webhook: "https://open.feishu.cn/open-apis/bot/v2/hook/d5366199-xxxx-xxxx-bd81-a57d1dd95de4"
```

## 2.10 SMS
Support SMS notification with multiple SMS service providers

- [Twilio](https://www.twilio.com/sms)
- [Vonage(Nexmo)](https://developer.vonage.com/messaging/sms/overview)
- [YunPian](https://www.yunpian.com/doc/en/domestic/list.html)

The plugin supports the following parameters:
 - `name`: A unique name for this notification endpoint
 - `provider`: SMS provider to use: **`yunpian`**, **`twilio`**, **`nexmo`**
 - `key`: An API key to use for the SMS service
 - `mobile`: The list of mobile numbers, separated by **`,`**, to send the notification
 - `sign`: Sign name, needed to register; usually the brand name

Example:
```YAML
# Notification Configuration
notify:
  sms:
    - name: "sms alert service"
      provider: "yunpian"
      key: xxxxxxxxxxxx # yunpian apikey
      mobile: 123456789,987654321 # mobile phone number, multiple phone number joint by `,`
      sign: "xxxxxxxx" # need to register; usually brand name
```

## 2.11 Log
Write notifications into a log file or send through syslog.

The plugin supports the following parameters:
 - `name`: A unique name for this notification endpoint
 - `file`: The path to the file that will hold the logs or **`syslog`**
 - `host`: Optionally the hostname or IP and port that your syslog server accepts messages, _syslog delivery is **NOT** supported on Windows hosts_
 - `network`: Optionally the network transport to use **`tcp`**, **`udp`**

**NOTE:**
> Windows platforms do not support syslog as notification method

Example:
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
```

## 2.12 Shell
Run a shell command to notify the result. (see [example](resources/scripts/notify/notify.sh))

The plugin supports the following parameters:
 - `name`: A unique name for this notification endpoint
 - `cmd`: The command to execute
 - `args`: Optional list of arguments as child items (see example below)
 - `env`: Optional list of environment variables to set when he command is executed (see example below)

Example:
```YAML
# Notification Configuration
notify:
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

## 2.13 RingCentral
This notification method utilizes the RingCentral webhooks to deliver status updates as messages.

The plugin supports the following parameters:
 - `name`: A unique name for this notification endpoint
 - `webhook`: The URL for the webhook

Example:
```YAML
# Notification Configuration
notify:
  ringcentral:
    - name: "MegaEase#Alert"
      webhook: "https://hooks.ringcentral.com/webhook/v2/.........."
```

# 3. Report

## 3.1 SLA Report Notification

EaseProbe supports minutely, hourly, daily, weekly, or monthly SLA reports.

```YAML
settings:
# SLA Report schedule
sla:
    #  minutely, hourly, daily, weekly (Sunday), monthly (Last Day), none
    schedule: "weekly"
    # UTC time, the format is 'hour:min:sec'
    time: "23:59"
```

## 3.2 SLA Live Report

You can query the SLA Live Report

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

  Refer to the [Global Setting Configuration](#710-global-setting-configuration) to see how to configure the access log.


## 3.3 SLA Data Persistence

EaseProbe would save the SLA statistics data on the disk.

The SLA data would be persisted in `$CWD/data/data.yaml` by default. If you want to configure the path, you can do it in the `settings` section.

When EaseProbe starts, it looks for the location of `data.yaml` and if found, loads the file and removes any probes that are no longer present in the configuration file. Setting a value of `"-"` for `data:` disables SLA persistence (eg `data: "-"`).

```YAML
settings:
sla:
    # SLA data persistence file path.
    # The default location is `$CWD/data/data.yaml`
    data: /path/to/data/file.yaml
```

For more information, please check the [Global Setting Configuration](#710-global-setting-configuration)


# 4. Channel

## 4.1 Overview

The Channel is used for connecting the Probers and the Notifiers. It can be configured for every Prober and Notifier.

This feature could help you group the Probers and Notifiers into a logical group.

> **Note**:
>
> If no Channel is defined on a probe or notify entry, then the default channel will be used. The default channel name is `__EaseProbe_Channel__`
>
> EaseProbe versions prior to v1.5.0, do not have support for the `channel` feature

## 4.2 Examples

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

# 5. Administration

There are some administration configuration options:

## 5.1 PID file

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

## 5.2 Log file Rotation

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

# 6. Prometheus Metrics Exporter

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

- Ping Probe
  - `sent`: Number of sent packets
  - `recv`: Number of received packets
  - `loss`: Packet loss percentage
  - `min_rtt`: Minimum round-trip time in milliseconds
  - `max_rtt`: Maximum round-trip time in milliseconds
  - `avg_rtt`: Average round-trip time in milliseconds
  - `stddev_rtt`: Standard deviation of round-trip time in milliseconds

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

![](./grafana.demo.png)

Refer to the [Global Setting Configuration](#710-global-setting-configuration) for further details on how to configure the HTTP server.


# 7. Configuration

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
- `failure` - number of consecutive failed probes needed to determine the status down, default: 1
- `success` - number of consecutive successful probes needed to determine the status up, default: 1

## 7.1 HTTP Probe Configuration

### 7.1.1 Basic HTTP Configuration

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
```

### 7.1.2 Complete HTTP Configuration

```yaml
http:
  # A completed HTTP Probe configuration
  - name: Special Website
    url: https://megaease.cn
    # Proxy setting, support sock5, http, https, for example:
    #   proxy: http://proxy.server:8080
    #   proxy: socks5://localhost:1085
    #   proxy: https://user:password@proxy.example.com:443
    # Also support `HTTP_PROXY` & `HTTPS_PROXY` environment variables
    proxy: http://proxy.server:8080
    # Request Method
    method: GET
    # Request Header
    headers:
      User-Agent: Customized User-Agent # default: "MegaEase EaseProbe / v1.6.0"
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
    regex: false # if true, the contain and not_contain will be treated as regular expression. default: false
    eval: # eval is a expression evaluation for HTTP response message
      doc: XML # support  XML, JSON, HTML, TEXT.
      expression: "x_time('//feed/updated') > '2022-07-01'" # the expression to evaluate.
    # configuration
    timeout: 10s # default is 30 seconds
```

> **Note**:
>
> The Regular Expression supported refer to https://github.com/google/re2/wiki/Syntax
> The XPath only supported 1.0/2.0) syntax. refer to https://www.w3.org/TR/xpath/, and the library is https://github.com/antchfx/xpath

### 7.1.3 Expression Evaluation

HTTP Probe supports two type of expression evaluation.

- `XML`, `JSON`, `HTML `: using the **XPath(1.0/2.0)** to extract the value
- `TEXT` : using the **Regression Expression** to extract the value

And the configuration can be two types as below:

**1) Variable Definition**

```yaml
  eval:
    doc: XML # support  XML, JSON, HTML, TEXT
    expression: "updated > '2022-07-01'"
    variables: # variables definition
        - name: updated # variable name
            type: time # variable type, support `int`, `float`, `bool`, `time` and `duration`.
            query: "//feed/updated" # the XPath query to get the variable value.
```

**2) Build-in XPath function Expression Evaluation**

you can just use the XPath build-in function in expression so simplify the configuration.

```yaml
  eval:
    doc: XML # support  XML, JSON, HTML, TEXT.
    expression: "x_time('//feed/updated') > '2022-07-01'" # the expression to evaluate.
```

Currently, EaseProbe supports the following XPath functions:
- `x_str` - get the string value from the XPath/RegExp query result.
- `x_int` - get the integer value from the XPath/RegExp query result.
- `x_float` - get the float value from the XPath/RegExp query result.
- `x_time` - get the time value from the XPath/RegExp query result.
- `x_duration` - get the duration value from the XPath/RegExp query result.

**3) Build-in Functions**

Currently, EaseProbe supports the following build-in functions:

- `strlen` - get the string length.
- `now` - get the current time.
- `duration` - get the duration value.


For examples:

check the `time` from response is 5 seconds later than the current time.

```yaml
eval:
   doc: HTML
   expression: "now() - x_time('//div[@id=\\'time\\']') > 5"
```


Check the duration from response is less than 1 second.

```yaml
eval:
    doc: HTML
    expression: "duration(rt) < duration('1s')"
    variables:
        - name: rt # variable name `rt` will be used in expression.
            type: duration # variable type is `duration`
            query: "//div[@id=\\'time\\']" # the XPath query the value.
```
Or

```yaml
eval:
    doc: HTML
    expression: "x_duration('//div[@id=\\'resp_time\\']') < duration('1s')"
```


**4) XPath Syntax Example**

Considering we have the following response:

```json
{
    "company": {
        "name": "MegaEase",
        "person": [{
                "name": "Bob",
                "email": "bob@example.com",
                "age": 35,
                "salary": 35000.12,
                "birth": "1984-10-12",
                "work": "40h",
                "fulltime": true
            },
            {
                "name": "Alice",
                "email": "alice@example.com",
                "age": 25,
                "salary": 25000.12,
                "birth": "1985-10-12",
                "work": "30h",
                "fulltime": false
            }
        ]
    }
}
```
Then, the extraction syntax as below:

```
"//name"                                        ==>  "MegaEase"
"//company/name"                                ==>  "MegaEase"
"//email"                                       ==>  "bob@example.com"
"//company/person/*[1]/name"                    ==>  "Bob"
"//company/person/*[2]/emai                     ==>  "alice@example.com"
"//company/person/*[last()]/name"               ==>  "Alice"
"//company/person/*[last()]/age"                ==>  "25"
"//company/person/*[salary=25000.12]/salary"    ==>  "25000.12"
"//company/person/*[name='Bob']/birth"          ==>  "1984-10-12"
"//company/person/*[name='Alice']/work"         ==>  "30h"
"//*/email[contains(.,'bob')]"                  ==>  "bob@example.com"
"//work",                                       ==>  "40h"
"//person/*[2]/fulltime"                        ==>  "false"
```

**5) Regression Expression Syntax Examples**

Considering we have the following response:

`name: Bob, email: bob@example.com, age: 35, salary: 35000.12, birth: 1984-10-12, work: 40h, fulltime: true`

Then, the extraction syntax as below:

```
"name: (?P<name>[a-zA-Z0-9 ]*)"           ==>  "Bob"
"email: (?P<email>[a-zA-Z0-9@.]*)"        ==>  "bob@example.com"
"age: (?P<age>[0-9]*)"                    ==>  "35"
"age: (?P<age>\\d+)"                      ==>  "35"
"salary: (?P<salary>[0-9.]*)"             ==>  "35000.12"
"salary: (?P<salary>\\d+\\.\\d+)"         ==>  "35000.12"
"birth: (?P<birth>[0-9-]*)"               ==>  "1984-10-12"
"birth: (?P<birth>\\d{4}-\\d{2}-\\d{2})"  ==>  "1984-10-12"
"work: (?P<work>\\d+[hms])"               ==>  "40h"
"fulltime: (?P<fulltime>true|false)"      ==>  "true"
```
> Notes
>
> Checking the unit test case in [`eval`](./eval/) package you can find more examples.

## 7.2 TCP Probe Configuration

```YAML
# TCP Probe Configuration
tcp:
  - name: SSH Service
    host: example.com:22
    timeout: 10s # default is 30 seconds
    interval: 2m # default is 60 seconds
    proxy: socks5://proxy.server:1080 # Optional. Only support socks5.
                                      # Also support the `ALL_PROXY` environment.
  - name: Kafka
    host: kafka.server:9093
```

## 7.3 Ping Probe Configuration

```YAML
ping:
  - name: Localhost
    host: 127.0.0.1
    count: 5 # number of packets to send, default: 3
    lost: 0.2 # 20% lost percentage threshold, mark it down if the loss is greater than this, default: 0
    privileged: true # if true, the ping will be executed with icmp, otherwise use udp, default: false
    timeout: 10s # default is 30 seconds
    interval: 2m # default is 60 seconds
```

## 7.4 Shell Command Probe Configuration

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
    not_contain: "failure" # response body must NOT contain this string, if it does the probe is considered failed.
    regex: false # if true, the `contain` and `not_contain` will be treated as regular expression. default: false

  # Run Zookeeper command `stat` to check the zookeeper status
  - name: Zookeeper (Local)
    cmd: "/bin/sh"
    args:
      - "-c"
      - "echo stat | nc 127.0.0.1 2181"
    contain: "Mode:"
```

> **Note**:
>
> The Regular Expression supported refer to https://github.com/google/re2/wiki/Syntax

## 7.5 SSH Command Probe Configuration

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
      passphrase: xxxxxxx  # PrivateKey password
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
      not_contain: "failure" # response body must NOT contain this string, if it does the probe is considered failed.
      regex: false # if true, the contain and not_contain will be treated as regular expression. default: false

    # Check the process status of `Kafka`
    - name:  Kafka (GCP)
      bastion: gcp         #  ◄------ bastion host id
      host: 172.10.1.100:22
      username: ubuntu
      key: /path/to/private.key
      cmd: "ps -ef | grep kafka"
```
> **Note**:
>
> The Regular Expression supported refer to https://github.com/google/re2/wiki/Syntax

## 7.6 TLS Probe Configuration

TLS ping to remote endpoint, can probe for revoked or expired certificates

```YAML
tls:
- name: expired test
    host: expired.badssl.com:443
    proxy: socks5://proxy.server:1080 # Optional. Only support socks5.
                                    # Also support the `ALL_PROXY` environment.
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

## 7.7 Host Resource Usage Probe Configuration

The host resource usage probe allows for collecting information and alerting when certain resource utilization thresholds are exceeded.

The resources currently monitored include CPU, memory and disk utilization. The probe status is considered as `down` when any value exceeds its defined threshold.

> **Note**:
> - The host running EaseProbe needs the following commands to be installed on the remote system that will be monitored: `top`, `df`, `free`, `awk`, `grep`, `tr`, and `hostname` (check the [source code](./probe/host/host.go) for more details on this works and/or modify its behavior).
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

## 7.8 Native Client Probe Configuration

Native Client probe uses the native GO SDK to communicate with the remote endpoints. Additionally to simple connectivity checks, you can also define key and data validity checks for EaseProbe, it will query for the given keys and verify the data stored on each service.

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
      "test:employee" : '{"name":"Hao Chen"}' # find the employee with name "Hao Chen"
      "test:product" : '{"name":"EaseProbe"}' # find the product with name "EaseProbe"

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
    data: # Optional, check the specific value in the path
      "/path/to/key": "value" # Check that the value of the `/path/to/key` is "value"
    # mTLS - Optional
    ca: /path/to/file.ca
    cert: /path/to/file.crt
    key: /path/to/file.key
```


## 7.9 Notification Configuration

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
      secret: "" # sign secret if set
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
  #  - EASEPROBE_TIME: time of probe time(formatted by timeformat configured in settings section)
  #  - EASEPROBE_TIMESTAMP: timestamp of probe time
  #  - EASEPROBE_MESSAGE: probe message
  # and offer two formats of string
  #  - EASEPROBE_JSON: the JSON format
  #  - EASEPROBE_CSV: the CSV format
  # The CSV format would be set for STDIN for the shell command.
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

> **Note**:
>
> 1) Setting the environment variables `$HTTP_PROXY` & `$HTTPS_PROXY` allows for configuring the proxy settings for all HTTP related webhook notifications such as discord, slack, telegram etc.
>
> 2) All of the notifications support the following optional configuration parameters.
>
>     ```YAML
>     dry: true # dry notification, print the Discord JSON in log(STDOUT)
>     timeout: 20s # the timeout send out notification, default: 30s
>     retry: # somehow the network is not good and needs to retry.
>       times: 3 # default: 3
>       interval: 10s # default: 5s
>     ```


## 7.10 Global Setting Configuration

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
    #  minutely, hourly, daily, weekly (Sunday), monthly (Last Day), none
    schedule : "daily"
    # the time to send the SLA report. Ignored on hourly and minutely schedules
    # - the format is 'hour:min:sec'.
    # - the timezone can be configured by `settings.timezone`, default is UTC.
    time: "23:59"
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
    failure: 2 # number of consecutive failed probes needed to determine the status down, default: 1
    success: 1 # number of consecutive successful probes needed to determine the status up, default: 1


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
# 8. Tools

## 8.1 EaseProbe JSON Schema

We have a JSON schema that can be used to validate your EaseProbe configuration. The schema can be found at [resources/schema.json](https://raw.githubusercontent.com/megaease/easeprobe/main/resources/schema.json).

The schema file can be generated at any time by running the following command:

```bash
$ easeprobe -j > schema.json
```

In order to use the schema with VSCode for validating your configuration, you need to install the [YAML extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) and add the following configuration to your `settings.json` file:

```json
{
  "yaml.schemas": {
    "https://raw.githubusercontent.com/megaease/easeprobe/main/resources/schema.json": [
      "/path/to/config.yaml"
    ]
  }
}
```
