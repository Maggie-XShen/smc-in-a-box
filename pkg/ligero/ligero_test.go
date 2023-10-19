package ligero

import (
	"fmt"
	"log"
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
	input := []int{10, 25, 35}
	zk, err := NewLigeroZK(3, 2, 6, 1, 41, 3)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	result, err := zk.prepare_extended_witness(input)

	fmt.Printf("%v", result)

	if err != nil {
		log.Fatal(err)
	}

}

func TestEncode_Extended_Witness(t *testing.T) {
	input := []int{10, 25, 35}
	zk, err := NewLigeroZK(3, 2, 6, 1, 41, 3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	extended_witness, err := zk.prepare_extended_witness(input)

	if err != nil {
		log.Fatal(err)
	}

	encode, err := zk.encode_extended_witness(extended_witness)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v", encode)

}

func TestGenerate_Code_Proof(t *testing.T) {
	input := []int{10, 25, 35}
	zk, err := NewLigeroZK(3, 2, 6, 1, 41, 3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	extended_witness, err := zk.prepare_extended_witness(input)

	if err != nil {
		log.Fatal(err)
	}

	encode, err := zk.encode_extended_witness(extended_witness)

	if err != nil {
		log.Fatal(err)
	}

	randomness1 := GenerateRandomness(zk.m*(1+zk.n_server), zk.q)
	q_code, err := zk.generate_code_proof(encode, randomness1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v", q_code)

}

func TestGenerate_MerkleTree(t *testing.T) {
	input := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	zk, err := NewLigeroZK(3, 2, 6, 1, 41, 3)
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
