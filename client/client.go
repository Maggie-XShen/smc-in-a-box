package client

import (
	"log"

	"example.com/SMC/pkg/packed"
)

type Client struct {
	cfg Config
}

func NewClient(conf Config) *Client {
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
