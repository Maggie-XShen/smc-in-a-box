package ligero

import (
	"fmt"
	"log"
	"testing"
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
	zk, err := NewLigeroZK(2, 6, 1, 41, 8)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	result, err := zk.prepare_extended_witness(input, 2, 6, 1, 41)

	fmt.Printf("%v", result)

	if err != nil {
		log.Fatal(err)
	}

}

func TestEncode_Extended_Witness(t *testing.T) {
	input := []int{10, 25, 35}
	lg, err := NewLigeroZK(2, 6, 1, 41)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	result, err := prepare_extended_witness(input, 2, 6, 1, 41)

	if err != nil {
		log.Fatal(err)
	}

	encode, err := encode_extended_witness(result, 8, 1, 41)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v", encode)

}

func TestGenerate_Code_Proof(t *testing.T) {
	input := []int{10, 25, 35}
	result, err := prepare_extended_witness(input, 2, 6, 1, 41)

	if err != nil {
		log.Fatal(err)
	}

	encode, err := encode_extended_witness(result, 8, 1, 41)

	if err != nil {
		log.Fatal(err)
	}

	q_code, err := generate_code_proof(encode, 41)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v", q_code)

}
