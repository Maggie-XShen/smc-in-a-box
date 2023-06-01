package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/SMC/pkg/packed"
)

type Configuration struct {
	CID     string
	URLs    []string
	N       int
	T       int
	K       int
	Q       int
	EID     string
	Secrets []int
}

type Client struct {
	CID string
}

type Msg struct {
	EID         string       `json:"EID"`
	CID         string       `json:"CID"`
	SecretShare packed.Share `json:"SecretShare"`
	Timestamp   string       `json:"Timestamp"`
	//Proof       string       `json:"Proof"`
	//Hash_proof  string       `json:"HashProof"`
	//Signature   string       `json:"Signature"`
}

func NewClient(id string) *Client {
	return &Client{CID: id}
}

func (c *Client) send(address string, data []byte) {
	req, err := http.NewRequest("POST", address, bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("impossible to build request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("impossible to send request: %s", err)
	}
	log.Printf("response Status:%s", res.Status)

	//defer res.Body.Close()
	//body, _ := io.ReadAll(res.Body)
	//fmt.Println("response Body:", string(body))

}

func (c *Client) generateShares(secrets []int, n, t, k, q int) ([]packed.Share, error) {
	//Todo: read parameters for packed secret sharing from file
	npss, err := packed.NewPackedSecretSharing(n, t, k, q)
	if err != nil {
		log.Fatal(err)
	}

	shares, err := npss.Split(secrets[:])
	if err != nil {
		log.Fatal(err)
	}

	return shares, nil
}

func (c *Client) write(eid string, cid string, share packed.Share) ([]byte, error) {
	currentTime := time.Now()
	msg := &Msg{
		EID:         eid,
		CID:         cid,
		SecretShare: share,
		Timestamp:   currentTime.Format("2006-01-02"), // Todo: decide time format
	}
	message, err := json.Marshal(msg)

	if err != nil {
		log.Fatalf("impossible to marshall response: %s", err)
	}

	return message, nil
}

func main() {
	// read from config file
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Configuration{}
	err := decoder.Decode(&config)
	if err != nil {
		log.Fatalf("unable to read from config file: %s", err)
	}

	// get client id from command, need to change to read from config
	cid := flag.String("cid", "c1", "client ID")
	flag.Parse()

	client := NewClient(*cid)

	secrets := config.Secrets

	//secrete sharing
	shares, _ := client.generateShares(secrets, config.N, config.T, config.K, config.Q)

	urls := config.URLs
	for i := 0; i < len(urls); i++ {
		message, _ := client.write(config.EID, *cid, shares[i])
		fmt.Println(string(message))

		client.send(urls[i], message)
	}

}
