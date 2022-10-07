# tailscale-router

## How to use
1. Create an app  `flyctl apps create my-unqique-tailscale-router-app-name`
2. Get a secret from the tailscale admin console: tailscale admin console > settings > keys > generate auth key _(you probably want to choose the reusable and ephemeral options)_
3. Set the token you get as a secret `flyctl secrets set TAILSCALE_AUTHKEY=thekeyyougot -a my-unqique-tailscale-router-app-name`
4. Build this repo `docker build -t registry.fly.io/my-unqique-tailscale-router-app-name:latest .`
5. Deploy a machine `flyctl m run registry.fly.io/my-unqique-tailscale-router-app-name:latest -a my-unqique-tailscale-router-app-name --cpus 1 --memory 256`
6. Follow steps `3` and `5` of https://tailscale.com/kb/1019/subnets/ to enable subnets for the machine that got automatically configured
7. Enjoy
