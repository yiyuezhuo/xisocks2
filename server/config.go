package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Token       string
	ListenIp    string
	ListenPort  int
	ForwardIp   string
	ForwardPort int
	Crt         string
	Key         string
}

func loadConfig(configPath string) Config {
	// loaf config from json
	jsonFile, err := os.Open(configPath)
	if err != nil {
		log.Panic(err)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var config Config
	json.Unmarshal(byteValue, &config)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(
		"Token:", config.Token,
		"ListenIp:", config.ListenIp,
		"ListenPort:", config.ListenPort,
		"ForwardIp", config.ForwardIp,
		"ForwardPort:", config.ForwardPort,
		"Crt:", config.Crt,
		"Key:", config.Key)

	return config
}
