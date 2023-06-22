package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"example.com/SMC/pkg/api"
	"example.com/SMC/pkg/message"
	"example.com/SMC/pkg/packed"
	"example.com/SMC/pkg/repository"
)

type Server struct {
	cfg     Config
	storage repository.Storage
}

func NewServer(conf Config, storage repository.Storage) *Server {
	return &Server{cfg: conf, storage: storage}
}

func (s *Server) HandelRegistration() {

	// Todo: decide how to get client registry information
	var regs []message.ClientRegistry
	regs = append(regs, message.ClientRegistry{Exp_ID: "exp1", Client_ID: "c1", Token: "tk1"})
	regs = append(regs, message.ClientRegistry{Exp_ID: "exp1", Client_ID: "c2", Token: "tk2"})
	//regs = append(regs, message.ClientRegistry{Exp_ID: "exp1", Client_ID: "c3", Token: "tk3"})

	clientService := api.NewClientService(s.storage)
	for _, reg := range regs {
		err := clientService.CreateClientRegistry(reg)
		if err != nil {
			log.Println("erro:", err)
			return
		}
	}

}

func (s *Server) clientRequestHandler(rw http.ResponseWriter, req *http.Request) {
	var request message.ClientRequest

	clientService := api.NewClientService(s.storage)
	err := clientService.CreateClient(request.ReadJson(req))

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (s *Server) expInforHandler(rw http.ResponseWriter, req *http.Request) {
	var exp message.OutputPartyRequest

	expService := api.NewExperimentService(s.storage)
	err := expService.CreateExp(exp.ReadJson(req))

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	result, _ := s.storage.GetAllExps()
	if len(result) == 3 {
		s.HandelRegistration()
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

			currentTime := time.Now()
			due, _ := time.Parse("2006-01-02 15:04:05", exp.Due)

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
				fmt.Printf("%+v\n", msg)
				writer := &msg
				message.Send(s.cfg.URL, writer.WriteJson())

				//set exp to completed
				err = s.storage.UpdateCompletedExperiment(exp.Exp_ID)
				if err != nil {
					log.Println("cannot set experiment to completed - error:", err)
				}

			}

		}

		// check if experiment is over and do something (e.g., remove exp and clients information from DB)

	}
}

func (s *Server) Start() {
	http.HandleFunc("/clientRequestSubmit/", s.clientRequestHandler)
	http.HandleFunc("/outputPartyRequestSubmit/", s.expInforHandler)

	log.Fatal(http.ListenAndServe(":"+s.cfg.Port, nil))

}
