package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/haxii/daemon"
	"github.com/haxii/fastproxy/bufiopool"
	"github.com/haxii/log"
)

var (
	configPath = flag.String("c", "config.json", "config file")
	_          = flag.String("s", daemon.UsageDefaultName, daemon.UsageMessage)
)

func main() {
	d := daemon.Make("-s", "socks-over-https", "socks proxy on http tunnel")
	d.Run(serve)
}

func serve() {
	flag.Parse()

	// init config
	config := makeConfig(*configPath)

	// setup logger
	isDebug := false
	if config.Log.Level == "debug" {
		isDebug = true
	}
	logger, err := log.MakeZeroLogger(isDebug, log.LoggingConfig{
		FileDir: config.Log.Dir,
	}, "socks-over-https")
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to make logging file in %s with error %s\n", config.Log.Dir, err)
		os.Exit(1)
	}
	defer logger.CloseLogger()

	// init bufio pool
	bufioPool := bufiopool.New(config.Settings.ReadBufferSize, config.Settings.WriteBufferSize)

	// init all socks5 servers
	if len(config.Proxies) == 0 {
		logger.Fatal("MGR", nil, "no proxy defined")
	}
	serverMap := make(map[string]*server)
	for _, proxyPair := range config.Proxies {
		if len(proxyPair.Socks.Address) == 0 {
			proxyPair.Socks.Address = "127.0.0.1"
		}
		if proxyPair.Socks.Port <= 0 || proxyPair.Socks.Port >= 0xffff ||
			proxyPair.HTTP.Port <= 0 || proxyPair.HTTP.Port >= 0xffff ||
			len(proxyPair.HTTP.Address) == 0 {
			logger.Fatal("MGR", err, "invalid proxy settings of %+v", proxyPair)
		}
		socksAddr := fmt.Sprintf("%s:%d", proxyPair.Socks.Address, proxyPair.Socks.Port)
		if _, exists := serverMap[socksAddr]; exists {
			logger.Fatal("MGR", nil, "duplicated socks proxy settings in %+v", proxyPair)
		}
		serverMap[socksAddr], err = newServer(bufioPool, logger, proxyPair.HTTP.Address,
			(uint16)(proxyPair.HTTP.Port), proxyPair.HTTP.User, proxyPair.HTTP.Pass, socksAddr,
			proxyPair.Socks.User, proxyPair.Socks.Pass)
		if err != nil {
			logger.Fatal("MGR", err, "fail to make socks server of %+v", proxyPair)
		}
	}

	// start servers
	var wg sync.WaitGroup
	wg.Add(len(serverMap))
	for _, s := range serverMap {
		logger.Debug("MGR", "start socks5 proxy on %s via %s",
			s.getSocks5Description(), s.getUpstreamDescription())
		go func(_s *server) {
			_s.listenAndServe()
			wg.Done()
		}(s)
	}
	wg.Wait()
}
