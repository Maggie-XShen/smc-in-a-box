package main

import (
	"flag"
	"log"
	"time"

	"example.com/SMC/cmd/server/config"
	"example.com/SMC/pkg/repository"
)

func main() {
	//read configuration
	confpath := flag.String("confpath", "config/server_template.json", "config file path")
	datapath := flag.String("datapath", "./data.json", "experiments data and registry data path")
	flag.Parse()
	conf := config.Load(*confpath)

	db, err := SetupDatabase(conf.Server_ID)
	if err != nil {
		log.Fatalf("Cannot set up database: %s", err)
	}

	storage := repository.NewStorage(db)

	s := NewServer(conf, storage)

	s.Read(*datapath)

	// set up ticker
	ticker := time.NewTicker(1 * time.Second)
	go s.WaitForEndOfExperiment(ticker)

	s.Start()

}
