/*
Provide common functions used in server
*/
package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"testing"
)

func assertString(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

func dummyHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("dummyHandler have been called")

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	io.WriteString(w, `{"fuck":"xi"}`)
}

func fakeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("fakeHandler have been called")

	w.WriteHeader(http.StatusOK)

	io.WriteString(w, "fake")
}

func startTempServer(handlerFunc func(http.ResponseWriter, *http.Request)) string {
	listener, err := net.Listen("tcp", "127.0.0.1:0") // 0 = select a free port. :0 = 0.0.0.0:0
	if err != nil {
		log.Panic(err)
	}

	tempAddr := listener.Addr().String()

	go http.Serve(listener, http.HandlerFunc(handlerFunc))

	return tempAddr
}
