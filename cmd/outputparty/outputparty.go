package main

import (
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
	os.Remove(db_name)

	// open a database
	db, err := gorm.Open(sqlite.Open(db_name), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	log.Println("Connection to Database Established")

	db.AutoMigrate(&repository.Experiment{})

	db.AutoMigrate(&repository.Server{})

	return db, nil
}

func (op *OutputParty) HandelExpInfor() {

	// TODO: decide how to get experiment information
	due := time.Now().Add(time.Hour * 10).Format("2006-01-02 15:04:05")

	var exps []message.OutputPartyRequest
	exps = append(exps, message.OutputPartyRequest{Exp_ID: "exp1", Due: due})
	exps = append(exps, message.OutputPartyRequest{Exp_ID: "exp2", Due: due})

	expService := NewExperimentService(op.storage)
	for _, exp := range exps {
		err := expService.CreateExp(exp)
		if err != nil {
			log.Println("erro:", err)
			return
		}

		msg := message.OutputPartyRequest{Exp_ID: exp.Exp_ID, Due: exp.Due}
		fmt.Printf("%+v\n", msg)
		writer := &msg
		for _, url := range op.cfg.URLs {
			message.Send(url, writer.WriteJson())
		}
	}

}

func (op *OutputParty) reveal(shares []packed.Share) ([]int, error) {
	//read parameters for packed secret sharing from file
	npss, err := packed.NewPackedSecretSharing(op.cfg.N, op.cfg.T, op.cfg.K, op.cfg.Q)
	if err != nil {
		log.Fatal(err)
	}

	result, err := npss.Reconstruct(shares)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("sum of secrets: %v", result)

	return result, nil
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

				op.reveal(shares)

				//set exp to completed
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
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)

}

func (op *OutputParty) Start() {
	http.HandleFunc("/serverRequestSubmit/", op.serverRequestHandler)

	log.Fatal(http.ListenAndServe(":"+op.cfg.Port, nil))

}
