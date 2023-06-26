package main

import (
	"log"

	"example.com/SMC/cmd/client/config"
	"example.com/SMC/pkg/packed"
)

type Client struct {
	cfg *config.Client
}

func NewClient(conf *config.Client) *Client {
	return &Client{cfg: conf}
}

func (c *Client) GenerateShares() ([]packed.Share, error) {
	npss, err := packed.NewPackedSecretSharing(c.cfg.N, c.cfg.T, c.cfg.K, c.cfg.Q)
	if err != nil {
		log.Fatal(err)
	}

	shares, err := npss.Split(c.cfg.Secrets[:])
	if err != nil {
		log.Fatal(err)
	}

	return shares, nil
}
