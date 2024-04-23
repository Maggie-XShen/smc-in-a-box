package generator

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type Server struct {
	Server_ID               string
	Token                   string
	Cert_path               string
	Key_path                string
	Port                    string
	Complaint_urls          []string
	Masked_share_urls       []string
	Dolev_complaint_urls    []string
	Dolev_masked_share_urls []string
	Share_Index             int
	N                       int
	T                       int
	Q                       int
	N_secrets               int
	M                       int
	N_open                  int
}

func GenerateServerConfigLocal(num int, ports []string, src string, des string) {
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

	config := Server{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("unable to read from server_template.json: %s", err)
		return
	}

	for i := 0; i < num; i++ {
		config.Server_ID = "s" + strconv.Itoa(i+1)
		config.Token = "stk" + strconv.Itoa(i+1)
		config.Port = ports[i]
		config.Share_Index = i + 1

		c_urls := make([]string, num-1)
		m_urls := make([]string, num-1)
		dc_urls := make([]string, num-1)
		dm_urls := make([]string, num-1)
		index := 0
		for j := 0; j < num; j++ {
			if j != i {
				c_url := "http://127.0.0.1:" + ports[j] + "/complaint/"
				c_urls[index] = c_url
				m_url := "http://127.0.0.1:" + ports[j] + "/maskedShare/"
				m_urls[index] = m_url
				dc_url := "http://127.0.0.1:" + ports[j] + "/dolevComplaint/"
				dc_urls[index] = dc_url
				dm_url := "http://127.0.0.1:" + ports[j] + "/dolevMaskedShare/"
				dm_urls[index] = dm_url

				index++

			}
		}

		config.Complaint_urls = c_urls
		config.Masked_share_urls = m_urls

		file, _ := json.MarshalIndent(config, "", " ")
		fileName := fmt.Sprintf("config_%s.json", config.Server_ID)
		filePath := filepath.Join(des, fileName)
		_ = os.WriteFile(filePath, file, 0644)
	}
}

func GenerateServerConfigCloud(num int, ip []string, src string, des string) {
	// Ensure the folder exists
	err := os.MkdirAll(des, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating folder:", err)
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

	config := Server{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("unable to read from server_template.json: %s", err)
		return
	}

	for i := 0; i < num; i++ {
		config.Server_ID = "s" + strconv.Itoa(i)
		config.Token = "stk" + strconv.Itoa(i)
		config.Share_Index = i

		c_urls := make([]string, num-1)
		m_urls := make([]string, num-1)
		dc_urls := make([]string, num-1)
		dm_urls := make([]string, num-1)
		index := 0
		for j := 0; j < num; j++ {
			if j != i {
				c_url := ip[j] + config.Port + "/complaint/"
				c_urls[index] = c_url
				m_url := ip[j] + config.Port + "/maskedShare/"
				m_urls[index] = m_url
				dc_url := ip[j] + config.Port + "/dolevComplaint/"
				dc_urls[index] = dc_url
				dm_url := ip[j] + config.Port + "/dolevMaskedShare/"
				dm_urls[index] = dm_url
				index++

			}
		}

		config.Complaint_urls = c_urls
		config.Masked_share_urls = m_urls

		file, _ := json.MarshalIndent(config, "", " ")
		fileName := fmt.Sprintf("config_%s.json", config.Server_ID)
		filePath := filepath.Join(des, fileName)
		_ = os.WriteFile(filePath, file, 0644)
	}
}
