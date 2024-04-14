package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Experiment struct {
	Exp_ID         string `json:"Exp_ID"`
	ClientShareDue string `json:"ClientShareDue"`
	ServerShareDue string `json:"ServerShareDue"`
}

func GenerateOPInput(exp_num int, start_time time.Time, t int, des string) {
	// Ensure the folder exists
	err := os.MkdirAll(des, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating folder:", err)
		return
	}

	// List to store data objects
	dataList := make([]Experiment, 0)
	for i := 0; i < exp_num; i++ {
		expID := "exp" + strconv.Itoa(i+1)
		client_share_due := start_time.UTC()
		server_share_due := client_share_due.Add(time.Duration(t) * time.Minute).String()

		data := Experiment{
			Exp_ID:         expID,
			ClientShareDue: client_share_due.String(),
			ServerShareDue: server_share_due,
		}

		dataList = append(dataList, data)
	}

	// Open a new file for writing
	fileName := "experiments.json"
	filePath := filepath.Join(des, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Create a JSON encoder
	encoder := json.NewEncoder(file)

	// Write data objects as a list in the file
	err = encoder.Encode(dataList)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

}
