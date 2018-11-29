# socks-over-HTTPS

SOCKS5 proxy over HTTP tunnel, which simply coverts a certain HTTPS proxy into SOCKS5 proxy

A typical HTTP proxy which can proxy HTTPS requests a.k.a. `HTTPS proxy` is mostly based on the [HTTP tunnel](HTTPS://en.wikipedia.org/wiki/HTTP_tunnel) by using the [CONNECT](HTTPS://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/CONNECT) method of HTTP.

For example, a typical protocol of https proxy request to https://example.com/some/path is

```
CONNECT example.com:443 HTTP/1.1
Host: example.com:443
User-Agent: some-user-agent
Proxy-Authorization: Basic dXNlcjpwYXNz
```

The proxy will open a **TCP tunnel** to `example.com:443` for the client and return

```
HTTP/1.1 200 Connection established
```

Then any traffic sent to the proxy will be redirected to the **TCP tunnel** opened by proxy.

We can turn the **TCP tunnel** above for SOCKS5 protocol's tunnel:

```
+--------+       +--------+        +-----------+      +--------+
|        +-------> socks5 +-------->           +------>        |
| client |       | over   |        |https proxy|      | server |
|        <-------+ https  <--------+           <------+        |
+--------+       +--------+        +-----------+      +--------+
```

According to [RFC1928](https://www.ietf.org/rfc/rfc1928.txt), the protocol of socks5 proxy request to https://example.com/some/path with https tunnel is

1) client sends the version identifier/method selection message to proxy

```
using socks v5, using 3 auth methods: no auth, GSSAPI and username/password
+-----+----------+----------+
| VER | NMETHODS | METHODS  |
+-----+----------+----------+
| 0x05|   0x03   | 0x000102 |
+-----+----------+----------+
```

2) server responds the version message

```
using socks v5, using the no auth method
+-----+--------+
| VER | METHOD |
+-----+--------+
| 0x05|  0x02  |
+-----+--------+
````

3) client sends the tunnel request to target server

```
CONNECT(CMD 1) to domain(ATYP 3) example.com with port 433
+-----+-----+-----+------+-------------+----------+
| VER | CMD | RSV | ATYP |  DST.ADDR   | DST.PORT |
+-----+-----+-----+------+-------------+----------+
| 0x05| 0x01| 0x00| 0x03 | example.com |   443    |
+-----+-----+-----+------+-------------+----------+
```

4) server gets the request and connects to the remote https proxy, then respond to client

- server connect to https proxy

```
CONNECT example.com:443 HTTP/1.1
Host: example.com:443
User-Agent: some-user-agent
Proxy-Authorization: Basic dXNlcjpwYXNz
```

- remote proxy responds

```
HTTP/1.1 200 Connection established
```

- respond success (REP 0) to client

```
+-----+-----+------+------+----------+----------+
| VER | REP |  RSV | ATYP | BND.ADDR | BND.PORT |
+-----+-----+------+------+----------+----------+
| 0x05| 0x00| 0x00 |  1   | 0.0.0.0  |   7648   |
+-----+-----+------+------+----------+----------+
```

5) data transfer

Any traffic from client in written into http tunnel opened by proxy, any traffic from tunnel is written to client.