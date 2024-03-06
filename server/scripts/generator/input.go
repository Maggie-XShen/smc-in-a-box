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
	Exp_ID            string `json:"Exp_ID"`
	ClientShareDue    string `json:"ClientShareDue"`
	ComplaintDue      string `json:"ComplaintDue"`
	ShareBroadcastDue string `json:"ShareBroadcastDue"`
	Owner             string `json:"Owner"`
}

func GenerateServerInput(exp_num int, start_time string, t1 int, t2 int, owner string, des string) {
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
		client_share_due, _ := time.Parse("2006-01-02 15:04:05", start_time)
		complaint_due := client_share_due.UTC().Add(time.Duration(t1) * time.Minute).Format("2006-01-02 15:04:05")
		share_broadcast_due := client_share_due.UTC().Add(time.Duration(t2) * time.Minute).Format("2006-01-02 15:04:05")

		data := Experiment{
			Exp_ID:            expID,
			ClientShareDue:    start_time,
			ComplaintDue:      complaint_due,
			ShareBroadcastDue: share_broadcast_due,
			Owner:             owner,
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
