package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"example.com/SMC/pkg/rss"
	"example.com/SMC/server/config"
	"example.com/SMC/server/sqlstore"
)

type Server struct {
	cfg   *config.Server
	store *sqlstore.DB
}

func NewServer(conf *config.Server) *Server {
	return &Server{cfg: conf, store: sqlstore.NewDB(conf.Server_ID)}
}

/*
*Set up experiments table before
client, server and outputparty start communicating
*
*/
func (s *Server) HandleExp(path string) {
	experiments := ReadServerInput(path)

	expService := NewExperimentService(s.store)

	for _, exp := range experiments {
		//exp.ClientShareDue = time.Now().UTC().Add(time.Duration(1) * time.Minute).Format("2006-01-02 15:04:05")
		err := expService.CreateExperiment(exp)
		if err != nil {
			log.Println("error:", err)
		}
	}

}

func (s *Server) clientRequestHandler(rw http.ResponseWriter, req *http.Request) {
	var request ClientRequest

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

/**
func (s *Server) expInforHandler(rw http.ResponseWriter, req *http.Request) {
	var exp utils.OutputPartyRequest

	expService := NewExperimentService(s.store)

	err := expService.CreateExperiment(exp.ReadJson(req))

	if err != nil {
		log.Printf("error: %s\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(rw, err)
		return
	}

	rw.WriteHeader(http.StatusOK)
}**/

func (s *Server) serverComplaintHandler(rw http.ResponseWriter, req *http.Request) {
	var request ComplaintRequest

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
	var request MaskedShareRequest

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

func count(complaints []sqlstore.Complaint) (int, int, int) {
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
				rootCount[key] = 1
			}
		}

	}

	for _, count := range rootCount {
		if count > maxCount {
			maxCount = count
		}
	}

	return num_isNotComplain, len(rootCount), maxCount
}

/**
func generateMaskedShareMap(input []sqlstore.MaskedShare) (map[int]map[string][]rss.Share, error) {
	if len(input) <= 0 {
		return nil, fmt.Errorf("masked shares get from table is empty")
	}
	inputMaskedShares := make(map[int]map[string][]rss.Share)
	for _, record := range input {
		v1, check1 := inputMaskedShares[record.Input_Index]
		if check1 {
			v2, check2 := v1[record.Server_ID]
			if check2 {
				inputMaskedShares[record.Input_Index][record.Server_ID] = append(v2, rss.Share{Index: record.Index, Value: record.Value})
			} else {
				inputMaskedShares[record.Input_Index][record.Server_ID] = []rss.Share{{Index: record.Index, Value: record.Value}}
			}
		} else {
			inputMaskedShares[record.Input_Index] = make(map[string][]rss.Share)
			inputMaskedShares[record.Input_Index][record.Server_ID] = []rss.Share{{Index: record.Index, Value: record.Value}}
		}
	}

	return inputMaskedShares, nil
}**/

func computeMajority(input map[string][]rss.Share) ([]rss.Share, error) {
	if len(input) <= 0 {
		return nil, fmt.Errorf("masked shares map is empty")
	}

	shareValue := make(map[int][]int)
	for _, shares := range input {
		for _, sh := range shares {
			v1, check1 := shareValue[sh.Index]
			if check1 {
				shareValue[sh.Index] = append(v1, sh.Value)
			} else {
				shareValue[sh.Index] = []int{sh.Value}
			}
		}
	}

	result := make([]rss.Share, len(shareValue))
	for index, values := range shareValue {
		val, err := FindMajority(values)
		if err != nil {
			return nil, err
		}

		result = append(result, rss.Share{Index: index, Value: val})
	}

	return result, nil
}

