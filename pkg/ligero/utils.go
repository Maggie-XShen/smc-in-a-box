package ligero

import (
	"fmt"
	"strings"
)

func ConvertToColumnwise(matrix [][]int) ([][]int, error) {
	if len(matrix) == 0 {
		return nil, fmt.Errorf("matrix cannot be empty")
	}
	result := make([][]int, len(matrix[0]))
	for j := 0; j < len(matrix[0]); j++ {
		temp := make([]int, len(matrix))
		for i := 0; i < len(matrix); i++ {
			temp[i] = matrix[i][j]
		}
		result[j] = temp
	}
	return result, nil
}

func ConvertColumnToString(list []int) (string, error) {
	if len(list) == 0 {
		return "", fmt.Errorf("list cannot be empty")
	}

	col := make([]string, len(list))
	for i := 0; i < len(list); i++ {
		col[i] = fmt.Sprintf("%064b", list[i])
	}
	//concatenate values in the column to a string
	concatenated := strings.Join(col, "")

	return concatenated, nil

}
