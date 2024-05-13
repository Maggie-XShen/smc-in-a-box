package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"

	"example.com/SMC/client/config"
	"example.com/SMC/pkg/ligero"
	"github.com/sirupsen/logrus"
)

type Client struct {
	cfg  *config.Client
	mode string
}

func NewClient(conf *config.Client, md string) *Client {
	return &Client{cfg: conf, mode: md}
}

func (c *Client) Run(inputpath string) {
	inputs := ReadClientInput(inputpath)
	urls := c.cfg.URLs

	zk, err := ligero.NewLigeroZK(c.cfg.N_secrets, c.cfg.M, c.cfg.N, c.cfg.T, c.cfg.Q, c.cfg.N_open)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	for _, input := range inputs {

		/**
		//test c1's input is malformed
		if c.cfg.Client_ID == "c1" {
			input.Secrets = []int{100}
		}**/

		proof_start := time.Now() //Proof generation start time

		proof, err := zk.GenerateProof(input.Secrets)

		proof_end := time.Since(proof_start) //Proof generation end time

		if err != nil {
			log.Fatal(err)
		}

		theorySingleProofBytes, theoryInputSharesBytes := zk.GetSize(*proof[0]) //Theoretical proof size

		logger.WithFields(logrus.Fields{
			"input":             input,
			"proof_time":        proof_end.String(),
			"proof_size":        big.NewInt(theorySingleProofBytes),
			"input_shares_size": big.NewInt(theoryInputSharesBytes),
		}).Info("")

		current_time := time.Now().UTC()
		var wg sync.WaitGroup
		for i := 0; i < len(urls); i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				/**
				  //test c1 not sending data to s1
				  if idx == 0 && c.cfg.Client_ID == "c1" || idx == 1 && c.cfg.Client_ID == "c2" {
				      return
				  }**/

				//test client's proof is malformed
				var msg ClientRequest
				if c.mode == "malicious" && idx == 0 {
					mal_proof := proof[idx]
					mal_proof.CodeTest = make([]int, len(proof[0].CodeTest))

					msg = ClientRequest{Exp_ID: input.Exp_ID, Client_ID: c.cfg.Client_ID, Token: c.cfg.Token, Proof: *mal_proof, Timestamp: current_time.String()}
				} else {
					msg = ClientRequest{Exp_ID: input.Exp_ID, Client_ID: c.cfg.Client_ID, Token: c.cfg.Token, Proof: *proof[idx], Timestamp: current_time.String()}
				}

				writer := &msg
				log.Printf("client %s is sending data of %s to server%d ...\n", msg.Client_ID, msg.Exp_ID, msg.Proof.Shares.PartyIndex)
				c.Send(urls[idx], writer.ToJson())
			}(i)

		}
		wg.Wait()
	}

}

func (c *Client) Send(address string, data []byte) {
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
