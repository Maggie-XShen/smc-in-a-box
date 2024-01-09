package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"example.com/SMC/client/config"
	"example.com/SMC/pkg/ligero"
)

type Client struct {
	cfg *config.Client
}

func NewClient(conf *config.Client) *Client {
	return &Client{cfg: conf}
}

/**
func (c *Client) PrepareShares(secrets []int) ([][]int, [][]rss.Share, error) {
	nrss, err := rss.NewReplicatedSecretSharing(c.cfg.N, c.cfg.T, c.cfg.Q)
	if err != nil {
		return nil, nil, err
	}

	size := len(secrets)
	vals_per_secret := make([][]int, size)
	shares_per_party := make([][]rss.Share, c.cfg.N)
	for i := 0; i < size; i++ {
		values, shares, err := nrss.Split(secrets[i])
		if err != nil {
			return nil, nil, err
		}
		vals_per_secret[i] = values
		for j := 0; j < len(shares); j++ {
			shares_per_party[j][i] = shares[j]
		}
	}

	return vals_per_secret, shares_per_party, nil
}

func (c *Client) GenerateZKP(secrets []int, values [][]int) (*ligero.Proof, error) {
	//NewLigeroZK(N_input, M, N_server, T, Q, N_open int)
	zk, err := ligero.NewLigeroZK(c.cfg.N_claims, c.cfg.M, c.cfg.N, c.cfg.T, c.cfg.Q, c.cfg.N_open)
	if err != nil {
		return nil, err
	}

	claims, err := FormClaims(secrets, values)
	if err != nil {
		return nil, err
	}

	proof, err := zk.Generate(claims)
	if err != nil {
		return nil, err
	}

	return proof, nil
}**/

func (c *Client) Run(inputpath string) {
	inputs := ReadClientInput(inputpath)
	urls := c.cfg.URLs

	for _, input := range inputs {
		zk, err := ligero.NewLigeroZK(c.cfg.N_secrets, c.cfg.M, c.cfg.N, c.cfg.T, c.cfg.Q, c.cfg.N_open)
		if err != nil {
			log.Fatalf("err: %v", err)
		}

		proof, err := zk.GenerateProof(input.Secrets)

		if err != nil {
			log.Fatal(err)
		}

		current_time := time.Now().Format("2006-01-02 15:04:05")
		for i := 0; i < len(urls); i++ {
			msg := ClientRequest{Exp_ID: input.Exp_ID, Client_ID: c.cfg.Client_ID, Token: c.cfg.Token, Proof: *proof[i], Timestamp: current_time}

			writer := &msg
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
		log.Fatalf("impossible to send http request: %s", err)
	}

	log.Printf("response Status:%s", res.Status)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if len(body) > 0 {
		fmt.Println("response Body:", string(body))
	}

}
