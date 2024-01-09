package ligero

import (
	"fmt"
	"log"
	"strings"

	"example.com/SMC/pkg/rss"
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

func ConvertSharesToColumnwise(shares [][]rss.Share) ([][]rss.Share, error) {
	if len(shares) == 0 {
		return nil, fmt.Errorf("shares cannot be empty")
	}
	result := make([][]rss.Share, len(shares[0]))
	for j := 0; j < len(shares[0]); j++ {
		temp := make([]rss.Share, len(shares))
		for i := 0; i < len(shares); i++ {
			temp[i] = shares[i][j]
		}
		result[j] = temp
	}
	return result, nil
}

func ConvertToByteArray(input []int) []byte {
	size := len(input)
	if size == 0 {
		log.Fatal("cannot convert empty integer array to byte array")
	}

	list := make([]string, size)
	for j := 0; j < size; j++ {
		list[j] = fmt.Sprintf("%064b", input[j])
	}

	//concatenate values in the column to a string
	concatenated := strings.Join(list, "")
	return []byte(concatenated)
}
