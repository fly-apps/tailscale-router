ARG goversion=latest
ARG alpineversion=latest

FROM docker.io/library/golang:${goversion} as builder
WORKDIR /app
COPY . ./
RUN env CGO_ENABLED=0 go build ./...

FROM docker.io/library/alpine:${alpineversion}
RUN apk --no-cache add ca-certificates iptables ip6tables bash bind-tools jq
ARG tailscaleversion
ARG dnsproxyversion
WORKDIR /app
ENV TSFILE=tailscale_${tailscaleversion}_amd64.tgz
ENV DNSPROXYVERSION=${dnsproxyversion}
ENV DNSPROXYFILE=dnsproxy-linux-amd64-${dnsproxyversion}.tar.gz
RUN wget https://pkgs.tailscale.com/stable/${TSFILE} \
  && tar xzf ${TSFILE} --strip-components=1 \
  && wget https://github.com/AdguardTeam/dnsproxy/releases/download/${DNSPROXYVERSION}/${DNSPROXYFILE} \
  && tar xzf ${DNSPROXYFILE} --strip-components=1
COPY --from=builder /app/tsrouter /app/tsrouter
COPY init.sh /
RUN mkdir -p /var/run/tailscale /var/cache/tailscale /var/lib/tailscale
ENTRYPOINT [ "/bin/sh", "/init.sh" ]
