# EaseProbe Roadmap
EaseProbe plans to align closely with the Cloud Native ecosystem in providing service probes and notifications.

- [EaseProbe Roadmap](#easeprobe-roadmap)
  - [Product Principles](#product-principles)
  - [Features](#features)
    - [Probe specific](#probe-specific)
    - [Notify specific](#notify-specific)
  - [Roadmap 2022](#roadmap-2022)
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

## Roadmap 2022
Some of the features that are planned/considered for 2022 can be broken down into three categories: *General*, *Probe*, *Notify*


### General
* Work on detailed documentation
* support for common daemon features
  * [x] megaease/easeprobe#75 add `daemon()` & `/var/run/easeprobe.pid` support
  * [x] megaease/easeprobe#75 add SIGHUP, and ensure it closes and re-opens of logfile to allow for `easeprobe.log` rotation
  * add syslog support as an alternative destination instead of `easeprobe.log` eg `logfile: syslog`
* Support for common `timeformat`, use standard timezone and `strftime` conversions, eg `timezone: [UTC|local|Europe/Athens]`, `timeformat: %F %R:%S UTC`
* Add opt-out options where appropriate
  * [x] megaease/easeprobe#75 add opt-out option for `logfile` option
  * Add opt-out option for SLA data persistence `data: false`
  * Make historical data configurable `history: false` and avoid creating backups of statistics

### Probes
* integrate automatic service discovery which includes probing details
* add plain old icmp ping probe (or read next)
* shell probe command must allow allocation or not of `stdout`/`stderr` or redirection of them (eg right now we are forced to create wrapper scripts to achieve this) This will also make ping by shell usable
* Add support to define desired notification on probes???
```yaml
tcp:
  - name: Memcached
    host: 10.7.0.253:11211
    notify: "Server #Alert" # notify.discord.name
```
* Add support for host group probes (eg 1 host definition with 4 services)
```yaml
name: MyServer
  probes:
    tcp:
      host: myserver:11211
    http:
      url: https://myserver.com
```

### Notify
* export the status and notification data to open formats so it can easily be integrated with 3rd party applications such as Prometheus and Graphana
* support for notify `triggers` to help on automation operations (eg. not only send a notification message but also call an API or a shell script to assist in service recovery?)
* Improve on capabilities of discord notify (eg configurable username `Username:  global.Prog,`)
* work on clear distinction between `host`, `ssh` and `shell` (certain areas seem overlapping):
  * add support for **`host: local`** keyword to monitor self
  * check that we are OS agnostic where possible and confirm OS specific operations are abstracted (such as `daemon_linux.go`, `daemon_darwin.go` etc)
  * split checks into their own functions so that the final commands to be send can be combined based on the `config.yaml`
  * add support for custom metrics and expand thresholds accordingly eg: number of process
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
