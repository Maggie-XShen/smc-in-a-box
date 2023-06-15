package config

import (
	"encoding/json"
	"log"
	"os"
)

type Configuration struct {
	Client_ID string
	URLs      []string //servers urls
	N         int
	T         int
	K         int
	Q         int
	Exp_ID    string
	Secrets   []int
}

func NewConfig() *Configuration {
	return &Configuration{}
}

func LoadConfig(path string) *Configuration {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("%s", err)
		return nil
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	config := Configuration{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("unable to read from config file: %s", err)
		return nil
	}
	return &config
}
