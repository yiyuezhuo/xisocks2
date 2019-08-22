# Yet a another toy proxy

Yet another toy proxy which provide authentication and removing verbose handshake procedure as much as possible, compared to [previous version](https://github.com/yiyuezhuo/xisocksGo).

## Protocol

This protocol is inspired by `v2ray` and `trojan`.

All socks5/https handshake will be treated successed to accelerate
transport. When proxed app such as browser, client start a TLS connection to
connect remote server outside GFW. When TLS connection have been established, 
a header with payload will be sent to server to specify TCP CONNECT destination 
and payload to reduce packet required to transport.

```
+----------------+--------------+--------------+-------------+---------------+
|   len(TOKEN)   |    TOKEN     |  len(host)   |     host    |   Payload     |
+-------------------------------+--------------+-----------------------------+
|     1 byte     | 1-255 bytes  |   1 byte     | 1-255 bytes |      *        |
+----------------+--------------+--------------+-------------+---------------+
```

## Fake website

When server take a wrong header format, it will forward the suspicious request to a local http server.
So that you and GFW internet agent may find that a website is hosted in that url if just access it from browser.

My personal interest is to show a sad panda to GFW internet agent, or you may launch a jupyter notebook server,
which provide powerfull screen-like terminor.

## Build

Install `github.com/gorilla/websocket`

Clone the project and `cd` into `$GOPATH/github.com/yiyuezhuo/`, then

```
make
```