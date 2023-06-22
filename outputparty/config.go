package outputparty

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	OutputParty_ID string
	Port           string
	N              int
	T              int
	K              int
	Q              int
	URLs           []string //servers urls
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
		log.Fatalf("unable to read from config file: %s", err)
		return nil
	}
	return &config
}
