ARG goversion
ARG alpineversion

FROM docker.io/library/golang:${goversion}-alpine as builder
WORKDIR /app
COPY . ./

RUN go mod download
RUN go build

FROM docker.io/library/alpine:${alpineversion}
RUN apk --no-cache add ca-certificates iptables ip6tables bash bind-tools jq

ARG tailscaleversion
ARG dnsproxyversion

WORKDIR /app
COPY . ./
ENV TSFILE=tailscale_${tailscaleversion}_amd64.tgz
ENV DNSPROXYFILE=dnsproxy-linux-amd64-v0.45.2.tar.gz
ENV DNSPROXYVERSION=${dnsproxyversion}
RUN wget https://pkgs.tailscale.com/stable/${TSFILE} && tar xzf ${TSFILE} --strip-components=1
RUN wget https://github.com/AdguardTeam/dnsproxy/releases/download/${DNSPROXYVERSION}/${DNSPROXYFILE} && tar xzf ${DNSPROXYFILE} --strip-components=1
COPY --from=builder /app/tailscale-router /app/tailscale-router
COPY . ./

RUN mkdir -p /var/run/tailscale /var/cache/tailscale /var/lib/tailscale

CMD ["/app/start.sh"]
