package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"time"

	"example.com/SMC/client/config"
	"example.com/SMC/pkg/ligero"
	"github.com/sirupsen/logrus"
)

type Client struct {
	cfg *config.Client
}

func NewClient(conf *config.Client) *Client {
	return &Client{cfg: conf}
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

		theorySingleProofBytes := big.NewInt(zk.GetProofSize(*proof[0])) //Theoretical proof size

		logger.WithFields(logrus.Fields{
			"input":      input,
			"proof_time": proof_end.String(),
			"proof_size": theorySingleProofBytes,
		}).Info("")

		current_time := time.Now().UTC().Format("2006-01-02 15:04:05")
		for i := 0; i < len(urls); i++ {

			/**
			//test c1 not sending data to s1
			if i == 0 && c.cfg.Client_ID == "c1" || i == 1 && c.cfg.Client_ID == "c2" {
				continue
			}**/

			/**
			//test c1's proof is malformed
			if c.cfg.Client_ID == "c1" {
				temp := proof[i].CodeTest[0]
				proof[i].CodeTest[0] = temp%c.cfg.Q - 3
			}**/

			msg := ClientRequest{Exp_ID: input.Exp_ID, Client_ID: c.cfg.Client_ID, Token: c.cfg.Token, Proof: *proof[i], Timestamp: current_time}

			writer := &msg

			log.Printf("client %s is sending data of %s to server%d ...\n", msg.Client_ID, msg.Exp_ID, msg.Proof.PartyShares[0].Index+1)
			c.Send(urls[i], writer.ToJson())

		}

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
