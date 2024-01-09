package main

import (
	"fmt"
	"log"

	"example.com/SMC/pkg/ligero"
)

func main() {
	zk, err := ligero.NewLigeroZK(3, 1, 6, 1, 41, 3)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	secrets := []int{1, 1, 0}

	proof, err := zk.GenerateProof(secrets)

	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(proof); i++ {

		verify, err := zk.VerifyProof(*proof[i])
		if err != nil {
			fmt.Printf("verification failed for party %d\n", *&proof[i].PartyShares[0].Index)
			log.Fatal(err)
		}
		if !verify {
			fmt.Printf("verification failed for party %d\n", *&proof[i].PartyShares[0].Index)
		}
		fmt.Printf("verification succeed for party %d\n", *&proof[i].PartyShares[0].Index)
	}

}