// When experiment's due is triggered, server broadcasts complaint message
func (s *Server) WaitForEndOfExperiment(ticker *time.Ticker) {

	for range ticker.C {

		experiments, err := s.store.GetAllExperiments()
		if err != nil {
			//log.Println("cannot retreive non-completed experiments - error:", err)
			panic(err)
		}

		for _, exp := range experiments {

			due, _ := time.Parse("2006-01-02 15:04:05", exp.ClientShareDue)
			currentTime := time.Now().UTC()
			//due := time.Now().UTC()
			//currentTime := due.Add(1 * time.Minute)

			if currentTime.After(due) {
				complaints, err := s.store.GetComplaintsPerServer(exp.Exp_ID, s.cfg.Server_ID)
				if err != nil {
					//log.Println("cannot retreive complaints records- error:", err)
					//continue
					panic(err)
				}

				if len(complaints) == 0 {
					log.Println("error: complaints table is empty")
					continue
				}

				var set []Complaint
				for _, comp := range complaints {
					set = append(set, Complaint{Client_ID: comp.Client_ID, Complain: comp.Complain, Root: comp.Root})
				}

				message := ComplaintRequest{
					Exp_ID:     exp.Exp_ID,
					Server_ID:  s.cfg.Server_ID,
					Complaints: set,
				}

				for _, address := range s.cfg.Complaint_urls {
					fmt.Printf("server %s sends complaints: %+v\n", s.cfg.Server_ID, message)
					writer := &message
					send(address, writer.ToJson())
				}

				//set round1 to completed
				err = s.store.UpdateRound1Completed(exp.Exp_ID)
				if err != nil {
					//log.Println("cannot set round1 to completed - error:", err)
					panic(err)
				}

			}

		}
	}
}

// When complaint clock is triggered, server computes valid client set, and invoke vss and broadcast masked shares
func (s *Server) WaitForEndOfComplaintBroadcast(ticker *time.Ticker) {

	for range ticker.C {

		experiments, err := s.store.GetExpsWithRound1Completed()
		if err != nil {
			//log.Println("cannot retreive non-completed experiments - error:", err)
			panic(err)
		}

		for _, exp := range experiments {

			due, _ := time.Parse("2006-01-02 15:04:05", exp.ComplaintDue)
			currentTime := time.Now().UTC()

			if currentTime.After(due) {
				//find dropout clients for the server
				dropout, err := s.store.GetDropoutClient(exp.Exp_ID)
				if err != nil {
					//log.Println("cannot retreive missing clients- error:", err)
					//continue
					panic(err)
				}

				//generate complaint of dropout client
				for _, client_id := range dropout {
					err := s.store.InsertClient(exp.Exp_ID, client_id)
					if err != nil {
						//log.Fatal("cannot insert missing client to the client table - error:", err)
						panic(err)
					}

					err = s.store.InsertComplaint(exp.Exp_ID, s.cfg.Server_ID, client_id, true, []byte("default"))
					if err != nil {
						//log.Fatal("cannot insert complaint of missing client to the complaint table - error:", err)
						panic(err)
					}
				}

				clients, err := s.store.GetClientsPerExperiment(exp.Exp_ID)
				if err != nil {
					//log.Fatal("cannot retreive clients records - error:", err)
					panic(err)
				}

				//generate valid client set and trigger mask generateion when condition meets
				for _, c := range clients {
					complaints, err := s.store.GetComplaintsPerClient(exp.Exp_ID, c.Client_ID)
					if err != nil {
						//log.Fatal("cannot retreive complaints records - error:", err)
						panic(err)
					}

					num_isNotComplain, rootCount, maxCount := count(complaints)

					if num_isNotComplain >= s.cfg.N-s.cfg.T && maxCount >= s.cfg.N-s.cfg.T {
						err = s.store.InsertValidClient(exp.Exp_ID, c.Client_ID)
						if err != nil {
							//log.Fatal("cannot create valid client record - error:", err)
							panic(err)
						}

						//generate mask and masked shares
						if num_isNotComplain < s.cfg.N || rootCount > 1 {
							clientShares, err := s.store.GetClientShares(exp.Exp_ID, c.Client_ID)
							if err != nil {
								//log.Fatal("cannot get client shares record - error:", err)
								panic(err)
							}

							for _, cs := range clientShares {
								key := 1 //TODO: change key to offline generated key
								crs := NewCryptoRandSource()
								crs.Seed(key, cs.Exp_ID, cs.Client_ID, cs.Input_Index, cs.Index)
								mask := int(crs.Int63(int64(s.cfg.Q)))

								err := s.store.InsertMask(cs.Exp_ID, cs.Client_ID, cs.Input_Index, cs.Index, mask)
								if err != nil {
									//log.Fatal("cannot add mask record to the table - error:", err)
									panic(err)
								}

								err = s.store.InsertMaskedShare(cs.Exp_ID, s.cfg.Server_ID, cs.Client_ID, cs.Input_Index, cs.Index, mask+cs.Value)
								if err != nil {
									//log.Fatal("cannot add masked share to the table - error:", err)
									panic(err)
								}

							}

						}

					}

				}

				maskedShares, err := s.store.GetMaskedSharesPerServer(exp.Exp_ID, s.cfg.Server_ID)
				if err != nil {
					//log.Fatal("cannot retreive masked shares record - error:", err)
					panic(err)

				}

				if len(maskedShares) > 0 {
					var set []MaskedShare
					for _, mask_sh := range maskedShares {
						set = append(set, MaskedShare{Client_ID: mask_sh.Client_ID, Input_Index: mask_sh.Input_Index, Index: mask_sh.Index, Value: mask_sh.Value})
					}

					message := MaskedShareRequest{
						Exp_ID:       exp.Exp_ID,
						Server_ID:    s.cfg.Server_ID,
						MaskedShares: set,
					}

					for _, address := range s.cfg.Masked_share_urls {
						log.Printf("server %s is sending to %s: %+v\n", s.cfg.Server_ID, address, message)
						writer := &message
						send(address, writer.ToJson())
					}

				}

				//set round2 to completed
				err = s.store.UpdateRound2Completed(exp.Exp_ID)
				if err != nil {
					//log.Println("cannot set round2 to completed - error:", err)
					panic(err)
				}

			}

		}
	}
}

