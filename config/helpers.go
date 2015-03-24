package config

import (
    "net"
    "errors"
)


func GetNetwork (o_net string) (network NetworkStruct, err error) {
    var on, n *net.IPNet

    _, on, err = net.ParseCIDR(o_net)
    if err != nil {
        Log.Fatal("net.ParseCIDR failed on " + o_net)
        return
    }

    for _, network = range Config.Networks {
        if network.Address == "dhcp" {
            continue
        }

        _, n, err = net.ParseCIDR(network.Address)
        if err != nil {
            Log.Fatal("net.ParseCIDR failed on " + network.Address)
            return
        }
        if n.Contains(on.IP) {
            return
        }
    }

    for _, network = range Config.Networks {
        if network.Address == "dhcp" {
            return
        }
    }

    network = NetworkStruct{}
    err = errors.New("No valid network configuration found")
    return
}


func GetVPNTunnel(netname string) (tunnel OpenVPNStruct, err error) {
    for _, tunnel = range Config.Tunnels {
        if tunnel.Network == netname {
            return
        }
    }

    tunnel = OpenVPNStruct{}
    return
}

func GetResolver() (resolver ResolverStruct, err error) {
    resolver = Config.Resolver
    return
}
