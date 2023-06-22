package outputparty

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

type OutputParty struct {
	cfg     Config
	storage repository.Storage
}

func NewOutputParty(conf Config, storage repository.Storage) *OutputParty {
	return &OutputParty{cfg: conf, storage: storage}
}

func (op *OutputParty) HandelExpInfor() {

	// TODO: decide how to get experiment information
	due := time.Now().Add(time.Hour * 10).Format("2006-01-02 15:04:05")

	var exps []message.OutputPartyRequest
	exps = append(exps, message.OutputPartyRequest{Exp_ID: "exp1", Due: due})
	exps = append(exps, message.OutputPartyRequest{Exp_ID: "exp2", Due: due})

	expService := api.NewExperimentService(op.storage)
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

	fmt.Printf("%v", result)

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

			currentTime := time.Now()
			due, _ := time.Parse("2006-01-02 15:04:05", exp.Due)

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
					continue
				}

				// reconstruct sum of secrets
				var shares []packed.Share
				for _, server := range servers {
					shares = append(shares, packed.Share{Value: server.SumShare_Value, Index: server.SumShare_Index})
				}

				fmt.Printf("%+v\n", shares)

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

	serverService := api.NewServerService(op.storage)
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
