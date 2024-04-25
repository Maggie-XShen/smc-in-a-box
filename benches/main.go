package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var client_header = []string{"id", "N_secrets", "N", "M", "N_open", "T", "start", "proof_size", "proof_time", "end"}
var server_header = []string{"id", "N_clients", "N_secrets", "N", "M", "N_open", "T", "client_share_due", "complaint_due", "share_broadcast_due", "start", "avg_verify_time", "num_client_received", "total_verify_time", "real_client_share_due", "real_complaint_due", "mask_share_time", "real_share_broadcast_due", "share_correction_time", "end"}
var op_header = []string{"id", "N_clients", "N_secrets", "N", "T", "client_share_due", "server_share_due", "start", "real_server_share_due", "reconstruction_time", "end"}
var server_mal_header = []string{"id", "N_clients", "N_secrets", "N", "M", "N_open", "T", "client_share_due", "complaint_due", "share_broadcast_due", "start", "avg_verify_time", "num_client_received", "real_client_share_due", "real_complaint_due", "mask_share_time", "real_share_broadcast_due", "share_correction_time", "end"}

func main() {
	logLocation := flag.String("logLocation", "./", "log folder path")
	party := flag.String("party", "", "party type")
	csvLocation := flag.String("csvLocation", "./", "csv file location")
	flag.Parse()

	time := time.Now().UTC()
	if *party == "client" {
		fileName := fmt.Sprintf("client_%s.csv", time.String())
		filePath := filepath.Join(*csvLocation, fileName)
		creatCSV(filePath, *logLocation, client_header)

	} else if *party == "server" {
		fileName := fmt.Sprintf("server_%s.csv", time.String())
		filePath := filepath.Join(*csvLocation, fileName)
		creatCSV(filePath, *logLocation, server_header)

	} else if *party == "outputparty" {
		fileName := fmt.Sprintf("op_%s.csv", time.String())
		filePath := filepath.Join(*csvLocation, fileName)
		creatCSV(filePath, *logLocation, op_header)
	} else if *party == "servermal" {
		fileName := fmt.Sprintf("server_mal_%s.csv", time.String())
		filePath := filepath.Join(*csvLocation, fileName)
		creatCSV(filePath, *logLocation, server_mal_header)

	} else {
		log.Println("party flog cannot be empty")
	}
}

// extractValue extracts the value associated with a given key from a JSON object
func extractValue(entry map[string]interface{}, key string) string {
	if val, ok := entry[key]; ok {
		switch v := val.(type) {
		case int:
			return strconv.Itoa(v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

func creatCSV(csvPath, logLocation string, header []string) {
	csvfile, err := os.Create(csvPath)
	if err != nil {
		log.Fatal(err)
	}
	defer csvfile.Close()

	writer := csv.NewWriter(csvfile)
	defer writer.Flush()

	err = writer.Write(header)
	if err != nil {
		log.Fatal(err)
	}

	logs, err := os.ReadDir(logLocation)
	if err != nil {
		log.Fatal(err)
	}

	for _, log := range logs {
		if strings.HasSuffix(log.Name(), ".log") {
			processLog(logLocation, log.Name(), header, writer)
		}
	}

}

func processLog(location, fileName string, headers []string, writer *csv.Writer) {
	// Open the log file for reading
	logfile, err := os.Open(location + fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer logfile.Close()

	// Initialize the row slice
	var row []string

	// Parse each line of the log file
	scanner := bufio.NewScanner(logfile)
	for scanner.Scan() {
		// Parse JSON from the log line
		var entry map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			panic(err)
		}

		// Extract values for selected headers and write to CSV
		for _, header := range headers {
			value := extractValue(entry, header)
			if value != "" {
				row = append(row, value)
			}
		}
	}

	// Write row to CSV
	if err := writer.Write(row); err != nil {
		panic(err)
	}

}
