package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"example.com/SMC/server/config"
	"github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/stat/combin"
)

var logger *logrus.Logger
var client_size = 6     //total number of clients (no dropout) per experiment
var bad_client_size = 1 // total number of bad client (no dropout) per experiment
var complaint_size int  //total number of complaints from all servers per experiment
var mask_share_size int //total number of masked share records per experiment
var p_sh int            //total number of shares per secret stored by each server

func main() {
	//read configuration
	confpath := flag.String("confpath", "../config/server.json", "config file path")
	inputpath := flag.String("inputpath", "experiments.json", "experiments file path")
	mode := flag.String("mode", "tls", "use tls")
	logpath := flag.String("logpath", "./", "server log path")
	flag.Parse()
	conf := config.Load(*confpath)

	complaint_size = client_size * conf.N
	p_sh = combin.Binomial(conf.N-1, conf.T)
	mask_share_size = conf.N * bad_client_size * conf.N_secrets * p_sh

	logger = logrus.New()
	logger.SetLevel(logrus.DebugLevel)

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
		"party":                   "server",
		"id":                      conf.Server_ID,
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
	}).Info("Server configuration")

	s := NewServer(conf)
	s.HandleExp(*inputpath)

	// set up ticker
	ticker := time.NewTicker(1 * time.Second)
	go s.WaitForEndOfExperiment(ticker)
	go s.WaitForEndOfComplaintBroadcast(ticker)
	go s.WaitForEndOfShareBroadcast(ticker)

	start := time.Now().UTC()
	logger.WithFields(logrus.Fields{
		"start time": start.String(),
	}).Info("Server started")

	if *mode == "tls" {
		s.StartTLS()
	} else {
		s.Start()
	}

}
