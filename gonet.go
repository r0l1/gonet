package main

/*

== start ==
1) Determine existence of bond0
1a) No -> exit with error
1b) Yes -> continue
2) Determine link status of bond0
2a) No -> Wait until timeout for link
2b) Yes -> continue
3) Perform dhcp probe
4) Is network a configured network?
4a) Yes -> Load settings for network
4b) No -> Load default settings
5) Configure ip addresses
6) Configure ip routing
7) Configure openvpn

== stop ==
1) Stop openvpn
2) Flush routes
3) Flush ip addresses
4) Bring all interfaces down

*/

import (
    "flag"
    "net"
    "github.com/r3boot/rlib/logger"
    "github.com/r3boot/rlib/network"
    "github.com/r3boot/rlib/vpn"
    "github.com/r3boot/gonet/config"
    "github.com/r3boot/gonet/lib"
)

// Application specific constants
const APP_NAME string = "gonet"
const APP_VERS string = "0.1"

// Default options constants
const D_CFG_FILE    string = "/etc/gonet.yml"
const D_INTERFACE   string = "bond0"
const D_DEBUG       bool = false
const D_CONNECT     bool = false
const D_DISCONNECT  bool = false

var cfg_file    = flag.String("f", D_CFG_FILE, "Path to the configuration file")
var intf_s      = flag.String("i", D_INTERFACE, "Interface to work on")
var debug       = flag.Bool("D", D_DEBUG, "Enable debugging output")
var connect     = flag.Bool("c", D_CONNECT, "Connect to uplink")
var disconnect  = flag.Bool("d", D_DISCONNECT, "Disconnect from uplink")

var Log logger.Log

var Uplink lib.Interface

func init () {
    var intf network.Interface
    var has_interface bool
    var err error

    flag.Parse()

    Log.UseDebug = *debug
    Log.UseVerbose = true
    Log.UseTimestamp = true
    Log.Debug("Logging initialized")

    config.Setup(Log)
    lib.Setup(Log)

    Log.Debug("Initializing bonding interface")
    interfaces, err := net.Interfaces()
    if err != nil {
        Log.Fatal("Failed to load interface list: " + err.Error())
    }

    has_interface = false
    for _, raw_intf := range interfaces {
        if raw_intf.Name == *intf_s {
            intf, err = network.InterfaceFactory(raw_intf)
            if err != nil {
                Log.Fatal("Failed to initialize interface: " + err.Error())
            }
            has_interface = true
            break
        }
    }

    if has_interface {
        Uplink = lib.Interface{intf, config.Config, vpn.OpenVPN{}, 0.0}
    } else {
        Log.Fatal("No usable interfaces found")
    }

    Log.Debug("Running " + APP_NAME + " v" + APP_VERS)
}


func main () {
    config.LoadConfig(*cfg_file)

    if *disconnect {
        Log.Info("Disconnecting from " + Uplink.Name)
        Uplink.Stop()
    } else if *connect {
        Log.Info("Trying to establish network connectivity via " + Uplink.Name)
        Uplink.Start()
    }
}
