package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"example.com/SMC/pkg/packed"
)

type Configuration struct {
	N int
	T int
	K int
	Q int
}

type Experiment struct {
	EID       string                  `json:"EID"`
	Due       string                  `json:"Due"`
	SumShares map[string]packed.Share `json:"SumShares"` // key is SID
}

type OutputParty struct {
	Exps map[string]Experiment
}

func NewOutputParty() *OutputParty {
	return &OutputParty{Exps: make(map[string]Experiment)}
}

func (op *OutputParty) reveal(shares []packed.Share, n, t, k, q int) ([]int, error) {
	//Todo: read parameters for packed secret sharing from file
	npss, err := packed.NewPackedSecretSharing(n, t, k, q)

	result, err := npss.Reconstruct(shares)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)

	return result, nil
}

func (op *OutputParty) serverDataHandler(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	t := struct {
		EID       string       `json:"EID"`
		SID       string       `json:"SID"`
		SumShares packed.Share `json:"SumShares"`
		Timestamp string       `json:"Timestamp"`
	}{}
	err := decoder.Decode(&t)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	//Todo: check validity of server

	//store server data to database
	exp, exists := op.Exps[t.EID]
	if exists {
		_, exists := exp.SumShares[t.SID]
		if !exists {
			exp.SumShares[t.SID] = t.SumShares
		} else {
			log.Println("Server data is discarded because it already exists!")
		}
	} else {
		log.Println("Experiment does not exist!")
	}
}

func main() {
	port := flag.String("port", ":8080", "the port on which the server will listen")

	outputParty := NewOutputParty()

	// Todo: set experiments information from file
	outputParty.Exps["exp1"] = Experiment{EID: "exp1", Due: "2023-06-01", SumShares: make(map[string]packed.Share)}

	http.HandleFunc("/serverDataSubmit/", outputParty.serverDataHandler)

	// Todo: read port number from file
	log.Fatal(http.ListenAndServe(*port, nil))

	// Todo: check if receive data from all servers
	//outputParty.reveal()

}
