package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"example.com/SMC/pkg/rss"
)

type ServerRequest struct {
	Exp_ID    string      `json:"Exp_ID "`
	Server_ID string      `json:"Server_ID"`
	Shares    []rss.Share `json:"Sum_Shares"`
	Timestamp string      `json:"Timestamp"`
}

type OutputPartyRequest struct {
	Exp_ID string `json:"Exp_ID"`
	Due    string `json:"Due"`
	Owner  string `json:"Owner"`
}

type ExpResult struct {
	Exp_ID string `json:"Exp_ID"`
	Result int    `json:"Result"`
}

func (op *OutputPartyRequest) ToJson() []byte {
	msg := &OutputPartyRequest{
		Exp_ID: op.Exp_ID,
		Due:    op.Due,
		Owner:  op.Owner,
	}
	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("Cannot marshall output party request: %s", err)
	}

	return message
}

func (s *ServerRequest) ReadJson(req *http.Request) ServerRequest {
	decoder := json.NewDecoder(req.Body)
	var t ServerRequest
	err := decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Cannot decode server request: %s", err)
	}
	return t
}

func readDataFromFile(filename string) ([]ExpResult, error) {
	var data []ExpResult

	// Check if the file exists
	if _, err := os.Stat(filename); err == nil {
		fileData, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		// Unmarshal the existing data into the slice
		if err := json.Unmarshal(fileData, &data); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func appendDataToFile(filename string, data []ExpResult) error {
	// Marshal the data to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Append the JSON data to the file
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return err
	}

	return nil
}

// write reconstructed result to the file
func WriteResult(id string, result int) {
	expResult := ExpResult{
		Exp_ID: id,
		Result: result,
	}

	// Read existing data
	existingData, err := readDataFromFile("result.json")
	if err != nil {
		panic(err)
	}

	// Append the new record to the existing data
	updatedData := append(existingData, expResult)

	// Write the updated data to the file
	if err := appendDataToFile("result.json", updatedData); err != nil {
		panic(err)
	}

}
