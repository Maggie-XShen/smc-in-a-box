package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/SMC/cmd/outputparty/config"
	"example.com/SMC/pkg/packed"
	"example.com/SMC/pkg/repository"
	"example.com/SMC/pkg/utils/message"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type OutputParty struct {
	cfg     *config.OutputParty
	storage *repository.Storage
}

func NewOutputParty(conf *config.OutputParty, storage *repository.Storage) *OutputParty {
	return &OutputParty{cfg: conf, storage: storage}
}

func SetupDatabase(oid string) (*gorm.DB, error) {
	db_name := fmt.Sprintf("%s.db", oid)

	// remove old database
	os.Remove(db_name) //TODO: need to remove

	// open a database
	db, err := gorm.Open(sqlite.Open(db_name), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	log.Printf("Connection to %s Database Established\n", db_name)

	db.AutoMigrate(&repository.Experiment{})

	db.AutoMigrate(&repository.Server{})

	return db, nil
}

func (op *OutputParty) ReadExperiments(path string) {

	// TODO: decide how to get experiment information
	type Tables struct {
		Experiments []message.OutputPartyRequest
	}

	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("%s", err)
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	tables := Tables{}
	err = decoder.Decode(&tables)
	if err != nil {
		log.Fatalf("unable to read from tables file: %s", err)
		return
	}

	expService := NewExperimentService(op.storage)
	for _, exp := range tables.Experiments {
		exp.Due = time.Now().Add(time.Hour * 10).Format("2006-01-02 15:04:05")
		err := expService.CreateExp(exp)
		if err != nil {
			log.Println("error:", err)
		}

		/**
		msg := message.OutputPartyRequest{Exp_ID: exp.Exp_ID, Due: exp.Due}
		fmt.Printf("%+v\n", msg)
		writer := &msg
		for _, url := range op.cfg.URLs {
			message.Send(url, writer.WriteJson())
		}**/
	}

}

func (op *OutputParty) reveal(shares []packed.Share) ([]int, error) {
	//read parameters for packed secret sharing from file
	npss, err := packed.NewPackedSecretSharing(op.cfg.N, op.cfg.T, op.cfg.K, op.cfg.Q)
	if err != nil {
		log.Fatal(err)
	}

	results, err := npss.Reconstruct(shares)
	if err != nil {
		log.Fatal(err)
	}

	return results, nil
}

func (op *OutputParty) WaitForEndOfExperiment(ticker *time.Ticker) {

	for range ticker.C {
		// this gets called every second
		experiments, err := op.storage.GetAllExps()
		if err != nil {
			log.Println("cannot retreive non-completed experiments - error:", err)
		}

		for _, exp := range experiments {

			due, _ := time.Parse("2006-01-02 15:04:05", exp.Due)
			//currentTime := time.Now()
			currentTime := due.Add(5 * time.Minute) //Todo: need to change back to time.Now()

			// check current time is pased due
			if currentTime.After(due) {
				servers, err := op.storage.GetAllServers(exp.Exp_ID)
				if err != nil {
					log.Println("cannot retreive servers records - error:", err)
					continue
				}
				// check if output party receives all servers' shares
				if len(servers) < op.cfg.N {
					// TODO: send message "did not get enough servers" to output party
					//log.Println("did not get enough servers")
					continue
				}

				// reconstruct sum of secrets
				var shares []packed.Share
				for _, server := range servers {
					shares = append(shares, packed.Share{Value: server.SumShare_Value, Index: server.SumShare_Index})
				}

				fmt.Printf("output party receives: %+v\n", shares)

				result, _ := op.reveal(shares)
				fmt.Printf("sum of secrets for %s : %v\n", exp.Exp_ID, result[0])

				//set experiments to completed
				err1 := op.storage.UpdateCompletedExperiment(exp.Exp_ID)
				if err1 != nil {
					log.Println("cannot set experiment to completed - error:", err1)
				}

			}

		}

		// Todo: check if experiment is over and do something (e.g., remove exp and clients information from DB)
	}
}

func (op *OutputParty) serverRequestHandler(rw http.ResponseWriter, req *http.Request) {
	var request message.ServerRequest

	serverService := NewServerService(op.storage)
	err := serverService.CreateServer(request.ReadJson(req))

	if err != nil {
		log.Printf("error: %s\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(rw, err)
		return
	}

	rw.WriteHeader(http.StatusOK)

}

func (op *OutputParty) Start() {
	http.HandleFunc("/serverRequestSubmit/", op.serverRequestHandler)

	log.Fatal(http.ListenAndServe(":"+op.cfg.Port, nil))

}
