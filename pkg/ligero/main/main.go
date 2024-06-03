package main

import (
	"fmt"
	"log"
	"time"

	"example.com/SMC/pkg/ligero"
)

var verify_start time.Time
var verify_end time.Duration

func main() {

	zk, err := ligero.NewLigeroZK(1000000, 2000, 4, 1, 41543, 240)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	secrets := make([]int, 100000)
	for i := 0; i < 100000; i++ {
		secrets[i] = 1
	}

	start := time.Now()
	proof, err := zk.GenerateProof(secrets)
	end := time.Since(start)
	proof_size, share_size := zk.GetSize(*proof[0])
	fmt.Printf("share size: %d, proof size: %d\n", share_size, proof_size)

	fmt.Printf("proof generation end: %v\n", end)

	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(proof); i++ {

		if i == 0 {
			verify_start = time.Now()
		}

		verify, err := zk.VerifyProof(*proof[i])
		if i == 0 {
			verify_end = time.Since(verify_start)
			fmt.Printf("proof verification end: %v\n", verify_end)
		}

		if err != nil {
			fmt.Printf("verification failed for party %d\n", proof[i].Shares.PartyIndex)
			log.Fatal(err)
		}
		if !verify {
			fmt.Printf("verification failed for party %d\n", proof[i].Shares.PartyIndex)
		}
		fmt.Printf("verification succeed for party %d\n", proof[i].Shares.PartyIndex)
	}

}
