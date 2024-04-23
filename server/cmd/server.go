package main

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"example.com/SMC/pkg/rss"
	"example.com/SMC/server/config"
	"example.com/SMC/server/sqlstore"
	"github.com/sirupsen/logrus"
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
		logger.WithFields(logrus.Fields{
			"exp_id":              exp.Exp_ID,
			"client_share_due":    exp.ClientShareDue,
			"complaint_due":       exp.ComplaintDue,
			"share_broadcast_due": exp.ShareBroadcastDue,
			"owner":               exp.Owner,
		}).Info("")

		err := expService.CreateExperiment(exp)
		if err != nil {
			log.Printf("%s cannot creat experiment - error: %s\n", s.cfg.Server_ID, err)
		}
	}

}

func (s *Server) clientRequestHandler(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)

	var request ClientRequest

	clientService := NewClientService(s.store)
	data := request.ReadJson(req)

	go func() {

		err := clientService.CreateClientShare(data, s.cfg)

		if err != nil {
			log.Printf("%s cannot create client share - error: %s\n", s.cfg.Server_ID, err)
		}

		client_count += 1
		if client_count == client_size {
			real_client_share_due = time.Now().UTC() // time to start the step of assemble complaints and broadcast without waiting
		}
	}()

}

func (s *Server) serverComplaintHandler(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)

	var request ComplaintRequest

	serverService := NewServerService(s.store)
	data := request.ReadJson(req)

	go func() {

		err := serverService.CreateComplaint(data)

		count := s.store.CountComplaintsPerExperiment(data.Exp_ID)

		if count == int64(complaint_size) {
			real_complaint_due = time.Now().UTC() //time to start the step of masked share generation without waiting
		}

		if err != nil {
			log.Printf("error: %s\n", err)
		}
	}()

}

func (s *Server) serverMaskedSharesHandler(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)

	var request MaskedShareRequest

	serverService := NewServerService(s.store)
	data := request.ReadJson(req)

	go func() {

		err := serverService.CreateMaskedShares(data)

		count := s.store.CountMaskedSharesPerExperiment(data.Exp_ID)

		if mask_share_size != 0 && count == int64(mask_share_size) {
			real_share_broadcast_due = time.Now().UTC() //time to start the step of share correction without waiting

		}

		if err != nil {
			log.Printf("error: %s\n", err)
		}
	}()

}

func (s *Server) dolevComplaintHandler(rw http.ResponseWriter, req *http.Request) {
	var request DolevComplaintRequest
	data := request.ReadJson(req)
	if data.Round_ID <= s.cfg.T+1 {
		if data.Round_ID == len(data.Signatures) {
			complaints := fmt.Sprintf("%+v", data.Msg.Complaints)
			set, _ := s.store.GetEchoComplaint(data.Msg.Server_ID, data.Msg.Exp_ID, complaints)
			//check if message already exist
			str := fmt.Sprintf("%+v", data.Msg)
			if len(set) == 0 && s.checkSigChain(str, data.Signatures) {
				s.store.InsertEchoComplaint(data.Msg.Server_ID, data.Msg.Exp_ID, complaints)
				s.dolevComplaintBroadcast(data.Round_ID+1, data.Msg, data.Signatures)
			}
		}

	}

}

func (s *Server) checkSigChain(msg string, sig_chain []Signature) bool {
	hashed := sha256.Sum256([]byte(fmt.Sprintf("%+v", msg)))
	for _, sig := range sig_chain {
		fileName := fmt.Sprintf("cert_%s.pem", sig.Server_ID)
		cert_path := filepath.Join("./rsa/", fileName)
		err := rsa.VerifyPKCS1v15(loadPublicKey(cert_path), crypto.SHA256, hashed[:], sig.Sig)
		if err != nil {
			log.Printf("Signature verification failed: %s\n", err)
			return false
		}
	}
	return true
}

