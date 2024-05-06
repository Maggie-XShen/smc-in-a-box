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

/**
func TestPrepare_Extended_Witness(t *testing.T) {
	zk, err := NewLigeroZK(7, 2, 6, 1, 41, 3)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	claims1 := []Claim{{Secret: 1, Shares: []int{9, 14, 15, 8, 0, 0, 14, 16, 11, 8, 12, 8, 9, 5, 8, 2, 7, 8, 11, 3, 3}}, {Secret: 0, Shares: []int{9, 14, 14, 4, 4, 15, 10, 10, 0, 13, 3, 13, 8, 3, 5, 14, 11, 15, 15, 12, 12}}, {Secret: 1, Shares: []int{9, 16, 1, 10, 9, 11, 9, 14, 3, 11, 12, 4, 10, 3, 10, 6, 13, 10, 6, 13, 8}}}
	result1, _ := zk.prepare_extended_witness(claims1)
	expected1 := [][]int{{0, 1, 0}, {11, 10, 5}, {21, 24, 26}, {31, 31, 13}, {8, 18, 26}, {39, 35, 31}, {15, 17, 18}}

	// Check if the result matches the expected output
	if !reflect.DeepEqual(result1, expected1) {
		t.Errorf("Expected %v, but got %v", expected1, result1)
	}

}**/

/**
func TestEncode_Extended_Witness(t *testing.T) {
	zk, err := NewLigeroZK(1, 1, 6, 1, 41, 3)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	nrss, err := rss.NewReplicatedSecretSharing(6, 1, 41)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	secrets := []int{0}
	sh, _, err := nrss.Split(secrets[0])
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	shares := make([][]rss.Share, 1)
	shares[0] = sh
	claims, err := FormClaims(secrets, shares)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	extended_witness, _ := zk.prepare_extended_witness(claims)

	if err != nil {
		log.Fatal(err)
	}

	key := make([]int, zk.n_shares+1)
	for k := 0; k < zk.n_shares+1; k++ {
		r, err := rand.Int(rand.Reader, big.NewInt(int64(zk.q)))
		if err != nil {
			log.Fatal(err)
		}
		key[k] = int(r.Int64())
	}
	encode, err := zk.encode_extended_witness(extended_witness, key)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v", encode)

}**/

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
	tree, leaves, _, err := zk.generate_merkletree(encoded_witeness_columnwise)
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

func TestGenerate(t *testing.T) {
	zk, err := NewLigeroZK(3, 1, 6, 1, 10631, 3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	secrets := []int{1, 0, 1}

	proof, err := zk.GenerateProof(secrets)

	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(proof); i++ {
		verify, err := zk.VerifyProof(*proof[i])
		if err != nil {
			log.Fatal(err)
		}
		if !verify {
			fmt.Println("Verification failed for party!")
		}
		fmt.Println("Verification succeed for party !")
	}

}

func BenchmarkGenerateProof(b *testing.B) {
	for i := 0; i < b.N; i++ {
		zk, err := NewLigeroZK(100, 4, 4, 1, 10631, 240)
		if err != nil {
			log.Fatalf("err: %v", err)
		}

		secrets := []int{0, 1, 1, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 1, 0, 0, 0, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 1, 1, 0, 0, 0, 1, 1, 0, 1, 0, 0, 0, 1, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 0, 0, 0, 1, 0, 1, 1, 0, 1, 0, 0, 0, 1, 1, 1, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 1, 0, 0, 1, 0, 1}

		proof, err := zk.GenerateProof(secrets)

		if err != nil {
			log.Fatal(err)
		}
		zk.GetSize(*proof[0])

		//fmt.Printf("%d\n", zk.GetProofSize(*proof[0]))
	}
}
