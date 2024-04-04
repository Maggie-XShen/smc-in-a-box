package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"example.com/SMC/client/config"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func main() {
	//read configuration
	confpath := flag.String("confpath", "../config/client.json", "config file path")
	inputpath := flag.String("inputpath", "input.json", "client input path")
	logpath := flag.String("logpath", "./", "client log path")
	flag.Parse()

	conf := config.Load(*confpath)

	logger = logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Log to a file
	fileName := fmt.Sprintf("%s.log", conf.Client_ID)
	filePath := filepath.Join(*logpath, fileName)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger.SetOutput(file)
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}

	logger.WithFields(logrus.Fields{
		"party":     "client",
		"id":        conf.Client_ID,
		"N":         conf.N,
		"T":         conf.T,
		"Q":         conf.Q,
		"N_secrets": conf.N_secrets,
		"M":         conf.M,
		"N_open":    conf.N_open,
		"URLs":      conf.URLs,
	}).Info("Client configuration")

	client := NewClient(conf)

	start := time.Now().UTC()
	logger.WithFields(logrus.Fields{
		"start time": start.String(),
	}).Info("Client started")

	client.Run(*inputpath)

	end := time.Since(start)
	logger.WithFields(logrus.Fields{
		"duration": end,
	}).Info("Client finished")

}
