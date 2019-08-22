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

const BUFFER_SIZE = 8192
const CONNECTION_POOL_SIZE = 16

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

	l, err := net.Listen("tcp", localAddr) // l mean Listener
	if err != nil {
		log.Panic(err)
	}
	defer l.Close()

	fmt.Println("Linsen to", localAddr)

	// building connection pool(channel with )
	connection_pool_channel = make(chan *websocket.Conn, CONNECTION_POOL_SIZE)
	for i := 0; i < CONNECTION_POOL_SIZE; i++ {
		go buildConnection()
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}

		go handleRequest(conn)
	}

}

func buildConnection() {
	for {
		// serial running to debug, lately we will introduce some "workers" to do it.
		u := url.URL{Scheme: "wss", Host: proxyURL, Path: "/"}
		remote_c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			//fmt.Println("Sleep 1 second to dial again", err)
			//fmt.Println("Sleep 1 second to dial again")
			time.Sleep(1)
			continue
		}
		connection_pool_channel <- remote_c
		//fmt.Println("build a TLS connection to remote")
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

	remote_c := <-connection_pool_channel
	//defer remote_c.Close()

	// build xi handshake packet
	//lenHost := len(host)
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

	//fmt.Println("message_buff:", len(message_buff), cap(message_buff), message_buff, string(message_buff))

	err = remote_c.WriteMessage(websocket.BinaryMessage, message_buff)
	if err != nil {
		fmt.Println("xi handshake failed", err)
		return
	}

	//proxy(conn, remote_c)
	common.Proxy(conn, remote_c)
}
