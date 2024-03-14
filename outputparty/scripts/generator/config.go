package generator

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
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

func GenerateOPConfig(n_op int, ports []string, src string, des string) {
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

	config := OutputParty{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("unable to read from outputparty_template.json: %s", err)
		return
	}

	for i := 0; i < n_op; i++ {
		config.OutputParty_ID = "op" + strconv.Itoa(i+1)
		config.Port = ports[i]

		file, _ := json.MarshalIndent(config, "", " ")
		fileName := fmt.Sprintf("config_%s.json", config.OutputParty_ID)
		filePath := filepath.Join(des, fileName)
		_ = os.WriteFile(filePath, file, 0644)
	}
}
