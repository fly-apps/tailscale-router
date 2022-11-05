FROM golang:1.19-alpine as builder
WORKDIR /app
COPY . ./

RUN go mod download
RUN go build

FROM alpine:latest
RUN apk update && apk add ca-certificates iptables ip6tables bash bind-tools jq && rm -rf /var/cache/apk/*

ARG tailscale_version=1.30.2
ARG dnsproxyversion=v0.45.2

WORKDIR /app
COPY . ./
ENV TSFILE=tailscale_${tailscale_version}_amd64.tgz
ENV DNSPROXYFILE=dnsproxy-linux-amd64-v0.45.2.tar.gz
ENV DNSPROXYVERSION=${dnsproxyversion}
RUN wget https://pkgs.tailscale.com/stable/${TSFILE} && tar xzf ${TSFILE} --strip-components=1
RUN wget https://github.com/AdguardTeam/dnsproxy/releases/download/${DNSPROXYVERSION}/${DNSPROXYFILE} && tar xzf ${DNSPROXYFILE} --strip-components=1
COPY --from=builder /app/tailscale-router /app/tailscale-router
COPY . ./

RUN mkdir -p /var/run/tailscale /var/cache/tailscale /var/lib/tailscale

CMD ["/app/start.sh"]
