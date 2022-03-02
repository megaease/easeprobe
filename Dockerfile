FROM golang:1.17-alpine as builder
WORKDIR /go/src/github.com/megaease/easeprobe/
COPY ./ /go/src/github.com/megaease/easeprobe/
RUN apk --no-cache add make git && make clean && make

FROM alpine:latest
WORKDIR /opt/
COPY --from=builder /go/src/github.com/megaease/easeprobe/build/bin/* ./
ENTRYPOINT ["/opt/easeprobe", "-f", "/opt/config.yaml"]