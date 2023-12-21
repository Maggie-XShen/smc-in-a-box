package main

import (
	"fmt"
	"log"

	"example.com/SMC/pkg/rss"
)

func main() {
	secret := 10

	rss, err := rss.NewReplicatedSecretSharing(4, 1, 17)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	shares, err := rss.Split(secret)
	if err != nil {
		log.Fatalf("Split(%v) failed with error %s", secret, err)
	}

	fmt.Printf("%v\n", shares)

	recon, err := rss.Reconstruct(shares)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	fmt.Printf("%d", recon)
}
