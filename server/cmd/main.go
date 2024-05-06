package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"example.com/SMC/server/config"
	"github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/stat/combin"
)

var logger *logrus.Logger
var client_size int     //total number of clients (no dropout) per experiment
var complaint_size int  //total number of complaints from all servers per experiment
var mask_share_size int //total number of masked share records per experiment
var real_client_share_due time.Time
var real_complaint_due time.Time
var real_share_broadcast_due time.Time
var total_verify_time time.Duration
var get_complaints_end time.Duration
var mask_share_end time.Duration
var share_correct_end time.Duration

func main() {
	//read configuration
	confpath := flag.String("confpath", "../config/server.json", "config file path")
	inputpath := flag.String("inputpath", "experiments.json", "experiments file path")
	mode := flag.String("mode", "tls", "use tls")
	logpath := flag.String("logpath", "./", "server log path")
	n_client := flag.Int("n_client", 0, "total client number")
	n_client_mal := flag.Int("n_client_mal", 0, "malicious client number")

	flag.Parse()

	if *n_client == 0 {
		log.Fatal("number of clients in command could not be 0")
	} else {
		client_size = *n_client
	}

	conf := config.Load(*confpath)

	complaint_size = client_size * conf.N
	bad_client_size := *n_client_mal          // total number of bad client (no dropout) per experiment
	p_sh := combin.Binomial(conf.N-1, conf.T) //total number of shares per secret stored by each server
	mask_share_size = conf.N * bad_client_size * conf.N_secrets * p_sh

	logger = logrus.New()
	formatter := &logrus.JSONFormatter{
		DisableTimestamp: true,
	}
	logger.SetFormatter(formatter)
	logger.SetLevel(logrus.DebugLevel)

	// Ensure the log folder exists
	err := os.MkdirAll(*logpath, os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating folder:%s", err)
		return
	}

	// Log to a file
	fileName := fmt.Sprintf("%s.log", conf.Server_ID)
	filePath := filepath.Join(*logpath, fileName)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger.SetOutput(file)
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}

	logger.WithFields(logrus.Fields{
		"id":                      conf.Server_ID,
		"N_clients":               client_size,
		"N":                       conf.N,
		"T":                       conf.T,
		"Q":                       conf.Q,
		"N_secrets":               conf.N_secrets,
		"M":                       conf.M,
		"N_open":                  conf.N_open,
		"Cert_path":               conf.Cert_path,
		"Key_path":                conf.Key_path,
		"Complaint_URLs":          conf.Complaint_urls,
		"Masked_share_urls":       conf.Masked_share_urls,
		"Dolev_complaint_urls":    conf.Dolev_complaint_urls,
		"Dolev_masked_share_urls": conf.Dolev_masked_share_urls,
	}).Info("")

	s := NewServer(conf)
	s.HandleExp(*inputpath)

	// set up ticker
	ticker := time.NewTicker(1 * time.Second)
	go s.WaitForEndOfExperiment(ticker)
	go s.WaitForEndOfComplaintBroadcast(ticker)
	go s.WaitForEndOfShareBroadcast(ticker)
	go s.Close(ticker)

	start := time.Now().UTC()
	logger.WithFields(logrus.Fields{
		"start": start.String(),
	}).Info("")

	if *mode == "tls" {
		s.StartTLS()
	} else {
		s.Start()
	}

}
