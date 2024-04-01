package main

import (
	"flag"
	"time"

	"example.com/SMC/server/config"
)

func main() {
	//read configuration
	confpath := flag.String("confpath", "../config/server.json", "config file path")
	inputpath := flag.String("inputpath", "experiments.json", "experiments file path")
	mode := flag.String("mode", "tls", "use tls")
	flag.Parse()
	conf := config.Load(*confpath)

	s := NewServer(conf)

	s.HandleExp(*inputpath)

	// set up ticker
	ticker := time.NewTicker(1 * time.Second)
	go s.WaitForEndOfExperiment(ticker)
	go s.WaitForEndOfComplaintBroadcast(ticker)
	go s.WaitForEndOfShareBroadcast(ticker)

	if *mode == "tls" {
		s.StartTLS()
	} else {
		s.Start()
	}

}