// When clock of masked share broadcast is triggered, server recover shares and aggregate shares
func (s *Server) WaitForEndOfShareBroadcast(ticker *time.Ticker) {
	for range ticker.C {

		experiments, err := s.store.GetExpsWithRound2Completed()
		if err != nil {
			//log.Println("cannot retreive non-completed experiments - error:", err)
			panic(err)
		}

		for _, exp := range experiments {

			due, _ := time.Parse("2006-01-02 15:04:05", exp.ShareBroadcastDue)
			currentTime := time.Now().UTC()

			if currentTime.After(due) {

				valid_clients, err := s.store.GetValidClientsPerExperiment(exp.Exp_ID)
				if err != nil {
					//log.Fatal("cannot retreive valid clients - error:", err)
					panic(err)
				}

				//prepare valid client set for share correction
				for _, vc := range valid_clients {
					notComplain, err := s.store.GetNoComplain(exp.Exp_ID, vc.Client_ID)
					if err != nil {
						//log.Fatal("cannot retreive complaint records where complaint is false - error:", err)
						panic(err)
					}

					if len(notComplain) < s.cfg.N {
						//build input shares map for each client
						inputMaskedShares := make(map[int]map[string][]rss.Share)
						for _, entry := range notComplain {
							masked_shares, _ := s.store.GetMaskedShares(exp.Exp_ID, entry.Server_ID, entry.Client_ID)
							for _, record := range masked_shares {
								v1, check1 := inputMaskedShares[record.Input_Index]
								if check1 {
									v2, check2 := v1[record.Server_ID]
									if check2 {
										inputMaskedShares[record.Input_Index][record.Server_ID] = append(v2, rss.Share{Index: record.Index, Value: record.Value})
									} else {
										inputMaskedShares[record.Input_Index][record.Server_ID] = []rss.Share{{Index: record.Index, Value: record.Value}}
									}
								} else {
									inputMaskedShares[record.Input_Index] = make(map[string][]rss.Share)
									inputMaskedShares[record.Input_Index][record.Server_ID] = []rss.Share{{Index: record.Index, Value: record.Value}}
								}
							}

						}

						//remove invalid client from valid set
						isRemoved := false
						for _, val := range inputMaskedShares {
							parties := []rss.Party{}
							for _, party := range val {
								parties = append(parties, rss.Party{Index: 0, Shares: party})
							}
							nrss, _ := rss.NewReplicatedSecretSharing(s.cfg.N, s.cfg.T, s.cfg.Q)

							_, err := nrss.Reconstruct(parties)
							if err != nil {
								log.Printf("reconstruct fail, need to remove client from valid set - error:%s", err)
								err = s.store.DeleteValidClient(exp.Exp_ID, vc.Client_ID)
								isRemoved = true
								if err != nil {
									//log.Fatal("cannot remove client from valid set - error:", err)
									panic(err)
								}
								break
							}

						}

						if !isRemoved {
							//check if server itself complains this valid client
							record, err := s.store.GetComplaint(exp.Exp_ID, s.cfg.Server_ID, vc.Client_ID)
							if err != nil {
								//log.Fatal("cannot retreive complaint record - error:", err)
								panic(err)
							}

							//share correction
							if record.Exp_ID != "" && record.Complain {
								for input_index, shares := range inputMaskedShares {
									masked_shares, err := computeMajority(shares)
									if err != nil {
										panic(err)
									}

									for _, sh := range masked_shares {
										mask, err := s.store.GetMask(exp.Exp_ID, vc.Client_ID, input_index, sh.Index)
										if err != nil {
											//log.Fatal("cannot get mask from table - error:", err)
											panic(err)
										}

										newShare := sh.Value - mask.Value
										err = s.store.UpdateClientShare(exp.Exp_ID, vc.Client_ID, input_index, sh.Index, newShare)
										if err != nil {
											//log.Fatal("cannot update client share - error:", err)
											panic(err)
										}

									}

								}
							}
						}
					}

				}

				clientShares, err := s.store.GetValidClientShares(exp.Exp_ID)
				if err != nil {
					//log.Fatal("cannot retreive valid client shares record - error:", err)
					panic(err)
				}

				//compute aggregated share
				aggreShares, err := aggregateShares(clientShares)
				if err != nil {
					panic(err)
				}

				/**
				//test s6 change aggregated share to invalid value
				if s.cfg.Server_ID == "s6" {
					aggreShares = []rss.Share{{Index: 0, Value: 27597}, {Index: 2, Value: 28090}, {Index: 3, Value: 35626}, {Index: 4, Value: 36324}, {Index: 5, Value: 38150}}
				}**/

				msg := AggregatedShareRequest{Exp_ID: exp.Exp_ID, Server_ID: s.cfg.Server_ID, Shares: aggreShares, Timestamp: time.Now().Format("2006-01-02 15:04:05")}
				fmt.Printf("server %s sends: %+v\n", s.cfg.Server_ID, msg)
				writer := &msg
				send(exp.Owner, writer.ToJson())

				//set round3 to completed
				err = s.store.UpdateRound3Completed(exp.Exp_ID)
				if err != nil {
					//log.Println("cannot set round3 to completed - error:", err)
					panic(err)
				}
			}

		}

	}
	//TODO: check if experiment is over and do something (e.g., remove exp and clients information from DB)

}

