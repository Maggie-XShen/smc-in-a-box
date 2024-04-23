package generator

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type Client struct {
	Client_ID string
	Token     string
	URLs      []string //servers urls
	N         int
	T         int
	Q         int
	N_secrets int
	M         int
	N_open    int
}

func GenerateClientConfig(client_num int, src string, des string) {
	// Ensure the folder exists
	err := os.MkdirAll(des, os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating folder:%s", err)
		return
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
		fileName := fmt.Sprintf("config_%s.json", config.Client_ID)
		filePath := filepath.Join(des, fileName)
		_ = os.WriteFile(filePath, file, 0644)
	}
}

func GenerateClientConfigCloud(client_num int, src string, des string) {
	// Ensure the folder exists
	err := os.MkdirAll(des, os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating folder:%s", err)
		return
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

	for i := 0; i < client_num; i++ {
		config.Client_ID = "c" + strconv.Itoa(i)
		config.Token = "t" + strconv.Itoa(i)
		//config.URLs = urls

		file, _ := json.MarshalIndent(config, "", " ")
		fileName := fmt.Sprintf("config_%s.json", config.Client_ID)
		filePath := filepath.Join(des, fileName)
		_ = os.WriteFile(filePath, file, 0644)
	}
}
