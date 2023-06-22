package server

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Server_ID   string
	Token       string
	Port        string
	Share_Index int
	URL         string
}

func NewConfig() *Config {
	return &Config{}
}

func LoadConfig(path string) *Config {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("%s", err)
		return nil
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("unable to read from server config file: %s", err)
		return nil
	}
	return &config
}
