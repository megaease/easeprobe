# EaseProbe Roadmap
EaseProbe plans to align closely with the Cloud Native ecosystem in providing service probes and notifications.

- [EaseProbe Roadmap](#easeprobe-roadmap)
  - [Product Principles](#product-principles)
  - [Features](#features)
    - [Probe specific](#probe-specific)
    - [Notify specific](#notify-specific)
  - [Roadmap](#roadmap)
    - [General](#general)
    - [Probes](#probes)
    - [Notify](#notify)


## Product Principles
EaseProbe tries to do two things but do them extremely well - **Probe** and **Notify**.

1. **Lightweight, Standalone and KISS**. Aims to be lightweight, with as few as possible external system dependencies and follow KISS ideas.

2. **Probe and Notify**. Aims to provide the ability to easily **probe** servers, services and API endpoints reliably and **notify** on status updates.

3. **Open & Extensible**. It aims to be **Open** and follow Open standards that allow it to be integrated into **extensible-development** platforms (such as EaseMesh).

4. **Cloud Native**. It is designed to be **cloud-native** compliant. It is scalable, resilient, manageable, and observable and it's easy to integrate with cloud-native architectures of any type.

5. **Predictability & Reliability**. EaseProbe is designed to follow best practices allowing predictable and reliable operations. Even when EaseProbe fails, it aims to fail in a predictable way which allows for easy and speedy troubleshooting of problems.


## Features
The project principles, of EaseProbe' features are separated into two main categories: probe-specific & notify-specific.

### Probe specific
* HTTP for testing connectivity and validity of responses
* TCP for testing connectivity
* Shell for executing custom probe scripts
* SSH for remote ssh commands
* Host for CPU, Memory, Disk usage metrics
* Clients for Redis, MySQL, MongoDB, PostgreSQL, Kafka, Zookeeper

### Notify specific
* Mail notification
* AWS SNS notification
* Log files
* Slack
* Discord
* Telegram
* WeChat Work
* Lark
* DingTalk
* Twilio
* Nexmo
* YunPian

... with new notification backends been added constantly.

## Roadmap
Some of the features that we plan to implement in the future fall under one of these categories: *General*, *Probe*, *Notify*

### General
* [x] Work on detailed documentation (megaease/easeprobe#210)
* [x] Improve test coverage (megaease/easeprobe#128 megaease/easeprobe#127 megaease/easeprobe#119 megaease/easeprobe#118 megaease/easeprobe#117)
* Improve 3rd party integrations and supports
  * [x] megaease/easeprobe#95 Prometheus compatible metrics
* Support for common daemon features
  * [ ] ability to send daemon to background without stdout logs
  * [x] megaease/easeprobe#129 add syslog support as an alternative destination instead of `easeprobe.log` eg `log: syslog`
  * [ ] introduce a control socket for running easeprobe instance with disable or enable probes and notify endpoints (maybe something like `/var/run/easeprobe.sock` that speaks HTTP (`dockerd` & `supervisord` does something like that).
  * [x] megaease/easeprobe#75 add `daemon()` & `/var/run/easeprobe.pid` support
  * [x] megaease/easeprobe#75 add SIGHUP, and ensure it closes and re-opens of logfile to allow for `easeprobe.log` rotation
* [ ] Support for common `timeformat`, use standard timezone and `strftime` conversions, eg `timezone: [UTC|local|Europe/Athens]`, `timeformat: %F %R:%S UTC`
* [x] Support the timezone configuration (megaease/easeprobe#167)
* Add opt-out options where appropriate
  * [x] megaease/easeprobe#75 add opt-out option for `log` option
  * [x] megaease/easeprobe#92 Add opt-out option for SLA data persistence `data: false`
  * [x] megaease/easeprobe#81 Make historical data configurable `history: false` and avoid creating backups of statistics

### Probes
* [ ] add automatic service discovery which includes probing details
* [x] add plain old `icmp` ping probe (megaease/easeprobe#251)
* [ ] `shell` probe command improvements in handling stdout/stderr
* [x] megaease/easeprobe#108 add support to define destination notification channels for probe (see https://github.com/megaease/easeprobe/discussions/82)
* [ ] work on cleaner distinction between `host`, `ssh` and `shell` (certain areas seem overlapping):
  * add support for **`host: local`** keyword to monitor self
  * check that we are OS agnostic where possible and confirm OS specific operations are abstracted (such as `daemon_linux.go`, `daemon_darwin.go` etc)
  * split hardcoded commands into their own configurable functions so that the final commands to be send can be combined based on `config.yaml` settings later on
* [ ] Add support for host group probes (eg 1 host definition with 4 services)
```yaml
name: MyServer
  probes:
    tcp:
      host: myserver:11211
    http:
      url: https://myserver.com
```
* [ ] add support for custom metrics and expand thresholds accordingly eg: number of process
```yaml
host:
  servers:
    - name : server
      host: ubuntu@172.20.2.202:22
      key: /path/to/server.pem
      metrics:
        numprocs: "ps axu | wc -l"
        cpu: true
      threshold:
        numprocs: 400 # custom metric
        cpu: 0.80  # cpu usage  80%
```

### Notify
* [x] megaease/easeprobe#130 extend notify to help on automation operations (eg. not only send a notification message but also call an API or a shell script to assist in service recovery)
* [x] megaease/easeprobe#108 support for notification groups or lists (see https://github.com/megaease/easeprobe/discussions/82)
* [ ] megaease/easeprobe#79 Improve on capabilities of discord and other similar notify (such as configurable username)
