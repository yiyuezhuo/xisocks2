package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yiyuezhuo/xisocks2/common"
	//"github.com/yiyuezhuo/xisocks2/common"
)

var localAddr, token, proxyURL string
var proxyPort int
var lenToken byte
var useConnectionPool, ResolveHTTP bool

var proxyAddr string

const BUFFER_SIZE = 8192
const CONNECTION_POOL_SIZE = 64

var connection_pool_channel chan *websocket.Conn

var numSocks5Connected, numHTTPConnected int

var configPath = flag.String("config", "config-client.json", "config path")

func main() {
	fmt.Println("Start client")
	flag.Parse()
	config := loadConfig(*configPath)
	localAddr = config.LocalIp + ":" + strconv.Itoa(config.LocalPort)
	token = config.Token
	lenToken = byte(len(config.Token))
	proxyURL = config.ProxyURL
	proxyPort = config.ProxyPort
	useConnectionPool = config.UseConnectionPool
	ResolveHTTP = config.ResolveHTTP

	proxyAddr = proxyURL + ":" + strconv.Itoa(proxyPort)

	l, err := net.Listen("tcp", localAddr) // l mean Listener
	if err != nil {
		log.Panic(err)
	}
	defer l.Close()

	fmt.Println("Linsen to", localAddr)

	// building connection pool(channel with )
	if useConnectionPool {
		connection_pool_channel = make(chan *websocket.Conn, CONNECTION_POOL_SIZE)
		for i := 0; i < CONNECTION_POOL_SIZE; i++ {
			go buildConnection(i)
		}
	}

	numSocks5Connected = 0
	numHTTPConnected = 0

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("numSocks5Connected:", numSocks5Connected,
				"numHTTPConnected", numHTTPConnected)
			log.Panic(err)
		}

		go handleRequest(conn, config)
	}

}

func buildConnection(worker_id int) {
	for {
		// serial running to debug, lately we will introduce some "workers" to do it.
		u := url.URL{Scheme: "wss", Host: proxyAddr, Path: "/"}
		remote_c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			//fmt.Println("Sleep 1 second to dial again", err)
			fmt.Println("worker:", worker_id, "Sleep 1 second to dial again")
			time.Sleep(1)
			continue
		}
		connection_pool_channel <- remote_c
		fmt.Println("worker:", worker_id, "build a TLS connection to remote")
	}

}

func handleRequest(conn net.Conn, config Config) {
	//defer conn.Close()
	// Close should be called in Proxy, but we should call it before any other return

	remote_host, payload, err := local_handshake(conn)
	host := []byte(remote_host)

	var remote_c *websocket.Conn
	//remote_c := <-connection_pool_channel
	if useConnectionPool {
		remote_c = <-connection_pool_channel
	} else {
		var u url.URL
		if config.TLS {
			u = url.URL{Scheme: "wss", Host: proxyAddr, Path: "/"}
		} else {
			u = url.URL{Scheme: "ws", Host: proxyAddr, Path: "/"}
		}
		remote_c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			fmt.Println("dial to proxy server fail", err)
			conn.Close()
			fmt.Println("remote_c:", remote_c)
			if remote_c != nil {
				remote_c.Close() // Close
			}
			return
		}
	}

	xi_header := common.XiHeader{
		LenToken: lenToken,
		Token:    []byte(token),
		LenHost:  byte(len(host)),
		Host:     host,
		Payload:  []byte(payload),
	}
	//common.DisplayXiHeader(xi_header)
	message_buff := common.BuildXiHandshake(xi_header)

	//fmt.Println("Connect to", string(host))

	//fmt.Println("message_buff:", len(message_buff), cap(message_buff), message_buff, string(message_buff))
	//common.DisplayXiHeader(xi_header)
	fmt.Println("message_buff len", len(message_buff))
	err = remote_c.WriteMessage(websocket.BinaryMessage, message_buff)
	if err != nil {
		fmt.Println("xi handshake failed", err)
		conn.Close()
		remote_c.Close()
		return
	}

	//proxy(conn, remote_c)
	common.Proxy(conn, remote_c)
	conn.Close()
	remote_c.Close()
}

func local_handshake(conn net.Conn) (remote_host string, payload []byte, err error) {
	var req *http.Request

	if !ResolveHTTP {
		remote_host, err = socks5_handshake(conn)
		numSocks5Connected++
		fmt.Println("numSocks5Connected: ", numSocks5Connected,
			conn.RemoteAddr(), "->", conn.LocalAddr())
	} else {
		buf := make([]byte, 1)
		conn.Read(buf)

		is := &InsertedStream{buf, conn}

		if buf[0] == 5 { // socks5
			remote_host, err = socks5_handshake(is)
			numSocks5Connected++
			fmt.Println("numSocks5Connected ->", numSocks5Connected,
				conn.RemoteAddr(), "->", conn.LocalAddr())
		} else {
			remote_host, req, err = http_handshake(is)
			numHTTPConnected++
			fmt.Println("numHTTPConnected ->", numHTTPConnected,
				conn.RemoteAddr(), "->", conn.LocalAddr())
		}
	}

	if err != nil {
		fmt.Println(err)
		return "", make([]byte, 0), err
	}

	if (req == nil) || (req.Method == "CONNECT") { // socks5 or https connect
		buf := make([]byte, BUFFER_SIZE)

		readLen, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			return "", make([]byte, 0), err
		}

		payload = buf[:readLen]
	} else {
		b := &bytes.Buffer{}
		err := req.Write(b)
		if err != nil {
			fmt.Println(err)
			return "", make([]byte, 0), err
		}

		payload = b.Bytes()
	}

	/*
		if req != nil {
			req.Body.Close()
		}
	*/

	return remote_host, payload, nil

}

type InsertedStream struct {
	inserted []byte
	conn     io.ReadWriter
}

func (is *InsertedStream) Read(buf []byte) (int, error) {
	lenRead := 0
	if len(is.inserted) > 0 {
		lenRead = copy(buf, is.inserted)

		is.inserted = is.inserted[lenRead:]

		buf = buf[lenRead:]
		if len(buf) == 0 {
			return lenRead, nil
		}
	}

	lenRead2, err := is.conn.Read(buf)

	return lenRead + lenRead2, err
}
func (is *InsertedStream) Write(buf []byte) (int, error) {
	return is.conn.Write(buf)
}
