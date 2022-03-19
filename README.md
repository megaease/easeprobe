# Ease Probe

Ease Probe is a simple, standalone, and lightWeight tool that can do health/status checking, written in Go.

Ease Probe supports the following probing methods:

- **HTTP**. Checking the HTTP status code, Support mTLS, HTTP Basic Auth, and can set the Request Header/Body.
- **TCP**. Just check the server can be connected successfully.
- **Shell**. Run a Shell command and check the result.

Ease Probe supports the following notifications:

- Email
- Slack
- Discord
- Log File

**Note**: The notification is the edge-trigger, only notified when the status changed.

And Ease Probe would send the Daily SLA Report at 00:00 UTC 

# Getting Start

## Build

Compiler `Go 1.17+`

Use `make` to make the binary file. the target is under the `build/bin` directory

```
make
```

## Run

Running the following command for local test

```
go run ./cmd/easeprobe/main.go -f config.yaml 
```

## Configuration

The following configuration is an example.

```YAML

# HTTP Probe Configuration

http:
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
    # configuration
    timeout: 10s # default is 30 seconds
    interval: 60s # default is 60 seconds

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
  - name: EaseService-Notification (AWS China)



# TCP Probe Configuration
tcp:
  - name: SSH Service (Invalid)
    host: error.megaease.cn:22
    timeout: 10s # default is 30 seconds
    interval: 2m # default is 60 seconds

  - name: Kafkak
    host: kafka.server:9093



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
    contain : "PONG"




# Notification Configuration
notify:
  # Notify to a local log file
  log:
    - file: "/tmp/easeprobe.log"
      dry: true  
  # Notify to Slack Channel
  slack:
    - webhook: "https://hooks.slack.com/services/........../....../....../"
      # dry: true   # dry notification, print the Slack JSON in log(STDOUT)
  # Notify to Discord Text Channel
  discord:
    - webhook: "https://discord.com/api/webhooks/...../....../"
      # dry: true # dry notification, print the Discord JSON in log(STDOUT)
      retry: # something the network is not good need to retry.
        times: 3
        interval: 10s
  email:
    - server: smtp.email.example.com:465
      username: user@example.com
      password: ********
      to: "user1@example.com;user2@example.com"
      # dry: true # dry notification, print the Email HTML in log(STDOUT)

# Global settings for all probers and notifiers.
settings:
  notify:
    # dry: true # Global settings for dry run 
    retry: # Global settings for retry 
      times: 5
      interval: 10s
  probe:
    interval: 1m # probe every minute
  # easeprobe program running log file.
  logfile: "test.log" 
  
  # Log Level Configuration
  # can be: panic, fatal, error, warn, info, debug.
  loglevel: "debug"

  # debug mode 
  # - true: the SLA report would be sent in every minute
  # - false: the SLA report would be sent in every day at 00:00 UTC time
  debug: false
 
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