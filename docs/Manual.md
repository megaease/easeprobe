<h1>User Manual</h1>

EaseProbe is a simple, standalone, and lightweight tool that can do health/status checking, written in Go.

![](./overview.png)

EaseProbe has the following major modules:

- **Probe**: It is used to check the health of the service.
- **Notification**: It is used to send the Probe notification.
- **Channel**: It is used to connect the probe and the notification.
- **Report**: It is used to generate the SLA report for all probe.
- **Metrics**: It is used to export the metrics data to Prometheus.


<h2>Table of Contents</h2>

- [1. Probe](#1-probe)
  - [1.1 Overview](#11-overview)
    - [1.1.1 General Settings](#111-general-settings)
    - [1.1.2 Alerting Interval](#112-alerting-interval)
      - [1.1.2.1 Regular Strategy](#1121-regular-strategy)
      - [1.1.2.2 Incremental Strategy](#1122-incremental-strategy)
      - [1.1.2.3 Exponential Strategy](#1123-exponential-strategy)
    - [1.1.3 Initial Fire Up](#113-initial-fire-up)
  - [1.2 HTTP](#12-http)
    - [1.2.1 Basic Configuration](#121-basic-configuration)
    - [1.2.2 Complete Configuration](#122-complete-configuration)
    - [1.2.3 Expression Evaluation](#123-expression-evaluation)
  - [1.3 TCP](#13-tcp)
  - [1.4 Ping](#14-ping)
  - [1.5 Shell](#15-shell)
  - [1.6 SSH](#16-ssh)
  - [1.7 TLS](#17-tls)
  - [1.8 Host](#18-host)
  - [1.9 Native Client](#19-native-client)
    - [1.9.1 Redis](#191-redis)
    - [1.9.2 MySQL](#192-mysql)
    - [1.9.3 MongoDB](#193-mongodb)
    - [1.9.4 Memcache](#194-memcache)
    - [1.9.5 Kafka](#195-kafka)
    - [1.9.6 PostgreSQL](#196-postgresql)
    - [1.9.7 Zookeeper](#197-zookeeper)
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
  - [6.1 General Metrics](#61-general-metrics)
  - [6.2 HTTP Probe](#62-http-probe)
  - [6.3 Ping Probe](#63-ping-probe)
  - [6.4 TLS Probe](#64-tls-probe)
  - [6.5 Shell \& SSH Probe](#65-shell--ssh-probe)
  - [6.6  Host Probe](#66--host-probe)
- [7. Configuration](#7-configuration)
  - [7.1 Probe Configuration](#71-probe-configuration)
  - [7.2 Notification Configuration](#72-notification-configuration)
  - [7.3 Global Setting Configuration](#73-global-setting-configuration)
- [8. Tools](#8-tools)
  - [8.1 EaseProbe JSON Schema](#81-easeprobe-json-schema)
  - [8.2 EaseProbe Deployment](#82-easeprobe-deployment)




# 1. Probe

EaseProbe supports these probing methods: **HTTP**, **TCP**, **TLS**, **Shell Command**, **SSH Command**, **Host Resource Usage**, and **Native Client**.

## 1.1 Overview

Each probe is identified by the method it supports (eg `http`), a unique name (across all probes in the configuration file) and the method specific parameters. For example:

```yaml
http:
  - name: Web Service
    url: http://example.com:1080
```
> **Note**:
>
> **Probe name must be unique.** if multiple probes using the same name then this could lead to corruption of the metrics data and/or the behavior of the application in non-deterministic way.

### 1.1.1 General Settings

All probes support the `timeout`, `interval`, `failure`, and `success` optional configuration parameters. For example:

```yaml
tcp:
  - name: Web Service
    host: example.com:1080
    timeout: 10s # the time out for all probes, default is 30 seconds
    interval: 2m # probe every minute for all probes, default is 60 seconds
    failure: 2 # number of consecutive failed probes needed to determine the status down, default: 1
    success: 1 # number of consecutive successful probes needed to determine the status up, default: 1
```

We can configure the general probe settings for all probes.

The following configuration is effective for all probe, unless the probe has its own configuration.

```yaml
settings:
  probe:
    timeout: 30s # the time out for all probes, default is 30 seconds
    interval: 1m # probe every minute for all probes, default is 60 seconds
    failure: 2 # number of consecutive failed probes needed to determine the status down, default: 1
    success: 1 # number of consecutive successful probes needed to determine the status up, default: 1
```

### 1.1.2 Alerting Interval

All of the Probe supports alerting interval configuration.

The alerting interval is the time EaseProbe keep sending service down alerting notification until the probe is back to normal.

The alerting interval is configured by three parameters: `strategy`, `factor` and `max`.

- `strategy` is the alerting strategy, it can be `regular`， `increment` and `exponent`. the default strategy is `regular`.
- `factor` is the alerting interval factor, it can be any positive integer. the default factor is 1.
- `max` is the maximum alerting times, it can be any positive integer. the default max is 1.

The following is an example of the alerting interval for HTTP Probe:
```yaml
  http:
    - name: Web Service
      url: http://example.com:1080
      alert:
        strategy: regular
        factor: 1
        max: 3
```

> **Note**:
>
> 1. The default strategy is regular and the max time is 1, this would be the same as the old behavior - edge trigger.
> 2. The alerting interval is aligned with the probe interval, it is triggered by the probe failure.
> 3. The alerting interval configuration can be configured in the global settings `settings.probe`, and the probe's own configuration will override the global settings.

#### 1.1.2.1 Regular Strategy

Notifications are sent at the same frequency as the probe time, continuing until the maximum number of notifications is reached.

The formula is:

`interval = factor`

For example:

```yaml
alert:
  strategy: regular
  factor: 1
  max: 10
```

- If the factor = 1, the interval is always 1.<br>
  So, the alert would be sent at failure of 1, 2，3，4，5，6，7...


- If the factor = 2, the interval is always 2.<br>
  So, the alert would be sent at failure of 1, 3, 5, 7, 9, 11, 13...

- If the factor = 3, the interval is always 3.<br>
  So, the alert would be sent at failure of 1, 4, 7, 10, 13, 16, 19...

#### 1.1.2.2 Incremental Strategy

Notifications are sent at increasing intervals until the maximum number of notifications is reached. With this strategy, the interval is increased linearly.

The formula is:

`interval = factor * (failure - 1) + 1`

For example:

```yaml
alert:
  strategy: increment
  factor: 1
  max: 10
```

- If the factor = 1, the interval is 1, 2, 3, 4, 5, 6, 7, 8, 9, 10.<br>
  So, the alert would be sent at failure of 1, 2, 4, 7, 11, 16, 22, 29, 37...

- If the factor = 2, the interval is 2, 4, 6, 8, 10, 12, 14, 16, 18, 20.<br>
  So, the alert would be sent at failure of 1, 3, 7, 13, 21, 31, 43, 57, 73...

- If the factor = 3, the interval is 3, 6, 9, 12, 15, 18, 21, 24, 27, 30.<br>
  So, the alert would be sent at failure of 1, 4, 10, 19, 31, 46, 64, 85, 109...


#### 1.1.2.3 Exponential Strategy

Notifications are sent at exponentially increasing intervals until the maximum number of notifications is reached. With this strategy, the interval is increased exponentially.

The formula is:

`interval =  failure times + factor * ( failure times - 1 )`

For example:

```yaml
alert:
  strategy: exponent
  factor: 1
  max: 10
```

- If the factor = 1, the interval is 1, 2, 4, 8, 16, 32, 64, 128...<br>
  So, the alert would be sent at failure of 1, 2, 4, 8, 16, 32, 64, 128, 256...

- If the factor = 2, the interval is 2, 6, 18, 54, 162, 486, 1458...<br>
  So, the alert would be sent at failure of 1, 3, 9, 27, 81, 243, 729, 2187...

- If the factor = 3, the interval is 3, 12, 48, 192, 768...<br>
  So, the alert would be sent at failure of 1, 4, 16, 64, 256, 1024...



### 1.1.3 Initial Fire Up

On application startup, the configured probes are scheduled for their initial fire up based on the following criteria:

-  Less than or equal to 60 total probers exist: the delay between initial prober fire-up is `1 second`
-  More than 60 total probers exist: the startup is scheduled based on the following equation `timeGap = DefaultProbeInterval / numProbes`


## 1.2 HTTP

HTTP probe using `http` identifier, it has the following features:

- Checking the HTTP status code range
- Support the response body check - JSON/XML/HTML Path checking
- Support mTLS and HTTP Basic Auth
- Can set the customized Request Header/Body.
- Support the HTTP Proxy

Setting the environment variables `$HTTP_PROXY` & `$HTTPS_PROXY` allows for configuring the proxy settings for all HTTP related probe (or you can set the `proxy` field in the probe configuration for specific probe).

```shell
export HTTPS_PROXY=socks5://127.0.0.1:1080
```

### 1.2.1 Basic Configuration

the following example configuration is a basic HTTP probe configuration, which only have `name` and `url`, the `name` is the probe name, and the `url` is the HTTP URL to be checked. If the HTTP status code is in `[0, 499]`, then the probe is considered as `UP`.

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

### 1.2.2 Complete Configuration

The following example configuration is a complete HTTP probe configuration, which has all the supported configuration parameters.

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

### 1.2.3 Expression Evaluation

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

## 1.3 TCP

TCP probe just simply check whether the TCP connection can be established or not.

The following is the configuration example, which has two TCP probes:
- **SSH Service**, it will check the TCP connection to `example.com:22` every 2 minutes with 10 second timeout via the proxy  `socks5://proxy.server:1080`.
- **Kafka**, which will check the TCP connection to `kafka.server:9093` by using the default interval and timeout settings.

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

## 1.4 Ping

Ping probe uses `ping` identifier, it just simply check whether can be pinged or not.

The following is the configuration example.

```YAML
ping:
  - name: Localhost
    host: 127.0.0.1
    count: 5 # number of packets to send, default: 3
    lost: 0.2 # 20% lost percentage threshold, mark it down if the loss is greater than this, default: 0
    privileged: true # if true, the ping will be executed with icmp, otherwise use udp, default: false (Note: On Windows platform, this must be set to True)
    timeout: 10s # default is 30 seconds
    interval: 2m # default is 60 seconds
```

## 1.5 Shell

The shell command probe uses `shell` identifier, it is used to execute a shell command and check the output.

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

## 1.6 SSH

SSH probe uses `ssh` identifier, it is similar to Shell probe.
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

## 1.7 TLS

TLS probe uses `tls` identifier, it pings to remote endpoint, can probe for revoked or expired certificates

```YAML
tls:
  - name: expired test
    host: expired.badssl.com:443
    proxy: socks5://proxy.server:1080 # Optional. Only support socks5.
                                      # Also support the `ALL_PROXY` environment.
    insecure_skip_verify: true # don't check cert validity
    expire_skip_verify: true   # don't check cert expire date
    alert_expire_before: 168h  # alert if cert expire date is before X, the value is a Duration,
                               # see https://pkg.go.dev/time#ParseDuration. example: 1h, 1m, 1s.
                               # expire_skip_verify must be false to use this feature.
    # root_ca_pem_path: /path/to/root/ca.pem # ignore if root_ca_pem is present
    # root_ca_pem: |
    #   -----BEGIN CERTIFICATE-----
  - name: untrust test
    host: untrusted-root.badssl.com:443
    # insecure_skip_verify: true # don't check cert validity
    # expire_skip_verify: true   # don't check cert expire date
    # root_ca_pem_path: /path/to/root/ca.pem # ignore if root_ca_pem is present
    # root_ca_pem: |
    #   -----BEGIN CERTIFICATE-----
```

## 1.8 Host

The host probe uses `host` identifier, it allows for collecting information and alerting when certain resource utilization thresholds are exceeded.

The resources currently monitored include CPU, memory and disk utilization. The probe status is considered as `down` when any value exceeds its defined threshold.

> **Note**:
> - The remote system to be monitored needs to have the following commands installed and available: `top`, `df`, `free`, `awk`, `grep`, `tr`, `cat` and `hostname`.
> - The disk usage check is limited to the root filesystem only with the following command `df -h /`.
> - The actual load would be divided by cpu core number, the threshold won't consider the cpu core number (requires proc filesystem support).

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
        load: # load average - Note: the actual load would be divided by cpu core number, the threshold won't consider the cpu core number.
          m1: 0.5  # 1 minute load average 0.5 (default: 0.8)
          m5: 0.9  # 5 minute load average 0.9 (default: 0.8)
          m15: 0.9 # 15 minute load average 0.9 (default: 0.8)

    # Using the default threshold
    # cpu 80%, mem 80%, disk 95% and 0.8 load average
    - name : My VPS
      host: user@example.com:22
      key: /Users/user/.ssh/id_rsa
```

## 1.9 Native Client

Native Client probe uses `client` identifier, it uses the native GO SDK to communicate with the remote endpoints. Additionally to simple connectivity checks, you can also define key and data validity checks for EaseProbe, it will query for the given keys and verify the data stored on each service.

Currently, support the following native client
  - **MySQL**. Connect to the MySQL server and run the `SHOW STATUS` SQL.
  - **Redis**. Connect to the Redis server and run the `PING` command.
  - **Memcache**. Connect to a Memcache server and run the `version` command or check based on key/value checks.
  - **MongoDB**. Connect to MongoDB server and just ping server.
  - **Kafka**. Connect to Kafka server and list all topics.
  - **PostgreSQL**. Connect to PostgreSQL server and run `SELECT 1` SQL.
  - **Zookeeper**. Connect to Zookeeper server and run `get /` command.

The following is an example for all native client probe configuration:

### 1.9.1 Redis

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

```

### 1.9.2 MySQL

```YAML
client:
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
```

### 1.9.3 MongoDB

```YAML
client:
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
```

### 1.9.4 Memcache

```YAML
client:
  - name: Memcache Native Client (local)
    driver: "memcache"
    host: "localhost:11211"
    timeout: 5s
    data:         # Optional
      key: val    # Check that key exists and its value is val
      "namespace:key": val # Namespaced keys enclosed in "
```

### 1.9.5 Kafka

```YAML
client:
  - name: Kafka Native Client (local)
    driver: "kafka"
    host: "localhost:9093"
    # mTLS - Optional
    ca: /path/to/file.ca
    cert: /path/to/file.crt
    key: /path/to/file.key
```

### 1.9.6 PostgreSQL

```YAML
client:
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
```

### 1.9.7 Zookeeper

```YAML
client:
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
     channels : [ Mgmt_Channel ]
shell:
   - name: probe B
     channels: [ Dev_Channel, QA_Channel ]
notify:
   - discord: Discord
     channels: [ Mgmt_Channel, Dev_Channel ]
   - email: Gmail
     channels: [ QA_Channel ]
```

Then, we will have the following diagram

```
┌─────────┐        ┌──────────────┐
│ Probe A ├───────►│ Mgmt_Channel ├──────────┐
└─────────┘        └──────────────┘          │
   http                                      │
                                             │
                   ┌──────────────┐     ┌────▼────┐
              ┌───►│  Dev_Channel ├────►│ Discord │
              │    └──────────────┘     └─────────┘
┌─────────┐   │
│ Probe B ├───┤
└─────────┘   │
   shell      │    ┌──────────────┐     ┌─────────┐
              └───►│  QA_Channel  ├────►│  Gmail  │
                   └──────────────┘     └─────────┘
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

If you want to disable the PID file, you can set it to "-" or "".

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

The following snapshot is the Grafana panel for host CPU metrics

![](./grafana.demo.png)

Refer to the [Global Setting Configuration](#710-global-setting-configuration) for further details on how to configure the HTTP server.

## 6.1 General Metrics

Currently, All of the Probers support the following metrics:

  - `total`: the total number of probes
  - `total_time`: the total time(seconds) of status up or down
  - `duration`: Probe duration in milliseconds
  - `status`: Probe status
  - `SLA`: Probe SLA percentage

And the different Probers have its own metrics.

## 6.2 HTTP Probe

The HTTP probe supports the following metrics:

  - `status_code`: HTTP status code
  - `content_len`: HTTP content length
  - `dns_duration`: DNS duration in milliseconds
  - `connect_duration`: TCP connection duration in milliseconds
  - `tls_duration`: TLS handshake duration in milliseconds
  - `send_duration`: HTTP send duration in milliseconds
  - `wait_duration`: HTTP wait duration in milliseconds
  - `transfer_duration`: HTTP transfer duration in milliseconds
  - `total_duration`: HTTP total duration in milliseconds

## 6.3 Ping Probe

The Ping probe supports the following metrics:

  - `sent`: Number of sent packets
  - `recv`: Number of received packets
  - `loss`: Packet loss percentage
  - `min_rtt`: Minimum round-trip time in milliseconds
  - `max_rtt`: Maximum round-trip time in milliseconds
  - `avg_rtt`: Average round-trip time in milliseconds
  - `stddev_rtt`: Standard deviation of round-trip time in milliseconds
  - `privileged`: Whether to use ICMP (`true`) or UDP (`false`) based ping probes, default `false`.

Please note that `privileged: true` requires administrative privileges such as `root` (for more details see https://github.com/prometheus-community/pro-bing#supported-operating-systems)

## 6.4 TLS Probe

The TLS probe supports the following metrics:

  - `earliest_cert_expiry`: last TLS chain expiry in timestamp seconds
  - `last_chain_expiry_timestamp_seconds`: earliest TLS cert expiry in Unix time

## 6.5 Shell & SSH Probe

The Shell & SSH probe supports the following metrics:

  - `exit_code`: exit code of the command
  - `output_len`: length of the output


## 6.6  Host Probe

The Host probe supports the following metrics:

  - `cpu`: CPU usage in percentage
  - `memory`: memory usage in percentage
  - `disk`: disk usage in percentage
  - `load`: load average for `m1`, `m5`, and `m15`


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

EaseProbe supports multiple configuration files, with the use of the command line flag `-f` which indicates the directory of the configuration files.

```shell
easeprobe -f /path/to/conf
```


The configuration file should be versioned, the version should be aligned with the EaseProbe binary version.

```yaml
version: v1.5.0
```

You can find the full configuration template [here](../resources/config.yaml)

## 7.1 Probe Configuration

```YAML
# --------------------- HTTP Probe Configuration ---------------------
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
    failure: 2 # number of consecutive failed probes needed to determine the status down, default: 1
    success: 1 # number of consecutive successful probes needed to determine the status up , default: 1


# --------------------- TCP Probe Configuration ---------------------

tcp:
  - name: SSH Service
    host: example.com:22
    timeout: 10s # default is 30 seconds
    interval: 2m # default is 60 seconds
    proxy: socks5://proxy.server:1080 # Optional. Only support socks5.
                                      # Also support the `ALL_PROXY` environment.
    failure: 2 # number of consecutive failed probes needed to determine the status down, default: 1
    success: 1 # number of consecutive successful probes needed to determine the status up, default: 1

--------------------- Ping Probe Configuration ---------------------

ping:
  - name: Localhost
    host: 127.0.0.1
    count: 5 # number of packets to send, default: 3
    lost: 0.2 # 20% lost percentage threshold, mark it down if the loss is greater than this, default: 0
    privileged: true # if true, the ping will be executed with icmp, otherwise use udp, default: false (Note: On Windows platform, this must be set to True)
    timeout: 10s # default is 30 seconds
    interval: 2m # default is 60 seconds

# --------------------- Shell Probe Configuration ---------------------

shell:
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


# --------------------- SSH Probe Configuration ---------------------

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


# --------------------- TLS Probe Configuration ---------------------

tls:
  - name: expired test
    host: expired.badssl.com:443
    proxy: socks5://proxy.server:1080 # Optional. Only support socks5.
                                      # Also support the `ALL_PROXY` environment.
    insecure_skip_verify: true # don't check cert validity
    expire_skip_verify: true   # don't check cert expire date
    alert_expire_before: 168h  # alert if cert expire date is before X, the value is a Duration,
                               # see https://pkg.go.dev/time#ParseDuration. example: 1h, 1m, 1s.
                               # expire_skip_verify must be false to use this feature.
    # root_ca_pem_path: /path/to/root/ca.pem # ignore if root_ca_pem is present
    # root_ca_pem: |
    #   -----BEGIN CERTIFICATE-----
  - name: untrust test
    host: untrusted-root.badssl.com:443
    # insecure_skip_verify: true # don't check cert validity
    # expire_skip_verify: true   # don't check cert expire date
    # root_ca_pem_path: /path/to/root/ca.pem # ignore if root_ca_pem is present
    # root_ca_pem: |
    #   -----BEGIN CERTIFICATE-----


# --------------------- Host Probe Configuration ---------------------

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
        load: # load average - Note: the actual load would be divided by cpu core number, the threshold won't consider the cpu core number.
          m1: 0.5  # 1 minute load average 0.5 (default: 0.8)
          m5: 0.9  # 5 minute load average 0.9 (default: 0.8)
          m15: 0.9 # 15 minute load average 0.9 (default: 0.8)

    # Using the default threshold
    # cpu 80%, mem 80%, disk 95% and 0.8 load average
    - name : My VPS
      host: user@example.com:22
      key: /Users/user/.ssh/id_rsa


# --------------------- Native Client Probe Configuration ---------------------

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

## 7.2 Notification Configuration

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


## 7.3 Global Setting Configuration

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
    alert: # alert interval for all probes
      strategy: "regular" # it can be "regular", "increment" or "exponent", default: "regular"
      factor: 1 # the factor of the interval, default: 1
      max: 5 # the max of the alert, default: 1


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

## 8.2 EaseProbe Deployment

EaseProbe can be deployed by Systemd, Docker, & Kubernetes.

You can find the details in [Deployment Guide](./Deployment.md)
