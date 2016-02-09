package lib

import (
	"github.com/r3boot/gonet/config"
	"github.com/r3boot/rlib/network"
	"github.com/r3boot/rlib/vpn"
	"net"
	"time"
)

type Interface struct {
	network.Interface
	Config  config.ConfigStruct
	OpenVPN vpn.OpenVPN
	Latency float64
}

func (intf *Interface) Start() {
	var ip, gw4, gw6, outside_gw net.IP
	var def_ipv4_dest, def_ipv6_dest, prefix *net.IPNet
	var resolvers []net.IP
	var tunnel config.OpenVPNStruct
	var err error

	i := *intf
	r := network.RIB{}

	_, o_network, err := network.GetOffer()
	if err != nil {
		Log.Warning(err)
		Log.Fatal("No DHCPOFFER received")
	}

	conn, err := config.GetNetwork(o_network.String())
	if err != nil {
		Log.Warning("No network settings found for " + o_network.String() + ", using defaults")
	}

	// IPv4 configuration
	if conn.Address == "dhcp" {
		Log.Debug("Configuring IPv4 using dhcp")

		if !i.Dhcpcd.IsRunning() {
			if err = i.Dhcpcd.Start(); err != nil {
				Log.Fatal("Failed to start dhcpcd: " + err.Error())
			}
		}

	} else {
		Log.Debug("Configuring IPv4 using static ip addresses")
		if i.Dhcpcd.IsRunning() {
			if err = i.Dhcpcd.Stop(); err != nil {
				Log.Fatal("Failed to stop dhcpcd: " + err.Error())
			}
		}

		if err = i.Ip.AddAddress(conn.Address, network.AF_INET); err != nil {
			Log.Fatal("Failed to add ip address: " + err.Error())
		}

		if _, def_ipv4_dest, err = net.ParseCIDR("0.0.0.0/0"); err != nil {
			Log.Fatal("net.ParseCIDR failed on : " + err.Error())
		}

		gw4 = net.ParseIP(conn.Gateway)
		if err = r.AddRoute(*def_ipv4_dest, gw4); err != nil {
			Log.Fatal("Failed to add default route: " + err.Error())
		}
	}

	// IPv6 configuration
	if conn.Address6 == "dhcp" {
		if err = i.RA.EnableRA(); err != nil {
			Log.Fatal("Failed to enable RA on " + i.Name + ": " + err.Error())
		}
	} else {
		if err = i.RA.DisableRA(); err != nil {
			Log.Fatal("Failed to disable RA on " + i.Name + ": " + err.Error())
		}

		if err = i.Ip.AddAddress(conn.Address6, network.AF_INET6); err != nil {
			Log.Fatal("Failed to add ip address: " + err.Error())
		}

		// Wait a bit for ND to operate
		time.Sleep(100 * time.Millisecond)

		if _, def_ipv6_dest, err = net.ParseCIDR("::/0"); err != nil {
			Log.Fatal("net.ParseCIDR failed on : " + err.Error())
		}

		gw6 = net.ParseIP(conn.Gateway6)
		if err = r.AddRoute(*def_ipv6_dest, gw6); err != nil {
			Log.Fatal("Failed to add default route: " + err.Error())
		}

	}

	if !i.HasUplink(outside_gw) {
		Log.Fatal("Failed to establish a network connection")
	}

	// Setup /etc/resolv.conf
	res, err := config.GetResolver()
	if err != nil {
		Log.Fatal("Failed to get resolver config")
	}

	for _, ip_s := range res.Nameservers {
		ip = net.ParseIP(ip_s)
		resolvers = append(resolvers, ip)
	}
	i.Resolvconf.Search = res.Search
	i.Resolvconf.Nameservers = resolvers
	i.Resolvconf.AddConfig()

	// Setup VPN tunnels (if needed)
	if outside_gw, err = r.GetDefaultGateway(network.AF_INET); err != nil {
		Log.Fatal("Failed to determine default gateway: " + err.Error())
	}

	if tunnel, err = config.GetVPNTunnel(conn.Name); err != nil {
		Log.Fatal("Failed to load tunnel config: " + err.Error())
	}

	if tunnel.Name != "" {
		Log.Debug("Starting OpenVPN tunnel for " + conn.Name)
		if i.OpenVPN, err = vpn.OpenVPNFactory(tunnel.Name); err != nil {
			Log.Fatal("Failed to generate OpenVPN struct: " + err.Error())
		}

		_, vpn_gw, err := net.ParseCIDR(i.OpenVPN.Remote.String() + "/32")
		if err != nil {
			Log.Fatal("Failed to determine remote for tunnel: " + err.Error())
		}
		r.AddRoute(*vpn_gw, outside_gw)

		if err = i.OpenVPN.Start(); err != nil {
			Log.Fatal("Failed to start OpenVPN tunnel: " + err.Error())
		}

		err = i.OpenVPN.Interface.Ip.AddAddress(tunnel.Address, network.AF_INET)
		if err != nil {
			Log.Fatal("Failed to add ip address: " + err.Error())
		}

		err = i.OpenVPN.Interface.Ip.AddAddress(tunnel.Address6, network.AF_INET6)
		if err != nil {
			Log.Fatal("Failed to add ip address: " + err.Error())
		}

		gw4 = net.ParseIP(tunnel.Gateway)
		gw6 = net.ParseIP(tunnel.Gateway6)

		for _, route := range tunnel.Routes {
			if _, prefix, err = net.ParseCIDR(route); err != nil {
				Log.Warning("Failed to add route: " + err.Error())
				continue
			}
			if len(prefix.IP) == net.IPv4len {
				r.AddRoute(*prefix, gw4)
			} else if len(prefix.IP) == net.IPv6len {
				r.AddRoute(*prefix, gw6)
			} else {
				Log.Warning("Unknown IP protocol: " + prefix.String())
			}
		}

	} else {
		Log.Debug("No VPN tunnels found for " + conn.Name)
	}

	*intf = i
}

func (intf *Interface) Stop() {
	var def_ipv4_dest, def_ipv6_dest *net.IPNet
	var err error

	i := *intf
	r := network.RIB{}

	i.OpenVPN.Stop()

	i.Resolvconf.RemoveConfig()

	if i.Dhcpcd.IsRunning() {
		i.Dhcpcd.Stop()
	}

	i.RA.DisableRA()

	if _, def_ipv4_dest, err = net.ParseCIDR("0.0.0.0/0"); err != nil {
		Log.Fatal("net.ParseCIDR failed on : " + err.Error())
	}
	r.RemoveRoute(*def_ipv4_dest)

	if _, def_ipv6_dest, err = net.ParseCIDR("::/0"); err != nil {
		Log.Fatal("net.ParseCIDR failed on : " + err.Error())
	}
	r.RemoveRoute(*def_ipv6_dest)

	i.Ip.FlushAllAddresses()

	*intf = i
}

func (intf *Interface) HasUplink(gateway net.IP) (reachable bool) {
	var err error

	i := *intf

	reachable = true
	return

	reachable, i.Latency, err = network.Arping(gateway, i.Name, 1)
	if err != nil {
		Log.Fatal("Failed to arping " + gateway.String() + ": " + err.Error())
	}

	*intf = i
	return
}
