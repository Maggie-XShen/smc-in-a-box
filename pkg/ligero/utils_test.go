package ligero

import (
	"reflect"
	"testing"
)

func Test_ConvertToColumnwise(t *testing.T) {

	tests := []struct {
		matrix   [][]int
		expected [][]int
		wantErr  bool
	}{
		// Test case 1: Valid input
		{[][]int{{1, 2, 3}, {4, 5, 6}}, [][]int{{1, 4}, {2, 5}, {3, 6}}, false},

		// Test case 2: Valid input
		{[][]int{{1}}, [][]int{{1}}, false},
	}

	for _, test := range tests {
		result, err := ConvertToColumnwise(test.matrix)

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

}
