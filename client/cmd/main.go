package main

import (
	"flag"
	"fmt"
	"time"

	"example.com/SMC/client"
	"example.com/SMC/pkg/message"
)

func main() {
	//read configuration
	confpath := flag.String("confpath", "../config.json", "config file path") // confpath := "config.json"
	flag.Parse()
	conf := client.LoadConfig(*confpath)

	client := client.NewClient(*conf)
	shares, _ := client.GenerateShares()

	urls := conf.URLs
	for i := 0; i < len(urls); i++ {
		current_time := time.Now().Format("2006-01-02 15:04:05")
		msg := message.ClientRequest{Exp_ID: conf.Exp_ID, Client_ID: conf.Client_ID, Token: conf.Token, Secret_Share: shares[i], Timestamp: current_time}
		fmt.Printf("%+v\n", msg)
		writer := &msg

		message.Send(urls[i], writer.WriteJson())
	}

}
