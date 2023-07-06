package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/SMC/cmd/server/config"
	"example.com/SMC/pkg/packed"
	"example.com/SMC/pkg/repository"
	"example.com/SMC/pkg/utils/message"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Server struct {
	cfg     *config.Server
	storage *repository.Storage
}

func NewServer(conf *config.Server, storage *repository.Storage) *Server {
	return &Server{cfg: conf, storage: storage}
}

func SetupDatabase(sid string) (*gorm.DB, error) {
	db_name := fmt.Sprintf("%s.db", sid)

	// remove old database
	os.Remove(db_name)

	// open a database
	db, err := gorm.Open(sqlite.Open(db_name), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	log.Printf("Connection to %s Database Established\n", db_name)

	db.AutoMigrate(&repository.Experiment{})

	db.AutoMigrate(&repository.Client{})

	db.AutoMigrate(&repository.ClientRegistry{})

	return db, nil
}

/*
*Set up experiments table and client registries table before
client, server and outputparty start communicating
*
*/
func (s *Server) Read(path string) {
	type Tables struct {
		Experiments       []message.OutputPartyRequest
		Client_registries []message.ClientRegistry
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
		log.Fatalf("unable to read experiments and registries data: %s", err)
		return
	}

	expService := NewExperimentService(s.storage)

	for _, exp := range tables.Experiments {
		exp.Due = time.Now().Add(time.Hour * 10).Format("2006-01-02 15:04:05")
		err := expService.CreateExp(exp)
		if err != nil {
			log.Println("error:", err)
		}
	}

	clientService := NewClientService(s.storage)
	for _, reg := range tables.Client_registries {
		err := clientService.CreateClientRegistry(reg)
		if err != nil {
			log.Println("error:", err)
		}
	}

}

func (s *Server) clientRequestHandler(rw http.ResponseWriter, req *http.Request) {
	var request message.ClientRequest

	clientService := NewClientService(s.storage)

	err := clientService.CreateClient(request.ReadJson(req))

	if err != nil {
		log.Printf("error: %s\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(rw, err)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (s *Server) expInforHandler(rw http.ResponseWriter, req *http.Request) {
	var exp message.OutputPartyRequest

	expService := NewExperimentService(s.storage)

	err := expService.CreateExp(exp.ReadJson(req))

	if err != nil {
		log.Printf("error: %s\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(rw, err)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (s *Server) addShares(shares []int) int {
	sumOfShares := 0

	for _, share := range shares {
		sumOfShares += share
	}
	return sumOfShares
}

func (s *Server) WaitForEndOfExperiment(ticker *time.Ticker) {

	for range ticker.C {

		experiments, err := s.storage.GetAllExps()
		if err != nil {
			log.Println("cannot retreive non-completed experiments - error:", err)
		}

		for _, exp := range experiments {

			due, _ := time.Parse("2006-01-02 15:04:05", exp.Due)
			//currentTime := time.Now()
			currentTime := due.Add(5 * time.Minute) //TODO: need to change back to time.Now()

			// check current time is pased due
			if currentTime.After(due) {
				clients, err := s.storage.GetAllClients(exp.Exp_ID)
				if err != nil {
					log.Println("cannot retreive clients records- error:", err)
					continue
				}

				// check if server receive all registered clients' shares
				registeredClients, err := s.storage.GetRegisteredClient(exp.Exp_ID)
				if err != nil {
					log.Println("cannot retreive registered clients- error:", err)
					continue
				}

				if len(clients) < len(registeredClients) {

					// TODO: send message "did not get enough clients" to output party
					//log.Println("did not get enough clients")
					continue
				}

				var shares []int
				for _, client := range clients {
					shares = append(shares, client.Share_Value)
				}

				// sum up the shares
				sumSharesValue := s.addShares(shares)

				// send to output party
				msg := message.ServerRequest{Exp_ID: exp.Exp_ID, Server_ID: s.cfg.Server_ID, Sum_Shares: packed.Share{Index: s.cfg.Share_Index, Value: sumSharesValue}, Timestamp: time.Now().Format("2006-01-02 15:04:05")}
				fmt.Printf("server %s sends: %+v\n", s.cfg.Server_ID, msg)
				writer := &msg
				message.Send(s.cfg.URL, writer.WriteJson())

				//set exp to completed
				err = s.storage.UpdateCompletedExperiment(exp.Exp_ID)
				if err != nil {
					log.Println("cannot set experiment to completed - error:", err)
				}

			}

		}

		//TODO: check if experiment is over and do something (e.g., remove exp and clients information from DB)

	}
}

func (s *Server) Start() {
	http.HandleFunc("/clientRequestSubmit/", s.clientRequestHandler)
	http.HandleFunc("/outputPartyRequestSubmit/", s.expInforHandler)

	log.Fatal(http.ListenAndServe(":"+s.cfg.Port, nil))

}
