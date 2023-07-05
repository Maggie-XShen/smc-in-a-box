package main

import (
	"flag"
	"log"
	"time"

	"example.com/SMC/cmd/outputparty/config"
	"example.com/SMC/pkg/repository"
)

func main() {
	//read configuration
	confpath := flag.String("confpath", "config/config.json", "config file path")
	exppath := flag.String("exppath", "./experiments_data.json", "expemeriment data path")
	flag.Parse()

	conf := config.Load(*confpath)

	db, err := SetupDatabase(conf.OutputParty_ID)
	if err != nil {
		log.Fatalf("Cannot set up database: %s", err)
	}

	storage := repository.NewStorage(db)

	op := NewOutputParty(conf, storage)

	//read experiment information from file to database
	op.ReadExperiments(*exppath) //TODO: decide how to get experiment information

	// set up ticker
	ticker := time.NewTicker(1 * time.Second)
	go op.WaitForEndOfExperiment(ticker)

	op.Start()

}