func (s *Server) dolevMaskedSharesHandler(rw http.ResponseWriter, req *http.Request) {
	var request DolevMaskedShareRequest
	data := request.ReadJson(req)
	if data.Round_ID <= s.cfg.T+1 {
		if data.Round_ID == len(data.Signatures) {
			mask_shares := fmt.Sprintf("%+v", data.Msg.MaskedShares)
			set, _ := s.store.GetEchoMaskedShare(data.Msg.Server_ID, data.Msg.Exp_ID, mask_shares)
			//check if message already exist
			str := fmt.Sprintf("%+v", data.Msg)
			if len(set) == 0 && s.checkSigChain(str, data.Signatures) {
				s.store.InsertEchoMaskedShare(data.Msg.Server_ID, data.Msg.Exp_ID, mask_shares)
				s.dolevMaskedShareBroadcast(data.Round_ID+1, data.Msg, data.Signatures)
			}
		}

	}
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

func computeMajority(input map[string][]rss.Share, t int) ([]rss.Share, error) {
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
		val, err := FindMajority(values, t)
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
			log.Printf("%s cannot retreive non-completed experiments - error: %s\n", s.cfg.Server_ID, err)
			continue
		}

		for _, exp := range experiments {

			due, _ := time.Parse("2006-01-02 15:04:05.999999999 +0000 UTC", exp.ClientShareDue)
			currentTime := time.Now().UTC()

			if currentTime.After(due) {

				complaints, err := s.store.GetComplaintsPerServer(exp.Exp_ID, s.cfg.Server_ID)
				if err != nil {
					log.Printf("%s cannot retreive complaints records - error: %s\n", s.cfg.Server_ID, err)
					continue
				}

				if len(complaints) == 0 {
					log.Printf("client share due passed, %s complaint table is empty\n", s.cfg.Server_ID)
					err = s.store.UpdateRound1Completed(exp.Exp_ID) //set round1 to completed
					if err != nil {
						log.Printf(" %s cannot set round1 to completed\n", s.cfg.Server_ID)
						panic(err)
					}
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

				var wg sync.WaitGroup
				for _, address := range s.cfg.Complaint_urls {
					wg.Add(1)
					go func(addr string) {
						defer wg.Done()
						log.Printf("server %s is sending complaints to %s\n", s.cfg.Server_ID, addr)
						writer := &message
						send(addr, writer.ToJson())
					}(address)
				}

				err = s.store.UpdateRound1Completed(exp.Exp_ID) //set round1 to completed
				if err != nil {
					log.Printf("error: %s cannot set round1 to completed\n", s.cfg.Server_ID)
					panic(err)
				}

				//s.dolevComplaintBroadcast(1, message, []Signature{})

			}

		}
	}
}

