package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Token             string
	LocalIp           string
	LocalPort         int
	ProxyURL          string
	ProxyPort         int
	UseConnectionPool bool
	ResolveHTTP       bool
	TLS               bool
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

	fmt.Printf("%#v\n", config)

	return config
}
