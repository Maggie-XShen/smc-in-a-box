package ligero

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	merkletree "github.com/wealdtech/go-merkletree"
)

/**
func TestRearrange_Extended_Witness(t *testing.T) {
	tests := []struct {
		input    []int
		m        int
		expected [][]int
		wantErr  bool
	}{
		// Test case 1: Valid input
		{[]int{1, 2, 3, 4, 5, 6}, 2, [][]int{{1, 2, 3}, {4, 5, 6}}, false},

		// Test case 1: Valid input
		{[]int{1, 2, 3, 4, 5}, 2, [][]int{{1, 2, 3}, {4, 5, 0}}, false},

		// Test case 3: Invalid input (len(slice)< m)
		{[]int{1}, 2, nil, true},
	}

	for _, test := range tests {
		result, err := rearrange_input(test.input, test.m)

		// Check if an error is expected
		if (err != nil) != test.wantErr {
			t.Errorf("Expected error: %v, but got error: %v", test.wantErr, err)
			continue
		}

		// Check if the result matches the expected output
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Expected %v, but got %v", test.expected, result)
		}
	}
}**/

func TestPrepare_Extended_Witness(t *testing.T) {
	zk, err := NewLigeroZK(3, 1, 6, 1, 41, 3)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	claims1 := []Claim{{Secrets: []int{0}, Shares: []int{11, 21, 31, 8, 39, 15}}, {Secrets: []int{1}, Shares: []int{10, 24, 31, 18, 35, 17}}, {Secrets: []int{0}, Shares: []int{5, 26, 13, 26, 31, 18}}}
	result1, _ := zk.prepare_extended_witness(claims1)
	expected1 := [][]int{{0, 1, 0}, {11, 10, 5}, {21, 24, 26}, {31, 31, 13}, {8, 18, 26}, {39, 35, 31}, {15, 17, 18}}

	// Check if the result matches the expected output
	if !reflect.DeepEqual(result1, expected1) {
		t.Errorf("Expected %v, but got %v", expected1, result1)
	}

	zk, err = NewLigeroZK(1, 1, 6, 1, 41, 3)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	claims2 := []Claim{{Secrets: []int{0, 1, 0}, Shares: []int{11, 21, 31, 8, 39, 15}}}
	result2, _ := zk.prepare_extended_witness(claims2)
	expected2 := [][]int{{0}, {1}, {0}, {11}, {21}, {31}, {8}, {39}, {15}}

	// Check if the result matches the expected output
	if !reflect.DeepEqual(result2, expected2) {
		t.Errorf("Expected %v, but got %v", expected2, result2)
	}

	zk, err = NewLigeroZK(4, 2, 6, 1, 41, 3)

	if err != nil {
		t.Fatal(err)
	}

	claims3 := []Claim{{
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
	}

	result3, _ := zk.prepare_extended_witness(claims3)
	expected3 := [][]int{{0, 1}, {1, 0}, {11, 10}, {21, 24}, {31, 31}, {8, 18}, {39, 35}, {15, 17}, {0, 1}, {0, 1}, {5, 11}, {26, 21}, {13, 31}, {26, 8}, {31, 39}, {18, 15}}

	// Check if the result matches the expected output
	if !reflect.DeepEqual(result3, expected3) {
		t.Errorf("Expected %v, but got %v", expected3, result3)
	}
}

func TestEncode_Extended_Witness(t *testing.T) {
	zk, err := NewLigeroZK(3, 1, 6, 1, 41, 3)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	claims := []Claim{{Secrets: []int{0}, Shares: []int{6, 9, 12, 15, 18, 21}}, {Secrets: []int{1}, Shares: []int{2, 23, 3, 24, 4, 25}}, {Secrets: []int{0}, Shares: []int{1, 22, 2, 23, 3, 24}}}
	extended_witness, _ := zk.prepare_extended_witness(claims)

	if err != nil {
		log.Fatal(err)
	}

	encode, err := zk.encode_extended_witness(extended_witness)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v", encode)

}

func TestGenerate_MerkleTree(t *testing.T) {
	input := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	zk, err := NewLigeroZK(3, 1, 6, 1, 41, 3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	encoded_witeness_columnwise, err := ConvertToColumnwise(input)
	if err != nil {
		log.Fatal(err)
	}

	//commit to the Extended Witness via Merkle Tree
	tree, leaves, err := zk.generate_merkletree(encoded_witeness_columnwise)
	if err != nil {
		log.Fatal(err)
	}
	//get root of merkletree
	root := tree.Root()
	fmt.Printf("leaves: %v\n", leaves)

	for _, leaf := range leaves {
		proof, err := tree.GenerateProof(leaf)
		if err != nil {
			panic(err)
		}

		// Verify the proof for each leaf
		verified, err := merkletree.VerifyProof(leaf, proof, root)
		if err != nil {
			panic(err)
		}
		if !verified {
			panic("failed to verify proof")
		}
	}

}

func TestGenerate_Code_Proof(t *testing.T) {

	zk, err := NewLigeroZK(3, 1, 6, 1, 41, 3)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	claims := []Claim{{Secrets: []int{0}, Shares: []int{6, 9, 12, 15, 18, 21}}, {Secrets: []int{1}, Shares: []int{2, 23, 3, 24, 4, 25}}, {Secrets: []int{0}, Shares: []int{1, 22, 2, 23, 3, 24}}}
	extended_witness, _ := zk.prepare_extended_witness(claims)

	if err != nil {
		log.Fatal(err)
	}

	encode, err := zk.encode_extended_witness(extended_witness)

	if err != nil {
		log.Fatal(err)
	}

	seed1 := GenerateRandomness(zk.l, zk.q)
	code_mask := zk.generate_mask(seed1)
	randomness1 := GenerateRandomness(zk.m*(1+zk.n_server), zk.q)
	q_code, err := zk.generate_code_proof(encode, randomness1, code_mask)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v", q_code)

}

func TestGenerate(t *testing.T) {
	zk, err := NewLigeroZK(3, 1, 6, 1, 41, 3)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	claims := []Claim{{Secrets: []int{0}, Shares: []int{6, 9, 12, 15, 18, 21}}, {Secrets: []int{1}, Shares: []int{2, 23, 3, 24, 4, 25}}, {Secrets: []int{0}, Shares: []int{1, 22, 2, 23, 3, 24}}}

	proof, err := zk.Generate(claims)

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
