package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

type Server struct {
	Server_ID   string
	Token       string
	Port        string
	Share_Index int
	URL         string
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

func Generate(num int, ports []string, src string) {
	//read template.json
	template, err := os.Open(src)
	if err != nil {
		log.Fatalf("%s", err)
		return
	}
	defer template.Close()

	decoder := json.NewDecoder(template)

	config := Server{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("unable to read from server_template.json: %s", err)
		return
	}

	for i := 1; i <= num; i++ {
		config.Server_ID = "s" + strconv.Itoa(i)
		config.Token = "stk" + strconv.Itoa(i)
		config.Port = ports[i]
		config.Share_Index = i

		file, _ := json.Marshal(config)
		file_name := fmt.Sprintf("config_%s.json", config.Server_ID)
		_ = os.WriteFile(file_name, file, 0644)
	}
}
