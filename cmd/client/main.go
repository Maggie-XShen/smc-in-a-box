package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"example.com/SMC/cmd/client/config"
	"example.com/SMC/pkg/message"
	"example.com/SMC/pkg/packed"
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
	//read configuration
	confpath := flag.String("confpath", "config/config.json", "config file path") // confpath := "config.json"
	flag.Parse()
	conf := config.LoadConfig(*confpath)

	client := NewClient(conf.Client_ID)
	shares, _ := client.generateShares(conf.Secrets, conf.N, conf.T, conf.K, conf.Q)

	urls := conf.URLs
	for i := 0; i < len(urls); i++ {
		current_time := time.Now().Format("2006-01-02")
		msg := message.Client_Msg{Exp_ID: conf.Exp_ID, Client_ID: conf.Client_ID, Secret_Share: shares[i], Timestamp: current_time}
		fmt.Printf("%+v\n", msg)
		writer := &msg

		message.Send(urls[i], writer.WriteToJson())
	}

}
