package config

import (
    "io/ioutil"
    "gopkg.in/yaml.v2"
    "github.com/r3boot/rlib/logger"
)


type ResolverStruct struct {
    Search      string
    Nameservers []string
}


type NetworkStruct struct {
    Name        string
    Address     string
    Gateway     string
    Address6    string
    Gateway6    string
}


type OpenVPNStruct struct {
    Network     string
    Name        string
    Address     string
    Gateway     string
    Address6    string
    Gateway6    string
    Routes      []string
}


type ConfigStruct struct {
    Resolver    ResolverStruct
    Networks    []NetworkStruct
    Tunnels     []OpenVPNStruct
}


var Log logger.Log
var Config ConfigStruct

func Setup(l logger.Log) {
    Log = l
}


func LoadConfig(file_name string) {
    var err error

    content, err := ioutil.ReadFile(file_name)
    if err != nil {
        Log.Fatal("Failed to read file: " + err.Error())
    }

    err = yaml.Unmarshal([]byte(content), &Config)
    if err != nil {
        Log.Fatal("Failed to unmarshall data: " + err.Error())
    }

    Log.Debug("Configuration loaded succesfully")

}
