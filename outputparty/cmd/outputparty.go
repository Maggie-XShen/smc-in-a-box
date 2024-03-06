package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"example.com/SMC/outputparty/config"
	"example.com/SMC/outputparty/sqlstore"
	"example.com/SMC/pkg/rss"
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
	//TODO: need to remove
	//time.Sleep(2 * time.Minute)
	for range ticker.C {
		experiments, err := op.store.GetAllExperiments()
		if err != nil {
			//log.Println("cannot retreive non-completed experiments - error:", err)
			panic(err)
		}

		for _, exp := range experiments {

			due, _ := time.Parse("2006-01-02 15:04:05", exp.ServerShareDue)
			currentTime := time.Now().UTC()

			if currentTime.After(due) {

				records, err := op.store.GetSharesPerExperiment(exp.Exp_ID)
				if err != nil {
					//log.Println("cannot retrieve servers records - error:", err)
					//continue
					panic(err)
				}

				if len(records) > 0 {
					serverShare := make(map[string][]rss.Share)
					for _, r := range records {
						v1, check1 := serverShare[r.Server_ID]
						if check1 {
							serverShare[r.Server_ID] = append(v1, rss.Share{Index: r.Index, Value: r.Value})
						} else {
							serverShare[r.Server_ID] = []rss.Share{{Index: r.Index, Value: r.Value}}
						}
					}

					// reconstruct sum of secrets
					var parties []rss.Party
					for _, server := range serverShare {
						parties = append(parties, rss.Party{Index: 0, Shares: server})
					}

					result, err := op.reveal(parties)
					if err != nil {
						panic(err)
					}

					fmt.Printf("sum of secrets for %s : %v\n", exp.Exp_ID, result)

					WriteResult(exp.Exp_ID, result)
				} else {
					log.Println("cannot compute the result since servers' shares are missing")
				}

				//set experiments to completed
				err1 := op.store.UpdateCompletedExperiment(exp.Exp_ID)
				if err1 != nil {
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
	err := serverService.CreateServerShare(request.ReadJson(req))

	if err != nil {
		log.Printf("error: %s\n", err)
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(rw, err)
		return
	}

	rw.WriteHeader(http.StatusOK)

}

func (op *OutputParty) Start() {
	http.HandleFunc("/serverShare/", op.serverRequestHandler)

	log.Fatal(http.ListenAndServe(":"+op.cfg.Port, nil))

}

/**
func (op *OutputParty) Start(certFile string, keyFile string) {
	http.HandleFunc("/serverRequestSubmit/", op.serverRequestHandler)

	log.Fatal(http.ListenAndServeTLS(":"+op.cfg.Port, certFile, keyFile, nil))

}**/
