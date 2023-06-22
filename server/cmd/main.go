package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"example.com/SMC/pkg/repository"
	"example.com/SMC/server"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetupDatabase(sid string) (*gorm.DB, error) {
	db_name := fmt.Sprintf("server-%s.db", sid)

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
	conf := server.LoadConfig(*confpath)

	db, err := SetupDatabase(conf.Server_ID)
	if err != nil {
		log.Fatalf("Cannot set up database: %s", err)
	}

	storage := repository.NewStorage(db)
	storage.Migrate()

	s := server.NewServer(*conf, *storage)

	// set up ticker
	ticker := time.NewTicker(1 * time.Second)
	go s.WaitForEndOfExperiment(ticker)

	s.Start()

}
