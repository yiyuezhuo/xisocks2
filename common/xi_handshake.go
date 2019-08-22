package common

import "fmt"

type XiHeader struct {
	LenToken byte
	Token    []byte
	LenHost  byte
	Host     []byte
	Payload  []byte
}

func DisplayXiHeader(xi_header XiHeader) {
	fmt.Println("LenToken:", xi_header.LenToken)
	fmt.Println("Token:", xi_header.Token, "parsed:", string(xi_header.Token))
	fmt.Println("LenHost:", xi_header.LenHost)
	fmt.Println("Host:", xi_header.Host, "parsed:", string(xi_header.Host))
	fmt.Println("Payload:", xi_header.Payload, "parsed:", string(xi_header.Payload))
}

func ParseXiHandshake(buf []byte, token string) (xi_header *XiHeader, err error) {

	LenToken := buf[0]
	buf = buf[1:]

	Token := buf[:LenToken]
	buf = buf[LenToken:]

	//fmt.Println("Compare:", string(Token), "vs", token)
	if string(Token) != token {
		fmt.Println("Compare:", string(Token), "vs", token, "token reject")
		return nil, fmt.Errorf("token reject")
	}

	LenHost := buf[0]
	buf = buf[1:]

	Host := buf[:LenHost]
	buf = buf[LenHost:]

	Payload := buf

	xi_header = &XiHeader{LenToken, Token, LenHost, Host, Payload}

	return
}

func BuildXiHandshake(xi_header XiHeader) []byte {
	lenToken := xi_header.LenToken
	token := xi_header.Token
	lenHost := xi_header.LenHost
	host := xi_header.Host
	payload := xi_header.Payload

	message_buff := make([]byte, 1+int(lenToken)+1+int(lenHost)+len(payload))

	it := message_buff
	it[0] = lenToken
	it = it[1:]
	//it[:lenToken] = token
	copy(it, token)
	it = it[lenToken:]
	it[0] = byte(lenHost)
	it = it[1:]
	//it[:lenHost] = host
	copy(it, host)
	it = it[lenHost:]
	//it[:] = payload
	copy(it, payload)

	return message_buff

}
