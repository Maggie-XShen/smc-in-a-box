package config

import (
	"encoding/json"
	"log"
	"os"
)

type Server struct {
	Server_ID         string
	Token             string
	Cert_path         string
	Key_path          string
	Port              string
	Other_server_urls []string
	Share_Index       int
	N                 int
	T                 int
	K                 int
	Q                 int
	N_claims          int
	M                 int
	N_open            int
}

func NewConfig() *Server {
	return &Server{}
}

func Load(path string) *Server {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("%s", err)
		return nil
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	config := Server{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("unable to read from server config file: %s", err)
		return nil
	}
	return &config
}
