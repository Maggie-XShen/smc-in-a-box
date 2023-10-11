package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"example.com/SMC/client/config"
	"example.com/SMC/pkg/packed"
)

type Client struct {
	cfg *config.Client
}

func NewClient(conf *config.Client) *Client {
	return &Client{cfg: conf}
}

func (c *Client) GenerateShares(secrets []int) ([]packed.Share, error) {
	npss, err := packed.NewPackedSecretSharing(c.cfg.N, c.cfg.T, c.cfg.K, c.cfg.Q)
	if err != nil {
		log.Fatal(err)
	}

	shares, err := npss.Split(secrets)
	if err != nil {
		log.Fatal(err)
	}

	return shares, nil
}

func (c *Client) Run(inputpath string) {
	inputs := ReadClientInput(inputpath)
	urls := c.cfg.URLs

	for _, input := range inputs {
		shares, _ := c.GenerateShares(input.Secrets)
		current_time := time.Now().Format("2006-01-02 15:04:05")
		for i := 0; i < len(urls); i++ {
			msg := ClientRequest{Exp_ID: input.Exp_ID, Client_ID: c.cfg.Client_ID, Token: c.cfg.Token, Secret_Share: shares[i], Timestamp: current_time}
			fmt.Printf("client sends: %+v\n", msg)
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
