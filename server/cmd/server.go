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

	"example.com/SMC/pkg/rss"
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
func (s *Server) HandleExp(path string) {
	type Table struct {
		Experiments []utils.OutputPartyRequest
		//Client_registries []utils.ClientRegistry
	}

	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("%s", err)
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	table := Table{}
	err = decoder.Decode(&table)
	if err != nil {
		log.Fatalf("unable to read experiments: %s", err)
		return
	}

	expService := NewExperimentService(s.store)

	for _, exp := range table.Experiments {
		exp.ClientShareDue = time.Now().UTC().Add(time.Duration(1) * time.Minute).Format("2006-01-02 15:04:05")
		err := expService.CreateExp(exp)
		if err != nil {
			log.Println("error:", err)
		}
	}

}

func (s *Server) clientRequestHandler(rw http.ResponseWriter, req *http.Request) {
	var request utils.ClientRequest

	clientService := NewClientService(s.store)

	err := clientService.CreateClientShare(request.ReadJson(req), s.cfg)

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

func (s *Server) serverComplaintHandler(rw http.ResponseWriter, req *http.Request) {
	var request utils.ComplaintRequest

	serverService := NewServerService(s.store)

	err := serverService.CreateComplaint(request.ReadJson(req))

	if err != nil {
		log.Printf("error: %s\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(rw, err)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (s *Server) serverMaskedSharesHandler(rw http.ResponseWriter, req *http.Request) {
	var request utils.MaskedShareRequest

	serverService := NewServerService(s.store)

	err := serverService.CreateMaskedShares(request.ReadJson(req))

	if err != nil {
		log.Printf("error: %s\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(rw, err)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func generateMask(keys []rss.Share, exp_id, client_id string, input_index, index int) int {
	// need to implement PRF
}

// When experiment's due is triggered, server broadcasts complaint message
func (s *Server) WaitForEndOfExperiment(ticker *time.Ticker) {

	for range ticker.C {

		experiments, err := s.store.GetAllExps()
		if err != nil {
			log.Println("cannot retreive non-completed experiments - error:", err)
		}

		for _, exp := range experiments {

			due, _ := time.Parse("2006-01-02 15:04:05", exp.ClientShareDue)
			currentTime := time.Now().UTC()
			//currentTime := due.Add(5 * time.Minute)

			// check current time is pased due
			if currentTime.After(due) {
				//broadcast complaints to other servers
				complaints, err := s.store.GetAllComplaintsPerServer(exp.Exp_ID, s.cfg.Server_ID)
				if err != nil {
					log.Println("cannot retreive complaints records- error:", err)
					continue
				}

				var set []utils.Complaint
				for _, comp := range complaints {
					set = append(set, utils.Complaint{Client_ID: comp.Client_ID, Complain: comp.Complain, Root: comp.Root})
				}

				message := utils.ComplaintRequest{
					Exp_ID:     exp.Exp_ID,
					Server_ID:  s.cfg.Server_ID,
					Complaints: set,
				}

				for _, address := range s.cfg.Other_server_urls {
					fmt.Printf("server %s sends: %+v\n", s.cfg.Server_ID, message)
					writer := &message
					send(address, writer.ToJson())
				}

				//set round1 to completed
				err = s.store.UpdateRound1Completed(exp.Exp_ID)
				if err != nil {
					log.Println("cannot set round1 to completed - error:", err)
				}

			}

		}
	}
}

// When complaint clock is triggered, server computes valid client set, and invoke vss and broadcast masked shares
func (s *Server) WaitForEndOfComplaintBroadcast(ticker *time.Ticker) {

	for range ticker.C {

		experiments, err := s.store.GetExpsWithSRound1Completed()
		if err != nil {
			log.Println("cannot retreive non-completed experiments - error:", err)
		}

		for _, exp := range experiments {

			due, _ := time.Parse("2006-01-02 15:04:05", exp.ComplaintDue)
			currentTime := time.Now().UTC()
			//currentTime := due.Add(5 * time.Minute)

			// check current time is passed complaint due
			if currentTime.After(due) {

				clients, err := s.store.GetAllClients(exp.Exp_ID)
				if err != nil {
					log.Println("cannot retreive clients records- error:", err)
					continue
				}

				for _, c := range clients {
					complaints, err := s.store.GetAllComplaintsPerClient(exp.Exp_ID, c.Client_ID)
					if err != nil {
						log.Println("cannot retreive complaints records- error:", err)
						continue
					}

					//check if client is valid
					num_isNotComplain := 0
					rootCount := make(map[string]int)
					var maxCount int

					for _, comp := range complaints {
						if !comp.Complain {
							num_isNotComplain += 1

							key := string(comp.Root)
							val, exist := rootCount[key]
							if exist {
								rootCount[key] = val + 1
							} else {
								rootCount[key] = val
							}
						}

					}

					for _, count := range rootCount {
						if count > maxCount {
							maxCount = count
						}
					}

					if num_isNotComplain >= s.cfg.N-s.cfg.T && maxCount >= s.cfg.N-s.cfg.T {
						//add valid client to the table
						err = s.store.InsertValidClient(exp.Exp_ID, c.Client_ID)
						if err != nil {
							log.Println("cannot create valid client record - error:", err)
							continue
						}

						/**
						comp, err := s.store.GetComplaint(exp.Exp_ID, s.cfg.Server_ID, c.Client_ID)
						if err != nil {
							log.Println("cannot get compliant record - error:", err)
							continue
						}**/

						//generate mask and broadcast masked shares to other servers
						if num_isNotComplain < s.cfg.N || len(rootCount) > 1 {
							clientShares, err := s.store.GetClientShares(exp.Exp_ID, c.Client_ID)
							if err != nil {
								log.Println("cannot get client shares record - error:", err)
								continue
							}

							for _, cs := range clientShares {
								//generate mask for a share
								mask := generateMask(keys, cs.Exp_ID, cs.Client_ID, cs.Input_Index, cs.Share_Index)

								//insert mask to mask table
								err := s.store.InsertMask(cs.Exp_ID, cs.Client_ID, cs.Input_Index, cs.Share_Index, mask)
								if err != nil {
									log.Println("cannot add mask record to the table- error:", err)
									continue
								}

								//insert masked share to table
								err = s.store.InsertMaskedShare(cs.Exp_ID, s.cfg.Server_ID, cs.Client_ID, cs.Input_Index, cs.Share_Index, mask+cs.Share_Value)
								if err != nil {
									log.Println("cannot add masked share to the table- error:", err)
									continue
								}

							}

						}

					}

				}

				maskedShares, err := s.store.GetMaskedSharesPerExp(exp.Exp_ID)
				if err != nil {
					log.Println("cannot retreive masked shares record- error:", err)
					continue
				}

				var set []utils.MaskedShare
				for _, mask_sh := range maskedShares {
					set = append(set, utils.MaskedShare{Client_ID: mask_sh.Client_ID, Input_Index: mask_sh.Input_Index, Index: mask_sh.Index, Value: mask_sh.Value})
				}

				message := utils.MaskedShareRequest{
					Exp_ID:       exp.Exp_ID,
					Server_ID:    s.cfg.Server_ID,
					MaskedShares: set,
				}

				//broadcast masked shares to other servers
				for _, address := range s.cfg.Other_server_urls {
					fmt.Printf("server %s sends: %+v\n", s.cfg.Server_ID, message)
					writer := &message
					send(address, writer.ToJson())
				}

				//set round2 to completed
				err = s.store.UpdateRound2Completed(exp.Exp_ID)
				if err != nil {
					log.Println("cannot set round2 to completed - error:", err)
				}

			}

		}
	}
}

// When clock of masked share broadcast is triggered, server recover shares and aggregate shares
func (s *Server) WaitForEndOfShareBroadcast(ticker *time.Ticker) {
	for range ticker.C {

		experiments, err := s.store.GetAllExps()
		if err != nil {
			log.Println("cannot retreive non-completed experiments - error:", err)
		}

		for _, exp := range experiments {

			due, _ := time.Parse("2006-01-02 15:04:05", exp.ShareBroadcastDue)
			currentTime := time.Now().UTC()
			//currentTime := due.Add(5 * time.Minute)

			// check current time is passed complaint due
			if currentTime.After(due) {
				//error correct

				//update valid client table or update client share table

				//compute aggregated share
				clientShares, err := s.store.GetValidClientShares(exp.Exp_ID)
				if err != nil {
					log.Println("cannot retreive valid client shares record- error:", err)
					continue
				}

				var shareSum []rss.Share
				aggreShare := make(map[int]int)
				for _, entry := range clientShares {
					val, exist := aggreShare[entry.Share_Index]
					if exist {
						aggreShare[entry.Share_Index] = val + entry.Share_Value
					} else {
						aggreShare[entry.Share_Index] = entry.Share_Value
					}
				}

				for key, val := range aggreShare {
					shareSum = append(shareSum, rss.Share{Index: key, Value: val})
				}

				// send to output party
				msg := utils.AggregatedShareRequest{Exp_ID: exp.Exp_ID, Server_ID: s.cfg.Server_ID, AggregatedShares: shareSum, Timestamp: time.Now().Format("2006-01-02 15:04:05")}
				fmt.Printf("server %s sends: %+v\n", s.cfg.Server_ID, msg)
				writer := &msg
				send(exp.Owner, writer.ToJson())

				//set round3 to completed
				err = s.store.UpdateRound3Completed(exp.Exp_ID)
				if err != nil {
					log.Println("cannot set round3 to completed - error:", err)
				}
			}

		}

	}
	//TODO: check if experiment is over and do something (e.g., remove exp and clients information from DB)

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
	http.HandleFunc("/serverDataSubmit/", s.serverComplaintHandler)
	http.HandleFunc("/serverDataSubmit/", s.serverMaskedSharesHandler)

	log.Fatal(http.ListenAndServe(":"+s.cfg.Port, nil))

}

/**
func (s *Server) Start(certFile string, keyFile string) {

	http.HandleFunc("/clientRequestSubmit/", s.clientRequestHandler)
	http.HandleFunc("/outputPartyRequestSubmit/", s.expInforHandler)

	log.Fatal(http.ListenAndServeTLS(":"+s.cfg.Port, certFile, keyFile, nil))

}**/
