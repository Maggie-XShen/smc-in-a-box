package main

import (
	"flag"
	"time"

	"example.com/SMC/outputparty/config"
)

func main() {
	//read configuration
	confpath := flag.String("confpath", "../config/outputparty.json", "config file path")
	inputpath := flag.String("inputpath", "experiments.json", "experiments infor path")
	mode := flag.String("mode", "tls", "use tls")
	flag.Parse()

	conf := config.Load(*confpath)

	op := NewOutputParty(conf)

	//read experiment information from file to database
	op.HandelExp(*inputpath)

	// set up ticker
	ticker := time.NewTicker(1 * time.Second)
	go op.WaitForEndOfExperiment(ticker)

	if *mode == "tls" {
		op.StartTLS(conf.Cert_path, conf.Key_path)
	} else {
		op.Start()
	}

}
