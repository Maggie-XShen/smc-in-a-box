package main

import (
	"encoding/json"
	"log"
	"os"

	"example.com/SMC/pkg/ligero"
)

type ClientRequest struct {
	Exp_ID    string       `json:"Exp_ID"`
	Client_ID string       `json:"Client_ID"`
	Token     string       `json:"Token"`
	Proof     ligero.Proof `json:"Proof"`
	Timestamp string       `json:"Timestamp"`
	//Hash_proof  string       `json:"HashProof"`
	//Signature   string       `json:"Signature"`
}

type Input struct {
	Exp_ID  string `json:"Exp_ID"`
	Secrets []int  `json:"Secrets"`
}

func (c *ClientRequest) ToJson() []byte {
	msg := &ClientRequest{
		Exp_ID:    c.Exp_ID,
		Client_ID: c.Client_ID,
		Proof:     c.Proof,
		Timestamp: c.Timestamp,
	}
	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("Cannot marshall client request: %s", err)
	}

	log.Printf("client %s is sending data of %s to server%d ...\n", msg.Client_ID, msg.Exp_ID, msg.Proof.PartyShares[0].Index)

	return message
}

func ReadClientInput(path string) []Input {
	jsonData, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("%s", err)
		return nil
	}

	var items []Input
	err = json.Unmarshal(jsonData, &items)
	if err != nil {
		log.Fatalf("%s", err)
		return nil
	}
	return items

}
