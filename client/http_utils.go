package main

import (
	"bufio"
	"io"
	"net/http"
	"strings"
)

func http_handshake(conn io.ReadWriter) (string, *http.Request, error) {
	var Host, Port, Method string

	reader := bufio.NewReader(conn)

	req, err := http.ReadRequest(reader)
	if err != nil {
		return "", nil, err
	}

	Method = req.Method

	if strings.Contains(req.URL.Host, ":") {
		host_port := strings.Split(req.URL.Host, ":")
		Host = host_port[0]
		Port = host_port[1]
	} else {
		Host = req.URL.Host
		Port = "80" // Is it possible that port take 443 in some situation?
	}

	host_port := Host + ":" + Port

	if Method == "CONNECT" {
		_, err := conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
		if err != nil {
			return "", nil, err
		}
	}
	return host_port, req, nil

}
