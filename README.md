### Go!net
Go!net is a no-frills network manager designed for roadwarriors. It uses dhcpcd
to detect the network you're connected to, and will be able to set per-network
configuration. Currently, it will perform the following tasks:

* Configure IPv4/IPv6 addresses (both via DHCP/RA and static)
* Setup a VPN using split-default routing
* Configuration of /etc/resolv.conf

### Interface teaming/bonding/aggregation
Go!net assumes that you're using bonding to perform failover between all
interfaces available on your system. This means that you are responsible for
configuring wireless, and choosing the failover policy.

To configure a failover interface, follow these steps:
```
# modprobe bonding mode=active-backup primary=<your primary nic> miimon=100
# /usr/bin/ip link set <your primary nic> up
# /usr/bin/ip link set <your secondary nic> up
# /usr/bin/ifenslave bond0 <your primary nic> <your secondary nic>
```
