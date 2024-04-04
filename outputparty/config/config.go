package config

import (
	"encoding/json"
	"log"
	"os"
)

type OutputParty struct {
	OutputParty_ID string
	Cert_path      string
	Key_path       string
	Port           string
	N              int
	T              int
	K              int
	Q              int
}

func Load(path string) *OutputParty {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("%s", err)
		return nil
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	config := OutputParty{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("unable to read from config file: %s", err)
		return nil
	}
	return &config
}
