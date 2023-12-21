package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"example.com/SMC/pkg/ligero"
	"example.com/SMC/pkg/packed"
)

type ClientRequest struct {
	Exp_ID       string       `json:"Exp_ID"`
	Client_ID    string       `json:"Client_ID"`
	Token        string       `json:"Token"`
	Secret_Share packed.Share `json:"Secret_Share"`
	Proof        ligero.Proof `json:"Proof"`
	Timestamp    string       `json:"Timestamp"`
	//Hash_proof  string       `json:"HashProof"`
	//Signature   string       `json:"Signature"`
}

type Input struct {
	Exp_ID  string `json:"Exp_ID"`
	Secrets []int  `json:"Secrets"`
}

func FormClaims(input []int, shares []packed.Share) ([]ligero.Claim, error) {
	if len(input) == 0 || len(shares) == 0 {
		return nil, fmt.Errorf("Invalid input when forming claims: Input is empty")
	}

	sh := make([]int, len(shares))
	for i := 0; i < len(shares); i++ {
		sh[i] = shares[i].Value
	}
	claims := []ligero.Claim{{Secrets: input, Shares: sh}}

	return claims, nil
}

func (c *ClientRequest) ToJson() []byte {
	msg := &ClientRequest{
		Exp_ID:       c.Exp_ID,
		Client_ID:    c.Client_ID,
		Secret_Share: c.Secret_Share,
		Proof:        c.Proof,
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
