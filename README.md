# Ease Probe

Simple, Standalone, health check tools, written in Go.

## Build

Compiler `Go 1.17+`

Using `make` to make the binary file. the target is under `build/bin` directory

```
make
```

## Test

Running the following command for locally test

```
go run ./cmd/easeprobe/main.go -f conf/test.yaml 
```

## Configuration

The following configuration is an example.

```yaml
http:
  - name: MegaEase Website (China)
    url: https://megaease.cn
    method: GET
    headers:
      X-head-one: xxxxxx
      X-head-two: yyyyyy
      X-head-THREE: zzzzzzX-
    content_encoding: text/json
    body: '{ "FirstName": "Mega", "LastName" : "Ease", "UserName" : "megaease", "Email" : "megaease@example.com"}'

    username: username
    password: password

    #ca: /path/to/file.ca
    #cert: /path/to/file.crt
    #key: /path/to/file.key

  - name: MegaEase Website (Global)
    url: https://megaease.com
    method: GET

tcp:
  - name: MegaEase SSH (China)
    host: ssh.megaease.cn:22
    timeout: 10s
  - name: MegaEase HTTP Service (China)
    host: megaease.cn:80

notify:
  log:
    - file: "/tmp/easeprobe.log"
  slack:
    - webhook: "https://hooks.slack.com/services/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"
  email:
    - server: smtp.exmail.qq.com:465
      username: noreply@megaease.com
      password: 644D4u43n
      to: "service@megaease.com;chenhao@megaease.com"

```