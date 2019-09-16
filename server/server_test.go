package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/yiyuezhuo/xisocks2/common"
)

/*
func startEcho() (string, chan []byte) {
	ch := make(chan []byte)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Panic(err)
	}

	tempAddr := listener.Addr().String()
	go func() {
		buf := make([]byte, 1024*32)
		for {
			conn, err := listener.Accept()
			readLen, err := conn.Read(buf)
			if err != nil {
				fmt.Println("Read err:", err)
			}
			ch <- buf[:readLen]
		}
	}()

	return tempAddr, ch
}
*/

func TestWrongRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	//forwardAddr = "127.0.0.1:12127"
	//go http.ListenAndServe(forwardAddr, http.HandlerFunc(dummyHandler))
	/*
		listener, err := net.Listen("tcp", "127.0.0.1:0") // 0 = select a free port. :0 = 0.0.0.0:0
		if err != nil {
			log.Panic(err)
		}
	*/

	forwardAddr = startTempServer(fakeHandler)
	fmt.Println("reset forwardAddr to", forwardAddr)

	//go http.Serve(listener, http.HandlerFunc(dummyHandler))

	handler := http.HandlerFunc(home)
	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Errorf("take %v want %v", rr.Code, 200)
	}
	b := rr.Body.Bytes()
	if string(b) != `fake` {
		t.Errorf("take %v want %v", string(b), `fake`)
	}
}

// prevent websocket: response does not implement http.Hijacker
/*
type ResponseWriter interface {
    Header() Header
    Write([]byte) (int, error)
    WriteHeader(statusCode int)
}
*/
/*
type MyRecorder struct{}

func (mr *MyRecorder) Header() http.Header {
	return http.Header{}
}
func (mr *MyRecorder) Write(inp []byte) (num int, err error) {
	return len(inp), nil
}
func (mr *MyRecorder) WriteHeader(statusCode int) {
}
*/

/*
func (mr *MyRecorder) Hijack(net.Conn, *bufio.ReadWriter, error) {
	net.Conn
}
*/

func TestHandshake(t *testing.T) {
	/*
		https://en.wikipedia.org/wiki/WebSocket

		Request

		GET /chat HTTP/1.1
		Host: server.example.com
		Upgrade: websocket
		Connection: Upgrade
		Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==
		Sec-WebSocket-Protocol: chat, superchat
		Sec-WebSocket-Version: 13
		Origin: http://example.com

		Response

		HTTP/1.1 101 Switching Protocols
		Upgrade: websocket
		Connection: Upgrade
		Sec-WebSocket-Accept: HSmrc0sMlYUkAGmm5OPpG2HaGWk=
		Sec-WebSocket-Protocol: chat
	*/

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	//rr := &MyRecorder{}

	//httptest.ResponseRecorder

	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "x3JJHMbDL1EzLkh9GBhXDw==")
	req.Header.Set("Sec-WebSocket-Protocol", "chat, superchat")
	req.Header.Set("Sec-WebSocket-Version", "13")

	//handler := http.HandlerFunc(home)
	//handler.ServeHTTP(rr, req)

	err := homeWithError(rr, req)

	fmt.Println("homeWithError expect a error:", err)

	if err == nil {
		t.Errorf("got nil want a error")
	}

	//response does not implement http.Hijacker
}

func TestGet(t *testing.T) {

	tempAddr := startTempServer(home)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://"+tempAddr, nil)

	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Key", "x3JJHMbDL1EzLkh9GBhXDw==")
	req.Header.Set("Sec-WebSocket-Protocol", "chat, superchat")
	req.Header.Set("Sec-WebSocket-Version", "13")

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	//fmt.Printf("%#v\n", resp)

	assertString(t, resp.Status, "101 Switching Protocols")
	assertString(t, resp.Header.Get("Upgrade"), "websocket")
	assertString(t, resp.Header.Get("Connection"), "Upgrade")
}

func dailAndReadFirstMessage(t *testing.T, proxyAddr, targetAddr, token string) ([]byte, error) {
	t.Helper()

	//targetUrl := "http://" + targetAddr
	targetUrl := targetAddr

	remote_c, _, err := websocket.DefaultDialer.Dial("ws://"+proxyAddr, nil)
	if err != nil {
		t.Errorf("ws dial fail: %v : %v", "ws://"+proxyAddr, err)
	}

	payload := []byte("GET / HTTP/1.1\r\n" + "User-Agent: xisocks2\r\n" +
		"Host: " + targetAddr + "\r\n" + "Accept: */*\r\n" + "\r\n")

	//fmt.Println("payload:", payload, string(payload))

	xi_header := common.XiHeader{
		LenToken: byte(len(token)),
		Token:    []byte(token),
		LenHost:  byte(len(targetUrl)),
		Host:     []byte(targetUrl),
		Payload:  []byte(payload),
	}

	message_buff := common.BuildXiHandshake(xi_header)
	err = remote_c.WriteMessage(websocket.BinaryMessage, message_buff)

	if err != nil {
		return nil, fmt.Errorf("WriteMessage to ws server fail: %v", err)
	}

	_, remote_buf, err := remote_c.ReadMessage()
	/*
		if err != nil {
			t.Errorf("ReadMessage error: %v", err)
		}
	*/

	return remote_buf, err
}

func TestWsDial(t *testing.T) {
	proxyAddr := startTempServer(home)
	targetAddr := startTempServer(dummyHandler)
	fakeAddr := startTempServer(fakeHandler)

	//fmt.Printf("proxyAddr %q targetAddr %q\n", proxyAddr, targetAddr)

	token = "fuckGFW"
	fmt.Printf("reset token to %q \n", token)

	forwardAddr = fakeAddr
	fmt.Println("reset forwardAddr to", forwardAddr)

	t.Run("use wrong token", func(t *testing.T) {
		fmt.Println("testing: use wrong token")
		//remote_buf, err := dailAndReadFirstMessage(t, proxyAddr, targetAddr, "fuckCCP")
		_, err := dailAndReadFirstMessage(t, proxyAddr, targetAddr, "fuckCCP")
		//fmt.Println("wrong return value:", remote_buf, string(remote_buf))
		if err == nil {
			t.Errorf("Wrong ws dial should return some error")
		}
	})

	t.Run("use correct token", func(t *testing.T) {
		fmt.Println("testing: use correct token")
		//remote_buf, err := dailAndReadFirstMessage(t, proxyAddr, targetAddr, token)
		_, err := dailAndReadFirstMessage(t, proxyAddr, targetAddr, token)
		//fmt.Println("correct return value:", remote_buf, string(remote_buf))
		if err != nil {
			t.Errorf("Use correct token encounter unexpected error: %v", err)
		}

	})

}
