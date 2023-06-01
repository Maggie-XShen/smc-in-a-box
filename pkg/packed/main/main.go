package main

import (
	"fmt"
	"log"

	"example.com/SMC/pkg/packed"
)

func main() {
	npss, err := packed.NewPackedSecretSharing(20, 8, 3, 41)
	if err != nil {
		log.Fatal(err)
	}

	secrets := [3]int{10, 25, 35}
	shares, err := npss.Split(secrets[:])
	if err != nil {
		log.Fatal(err)
	}

	result, err := npss.Reconstruct(shares)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)

	parts := shares[:11]
	result1, err := npss.Reconstruct(parts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result1)

}
