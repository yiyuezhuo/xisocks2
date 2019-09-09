package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestForward(t *testing.T) {
	forwardAddr = startTempServer(fakeHandler)
	fmt.Println("reset forwardAddr to", forwardAddr)

	hostAddr := startTempServer(home)

	resp, err := http.Get("http://" + hostAddr)
	if err != nil {
		t.Errorf("fail to get: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("resp: %#v\n", resp)

	if resp.StatusCode != 200 {
		t.Errorf("got %v want %v", resp.StatusCode, 200)
	}

	//buf := make([]byte, 2048)
	//num, err := resp.Body.Read(buf)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Unexpected error when read body: %v", err)
	}
	got := string(body)
	if got != "fake" {
		t.Errorf("got %v want %v", got, "fake")
	}

}
