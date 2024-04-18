package main

import (
	"fmt"
	"log"
	"time"

	"example.com/SMC/pkg/ligero"
)

func main() {

	zk, err := ligero.NewLigeroZK(100, 4, 4, 1, 10631, 240)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	secrets := []int{0, 1, 1, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 1, 1, 0, 0, 0, 1, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 1, 0, 1, 1, 0, 1, 0, 0, 0, 1, 1, 1, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 1, 0, 1}

	start := time.Now()
	proof, err := zk.GenerateProof(secrets)
	end := time.Since(start)
	fmt.Printf("main end: %v\n", end)

	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(proof); i++ {

		verify, err := zk.VerifyProof(*proof[i])
		if err != nil {
			fmt.Printf("verification failed for party %d\n", proof[i].PartyShares[0].Index)
			log.Fatal(err)
		}
		if !verify {
			fmt.Printf("verification failed for party %d\n", proof[i].PartyShares[0].Index)
		}
		fmt.Printf("verification succeed for party %d\n", proof[i].PartyShares[0].Index)
	}

}
