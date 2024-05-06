package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/SMC/outputparty/config"
	"example.com/SMC/outputparty/sqlstore"
	"example.com/SMC/pkg/rss"
	"github.com/sirupsen/logrus"
)

type OutputParty struct {
	cfg   *config.OutputParty
	store *sqlstore.DB
}

func NewOutputParty(conf *config.OutputParty) *OutputParty {
	return &OutputParty{cfg: conf, store: sqlstore.NewDB(conf.OutputParty_ID)}
}

func (op *OutputParty) HandelExp(path string) {
	experiments := ReadOutputPartyInput(path)

	expService := NewExperimentService(op.store)
	for _, exp := range experiments {
		logger.WithFields(logrus.Fields{
			"exp_id":           exp.Exp_ID,
			"client_share_due": exp.ClientShareDue,
			"server_share_due": exp.ServerShareDue,
		}).Info("")

		err := expService.CreateExperiment(exp)
		if err != nil {
			panic(err)
		}
	}

}

/**
func (op *OutputParty) send(address string, data []byte) {
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

}**/

func (op *OutputParty) reveal(parties []rss.Party) (int, error) {
	nrss, err := rss.NewReplicatedSecretSharing(op.cfg.N, op.cfg.T, op.cfg.Q)
	if err != nil {
		panic(err)
	}

	results, err := nrss.Reconstruct(parties)
	if err != nil {
		panic(err)
	}

	return results, nil
}

func (op *OutputParty) WaitForEndOfExperiment(ticker *time.Ticker) {
	for range ticker.C {
		experiments, err := op.store.GetAllExperiments()
		if err != nil {
			//log.Println("cannot retreive non-completed experiments - error:", err)
			panic(err)
		}

		for _, exp := range experiments {

			due, _ := time.Parse("2006-01-02 15:04:05.999999999 +0000 UTC", exp.ServerShareDue)
			currentTime := time.Now().UTC()

			if currentTime.After(due) {
				list, err := op.store.GetSharesPerExperiment(exp.Exp_ID)
				if err != nil {
					//log.Println("cannot retrieve servers records - error:", err)
					//continue
					panic(err)
				}

				inputShares := make(map[int]map[string][]rss.Share)
				for _, record := range list {
					v1, check1 := inputShares[record.Input_Index]
					if check1 {
						v2, check2 := v1[record.Server_ID]
						if check2 {
							inputShares[record.Input_Index][record.Server_ID] = append(v2, rss.Share{Index: record.Index, Value: record.Value})
						} else {
							inputShares[record.Input_Index][record.Server_ID] = []rss.Share{{Index: record.Index, Value: record.Value}}
						}
					} else {
						inputShares[record.Input_Index] = make(map[string][]rss.Share)
						inputShares[record.Input_Index][record.Server_ID] = []rss.Share{{Index: record.Index, Value: record.Value}}
					}
				}

				// reconstruct sum of secrets
				nrss, err := rss.NewReplicatedSecretSharing(op.cfg.N, op.cfg.T, op.cfg.Q)
				if err != nil {
					panic(err)
				}

				result := make([]int, op.cfg.N_secrets)
				for input_index, list := range inputShares {
					size := len(list)
					servers := make([]rss.Party, size)
					i := 0
					for _, shares := range list {
						servers[i] = rss.Party{Index: 0, Shares: shares}
						i++
					}
					sum, err := nrss.Reconstruct(servers)
					if err != nil {
						panic(err)
					}

					result[input_index] = sum

				}

				reconstruction_start, _ := time.Parse("2006-01-02 15:04:05.999999999 +0000 UTC", exp.ServerShareDue)
				reconstruction_end = time.Since(reconstruction_start)

				logger.WithFields(logrus.Fields{
					"exp_id": exp.Exp_ID,
					"result": result,
				}).Info("")

				//fmt.Printf("sum of secrets for %s : %v\n", exp.Exp_ID, result)

				WriteResult(exp.Exp_ID, result)

				err = op.store.UpdateCompletedExperiment(exp.Exp_ID) //set experiments to completed
				if err != nil {
					//log.Println("cannot set experiment to completed - error:", err1)
					panic(err)
				}

			}

		}

		// Todo: check if experiment is over and do something (e.g., remove exp and clients information from DB)
	}
}

func (op *OutputParty) serverRequestHandler(rw http.ResponseWriter, req *http.Request) {
	var request AggregatedShareRequest

	serverService := NewServerService(op.store)
	data := request.ReadJson(req)
	err := serverService.CreateServerShare(data)

	if err != nil {
		log.Printf("error: %s\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(rw, err)
		return
	}

	count := op.store.CountSharesPerExperiment(data.Exp_ID)

	if count == int64(n_sh) {
		real_server_share_due = time.Now().UTC() //ideal server share due is when all server shares arrived at output party

	}

	rw.WriteHeader(http.StatusOK)

}

func (op *OutputParty) Start() {
	http.HandleFunc("/serverShare/", op.serverRequestHandler)

	log.Fatal(http.ListenAndServe(":"+op.cfg.Port, nil))

}

func (op *OutputParty) StartTLS(certFile string, keyFile string) {
	http.HandleFunc("/serverShare/", op.serverRequestHandler)

	log.Fatal(http.ListenAndServeTLS(":"+op.cfg.Port, certFile, keyFile, nil))

}

func (op *OutputParty) Close(ticker *time.Ticker) {
	for range ticker.C {

		all, err := op.store.GetAllExperiments()
		if err != nil {
			//log.Println("cannot retreive non-completed experiments - error:", err)
			panic(err)
		}

		if len(all) == 0 {
			end := time.Now().UTC()
			logger.WithFields(logrus.Fields{
				"real_server_share_due": real_server_share_due.String(),
				"reconstruction_time":   reconstruction_end.String(), //time from output party started to reconstruction of the experiment is done
				"end":                   end.String(),
			}).Info("")
			log.Printf("%s is finishing\n", op.cfg.OutputParty_ID)
			os.Exit(0)
		}
	}

}
