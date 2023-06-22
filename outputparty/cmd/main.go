package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"example.com/SMC/outputparty"
	"example.com/SMC/pkg/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetupDatabase(oid string) (*gorm.DB, error) {
	db_name := fmt.Sprintf("outputparty-%s.db", oid)

	// remove old database
	os.Remove(db_name)

	// open a database
	db, err := gorm.Open(sqlite.Open(db_name), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	log.Println("Connection to Database Established")

	return db, nil
}

func main() {
	//read configuration
	confpath := flag.String("confpath", "../config.json", "config file path")
	flag.Parse()

	conf := outputparty.LoadConfig(*confpath)

	db, err := SetupDatabase(conf.OutputParty_ID)
	if err != nil {
		log.Fatalf("Cannot set up database: %s", err)
	}

	storage := repository.NewStorage(db)
	storage.Migrate()

	op := outputparty.NewOutputParty(*conf, *storage)

	//read experiment information from file to database
	op.HandelExpInfor() //TODO: decide how to get experiment information

	op.Start()

	// set up ticker
	ticker := time.NewTicker(1 * time.Second)
	go op.WaitForEndOfExperiment(ticker)

}
