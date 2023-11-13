package main

import (
	"fmt"
	"log"

	"example.com/SMC/pkg/ligero"
)

func main() {
	input := []int{0}
	//NewLigeroZK(N_input, M, N_server, T, Q, N_open int)
	zk, err := ligero.NewLigeroZK(1, 1, 6, 1, 41, 3)

	if err != nil {
		log.Fatal(err)
	}

	proof, err := zk.Generate(input)

	fmt.Printf("merkle root: %v\n", proof.MerkleRoot)
	fmt.Printf("q_code: %v\n", proof.Q_code)
	fmt.Printf("q_quadra: %v\n", proof.Q_quadra)
	fmt.Printf("q_linear: %v\n", proof.Q_linear)
	fmt.Printf("column check: %v\n", proof.ColumnCheck)

	if err != nil {
		log.Fatal(err)
	}

	verify, err := zk.Verify(*proof)
	if err != nil {
		log.Fatal(err)
	}
	if !verify {
		fmt.Println("failed verifification!")
	}
	fmt.Println("verification succeed!")

}
