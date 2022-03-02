FROM golang:1.17 as builder
WORKDIR /go/src/github.com/megaease/easeprobe/
RUN make

FROM alpine:latest
WORKDIR /opt/
COPY --from=builder /go/src/github.com/megaease/easeprobe/build/bin/* ./
ENTRYPOINT ["/opt/easeprobe", "-f", "/opt/config.yaml"]