func aggregateShares(clientShares []sqlstore.ClientShare) ([]rss.Share, error) {
	if len(clientShares) == 0 {
		return nil, fmt.Errorf("client shares are empty: no valid client exists")
	}

	var result []rss.Share
	aggreShare := make(map[int]int)
	for _, entry := range clientShares {
		val, exist := aggreShare[entry.Index]
		if exist {
			aggreShare[entry.Index] = val + entry.Value
		} else {
			aggreShare[entry.Index] = entry.Value
		}
	}

	for key, val := range aggreShare {
		result = append(result, rss.Share{Index: key, Value: val})
	}

	return result, nil
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
		log.Printf("impossible to send http request: %s", err)
	} else {
		log.Printf("response Status:%s", res.Status)

		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		if len(body) > 0 {
			fmt.Println("response Body:", string(body))
		}

	}

}

func (s *Server) Start() {

	http.HandleFunc("/client/", s.clientRequestHandler)
	//http.HandleFunc("/experiment/", s.expInforHandler)
	http.HandleFunc("/complaint/", s.serverComplaintHandler)
	http.HandleFunc("/maskedShare/", s.serverMaskedSharesHandler)

	log.Fatal(http.ListenAndServe(":"+s.cfg.Port, nil))

}

func (s *Server) StartTLS(certFile string, keyFile string) {

	http.HandleFunc("/client/", s.clientRequestHandler)
	//http.HandleFunc("/outputParty", s.expInforHandler)

	log.Fatal(http.ListenAndServeTLS(":"+s.cfg.Port, certFile, keyFile, nil))

}
