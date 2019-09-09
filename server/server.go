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

func homeWithError(w http.ResponseWriter, r *http.Request) error {
	//fmt.Println("A request is detected")

	if !tokenListContainsValue(r.Header, "Connection", "upgrade") {
		log.Println("Non-ws connection detected, forward to", forwardAddr)
		forward(forwardAddr, w, r)
		return nil
	}

	c, err := upgrader.Upgrade(w, r, nil) // upgrade from http to websocket connection(c)
	if err != nil {
		return fmt.Errorf("Error upgrade: %v", err)
	}
	defer c.Close()

	_, message, err := c.ReadMessage()
	if err != nil {
		return fmt.Errorf("Error c.ReadMessage: %v", err)
	}

	xi_header, err := common.ParseXiHandshake(message, token)
	if err != nil {
		return fmt.Errorf("Xi Handshake failed %v Forward it to %v", err, forwardAddr)
	}

	//fmt.Println("Xi Handshake accept")
	//common.DisplayXiHeader(*xi_header)
	//fmt.Printf("Dial to %q \n", string(xi_header.Host))

	target_conn, err := net.Dial("tcp", string(xi_header.Host))
	if err != nil {
		return fmt.Errorf("Dial fail %v", err)
	}

	target_conn.Write(xi_header.Payload)

	//fmt.Println("start proxy")
	common.Proxy(target_conn, c)
	//fmt.Println("end proxy")

	return nil
}

func home(w http.ResponseWriter, r *http.Request) {
	err := homeWithError(w, r)
	if err != nil {
		fmt.Println("home:", err)
	}
}
