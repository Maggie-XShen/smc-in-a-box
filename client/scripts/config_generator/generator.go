package config_generator

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

type Client struct {
	Client_ID string
	Token     string
	URLs      []string //servers urls
	N         int
	T         int
	K         int
	Q         int
	M         int
	N_open    int
}

func GenerateClientConfig(client_num int, src string, des string) {
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

	config := Client{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("unable to read from outputparty_template.json: %s", err)
		return
	}

	for i := 1; i <= client_num; i++ {
		config.Client_ID = "c" + strconv.Itoa(i)
		config.Token = "t" + strconv.Itoa(i)
		//config.URLs = urls

		file, _ := json.MarshalIndent(config, "", " ")
		file_name := fmt.Sprintf("config_%s.json", config.Client_ID)
		_ = os.WriteFile(des+file_name, file, 0644)
	}
}
