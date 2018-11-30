package main

import (
	"bytes"
	"context"
	"errors"
	"net"

	stdlog "log"

	"github.com/armon/go-socks5"
	"github.com/haxii/fastproxy/bufiopool"
	"github.com/haxii/fastproxy/superproxy"
	"github.com/haxii/log"
)

type stdLogWriter struct {
	addr   string
	logger log.Logger
}

func (w *stdLogWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}
	if p[len(p)-1] == '\n' {
		p = p[:len(p)-1]
	}
	who := "socks5://" + w.addr
	if bytes.HasPrefix(p, []byte("[ERR] ")) {
		p = p[6:]
		w.logger.Error(who, nil, "%s", p)
	} else {
		w.logger.Info(who, "%s", p)
	}
	return len(p), nil
}

type server struct {
	socks5Server    *socks5.Server
	socks5ProxyAddr string

	upstream *superproxy.SuperProxy

	bufioPool *bufiopool.Pool
	logger    log.Logger
}

var (
	errNoAvailableUpstream = errors.New("no avaliable upstream")
	errInvalidBufioPool    = errors.New("invalid bufio pool")
)

func newServer(pool *bufiopool.Pool, logger log.Logger, httpProxyHost string, httpProxyPort uint16,
	httpProxyUser, httpProxyPass, socks5ProxyAddr, socks5ProxyUser, socks5ProxyPass string) (*server, error) {
	if pool == nil {
		return nil, errInvalidBufioPool
	}
	s := &server{bufioPool: pool, logger: logger, socks5ProxyAddr: socks5ProxyAddr}

	// set up super proxy
	upstream, err := superproxy.NewSuperProxy(httpProxyHost, httpProxyPort,
		superproxy.ProxyTypeHTTP, httpProxyUser, httpProxyPass, "")
	if err != nil {
		return nil, err
	}
	s.upstream = upstream

	// set up socks server
	conf := &socks5.Config{
		Dial:   s.httpTunnelDialer,
		Logger: stdlog.New(&stdLogWriter{socks5ProxyAddr, logger}, "", 0),
	}
	if len(socks5ProxyUser) > 0 && len(socks5ProxyPass) > 0 {
		creds := socks5.StaticCredentials{
			socks5ProxyUser: socks5ProxyPass,
		}
		authenticator := socks5.UserPassAuthenticator{Credentials: creds}
		conf.AuthMethods = []socks5.Authenticator{authenticator}
	}
	socks5Server, err := socks5.New(conf)
	if err != nil {
		return nil, err
	}
	s.socks5Server = socks5Server

	return s, nil
}

func (s *server) getSocks5Description() string {
	return s.socks5ProxyAddr
}

func (s *server) getUpstreamDescription() string {
	return s.upstream.HostWithPort()
}
func (s *server) listenAndServe() error {
	return s.socks5Server.ListenAndServe("tcp", s.socks5ProxyAddr)
}

func (s *server) httpTunnelDialer(ctx context.Context, network, addr string) (net.Conn, error) {
	s.logger.Debug("http://"+s.upstream.HostWithPort(),
		"tunnel to %s from socks5://%s", addr, s.socks5ProxyAddr)
	if s.upstream == nil {
		return nil, errNoAvailableUpstream
	}
	return s.upstream.MakeTunnel(s.bufioPool, addr)
}
