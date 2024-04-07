package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"example.com/SMC/outputparty/config"
	"github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/stat/combin"
)

var logger *logrus.Logger
var p_sh int //total number of shares per secret stored by each server
var n_sh int //total number of shares a secret splits to

func main() {
	//read configuration
	confpath := flag.String("confpath", "../config/outputparty.json", "config file path")
	inputpath := flag.String("inputpath", "experiments.json", "experiments infor path")
	mode := flag.String("mode", "tls", "use tls")
	logpath := flag.String("logpath", "./", "outputparty log path")
	flag.Parse()

	conf := config.Load(*confpath)
	p_sh = combin.Binomial(conf.N-1, conf.T) //compute total number of shares each party has
	n_sh = p_sh * conf.N

	logger = logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Ensure the log folder exists
	err := os.MkdirAll(*logpath, os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating folder:%s", err)
		return
	}

	// Log to a file
	fileName := fmt.Sprintf("%s.log", conf.OutputParty_ID)
	filePath := filepath.Join(*logpath, fileName)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger.SetOutput(file)
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}
	logger.WithFields(logrus.Fields{
		"party":     "output party",
		"id":        conf.OutputParty_ID,
		"N":         conf.N,
		"T":         conf.T,
		"Q":         conf.Q,
		"N_secrets": conf.K,
		"Port":      conf.Port,
	}).Info("Output Party configuration")

	op := NewOutputParty(conf)

	op.HandelExp(*inputpath) //read experiment information from file to database

	ticker := time.NewTicker(1 * time.Second) // set up ticker
	go op.WaitForEndOfExperiment(ticker)
	go op.Close(ticker)

	start := time.Now().UTC()
	logger.WithFields(logrus.Fields{
		"start time": start.String(),
	}).Info("Output party started")

	if *mode == "tls" {
		op.StartTLS(conf.Cert_path, conf.Key_path)
	} else {
		op.Start()
	}

}
