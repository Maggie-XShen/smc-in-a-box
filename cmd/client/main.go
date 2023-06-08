package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"example.com/SMC/cmd/client/config"
	"example.com/SMC/pkg/packed"
	"example.com/SMC/pkg/utils"
)

type Client struct {
	Client_ID string
}

func NewClient(id string) *Client {
	return &Client{Client_ID: id}
}

func (c *Client) generateShares(secrets []int, n, t, k, q int) ([]packed.Share, error) {
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

func main() {
	// read parameters from config file and command
	conf := config.GetConfig("config/config.json")

	cid := flag.String("cid", "c1", "client ID") // Todo: read from config
	flag.Parse()

	client := NewClient(*cid)
	shares, _ := client.generateShares(conf.Secrets, conf.N, conf.T, conf.K, conf.Q)

	urls := conf.URLs
	for i := 0; i < len(urls); i++ {
		current_time := time.Now().Format("2006-01-02")
		msg := utils.Client_Msg{Exp_ID: conf.Exp_ID, Client_ID: *cid, Secret_Share: shares[i], Timestamp: current_time}
		fmt.Printf("%+v\n", msg)
		writer := &msg

		utils.Send(urls[i], writer.WriteToJson())
	}

}
