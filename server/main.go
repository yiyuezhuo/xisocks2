package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"unicode/utf8"

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

	if !tokenListContainsValue(r.Header, "Connection", "upgrade") {
		log.Println("forward to", forwardAddr)
		forward(w, r)
	}

	c, err := upgrader.Upgrade(w, r, nil) // upgrade from http to websocket connection(c)
	if err != nil {
		log.Println("Error upgrade:", err)
		//forward(w, r)
		return
	}
	defer c.Close()

	_, message, err := c.ReadMessage()
	if err != nil {
		log.Println("Error c.ReadMessage:", err)
		//forward(w, r)
		return
	}

	xi_header, err := common.ParseXiHandshake(message, token)
	if err != nil {
		fmt.Println("Xi Handshake failed", err, "Forward it to", forwardAddr)
		//forward(w, r)
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

func forward(w http.ResponseWriter, req *http.Request) {
	// https://stackoverflow.com/questions/34724160/go-http-send-incoming-http-request-to-an-other-server-using-client-do
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// you can reassign the body if you need to parse it as multipart
	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	// create a new url from the raw RequestURI sent by the client
	url := fmt.Sprintf("%s://%s%s", "http", forwardAddr, "")

	proxyReq, err := http.NewRequest(req.Method, url, bytes.NewReader(body))

	// We may want to filter some headers, otherwise we could just use a shallow copy
	// proxyReq.Header = req.Header
	proxyReq.Header = make(http.Header)
	for h, val := range req.Header {
		proxyReq.Header[h] = val
	}

	httpClient := &http.Client{}

	resp, err := httpClient.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// https://stackoverflow.com/questions/28891531/piping-http-response-to-http-responsewriter
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
	io.Copy(w, resp.Body)
	//resp.Body.Close()

}

// Follwing code are brought from github.com/gorilla/websocket/util since it's a private function
func tokenListContainsValue(header http.Header, name string, value string) bool {

headers:
	for _, s := range header[name] {
		for {
			var t string
			t, s = nextToken(skipSpace(s))
			if t == "" {
				continue headers
			}
			s = skipSpace(s)
			if s != "" && s[0] != ',' {
				continue headers
			}
			if equalASCIIFold(t, value) {
				return true
			}
			if s == "" {
				continue headers
			}
			s = s[1:]
		}
	}
	return false
}

// skipSpace returns a slice of the string s with all leading RFC 2616 linear
// whitespace removed.
func skipSpace(s string) (rest string) {
	i := 0
	for ; i < len(s); i++ {
		if b := s[i]; b != ' ' && b != '\t' {
			break
		}
	}
	return s[i:]
}

// nextToken returns the leading RFC 2616 token of s and the string following
// the token.
func nextToken(s string) (token, rest string) {
	i := 0
	for ; i < len(s); i++ {
		if !isTokenOctet[s[i]] {
			break
		}
	}
	return s[:i], s[i:]
}

// equalASCIIFold returns true if s is equal to t with ASCII case folding as
// defined in RFC 4790.
func equalASCIIFold(s, t string) bool {
	for s != "" && t != "" {
		sr, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		tr, size := utf8.DecodeRuneInString(t)
		t = t[size:]
		if sr == tr {
			continue
		}
		if 'A' <= sr && sr <= 'Z' {
			sr = sr + 'a' - 'A'
		}
		if 'A' <= tr && tr <= 'Z' {
			tr = tr + 'a' - 'A'
		}
		if sr != tr {
			return false
		}
	}
	return s == t
}

var isTokenOctet = [256]bool{
	'!':  true,
	'#':  true,
	'$':  true,
	'%':  true,
	'&':  true,
	'\'': true,
	'*':  true,
	'+':  true,
	'-':  true,
	'.':  true,
	'0':  true,
	'1':  true,
	'2':  true,
	'3':  true,
	'4':  true,
	'5':  true,
	'6':  true,
	'7':  true,
	'8':  true,
	'9':  true,
	'A':  true,
	'B':  true,
	'C':  true,
	'D':  true,
	'E':  true,
	'F':  true,
	'G':  true,
	'H':  true,
	'I':  true,
	'J':  true,
	'K':  true,
	'L':  true,
	'M':  true,
	'N':  true,
	'O':  true,
	'P':  true,
	'Q':  true,
	'R':  true,
	'S':  true,
	'T':  true,
	'U':  true,
	'W':  true,
	'V':  true,
	'X':  true,
	'Y':  true,
	'Z':  true,
	'^':  true,
	'_':  true,
	'`':  true,
	'a':  true,
	'b':  true,
	'c':  true,
	'd':  true,
	'e':  true,
	'f':  true,
	'g':  true,
	'h':  true,
	'i':  true,
	'j':  true,
	'k':  true,
	'l':  true,
	'm':  true,
	'n':  true,
	'o':  true,
	'p':  true,
	'q':  true,
	'r':  true,
	's':  true,
	't':  true,
	'u':  true,
	'v':  true,
	'w':  true,
	'x':  true,
	'y':  true,
	'z':  true,
	'|':  true,
	'~':  true,
}