// When complaint clock is triggered, server computes valid client set, and invoke vss and broadcast masked shares
func (s *Server) WaitForEndOfComplaintBroadcast(ticker *time.Ticker) {

	for range ticker.C {

		experiments, err := s.store.GetExpsWithRound1Completed()
		if err != nil {
			log.Printf("%s cannot retreive non-completed experiments - error: %s\n", s.cfg.Server_ID, err)
			continue
		}

		for _, exp := range experiments {

			due, _ := time.Parse("2006-01-02 15:04:05.999999999 +0000 UTC", exp.ComplaintDue)
			currentTime := time.Now().UTC()

			if currentTime.After(due) {
				mask_share_start := time.Now() //masked share generation start time

				//find dropout clients for the server
				dropout, err := s.store.GetDropoutClient(exp.Exp_ID)
				if err != nil {
					log.Printf("%s cannot retreive missing clients- error: %s\n", s.cfg.Server_ID, err)
					continue

				}

				//generate complaint of dropout client
				for _, client_id := range dropout {
					err := s.store.InsertClient(exp.Exp_ID, client_id)
					if err != nil {
						log.Printf("%s cannot insert missing client to the client table\n", s.cfg.Server_ID)
						panic(err)
					}

					err = s.store.InsertComplaint(exp.Exp_ID, s.cfg.Server_ID, client_id, true, []byte("default"))
					if err != nil {
						log.Printf("%s cannot insert complaint of missing client to the complaint table\n", s.cfg.Server_ID)
						panic(err)
					}
				}

				clients, err := s.store.GetClientsPerExperiment(exp.Exp_ID)
				if err != nil {
					log.Printf("%s cannot retreive clients records\n", s.cfg.Server_ID)
					panic(err)
				}

				//generate valid client set and trigger mask generateion when condition meets
				for _, c := range clients {
					complaints, err := s.store.GetComplaintsPerClient(exp.Exp_ID, c.Client_ID)
					if err != nil {
						log.Printf("%s cannot retreive complaints records\n", s.cfg.Server_ID)
						panic(err)
					}

					num_isNotComplain, rootCount, maxCount := count(complaints)

					if num_isNotComplain >= s.cfg.N-s.cfg.T && maxCount >= s.cfg.N-s.cfg.T {
						err = s.store.InsertValidClient(exp.Exp_ID, c.Client_ID)
						if err != nil {
							log.Printf("%s cannot create valid client record\n", s.cfg.Server_ID)
							panic(err)
						}

						//generate mask and masked shares
						if num_isNotComplain < s.cfg.N || rootCount > 1 {
							clientShares, err := s.store.GetClientShares(exp.Exp_ID, c.Client_ID)
							if err != nil {
								log.Printf("%s cannot get client shares record\n", s.cfg.Server_ID)
								panic(err)
							}

							for _, cs := range clientShares {
								key := 1 //TODO: change key to offline generated key
								crs := NewCryptoRandSource()
								crs.Seed(key, cs.Exp_ID, cs.Client_ID, cs.Input_Index, cs.Index)
								mask := int(crs.Int63(int64(s.cfg.Q)))

								err := s.store.InsertMask(cs.Exp_ID, cs.Client_ID, cs.Input_Index, cs.Index, mask)
								if err != nil {
									log.Printf("%s cannot add mask record to the table\n", s.cfg.Server_ID)
									panic(err)
								}

								err = s.store.InsertMaskedShare(cs.Exp_ID, s.cfg.Server_ID, cs.Client_ID, cs.Input_Index, cs.Index, mask+cs.Value)
								if err != nil {
									log.Printf("%s cannot add masked share to the table\n", s.cfg.Server_ID)
									panic(err)
								}

							}

						}

					}

				}

				mask_share_end = time.Since(mask_share_start) //masked share generation computing time

				maskedShares, err := s.store.GetMaskedSharesPerServer(exp.Exp_ID, s.cfg.Server_ID)
				if err != nil {
					log.Printf("%s cannot retreive masked shares record\n", s.cfg.Server_ID)
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

					var wg sync.WaitGroup
					for _, address := range s.cfg.Masked_share_urls {
						wg.Add(1)
						go func(addr string) {
							defer wg.Done()
							log.Printf("server %s is sending masked shares to %s\n", s.cfg.Server_ID, addr)
							writer := &message
							send(addr, writer.ToJson())
						}(address)
					}

					//TODO:dolev strong broadcast
					//s.dolevMaskedShareBroadcast(1, message, []Signature{})

				}

				err = s.store.UpdateRound2Completed(exp.Exp_ID) //set round2 to completed
				if err != nil {
					log.Printf("%s cannot set round2 to completed\n", s.cfg.Server_ID)
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
			log.Printf("%s cannot retreive non-completed experiments - error: %s\n", s.cfg.Server_ID, err)
			continue
		}

		for _, exp := range experiments {

			due, _ := time.Parse("2006-01-02 15:04:05.999999999 +0000 UTC", exp.ShareBroadcastDue)
			currentTime := time.Now().UTC()

			if currentTime.After(due) {

				share_correct_start := time.Now() //share correction start time

				valid_clients, err := s.store.GetValidClientsPerExperiment(exp.Exp_ID)
				if err != nil {
					log.Printf("%s cannot retreive valid clients - error: %s\n", s.cfg.Server_ID, err)
					continue
				}

				//prepare valid client set for share correction
				for _, vc := range valid_clients {
					notComplain, err := s.store.GetNoComplain(exp.Exp_ID, vc.Client_ID)
					if err != nil {
						log.Printf("%s cannot retreive complaint records where complaint is false\n", s.cfg.Server_ID)
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
						for _, list := range inputMaskedShares {
							parties := []rss.Party{}
							for _, shares := range list {
								parties = append(parties, rss.Party{Index: 0, Shares: shares})
							}
							nrss, _ := rss.NewReplicatedSecretSharing(s.cfg.N, s.cfg.T, s.cfg.Q)

							_, err := nrss.Reconstruct(parties)
							if err != nil {
								log.Printf("%s reconstruct fail, need to remove client from valid set - err: %s\n", s.cfg.Server_ID, err)
								err = s.store.DeleteValidClient(exp.Exp_ID, vc.Client_ID)
								isRemoved = true
								if err != nil {
									log.Printf("%s cannot remove client from valid set\n", s.cfg.Server_ID)
									panic(err)
								}
								break
							}

						}

						if !isRemoved {
							//check if server itself complains this valid client
							record, err := s.store.GetComplaint(exp.Exp_ID, s.cfg.Server_ID, vc.Client_ID)
							if err != nil {
								log.Printf("%s cannot retreive complaint record\n", s.cfg.Server_ID)
								panic(err)
							}

							//share correction
							if record.Exp_ID != "" && record.Complain {
								for input_index, shares := range inputMaskedShares {
									masked_shares, err := computeMajority(shares, s.cfg.T)
									if err != nil {
										panic(err)
									}

									for _, sh := range masked_shares {
										mask, err := s.store.GetMask(exp.Exp_ID, vc.Client_ID, input_index, sh.Index)
										if err != nil {
											log.Printf("%s cannot get mask from table\n", s.cfg.Server_ID)
											panic(err)
										}

										newShare := sh.Value - mask.Value
										err = s.store.UpdateClientShare(exp.Exp_ID, vc.Client_ID, input_index, sh.Index, newShare)
										if err != nil {
											log.Printf("%s cannot update client share\n", s.cfg.Server_ID)
											panic(err)
										}

									}

								}
							}
						}
					}

				}

				share_correct_end = time.Since(share_correct_start) //share correction computing time

				clientShares, err := s.store.GetValidClientShares(exp.Exp_ID)
				if err != nil {
					log.Printf("%s cannot retreive valid client shares record\n", s.cfg.Server_ID)
					panic(err)
				}

				//compute aggregated share
				aggreShares, err := s.aggregateShares(clientShares)
				if err != nil {
					panic(err)
				}

				/**
				//test s6 change aggregated share to invalid value
				if s.cfg.Server_ID == "s6" {
					aggreShares = []rss.Share{{Index: 0, Value: 27597}, {Index: 2, Value: 28090}, {Index: 3, Value: 35626}, {Index: 4, Value: 36324}, {Index: 5, Value: 38150}}
				}**/

				msg := AggregatedShareRequest{Exp_ID: exp.Exp_ID, Server_ID: s.cfg.Server_ID, Shares: aggreShares, Timestamp: time.Now().UTC().Format("2006-01-02 15:04:05")}
				log.Printf("server %s is sending aggregated shares: %+v\n", s.cfg.Server_ID, msg)
				writer := &msg
				send(exp.Owner, writer.ToJson())

				//set round3 to completed
				err = s.store.UpdateRound3Completed(exp.Exp_ID)
				if err != nil {
					log.Printf("%s cannot set round3 to completed\n", s.cfg.Server_ID)
					panic(err)
				}
			}

		}

	}
	//TODO: check if experiment is over and do something (e.g., remove exp and clients information from DB)

}

func (s *Server) aggregateShares(clientShares []sqlstore.ClientShare) ([]rss.Party, error) {
	if len(clientShares) == 0 {
		return nil, fmt.Errorf("client shares are empty: no valid client exists")
	}

	aggreShare := make(map[int]map[int]int) //(intput_index,(share_index, aggregateed_sh))
	for _, record := range clientShares {
		v1, check1 := aggreShare[record.Input_Index]
		if check1 {
			v2, check2 := v1[record.Index]
			if check2 {
				aggreShare[record.Input_Index][record.Index] = v2 + record.Value
			} else {
				aggreShare[record.Input_Index][record.Index] = record.Value
			}
		} else {
			aggreShare[record.Input_Index] = make(map[int]int)
			aggreShare[record.Input_Index][record.Index] = record.Value
		}
	}

	result := make([]rss.Party, s.cfg.N_secrets)
	for i := 0; i < len(aggreShare); i++ {
		var shares []rss.Share
		for index, value := range aggreShare[i] {
			shares = append(shares, rss.Share{Index: index, Value: value})
		}
		result[i] = rss.Party{Index: 0, Shares: shares}
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
		log.Printf("impossible to send http request: %s\n", err)
	} else {
		log.Printf("response Status:%s\n", res.Status)

		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		if len(body) > 0 {
			fmt.Println("response Body:", string(body))
		}

	}

}

func (s *Server) dolevComplaintBroadcast(round int, msg ComplaintRequest, sig_chain []Signature) {
	// Hash the message using SHA-256
	hashed := sha256.Sum256([]byte(fmt.Sprintf("%+v", msg)))

	// Sign the hashed message using RSA private key
	fileName := fmt.Sprintf("priv_%s.pem", s.cfg.Server_ID)
	priv_path := filepath.Join("./rsa/", fileName)
	sig, err := rsa.SignPKCS1v15(rand.Reader, s.loadPrivateKey(priv_path), crypto.SHA256, hashed[:])
	if err != nil {
		panic(err)
	}

	sig_chain = append(sig_chain, Signature{Sig: sig, Server_ID: s.cfg.Server_ID})

	ds_message := DolevComplaintRequest{
		Round_ID:   round,
		Server_ID:  s.cfg.Server_ID,
		Msg:        msg,
		Signatures: sig_chain,
	}

	for _, address := range s.cfg.Dolev_complaint_urls {
		log.Printf("server %s Dolev-Strong broadcasts complaints: %+v\n", s.cfg.Server_ID, msg)
		writer := &ds_message
		send(address, writer.ToJson())
	}
}

func (s *Server) dolevMaskedShareBroadcast(round int, msg MaskedShareRequest, sig_chain []Signature) {
	// Hash the message using SHA-256
	hashed := sha256.Sum256([]byte(fmt.Sprintf("%+v", msg)))

	// Sign the hashed message using RSA private key
	fileName := fmt.Sprintf("priv_+%s.pem", s.cfg.Server_ID)
	priv_path := filepath.Join("./rsa/", fileName)
	sig, err := rsa.SignPKCS1v15(rand.Reader, s.loadPrivateKey(priv_path), crypto.SHA256, hashed[:])
	if err != nil {
		panic(err)
	}

	sig_chain = append(sig_chain, Signature{Sig: sig, Server_ID: s.cfg.Server_ID})

	ds_message := DolevMaskedShareRequest{
		Round_ID:   round,
		Server_ID:  s.cfg.Server_ID,
		Msg:        msg,
		Signatures: sig_chain,
	}

	for _, address := range s.cfg.Dolev_masked_share_urls {
		log.Printf("server %s Dolev-Strong broadcast: %+v\n", s.cfg.Server_ID, msg)
		writer := &ds_message
		send(address, writer.ToJson())
	}
}

func loadPublicKey(Cert_path string) *rsa.PublicKey {
	// Read the fullchain.pem file
	certPEMBlock, err := os.ReadFile(Cert_path)
	if err != nil {
		log.Fatal(err)
	}

	// Decode PEM encoded data
	pemBlock, _ := pem.Decode(certPEMBlock)
	if pemBlock == nil {
		log.Fatal("Failed to parse certificate PEM")
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	return cert.PublicKey.(*rsa.PublicKey)
}

func (s *Server) loadPrivateKey(path string) *rsa.PrivateKey {
	// Read the privkey.pem file
	privKeyPEMBlock, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	// Decode PEM encoded data
	pemBlock, _ := pem.Decode(privKeyPEMBlock)
	if pemBlock == nil {
		log.Fatal("Failed to parse private key PEM")
	}

	// Parse the private key
	privateKey, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	return privateKey.(*rsa.PrivateKey)
}

func (s *Server) Start() {

	http.HandleFunc("/client/", s.clientRequestHandler)
	http.HandleFunc("/complaint/", s.serverComplaintHandler)
	http.HandleFunc("/maskedShare/", s.serverMaskedSharesHandler)
	http.HandleFunc("/dolevComplaint/", s.dolevComplaintHandler)
	http.HandleFunc("/dolevMaskedShare/", s.dolevMaskedSharesHandler)

	log.Fatal(http.ListenAndServe(":"+s.cfg.Port, nil))

}

func (s *Server) StartTLS() {

	http.HandleFunc("/client/", s.clientRequestHandler)
	http.HandleFunc("/complaint/", s.serverComplaintHandler)
	http.HandleFunc("/maskedShare/", s.serverMaskedSharesHandler)
	http.HandleFunc("/dolevComplaint/", s.dolevComplaintHandler)
	http.HandleFunc("/dolevMaskedShare/", s.dolevMaskedSharesHandler)

	log.Fatal(http.ListenAndServeTLS(":"+s.cfg.Port, s.cfg.Cert_path, s.cfg.Key_path, nil))

}

func (s *Server) Close(ticker *time.Ticker) {
	for range ticker.C {
		finished, err := s.store.GetExpsWithRound3Completed()
		if err != nil {
			log.Printf("%s cannot retreive completed experiments - error: %s\n", s.cfg.Server_ID, err)
			continue
		}

		all, err := s.store.GetExperimentCount()
		if err != nil {
			log.Printf("%scannot retreive non-completed experiments - error: %s\n", s.cfg.Server_ID, err)
			continue
		}

		if int64(len(finished)) == all {
			end := time.Since(start)

			avg := float64(total_verify_time.Milliseconds()) / float64(client_size)

			avg_verify_time := time.Duration(avg) * time.Millisecond

			logger.WithFields(logrus.Fields{
				"real_client_share_due":    real_client_share_due.String(),
				"avg_verify_time":          avg_verify_time.String(),
				"total_verify_time":        total_verify_time.String(),
				"num_client_received":      client_count,
				"real_complaint_due":       real_complaint_due.String(),
				"mask_share_time":          mask_share_end.String(),
				"real_share_broadcast_due": real_share_broadcast_due.String(),
				"share_correction_time":    share_correct_end.String(),
				"end":                      end.String(),
			}).Info("")
			log.Printf("%s is finishing\n", s.cfg.Server_ID)
			os.Exit(0)
		}
	}

}
