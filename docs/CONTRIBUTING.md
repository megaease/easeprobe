# CONTRIBUTING
So you've decided to actively join the development of EaseProbe!!! First let us Welcome you and **Thank You** for your interest.

The following document will try to provide some general guidelines and information about the development processes to help your on-boarding experience.

This document is constantly under change with new things been added as the project grows.

## General guidelines
The following are some guidelines that will help in getting your code contributions accepted easier and faster.

* For code changes, make sure the style complies with the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
* Write a small description for your PR, so the reviewers know what this is about and what to expect
* Make sure to address any linting or test failures your PR introduces
* If you contribute a new feature
    * update or create appropriate tests for it
    * update the README accordingly
* If you introduced new files make sure to include the license on top of each new `.go` file. Running `resources/scripts/copyright.sh` will do that for you

## How EaseProbe Works
Basically, the EaseProbe have the following major code organization.

### 1) Probe

- **Probe Interface**. Every Probe must implement the `Probe` interface, which is defined in `probe/probe.go`
- **Base Probe**. The `probe/base/base.go` helps to implement most `Probe` interface methods, like a Probe framework.
- **Concrete Probe**. Every concrete probe must do the following works
    - `Config()` - provide probe's `kind`, `tag` and other information.
    - `DoProbe()` - implement the actual code that does the probing.
- **Probe Configuration**. The probe configuration need to be added in the `Conf` struct ( `conf/conf.go` ).
- **Probe Result** The probe result and the relevant statistics are stored in a `Result` object, this is defined in `probe/result.go`

### 2) Notify

- **Notify Interface**. Every Notify must implement the `Notify` interface, which is defined in `notify/notify.go`.
- **Base Notify**. The `notify/base/base.go` help to implement most `Notify` interface methods, like a Notify framework.
- **Concrete Notify**. Every concrete notifier must do the following works
    - `Config()` - provide notify's `kind`, report format, send function and other information.
    - `SendXXXNotification(title, message)` - provide the implementation of how the notifier will perform the notification.
-  **Formation**. The formatting of the notification message, such as HTML, Markdown, JSON, Text or use a completely custom formatter. These formations are defined in `report/result.go`


### 3) Channel

The Channel is used for connect the Probe and Notify. The `channel` package organizes the Probe and Notify associations.

- **Probe**. All of the prober will be run by `runProbers()` function in `cmd/easeprobe/probe.go`
- **Notify**. The Notify would watch the Probe's Result in `channel.WatchAllEvents()` which in `channel/manager.go`.


### 4) SLA Report

The SLA Report contains the results of all Probers which include:

- **SLA Data**. SLA data (in `probe/data.go`) that are (optionally) persistent in a file.
- **Web Server**. the Web server for SLA report is in `web` package, it includes the HTML web page and the Restful API. This uses of Chi framework.
- **Report Format**. there are several formations for SLA report, the code is in `report/sla.go`
- **Scheduler**. takes care of the SLA report scheduling requirements (such as Daily/Weekly/Monthly) and can be found in `cmd/easeprobe/report.go`

### 5) Metrics (Prometheus)

EaseProbe supports exporting of probe metrics to prometheus. The package `metric/prometheus.go` help to create and register the metric.


- **Base Metrics**. We implement some common metrics for all of Probers `probe/base/metrics.go` - Total Probe Times(Counter), Probe Duration(Gauge), Status(Gauge), and SLA(Gauge).
- **Customize Metrics**. If we need to deal with the Prober specific metrics, we can do the following works
    - **Define Metrics**. just create the `metrics.go` in probe which define the metrics struct and how to create it.
    - **Initialize Metrics**. initialize the metrics in `Probe.Config()`function
    - **Export Metrics**. create a new `Probe.ExportMetrics()` function and call it in `Probe.DoProbe()` function.

Examples
- Host Probe metrics: `probe/host/metrics.go`
- HTTP Probe metrics: `probe/http/metrics.go`


## Source Code Files
* `cmd/easeprobe` This is the `main` package for the EaseProbe main function.
    * `main.go` - `main()` function to do everything.
    * `probe.go` - Initialize and configure the Probers and start up all of probers
    * `notify.go` - Initialize and configure the Notifiers.
    * `channel.go` - Initialize and configure the Channels.
    * `report.go` - SLA report.
* `conf`. This package is for EaseProbe configuration and Log file.
    * `conf.go` - EaseProbe configuration file processing.
    * `log.go` - EaseProbe application log and access log.
* `daemon`. This package is about EaseProbe daemon operations, such as: create the pid file.
    * `daemon.go` -  General operations for all platforms
    * `daemon_windows.go` - For Windows platform
    * `daemon_linux.go` - For Linux platform
    * `daemon_unix.go` - For all non-Windows and non-Linux platform such as: macOS & OpenBSD
* `global`. This package defines the Global variables, structures and functions.
    * `global.go` - Default Value, Retry, TLS structure, Working Directory...
    * `easeprobe.go` - Holds the EaseProbe Name, Icon, Version, Host information.
    * `probe.go` - The common configuration for all Probers.
    * `notify.go` - The common configuration for all Notifiers.
* `metric`. This package is for external metric report, like prometheus.
    * `prometheus.go` - The wrapper of prometheus metrics export.
* `channel`. Channel object and management
    * `channel.go` - The channel object.
    * `manager.go` - manage all of channels
* `eval`. This package is for evaluating the expression.
    * `eval.go` - The expression evaluation.
    * `extract.go` - Extract the value from the document.(using XPath or Regex)
    * `types.go` - The value's type.
* `notify`. All of notifiers
    * `notify.go` - The notify interface definition.
    * `base/` - The notify framework for all notify.
* `probe`. All of probers.
    * `probe.go` - the probe interface definition.
    * `base/` - The probe framework for all notification
    * `result.go` - The probe result object, includes all of the probe result, like status, start probe time, round trip time, statistics, etc.
    * `data.go` - The probe result persistent.
    * `status.go` - The probe status definition: `StatusUp`, `StatusDown`, `StatusUnknown`, ... etc.
* `report`. SLA Report formation.
    * `types.go` - The format of report
    * `sla.go` - SLA report for all of format.
    * `result.go`- Probe Result for all of format.
* `web`. This package is the web server for EaseProbe SLA report and API
    * `server.go` - Web Server
    * `log.go` - Access Log
​
## Makefile
The project provides a `Makefile` with the following targets
* `build`: compile the project and produce the EaseProbe binary
* `test`: runs `go test`
* `docker`: builds the docker images for EaseProbe (using `resources/Dockerfile`)
* `clean`: cleans up files created during build
​
​
## Github Actions
The project currently has 2 Github Workflows that are responsible for checking the code and assisting in releases of new versions.

- `code.yaml` performs code linting checks on files
- `license.yaml` checks for license existence in source code files
- `test.yaml` performs unit and coverage testing
- `release.yaml` would do release work, build the binaries and docker images

**Note**: `goreleaser` uses the configuration file at `.goreleaser.yaml`, and builds the release using the docker file at `resources/Dockerfile.goreleaser`
