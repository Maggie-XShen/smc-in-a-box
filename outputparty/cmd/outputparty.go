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

	"example.com/SMC/outputparty/config"
	"example.com/SMC/outputparty/public/utils"
	"example.com/SMC/outputparty/sqlstore"
	"example.com/SMC/pkg/packed"
)

type OutputParty struct {
	cfg   *config.OutputParty
	store *sqlstore.SqlStore
}

func NewOutputParty(conf *config.OutputParty) *OutputParty {
	return &OutputParty{cfg: conf, store: sqlstore.New(conf.OutputParty_ID)}
}

func (op *OutputParty) HandelExp(path string) {

	// TODO: decide how to get experiment information
	type Tables struct {
		Experiments []utils.OutputPartyRequest
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

	expService := NewExperimentService(op.store)
	for _, exp := range tables.Experiments {
		exp.Due = time.Now().Add(time.Hour * 10).Format("2006-01-02 15:04:05")
		err := expService.CreateExp(exp)
		if err != nil {
			log.Println("error:", err)
		}

		/**
		msg := utils.OutputPartyRequest{Exp_ID: exp.Exp_ID, Due: exp.Due}
		fmt.Printf("%+v\n", msg)
		writer := &msg
		for _, url := range op.cfg.URLs {
			op.send(url, writer.WriteJson())
		}**/
	}

}

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
		experiments, err := op.store.GetAllExps()
		if err != nil {
			log.Println("cannot retreive non-completed experiments - error:", err)
		}

		for _, exp := range experiments {

			due, _ := time.Parse("2006-01-02 15:04:05", exp.Due)
			//currentTime := time.Now()
			currentTime := due.Add(5 * time.Minute) //Todo: need to change back to time.Now()

			// check current time is pased due
			if currentTime.After(due) {
				servers, err := op.store.GetAllServers(exp.Exp_ID)
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
				//TODO: write result to file
				utils.WriteResult(exp.Exp_ID, result)

				//set experiments to completed
				err1 := op.store.UpdateCompletedExperiment(exp.Exp_ID)
				if err1 != nil {
					log.Println("cannot set experiment to completed - error:", err1)
				}

			}

		}

		// Todo: check if experiment is over and do something (e.g., remove exp and clients information from DB)
	}
}

func (op *OutputParty) serverRequestHandler(rw http.ResponseWriter, req *http.Request) {
	var request utils.ServerRequest

	serverService := NewServerService(op.store)
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

/**
func (op *OutputParty) Start(certFile string, keyFile string) {
	http.HandleFunc("/serverRequestSubmit/", op.serverRequestHandler)

	log.Fatal(http.ListenAndServeTLS(":"+op.cfg.Port, certFile, keyFile, nil))

}**/
