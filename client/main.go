package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yiyuezhuo/xisocks2/common"
	//"github.com/yiyuezhuo/xisocks2/common"
)

var localAddr, token, proxyURL string
var lenToken byte
var useConnectionPool bool

const BUFFER_SIZE = 8192
const CONNECTION_POOL_SIZE = 64

var connection_pool_channel chan *websocket.Conn

var configPath = flag.String("config", "config-client.json", "config path")

func main() {
	fmt.Println("Start client")
	flag.Parse()
	config := loadConfig(*configPath)
	localAddr = config.LocalIp + ":" + strconv.Itoa(config.LocalPort)
	token = config.Token
	lenToken = byte(len(config.Token))
	proxyURL = config.ProxyURL
	useConnectionPool = config.UseConnectionPool

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

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}

		go handleRequest(conn)
	}

}

func buildConnection(worker_id int) {
	for {
		// serial running to debug, lately we will introduce some "workers" to do it.
		u := url.URL{Scheme: "wss", Host: proxyURL, Path: "/"}
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

func handleRequest(conn net.Conn) {
	//defer conn.Close()

	remote_host, err := socks5_handshake(conn)
	if err != nil {
		fmt.Println(err)
		return
	}

	host := []byte(remote_host)

	buf := make([]byte, BUFFER_SIZE)

	readLen, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		return
	}

	var remote_c *websocket.Conn
	//remote_c := <-connection_pool_channel
	if useConnectionPool {
		remote_c = <-connection_pool_channel
	} else {
		u := url.URL{Scheme: "wss", Host: proxyURL, Path: "/"}
		remote_c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			fmt.Println("dial to proxy server fail", err)
			return
		}
	}

	payload := buf[:readLen]

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
		return
	}

	//proxy(conn, remote_c)
	common.Proxy(conn, remote_c)
}
