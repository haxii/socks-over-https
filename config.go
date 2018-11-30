package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type proxy struct {
	Address string
	Port    int
	User    string
	Pass    string
}

type proxyPair struct {
	Socks proxy
	HTTP  proxy
}

type config struct {
	Log struct {
		Dir   string
		Level string
	}
	Settings struct {
		ReadBufferSize  int
		WriteBufferSize int
	}
	Proxies []proxyPair
}

func makeConfig(configPath string) *config {
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to read config file %s with error %s\n", configPath, err)
		os.Exit(1)
	}
	config := &config{}
	err = json.Unmarshal(configData, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to parse config file %s with error %s\n", configPath, err)
		os.Exit(1)
	}
	if len(config.Log.Dir) == 0 {
		fmt.Fprintf(os.Stderr, "no log path specifled\n")
		os.Exit(1)
	}
	return config
}
