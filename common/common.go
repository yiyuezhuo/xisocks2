package common

import (
	"fmt"
	"io"
	"net"

	"github.com/gorilla/websocket"
)

const BUFFER_SIZE = 8192

func Proxy(local_c net.Conn, remote_c *websocket.Conn) {
	defer local_c.Close()
	defer remote_c.Close()

	buf := make([]byte, BUFFER_SIZE)
	go func() {
		for {
			readLen, err := local_c.Read(buf)
			if err != nil {
				//io.Copy(local_c, remote_c)
				if err == io.EOF {
					return
				}
				fmt.Println("proxyed->local error:", err)
				return
			}
			err = remote_c.WriteMessage(websocket.BinaryMessage, buf[:readLen])
			if err != nil {
				if err == io.EOF {
					return
				}
				fmt.Println("local->remote error:", err)
				return
			}
		}
	}()

	for {
		_, remote_buf, err := remote_c.ReadMessage()
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Println("remote->local error:", err)
			return
		}
		writeLen, err := local_c.Write(remote_buf)
		if writeLen != len(remote_buf) {
			fmt.Println("Write fail", writeLen, "vs", remote_buf)
			return
		}
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Println("local->proxyed error:", err)
			return
		}
	}

}
