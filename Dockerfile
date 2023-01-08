FROM golang:1.18 as build
ADD . /src
WORKDIR /src
RUN go build -ldflags='-w -s' -o /opt/init ./cmd/init
RUN go build -ldflags='-w -s' -o /opt/init-exec ./cmd/init-exec

FROM debian:bookworm-slim as base
ARG DEBIAN_FRONTEND=noninteractive
RUN apt update && apt install -y ntp openssh-server rsyslog stress-ng

FROM base
ADD config/init.yaml /etc/init.yaml
COPY --from=build /opt/init /sbin/init
COPY --from=build /opt/init-exec /sbin/init-exec
ENTRYPOINT ["/sbin/init"]
