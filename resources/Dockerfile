FROM golang:1.20.4-alpine as builder
WORKDIR /go/src/github.com/megaease/easeprobe/
COPY ./ /go/src/github.com/megaease/easeprobe/
RUN --mount=type=cache,target=/var/cache/apk \
    apk --no-cache add make git gcc libc-dev

COPY go.mod go.mod
COPY go.sum go.sum
COPY . .

RUN --mount=type=cache,mode=0777,id=gomodcache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make

FROM alpine:latest
RUN apk update && apk add tini tzdata busybox-extras curl redis
WORKDIR /opt/
COPY --from=builder /go/src/github.com/megaease/easeprobe/build/bin/* ./
COPY --from=builder /go/src/github.com/megaease/easeprobe/resources/scripts/entrypoint.sh /
ENV PATH /opt/:$PATH
ENV PROBE_CONFIG /opt/config.yaml
ENTRYPOINT ["/sbin/tini", "--", "/entrypoint.sh"]
