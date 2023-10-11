package main

import (
	"encoding/json"
	"log"
	"os"

	"example.com/SMC/pkg/packed"
)

type ClientRequest struct {
	Exp_ID       string       `json:"Exp_ID"`
	Client_ID    string       `json:"Client_ID"`
	Token        string       `json:"Token"`
	Secret_Share packed.Share `json:"Secret_Share"`
	Timestamp    string       `json:"Timestamp"`
	//Proof       string       `json:"Proof"`
	//Hash_proof  string       `json:"HashProof"`
	//Signature   string       `json:"Signature"`
}

type Input struct {
	Exp_ID  string `json:"Exp_ID"`
	Secrets []int  `json:"Secrets"`
}

func (c *ClientRequest) ToJson() []byte {
	msg := &ClientRequest{
		Exp_ID:       c.Exp_ID,
		Client_ID:    c.Client_ID,
		Secret_Share: c.Secret_Share,
		Timestamp:    c.Timestamp,
	}
	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("Cannot marshall client request: %s", err)
	}

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
