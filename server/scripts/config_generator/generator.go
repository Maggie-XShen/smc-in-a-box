package config_generator

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
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

func GenerateServerConfig(num int, ports []string, src string, des string) {
	//check if destination directory exists
	if _, err := os.Stat(des); os.IsNotExist(err) {
		err := os.Mkdir(des, os.ModePerm)
		if err != nil {
			log.Println(err)
		}

	}

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
		config.Port = ports[i-1]
		config.Share_Index = i

		file, _ := json.MarshalIndent(config, "", " ")
		file_name := fmt.Sprintf("config_%s.json", config.Server_ID)
		_ = os.WriteFile(des+file_name, file, 0644)
	}
}
