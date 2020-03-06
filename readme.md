![Go](https://github.com/yiyuezhuo/xisocks2/workflows/Go/badge.svg)
[![Build Status](https://travis-ci.org/yiyuezhuo/xisocks2.svg?branch=master)](https://travis-ci.org/yiyuezhuo/xisocks2)
[![Coverage Status](https://coveralls.io/repos/github/yiyuezhuo/xisocks2/badge.svg?branch=master)](https://coveralls.io/github/yiyuezhuo/xisocks2?branch=master)


# Yet another toy proxy

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

```
$ go get github.com/gorilla/websocket
$ go get github.com/yiyuezhuo/xisocks2
```

Enter project root

```
$ make
```

## Usage in PC

* Download respective version into your local computer(client) and VPS(server) from [release page](https://github.com/yiyuezhuo/xisocks2/releases).
* In client, replace config-client.json item value corresponding to ProxyURL with your hostname such as "xisocks2.com", which have been "protected" by CDN such as cloudflare. Thus CCP internet cops can't find your real IP. 
* See some TLS [tutorials](https://guide.v2fly.org/en_US/advanced/tls.html#register-a-domain) to get `server.crt` and `server.key` and placing them into server root.
* In client, run `client.exe`(windows) or `./client`(linux)
* In server, run `server.exe`(windows) or `./server`(linux)

## Usage in Android

In addtiontion to PC usage, the android version is just compiled using `GOARCH=arm64` and `GOOS=linux`. 
No UI is provided, you can use 
[termux](https://play.google.com/store/apps/details?id=com.termux&hl=en_US) 
and 
[Postern](https://play.google.com/store/apps/details?id=com.tunnelworkshop.postern&hl=en_US)
to help you use `xisocks2` just like usage for other console-oriented applications such as `v2ray-core`. Note `Data Sniffer` in `Rule` in `Postern`should be enable.