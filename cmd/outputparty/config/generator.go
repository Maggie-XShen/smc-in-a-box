package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

func GenerateOPConfig(op_num int, src string, des string) {
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

	config := OutputParty{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("unable to read from outputparty_template.json: %s", err)
		return
	}

	for i := 1; i <= op_num; i++ {
		config.OutputParty_ID = "op" + strconv.Itoa(i)
		//config.URLs = urls

		file, _ := json.MarshalIndent(config, "", " ")
		file_name := fmt.Sprintf("config_%s.json", config.OutputParty_ID)
		_ = os.WriteFile(des+file_name, file, 0644)
	}
}
