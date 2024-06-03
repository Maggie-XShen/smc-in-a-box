package ligero

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestMulMatrix(t *testing.T) {
	tests := []struct {
		matrix1  [][]int
		matrix2  [][]int
		q        int
		expected [][]int
		wantErr  error
	}{
		// Test case 1: Valid input
		{[][]int{{1, 2, 3}}, [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}, 41, [][]int{{30, 36, 1}}, nil},
		// Test case 2: Valid input
		{[][]int{{0, 1, 2}, {4, 5, 6}, {8, 9, 10}}, [][]int{{10, 11}, {13, 14}, {16, 17}}, 41, [][]int{{4, 7}, {37, 11}, {29, 15}}, nil},
		// Test case 3: Invalid input
		{[][]int{{1, 2}}, [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}, 41, nil, fmt.Errorf("Matrix multiplication is not possible. The number of columns in the first matrix must be equal to the number of rows in the second matrix.")},
	}

	for _, test := range tests {
		result, err := MulMatrix(test.matrix1, test.matrix2, test.q)

		// Check if an error is expected
		if err != nil {
			if errors.Is(test.wantErr, err) {
				t.Errorf("Expected error: %v, but got error: %v", test.wantErr, err)
				continue
			}
		}

		// Check if the result matches the expected output
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Expected %v, but got %v", test.expected, result)
		}
	}
}
