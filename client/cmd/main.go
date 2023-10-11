package main

import (
	"flag"

	"example.com/SMC/client/config"
)

func main() {
	//read configuration
	confpath := flag.String("confpath", "../config/client.json", "config file path")
	inputpath := flag.String("inputpath", "input.json", "client input path")
	flag.Parse()

	conf := config.Load(*confpath)

	client := NewClient(conf)

	client.Run(*inputpath)

}
