# tailscale-router

[![CI](https://github.com/ananthb/tailscale-router/actions/workflows/ci.yml/badge.svg)](https://github.com/ananthb/tailscale-router/actions/workflows/ci.yml)

## How to use

1. Create an app  `flyctl apps create app-name`
2. Get a secret from the tailscale admin console: tailscale admin console > settings > keys > `generate auth key` (you probably want to choose the reusable and ephemeral options).
3. Set the token you get as a secret `flyctl secrets set TAILSCALE_AUTHKEY=thekeyyougot -a app-name`.
4. Deploy a machine `flyctl m run ghcr.io/ananthb/tailscale-router:latest -a app-name --cpus 1 --memory 256`
5. Follow steps `3` and `5` of <https://tailscale.com/kb/1019/subnets/> to enable subnets for the machine that got automatically configured.
6. Enjoy

## Test it Out

You can test if it's working by finding the IP address of your new Fly.io app and using `dig`:

```bash
# Get the IP address of your app:
flyctl m list -a app-name

# Use dig to test DNS queries the DNS proxy setup in this repository
dig @<your-app-ip-address-here> aaaa app-name.internal
```

## DNS Setup

You can enable split DNS in your Tailscale settings to automatically resolve `*.internal` addresses through the DNS proxy setup in your new Fly.io app.

Tailscale documentation for that is [found here](https://tailscale.com/kb/1054/dns/).

1. Add a nameserver
2. Use the IP address of your new Fly.io app
3. Restrict to search domains, and use search domain `internal`

Then addresses should resolve! Maybe use `curl` to make an HTTP request to one of your apps. Be sure to use the `internal_port` of your application:

```bash
curl http://some-fly-app.internal:8080
```
