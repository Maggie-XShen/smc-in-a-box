package main

import (
	"fmt"
	"log"

	"example.com/SMC/pkg/ligero"
)

func main() {
	//NewLigeroZK(N_input, M, N_server, T, Q, N_open int)

	zk, err := ligero.NewLigeroZK(1, 1, 6, 1, 41, 3)
	//zk, err := ligero.NewLigeroZK(4, 2, 6, 1, 41, 3)

	if err != nil {
		log.Fatal(err)
	}

	/**
	claims := []ligero.Claim{{
		Secrets: []int{0, 1},
		Shares:  []int{11, 21, 31, 8, 39, 15},
	}, {
		Secrets: []int{1, 0},
		Shares:  []int{10, 24, 31, 18, 35, 17},
	}, {
		Secrets: []int{0, 0},
		Shares:  []int{5, 26, 13, 26, 31, 18},
	},
		{
			Secrets: []int{1, 1},
			Shares:  []int{11, 21, 31, 8, 39, 15}},
	}**/
	claims := []ligero.Claim{{
		Secrets: []int{0},
		Shares:  []int{6, 9, 12, 15, 18, 21},
	}}

	proof, err := zk.Generate(claims)

	//fmt.Printf("merkle root: %v\n", proof.MerkleRoot)
	//fmt.Printf("len: %d\n", len(proof.MerkleRoot))

	//fmt.Printf("Column check: %v\n", proof.ColumnCheck)
	//fmt.Printf("len auth path: %v\n", len(proof.ColumnCheck[0].Authpath))

	//fmt.Printf("q_code: %v\n", proof.Q_code)
	//fmt.Printf("q_quadra: %v\n", proof.Q_quadra)
	fmt.Printf("q_linear: %v\n", proof.LinearTest)
	//fmt.Printf("column check: %v\n", proof.ColumnCheck)

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
