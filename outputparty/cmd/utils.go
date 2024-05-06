package main

import (
	"compress/gzip"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"example.com/SMC/pkg/rss"
)

type AggregatedShareRequest struct {
	Exp_ID    string      `json:"Exp_ID "`
	Server_ID string      `json:"Server_ID"`
	Timestamp string      `json:"Timestamp"`
	Shares    []rss.Party `json:"Shares"`
}

type OutputPartyRequest struct {
	Exp_ID         string `json:"Exp_ID"`
	ClientShareDue string `json:"ClientShareDue"`
	Owner          string `json:"Owner"`
}

type Experiment struct {
	Exp_ID         string
	ClientShareDue string
	ServerShareDue string
}

type ExpResult struct {
	Exp_ID string `json:"Exp_ID"`
	Result []int  `json:"Result"`
}

func (op *OutputPartyRequest) ToJson() []byte {
	msg := &OutputPartyRequest{
		Exp_ID:         op.Exp_ID,
		ClientShareDue: op.ClientShareDue,
		Owner:          op.Owner,
	}
	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("Cannot marshall output party request: %s", err)
	}

	return message
}

func (s *AggregatedShareRequest) ReadJson(req *http.Request) AggregatedShareRequest {
	// Decompress the data using Gzip
	gzipReader, err := gzip.NewReader(req.Body)
	if err != nil {
		log.Fatalf("Cannot decompress server request: %s", err)
	}
	defer gzipReader.Close()

	decoder := json.NewDecoder(gzipReader)

	var t AggregatedShareRequest
	err = decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Cannot decode aggregated share request: %s", err)
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
func WriteResult(id string, result []int) {
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

func ReadOutputPartyInput(path string) []Experiment {
	jsonData, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("%s", err)
		return nil
	}

	var items []Experiment
	err = json.Unmarshal(jsonData, &items)
	if err != nil {
		log.Fatalf("%s", err)
		return nil
	}
	return items

}
