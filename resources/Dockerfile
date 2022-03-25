FROM golang:1.17-alpine as builder
WORKDIR /go/src/github.com/megaease/easeprobe/
COPY ./ /go/src/github.com/megaease/easeprobe/
RUN apk --no-cache add make git && make clean && make

FROM alpine:latest
RUN apk update && apk add tini tzdata busybox-extras curl redis
WORKDIR /opt/
COPY --from=builder /go/src/github.com/megaease/easeprobe/build/bin/* ./
COPY --from=builder /go/src/github.com/megaease/easeprobe/resources/scripts/entrypoint.sh /
ENV PATH /opt/:$PATH
ENTRYPOINT ["/sbin/tini", "--", "/entrypoint.sh"]
