<h1>EaseProbe</h1>

[![Go Report Card](https://goreportcard.com/badge/github.com/megaease/easeprobe)](https://goreportcard.com/report/github.com/megaease/easeprobe)
[![codecov](https://codecov.io/gh/megaease/easeprobe/branch/main/graph/badge.svg?token=L7SR8X6SRN)](https://codecov.io/gh/megaease/easeprobe)
[![Build](https://github.com/megaease/easeprobe/actions/workflows/test.yaml/badge.svg)](https://github.com/megaease/easeprobe/actions/workflows/test.yaml)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/megaease/easeprobe)](https://github.com/megaease/easeprobe/blob/main/go.mod)
[![Join MegaEase Slack](https://img.shields.io/badge/slack-megaease-brightgreen?logo=slack)](https://join.slack.com/t/openmegaease/shared_invite/zt-upo7v306-lYPHvVwKnvwlqR0Zl2vveA)


EaseProbe is a simple, standalone, and lightweight tool that can do health/status checking, written in Go.

![](docs/overview.png)

<h2>Table of Contents</h2>

- [1. Introduction](#1-introduction)
  - [1.1 Probe](#11-probe)
  - [1.2 Notification](#12-notification)
  - [1.3 Report](#13-report)
- [2. Getting Started](#2-getting-started)
  - [2.1 Build](#21-build)
  - [2.2 Configure](#22-configure)
  - [2.3 Run](#23-run)
- [3. User Manual](#3-user-manual)
- [4. Benchmark](#4-benchmark)
- [5. Contributing](#5-contributing)
- [6. Community](#6-community)
- [7. License](#7-license)


# 1. Introduction

EaseProbe is designed to do three kinds of work - **Probe**, **Notify**, and **Report**.

## 1.1 Probe

EaseProbe supports the following probing methods:

- **HTTP**. Checking the HTTP status code, Support mTLS, HTTP Basic Auth, and can set the Request Header/Body. ( [HTTP Probe Configuration](./docs/Manual.md#71-http-probe-configuration) )
- **TCP**. Just simply check whether the TCP connection can be established or not. ( [TCP Probe Configuration](./docs/Manual.md#72-tcp-probe-configuration) )
- **Shell**. Run a Shell command and check the result. ( [Shell Command Probe Configuration](./docs/Manual.md#73-shell-command-probe-configuration) )
- **SSH**. Run a remote command via SSH and check the result. Support the bastion/jump server  ([SSH Command Probe Configuration](./docs/Manual.md#74-ssh-command-probe-configuration))
- **TLS**. Ping the remote endpoint, can probe for revoked or expired certificates ( [TLS Probe Configuration](./docs/Manual.md#75-tls-probe-configuration) )
- **Host**. Run an SSH command on a remote host and check the CPU, Memory, and Disk usage. ( [Host Load Probe](./docs/Manual.md#76-host-resource-usage-probe-configuration) )
- **Client**. Currently, support the following native client. Support the mTLS. ( refer to: [Native Client Probe Configuration](./docs/Manual.md#77-native-client-probe-configuration) )
  - **MySQL**. Connect to the MySQL server and run the `SHOW STATUS` SQL.
  - **Redis**. Connect to the Redis server and run the `PING` command.
  - **Memcache**. Connect to a Memcache server and run the `version` command or check based on key/value checks.
  - **MongoDB**. Connect to MongoDB server and just ping server.
  - **Kafka**. Connect to Kafka server and list all topics.
  - **PostgreSQL**. Connect to PostgreSQL server and run `SELECT 1` SQL.
  - **Zookeeper**. Connect to Zookeeper server and run `get /` command.

## 1.2 Notification

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

Check the  [Notification Configuration](./docs/Manual.md#78-notification-configuration) to see how to configure it.

## 1.3 Report

- **SLA Report Notify**. EaseProbe would send the daily, weekly, or monthly SLA report.
- **SLA Live Report**. The EaseProbe would listen on the `0.0.0.0:8181` port by default. And you can access the Live SLA report by the following URL: HTML: `http://localhost:8181/`& JSON: `http://localhost:8181/api/v1/sla`
- **SLA Data Persistence**. The SLA data would be persisted in `$CWD/data/data.yaml` by default. If you want to configure the path, you can do it in the `settings` section.

For more information, please check the [Global Setting Configuration](./docs/Manual.md#79-global-setting-configuration)

# 2. Getting Started

You can get started with EaseProbe, by any of the following methods:
* Download the release for your platform from https://github.com/megaease/easeprobe/releases
* Use the available EaseProbe docker image `docker run -it megaease/easeprobe`
* Build `easeprobe` from sources

## 2.1 Build

Compiler `Go 1.18+` (Generics Programming Support)

Use `make` to build and produce the `easeprobe` binary file. The executable is produced under the `build/bin` directory

```shell
$ make
```
## 2.2 Configure

Read the [User Manual](./docs/Manual.md) to learn how to configure EaseProbe.

Create the configuration file - `$CWD/config.yaml`.

The following is an example of a simple configuration file to get started:

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
Or using this configuration template:
https://github.com/megaease/easeprobe/resources/config.yaml

## 2.3 Run

Running the following command for the local test

```shell
$ build/bin/easeprobe -f config.yaml
```
* `-f` configuration file or URL. Can also be achieved by setting the environment variable `PROBE_CONFIG`
* `-d` dry run. Can also be achieved by setting the environment variable `PROBE_DRY`


# 3. User Manual

Refer to - [User Manual](./docs/Manual.md)

# 4. Benchmark

Refer to - [Benchmark Report](./docs/Benchmark.md)

# 5. Contributing

If you're interested in contributing to the project, please spare a moment to read our [CONTRIBUTING Guide](./docs/CONTRIBUTING.md)


# 6. Community

- Join Slack [Workspace](https://join.slack.com/t/openmegaease/shared_invite/zt-upo7v306-lYPHvVwKnvwlqR0Zl2vveA) for requirements, issues, and development.
- [MegaEase on Twitter](https://twitter.com/megaease)

# 7. License

EaseProbe is under the Apache 2.0 license. See the [LICENSE](./LICENSE) file for details.