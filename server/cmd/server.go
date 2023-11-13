package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/SMC/pkg/ligero"
	"example.com/SMC/pkg/packed"
	"example.com/SMC/server/config"
	"example.com/SMC/server/public/utils"
	"example.com/SMC/server/sqlstore"
)

type Server struct {
	cfg   *config.Server
	store *sqlstore.SqlStore
}

func NewServer(conf *config.Server) *Server {
	return &Server{cfg: conf, store: sqlstore.New(conf.Server_ID)}
}

/*
*Set up experiments table and client registries table before
client, server and outputparty start communicating
*
*/
func (s *Server) HandleExpAndRegistry(path string) {
	type Tables struct {
		Experiments       []utils.OutputPartyRequest
		Client_registries []utils.ClientRegistry
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

	expService := NewExperimentService(s.store)

	for _, exp := range tables.Experiments {
		exp.Due = time.Now().UTC().Add(time.Duration(1) * time.Minute).Format("2006-01-02 15:04:05")
		err := expService.CreateExp(exp)
		if err != nil {
			log.Println("error:", err)
		}
	}

	clientService := NewClientService(s.store)
	for _, reg := range tables.Client_registries {
		err := clientService.CreateClientRegistry(reg)
		if err != nil {
			log.Println("error:", err)
		}
	}

}

func (s *Server) clientRequestHandler(rw http.ResponseWriter, req *http.Request) {
	var request utils.ClientRequest
	data := request.ReadJson(req)

	zk, err := ligero.NewLigeroZK(s.cfg.K, s.cfg.M, s.cfg.N, s.cfg.T, s.cfg.Q, s.cfg.N_open)
	if err != nil {
		log.Fatal(err)
	}

	//check client's proof
	verify, err := zk.Verify(data.Proof)
	if err != nil {
		log.Fatal(err)
	}
	if !verify {
		panic("failed verifification!")
	}
	fmt.Println("verification succeed!")

	clientService := NewClientService(s.store)

	err = clientService.CreateClient(data)

	if err != nil {
		log.Printf("error: %s\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(rw, err)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (s *Server) expInforHandler(rw http.ResponseWriter, req *http.Request) {
	var exp utils.OutputPartyRequest

	expService := NewExperimentService(s.store)

	err := expService.CreateExp(exp.ReadJson(req))

	if err != nil {
		log.Printf("error: %s\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(rw, err)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (s *Server) serverDataHandler(rw http.ResponseWriter, req *http.Request) {
	var request utils.ClientSet

	serverService := NewServerService(s.store)

	err := serverService.CreateClientSet(request.ReadJson(req))

	if err != nil {
		log.Printf("error: %s\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(rw, err)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func addShares(clients []sqlstore.ClientShare) int {
	sumOfShares := 0

	for _, client := range clients {
		sumOfShares += client.Share_Value
	}
	return sumOfShares
}

func (s *Server) WaitForEndOfExperiment(ticker *time.Ticker) {

	for range ticker.C {

		experiments, err := s.store.GetAllExps()
		if err != nil {
			log.Println("cannot retreive non-completed experiments - error:", err)
		}

		for _, exp := range experiments {

			due, _ := time.Parse("2006-01-02 15:04:05", exp.Due)
			currentTime := time.Now().UTC()
			//currentTime := due.Add(5 * time.Minute)

			// check current time is pased due
			if currentTime.After(due) {
				clients, err := s.store.GetAllClients(exp.Exp_ID)
				if err != nil {
					log.Println("cannot retreive clients records- error:", err)
					continue
				}

				// check if server receive all registered clients' shares
				/**
				registeredClients, err := s.storage.GetRegisteredClient(exp.Exp_ID)
				if err != nil {
					log.Println("cannot retreive registered clients- error:", err)
					continue
				}

				if len(clients) < len(registeredClients) {

					// TODO: send message "did not get enough clients" to output party
					//log.Println("did not get enough clients")
					continue
				}**/

				//write client set to the table
				var set []string
				for _, client := range clients {
					set = append(set, client.Client_ID)
				}

				entry := utils.GenerateClientSetRecord(exp.Exp_ID, s.cfg.Server_ID, set)

				serverService := NewServerService(s.store)
				err = serverService.CreateClientSet(entry)
				if err != nil {
					log.Println("cannot create client set entry for server itself- error:", err)
				}

				//send client set to other servers
				for _, address := range s.cfg.Other_server_urls {
					fmt.Printf("server %s sends: %+v\n", s.cfg.Server_ID, entry)
					writer := &entry
					send(address, writer.ToJson())
				}

				//set server round to completed
				err = s.store.UpdateHalfCompletedExperiment(exp.Exp_ID)
				if err != nil {
					log.Println("cannot set experiment to completed - error:", err)
				}

			}

		}
	}
}

func (s *Server) WaitForEndOfServersRound(ticker *time.Ticker) {
	for range ticker.C {

		experiments, err := s.store.GetAllExpsWithServerRoundCompleted()
		if err != nil {
			log.Println("cannot retreive non-completed experiments - error:", err)
		}

		for _, exp := range experiments {

			//check if server received all other servers' client set
			sets, err := s.store.GeAllClientSets(exp.Exp_ID)
			if err != nil {
				log.Println("cannot retreive client sets- error:", err)
				continue
			}

			fmt.Printf("order of set: %d\n", len(sets))
			if len(sets) < len(s.cfg.Other_server_urls)+1 {
				//log.Println("did not get enough client sets")
				continue
			}

			//find intersection of client sets
			intersection := s.findClientSetsIntersection(exp.Exp_ID, sets)

			// sum up the shares
			sumSharesValue := addShares(intersection)

			// send to output party
			msg := utils.ServerRequest{Exp_ID: exp.Exp_ID, Server_ID: s.cfg.Server_ID, Sum_Shares: packed.Share{Index: s.cfg.Share_Index, Value: sumSharesValue}, Timestamp: time.Now().Format("2006-01-02 15:04:05")}
			fmt.Printf("server %s sends: %+v\n", s.cfg.Server_ID, msg)
			writer := &msg
			send(exp.Owner, writer.ToJson())

			//set exp to completed
			err = s.store.UpdateCompletedExperiment(exp.Exp_ID)
			if err != nil {
				log.Println("cannot set experiment to completed - error:", err)
			}

		}

	}
	//TODO: check if experiment is over and do something (e.g., remove exp and clients information from DB)

}

func (s *Server) findClientSetsIntersection(exp_id string, sets []sqlstore.ClientSet) []sqlstore.ClientShare {
	// Create a map to keep track of the occurrences of each client
	countMap := make(map[string]int)

	// Iterate through each set
	for _, set := range sets {
		var ids []string
		json.Unmarshal(set.Clients, &ids)
		for _, client_id := range ids {
			countMap[client_id]++
		}
	}

	var intersection []sqlstore.ClientShare
	for client_id, count := range countMap {
		if count == len(sets) {
			client, err := s.store.GetClient(exp_id, client_id)
			if err != nil {
				log.Println("cannot retreive client records- error:", err)
				continue
			}
			intersection = append(intersection, *client)
		}
	}

	return intersection
}

func send(address string, data []byte) {
	req, err := http.NewRequest("POST", address, bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("impossible to build http post request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("impossible to send http request: %s", err)
	}

	log.Printf("response Status:%s", res.Status)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if len(body) > 0 {
		fmt.Println("response Body:", string(body))
	}

}

func (s *Server) Start() {

	http.HandleFunc("/clientRequestSubmit/", s.clientRequestHandler)
	http.HandleFunc("/outputPartyRequestSubmit/", s.expInforHandler)
	http.HandleFunc("/serverDataSubmit/", s.serverDataHandler)

	log.Fatal(http.ListenAndServe(":"+s.cfg.Port, nil))

}

/**
func (s *Server) Start(certFile string, keyFile string) {

	http.HandleFunc("/clientRequestSubmit/", s.clientRequestHandler)
	http.HandleFunc("/outputPartyRequestSubmit/", s.expInforHandler)

	log.Fatal(http.ListenAndServeTLS(":"+s.cfg.Port, certFile, keyFile, nil))

}**/
