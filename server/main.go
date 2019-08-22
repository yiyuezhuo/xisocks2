package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/yiyuezhuo/xisocks2/common"
	//"github.com/yiyuezhuo/xisocks2/common"
)

var listenAddr, forwardAddr, token string
var upgrader = websocket.Upgrader{} // upgrader := websocket.Upgrader{} can't be used outside a function

var configPath = flag.String("config", "config-server.json", "config path")

func main() {
	fmt.Println("start server")
	flag.Parse()

	config := loadConfig(*configPath)

	listenAddr = config.ListenIp + ":" + strconv.Itoa(config.ListenPort)
	forwardAddr = config.ForwardIp + ":" + strconv.Itoa(config.ForwardPort)
	token = config.Token

	fmt.Println("Lisnten to:", listenAddr, "Forward suspicious requests to:", forwardAddr)

	http.HandleFunc("/", home)

	log.Fatal(http.ListenAndServeTLS(listenAddr, config.Crt, config.Key, nil))
}

func home(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("A request is detected")

	c, err := upgrader.Upgrade(w, r, nil) // upgrade from http to websocket connection(c)
	if err != nil {
		log.Println("Error upgrade:", err)
		return
	}
	defer c.Close()

	_, message, err := c.ReadMessage()
	if err != nil {
		log.Println("Error c.ReadMessage:", err)
		return
	}

	xi_header, err := common.ParseXiHandshake(message, token)
	if err != nil {
		fmt.Println("Xi Handshake failed", err, "Forward it to", forwardAddr)
		//forward()
		return
	}

	//common.DisplayXiHeader(*xi_header)

	//fmt.Println(xi_header)

	target_conn, err := net.Dial("tcp", string(xi_header.Host))
	if err != nil {
		fmt.Println("Dial fail", err)
		return
	}

	target_conn.Write(xi_header.Payload)

	common.Proxy(target_conn, c)
}
