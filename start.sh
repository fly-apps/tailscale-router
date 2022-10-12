#!/bin/sh
echo 'net.ipv4.ip_forward = 1' | tee -a /etc/sysctl.conf
echo 'net.ipv6.conf.all.forwarding = 1' | tee -a /etc/sysctl.conf
sysctl -p /etc/sysctl.conf

/app/tailscaled --state=/var/lib/tailscale/tailscaled.state --socket=/var/run/tailscale/tailscaled.sock &
/app/tailscale-router

if [ $? -eq 0 ]
then
  /app/linux-amd64/dnsproxy -u fdaa::3
else
  exit 1
fi


tail -f /dev/null
