package config

import (
	"encoding/json"
	"log"
	"os"
)

type Configuration struct {
	Client_ID string
	URLs      []string
	N         int
	T         int
	K         int
	Q         int
	Exp_ID    string
	Secrets   []int
}

func GetConfig(path string) Configuration {
	file, error := os.Open(path)
	if error != nil {
		log.Fatalf("%s", error)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Configuration{}
	err := decoder.Decode(&config)
	if err != nil {
		log.Fatalf("unable to read from config file: %s", err)
	}

	return config
}
