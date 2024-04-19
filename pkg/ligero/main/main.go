package main

import (
	"fmt"
	"log"
	"time"

	"example.com/SMC/pkg/ligero"
)

var verify_end time.Duration

func main() {

	zk, err := ligero.NewLigeroZK(100, 4, 4, 1, 10631, 240)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	secrets := []int{0, 1, 1, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 1, 1, 0, 0, 0, 1, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 1, 0, 1, 1, 0, 1, 0, 0, 0, 1, 1, 1, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 1, 0, 1}

	/**
	secrets := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		secrets[i] = 1
	}**/

	start := time.Now()
	proof, err := zk.GenerateProof(secrets)
	end := time.Since(start)

	fmt.Printf("proof generation end: %v\n", end)

	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(proof); i++ {

		verify_start := time.Now()
		verify, err := zk.VerifyProof(*proof[i])
		verify_end = time.Since(verify_start)
		if err != nil {
			fmt.Printf("verification failed for party %d\n", proof[i].PartyShares[0].Index)
			log.Fatal(err)
		}
		if !verify {
			fmt.Printf("verification failed for party %d\n", proof[i].PartyShares[0].Index)
		}
		fmt.Printf("verification succeed for party %d\n", proof[i].PartyShares[0].Index)
	}
	fmt.Printf("proof verification end: %v\n", verify_end)
}
