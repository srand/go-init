FROM golang:1.18 as build
ADD . /src
WORKDIR /src
RUN go build -ldflags='-w -s' -o /opt/init ./cmd/init 

FROM debian:bookworm-slim as base
RUN apt update && apt install -y ntp openssh-server rsyslog

FROM base
ADD config/example.yaml /etc/init.yaml
COPY --from=build /opt/init /sbin/init
ENTRYPOINT ["/sbin/init"]